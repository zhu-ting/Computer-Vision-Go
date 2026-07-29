package repository

import (
	"github.com/tingzhu/cv-review/backend/internal/database"
	"github.com/tingzhu/cv-review/backend/internal/model"
)

// GetRandomQuestions fetches `count` questions at random, eagerly loading
// their Options. It excludes questions whose GroupID appears in excludeGroupIDs
// (not currently used, but useful for future "don't repeat" logic).
//
// PostgreSQL's RANDOM() ordering is fine for small-to-medium tables;
// for large-scale production, consider TABLESAMPLE or a dedicated
// random-selection strategy.
func GetRandomQuestions(count int) ([]model.Question, error) {
	var questions []model.Question
	err := database.DB.
		Preload("Options").                      // eager-load to avoid N+1
		Order("RANDOM()").                       // PostgreSQL random ordering
		Limit(count).
		Find(&questions).Error
	if err != nil {
		return nil, err
	}
	return questions, nil
}

// FindOptionsByQuestionID returns all options for a given question,
// ordered by their original sort_order. Used by the service layer
// to get the option IDs before shuffling.
func FindOptionsByQuestionID(questionID uint) ([]model.Option, error) {
	var options []model.Option
	err := database.DB.
		Where("question_id = ?", questionID).
		Order("sort_order ASC").
		Find(&options).Error
	if err != nil {
		return nil, err
	}
	return options, nil
}

// CreateExamWithQuestions inserts an Exam and its ExamQuestions in a single
// transaction. If any insertion fails, the entire operation rolls back.
func CreateExamWithQuestions(exam *model.Exam, examQuestions []model.ExamQuestion) error {
	return database.DB.Transaction(func(tx *gormDB) error {
		if err := tx.Create(exam).Error; err != nil {
			return err
		}
		for i := range examQuestions {
			examQuestions[i].ExamID = exam.ID
			if err := tx.Create(&examQuestions[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
