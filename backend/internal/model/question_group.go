package model

import "time"

// QuestionGroup represents a logical question that can have multiple versions.
// Notes are tied to this group, not to individual question versions,
// so user notes survive content updates.
type QuestionGroup struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"not null;size:500" json:"title"`    // short topic summary
	Topic     string    `gorm:"not null;size:200;index" json:"topic"` // e.g. "Edge Detection", "CNN"
	Difficulty string   `gorm:"not null;size:20;index" json:"difficulty"` // "easy", "medium", "hard"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Associations
	Questions []Question `gorm:"foreignKey:GroupID" json:"questions,omitempty"`
	Notes     []UserNote `gorm:"foreignKey:GroupID" json:"notes,omitempty"`
}

func (QuestionGroup) TableName() string { return "question_groups" }
