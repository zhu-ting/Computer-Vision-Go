package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tingzhu/cv-review/backend/internal/model"
	"github.com/tingzhu/cv-review/backend/internal/repository"
	"gorm.io/gorm"
)

// ── Sentinel errors ──────────────────────────────────────────────

var (
	ErrModuleNotFound        = errors.New("module not found")
	ErrModuleExists          = errors.New("module with this name already exists")
	ErrQuestionGroupNotFound = errors.New("question group not found")
	ErrInvalidOptions        = errors.New("options must have at least 2 entries and exactly one correct answer")
)

// ── DTOs ─────────────────────────────────────────────────────────

// ModuleResponse is the public representation of a module.
type ModuleResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QuestionGroupResponse is a question group summary with question count.
type QuestionGroupResponse struct {
	ID            uint   `json:"id"`
	ModuleID      uint   `json:"module_id"`
	Title         string `json:"title"`
	Topic         string `json:"topic"`
	Difficulty    string `json:"difficulty"`
	QuestionCount int64  `json:"question_count"`
}

// OptionAdminResponse includes is_correct — for admin use only.
type OptionAdminResponse struct {
	ID        uint   `json:"id"`
	Content   string `json:"content"`
	IsCorrect bool   `json:"is_correct"`
	SortOrder int    `json:"sort_order"`
}

// QuestionAdminResponse includes analysis and correct answers —
// for admin use only, not for active exams.
type QuestionAdminResponse struct {
	ID       uint                  `json:"id"`
	GroupID  uint                  `json:"group_id"`
	Content  string                `json:"content"`
	Analysis string                `json:"analysis"`
	Version  int                   `json:"version"`
	Options  []OptionAdminResponse `json:"options"`
}

// OptionInput is used when creating a question with options.
type OptionInput struct {
	Content   string `json:"content" binding:"required"`
	IsCorrect bool   `json:"is_correct"`
	SortOrder int    `json:"sort_order"`
}

// ── Module service ───────────────────────────────────────────────

// ListModules returns all modules alphabetically.
func ListModules() ([]ModuleResponse, error) {
	modules, err := repository.ListModules()
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}
	responses := make([]ModuleResponse, 0, len(modules))
	for _, m := range modules {
		responses = append(responses, ModuleResponse{
			ID:        m.ID,
			Name:      m.Name,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}
	return responses, nil
}

// CreateModule creates a new module with the given name.
// Returns ErrModuleExists if a module with the same name already exists.
func CreateModule(name string) (*ModuleResponse, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("module name must not be empty")
	}

	// Duplicate check
	if _, err := repository.FindModuleByName(name); err == nil {
		return nil, ErrModuleExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("check module existence: %w", err)
	}

	m := &model.Module{Name: name}
	if err := repository.CreateModule(m); err != nil {
		return nil, fmt.Errorf("create module: %w", err)
	}

	return &ModuleResponse{
		ID:        m.ID,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}, nil
}

// ── Question group service ───────────────────────────────────────

// ListQuestionGroups returns question groups, with question counts.
// If moduleID is non-nil, filters to that module.
func ListQuestionGroups(moduleID *uint) ([]QuestionGroupResponse, error) {
	groups, err := repository.ListQuestionGroups(moduleID)
	if err != nil {
		return nil, fmt.Errorf("list question groups: %w", err)
	}

	counts, err := repository.CountQuestionsByGroup()
	if err != nil {
		return nil, fmt.Errorf("count questions: %w", err)
	}

	responses := make([]QuestionGroupResponse, 0, len(groups))
	for _, g := range groups {
		responses = append(responses, QuestionGroupResponse{
			ID:            g.ID,
			ModuleID:      g.ModuleID,
			Title:         g.Title,
			Topic:         g.Topic,
			Difficulty:    g.Difficulty,
			QuestionCount: counts[g.ID],
		})
	}
	return responses, nil
}

// CreateQuestionGroup creates a new question group under a module.
func CreateQuestionGroup(moduleID uint, title, topic, difficulty string) (*QuestionGroupResponse, error) {
	title = strings.TrimSpace(title)
	topic = strings.TrimSpace(topic)
	difficulty = strings.TrimSpace(difficulty)

	// Validate the module exists
	exists, err := repository.ModuleExists(moduleID)
	if err != nil {
		return nil, fmt.Errorf("check module: %w", err)
	}
	if !exists {
		return nil, ErrModuleNotFound
	}

	g := &model.QuestionGroup{
		ModuleID:   moduleID,
		Title:      title,
		Topic:      topic,
		Difficulty: difficulty,
	}
	if err := repository.CreateQuestionGroup(g); err != nil {
		return nil, fmt.Errorf("create question group: %w", err)
	}

	return &QuestionGroupResponse{
		ID:         g.ID,
		ModuleID:   g.ModuleID,
		Title:      g.Title,
		Topic:      g.Topic,
		Difficulty: g.Difficulty,
	}, nil
}

// ── Question service ────────────────────────────────────────────

// ListQuestions returns questions, optionally filtered by group or module.
func ListQuestions(groupID, moduleID *uint) ([]QuestionAdminResponse, error) {
	var questions []model.Question
	var err error

	if moduleID != nil {
		questions, err = repository.ListQuestionsByModule(*moduleID)
	} else {
		questions, err = repository.ListQuestions(groupID)
	}
	if err != nil {
		return nil, fmt.Errorf("list questions: %w", err)
	}

	responses := make([]QuestionAdminResponse, 0, len(questions))
	for i := range questions {
		responses = append(responses, toQuestionAdminResponse(&questions[i]))
	}
	return responses, nil
}

// CreateQuestion creates a question with options under a group or module.
// When moduleID is provided, a default group is auto-created (or reused) for that module.
// Validates: group exists, at least 2 options, exactly 1 correct.
func CreateQuestion(groupID, moduleID *uint, content, analysis string, options []OptionInput) (*QuestionAdminResponse, error) {
	content = strings.TrimSpace(content)
	analysis = strings.TrimSpace(analysis)

	// Resolve group ID from module ID if needed
	var resolvedGroupID uint
	if moduleID != nil {
		// Validate the module exists
		exists, err := repository.ModuleExists(*moduleID)
		if err != nil {
			return nil, fmt.Errorf("check module: %w", err)
		}
		if !exists {
			return nil, ErrModuleNotFound
		}

		group, err := repository.FindOrCreateDefaultGroup(*moduleID)
		if err != nil {
			return nil, fmt.Errorf("find or create default group: %w", err)
		}
		resolvedGroupID = group.ID
	} else if groupID != nil {
		resolvedGroupID = *groupID
	}

	// Validate group exists
	exists, err := repository.QuestionGroupExists(resolvedGroupID)
	if err != nil {
		return nil, fmt.Errorf("check question group: %w", err)
	}
	if !exists {
		return nil, ErrQuestionGroupNotFound
	}

	// Validate options
	if len(options) < 2 {
		return nil, ErrInvalidOptions
	}
	correctCount := 0
	for _, o := range options {
		if o.IsCorrect {
			correctCount++
		}
	}
	if correctCount != 1 {
		return nil, ErrInvalidOptions
	}

	// Determine next version
	version, err := repository.MaxQuestionVersion(resolvedGroupID)
	if err != nil {
		return nil, fmt.Errorf("get max version: %w", err)
	}
	version++

	// Build model objects
	q := &model.Question{
		GroupID:  resolvedGroupID,
		Content:  content,
		Analysis: analysis,
		Version:  version,
	}

	opts := make([]model.Option, 0, len(options))
	for i, in := range options {
		sortOrder := in.SortOrder
		if sortOrder == 0 {
			sortOrder = i + 1
		}
		opts = append(opts, model.Option{
			Content:   strings.TrimSpace(in.Content),
			IsCorrect: in.IsCorrect,
			SortOrder: sortOrder,
		})
	}

	if err := repository.CreateQuestionWithOptions(q, opts); err != nil {
		return nil, fmt.Errorf("create question: %w", err)
	}

	// Re-fetch to get IDs and timestamps
	questions, err := repository.ListQuestions(&resolvedGroupID)
	if err != nil {
		return nil, fmt.Errorf("re-fetch question: %w", err)
	}
	for i := range questions {
		if questions[i].ID == q.ID {
			r := toQuestionAdminResponse(&questions[i])
			return &r, nil
		}
	}

	return nil, fmt.Errorf("created question not found after insert")
}

// DeleteQuestion deletes a question and its options by ID.
func DeleteQuestion(id uint) error {
	if err := repository.DeleteQuestion(id); err != nil {
		return fmt.Errorf("delete question: %w", err)
	}
	return nil
}

// ── Helpers ──────────────────────────────────────────────────────

func toQuestionAdminResponse(q *model.Question) QuestionAdminResponse {
	opts := make([]OptionAdminResponse, 0, len(q.Options))
	for _, o := range q.Options {
		opts = append(opts, OptionAdminResponse{
			ID:        o.ID,
			Content:   o.Content,
			IsCorrect: o.IsCorrect,
			SortOrder: o.SortOrder,
		})
	}
	return QuestionAdminResponse{
		ID:       q.ID,
		GroupID:  q.GroupID,
		Content:  q.Content,
		Analysis: q.Analysis,
		Version:  q.Version,
		Options:  opts,
	}
}
