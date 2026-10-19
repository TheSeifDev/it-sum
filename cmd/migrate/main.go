package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

var dbURLRegex = regexp.MustCompile(`postgres(ql)?://[^\s'"]+`)

func main() {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required for migrations")
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		log.Fatalf("Invalid migrations directory: %v", err)
	}
	// Format as file:// URL with forward slashes for cross-platform compatibility
	sourceURL := fmt.Sprintf("file://%s", filepath.ToSlash(absPath))

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		log.Fatalf("Migration initialization failed: %v", sanitizeError(err))
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Printf("Error closing migration source: %v", sanitizeError(sourceErr))
		}
		if dbErr != nil {
			log.Printf("Error closing migration database: %v", sanitizeError(dbErr))
		}
	}()

	command := args[0]
	switch command {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("No pending migrations to apply (database schema is up to date)")
				return
			}
			if errors.Is(err, os.ErrNotExist) || isNoMigrationsError(err) {
				log.Println("No migration files found in migrations directory")
				return
			}
			log.Fatalf("Migration failed: %v", sanitizeError(err))
		}
		log.Println("Migrations applied successfully")

	case "down":
		// Roll back 1 step by default to prevent accidental complete drops
		steps := 1
		if len(args) > 1 && args[1] == "all" {
			if err := m.Down(); err != nil {
				if errors.Is(err, migrate.ErrNoChange) {
					log.Println("No migrations to revert")
					return
				}
				if isNoMigrationsError(err) {
					log.Println("No migrations found to revert")
					return
				}
				log.Fatalf("Migration down failed: %v", sanitizeError(err))
			}
			log.Println("All migrations reverted successfully")
			return
		}

		if err := m.Steps(-steps); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("No migrations to revert")
				return
			}
			if isNoMigrationsError(err) {
				log.Println("No migrations found to revert")
				return
			}
			log.Fatalf("Migration rollback failed: %v", sanitizeError(err))
		}
		log.Printf("Reverted %d migration step(s) successfully\n", steps)

	case "step":
		if len(args) < 2 {
			log.Fatal("Usage: migrate step <n> (positive for up, negative for down)")
		}
		n, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("Invalid step count %q: must be an integer", args[1])
		}
		if err := m.Steps(n); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("No migration change applied")
				return
			}
			log.Fatalf("Migration step failed: %v", sanitizeError(err))
		}
		log.Printf("Applied %d migration step(s) successfully\n", n)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				log.Println("No migrations have been applied yet")
				return
			}
			log.Fatalf("Failed to fetch migration version: %v", sanitizeError(err))
		}
		log.Printf("Current migration version: %d (dirty: %t)\n", version, dirty)

	default:
		log.Printf("Unknown migration command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func isNoMigrationsError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, os.ErrNotExist) || errors.Is(err, migrate.ErrNilVersion)
}

func sanitizeError(err error) error {
	if err == nil {
		return nil
	}
	// Redact connection strings or passwords from the error text
	msg := err.Error()
	sanitized := dbURLRegex.ReplaceAllString(msg, "postgres://[REDACTED]")
	return errors.New(sanitized)
}

func printUsage() {
	fmt.Println("Usage: go run ./cmd/migrate [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up          Apply all pending migrations")
	fmt.Println("  down        Roll back the last migration step")
	fmt.Println("  down all    Roll back all migrations")
	fmt.Println("  step <n>    Apply <n> migrations (positive for up, negative for down)")
	fmt.Println("  version     Print the current migration version")
}
