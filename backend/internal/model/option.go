package model

// Option is one of the answer choices for a specific Question version.
// The is_correct flag is NEVER sent to the frontend during an active exam —
// the service layer strips it before serialization.
type Option struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	QuestionID uint   `gorm:"not null;index" json:"question_id"`
	Content    string `gorm:"type:text;not null" json:"content"`    // the option text
	IsCorrect  bool   `gorm:"not null;default:false" json:"is_correct"` // hidden from exam API
	SortOrder  int    `gorm:"not null;default:0" json:"sort_order"` // original ordering

	// Association
	Question Question `gorm:"foreignKey:QuestionID" json:"-"`
}

func (Option) TableName() string { return "options" }
