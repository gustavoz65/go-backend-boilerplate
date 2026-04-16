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
	ErrProviderNotFound      = errors.New("oauth provider not found")
	ErrProviderAlreadyLinked = errors.New("provider already linked to this account")
	ErrProviderUserExists    = errors.New("provider user already linked to another account")
)

type OAuthProviderRepository struct {
	*BaseRepository
}

func NewOAuthProviderRepository(db *database.Database, logger *zerolog.Logger) *OAuthProviderRepository {
	return &OAuthProviderRepository{
		BaseRepository: NewBaseRepository(db, logger),
	}
}

// Create cria novo provider vinculado
func (r *OAuthProviderRepository) Create(ctx context.Context, provider *model.OAuthProvider) error {
	provider.ID = uuid.New()
	provider.CreatedAt = time.Now()
	provider.UpdatedAt = time.Now()

	query := `
		INSERT INTO user_oauth_providers (
			id, user_id, provider, provider_user_id, provider_email,
			provider_name, provider_avatar_url, is_primary, metadata,
			last_login_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.ExecContext(ctx, query,
		provider.ID.String(),
		provider.UserID.String(),
		provider.Provider,
		provider.ProviderID,
		NullString(provider.ProviderEmail),
		provider.ProviderName,
		NullString(provider.ProviderAvatarURL),
		provider.IsPrimary,
		NullString(provider.Metadata),
		provider.LastLoginAt,
		provider.CreatedAt,
		provider.UpdatedAt,
	)

	if err != nil {
		if isDuplicateKeyError(err) {
			if contains(err.Error(), "uk_user_provider") {
				return ErrProviderAlreadyLinked
			}
			if contains(err.Error(), "uk_provider_user") {
				return ErrProviderUserExists
			}
		}
		return fmt.Errorf("failed to create oauth provider: %w", err)
	}

	return nil
}

// GetByUserIDAndProvider busca provider específico de um usuário
func (r *OAuthProviderRepository) GetByUserIDAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*model.OAuthProvider, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, provider_email,
			provider_name, provider_avatar_url, is_primary, metadata,
			last_login_at, created_at, updated_at
		FROM user_oauth_providers
		WHERE user_id = ? AND provider = ?
	`

	p := &model.OAuthProvider{}
	var providerEmail, providerAvatarURL, metadata sql.NullString
	var lastLoginAt sql.NullTime

	err := r.QueryRowContext(ctx, query, userID.String(), provider).Scan(
		&p.ID,
		&p.UserID,
		&p.Provider,
		&p.ProviderID,
		&providerEmail,
		&p.ProviderName,
		&providerAvatarURL,
		&p.IsPrimary,
		&metadata,
		&lastLoginAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProviderNotFound
		}
		return nil, fmt.Errorf("failed to get oauth provider: %w", err)
	}

	p.ProviderEmail = StringPtr(providerEmail)
	p.ProviderAvatarURL = StringPtr(providerAvatarURL)
	p.Metadata = StringPtr(metadata)
	if lastLoginAt.Valid {
		p.LastLoginAt = lastLoginAt.Time
	}

	return p, nil
}

// GetByProviderAndProviderUserID busca por provider + uid externo (para descobrir se já existe)
func (r *OAuthProviderRepository) GetByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*model.OAuthProvider, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, provider_email,
			provider_name, provider_avatar_url, is_primary, metadata,
			last_login_at, created_at, updated_at
		FROM user_oauth_providers
		WHERE provider = ? AND provider_user_id = ?
	`

	p := &model.OAuthProvider{}
	var providerEmail, providerAvatarURL, metadata sql.NullString
	var lastLoginAt sql.NullTime

	err := r.QueryRowContext(ctx, query, provider, providerUserID).Scan(
		&p.ID,
		&p.UserID,
		&p.Provider,
		&p.ProviderID,
		&providerEmail,
		&p.ProviderName,
		&providerAvatarURL,
		&p.IsPrimary,
		&metadata,
		&lastLoginAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProviderNotFound
		}
		return nil, fmt.Errorf("failed to get oauth provider: %w", err)
	}

	p.ProviderEmail = StringPtr(providerEmail)
	p.ProviderAvatarURL = StringPtr(providerAvatarURL)
	p.Metadata = StringPtr(metadata)
	if lastLoginAt.Valid {
		p.LastLoginAt = lastLoginAt.Time
	}

	return p, nil
}

// ListByUserID lista todos os providers vinculados a um usuário
func (r *OAuthProviderRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.OAuthProvider, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, provider_email,
			provider_name, provider_avatar_url, is_primary, metadata,
			last_login_at, created_at, updated_at
		FROM user_oauth_providers
		WHERE user_id = ?
		ORDER BY is_primary DESC, created_at ASC
	`

	rows, err := r.QueryContext(ctx, query, userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list oauth providers: %w", err)
	}
	defer rows.Close()

	var providers []*model.OAuthProvider
	for rows.Next() {
		p := &model.OAuthProvider{}
		var providerEmail, providerAvatarURL, metadata sql.NullString
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Provider,
			&p.ProviderID,
			&providerEmail,
			&p.ProviderName,
			&providerAvatarURL,
			&p.IsPrimary,
			&metadata,
			&lastLoginAt,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan oauth provider: %w", err)
		}

		p.ProviderEmail = StringPtr(providerEmail)
		p.ProviderAvatarURL = StringPtr(providerAvatarURL)
		p.Metadata = StringPtr(metadata)
		if lastLoginAt.Valid {
			p.LastLoginAt = lastLoginAt.Time
		}

		providers = append(providers, p)
	}

	return providers, nil
}

// Delete remove vinculo de provider
func (r *OAuthProviderRepository) Delete(ctx context.Context, userID uuid.UUID, provider string) error {
	query := `DELETE FROM user_oauth_providers WHERE user_id = ? AND provider = ?`

	result, err := r.ExecContext(ctx, query, userID.String(), provider)
	if err != nil {
		return fmt.Errorf("failed to delete oauth provider: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrProviderNotFound
	}

	return nil
}

// UpdateLastLogin atualiza timestamp de último login via provider
func (r *OAuthProviderRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE user_oauth_providers SET last_login_at = ? WHERE id = ?`

	_, err := r.ExecContext(ctx, query, time.Now(), id.String())
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// CheckUserHasPassword verifica se usuário tem senha (permite desvincular)
func (r *OAuthProviderRepository) CheckUserHasPassword(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `SELECT password_hash FROM users WHERE id = ?`

	var passwordHash string
	err := r.QueryRowContext(ctx, query, userID.String()).Scan(&passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrUserNotFound
		}
		return false, fmt.Errorf("failed to check user password: %w", err)
	}

	// Se password_hash não é vazio, tem senha
	return passwordHash != "", nil
}

// CountProvidersByUserID conta quantos providers um usuário tem
func (r *OAuthProviderRepository) CountProvidersByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM user_oauth_providers WHERE user_id = ?`

	var count int
	err := r.QueryRowContext(ctx, query, userID.String()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count providers: %w", err)
	}

	return count, nil
}
