// Package database handles the PostgreSQL connection via GORM,
// runs auto-migration for all models, and triggers seed data loading
// when the database is empty.
package database

import (
	"log"
	"time"

	"github.com/tingzhu/cv-review/backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the shared database connection pool.
// It is safe for concurrent use by multiple goroutines.
var DB *gorm.DB

// Connect establishes the connection to PostgreSQL, runs auto-migration,
// and seeds the database if it is empty.
//
// The DSN format expected by the postgres driver:
//
//	host=localhost port=5432 user=cvreview password=secret dbname=cvreview sslmode=disable
func Connect(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Log slow queries (over 200ms) to help catch missing indexes early.
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Configure the underlying connection pool.
	// These values are conservative for a dev/small-deploy setup;
	// tune them based on production load.
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)                // max simultaneous connections
	sqlDB.SetMaxIdleConns(5)                 // keep up to 5 idle connections ready
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // recycle connections after 5 min

	log.Println("[db] connection established")

	// ── Auto-migration ─────────────────────────────────────────
	// AutoMigrate creates tables, adds missing columns, and creates
	// indexes — but it will NOT delete or rename existing columns.
	// For production, use versioned migrations instead.
	if err := DB.AutoMigrate(
		&model.Module{},
		&model.QuestionGroup{},
		&model.Question{},
		&model.Option{},
		&model.Exam{},
		&model.ExamQuestion{},
		&model.UserAnswer{},
		&model.UserNote{},
	); err != nil {
		log.Fatalf("failed to auto-migrate: %v", err)
	}
	log.Println("[db] auto-migration complete")

	// ── Seed data ──────────────────────────────────────────────
	// Idempotent: only inserts if the questions table is empty.
	SeedIfEmpty()
}

// Close gracefully shuts down the database connection pool.
// Call this during server shutdown (e.g., via defer in main).
func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Printf("[db] error getting sql.DB for close: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("[db] error closing connection: %v", err)
		}
	}
	log.Println("[db] connection closed")
}
