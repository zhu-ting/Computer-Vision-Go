// Package router sets up the Gin engine and registers all route groups.
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tingzhu/cv-review/backend/internal/handler"
)

// Setup creates and configures the Gin engine with middleware and routes.
func Setup() *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Exams
		exams := v1.Group("/exams")
		{
			exams.POST("", handler.CreateExam)
			exams.PATCH("/:id/answers", handler.SaveAnswers)
			exams.POST("/:id/submit", handler.SubmitExam)
		}

		// Notes — keyed by question group_id, not question version
		notes := v1.Group("/notes")
		{
			notes.GET("", handler.ListNotes)
			notes.GET("/:group_id", handler.GetNote)
			notes.PUT("/:group_id", handler.SaveNote)
			notes.DELETE("/:group_id", handler.DeleteNote)
		}
	}

	return r
}
