package repository

import (
	"context"

	"github.com/google/uuid"

	models "github.com/example/go-backend-boilerplate/internal/model"
)

type Filter struct {
	Limit  int
	Offset int
}

type AuditLogRepository interface {
	Create(ctx context.Context, auditLog *models.AuditLog) error
	FindByUser(ctx context.Context, userID string, filters Filter) ([]*models.AuditLog, error)
	FindByEntity(ctx context.Context, entityType string, entityID uuid.UUID, limit int) ([]*models.AuditLog, error)
	CountByUser(ctx context.Context, userID string) (int, error)
}
