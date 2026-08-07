package model

import "time"

// Module is a top-level theme (e.g., "week7_deep_learning") that
// groups related QuestionGroups together. The quiz system can filter
// question groups and exam generation by module.
//
// Relationship: Module 1──N QuestionGroup (linked via QuestionGroup.ModuleID)
// No GORM FK constraint is defined because legacy rows use module_id = 0.
type Module struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null;uniqueIndex;size:100" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Module) TableName() string { return "modules" }
