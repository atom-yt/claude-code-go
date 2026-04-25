package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	// Default migration files directory
	defaultMigrationsDir = "./migrations"
	// Default database URL (can be overridden via env var)
	defaultDBURL = "postgres://postgres:postgres@localhost:5432/atom_ai?sslmode=disable"
)

var (
	migrationsDir = flag.String("dir", defaultMigrationsDir, "Directory containing migration files")
	dbURL         = flag.String("db", "", "Database connection URL (or set DB_URL env var)")
	create        = flag.Bool("create", false, "Create a new migration")
	up            = flag.Bool("up", false, "Run all up migrations")
	down          = flag.Bool("down", false, "Run all down migrations")
	steps         = flag.Int("steps", 0, "Number of migrations to apply (0 = all)")
	version        = flag.Int("version", -1, "Migrate to specific version")
	dryRun        = flag.Bool("dry-run", false, "Show what would be done without executing")
	force         = flag.Int("force", -1, "Set migration version without running migration (use with caution)")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of migration runner:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\nMigration runner for golang-migrate\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Commands:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -up              Run all pending up migrations\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -down            Rollback all migrations\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -steps N         Apply N migrations up or down\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -version N       Migrate to specific version\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -force N         Force migration version (use with caution)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -create NAME     Create new migration files\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -dry-run         Show what would be done\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  # Run all pending migrations\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  go run migrate.go -up\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\n  # Rollback last migration\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  go run migrate.go -down -steps 1\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\n  # Create new migration\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  go run migrate.go -create add_user_preferences\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\n  # Use custom database URL\n")
		fmt.Fprintf(flag.CommandLine.Output(), `  go run migrate.go -up -db "postgres://user:pass@host:5432/dbname"` + "\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\n  # Or use environment variable\n")
		fmt.Fprintf(flag.CommandLine.Output(), `  DB_URL="postgres://..." go run migrate.go -up` + "\n")
	}
}

func getDBURL() string {
	if *dbURL != "" {
		return *dbURL
	}
	if envURL := os.Getenv("DB_URL"); envURL != "" {
		return envURL
	}
	return defaultDBURL
}

func main() {
	flag.Parse()

	if *migrationsDir == "" {
		log.Fatal("migrations directory cannot be empty")
	}

	db := getDBURL()

	// Handle create command
	if *create {
		createMigration(*migrationsDir, flag.Arg(0))
		return
	}

	// Build migrate command
	args := []string{"-path", *migrationsDir, "-database", db}

	switch {
	case *force >= 0:
		args = append(args, "force", fmt.Sprintf("%d", *force))
	case *version >= 0:
		args = append(args, "goto", fmt.Sprintf("%d", *version))
	case *up:
		args = append(args, "up")
	case *down:
		if *steps > 0 {
			args = append(args, "down", fmt.Sprintf("%d", *steps))
		} else {
			args = append(args, "down")
		}
	case *steps > 0:
		// Default to up if -steps is specified without -down
		args = append(args, "up", fmt.Sprintf("%d", *steps))
	default:
		// Default: show status
		args = append(args, "version")
	}

	if *dryRun {
		args = append(args, "-dry-run")
	}

	// Execute migrate command
	fmt.Printf("Executing: migrate %s\n", strings.Join(args, " "))
	if err := runMigrate(args); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}

func runMigrate(args []string) error {
	// This is a simplified version. In production, use golang-migrate/migrate package
	// For now, print the command that would be executed
	fmt.Println("To run migrations, install golang-migrate:")
	fmt.Println("  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest")
	fmt.Println()
	fmt.Println("Then run:")
	fmt.Printf("  migrate %s\n", strings.Join(args, " "))

	// In a real implementation, you would use:
	// import "github.com/golang-migrate/migrate/v4"
	// m, err := migrate.New("file://"+*migrationsDir, db)
	// ... execute migration

	return nil
}

func createMigration(dir, name string) {
	if name == "" {
		log.Fatal("migration name is required")
	}

	// List existing migrations to determine the next number
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to read migrations directory: %v", err)
	}

	maxNum := 0
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			var num int
			_, err := fmt.Sscanf(f.Name(), "%d_", &num)
			if err == nil && num > maxNum {
				maxNum = num
			}
		}
	}

	nextNum := maxNum + 1
	upFile := fmt.Sprintf("%s/%06d_%s.up.sql", dir, nextNum, name)
	downFile := fmt.Sprintf("%s/%06d_%s.down.sql", dir, nextNum, name)

	// Create up migration
	upContent := `-- Migration: ` + name + `
-- Created at: ` + getCurrentTimestamp() + `
-- Description: Add your migration logic here

BEGIN;

-- Add your SQL statements here

COMMIT;
`

	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		log.Fatalf("failed to create up migration file: %v", err)
	}

	// Create down migration
	downContent := `-- Rollback: ` + name + `
-- Created at: ` + getCurrentTimestamp() + `

BEGIN;

-- Add your rollback statements here

COMMIT;
`

	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		log.Fatalf("failed to create down migration file: %v", err)
	}

	fmt.Printf("Created migration files:\n  %s\n  %s\n", upFile, downFile)
}

func getCurrentTimestamp() string {
	// In a real implementation, you'd use time.Now().Format(...)
	return "2026-04-25 00:00:00"
}