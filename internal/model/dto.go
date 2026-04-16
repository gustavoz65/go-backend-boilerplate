package model

import "time"

// Auth DTOs

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
	Phone     string `json:"phone,omitempty" validate:"omitempty,max=20"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
}

type SetPasswordRequest struct {
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

type SocialLoginRequest struct {
	Provider   string `json:"provider" validate:"required,oneof=google facebook github"`
	IDToken    string `json:"id_token" validate:"required"`
	DeviceInfo string `json:"device_info,omitempty" validate:"omitempty,max=500"`
}

type LinkProviderRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google facebook github"`
	IDToken  string `json:"id_token" validate:"required"`
}

type LinkedProviderResponse struct {
	Provider  string    `json:"provider"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	IsPrimary bool      `json:"is_primary"`
	LinkedAt  time.Time `json:"linked_at"`
}

type ListProvidersResponse struct {
	Providers   []LinkedProviderResponse `json:"providers"`
	HasPassword bool                     `json:"has_password"`
}

// User DTOs

type UpdateUserRequest struct {
	FirstName         *string `json:"first_name,omitempty" validate:"omitempty,min=2,max=100"`
	LastName          *string `json:"last_name,omitempty" validate:"omitempty,min=2,max=100"`
	Phone             *string `json:"phone,omitempty" validate:"omitempty,max=20"`
	PreferredCurrency *string `json:"preferred_currency,omitempty" validate:"omitempty,len=3"`
	PreferredLanguage *string `json:"preferred_language,omitempty" validate:"omitempty,max=5"`
	Timezone          *string `json:"timezone,omitempty" validate:"omitempty,max=50"`
}

type UpdateUserSettingsRequest struct {
	NotificationEmail *bool   `json:"notification_email,omitempty"`
	NotificationPush  *bool   `json:"notification_push,omitempty"`
	NotificationSMS   *bool   `json:"notification_sms,omitempty"`
	WeeklySummary     *bool   `json:"weekly_summary,omitempty"`
	ProductUpdates    *bool   `json:"product_updates,omitempty"`
	Theme             *string `json:"theme,omitempty" validate:"omitempty,oneof=light dark system"`
}

// Generic response DTOs

type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasMore    bool  `json:"has_more"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}
