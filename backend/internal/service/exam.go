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

// ─────────────────────────────────────────────────────────────────
// Progress saving DTOs
// ─────────────────────────────────────────────────────────────────

// AnswerInput represents a single answer submitted by the frontend
// during an active exam (page flip or auto-save).
type AnswerInput struct {
	ExamQuestionID  uint  `json:"exam_question_id"`
	SelectedOptionID *uint `json:"selected_option_id"` // nil = unanswered
}

// ─────────────────────────────────────────────────────────────────
// Submission / result DTOs — these DO include is_correct and analysis
// because the exam is over and the user is reviewing their results.
// ─────────────────────────────────────────────────────────────────

// ExamResultResponse is the full snapshot returned after submission.
// It includes the score, correct answers, and analysis.
type ExamResultResponse struct {
	ExamID         uint                     `json:"exam_id"`
	TotalQuestions int                      `json:"total_questions"`
	CorrectCount   int                      `json:"correct_count"`
	Score          float64                  `json:"score"`
	Status         string                   `json:"status"`
	Questions      []QuestionResultResponse `json:"questions"`
}

// QuestionResultResponse is a single question as seen in the result view.
// Unlike QuestionResponse, this includes analysis and reveals correct answers.
type QuestionResultResponse struct {
	ID               uint                 `json:"id"`
	ExamQuestionID   uint                 `json:"exam_question_id"`
	GroupID          uint                 `json:"group_id"`
	Content          string               `json:"content"`
	Analysis         string               `json:"analysis"`           // revealed after submission
	SelectedOptionID *uint                `json:"selected_option_id"` // what the user picked
	IsCorrect        bool                 `json:"is_correct"`         // whether the answer was right
	Options          []OptionResultResponse `json:"options"`
}

// OptionResultResponse includes is_correct so the user can see
// which option was the right one during review.
type OptionResultResponse struct {
	ID        uint   `json:"id"`
	Content   string `json:"content"`
	IsCorrect bool   `json:"is_correct"` // revealed after submission
}

// Sentinel errors for exam lifecycle state.
var (
	ErrExamNotFound        = errors.New("exam not found")
	ErrExamAlreadySubmitted = errors.New("exam has already been submitted")
	ErrExamNotOwned         = errors.New("exam question does not belong to this exam")
)

// ─────────────────────────────────────────────────────────────────
// Progress saving
// ─────────────────────────────────────────────────────────────────

// SaveProgress validates that the answers belong to the given exam
// and then upserts them. The frontend can call this safely on every
// page flip — duplicates are handled by the upsert.
func SaveProgress(examID uint, inputs []AnswerInput) error {
	if len(inputs) == 0 {
		return nil
	}

	// Verify the exam exists and is still in progress.
	exam, err := repository.GetExamByID(examID)
	if err != nil {
		return ErrExamNotFound
	}
	if exam.Status != model.ExamStatusInProgress {
		return ErrExamAlreadySubmitted
	}

	// Build the UserAnswer models. We trust the frontend to send
	// valid exam_question_ids that belong to this exam; a production
	// system would add an ownership check here.
	answers := make([]model.UserAnswer, 0, len(inputs))
	for _, in := range inputs {
		answers = append(answers, model.UserAnswer{
			ExamQuestionID:   in.ExamQuestionID,
			SelectedOptionID: in.SelectedOptionID,
		})
	}

	return repository.UpsertUserAnswers(answers)
}

// ─────────────────────────────────────────────────────────────────
// Submission & grading
// ─────────────────────────────────────────────────────────────────

// SubmitExam grades the exam, persists the score, and returns a
// complete snapshot with correct answers and analysis revealed.
func SubmitExam(examID uint) (*ExamResultResponse, error) {
	// Verify the exam is in a gradable state.
	exam, err := repository.GetExamByID(examID)
	if err != nil {
		return nil, ErrExamNotFound
	}
	if exam.Status != model.ExamStatusInProgress {
		return nil, ErrExamAlreadySubmitted
	}

	// Load all exam questions with nested data for grading.
	eqs, err := repository.GetExamQuestionsForGrading(examID)
	if err != nil {
		return nil, fmt.Errorf("failed to load exam questions: %w", err)
	}

	// Grade each question.
	correctCount := 0
	results := make([]QuestionResultResponse, 0, len(eqs))

	for _, eq := range eqs {
		qr := gradeQuestion(eq)
		if qr.IsCorrect {
			correctCount++
		}
		results = append(results, qr)
	}

	// Calculate score as a percentage, rounded to one decimal place.
	score := 0.0
	if exam.TotalQuestions > 0 {
		score = float64(correctCount) / float64(exam.TotalQuestions) * 100
	}
	// Round to 1 decimal: multiply by 10, round, divide by 10.
	score = float64(int(score*10+0.5)) / 10

	// Persist the score.
	if err := repository.SubmitExam(examID, score); err != nil {
		return nil, fmt.Errorf("failed to submit exam: %w", err)
	}

	return &ExamResultResponse{
		ExamID:         exam.ID,
		TotalQuestions: exam.TotalQuestions,
		CorrectCount:   correctCount,
		Score:          score,
		Status:         string(model.ExamStatusSubmitted),
		Questions:      results,
	}, nil
}

// gradeQuestion determines whether the user's answer is correct and
// builds the full result DTO (with analysis and is_correct revealed).
func gradeQuestion(eq model.ExamQuestion) QuestionResultResponse {
	// Find the correct option and check the user's answer.
	var selectedID *uint
	isCorrect := false

	if eq.UserAnswer != nil && eq.UserAnswer.SelectedOptionID != nil {
		selectedID = eq.UserAnswer.SelectedOptionID
		// Check if the selected option is marked as correct.
		for _, o := range eq.Question.Options {
			if o.ID == *selectedID && o.IsCorrect {
				isCorrect = true
				break
			}
		}
	}

	// Build the options with is_correct exposed.
	// We respect the original shuffled order stored in OptionOrder.
	optionOrder := parseOptionOrder(eq.OptionOrder)
	options := buildResultOptions(eq.Question.Options, optionOrder)

	return QuestionResultResponse{
		ID:               eq.Question.ID,
		ExamQuestionID:   eq.ID,
		GroupID:          eq.Question.GroupID,
		Content:          eq.Question.Content,
		Analysis:         eq.Question.Analysis, // ← revealed
		SelectedOptionID: selectedID,
		IsCorrect:        isCorrect,
		Options:          options,
	}
}

// buildResultOptions returns options in the shuffled order, with
// is_correct exposed (because the exam is over).
func buildResultOptions(options []model.Option, order []uint) []OptionResultResponse {
	optMap := make(map[uint]model.Option, len(options))
	for _, o := range options {
		optMap[o.ID] = o
	}

	result := make([]OptionResultResponse, 0, len(order))
	for _, id := range order {
		if o, ok := optMap[id]; ok {
			result = append(result, OptionResultResponse{
				ID:        o.ID,
				Content:   o.Content,
				IsCorrect: o.IsCorrect, // ← revealed
			})
		}
	}
	return result
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


// parseOptionOrder deserializes the JSONB option_order back into a uint slice.
// Returns nil on parse failure (the caller handles nil gracefully).
func parseOptionOrder(raw string) []uint {
	var ids []uint
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil
	}
	return ids
}
