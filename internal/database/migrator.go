package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net"
	"strconv"

	"github.com/rs/zerolog"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func Migrate(ctx context.Context, logger *zerolog.Logger, cfg *config.Config) error {
	hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))

	// MySQL DSN format
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&multiStatements=true",
		cfg.Database.User,
		cfg.Database.Password,
		hostPort,
		cfg.Database.Name,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	driver, err := mysql.WithInstance(db, &mysql.Config{
		MigrationsTable: "schema_migrations",
		DatabaseName:    cfg.Database.Name,
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Cria o source driver usando o sistema de arquivos embutido
	sourceDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "mysql", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current migration version: %d", err)
	}

	if dirty {
		logger.Error().Int("version", int(version)).Msg("Database is dirty")
		return fmt.Errorf("database is in a dirty state at version %d", version)
	}

	logger.Info().Msg("Applying migrations...")
	err = m.Up()

	if err != nil {
		if err == migrate.ErrNoChange {
			logger.Info().Msg("No pending migrations")
			return nil
		}
		return fmt.Errorf("failed to apply migrations:%d", err)
	}

	newVersion, _, _ := m.Version()
	logger.Info().
		Int("from", int(version)).
		Int("to", int(newVersion)).
		Msg("Migrations applied successfully")

	return nil
}
