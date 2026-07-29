package model

// ExamQuestion is the immutable snapshot linking an exam to a specific
// question version. It stores the shuffled option order so the same
// shuffle is preserved for the lifetime of this exam session.
type ExamQuestion struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	ExamID      uint   `gorm:"not null;index" json:"exam_id"`
	QuestionID  uint   `gorm:"not null;index" json:"question_id"`     // pinned to a specific version
	OptionOrder string `gorm:"type:jsonb;not null" json:"option_order"` // JSON array of shuffled option IDs, e.g. [3,1,4,2]

	// Associations
	Exam       Exam        `gorm:"foreignKey:ExamID" json:"-"`
	Question   Question    `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	UserAnswer *UserAnswer `gorm:"foreignKey:ExamQuestionID" json:"user_answer,omitempty"`
}

func (ExamQuestion) TableName() string { return "exam_questions" }
