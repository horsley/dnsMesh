package database

import (
	"dnsmesh/internal/models"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize initializes the database connection
func Initialize() error {
	sqlitePath := getEnv("SQLITE_PATH", "data/dnsmesh.db")
	if err := ensureSQLiteDir(sqlitePath); err != nil {
		return fmt.Errorf("failed to prepare sqlite directory: %w", err)
	}
	dsn := sqliteDSN(sqlitePath)

	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to access sqlite db handle: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)

	log.Println("Database connected successfully")

	// Run migrations
	if err := migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Clean up empty providers
	if err := cleanupEmptyProviders(); err != nil {
		return fmt.Errorf("failed to cleanup empty providers: %w", err)
	}

	return nil
}

// migrate runs database migrations
func migrate() error {
	log.Println("Running database migrations...")

	err := DB.AutoMigrate(
		&models.Provider{},
		&models.DNSRecord{},
		&models.AuditLog{},
	)

	if err != nil {
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}

// cleanupEmptyProviders removes providers that have no associated DNS records
func cleanupEmptyProviders() error {
	// Find all providers that have no DNS records
	var emptyProviders []models.Provider

	err := DB.Where("NOT EXISTS (SELECT 1 FROM dns_records WHERE dns_records.provider_id = providers.id)").
		Find(&emptyProviders).Error

	if err != nil {
		return err
	}

	if len(emptyProviders) > 0 {
		// Delete empty providers
		for _, provider := range emptyProviders {
			if err := DB.Delete(&provider).Error; err != nil {
				log.Printf("Failed to delete empty provider %d (%s): %v", provider.ID, provider.Name, err)
				continue
			}
			log.Printf("Deleted empty provider: %s (ID: %d)", provider.Name, provider.ID)
		}
		log.Printf("Cleanup completed: removed %d empty provider(s)", len(emptyProviders))
	}

	return nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func ensureSQLiteDir(path string) error {
	filePath := sqliteFilePath(path)
	if filePath == ":memory:" {
		return nil
	}
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func sqliteFilePath(path string) string {
	if idx := strings.Index(path, "?"); idx >= 0 {
		return path[:idx]
	}
	return path
}

func sqliteDSN(path string) string {
	if strings.Contains(path, "_fk=") {
		return path
	}
	if strings.Contains(path, "?") {
		return path + "&_fk=1"
	}
	return path + "?_fk=1"
}
