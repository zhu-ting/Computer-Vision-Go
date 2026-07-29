// Package router sets up the Gin engine and registers all route groups.
// Each feature (exams, notes, etc.) gets its own route group.
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tingzhu/cv-review/backend/internal/handler"
)

// Setup creates and configures the Gin engine with middleware and routes.
func Setup() *gin.Engine {
	r := gin.Default()

	// ── API v1 routes ──────────────────────────────────────
	v1 := r.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Exams — generation, progress saving, and submission
		exams := v1.Group("/exams")
		{
			exams.POST("", handler.CreateExam) // POST /api/v1/exams
		}
	}

	return r
}
