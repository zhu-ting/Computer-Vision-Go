package model

import "time"

// Question is a specific version of a question within a QuestionGroup.
// When a question is updated, a new Question row is inserted with an
// incremented version number — this preserves historical exam snapshots
// that reference the old version.
type Question struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	GroupID   uint   `gorm:"not null;index" json:"group_id"`
	Content   string `gorm:"type:text;not null" json:"content"`   // the question text
	Analysis  string `gorm:"type:text;not null" json:"analysis"`  // answer explanation
	Version   int    `gorm:"not null;default:1" json:"version"`   // version number within the group
	CreatedAt time.Time `json:"created_at"`

	// Associations
	Group   QuestionGroup `gorm:"foreignKey:GroupID" json:"-"`
	Options []Option      `gorm:"foreignKey:QuestionID" json:"options,omitempty"`
}

func (Question) TableName() string { return "questions" }
