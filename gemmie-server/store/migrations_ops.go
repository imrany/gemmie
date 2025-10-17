package store

import (
	"embed"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// RunMigrations executes database migrations
func RunMigrations() error {
	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		return err
	}

	sourceDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	// Get current version
	currentVersion, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		slog.Warn("Could not determine current migration version", "error", err)
	}

	if dirty {
		slog.Warn("Database is in dirty state, attempting to force version", "version", currentVersion)
		// Try to force the version to recover from dirty state
		if err := m.Force(int(currentVersion)); err != nil {
			slog.Error("Failed to force version", "error", err)
		}
	}

	// Run all up migrations automatically
	slog.Info("Running database migrations...")
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			slog.Info("Database schema is up to date")
			return nil
		}
		return err
	}

	// Get new version after migration
	newVersion, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return err
	}

	slog.Info("Database migrations completed successfully", 
		"previous_version", currentVersion, 
		"current_version", newVersion)
	return nil
}

// MigrateUp runs all pending migrations
func MigrateUp() error {
	return RunMigrations()
}

// MigrateDown rolls back the last migration (use with caution)
func MigrateDown() error {
	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		return err
	}

	sourceDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	slog.Info("Rolling back last migration...")
	return m.Steps(-1)
}

// MigrateSteps migrates up or down by the specified number of steps
// Positive number = migrate up, Negative number = migrate down
func MigrateSteps(steps int) error {
	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		return err
	}

	sourceDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	slog.Info("Migrating by steps", "steps", steps)
	return m.Steps(steps)
}

// MigrateTo migrates to a specific version
func MigrateTo(version uint) error {
	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		return err
	}

	sourceDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	slog.Info("Migrating to version", "version", version)
	return m.Migrate(version)
}

// GetMigrationVersion returns the current migration version and dirty state
func GetMigrationVersion() (uint, bool, error) {
	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		return 0, false, err
	}

	sourceDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return 0, false, err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return 0, false, err
	}

	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	return version, dirty, err
}

// ForceMigrationVersion forces the migration version (use when in dirty state)
func ForceMigrationVersion(version int) error {
	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		return err
	}

	sourceDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	slog.Warn("Forcing migration version", "version", version)
	return m.Force(version)
}