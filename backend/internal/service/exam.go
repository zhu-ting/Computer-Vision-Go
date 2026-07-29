package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/tingzhu/cv-review/backend/internal/model"
	"github.com/tingzhu/cv-review/backend/internal/repository"
	"github.com/tingzhu/cv-review/backend/pkg/shuffle"
)

// ErrNoQuestions is returned when the database has no questions to generate
// an exam. The handler maps this to HTTP 503.
var ErrNoQuestions = errors.New("no questions available in the database")

// ─────────────────────────────────────────────────────────────────
// DTOs — these are what the frontend receives.
// They intentionally omit is_correct and analysis.
// ─────────────────────────────────────────────────────────────────

// ExamResponse is the safe-for-frontend view of a newly created exam.
type ExamResponse struct {
	ExamID         uint               `json:"exam_id"`
	TotalQuestions int                `json:"total_questions"`
	Questions      []QuestionResponse `json:"questions"`
}

// QuestionResponse represents a question as seen during an active exam.
// It deliberately excludes is_correct (on options) and analysis (on question).
type QuestionResponse struct {
	ID             uint             `json:"id"`
	ExamQuestionID uint             `json:"exam_question_id"`
	GroupID        uint             `json:"group_id"`
	Content        string           `json:"content"`
	Options        []OptionResponse `json:"options"`
}

// OptionResponse is a single answer choice. Only id and content are
// exposed — is_correct is stripped here, not hidden by the frontend.
type OptionResponse struct {
	ID      uint   `json:"id"`
	Content string `json:"content"`
}

// ─────────────────────────────────────────────────────────────────
// Service logic
// ─────────────────────────────────────────────────────────────────

// GenerateExam creates a new exam session with randomly selected questions
// and shuffled option order. It returns a response that is safe to send
// to the frontend (no correct-answer data leaked).
//
// The shuffle result is persisted in ExamQuestion.OptionOrder so that
// the same order is preserved across page saves, page loads, and grading.
func GenerateExam(questionCount int) (*ExamResponse, error) {
	// ── Step 1: Pick random questions ──────────────────────────
	questions, err := repository.GetRandomQuestions(questionCount)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch questions: %w", err)
	}
	if len(questions) == 0 {
		return nil, ErrNoQuestions
	}

	// ── Step 2: Prepare exam record ────────────────────────────
	exam := &model.Exam{
		TotalQuestions: questionCount,
		Status:         model.ExamStatusInProgress,
		StartedAt:      time.Now(),
	}

	// ── Step 3: Shuffle options & build snapshots ──────────────
	examQuestions := make([]model.ExamQuestion, 0, len(questions))
	responses := make([]QuestionResponse, 0, len(questions))

	for _, q := range questions {
		// Collect option IDs and shuffle them.
		// We shuffle IDs, not the full Option structs, because the
		// ExamQuestion stores the order as a JSON array of IDs.
		optionIDs := make([]uint, len(q.Options))
		for i, o := range q.Options {
			optionIDs[i] = o.ID
		}
		shuffledIDs := shuffle.Options(optionIDs)

		// Serialize the shuffled order into JSON for storage.
		orderJSON, err := json.Marshal(shuffledIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal option order: %w", err)
		}

		examQuestion := model.ExamQuestion{
			QuestionID:  q.ID, // pinned to this specific version
			OptionOrder: string(orderJSON),
		}
		examQuestions = append(examQuestions, examQuestion)

		// Build the frontend-safe response.
		// We reorder the options to match the shuffled IDs,
		// and strip is_correct from every option.
		qResponse := buildQuestionResponse(q, examQuestion, shuffledIDs)
		responses = append(responses, qResponse)
	}

	// ── Step 4: Persist in a transaction ───────────────────────
	if err := repository.CreateExamWithQuestions(exam, examQuestions); err != nil {
		return nil, fmt.Errorf("failed to create exam: %w", err)
	}

	// Assign the ExamQuestion IDs (set during insertion) to the response.
	for i := range responses {
		responses[i].ExamQuestionID = examQuestions[i].ID
	}

	return &ExamResponse{
		ExamID:         exam.ID,
		TotalQuestions: questionCount,
		Questions:      responses,
	}, nil
}

// buildQuestionResponse constructs a QuestionResponse where the options
// appear in the shuffled order and every is_correct field is omitted.
func buildQuestionResponse(q model.Question, eq model.ExamQuestion, shuffledIDs []uint) QuestionResponse {
	// Build a lookup map: optionID → Option
	optMap := make(map[uint]model.Option, len(q.Options))
	for _, o := range q.Options {
		optMap[o.ID] = o
	}

	// Emit options in shuffled order, without IsCorrect.
	options := make([]OptionResponse, 0, len(shuffledIDs))
	for _, id := range shuffledIDs {
		if o, ok := optMap[id]; ok {
			options = append(options, OptionResponse{
				ID:      o.ID,
				Content: o.Content,
				// is_correct intentionally NOT exposed
			})
		}
	}

	return QuestionResponse{
		ID:             q.ID,
		ExamQuestionID: eq.ID,
		GroupID:        q.GroupID,
		Content:        q.Content,
		// Analysis intentionally NOT exposed
		Options: options,
	}
}
