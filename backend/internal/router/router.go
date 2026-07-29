// Package router sets up the Gin engine and registers all route groups.
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
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Exams
		exams := v1.Group("/exams")
		{
			exams.POST("", handler.CreateExam)              // generate new exam
			exams.PATCH("/:id/answers", handler.SaveAnswers) // incremental save
			exams.POST("/:id/submit", handler.SubmitExam)    // submit & grade
		}
	}

	return r
}
