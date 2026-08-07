package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tingzhu/cv-review/backend/internal/service"
)

// ── Request structs ──────────────────────────────────────────────

// CreateModuleRequest is the JSON body for POST /api/v1/modules.
type CreateModuleRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateQuestionGroupRequest is the JSON body for POST /api/v1/question-groups.
type CreateQuestionGroupRequest struct {
	ModuleID   uint   `json:"module_id" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Topic      string `json:"topic" binding:"required"`
	Difficulty string `json:"difficulty" binding:"required,oneof=easy medium hard"`
}

// CreateQuestionRequest is the JSON body for POST /api/v1/questions.
type CreateQuestionRequest struct {
	GroupID  uint                 `json:"group_id" binding:"required"`
	Content  string               `json:"content" binding:"required"`
	Analysis string               `json:"analysis" binding:"required"`
	Options  []service.OptionInput `json:"options" binding:"required,min=2,dive"`
}

// ── Module handlers ──────────────────────────────────────────────

// ListModules handles GET /api/v1/modules.
func ListModules(c *gin.Context) {
	modules, err := service.ListModules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch modules"})
		return
	}
	c.JSON(http.StatusOK, modules)
}

// CreateModule handles POST /api/v1/modules.
func CreateModule(c *gin.Context) {
	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	mod, err := service.CreateModule(req.Name)
	if err != nil {
		if errors.Is(err, service.ErrModuleExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create module"})
		return
	}

	c.JSON(http.StatusCreated, mod)
}

// ── Question group handlers ──────────────────────────────────────

// ListQuestionGroups handles GET /api/v1/question-groups?module_id=.
func ListQuestionGroups(c *gin.Context) {
	moduleID, err := parseOptionalUintQuery(c, "module_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid module_id"})
		return
	}

	groups, err := service.ListQuestionGroups(moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch question groups"})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// CreateQuestionGroup handles POST /api/v1/question-groups.
func CreateQuestionGroup(c *gin.Context) {
	var req CreateQuestionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module_id, title, topic, and difficulty (easy/medium/hard) are required"})
		return
	}

	group, err := service.CreateQuestionGroup(req.ModuleID, req.Title, req.Topic, req.Difficulty)
	if err != nil {
		if errors.Is(err, service.ErrModuleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create question group"})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// ── Question handlers ────────────────────────────────────────────

// ListQuestions handles GET /api/v1/questions?group_id=.
func ListQuestions(c *gin.Context) {
	groupID, err := parseOptionalUintQuery(c, "group_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	questions, err := service.ListQuestions(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch questions"})
		return
	}
	c.JSON(http.StatusOK, questions)
}

// CreateQuestion handles POST /api/v1/questions.
func CreateQuestion(c *gin.Context) {
	var req CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id, content, analysis, and at least 2 options are required"})
		return
	}

	q, err := service.CreateQuestion(req.GroupID, req.Content, req.Analysis, req.Options)
	if err != nil {
		if errors.Is(err, service.ErrQuestionGroupNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "question group not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidOptions) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create question"})
		return
	}

	c.JSON(http.StatusCreated, q)
}

// ── Helpers ──────────────────────────────────────────────────────

// parseOptionalUintQuery returns nil if the query param is absent or empty.
func parseOptionalUintQuery(c *gin.Context, key string) (*uint, error) {
	raw := c.Query(key)
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil, err
	}
	u := uint(v)
	return &u, nil
}
