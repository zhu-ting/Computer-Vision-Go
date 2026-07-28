// Package database handles the PostgreSQL connection via GORM.
// This file contains stubs that will be fully implemented in Commit 2.
package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is the shared database connection pool.
// It's safe for concurrent use by multiple goroutines.
var DB *gorm.DB

// Connect establishes the connection to PostgreSQL and runs migrations.
func Connect(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("database connection established")
}

// Close gracefully shuts down the database connection pool.
func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
	log.Println("database connection closed")
}
