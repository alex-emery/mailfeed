package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/alex-emery/mailfeed/database/sqlc"
	msql "github.com/alex-emery/mailfeed/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

type Database struct {
	*sqlc.Queries
}

func New(logger *zap.Logger, filepath string) (Database, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return Database{}, fmt.Errorf("failed to open database: %w", err)
	}

	if err := Migrate(logger, db); err != nil {
		return Database{}, fmt.Errorf("failed to migrate database: %w", err)
	}

	queries := sqlc.New(db)
	return Database{Queries: queries}, nil

}

func Migrate(logger *zap.Logger, db *sql.DB) error {
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}

	d, err := iofs.New(msql.Migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "sql", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	m.Log = &migrationLogger{logger: logger}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migration: %w", err)
	}

	return nil
}

type migrationLogger struct {
	logger *zap.Logger
}

func (m *migrationLogger) Printf(format string, v ...interface{}) {
	m.logger.Info(fmt.Sprintf(format, v...))
}

func (m *migrationLogger) Verbose() bool {
	return true
}
