// Pacote auth implementa o módulo de autenticação (registro, login, tokens e login social).
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/config"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/lib/firebase"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/lib/hasher"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/model"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotActive      = errors.New("user account is not active")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrPasswordMismatch   = errors.New("current password is incorrect")
	ErrWeakPassword       = errors.New("password does not meet requirements")
	ErrPasswordAlreadySet = errors.New("user already has a password set")
)

type JWTClaims struct {
	UserID uuid.UUID      `json:"user_id"`
	Email  string         `json:"email"`
	Role   model.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type Service struct {
	userRepo       *repository.UserRepository
	providerRepo   *repository.OAuthProviderRepository
	firebaseClient *firebase.Client
	config         *config.Config
	logger         *zerolog.Logger
}

func NewService(
	userRepo *repository.UserRepository,
	providerRepo *repository.OAuthProviderRepository,
	firebaseClient *firebase.Client,
	cfg *config.Config,
	logger *zerolog.Logger,
) *Service {
	return &Service{
		userRepo:       userRepo,
		providerRepo:   providerRepo,
		firebaseClient: firebaseClient,
		config:         cfg,
		logger:         logger,
	}
}

// Register creates a new user account and returns authentication tokens
func (s *Service) Register(ctx context.Context, req *model.RegisterRequest) (*model.LoginResponse, error) {
	// Hash password
	passwordHash, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Email:             req.Email,
		PasswordHash:      passwordHash,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		PreferredCurrency: "BRL",
		PreferredLanguage: "pt-BR",
		Timezone:          "America/Sao_Paulo",
		Role:              model.UserRoleUser,
		IsActive:          true,
	}

	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			// Verificar se o usuário existente está inativo (soft-deleted)
			existingUser, getErr := s.userRepo.GetByEmail(ctx, req.Email)
			if getErr == nil && !existingUser.IsActive {
				// Reativar e atualizar credenciais
				if err := s.userRepo.Reactivate(ctx, existingUser.ID); err != nil {
					return nil, fmt.Errorf("failed to reactivate user: %w", err)
				}
				if err := s.userRepo.UpdatePassword(ctx, existingUser.ID, passwordHash); err != nil {
					return nil, fmt.Errorf("failed to update password: %w", err)
				}
				existingUser.IsActive = true
				existingUser.FirstName = req.FirstName
				existingUser.LastName = req.LastName
				user = existingUser
				s.logger.Info().Str("user_id", user.ID.String()).Msg("user reactivated via re-registration")
			} else {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Create default settings apenas para usuários novos
		_, err = s.userRepo.CreateDefaultSettings(ctx, user.ID)
		if err != nil {
			s.logger.Warn().Err(err).Msg("failed to create default settings for user")
		}
	}

	// Generate tokens for auto-login after registration
	accessToken, expiresAt, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshTokenHash, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session
	session := &model.UserSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        time.Now().Add(time.Duration(s.config.Auth.RefreshTokenDuration) * time.Hour),
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Msg("new user registered")

	return &model.LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, req *model.LoginRequest, ipAddress, userAgent string) (*model.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.IsActive {
		return nil, ErrUserNotActive
	}

	valid, needsMigration := s.checkPassword(req.Password, user.PasswordHash)
	if !valid {
		return nil, ErrInvalidCredentials
	}

	// Migração silenciosa bcrypt → argon2id
	if needsMigration {
		if newHash, hashErr := s.hashPassword(req.Password); hashErr != nil {
			s.logger.Warn().Err(hashErr).Str("user_id", user.ID.String()).Msg("failed to generate argon2id hash during migration")
		} else if updateErr := s.userRepo.UpdatePassword(ctx, user.ID, newHash); updateErr != nil {
			s.logger.Warn().Err(updateErr).Str("user_id", user.ID.String()).Msg("failed to migrate password to argon2id")
		} else {
			s.logger.Info().Str("user_id", user.ID.String()).Msg("password migrated from bcrypt to argon2id")
		}
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	accessToken, expiresAt, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshTokenHash, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session
	session := &model.UserSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		IPAddress:        &ipAddress,
		UserAgent:        &userAgent,
		ExpiresAt:        time.Now().Add(time.Duration(s.config.Auth.RefreshTokenDuration) * time.Hour),
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Msg("user logged in")

	return &model.LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// RefreshToken refreshes the access token using a refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*model.RefreshTokenResponse, error) {
	tokenHash := s.hashRefreshToken(refreshToken)

	session, err := s.userRepo.GetSessionByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) ||
			errors.Is(err, repository.ErrSessionExpired) ||
			errors.Is(err, repository.ErrSessionRevoked) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.IsActive {
		return nil, ErrUserNotActive
	}

	// Revoke old session
	_ = s.userRepo.RevokeSession(ctx, session.ID)

	// Generate new tokens
	accessToken, expiresAt, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, newRefreshTokenHash, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create new session
	newSession := &model.UserSession{
		UserID:           user.ID,
		RefreshTokenHash: newRefreshTokenHash,
		IPAddress:        session.IPAddress,
		UserAgent:        session.UserAgent,
		ExpiresAt:        time.Now().Add(time.Duration(s.config.Auth.RefreshTokenDuration) * time.Hour),
	}

	if err := s.userRepo.CreateSession(ctx, newSession); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &model.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Logout revokes the user's refresh token
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := s.hashRefreshToken(refreshToken)

	session, err := s.userRepo.GetSessionByToken(ctx, tokenHash)
	if err != nil {
		// Ignore errors, just return success
		return nil
	}

	return s.userRepo.RevokeSession(ctx, session.ID)
}

// LogoutAll revokes all sessions for a user
func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.RevokeAllUserSessions(ctx, userID)
}

// SetPassword defines a password for users who logged in via social provider and have no password yet
func (s *Service) SetPassword(ctx context.Context, userID uuid.UUID, req *model.SetPasswordRequest) error {
	hasPassword, err := s.providerRepo.CheckUserHasPassword(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check user password: %w", err)
	}

	if hasPassword {
		return ErrPasswordAlreadySet
	}

	newPasswordHash, err := s.hashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, newPasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Msg("user set password for the first time")

	return nil
}

// ChangePassword changes the user's password
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, req *model.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if valid, _ := s.checkPassword(req.CurrentPassword, user.PasswordHash); !valid {
		return ErrPasswordMismatch
	}

	newPasswordHash, err := s.hashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, newPasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke all sessions to force re-login
	_ = s.userRepo.RevokeAllUserSessions(ctx, userID)

	s.logger.Info().
		Str("user_id", userID.String()).
		Msg("user changed password")

	return nil
}

// ValidateAccessToken validates an access token and returns the claims
func (s *Service) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Auth.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// Helper methods

func (s *Service) hashPassword(password string) (string, error) {
	return hasher.HashArgon2(password)
}

// checkPassword verifica a senha e indica se precisa migrar de bcrypt para argon2id.
// needsMigration é true somente quando o hash era bcrypt e a senha está correta.
func (s *Service) checkPassword(password, hash string) (valid bool, needsMigration bool) {
	if hasher.IsArgon2Hash(hash) {
		return hasher.VerifyArgon2(password, hash), false
	}
	// Hash legado bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false, false
	}
	return true, true
}

func (s *Service) generateAccessToken(user *model.User) (string, int64, error) {
	expiresAt := time.Now().Add(time.Duration(s.config.Auth.AccessTokenDuration) * time.Minute)

	claims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.Auth.Issuer,
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.Auth.SecretKey))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

func (s *Service) generateRefreshToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	token := base64.URLEncoding.EncodeToString(bytes)
	hash := s.hashRefreshToken(token)

	return token, hash, nil
}

func (s *Service) hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// Social Login Methods

var (
	ErrFirebaseNotConfigured    = errors.New("firebase is not configured")
	ErrInvalidIDToken           = errors.New("invalid firebase ID token")
	ErrCannotUnlinkLastProvider = errors.New("cannot unlink last authentication method")
)

// SocialLogin autentica usuário via Firebase (cria se não existir)
func (s *Service) SocialLogin(ctx context.Context, req *model.SocialLoginRequest, ipAddress, userAgent string) (*model.LoginResponse, error) {
	if s.firebaseClient == nil {
		return nil, ErrFirebaseNotConfigured
	}

	// Validar ID token com Firebase
	token, err := s.firebaseClient.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to verify firebase token")
		return nil, ErrInvalidIDToken
	}

	providerUserID := token.GetUID()
	email := token.GetEmail()
	name := token.GetName()
	picture := token.GetPicture()

	// Verificar se provider já existe
	provider, err := s.providerRepo.GetByProviderAndProviderUserID(ctx, req.Provider, providerUserID)
	if err != nil && !errors.Is(err, repository.ErrProviderNotFound) {
		return nil, fmt.Errorf("failed to check provider: %w", err)
	}

	var user *model.User

	if provider != nil {
		// Provider já existe, fazer login
		user, err = s.userRepo.GetByIDIncludingInactive(ctx, provider.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		if !user.IsActive {
			// Reativar conta: identidade verificada pelo Firebase
			if err := s.userRepo.Reactivate(ctx, user.ID); err != nil {
				return nil, fmt.Errorf("failed to reactivate user: %w", err)
			}
			user.IsActive = true
			s.logger.Info().Str("user_id", user.ID.String()).Msg("user reactivated via social login")
		}

		// Atualizar avatar se o usuário ainda não tiver um
		if user.AvatarURL == nil && picture != "" {
			if err := s.userRepo.UpdateAvatarURL(ctx, user.ID, picture); err == nil {
				user.AvatarURL = &picture
			}
		}

		// Atualizar last login do provider
		_ = s.providerRepo.UpdateLastLogin(ctx, provider.ID)
	} else {
		// Provider não existe, criar novo usuário
		firstName, lastName := splitName(name)

		user = &model.User{
			Email:             email,
			PasswordHash:      "", // Sem senha para login social
			FirstName:         firstName,
			LastName:          lastName,
			PreferredCurrency: "BRL",
			PreferredLanguage: "pt-BR",
			Timezone:          "America/Sao_Paulo",
			Role:              model.UserRoleUser,
			EmailVerified:     true, // Firebase já verificou
			IsActive:          true,
		}

		if picture != "" {
			user.AvatarURL = &picture
		}

		// Verificar se já existe usuário com esse email
		existingUser, err := s.userRepo.GetByEmail(ctx, email)
		if err == nil {
			// Usuário já existe, vincular provider
			user = existingUser
			// Reativar se a conta estiver inativa
			if !user.IsActive {
				if err := s.userRepo.Reactivate(ctx, user.ID); err != nil {
					return nil, fmt.Errorf("failed to reactivate user: %w", err)
				}
				user.IsActive = true
				s.logger.Info().Str("user_id", user.ID.String()).Msg("user reactivated via social login (email match)")
			}
			// Atualizar avatar se ainda não tiver
			if user.AvatarURL == nil && picture != "" {
				if err := s.userRepo.UpdateAvatarURL(ctx, user.ID, picture); err == nil {
					user.AvatarURL = &picture
				}
			}
		} else if !errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("failed to check existing user: %w", err)
		} else {
			// Criar novo usuário
			if err := s.userRepo.Create(ctx, user); err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}

			// Criar settings padrão
			_, err = s.userRepo.CreateDefaultSettings(ctx, user.ID)
			if err != nil {
				s.logger.Warn().Err(err).Msg("failed to create default settings")
			}
		}

		// Criar provider vinculado
		provider = &model.OAuthProvider{
			UserID:            user.ID,
			Provider:          req.Provider,
			ProviderID:        providerUserID,
			ProviderEmail:     &email,
			ProviderName:      name,
			ProviderAvatarURL: &picture,
			IsPrimary:         true,
			LastLoginAt:       time.Now(),
		}

		if err := s.providerRepo.Create(ctx, provider); err != nil {
			return nil, fmt.Errorf("failed to create provider: %w", err)
		}

		s.logger.Info().
			Str("user_id", user.ID.String()).
			Str("provider", req.Provider).
			Msg("new user registered via social login")
	}

	// Atualizar last login do usuário
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Gerar tokens JWT
	accessToken, expiresAt, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshTokenHash, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Criar sessão
	session := &model.UserSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		DeviceInfo:       &req.DeviceInfo,
		IPAddress:        &ipAddress,
		UserAgent:        &userAgent,
		ExpiresAt:        time.Now().Add(time.Duration(s.config.Auth.RefreshTokenDuration) * time.Hour),
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("provider", req.Provider).
		Msg("user logged in via social")

	return &model.LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// LinkProvider vincula um provider OAuth a uma conta existente
func (s *Service) LinkProvider(ctx context.Context, userID uuid.UUID, req *model.LinkProviderRequest) error {
	if s.firebaseClient == nil {
		return ErrFirebaseNotConfigured
	}

	// Validar ID token com Firebase
	token, err := s.firebaseClient.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to verify firebase token")
		return ErrInvalidIDToken
	}

	providerUserID := token.GetUID()
	email := token.GetEmail()
	name := token.GetName()
	picture := token.GetPicture()

	// Verificar se provider já está vinculado a este usuário
	_, err = s.providerRepo.GetByUserIDAndProvider(ctx, userID, req.Provider)
	if err == nil {
		return repository.ErrProviderAlreadyLinked
	}
	if !errors.Is(err, repository.ErrProviderNotFound) {
		return fmt.Errorf("failed to check provider: %w", err)
	}

	// Verificar se Firebase UID já está vinculado a outra conta
	existingProvider, err := s.providerRepo.GetByProviderAndProviderUserID(ctx, req.Provider, providerUserID)
	if err != nil && !errors.Is(err, repository.ErrProviderNotFound) {
		return fmt.Errorf("failed to check provider user: %w", err)
	}
	if existingProvider != nil {
		return repository.ErrProviderUserExists
	}

	// Criar registro de provider vinculado
	provider := &model.OAuthProvider{
		UserID:            userID,
		Provider:          req.Provider,
		ProviderID:        providerUserID,
		ProviderEmail:     &email,
		ProviderName:      name,
		ProviderAvatarURL: &picture,
		IsPrimary:         false,
		LastLoginAt:       time.Now(),
	}

	if err := s.providerRepo.Create(ctx, provider); err != nil {
		return fmt.Errorf("failed to link provider: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Str("provider", req.Provider).
		Msg("provider linked to account")

	return nil
}

// UnlinkProvider remove vinculo de provider
func (s *Service) UnlinkProvider(ctx context.Context, userID uuid.UUID, provider string) error {
	// Verificar se provider existe
	_, err := s.providerRepo.GetByUserIDAndProvider(ctx, userID, provider)
	if err != nil {
		if errors.Is(err, repository.ErrProviderNotFound) {
			return err
		}
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Verificar se usuário tem senha
	hasPassword, err := s.providerRepo.CheckUserHasPassword(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check password: %w", err)
	}

	// Contar quantos providers o usuário tem
	providerCount, err := s.providerRepo.CountProvidersByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to count providers: %w", err)
	}

	// Não permitir desvincular se for o único método de autenticação
	if !hasPassword && providerCount <= 1 {
		return ErrCannotUnlinkLastProvider
	}

	// Deletar provider
	if err := s.providerRepo.Delete(ctx, userID, provider); err != nil {
		return fmt.Errorf("failed to unlink provider: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Str("provider", provider).
		Msg("provider unlinked from account")

	return nil
}

// GetLinkedProviders lista providers vinculados
func (s *Service) GetLinkedProviders(ctx context.Context, userID uuid.UUID) (*model.ListProvidersResponse, error) {
	// Buscar todos os providers do usuário
	providers, err := s.providerRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	// Verificar se tem senha
	hasPassword, err := s.providerRepo.CheckUserHasPassword(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check password: %w", err)
	}

	// Converter para DTO
	linkedProviders := make([]model.LinkedProviderResponse, 0, len(providers))
	for _, p := range providers {
		email := ""
		if p.ProviderEmail != nil {
			email = *p.ProviderEmail
		}
		avatarURL := ""
		if p.ProviderAvatarURL != nil {
			avatarURL = *p.ProviderAvatarURL
		}

		linkedProviders = append(linkedProviders, model.LinkedProviderResponse{
			Provider:  p.Provider,
			Email:     email,
			Name:      p.ProviderName,
			AvatarURL: avatarURL,
			IsPrimary: p.IsPrimary,
			LinkedAt:  p.CreatedAt,
		})
	}

	return &model.ListProvidersResponse{
		Providers:   linkedProviders,
		HasPassword: hasPassword,
	}, nil
}

// Helper function to split full name into first and last name
func splitName(fullName string) (string, string) {
	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) == 0 {
		return "User", "User"
	}
	if len(parts) == 1 {
		return parts[0], parts[0]
	}
	return parts[0], strings.Join(parts[1:], " ")
}
