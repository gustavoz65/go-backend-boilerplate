package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/database"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/model"
)

type AuditLogRepositoryImpl struct {
	*BaseRepository
}

func NewAuditLogRepository(db *database.Database, logger *zerolog.Logger) *AuditLogRepositoryImpl {
	return &AuditLogRepositoryImpl{
		BaseRepository: NewBaseRepository(db, logger),
	}
}

// Create creates a new audit log entry
func (r *AuditLogRepositoryImpl) Create(ctx context.Context, auditLog *model.AuditLog) error {
	auditLog.ID = uuid.New()
	auditLog.CreatedAt = time.Now()

	query := `
		INSERT INTO audit_logs (
			id, user_id, action, entity_type, entity_id,
			old_values, new_values, ip_address, user_agent, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var userID, entityID, ipAddress, userAgent sql.NullString

	if auditLog.UserID != nil {
		userID = sql.NullString{String: auditLog.UserID.String(), Valid: true}
	}

	if auditLog.EntityID != nil {
		entityID = sql.NullString{String: auditLog.EntityID.String(), Valid: true}
	}

	if auditLog.IPAddress != nil {
		ipAddress = sql.NullString{String: *auditLog.IPAddress, Valid: true}
	}

	if auditLog.UserAgent != nil {
		userAgent = sql.NullString{String: *auditLog.UserAgent, Valid: true}
	}

	_, err := r.ExecContext(ctx, query,
		auditLog.ID.String(),
		userID,
		auditLog.Action,
		auditLog.EntityType,
		entityID,
		auditLog.OldValues,
		auditLog.NewValues,
		ipAddress,
		userAgent,
		auditLog.CreatedAt,
	)

	if err != nil {
		r.logger.Error().Err(err).Msg("failed to create audit log")
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// FindByUser retrieves audit logs for a specific user with filters
func (r *AuditLogRepositoryImpl) FindByUser(ctx context.Context, userID string, filters Filter) ([]*model.AuditLog, error) {
	query := `
		SELECT
			id, user_id, action, entity_type, entity_id,
			old_values, new_values, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.QueryContext(ctx, query, userID, filters.Limit, filters.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		log := &model.AuditLog{}
		var userID, entityID, ipAddress, userAgent sql.NullString
		var oldValues, newValues sql.NullString

		err := rows.Scan(
			&log.ID,
			&userID,
			&log.Action,
			&log.EntityType,
			&entityID,
			&oldValues,
			&newValues,
			&ipAddress,
			&userAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Handle nullable UUID fields
		if userID.Valid {
			uid, _ := uuid.Parse(userID.String)
			log.UserID = &uid
		}

		if entityID.Valid {
			eid, _ := uuid.Parse(entityID.String)
			log.EntityID = &eid
		}

		// Handle nullable string fields
		if ipAddress.Valid {
			log.IPAddress = &ipAddress.String
		}

		if userAgent.Valid {
			log.UserAgent = &userAgent.String
		}

		// Handle JSON fields
		if oldValues.Valid {
			if err := log.OldValues.Scan([]byte(oldValues.String)); err != nil {
				r.logger.Error().Err(err).Msg("failed to parse old_values JSON")
			}
		}

		if newValues.Valid {
			if err := log.NewValues.Scan([]byte(newValues.String)); err != nil {
				r.logger.Error().Err(err).Msg("failed to parse new_values JSON")
			}
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return logs, nil
}

// FindByEntity retrieves audit logs for a specific entity
func (r *AuditLogRepositoryImpl) FindByEntity(ctx context.Context, entityType string, entityID uuid.UUID, limit int) ([]*model.AuditLog, error) {
	query := `
		SELECT
			id, user_id, action, entity_type, entity_id,
			old_values, new_values, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.QueryContext(ctx, query, entityType, entityID.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs by entity: %w", err)
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		log := &model.AuditLog{}
		var userID, entityIDStr, ipAddress, userAgent sql.NullString
		var oldValues, newValues sql.NullString

		err := rows.Scan(
			&log.ID,
			&userID,
			&log.Action,
			&log.EntityType,
			&entityIDStr,
			&oldValues,
			&newValues,
			&ipAddress,
			&userAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Handle nullable fields
		if userID.Valid {
			uid, _ := uuid.Parse(userID.String)
			log.UserID = &uid
		}

		if entityIDStr.Valid {
			eid, _ := uuid.Parse(entityIDStr.String)
			log.EntityID = &eid
		}

		if ipAddress.Valid {
			log.IPAddress = &ipAddress.String
		}

		if userAgent.Valid {
			log.UserAgent = &userAgent.String
		}

		if oldValues.Valid {
			log.OldValues.Scan([]byte(oldValues.String))
		}

		if newValues.Valid {
			log.NewValues.Scan([]byte(newValues.String))
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// CountByUser returns the total count of audit logs for a user
func (r *AuditLogRepositoryImpl) CountByUser(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM audit_logs WHERE user_id = ?`

	var count int
	err := r.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}
