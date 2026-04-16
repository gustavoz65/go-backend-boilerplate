package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/example/go-backend-boilerplate/internal/database"
	"github.com/example/go-backend-boilerplate/internal/model"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

type NotificationRepository struct {
	*BaseRepository
}

func NewNotificationRepository(db *database.Database, logger *zerolog.Logger) *NotificationRepository {
	return &NotificationRepository{
		BaseRepository: NewBaseRepository(db, logger),
	}
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	notification.ID = uuid.New()
	notification.CreatedAt = time.Now()

	query := `
		INSERT INTO notifications (
			id, user_id, type, title, message, data, is_read,
			sent_via, scheduled_for, sent_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var dataJSON, sentViaJSON []byte
	if notification.Data != nil {
		dataJSON = []byte(*notification.Data)
	}
	if len(notification.SentVia) > 0 {
		sentViaJSON, _ = json.Marshal(notification.SentVia)
	} else {
		sentViaJSON = []byte(`["in_app"]`)
	}

	var scheduledFor, sentAt sql.NullTime
	if notification.ScheduledFor != nil {
		scheduledFor = sql.NullTime{Time: *notification.ScheduledFor, Valid: true}
	}
	if notification.SentAt != nil {
		sentAt = sql.NullTime{Time: *notification.SentAt, Valid: true}
	}

	_, err := r.ExecContext(ctx, query,
		notification.ID.String(),
		notification.UserID.String(),
		notification.Type,
		notification.Title,
		notification.Message,
		dataJSON,
		notification.IsRead,
		sentViaJSON,
		scheduledFor,
		sentAt,
		notification.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetByID retrieves a notification by ID
func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, is_read, read_at,
			sent_via, scheduled_for, sent_at, created_at
		FROM notifications
		WHERE id = ?
	`

	return r.scanNotification(r.QueryRowContext(ctx, query, id.String()))
}

// GetByIDAndUser retrieves a notification by ID ensuring it belongs to the user
func (r *NotificationRepository) GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*model.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, is_read, read_at,
			sent_via, scheduled_for, sent_at, created_at
		FROM notifications
		WHERE id = ? AND user_id = ?
	`

	return r.scanNotification(r.QueryRowContext(ctx, query, id.String(), userID.String()))
}

// GetAllByUser retrieves all notifications for a user
func (r *NotificationRepository) GetAllByUser(ctx context.Context, userID uuid.UUID, page PaginationParams) ([]*model.Notification, int64, error) {
	page.Validate([]string{"created_at", "is_read"})

	// Count total
	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = ?`
	var total int64
	err := r.QueryRowContext(ctx, countQuery, userID.String()).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Get data
	query := fmt.Sprintf(`
		SELECT id, user_id, type, title, message, data, is_read, read_at,
			sent_via, scheduled_for, sent_at, created_at
		FROM notifications
		WHERE user_id = ?
		%s
		%s
	`, page.BuildOrderClause(), page.BuildLimitClause())

	rows, err := r.QueryContext(ctx, query, userID.String())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer rows.Close()

	notifications, err := r.scanNotifications(rows)
	if err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// GetUnreadByUser retrieves unread notifications for a user
func (r *NotificationRepository) GetUnreadByUser(ctx context.Context, userID uuid.UUID) ([]*model.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, is_read, read_at,
			sent_via, scheduled_for, sent_at, created_at
		FROM notifications
		WHERE user_id = ? AND is_read = FALSE
		ORDER BY created_at DESC
		LIMIT 50
	`

	rows, err := r.QueryContext(ctx, query, userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get unread notifications: %w", err)
	}
	defer rows.Close()

	return r.scanNotifications(rows)
}

// GetUnreadCount returns the count of unread notifications
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = FALSE`

	var count int64
	err := r.QueryRowContext(ctx, query, userID.String()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = TRUE, read_at = ?
		WHERE id = ? AND user_id = ?
	`

	result, err := r.ExecContext(ctx, query, time.Now(), id.String(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = TRUE, read_at = ?
		WHERE user_id = ? AND is_read = FALSE
	`

	_, err := r.ExecContext(ctx, query, time.Now(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// MarkAsSent marks a notification as sent
func (r *NotificationRepository) MarkAsSent(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE notifications SET sent_at = ? WHERE id = ?`

	_, err := r.ExecContext(ctx, query, time.Now(), id.String())
	if err != nil {
		return fmt.Errorf("failed to mark notification as sent: %w", err)
	}

	return nil
}

// GetPendingScheduled retrieves notifications scheduled to be sent
func (r *NotificationRepository) GetPendingScheduled(ctx context.Context) ([]*model.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, is_read, read_at,
			sent_via, scheduled_for, sent_at, created_at
		FROM notifications
		WHERE sent_at IS NULL AND scheduled_for IS NOT NULL AND scheduled_for <= NOW()
		ORDER BY scheduled_for ASC
		LIMIT 100
	`

	rows, err := r.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending scheduled notifications: %w", err)
	}
	defer rows.Close()

	return r.scanNotifications(rows)
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = ? AND user_id = ?`

	result, err := r.ExecContext(ctx, query, id.String(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

// DeleteOldNotifications deletes notifications older than specified days
func (r *NotificationRepository) DeleteOldNotifications(ctx context.Context, days int) (int64, error) {
	query := `DELETE FROM notifications WHERE created_at < DATE_SUB(NOW(), INTERVAL ? DAY)`

	result, err := r.ExecContext(ctx, query, days)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old notifications: %w", err)
	}

	return result.RowsAffected()
}

// Helper functions

func (r *NotificationRepository) scanNotification(row *sql.Row) (*model.Notification, error) {
	notification := &model.Notification{}
	var data sql.NullString
	var readAt, scheduledFor, sentAt sql.NullTime
	var sentVia []byte

	err := row.Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&data,
		&notification.IsRead,
		&readAt,
		&sentVia,
		&scheduledFor,
		&sentAt,
		&notification.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to scan notification: %w", err)
	}

	notification.Data = StringPtr(data)

	if readAt.Valid {
		notification.ReadAt = &readAt.Time
	}
	if scheduledFor.Valid {
		notification.ScheduledFor = &scheduledFor.Time
	}
	if sentAt.Valid {
		notification.SentAt = &sentAt.Time
	}

	if len(sentVia) > 0 {
		_ = json.Unmarshal(sentVia, &notification.SentVia)
	}

	return notification, nil
}

func (r *NotificationRepository) scanNotifications(rows *sql.Rows) ([]*model.Notification, error) {
	var notifications []*model.Notification

	for rows.Next() {
		notification := &model.Notification{}
		var data sql.NullString
		var readAt, scheduledFor, sentAt sql.NullTime
		var sentVia []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&data,
			&notification.IsRead,
			&readAt,
			&sentVia,
			&scheduledFor,
			&sentAt,
			&notification.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		notification.Data = StringPtr(data)

		if readAt.Valid {
			notification.ReadAt = &readAt.Time
		}
		if scheduledFor.Valid {
			notification.ScheduledFor = &scheduledFor.Time
		}
		if sentAt.Valid {
			notification.SentAt = &sentAt.Time
		}

		if len(sentVia) > 0 {
			_ = json.Unmarshal(sentVia, &notification.SentVia)
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}
