package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/example/go-backend-boilerplate/internal/model"
	"github.com/example/go-backend-boilerplate/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
	logger   *zerolog.Logger
}

func NewUserService(userRepo *repository.UserRepository, logger *zerolog.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByEmail retrieves a user by email
func (s *UserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update updates a user profile
func (s *UserService) Update(ctx context.Context, userID uuid.UUID, req *model.UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.PreferredCurrency != nil {
		user.PreferredCurrency = *req.PreferredCurrency
	}
	if req.PreferredLanguage != nil {
		user.PreferredLanguage = *req.PreferredLanguage
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Msg("user profile updated")

	return user, nil
}

// CompleteOnboarding marks the user's onboarding as completed
func (s *UserService) CompleteOnboarding(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.OnboardingCompleted = true

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to complete onboarding: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Msg("user onboarding completed")

	return user, nil
}

// Deactivate deactivates a user account
func (s *UserService) Deactivate(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.Deactivate(ctx, userID); err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Msg("user account deactivated")

	return nil
}

// GetSettings retrieves user settings
func (s *UserService) GetSettings(ctx context.Context, userID uuid.UUID) (*model.UserSettings, error) {
	settings, err := s.userRepo.GetSettings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}
	return settings, nil
}

// UpdateSettings updates user settings
func (s *UserService) UpdateSettings(ctx context.Context, userID uuid.UUID, req *model.UpdateUserSettingsRequest) (*model.UserSettings, error) {
	settings, err := s.userRepo.GetSettings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	// Apply updates
	if req.NotificationEmail != nil {
		settings.NotificationEmail = *req.NotificationEmail
	}
	if req.NotificationPush != nil {
		settings.NotificationPush = *req.NotificationPush
	}
	if req.NotificationSMS != nil {
		settings.NotificationSMS = *req.NotificationSMS
	}
	if req.WeeklySummary != nil {
		settings.WeeklySummary = *req.WeeklySummary
	}
	if req.ProductUpdates != nil {
		settings.ProductUpdates = *req.ProductUpdates
	}
	if req.Theme != nil {
		settings.Theme = *req.Theme
	}

	if err := s.userRepo.UpdateSettings(ctx, settings); err != nil {
		return nil, fmt.Errorf("failed to update user settings: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Msg("user settings updated")

	return settings, nil
}
