package service

import (
	"errors"
	"time"

	"github.com/tingzhu/cv-review/backend/internal/model"
	"github.com/tingzhu/cv-review/backend/internal/repository"
	"gorm.io/gorm"
)

// ── DTO ──────────────────────────────────────────────────────────

// NoteResponse is the public representation of a user note.
// It includes the group_id (which is how notes are addressed) and
// the content + timestamps.
type NoteResponse struct {
	ID        uint      `json:"id"`
	GroupID   uint      `json:"group_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ── Service ──────────────────────────────────────────────────────

// GetNote returns the note for a question group, or nil if none exists.
func GetNote(groupID uint) (*NoteResponse, error) {
	note, err := repository.GetNoteByGroupID(groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // no note yet — not an error
		}
		return nil, err
	}
	return toNoteResponse(note), nil
}

// SaveNote creates or updates a note for the given question group.
// Because notes are tied to group_id (not question_id), they survive
// content updates — the user's notes persist across question versions.
func SaveNote(groupID uint, content string) (*NoteResponse, error) {
	note := &model.UserNote{
		GroupID: groupID,
		Content: content,
	}
	if err := repository.UpsertNote(note); err != nil {
		return nil, err
	}

	// Re-fetch to get the correct ID and timestamps after upsert.
	saved, err := repository.GetNoteByGroupID(groupID)
	if err != nil {
		return nil, err
	}
	return toNoteResponse(saved), nil
}

// DeleteNote removes the note for a question group.
func DeleteNote(groupID uint) error {
	err := repository.DeleteNoteByGroupID(groupID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return repository.ErrNoteNotFound
	}
	return err
}

// ListNotes returns all notes, newest first.
func ListNotes() ([]NoteResponse, error) {
	notes, err := repository.ListAllNotes()
	if err != nil {
		return nil, err
	}
	responses := make([]NoteResponse, 0, len(notes))
	for _, n := range notes {
		responses = append(responses, *toNoteResponse(&n))
	}
	return responses, nil
}

func toNoteResponse(n *model.UserNote) *NoteResponse {
	return &NoteResponse{
		ID:        n.ID,
		GroupID:   n.GroupID,
		Content:   n.Content,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}
