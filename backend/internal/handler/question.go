package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tingzhu/cv-review/backend/internal/service"
)

// ── Request DTOs ─────────────────────────────────────────────────

// CreateQuestionRequest is the JSON body for POST /api/v1/questions.
type CreateQuestionRequest struct {
	GroupID         uint                     `json:"group_id"`
	GroupTitle      string                   `json:"group_title"`
	GroupTopic      string                   `json:"group_topic"`
	GroupDifficulty string                   `json:"group_difficulty"`
	Content         string                   `json:"content" binding:"required"`
	Analysis        string                   `json:"analysis" binding:"required"`
	Options         []CreateOptionRequest    `json:"options" binding:"required,min=1,dive"`
}

// CreateOptionRequest is a single option within a create/update request.
type CreateOptionRequest struct {
	Content   string `json:"content" binding:"required"`
	IsCorrect bool   `json:"is_correct"`
	SortOrder int    `json:"sort_order"`
}

// UpdateQuestionRequest is the JSON body for PUT /api/v1/questions/:id.
type UpdateQuestionRequest struct {
	Content  string                `json:"content" binding:"required"`
	Analysis string                `json:"analysis" binding:"required"`
	Options  []CreateOptionRequest `json:"options" binding:"required,min=1,dive"`
}

// ── Handlers ─────────────────────────────────────────────────────

// ListQuestions handles GET /api/v1/questions.
// Supports ?page=1&page_size=20 query parameters.
func ListQuestions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := service.ListQuestions(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch questions"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetQuestion handles GET /api/v1/questions/:id.
func GetQuestion(c *gin.Context) {
	id, err := parseQuestionID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question ID"})
		return
	}

	question, err := service.GetQuestion(id)
	if err != nil {
		if errors.Is(err, service.ErrQuestionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch question"})
		return
	}

	c.JSON(http.StatusOK, question)
}

// CreateQuestion handles POST /api/v1/questions.
func CreateQuestion(c *gin.Context) {
	var req CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "content, analysis, and at least one option are required",
		})
		return
	}

	// Map handler request → service input.
	input := service.CreateQuestionInput{
		GroupID:         req.GroupID,
		GroupTitle:      req.GroupTitle,
		GroupTopic:      req.GroupTopic,
		GroupDifficulty: req.GroupDifficulty,
		Content:         req.Content,
		Analysis:        req.Analysis,
		Options:         mapOptions(req.Options),
	}

	question, err := service.CreateQuestion(input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGroupNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "question group not found"})
		case errors.Is(err, service.ErrNoCorrectOption):
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one option must be marked as correct"})
		case errors.Is(err, service.ErrNoOptions):
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one option is required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create question"})
		}
		return
	}

	c.JSON(http.StatusCreated, question)
}

// UpdateQuestion handles PUT /api/v1/questions/:id.
// Creates a new version of the question instead of mutating the old one.
func UpdateQuestion(c *gin.Context) {
	id, err := parseQuestionID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question ID"})
		return
	}

	var req UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "content, analysis, and at least one option are required",
		})
		return
	}

	input := service.UpdateQuestionInput{
		Content:  req.Content,
		Analysis: req.Analysis,
		Options:  mapOptions(req.Options),
	}

	question, err := service.UpdateQuestion(id, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrQuestionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		case errors.Is(err, service.ErrNoCorrectOption):
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one option must be marked as correct"})
		case errors.Is(err, service.ErrNoOptions):
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one option is required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update question"})
		}
		return
	}

	c.JSON(http.StatusOK, question)
}

// DeleteQuestion handles DELETE /api/v1/questions/:id.
func DeleteQuestion(c *gin.Context) {
	id, err := parseQuestionID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question ID"})
		return
	}

	if err := service.DeleteQuestion(id); err != nil {
		if errors.Is(err, service.ErrQuestionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete question"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ListGroups handles GET /api/v1/groups.
func ListGroups(c *gin.Context) {
	groups, err := service.ListGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch groups"})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// ── Helpers ──────────────────────────────────────────────────────

// parseQuestionID extracts the question ID from the route parameter ":id".
func parseQuestionID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// mapOptions converts handler request options to service input options.
func mapOptions(opts []CreateOptionRequest) []service.CreateOptionInput {
	inputs := make([]service.CreateOptionInput, 0, len(opts))
	for _, o := range opts {
		inputs = append(inputs, service.CreateOptionInput{
			Content:   o.Content,
			IsCorrect: o.IsCorrect,
			SortOrder: o.SortOrder,
		})
	}
	return inputs
}
