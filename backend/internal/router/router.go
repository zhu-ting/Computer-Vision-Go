// Package router sets up the Gin engine and registers all route groups.
package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tingzhu/cv-review/backend/internal/handler"
)

// Setup creates and configures the Gin engine with middleware and routes.
func Setup() *gin.Engine {
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true

	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}

    r.Use(cors.New(config))

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

		// Modules (themes) + question bank catalog
		modules := v1.Group("/modules")
		{
			modules.GET("", handler.ListModules)
			modules.POST("", handler.CreateModule)
		}

		questionGroups := v1.Group("/question-groups")
		{
			questionGroups.GET("", handler.ListQuestionGroups)
			questionGroups.POST("", handler.CreateQuestionGroup)
		}

		questions := v1.Group("/questions")
		{
			questions.GET("", handler.ListQuestions)
			questions.POST("", handler.CreateQuestion)
		}
	}

	return r
}
