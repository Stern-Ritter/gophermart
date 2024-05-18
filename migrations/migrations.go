package migrations

import (
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed "*.sql"
var Migrations embed.FS

func Migrate(databaseURL string, dialect string, driverName string) error {
	goose.SetBaseFS(Migrations)
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("goose failed to set %s dialect: %w", dialect, err)
	}

	db, err := goose.OpenDBWithDriver(driverName, databaseURL)
	if err != nil {
		return fmt.Errorf("goose failed to open database connection: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose failed to up migrations: %w", err)
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("goose failed to close database connection: %w", err)
	}

	return nil
}
