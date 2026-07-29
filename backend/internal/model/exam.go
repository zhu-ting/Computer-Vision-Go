package model

import "time"

// ExamStatus represents the lifecycle of an exam session.
type ExamStatus string

const (
	ExamStatusInProgress ExamStatus = "in_progress"
	ExamStatusSubmitted  ExamStatus = "submitted"
)

// Exam is an exam session created when a user starts a new practice set.
// The Score is only set after submission.
type Exam struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	TotalQuestions int       `gorm:"not null" json:"total_questions"`
	Score         *float64   `gorm:"type:decimal(5,2)" json:"score,omitempty"` // nil until submitted
	Status        ExamStatus `gorm:"not null;size:20;index;default:in_progress" json:"status"`
	StartedAt     time.Time  `gorm:"not null" json:"started_at"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty"` // nil until submitted

	// Associations
	ExamQuestions []ExamQuestion `gorm:"foreignKey:ExamID" json:"exam_questions,omitempty"`
}

func (Exam) TableName() string { return "exams" }
