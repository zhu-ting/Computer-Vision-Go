// Package model holds the GORM model definitions for all database tables.
//
// Table relationships (simplified):
//
//	question_groups 1──N questions 1──N options
//	exams 1──N exam_questions (snapshot) N──1 questions
//	exam_questions 1──0..1 user_answers
//	question_groups 1──0..1 user_notes
//
// The distinction between question_groups and questions is intentional:
// when a question's content is updated, we insert a new row in questions
// (with version+1) rather than mutating the existing row. Exam snapshots
// (exam_questions) reference the specific question version that was active
// when the exam was taken.
package model
