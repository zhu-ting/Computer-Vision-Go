package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tingzhu/cv-review/backend/internal/service"
)

// CreateExamRequest is the expected JSON body for POST /api/v1/exams.
// QuestionCount must be one of the allowed values.
type CreateExamRequest struct {
	QuestionCount int `json:"question_count" binding:"required,min=1"`
}

// AllowedQuestionCounts defines the valid numbers of questions a user
// can request for a single exam session.
var AllowedQuestionCounts = map[int]bool{
	10: true,
	20: true,
	30: true,
	40: true,
	50: true,
}

// CreateExam handles POST /api/v1/exams.
// It generates a new exam session with randomly selected questions,
// shuffles the options for each question, and returns a response
// that contains NO correct-answer data.
func CreateExam(c *gin.Context) {
	var req CreateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "question_count must be an integer",
		})
		return
	}

	// Validate the question count against the allowed set.
	// We validate here (not in the binding tag) so we can return a
	// clear error message listing valid values.
	if !AllowedQuestionCounts[req.QuestionCount] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":            "question_count must be one of 10, 20, 30, 40, 50",
			"allowed_values":   []int{10, 20, 30, 40, 50},
		})
		return
	}

	// Delegate to the service layer — the handler knows nothing about
	// shuffling, DTOs, or database queries.
	exam, err := service.GenerateExam(req.QuestionCount)
	if err != nil {
		// Distinguish between "no questions" (misconfiguration) and
		// actual server errors (database down, etc.).
		if errors.Is(err, service.ErrNoQuestions) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "no questions available — seed the database first",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate exam",
		})
		return
	}

	c.JSON(http.StatusCreated, exam)
}
