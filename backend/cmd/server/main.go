// Package main is the entry point for the CV Review backend server.
// It wires together the database connection, services, and HTTP router,
// then starts listening on the configured port.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tingzhu/cv-review/backend/internal/database"
	"github.com/tingzhu/cv-review/backend/internal/router"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// ── Configuration ─────────────────────────────────────────────
	port := envOrDefault("API_PORT", "8080")

	// Build the PostgreSQL DSN (Data Source Name) from environment
	// variables. In production, sensitive values come from Docker secrets
	// or a vault; .env files are for local development only.
	sslMode := envOrDefault("DB_SSLMODE", "require")

    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        envOrDefault("DB_HOST", "localhost"),
        envOrDefault("DB_PORT", "5432"),
        envOrDefault("DB_USER", "cvreview"),
        envOrDefault("DB_PASSWORD", "cvreview"),
        envOrDefault("DB_NAME", "cvreview"),
        sslMode,
    )

	// ── Database ──────────────────────────────────────────────────
	database.Connect(dsn)
	defer database.Close()

	// ── Router ────────────────────────────────────────────────────
	r := router.Setup()

	// ── Start server ──────────────────────────────────────────────
	log.Printf("CV Review API listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

// envOrDefault returns the value of the environment variable named by key,
// or fallback if the variable is unset or empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
