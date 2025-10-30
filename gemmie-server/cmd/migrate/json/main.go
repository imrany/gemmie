package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/imrany/gemmie/gemmie-server/store"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Legacy JSON storage structure
type LegacyStorage struct {
	Users        map[string]store.User        `json:"users"`
	UserData     map[string]store.UserData    `json:"user_data"`
	Transactions map[string]store.Transaction `json:"transactions"`
	Mu           sync.RWMutex                 `json:"-"`
}

func main() {
	// Load .env if present
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, using defaults")
	} else {
		slog.Info(".env file loaded successfully")
	}

	// Command line flags
	jsonFile := pflag.String("json", "./gemmie_data.json", "Path to JSON file")
	pflag.String("db-host", "localhost", "Database host")
	pflag.String("db-port", "5432", "Database port")
	pflag.String("db-user", "", "Database user")
	pflag.String("db-password", "", "Database password")
	pflag.String("db-name", "gemmie", "Database name")
	pflag.String("db-sslmode", "disable", "Database SSL mode")
	dryRun := pflag.Bool("dry-run", false, "Perform a dry run without inserting data")

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

	// Get values from viper (which checks env vars first, then flags)
	finalDBHost := viper.GetString("DB_HOST")
	finalDBPort := viper.GetString("DB_PORT")
	finalDBUser := viper.GetString("DB_USER")
	finalDBPassword := viper.GetString("DB_PASSWORD")
	finalDBName := viper.GetString("DB_NAME")
	finalDBSSLMode := viper.GetString("DB_SSLMODE")

	// Validate required parameters
	if finalDBUser == "" || finalDBPassword == "" {
		log.Fatal("Database user and password are required (via flags or environment variables)")
	}

	fmt.Println("=== JSON to PostgreSQL Migration Tool ===")
	fmt.Printf("JSON File: %s\n", *jsonFile)
	fmt.Printf("Database: %s@%s:%s/%s\n", finalDBUser, finalDBHost, finalDBPort, finalDBName)
	fmt.Printf("Dry Run: %v\n\n", *dryRun)

	// Step 1: Read JSON file
	fmt.Println("Step 1: Reading JSON file...")
	data, err := os.ReadFile(*jsonFile)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	// Step 2: Parse JSON
	fmt.Println("Step 2: Parsing JSON data...")
	var legacy LegacyStorage
	if err := json.Unmarshal(data, &legacy); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	fmt.Printf("Found:\n")
	fmt.Printf("  - %d users\n", len(legacy.Users))
	fmt.Printf("  - %d user data records\n", len(legacy.UserData))
	fmt.Printf("  - %d transactions\n\n", len(legacy.Transactions))

	if *dryRun {
		fmt.Println("DRY RUN MODE - No data will be inserted")
		displaySampleData(legacy)
		return
	}

	// Step 3: Connect to PostgreSQL using viper values
	fmt.Println("Step 3: Connecting to PostgreSQL...")
	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		finalDBHost, finalDBPort, finalDBUser, finalDBPassword, finalDBName, finalDBSSLMode,
	)

	if err := store.InitStorage(connString); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()
	fmt.Println("Connected successfully!")

	// Step 4: Migrate Users
	fmt.Println("Step 4: Migrating users...")
	userCount := 0
	userErrors := 0
	for userID, user := range legacy.Users {
		if err := store.CreateUser(user); err != nil {
			fmt.Printf("  ⚠ Failed to migrate user %s (%s): %v\n", userID, user.Email, err)
			userErrors++
		} else {
			userCount++
		}
	}
	fmt.Printf("  ✓ Migrated %d/%d users successfully", userCount, len(legacy.Users))
	if userErrors > 0 {
		fmt.Printf(" (%d errors)", userErrors)
	}
	fmt.Println()

	// Step 5: Migrate User Data
	fmt.Println("Step 5: Migrating user data...")
	userDataCount := 0
	userDataErrors := 0
	for userID, userData := range legacy.UserData {
		if err := store.CreateUserData(userData); err != nil {
			fmt.Printf("  ⚠ Failed to migrate user data for %s: %v\n", userID, err)
			userDataErrors++
		} else {
			userDataCount++
		}
	}
	fmt.Printf("  ✓ Migrated %d/%d user data records successfully", userDataCount, len(legacy.UserData))
	if userDataErrors > 0 {
		fmt.Printf(" (%d errors)", userDataErrors)
	}
	fmt.Println()

	// Step 6: Migrate Transactions
	fmt.Println("Step 6: Migrating transactions...")
	txCount := 0
	txErrors := 0
	for txID, transaction := range legacy.Transactions {
		if err := store.CreateTransaction(transaction); err != nil {
			fmt.Printf("  ⚠ Failed to migrate transaction %s: %v\n", txID, err)
			txErrors++
		} else {
			txCount++
		}
	}
	fmt.Printf("  ✓ Migrated %d/%d transactions successfully", txCount, len(legacy.Transactions))
	if txErrors > 0 {
		fmt.Printf(" (%d errors)", txErrors)
	}
	fmt.Println()

	// Step 7: Create backup of JSON file
	fmt.Println("Step 7: Creating backup of JSON file...")
	backupFile := *jsonFile + ".migrated." + time.Now().Format("20060102_150405")
	if err := os.Rename(*jsonFile, backupFile); err != nil {
		fmt.Printf("  ⚠ Failed to create backup: %v\n", err)
	} else {
		fmt.Printf("  ✓ Backup created: %s\n", backupFile)
	}

	// Summary
	fmt.Println("=== Migration Complete ===")
	fmt.Printf("Successfully migrated:\n")
	fmt.Printf("  - Users: %d/%d\n", userCount, len(legacy.Users))
	fmt.Printf("  - User Data: %d/%d\n", userDataCount, len(legacy.UserData))
	fmt.Printf("  - Transactions: %d/%d\n", txCount, len(legacy.Transactions))

	totalErrors := userErrors + userDataErrors + txErrors
	if totalErrors > 0 {
		fmt.Printf("⚠ Total errors: %d\n", totalErrors)
		fmt.Println("Review the errors above for details.")
	} else {
		fmt.Println("✓ All data migrated successfully!")
	}
}

func displaySampleData(legacy LegacyStorage) {
	fmt.Println("=== Sample Data Preview ===")

	// Show first user
	if len(legacy.Users) > 0 {
		fmt.Println("Sample User:")
		for _, user := range legacy.Users {
			fmt.Printf("  ID: %s\n", user.ID)
			fmt.Printf("  Username: %s\n", user.Username)
			fmt.Printf("  Email: %s\n", user.Email)
			fmt.Printf("  Created: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Plan: %s\n", user.Plan)
			fmt.Printf("  Email Verified: %v\n", user.EmailVerified)
			fmt.Printf("  Response Mode: %s\n", user.ResponseMode)
			break
		}
	}

	// Show first user data
	if len(legacy.UserData) > 0 {
		fmt.Println("Sample User Data:")
		for _, userData := range legacy.UserData {
			fmt.Printf("  User ID: %s\n", userData.UserID)
			fmt.Printf("  Chats Length: %d characters\n", len(userData.Chats))
			fmt.Printf("  Updated: %s\n", userData.UpdatedAt.Format("2006-01-02 15:04:05"))
			break
		}
	}

	// Show first transaction
	if len(legacy.Transactions) > 0 {
		fmt.Println("Sample Transaction:")
		for _, tx := range legacy.Transactions {
			fmt.Printf("  ID: %s\n", tx.ID)
			fmt.Printf("  Amount: %d\n", tx.Amount)
			fmt.Printf("  Phone: %s\n", tx.PhoneNumber)
			fmt.Printf("  Status: %s\n", tx.Status)
			fmt.Printf("  Created: %s\n", tx.CreatedAt.Format("2006-01-02 15:04:05"))
			break
		}
	}

	fmt.Println("To proceed with migration, run without --dry-run flag")
}
