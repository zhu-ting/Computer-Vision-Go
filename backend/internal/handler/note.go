package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tingzhu/cv-review/backend/internal/repository"
	"github.com/tingzhu/cv-review/backend/internal/service"
)

// SaveNoteRequest is the JSON body for PUT /api/v1/notes/:group_id.
type SaveNoteRequest struct {
	Content string `json:"content" binding:"required"`
}

// ── Handlers ─────────────────────────────────────────────────────

// GetNote handles GET /api/v1/notes/:group_id.
func GetNote(c *gin.Context) {
	groupID, err := parseGroupID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	note, err := service.GetNote(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch note"})
		return
	}
	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// SaveNote handles PUT /api/v1/notes/:group_id.
// Creates a new note or updates the existing one (upsert).
func SaveNote(c *gin.Context) {
	groupID, err := parseGroupID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	var req SaveNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "content is required",
		})
		return
	}

	note, err := service.SaveNote(groupID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save note"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// DeleteNote handles DELETE /api/v1/notes/:group_id.
func DeleteNote(c *gin.Context) {
	groupID, err := parseGroupID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	if err := service.DeleteNote(groupID); err != nil {
		if errors.Is(err, repository.ErrNoteNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ListNotes handles GET /api/v1/notes.
func ListNotes(c *gin.Context) {
	notes, err := service.ListNotes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch notes"})
		return
	}
	c.JSON(http.StatusOK, notes)
}

// parseGroupID extracts the group_id from the route parameter.
func parseGroupID(c *gin.Context) (uint, error) {
	idStr := c.Param("group_id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
