package schema

import (
	"context"
	_ "embed" // Calls init functions
	"fmt"

	"github.com/ardanlabs/darwin"
	"github.com/jinchi2013/service/busniess/sys/database"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed sql/schema.sql
	schemaDoc string

	//go:embed sql/delete.sql
	deleteDoc string
)

// Migrate attempts to bring the schema for db up to date
// with the migrations defined in the package
func Migrate(ctx context.Context, db *sqlx.DB) error {
	if err := database.StatusCheck(ctx, db); err != nil {
		return fmt.Errorf("status check database: %w", err)
	}

	driver, err := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})

	if err != nil {
		return fmt.Errorf("construct darwin driver: %w", err)
	}

	d := darwin.New(driver, darwin.ParseMigrations(schemaDoc))

	return d.Migrate()
}

func DeleteAll(db *sqlx.DB) error {
	// Tx is an in-progress database transaction.
	// A transaction must end with a call to Commit or Rollback.
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(deleteDoc); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}
