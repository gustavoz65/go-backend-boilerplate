package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/example/go-backend-boilerplate/internal/model"
	"github.com/example/go-backend-boilerplate/internal/repository"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	logger           *zerolog.Logger
}

func NewNotificationService(
	notificationRepo *repository.NotificationRepository,
	logger *zerolog.Logger,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		logger:           logger,
	}
}

func (s *NotificationService) Create(ctx context.Context, notification *model.Notification) error {
	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.Info().
		Str("user_id", notification.UserID.String()).
		Str("notification_id", notification.ID.String()).
		Str("type", string(notification.Type)).
		Msg("notification created")

	return nil
}

func (s *NotificationService) CreateBulk(ctx context.Context, userIDs []uuid.UUID, notificationType model.NotificationType, title, message string, data interface{}) error {
	var dataJSON *string
	if data != nil {
		bytes, err := json.Marshal(data)
		if err == nil {
			str := string(bytes)
			dataJSON = &str
		}
	}

	for _, userID := range userIDs {
		notification := &model.Notification{
			UserID:  userID,
			Type:    notificationType,
			Title:   title,
			Message: message,
			Data:    dataJSON,
		}

		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			s.logger.Error().Err(err).
				Str("user_id", userID.String()).
				Msg("failed to create notification for user")
		}
	}

	return nil
}

func (s *NotificationService) GetByID(ctx context.Context, userID, notificationID uuid.UUID) (*model.Notification, error) {
	notification, err := s.notificationRepo.GetByIDAndUser(ctx, notificationID, userID)
	if err != nil {
		return nil, err
	}
	return notification, nil
}

func (s *NotificationService) GetAll(ctx context.Context, userID uuid.UUID, page, pageSize int) (*model.PaginatedResponse[*model.Notification], error) {
	pagination := repository.PaginationParams{
		Page:     page,
		PageSize: pageSize,
		SortBy:   "created_at",
		SortDir:  "DESC",
	}
	pagination.Validate([]string{"created_at", "is_read"})

	notifications, total, err := s.notificationRepo.GetAllByUser(ctx, userID, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &model.PaginatedResponse[*model.Notification]{
		Data:       notifications,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
		HasMore:    page < totalPages,
	}, nil
}

func (s *NotificationService) GetUnread(ctx context.Context, userID uuid.UUID) ([]*model.Notification, error) {
	notifications, err := s.notificationRepo.GetUnreadByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread notifications: %w", err)
	}
	return notifications, nil
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.notificationRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	if err := s.notificationRepo.MarkAsRead(ctx, notificationID, userID); err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if err := s.notificationRepo.MarkAllAsRead(ctx, userID); err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Msg("all notifications marked as read")

	return nil
}

func (s *NotificationService) Delete(ctx context.Context, userID, notificationID uuid.UUID) error {
	if err := s.notificationRepo.Delete(ctx, notificationID, userID); err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Str("notification_id", notificationID.String()).
		Msg("notification deleted")

	return nil
}

func (s *NotificationService) ProcessScheduledNotifications(ctx context.Context) error {
	notifications, err := s.notificationRepo.GetPendingScheduled(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending scheduled notifications: %w", err)
	}

	for _, notification := range notifications {
		if err := s.notificationRepo.MarkAsSent(ctx, notification.ID); err != nil {
			s.logger.Error().Err(err).
				Str("notification_id", notification.ID.String()).
				Msg("failed to mark notification as sent")
		}
	}

	return nil
}

func (s *NotificationService) CleanupOldNotifications(ctx context.Context, days int) (int64, error) {
	deleted, err := s.notificationRepo.DeleteOldNotifications(ctx, days)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old notifications: %w", err)
	}

	if deleted > 0 {
		s.logger.Info().
			Int64("deleted_count", deleted).
			Int("days_old", days).
			Msg("cleaned up old notifications")
	}

	return deleted, nil
}
