package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/example/go-backend-boilerplate/internal/database"
	"github.com/example/go-backend-boilerplate/internal/model"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrSessionNotFound   = errors.New("session not found")
	ErrSessionExpired    = errors.New("session expired")
	ErrSessionRevoked    = errors.New("session revoked")
)

type UserRepository struct {
	*BaseRepository
}

func NewUserRepository(db *database.Database, logger *zerolog.Logger) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository(db, logger),
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (
			id, email, password_hash, first_name, last_name, phone, avatar_url,
			preferred_currency, preferred_language, timezone, role,
			email_verified, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.ExecContext(ctx, query,
		user.ID.String(),
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		NullString(user.Phone),
		NullString(user.AvatarURL),
		user.PreferredCurrency,
		user.PreferredLanguage,
		user.Timezone,
		user.Role,
		user.EmailVerified,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, avatar_url,
			preferred_currency, preferred_language, timezone, role,
			email_verified, email_verified_at, last_login_at, is_active,
			created_at, updated_at
		FROM users
		WHERE id = ? AND is_active = TRUE
	`

	user := &model.User{}
	var phone, avatarURL sql.NullString
	var emailVerifiedAt, lastLoginAt sql.NullTime

	err := r.QueryRowContext(ctx, query, id.String()).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&phone,
		&avatarURL,
		&user.PreferredCurrency,
		&user.PreferredLanguage,
		&user.Timezone,
		&user.Role,
		&user.EmailVerified,
		&emailVerifiedAt,
		&lastLoginAt,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	user.Phone = StringPtr(phone)
	user.AvatarURL = StringPtr(avatarURL)
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, avatar_url,
			preferred_currency, preferred_language, timezone, role,
			email_verified, email_verified_at, last_login_at, is_active,
			created_at, updated_at
		FROM users
		WHERE email = ?
	`

	user := &model.User{}
	var phone, avatarURL sql.NullString
	var emailVerifiedAt, lastLoginAt sql.NullTime

	err := r.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&phone,
		&avatarURL,
		&user.PreferredCurrency,
		&user.PreferredLanguage,
		&user.Timezone,
		&user.Role,
		&user.EmailVerified,
		&emailVerifiedAt,
		&lastLoginAt,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user.Phone = StringPtr(phone)
	user.AvatarURL = StringPtr(avatarURL)
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users SET
			first_name = ?, last_name = ?, phone = ?, avatar_url = ?,
			preferred_currency = ?, preferred_language = ?, timezone = ?,
			email_verified = ?, email_verified_at = ?, last_login_at = ?,
			is_active = ?, updated_at = ?
		WHERE id = ?
	`

	var emailVerifiedAt, lastLoginAt sql.NullTime
	if user.EmailVerifiedAt != nil {
		emailVerifiedAt = sql.NullTime{Time: *user.EmailVerifiedAt, Valid: true}
	}
	if user.LastLoginAt != nil {
		lastLoginAt = sql.NullTime{Time: *user.LastLoginAt, Valid: true}
	}

	result, err := r.ExecContext(ctx, query,
		user.FirstName,
		user.LastName,
		NullString(user.Phone),
		NullString(user.AvatarURL),
		user.PreferredCurrency,
		user.PreferredLanguage,
		user.Timezone,
		user.EmailVerified,
		emailVerifiedAt,
		lastLoginAt,
		user.IsActive,
		user.UpdatedAt,
		user.ID.String(),
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`

	result, err := r.ExecContext(ctx, query, passwordHash, time.Now(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateAvatarURL updates the avatar URL for a user.
func (r *UserRepository) UpdateAvatarURL(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	query := `UPDATE users SET avatar_url = ?, updated_at = ? WHERE id = ?`

	_, err := r.ExecContext(ctx, query, avatarURL, time.Now(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to update avatar URL: %w", err)
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_login_at = ? WHERE id = ?`

	_, err := r.ExecContext(ctx, query, time.Now(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// VerifyEmail marks user email as verified
func (r *UserRepository) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET email_verified = TRUE, email_verified_at = ? WHERE id = ?`

	_, err := r.ExecContext(ctx, query, time.Now(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	return nil
}

// Deactivate soft deletes a user
func (r *UserRepository) Deactivate(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET is_active = FALSE, updated_at = ? WHERE id = ?`

	_, err := r.ExecContext(ctx, query, time.Now(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

// Reactivate reativates a soft-deleted user
func (r *UserRepository) Reactivate(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET is_active = TRUE, updated_at = ? WHERE id = ?`

	_, err := r.ExecContext(ctx, query, time.Now(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to reactivate user: %w", err)
	}

	return nil
}

// GetByIDIncludingInactive retrieves a user by ID regardless of active status
func (r *UserRepository) GetByIDIncludingInactive(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, avatar_url,
			preferred_currency, preferred_language, timezone, role,
			email_verified, email_verified_at, last_login_at, is_active,
			created_at, updated_at
		FROM users
		WHERE id = ?
	`

	user := &model.User{}
	var phone, avatarURL sql.NullString
	var emailVerifiedAt, lastLoginAt sql.NullTime

	err := r.QueryRowContext(ctx, query, id.String()).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&phone,
		&avatarURL,
		&user.PreferredCurrency,
		&user.PreferredLanguage,
		&user.Timezone,
		&user.Role,
		&user.EmailVerified,
		&emailVerifiedAt,
		&lastLoginAt,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	user.Phone = StringPtr(phone)
	user.AvatarURL = StringPtr(avatarURL)
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// Session Methods

// CreateSession creates a new user session
func (r *UserRepository) CreateSession(ctx context.Context, session *model.UserSession) error {
	session.ID = uuid.New()
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()

	query := `
		INSERT INTO user_sessions (
			id, user_id, refresh_token_hash, device_info, ip_address,
			user_agent, expires_at, is_revoked, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.ExecContext(ctx, query,
		session.ID.String(),
		session.UserID.String(),
		session.RefreshTokenHash,
		NullString(session.DeviceInfo),
		NullString(session.IPAddress),
		NullString(session.UserAgent),
		session.ExpiresAt,
		session.IsRevoked,
		session.CreatedAt,
		session.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByToken retrieves a session by refresh token hash
func (r *UserRepository) GetSessionByToken(ctx context.Context, tokenHash string) (*model.UserSession, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, device_info, ip_address,
			user_agent, expires_at, is_revoked, created_at, updated_at
		FROM user_sessions
		WHERE refresh_token_hash = ?
	`

	session := &model.UserSession{}
	var deviceInfo, ipAddress, userAgent sql.NullString

	err := r.QueryRowContext(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&deviceInfo,
		&ipAddress,
		&userAgent,
		&session.ExpiresAt,
		&session.IsRevoked,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	session.DeviceInfo = StringPtr(deviceInfo)
	session.IPAddress = StringPtr(ipAddress)
	session.UserAgent = StringPtr(userAgent)

	if session.IsRevoked {
		return nil, ErrSessionRevoked
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	return session, nil
}

// RevokeSession revokes a session
func (r *UserRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	query := `UPDATE user_sessions SET is_revoked = TRUE, updated_at = ? WHERE id = ?`

	_, err := r.ExecContext(ctx, query, time.Now(), sessionID.String())
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *UserRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE user_sessions SET is_revoked = TRUE, updated_at = ? WHERE user_id = ?`

	_, err := r.ExecContext(ctx, query, time.Now(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to revoke all sessions: %w", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (r *UserRepository) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	query := `DELETE FROM user_sessions WHERE expires_at < ? OR is_revoked = TRUE`

	result, err := r.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup sessions: %w", err)
	}

	return result.RowsAffected()
}

// User Settings Methods

// GetSettings retrieves user settings
func (r *UserRepository) GetSettings(ctx context.Context, userID uuid.UUID) (*model.UserSettings, error) {
	query := `
		SELECT id, user_id, notification_email, notification_push, notification_sms,
			weekly_summary, product_updates, theme, created_at, updated_at
		FROM user_settings
		WHERE user_id = ?
	`

	settings := &model.UserSettings{}
	err := r.QueryRowContext(ctx, query, userID.String()).Scan(
		&settings.ID,
		&settings.UserID,
		&settings.NotificationEmail,
		&settings.NotificationPush,
		&settings.NotificationSMS,
		&settings.WeeklySummary,
		&settings.ProductUpdates,
		&settings.Theme,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create default settings
			return r.CreateDefaultSettings(ctx, userID)
		}
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	return settings, nil
}

// CreateDefaultSettings creates default settings for a user
func (r *UserRepository) CreateDefaultSettings(ctx context.Context, userID uuid.UUID) (*model.UserSettings, error) {
	settings := &model.UserSettings{
		ID:                uuid.New(),
		UserID:            userID,
		NotificationEmail: true,
		NotificationPush:  true,
		NotificationSMS:   false,
		WeeklySummary:     true,
		ProductUpdates:    true,
		Theme:             "system",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	query := `
		INSERT INTO user_settings (
			id, user_id, notification_email, notification_push, notification_sms,
			weekly_summary, product_updates, theme, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.ExecContext(ctx, query,
		settings.ID.String(),
		settings.UserID.String(),
		settings.NotificationEmail,
		settings.NotificationPush,
		settings.NotificationSMS,
		settings.WeeklySummary,
		settings.ProductUpdates,
		settings.Theme,
		settings.CreatedAt,
		settings.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user settings: %w", err)
	}

	return settings, nil
}

// UpdateSettings updates user settings
func (r *UserRepository) UpdateSettings(ctx context.Context, settings *model.UserSettings) error {
	settings.UpdatedAt = time.Now()

	query := `
		UPDATE user_settings SET
			notification_email = ?, notification_push = ?, notification_sms = ?,
			weekly_summary = ?, product_updates = ?, theme = ?, updated_at = ?
		WHERE user_id = ?
	`

	_, err := r.ExecContext(ctx, query,
		settings.NotificationEmail,
		settings.NotificationPush,
		settings.NotificationSMS,
		settings.WeeklySummary,
		settings.ProductUpdates,
		settings.Theme,
		settings.UpdatedAt,
		settings.UserID.String(),
	)

	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	return nil
}

// Helper function to check for duplicate key error
func isDuplicateKeyError(err error) bool {
	return err != nil && (contains(err.Error(), "Duplicate entry") || contains(err.Error(), "duplicate key"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
