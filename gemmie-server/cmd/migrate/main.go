package main

import (
	"fmt"
	"log"

	"github.com/imrany/gemmie/gemmie-server/store"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	// Load .env if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Command line flags
	command := pflag.String("command", "version", "Migration command: up, down, steps, goto, version, force, reset")
	steps := pflag.Int("steps", 1, "Number of steps for migration (used with 'steps' command)")
	target := pflag.Uint("target", 0, "Target version (used with 'goto' command)")
	forceVersion := pflag.Int("version", 0, "Version to force (used with 'force' command)")

	pflag.String("db-host", "localhost", "Database host")
	pflag.Int("db-port", 5432, "Database port")
	pflag.String("db-user", "", "Database user")
	pflag.String("db-password", "", "Database password")
	pflag.String("db-name", "", "Database name")
	pflag.String("db-sslmode", "disable", "Database SSL mode")

	pflag.Parse()

	// Set up viper to read from environment
	viper.AutomaticEnv()

	// Bind flags to viper, but environment variables take precedence
	viper.BindPFlag("DB_HOST", pflag.Lookup("db-host"))
	viper.BindPFlag("DB_PORT", pflag.Lookup("db-port"))
	viper.BindPFlag("DB_USER", pflag.Lookup("db-user"))
	viper.BindPFlag("DB_PASSWORD", pflag.Lookup("db-password"))
	viper.BindPFlag("DB_NAME", pflag.Lookup("db-name"))
	viper.BindPFlag("DB_SSLMODE", pflag.Lookup("db-sslmode"))

	// Get final values
	dbDSN := viper.GetString("DB_DSN")
	if dbDSN == "" {
		dbDSN = viper.GetString("DSN")
	}
	finalDBHost := viper.GetString("DB_HOST")
	finalDBPort := viper.GetString("DB_PORT")
	finalDBUser := viper.GetString("DB_USER")
	finalDBPassword := viper.GetString("DB_PASSWORD")
	finalDBName := viper.GetString("DB_NAME")
	finalDBSSLMode := viper.GetString("DB_SSLMODE")

	var connString string

	// Use DB_DSN if available, otherwise construct from individual parameters
	if dbDSN != "" {
		connString = dbDSN
	} else {
		// Validate required parameters
		if finalDBUser == "" || finalDBPassword == "" {
			log.Fatal("Database user and password are required")
		}

		connString = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			finalDBHost, finalDBPort, finalDBUser, finalDBPassword, finalDBName, finalDBSSLMode,
		)
	}

	fmt.Println("=== Database Migration Tool ===")
	if dbDSN == "" {
		fmt.Printf("Database: %s@%s:%s/%s\n", finalDBUser, finalDBHost, finalDBPort, finalDBName)
	} else {
		fmt.Printf("Database: Using DSN\n")
	}
	fmt.Printf("Command: %s\n\n", *command)

	// Initialize connection (but don't auto-migrate)
	if err := store.InitStorageWithoutMigration(connString); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()

	// Check for dirty state before executing commands
	version, dirty, err := store.GetMigrationVersion()
	if err != nil {
		log.Printf("Warning: Could not get migration version: %v", err)
	} else if dirty && *command != "force" && *command != "version" {
		fmt.Println("\n⚠ ERROR: Database is in DIRTY state!")
		fmt.Printf("Current version: %d (dirty)\n\n", version)
		fmt.Println("A previous migration failed and needs to be resolved.")
		fmt.Println("\nOptions to fix:")
		fmt.Println("1. Fix the failed migration file and run:")
		fmt.Printf("   --command=force --version=%d\n", version-1)
		fmt.Println("   (This will reset to the previous version)")
		fmt.Println("\n2. Or force to the current version if migration actually succeeded:")
		fmt.Printf("   --command=force --version=%d\n", version)
		fmt.Println("\n3. Check migration status:")
		fmt.Println("   --command=version")
		log.Fatal("\nCannot proceed with dirty database state")
	}

	// Execute command
	switch *command {
	case "up":
		fmt.Println("Running all pending migrations...")
		if err := store.MigrateUp(); err != nil {
			log.Printf("\n✗ Migration failed: %v\n", err)
			showVersion()
			fmt.Println("\n⚠ The database is now in a DIRTY state.")
			fmt.Println("To recover, fix the migration file and run:")
			fmt.Printf("  --command=force --version=%d\n", getCurrentVersion())
			log.Fatal("Migration aborted")
		}
		fmt.Println("✓ Migrations completed successfully")
		showVersion()

	case "down":
		fmt.Println("Rolling back last migration...")
		if err := store.MigrateDown(); err != nil {
			log.Printf("\n✗ Rollback failed: %v\n", err)
			showVersion()
			fmt.Println("\n⚠ The database may be in a DIRTY state.")
			fmt.Println("Check the version and use 'force' if needed.")
			log.Fatal("Rollback aborted")
		}
		fmt.Println("✓ Rollback completed successfully")
		showVersion()

	case "steps":
		fmt.Printf("Migrating by %d steps...\n", *steps)
		if err := store.MigrateSteps(*steps); err != nil {
			log.Printf("\n✗ Migration failed: %v\n", err)
			showVersion()
			fmt.Println("\n⚠ The database is now in a DIRTY state.")
			fmt.Println("To recover, fix the migration file and run:")
			fmt.Printf("  --command=force --version=%d\n", getCurrentVersion())
			log.Fatal("Migration aborted")
		}
		fmt.Println("✓ Migration completed successfully")
		showVersion()

	case "goto":
		if *target == 0 {
			log.Fatal("Please specify target version with -target flag")
		}
		fmt.Printf("Migrating to version %d...\n", *target)
		if err := store.MigrateTo(*target); err != nil {
			log.Printf("\n✗ Migration failed: %v\n", err)
			showVersion()
			fmt.Println("\n⚠ The database is now in a DIRTY state.")
			fmt.Println("To recover, fix the migration file and run:")
			fmt.Printf("  --command=force --version=%d\n", getCurrentVersion())
			log.Fatal("Migration aborted")
		}
		fmt.Println("✓ Migration completed successfully")
		showVersion()

	case "version":
		showVersion()

	case "force":
		if *forceVersion < 0 {
			log.Fatal("Please specify version with -version flag")
		}
		fmt.Printf("⚠ WARNING: Forcing migration version to %d\n", *forceVersion)
		fmt.Println("This should only be used to recover from a dirty state.")
		fmt.Println("\nThis will:")
		fmt.Println("- Set the version to", *forceVersion)
		fmt.Println("- Mark the database as clean")
		fmt.Println("- NOT run any migrations")
		fmt.Println("\nMake sure you've manually fixed any issues before proceeding.")

		if err := store.ForceMigrationVersion(*forceVersion); err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		fmt.Println("✓ Version forced successfully")
		showVersion()

	case "reset":
		fmt.Println("⚠ WARNING: This will reset to version 0 (no migrations applied)")
		fmt.Println("Use this only if you want to start fresh.")
		if err := store.ForceMigrationVersion(0); err != nil {
			log.Fatalf("Reset failed: %v", err)
		}
		fmt.Println("✓ Database reset to version 0")
		showVersion()

	default:
		log.Fatalf("Unknown command: %s\nAvailable commands: up, down, steps, goto, version, force, reset", *command)
	}
}

func getCurrentVersion() int {
	version, _, err := store.GetMigrationVersion()
	if err != nil {
		return 0
	}
	// Return previous version since current one failed
	if version > 0 {
		return int(version - 1)
	}
	return 0
}

func showVersion() {
	version, dirty, err := store.GetMigrationVersion()
	if err != nil {
		log.Printf("Warning: Could not get version: %v", err)
		return
	}

	fmt.Println("\n--- Current Status ---")
	fmt.Printf("Version: %d\n", version)
	if dirty {
		fmt.Println("State: ⚠ DIRTY (migration failed, needs attention)")
		fmt.Println("\nTo recover from dirty state:")
		fmt.Println("1. Fix the failed migration file")
		if version > 0 {
			fmt.Printf("2. Force to previous version: --command=force --version=%d\n", version-1)
		} else {
			fmt.Println("2. Force to version 0: --command=force --version=0")
		}
		fmt.Println("3. Run migrations again: --command=up")
	} else {
		fmt.Println("State: ✓ Clean")
	}
}
