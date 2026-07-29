package repository

import (
	"errors"

	"github.com/tingzhu/cv-review/backend/internal/database"
	"github.com/tingzhu/cv-review/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetNoteByGroupID fetches the note for a given question group.
// Returns gorm.ErrRecordNotFound if no note exists.
func GetNoteByGroupID(groupID uint) (*model.UserNote, error) {
	var note model.UserNote
	err := database.DB.Where("group_id = ?", groupID).First(&note).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

// UpsertNote inserts a new note or updates the existing one for the
// same group_id. This means the frontend can always use PUT without
// caring whether the note already exists.
func UpsertNote(note *model.UserNote) error {
	return database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"content", "updated_at"}),
	}).Create(note).Error
}

// DeleteNoteByGroupID removes the note for a given question group.
// It returns an error if no note exists for that group.
func DeleteNoteByGroupID(groupID uint) error {
	result := database.DB.Where("group_id = ?", groupID).Delete(&model.UserNote{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListAllNotes returns every note, ordered by most recently updated first.
func ListAllNotes() ([]model.UserNote, error) {
	var notes []model.UserNote
	err := database.DB.Order("updated_at DESC").Find(&notes).Error
	if err != nil {
		return nil, err
	}
	return notes, nil
}

// ErrNoteNotFound is a package-level sentinel the service layer can check.
var ErrNoteNotFound = errors.New("note not found")
