package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tingzhu/cv-review/backend/internal/service"
)

// CreateExamRequest is the expected JSON body for POST /api/v1/exams.
type CreateExamRequest struct {
	QuestionCount int `json:"question_count" binding:"required,min=1"`
}

// SaveAnswersRequest is the JSON body for incremental progress saves.
type SaveAnswersRequest struct {
	Answers []service.AnswerInput `json:"answers" binding:"required,dive"`
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
// Generates a new exam session with randomly selected questions,
// shuffles options, and returns a response with NO correct-answer data.
func CreateExam(c *gin.Context) {
	var req CreateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "question_count must be an integer",
		})
		return
	}

	if !AllowedQuestionCounts[req.QuestionCount] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "question_count must be one of 10, 20, 30, 40, 50",
			"allowed_values": []int{10, 20, 30, 40, 50},
		})
		return
	}

	exam, err := service.GenerateExam(req.QuestionCount)
	if err != nil {
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

// SaveAnswers handles PATCH /api/v1/exams/:id/answers.
// The frontend calls this on every page flip to persist the user's
// current selections. The endpoint is idempotent — repeated calls
// with the same data are safe.
func SaveAnswers(c *gin.Context) {
	examID, err := parseExamID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam ID"})
		return
	}

	var req SaveAnswersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "answers must be an array of {exam_question_id, selected_option_id}",
		})
		return
	}

	if err := service.SaveProgress(examID, req.Answers); err != nil {
		switch {
		case errors.Is(err, service.ErrExamNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "exam not found"})
		case errors.Is(err, service.ErrExamAlreadySubmitted):
			c.JSON(http.StatusConflict, gin.H{"error": "exam has already been submitted"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save answers"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "saved"})
}

// SubmitExam handles POST /api/v1/exams/:id/submit.
// Grades the exam, records the score, and returns a full snapshot
// with correct answers and analysis revealed.
func SubmitExam(c *gin.Context) {
	examID, err := parseExamID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam ID"})
		return
	}

	result, err := service.SubmitExam(examID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrExamNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "exam not found"})
		case errors.Is(err, service.ErrExamAlreadySubmitted):
			c.JSON(http.StatusConflict, gin.H{"error": "exam has already been submitted"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit exam"})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// parseExamID extracts the exam ID from the route parameter ":id".
func parseExamID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
