package database

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/tingzhu/cv-review/backend/internal/model"
	"gorm.io/gorm"
)

// seedData mirrors the structure of backend/data/seed_data.json.
type seedData struct {
	QuestionGroups []seedGroup `json:"question_groups"`
}

type seedGroup struct {
	Title      string        `json:"title"`
	Topic      string        `json:"topic"`
	Difficulty string        `json:"difficulty"`
	Questions  []seedQuestion `json:"questions"`
}

type seedQuestion struct {
	Content  string       `json:"content"`
	Analysis string       `json:"analysis"`
	Version  int          `json:"version"`
	Options  []seedOption `json:"options"`
}

type seedOption struct {
	Content   string `json:"content"`
	IsCorrect bool   `json:"is_correct"`
	SortOrder int    `json:"sort_order"`
}

// SeedIfEmpty checks whether the database has any questions, and if not,
// reads seed_data.json and populates the initial question bank.
//
// It is idempotent: running it multiple times (e.g., across server restarts)
// will not duplicate data because it skips seeding when questions already exist.
func SeedIfEmpty() {
	var count int64
	if err := DB.Model(&model.Question{}).Count(&count).Error; err != nil {
		log.Printf("[seed] failed to check question count: %v — skipping seed", err)
		return
	}
	if count > 0 {
		log.Printf("[seed] database already has %d questions — skipping seed", count)
		return
	}

	seedPath := resolveSeedPath()
	log.Printf("[seed] loading seed data from %s", seedPath)

	data, err := os.ReadFile(seedPath)
	if err != nil {
		log.Printf("[seed] failed to read seed file: %v", err)
		return
	}

	var seed seedData
	if err := json.Unmarshal(data, &seed); err != nil {
		log.Printf("[seed] failed to parse seed file: %v", err)
		return
	}

	inserted := insertSeedData(seed)
	log.Printf("[seed] done: %d groups, %d questions, %d options inserted",
		len(seed.QuestionGroups), inserted.questions, inserted.options)
}

// resolveSeedPath returns the absolute path to seed_data.json.
// It tries several locations so it works both in local development
// (go run ./cmd/server) and inside the Docker container.
func resolveSeedPath() string {
	// When running inside a Docker container, the binary and data
	// are placed in the same /app directory.
	if _, err := os.Stat("/app/data/seed_data.json"); err == nil {
		return "/app/data/seed_data.json"
	}

	// In local development, resolve relative to this source file.
	// runtime.Caller(0) gives us the path to seed.go, so we walk up
	// two levels to reach the project root (backend/).
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "backend/data/seed_data.json"
	}
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	return filepath.Join(root, "data", "seed_data.json")
}

type seedCounts struct {
	questions int
	options   int
}

// insertSeedData writes seed data into the database inside a transaction,
// so a partial failure rolls back cleanly and leaves no orphaned rows.
func insertSeedData(seed seedData) seedCounts {
	var counts seedCounts

	err := DB.Transaction(func(tx *gorm.DB) error {
		for _, g := range seed.QuestionGroups {
			group := model.QuestionGroup{
				Title:      g.Title,
				Topic:      g.Topic,
				Difficulty: g.Difficulty,
			}
			if err := tx.Create(&group).Error; err != nil {
				return err
			}

			for _, q := range g.Questions {
				question := model.Question{
					GroupID:  group.ID,
					Content:  q.Content,
					Analysis: q.Analysis,
					Version:  q.Version,
				}
				if err := tx.Create(&question).Error; err != nil {
					return err
				}
				counts.questions++

				for _, o := range q.Options {
					option := model.Option{
						QuestionID: question.ID,
						Content:    o.Content,
						IsCorrect:  o.IsCorrect,
						SortOrder:  o.SortOrder,
					}
					if err := tx.Create(&option).Error; err != nil {
						return err
					}
					counts.options++
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("[seed] transaction failed, rolled back: %v", err)
	}
	return counts
}
