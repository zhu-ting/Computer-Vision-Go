package model

import "time"

// UserNote is a personal note attached to a QuestionGroup.
// It is NOT tied to a specific question version, so it survives
// content updates.
type UserNote struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GroupID   uint      `gorm:"not null;uniqueIndex" json:"group_id"` // one note per group
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Association
	Group QuestionGroup `gorm:"foreignKey:GroupID" json:"-"`
}

func (UserNote) TableName() string { return "user_notes" }
