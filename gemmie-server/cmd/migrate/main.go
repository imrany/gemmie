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
	command := pflag.String("command", "version", "Migration command: up, down, steps, goto, version, force")
	steps := pflag.Int("steps", 1, "Number of steps for migration (used with 'steps' command)")
	target := pflag.Uint("target", 0, "Target version (used with 'goto' command)")
	forceVersion := pflag.Int("version", 0, "Version to force (used with 'force' command)")

	pflag.String("db-host", "localhost", "Database host")
	pflag.String("db-port", "5432", "Database port")
	pflag.String("db-user", "", "Database user")
	pflag.String("db-password", "", "Database password")
	pflag.String("db-name", "gemmie", "Database name")
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

	// Execute command
	switch *command {
	case "up":
		fmt.Println("Running all pending migrations...")
		if err := store.MigrateUp(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✓ Migrations completed successfully")
		showVersion()

	case "down":
		fmt.Println("Rolling back last migration...")
		if err := store.MigrateDown(); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("✓ Rollback completed successfully")
		showVersion()

	case "steps":
		fmt.Printf("Migrating by %d steps...\n", *steps)
		if err := store.MigrateSteps(*steps); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✓ Migration completed successfully")
		showVersion()

	case "goto":
		if *target == 0 {
			log.Fatal("Please specify target version with -target flag")
		}
		fmt.Printf("Migrating to version %d...\n", *target)
		if err := store.MigrateTo(*target); err != nil {
			log.Fatalf("Migration failed: %v", err)
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
		if err := store.ForceMigrationVersion(*forceVersion); err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		fmt.Println("✓ Version forced successfully")
		showVersion()

	default:
		log.Fatalf("Unknown command: %s\nAvailable commands: up, down, steps, goto, version, force", *command)
	}
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
		fmt.Println("Run with -command=force -version=X to recover")
	} else {
		fmt.Println("State: ✓ Clean")
	}
}
