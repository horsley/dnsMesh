package main

import (
	"dnsmesh/internal/models"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const batchSize = 200

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	if err := run(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}

func run() error {
	postgresDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "dnsmesh"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", "dnsmesh"),
		getEnv("DB_SSLMODE", "disable"),
	)

	log.Println("Connecting to Postgres...")
	postgresDB, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	sqlitePath := getEnv("SQLITE_PATH", "data/dnsmesh.db")
	if err := ensureSQLiteDir(sqlitePath); err != nil {
		return fmt.Errorf("failed to prepare sqlite directory: %w", err)
	}
	if err := ensureSQLiteDoesNotExist(sqlitePath); err != nil {
		return err
	}

	log.Println("Connecting to SQLite...")
	sqliteDB, err := gorm.Open(sqlite.Open(sqliteDSN(sqlitePath)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to sqlite: %w", err)
	}

	sqlDB, err := sqliteDB.DB()
	if err != nil {
		return fmt.Errorf("failed to access sqlite db handle: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := sqliteDB.AutoMigrate(&models.Provider{}, &models.DNSRecord{}, &models.AuditLog{}); err != nil {
		return fmt.Errorf("failed to migrate sqlite schema: %w", err)
	}

	if err := sqliteDB.Transaction(func(tx *gorm.DB) error {
		if err := copyProviders(postgresDB, tx); err != nil {
			return err
		}
		if err := copyRecords(postgresDB, tx); err != nil {
			return err
		}
		if err := copyAuditLogs(postgresDB, tx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	log.Printf("Migration completed. SQLite database created at %s", sqliteFilePath(sqlitePath))
	return nil
}

func copyProviders(source *gorm.DB, target *gorm.DB) error {
	var providers []models.Provider
	if err := source.Find(&providers).Error; err != nil {
		return fmt.Errorf("failed to read providers: %w", err)
	}
	if len(providers) == 0 {
		log.Println("No providers found to migrate")
		return nil
	}
	if err := target.Select("*").Omit("DNSRecords").CreateInBatches(&providers, batchSize).Error; err != nil {
		return fmt.Errorf("failed to insert providers: %w", err)
	}
	log.Printf("Migrated %d providers", len(providers))
	return nil
}

func copyRecords(source *gorm.DB, target *gorm.DB) error {
	var records []models.DNSRecord
	if err := source.Find(&records).Error; err != nil {
		return fmt.Errorf("failed to read DNS records: %w", err)
	}
	if len(records) == 0 {
		log.Println("No DNS records found to migrate")
		return nil
	}
	if err := target.Select("*").Omit("Provider").CreateInBatches(&records, batchSize).Error; err != nil {
		return fmt.Errorf("failed to insert DNS records: %w", err)
	}
	log.Printf("Migrated %d DNS records", len(records))
	return nil
}

func copyAuditLogs(source *gorm.DB, target *gorm.DB) error {
	var logs []models.AuditLog
	if err := source.Find(&logs).Error; err != nil {
		return fmt.Errorf("failed to read audit logs: %w", err)
	}
	if len(logs) == 0 {
		log.Println("No audit logs found to migrate")
		return nil
	}
	if err := target.Select("*").CreateInBatches(&logs, batchSize).Error; err != nil {
		return fmt.Errorf("failed to insert audit logs: %w", err)
	}
	log.Printf("Migrated %d audit logs", len(logs))
	return nil
}

func ensureSQLiteDoesNotExist(path string) error {
	filePath := sqliteFilePath(path)
	if filePath == ":memory:" {
		return nil
	}
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("sqlite database already exists: %s", filePath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to access sqlite path: %w", err)
	}
	return nil
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

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
