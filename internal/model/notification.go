package model

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeSystem NotificationType = "system"
	NotificationTypeInfo   NotificationType = "info"
	NotificationTypeAlert  NotificationType = "alert"
)

type NotificationChannel string

const (
	NotificationChannelInApp NotificationChannel = "in_app"
	NotificationChannelEmail NotificationChannel = "email"
	NotificationChannelPush  NotificationChannel = "push"
	NotificationChannelSMS   NotificationChannel = "sms"
)

type Notification struct {
	ID           uuid.UUID        `json:"id" db:"id"`
	UserID       uuid.UUID        `json:"user_id" db:"user_id"`
	Type         NotificationType `json:"type" db:"type"`
	Title        string           `json:"title" db:"title"`
	Message      string           `json:"message" db:"message"`
	Data         *string          `json:"data,omitempty" db:"data"`
	IsRead       bool             `json:"is_read" db:"is_read"`
	ReadAt       *time.Time       `json:"read_at,omitempty" db:"read_at"`
	SentVia      []string         `json:"sent_via" db:"sent_via"`
	ScheduledFor *time.Time       `json:"scheduled_for,omitempty" db:"scheduled_for"`
	SentAt       *time.Time       `json:"sent_at,omitempty" db:"sent_at"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
}

func (n *Notification) MarkAsRead() {
	n.IsRead = true
	now := time.Now()
	n.ReadAt = &now
}

func (n *Notification) IsPending() bool {
	return n.SentAt == nil
}

func (n *Notification) IsScheduled() bool {
	return n.ScheduledFor != nil && n.SentAt == nil
}

func (n *Notification) ShouldSendNow() bool {
	if n.SentAt != nil {
		return false
	}
	if n.ScheduledFor == nil {
		return true
	}
	return time.Now().After(*n.ScheduledFor)
}

// Optional template helpers for applications built from this boilerplate.
type NotificationTemplate struct {
	Type    NotificationType
	Title   string
	Message string
}

var NotificationTemplates = map[NotificationType]NotificationTemplate{
	NotificationTypeSystem: {Type: NotificationTypeSystem, Title: "System", Message: "%s"},
	NotificationTypeInfo:   {Type: NotificationTypeInfo, Title: "Info", Message: "%s"},
	NotificationTypeAlert:  {Type: NotificationTypeAlert, Title: "Alert", Message: "%s"},
}
