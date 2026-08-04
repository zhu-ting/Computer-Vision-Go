package repository

import (
	"errors"

	"github.com/tingzhu/cv-review/backend/internal/database"
	"github.com/tingzhu/cv-review/backend/internal/model"
	"gorm.io/gorm"
)

// ── Question queries ────────────────────────────────────────────

// ListQuestions returns a paginated slice of questions, newest first,
// preloading their Group and Options. It also returns the total count
// for pagination metadata.
func ListQuestions(offset, limit int) ([]model.Question, int64, error) {
	var questions []model.Question
	var total int64

	if err := database.DB.Model(&model.Question{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := database.DB.
		Preload("Group").
		Preload("Options").
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&questions).Error
	if err != nil {
		return nil, 0, err
	}

	return questions, total, nil
}

// GetQuestionByID fetches a single question with its Group and Options
// eagerly loaded. Returns gorm.ErrRecordNotFound if not found.
func GetQuestionByID(id uint) (*model.Question, error) {
	var question model.Question
	err := database.DB.
		Preload("Group").
		Preload("Options").
		First(&question, id).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

// GetMaxVersionForGroup returns the highest version number among all
// questions in a group. Returns 0 if the group has no questions.
func GetMaxVersionForGroup(groupID uint) (int, error) {
	var maxVersion int
	err := database.DB.
		Model(&model.Question{}).
		Where("group_id = ?", groupID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error
	if err != nil {
		return 0, err
	}
	return maxVersion, nil
}

// CountQuestionsByGroupID returns how many questions belong to a group.
func CountQuestionsByGroupID(groupID uint) (int64, error) {
	var count int64
	err := database.DB.
		Model(&model.Question{}).
		Where("group_id = ?", groupID).
		Count(&count).Error
	return count, err
}

// ── Writes ──────────────────────────────────────────────────────

// CreateQuestionWithOptions inserts a question and its options inside
// a transaction. The caller is responsible for setting question.GroupID
// and all option fields before calling this.
func CreateQuestionWithOptions(question *model.Question, options []model.Option) error {
	return database.DB.Transaction(func(tx *gormDB) error {
		if err := tx.Create(question).Error; err != nil {
			return err
		}
		for i := range options {
			options[i].QuestionID = question.ID
			if err := tx.Create(&options[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ── Deletes ─────────────────────────────────────────────────────

// DeleteQuestionByID removes a question and its options.
// Options are deleted explicitly to avoid relying on DB-level CASCADE.
// Returns gorm.ErrRecordNotFound if the question doesn't exist.
func DeleteQuestionByID(id uint) error {
	return database.DB.Transaction(func(tx *gormDB) error {
		// Delete options first (FK constraint).
		if err := tx.Where("question_id = ?", id).Delete(&model.Option{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&model.Question{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

// DeleteGroupIfEmpty removes a question group only if it has no
// remaining questions. Returns nil if the group still has questions
// (not an error — it's intentionally skipped).
func DeleteGroupIfEmpty(groupID uint) error {
	count, err := CountQuestionsByGroupID(groupID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // group still has questions, skip
	}

	result := database.DB.Delete(&model.QuestionGroup{}, groupID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// CreateGroup inserts a new question group.
func CreateGroup(group *model.QuestionGroup) error {
	return database.DB.Create(group).Error
}

// ── QuestionGroup queries ───────────────────────────────────────

// ListAllGroups returns every question group, ordered alphabetically
// by topic then title.
func ListAllGroups() ([]model.QuestionGroup, error) {
	var groups []model.QuestionGroup
	err := database.DB.
		Order("topic ASC, title ASC").
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// GetGroupByID fetches a single question group by ID.
func GetGroupByID(id uint) (*model.QuestionGroup, error) {
	var group model.QuestionGroup
	err := database.DB.First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// ── Sentinel errors ─────────────────────────────────────────────

var (
	ErrQuestionNotFound = errors.New("question not found")
	ErrGroupNotFound    = errors.New("question group not found")
)
