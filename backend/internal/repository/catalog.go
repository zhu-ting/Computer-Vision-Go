package repository

import (
	"github.com/tingzhu/cv-review/backend/internal/database"
	"github.com/tingzhu/cv-review/backend/internal/model"
)

// ── Modules ──────────────────────────────────────────────────────

// ListModules returns all modules ordered by name.
func ListModules() ([]model.Module, error) {
	var modules []model.Module
	err := database.DB.Order("name ASC").Find(&modules).Error
	if err != nil {
		return nil, err
	}
	return modules, nil
}

// FindModuleByName looks up a module by exact name (case-sensitive).
// Returns gorm.ErrRecordNotFound if no match.
func FindModuleByName(name string) (*model.Module, error) {
	var m model.Module
	err := database.DB.Where("name = ?", name).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// CreateModule inserts a new module.
func CreateModule(m *model.Module) error {
	return database.DB.Create(m).Error
}

// ModuleExists checks whether a module with the given ID exists.
func ModuleExists(id uint) (bool, error) {
	var count int64
	err := database.DB.Model(&model.Module{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ── Question Groups ──────────────────────────────────────────────

// ListQuestionGroups returns question groups, optionally filtered by module.
// When moduleID is non-nil, only groups belonging to that module are returned.
func ListQuestionGroups(moduleID *uint) ([]model.QuestionGroup, error) {
	var groups []model.QuestionGroup
	q := database.DB.Order("id ASC")
	if moduleID != nil {
		q = q.Where("module_id = ?", *moduleID)
	}
	err := q.Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// CountQuestionsByGroup returns a map from group_id to question count.
func CountQuestionsByGroup() (map[uint]int64, error) {
	type row struct {
		GroupID uint
		Count   int64
	}
	var rows []row
	err := database.DB.Model(&model.Question{}).
		Select("group_id, count(*) as count").
		Group("group_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]int64, len(rows))
	for _, r := range rows {
		result[r.GroupID] = r.Count
	}
	return result, nil
}

// QuestionGroupExists checks whether a question group with the given ID exists.
func QuestionGroupExists(id uint) (bool, error) {
	var count int64
	err := database.DB.Model(&model.QuestionGroup{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// CreateQuestionGroup inserts a new question group.
func CreateQuestionGroup(g *model.QuestionGroup) error {
	return database.DB.Create(g).Error
}

// ── Questions ────────────────────────────────────────────────────

// ListQuestions returns questions, optionally filtered by group,
// with options eagerly loaded in sort_order.
func ListQuestions(groupID *uint) ([]model.Question, error) {
	var questions []model.Question
	q := database.DB.
		Preload("Options", func(db *gormDB) *gormDB {
			return db.Order("sort_order ASC")
		}).
		Order("version DESC")
	if groupID != nil {
		q = q.Where("group_id = ?", *groupID)
	}
	err := q.Find(&questions).Error
	if err != nil {
		return nil, err
	}
	return questions, nil
}

// MaxQuestionVersion returns the highest version number for a question group,
// or 0 if no questions exist yet.
func MaxQuestionVersion(groupID uint) (int, error) {
	var maxVersion int
	err := database.DB.Model(&model.Question{}).
		Select("COALESCE(MAX(version), 0)").
		Where("group_id = ?", groupID).
		Scan(&maxVersion).Error
	return maxVersion, err
}

// CreateQuestionWithOptions inserts a question and its options in a transaction.
func CreateQuestionWithOptions(q *model.Question, opts []model.Option) error {
	return database.DB.Transaction(func(tx *gormDB) error {
		if err := tx.Create(q).Error; err != nil {
			return err
		}
		for i := range opts {
			opts[i].QuestionID = q.ID
			if err := tx.Create(&opts[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
