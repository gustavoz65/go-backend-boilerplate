// Pacote database fornece utilitários de conexão e consulta ao banco de dados.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/rs/zerolog"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	DB   *sql.DB
	log  *zerolog.Logger
	Pool *sql.DB
}

type multiTracer struct {
	tracers []any
}

const DatabasePingTimeout = 10

func New(cfg *config.Config, logger *zerolog.Logger) (*Database, error) {
	hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))

	// MySQL DSN format: user:password@tcp(host:port)/dbname?parseTime=true
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		hostPort,
		cfg.Database.Name,
	)

	if cfg.Database.SSLMode != "disable" && cfg.Database.SSLMode != "" {
		dsn += "&tls=" + cfg.Database.SSLMode
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(cfg.Database.ConnMaxIdleTime) * time.Second)

	database := &Database{
		DB:   db,
		log:  logger,
		Pool: db,
	}

	ctx, cancel := context.WithTimeout(context.Background(), DatabasePingTimeout*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().Msg("connected to the MySQL database")

	return database, nil
}

func (db *Database) Close() error {
	db.log.Info().Msg("closing database connection pool")
	return db.DB.Close()
}

// GetDB returns the underlying sql.DB instance
func (db *Database) GetDB() *sql.DB {
	return db.DB
}

// BeginTx starts a new transaction
func (db *Database) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return db.DB.BeginTx(ctx, nil)
}

// ExecContext executes a query without returning rows
func (db *Database) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows
func (db *Database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns a single row
func (db *Database) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

// PrepareContext creates a prepared statement
func (db *Database) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return db.DB.PrepareContext(ctx, query)
}

// Health checks the database connection health
func (db *Database) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return db.DB.PingContext(ctx)
}

// Stats returns database connection pool statistics
func (db *Database) Stats() sql.DBStats {
	return db.DB.Stats()
}
