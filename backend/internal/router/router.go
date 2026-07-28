// Package router sets up the Gin engine and registers all route groups.
// Each feature (exams, notes, etc.) will get its own route group in later commits.
package router

import "github.com/gin-gonic/gin"

// Setup creates and configures the Gin engine with middleware and routes.
func Setup() *gin.Engine {
	r := gin.Default()

	// ── API v1 routes ──────────────────────────────────────
	v1 := r.Group("/api/v1")
	{
		// Health check — already wired in main.go, but having it
		// here as well keeps route definitions centralized.
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	return r
}
