package model

import (
	"time"

	"github.com/google/uuid"
)

type OAuthProvider struct {
	ID                uuid.UUID `json:"id" db:"id"`
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	Provider          string    `json:"provider" db:"provider"`
	ProviderID        string    `json:"provider_id" db:"provider_id"`
	ProviderEmail     *string   `json:"provider_email,omitempty" db:"provider_email"`
	ProviderName      string    `json:"provider_name" db:"provider_name"`
	ProviderAvatarURL *string   `json:"provider_avatar_url,omitempty" db:"provider_avatar_url"`
	IsPrimary         bool      `json:"is_primary" db:"is_primary"`
	Metadata          *string   `json:"metadata,omitempty" db:"metadata"`
	LastLoginAt       time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

const (
	ProviderGoogle   = "google"
	ProviderFacebook = "facebook"
	ProviderGithub   = "github"
	ProviderTwitter  = "twitter"
)

func (p *OAuthProvider) IsGoogle() bool {
	return p.Provider == ProviderGoogle
}

func (p *OAuthProvider) IsFacebook() bool {
	return false
	// futuramente será implementado
}

func (p *OAuthProvider) IsGithub() bool {
	return false
	//futuramente será implementado
}
