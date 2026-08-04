package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/tingzhu/cv-review/backend/internal/model"
	"github.com/tingzhu/cv-review/backend/internal/repository"
	"gorm.io/gorm"
)

// ── Sentinel errors ──────────────────────────────────────────────

var (
	ErrQuestionNotFound = errors.New("question not found")
	ErrGroupNotFound    = errors.New("question group not found")
	ErrNoCorrectOption  = errors.New("at least one option must be marked as correct")
	ErrNoOptions        = errors.New("at least one option is required")
)

// ── DTOs ─────────────────────────────────────────────────────────

// QuestionDetailResponse is a single question with all fields exposed.
// This is the admin/management view — is_correct IS included.
type QuestionDetailResponse struct {
	ID        uint                 `json:"id"`
	GroupID   uint                 `json:"group_id"`
	Group     *GroupResponse       `json:"group,omitempty"`
	Content   string               `json:"content"`
	Analysis  string               `json:"analysis"`
	Version   int                  `json:"version"`
	CreatedAt time.Time            `json:"created_at"`
	Options   []OptionDetailResponse `json:"options"`
}

// OptionDetailResponse is an option with is_correct exposed (admin view).
type OptionDetailResponse struct {
	ID         uint   `json:"id"`
	QuestionID uint   `json:"question_id"`
	Content    string `json:"content"`
	IsCorrect  bool   `json:"is_correct"`
	SortOrder  int    `json:"sort_order"`
}

// QuestionListResponse is a paginated list of questions with metadata.
type QuestionListResponse struct {
	Questions  []QuestionDetailResponse `json:"questions"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

// GroupResponse is a lightweight question-group summary.
type GroupResponse struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	Topic      string `json:"topic"`
	Difficulty string `json:"difficulty"`
}

// ── Input DTOs ───────────────────────────────────────────────────

// CreateQuestionInput is the payload for creating a new question.
// Either GroupID (existing group) OR group metadata (new group) must be provided.
type CreateQuestionInput struct {
	GroupID         uint                `json:"group_id"`
	GroupTitle      string              `json:"group_title"`
	GroupTopic      string              `json:"group_topic"`
	GroupDifficulty string              `json:"group_difficulty"`
	Content         string              `json:"content"`
	Analysis        string              `json:"analysis"`
	Options         []CreateOptionInput `json:"options"`
}

// CreateOptionInput is a single option within a create request.
type CreateOptionInput struct {
	Content   string `json:"content"`
	IsCorrect bool   `json:"is_correct"`
	SortOrder int    `json:"sort_order"`
}

// UpdateQuestionInput is the payload for updating a question.
// Group fields are omitted — updates always stay in the same group
// and create a new version.
type UpdateQuestionInput struct {
	Content  string              `json:"content"`
	Analysis string              `json:"analysis"`
	Options  []CreateOptionInput `json:"options"`
}

// ── Service functions ────────────────────────────────────────────

// ListQuestions returns a paginated list of all questions, newest first.
func ListQuestions(page, pageSize int) (*QuestionListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	questions, total, err := repository.ListQuestions(offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list questions: %w", err)
	}

	responses := make([]QuestionDetailResponse, 0, len(questions))
	for _, q := range questions {
		responses = append(responses, toQuestionDetailResponse(&q))
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &QuestionListResponse{
		Questions:  responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetQuestion returns a single question by ID with all details.
func GetQuestion(id uint) (*QuestionDetailResponse, error) {
	q, err := repository.GetQuestionByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("failed to get question: %w", err)
	}
	resp := toQuestionDetailResponse(q)
	return &resp, nil
}

// CreateQuestion creates a new question (and optionally a new group).
// It validates that at least one correct option exists and that options
// are non-empty. The first question in a new group starts at version 1.
func CreateQuestion(input CreateQuestionInput) (*QuestionDetailResponse, error) {
	// ── Validate ──────────────────────────────────────────────
	if err := validateOptions(input.Options); err != nil {
		return nil, err
	}

	// ── Resolve or create group ───────────────────────────────
	var groupID uint

	if input.GroupID != 0 {
		// Use existing group.
		group, err := repository.GetGroupByID(input.GroupID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrGroupNotFound
			}
			return nil, fmt.Errorf("failed to fetch group: %w", err)
		}
		groupID = group.ID
	} else if input.GroupTitle != "" {
		// Create a new group.
		group, err := createGroup(input.GroupTitle, input.GroupTopic, input.GroupDifficulty)
		if err != nil {
			return nil, err
		}
		groupID = group.ID
	} else {
		return nil, fmt.Errorf("either group_id or group_title must be provided")
	}

	// ── Build question ────────────────────────────────────────
	question := model.Question{
		GroupID:  groupID,
		Content:  input.Content,
		Analysis: input.Analysis,
		Version:  1,
	}

	options := make([]model.Option, 0, len(input.Options))
	for _, o := range input.Options {
		options = append(options, model.Option{
			Content:   o.Content,
			IsCorrect: o.IsCorrect,
			SortOrder: o.SortOrder,
		})
	}

	// ── Persist ───────────────────────────────────────────────
	if err := repository.CreateQuestionWithOptions(&question, options); err != nil {
		return nil, fmt.Errorf("failed to create question: %w", err)
	}

	// Re-fetch to get populated fields (Group, Options with IDs).
	created, err := repository.GetQuestionByID(question.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created question: %w", err)
	}
	resp := toQuestionDetailResponse(created)
	return &resp, nil
}

// UpdateQuestion creates a NEW version of the question in the same group.
// The old question row is preserved so that existing exam snapshots
// still reference the original version.
func UpdateQuestion(id uint, input UpdateQuestionInput) (*QuestionDetailResponse, error) {
	// ── Validate ──────────────────────────────────────────────
	if err := validateOptions(input.Options); err != nil {
		return nil, err
	}

	// ── Fetch the existing question ───────────────────────────
	existing, err := repository.GetQuestionByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("failed to fetch question: %w", err)
	}

	// ── Compute next version ──────────────────────────────────
	maxVersion, err := repository.GetMaxVersionForGroup(existing.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to compute version: %w", err)
	}
	nextVersion := maxVersion + 1

	// ── Build new version ─────────────────────────────────────
	question := model.Question{
		GroupID:  existing.GroupID,
		Content:  input.Content,
		Analysis: input.Analysis,
		Version:  nextVersion,
	}

	options := make([]model.Option, 0, len(input.Options))
	for _, o := range input.Options {
		options = append(options, model.Option{
			Content:   o.Content,
			IsCorrect: o.IsCorrect,
			SortOrder: o.SortOrder,
		})
	}

	// ── Persist ───────────────────────────────────────────────
	if err := repository.CreateQuestionWithOptions(&question, options); err != nil {
		return nil, fmt.Errorf("failed to create new question version: %w", err)
	}

	created, err := repository.GetQuestionByID(question.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated question: %w", err)
	}
	resp := toQuestionDetailResponse(created)
	return &resp, nil
}

// DeleteQuestion deletes a question and its options. If the question
// was the last one in its group, the group is also cleaned up.
func DeleteQuestion(id uint) error {
	// Fetch first to get the GroupID for cleanup.
	q, err := repository.GetQuestionByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrQuestionNotFound
		}
		return fmt.Errorf("failed to fetch question: %w", err)
	}
	groupID := q.GroupID

	if err := repository.DeleteQuestionByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrQuestionNotFound
		}
		return fmt.Errorf("failed to delete question: %w", err)
	}

	// Clean up the group if it's now empty.
	if err := repository.DeleteGroupIfEmpty(groupID); err != nil {
		return fmt.Errorf("failed to clean up group: %w", err)
	}

	return nil
}

// ListGroups returns all question groups, ordered by topic then title.
func ListGroups() ([]GroupResponse, error) {
	groups, err := repository.ListAllGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	responses := make([]GroupResponse, 0, len(groups))
	for _, g := range groups {
		responses = append(responses, GroupResponse{
			ID:         g.ID,
			Title:      g.Title,
			Topic:      g.Topic,
			Difficulty: g.Difficulty,
		})
	}
	return responses, nil
}

// ── Private helpers ──────────────────────────────────────────────

// createGroup inserts a new QuestionGroup and returns it.
func createGroup(title, topic, difficulty string) (*model.QuestionGroup, error) {
	if topic == "" {
		return nil, fmt.Errorf("group_topic is required when creating a new group")
	}
	if difficulty == "" {
		difficulty = "medium"
	}

	group := &model.QuestionGroup{
		Title:      title,
		Topic:      topic,
		Difficulty: difficulty,
	}

	if err := repository.CreateGroup(group); err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}
	return group, nil
}

// validateOptions checks that options are non-empty and that exactly
// one option is marked as correct.
func validateOptions(opts []CreateOptionInput) error {
	if len(opts) == 0 {
		return ErrNoOptions
	}

	correctCount := 0
	for _, o := range opts {
		if o.IsCorrect {
			correctCount++
		}
	}
	if correctCount == 0 {
		return ErrNoCorrectOption
	}
	// Note: we allow multiple correct options — some question types
	// (e.g., "select all that apply") may need them.

	return nil
}

// toQuestionDetailResponse maps a model.Question (with preloaded
// Group and Options) to the admin-facing DTO.
func toQuestionDetailResponse(q *model.Question) QuestionDetailResponse {
	var groupResp *GroupResponse
	if q.Group.ID != 0 {
		groupResp = &GroupResponse{
			ID:         q.Group.ID,
			Title:      q.Group.Title,
			Topic:      q.Group.Topic,
			Difficulty: q.Group.Difficulty,
		}
	}

	opts := make([]OptionDetailResponse, 0, len(q.Options))
	for _, o := range q.Options {
		opts = append(opts, OptionDetailResponse{
			ID:         o.ID,
			QuestionID: o.QuestionID,
			Content:    o.Content,
			IsCorrect:  o.IsCorrect,
			SortOrder:  o.SortOrder,
		})
	}

	return QuestionDetailResponse{
		ID:        q.ID,
		GroupID:   q.GroupID,
		Group:     groupResp,
		Content:   q.Content,
		Analysis:  q.Analysis,
		Version:   q.Version,
		CreatedAt: q.CreatedAt,
		Options:   opts,
	}
}
