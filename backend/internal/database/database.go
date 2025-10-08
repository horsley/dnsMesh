package database

import (
	"dnsmesh/internal/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize initializes the database connection
func Initialize() error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "dnsmesh"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", "dnsmesh"),
		getEnv("DB_SSLMODE", "disable"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")

	// Run migrations
	if err := migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create default admin user if not exists
	if err := createDefaultUser(); err != nil {
		return fmt.Errorf("failed to create default user: %w", err)
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
		&models.User{},
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

// createDefaultUser creates the default admin user if it doesn't exist
func createDefaultUser() error {
	var count int64
	DB.Model(&models.User{}).Count(&count)

	if count == 0 {
		username := getEnv("ADMIN_USERNAME", "admin")
		password := getEnv("ADMIN_PASSWORD", "admin")

		user := models.User{
			Username: username,
		}

		if err := user.HashPassword(password); err != nil {
			return err
		}

		if err := DB.Create(&user).Error; err != nil {
			return err
		}

		log.Printf("Default admin user created: %s", username)
	}

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
