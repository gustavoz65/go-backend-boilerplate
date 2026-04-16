// Pacote repository fornece implementações da camada de acesso a dados.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/rs/zerolog"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/database"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	db     *database.Database
	logger *zerolog.Logger
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *database.Database, logger *zerolog.Logger) *BaseRepository {
	return &BaseRepository{
		db:     db,
		logger: logger,
	}
}

// GetDB returns the database instance
func (r *BaseRepository) GetDB() *database.Database {
	return r.db
}

// GetLogger returns the logger instance
func (r *BaseRepository) GetLogger() *zerolog.Logger {
	return r.logger
}

// ExecContext executes a query without returning rows
func (r *BaseRepository) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.db.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows
func (r *BaseRepository) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return r.db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns a single row
func (r *BaseRepository) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return r.db.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a new transaction
func (r *BaseRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx)
}

// WithTx executes a function within a transaction
func (r *BaseRepository) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.Error().Err(rbErr).Msg("failed to rollback transaction")
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page     int
	PageSize int
	SortBy   string
	SortDir  string
}

// DefaultPagination returns default pagination params
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Page:     1,
		PageSize: 20,
		SortBy:   "created_at",
		SortDir:  "DESC",
	}
}

// Offset calculates the offset for pagination
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Validate validates and sanitizes pagination params
func (p *PaginationParams) Validate(allowedSortFields []string) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}

	// Validate sort direction
	p.SortDir = strings.ToUpper(p.SortDir)
	if p.SortDir != "ASC" && p.SortDir != "DESC" {
		p.SortDir = "DESC"
	}

	// Validate sort field
	if p.SortBy == "" {
		p.SortBy = "created_at"
	} else {
		valid := false
		for _, f := range allowedSortFields {
			if f == p.SortBy {
				valid = true
				break
			}
		}
		if !valid {
			p.SortBy = "created_at"
		}
	}
}

// BuildOrderClause builds an ORDER BY clause
func (p PaginationParams) BuildOrderClause() string {
	return fmt.Sprintf("ORDER BY %s %s", p.SortBy, p.SortDir)
}

// BuildLimitClause builds a LIMIT clause for pagination
func (p PaginationParams) BuildLimitClause() string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", p.PageSize, p.Offset())
}

// NullString converts a *string to sql.NullString
func NullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// NullInt64 converts a *int64 to sql.NullInt64
func NullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

// NullInt32 converts a *int to sql.NullInt32
func NullInt32(i *int) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: int32(*i), Valid: true}
}

// NullFloat64 converts a *float64 to sql.NullFloat64
func NullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

// StringPtr converts sql.NullString to *string
func StringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// Int64Ptr converts sql.NullInt64 to *int64
func Int64Ptr(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

// IntPtr converts sql.NullInt32 to *int
func IntPtr(ni sql.NullInt32) *int {
	if !ni.Valid {
		return nil
	}
	i := int(ni.Int32)
	return &i
}

// Float64Ptr converts sql.NullFloat64 to *float64
func Float64Ptr(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}
