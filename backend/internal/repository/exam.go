package repository

import (
	"time"

	"github.com/tingzhu/cv-review/backend/internal/database"
	"github.com/tingzhu/cv-review/backend/internal/model"
	"gorm.io/gorm/clause"
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

// ─────────────────────────────────────────────────────────────────
// Progress saving & submission
// ─────────────────────────────────────────────────────────────────

// UpsertUserAnswers inserts or updates user answers in batch.
// GORM's OnConflict clause handles the upsert: if a row with the
// same exam_question_id already exists, it updates selected_option_id.
// This lets the frontend call the save endpoint repeatedly without
// worrying about whether it's a create or an update.
func UpsertUserAnswers(answers []model.UserAnswer) error {
	if len(answers) == 0 {
		return nil
	}
	return database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "exam_question_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"selected_option_id", "updated_at"}),
	}).Create(&answers).Error
}

// GetExamByID fetches an exam by its primary key.
// Returns gorm.ErrRecordNotFound if the exam does not exist.
func GetExamByID(id uint) (*model.Exam, error) {
	var exam model.Exam
	err := database.DB.First(&exam, id).Error
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

// GetExamQuestionsForGrading loads all exam_questions for an exam,
// eagerly fetching the nested data needed to compute a score:
// question → options, and the user's answer.
//
// Preload nesting: ExamQuestion.Question.Options loads two levels deep.
func GetExamQuestionsForGrading(examID uint) ([]model.ExamQuestion, error) {
	var eqs []model.ExamQuestion
	err := database.DB.
		Where("exam_id = ?", examID).
		Preload("Question.Options").
		Preload("UserAnswer").
		Find(&eqs).Error
	if err != nil {
		return nil, err
	}
	return eqs, nil
}

// SubmitExam marks an exam as submitted and records the score.
// It uses a struct update to zero-value-safe fields only.
func SubmitExam(examID uint, score float64) error {
	now := time.Now()
	return database.DB.Model(&model.Exam{}).Where("id = ?", examID).Updates(map[string]interface{}{
		"status":       model.ExamStatusSubmitted,
		"score":        score,
		"submitted_at": now,
	}).Error
}
