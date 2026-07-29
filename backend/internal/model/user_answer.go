package model

import "time"

// UserAnswer records which option the user selected for a given exam question.
// It is updated incrementally as the user navigates between pages.
// selected_option_id is nil until the user makes a choice.
type UserAnswer struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ExamQuestionID  uint      `gorm:"uniqueIndex;not null" json:"exam_question_id"`
	SelectedOptionID *uint    `json:"selected_option_id,omitempty"` // nil = unanswered
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	ExamQuestion   ExamQuestion `gorm:"foreignKey:ExamQuestionID" json:"-"`
	SelectedOption *Option      `gorm:"foreignKey:SelectedOptionID" json:"-"`
}

func (UserAnswer) TableName() string { return "user_answers" }
