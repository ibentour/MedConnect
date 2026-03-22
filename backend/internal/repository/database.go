// Package repository handles database connection, migrations, and seeding.
package repository

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"medconnect-oriental/backend/internal/models"
)

// ConnectDatabase establishes a PostgreSQL connection with GORM and
// runs auto-migrations for all models.
func ConnectDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ Database connection established")

	// ── DROP Obsolete Columns ──────────────────
	// GORM doesn't drop columns automagically. We must remove
	// the old 'age' column manually because it is NOT NULL and is blocking
	// the insertion of new patients since the code no longer provides it.
	// if err := db.Exec("ALTER TABLE patients DROP COLUMN IF EXISTS age;").Error; err != nil {
	// 	log.Printf("[DB WARNING] Failed to drop obsolete 'age' column: %v", err)
	// }

	// Auto-migrate all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Department{},
		&models.Patient{},
		&models.Referral{},
		&models.Attachment{},
		&models.AuditLog{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}

	log.Println("✅ Database migrations completed")

	return db, nil
}
