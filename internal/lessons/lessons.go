// Package lessons loads the static lesson content under
// data/lessons/python and syncs each lesson's intro challenge into the
// challenges table.
package lessons

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/fred-bonn/srlcode/internal/database"
)

type Lesson struct {
	ID           string
	Number       int
	Topics       []string
	Difficulty   int
	Content      string
	Assignment   string
	StartCode    string
	TestCode     string
	SolutionCode string
}

type metadata struct {
	Difficulty int      `json:"difficulty"`
	Topics     []string `json:"topics"`
}

func Load(dir string) ([]Lesson, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading lesson directory: %w", err)
	}
	var lessons []Lesson
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		lesson, err := loadLesson(filepath.Base(dir), filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("loading lesson %s: %w", entry.Name(), err)
		}
		lessons = append(lessons, lesson)
	}
	sort.Slice(lessons, func(i, j int) bool { return lessons[i].Number < lessons[j].Number })
	for i, lesson := range lessons {
		if lesson.Number != i+1 {
			return nil, fmt.Errorf("lesson numbers must be contiguous from 1, got %d at position %d", lesson.Number, i+1)
		}
	}
	return lessons, nil
}

func loadLesson(lang, dir string) (Lesson, error) {
	number, err := strconv.Atoi(strings.SplitN(filepath.Base(dir), "-", 2)[0])
	if err != nil {
		return Lesson{}, fmt.Errorf("parsing lesson number: %w", err)
	}
	var meta metadata
	if err := readFile(filepath.Join(dir, "metadata.json"), &meta); err != nil {
		return Lesson{}, err
	}
	if len(meta.Topics) == 0 {
		return Lesson{}, fmt.Errorf("no topics in metadata")
	}
	lesson := Lesson{
		ID:         lang + "-" + filepath.Base(dir),
		Number:     number,
		Topics:     meta.Topics,
		Difficulty: meta.Difficulty,
	}
	texts, err := readTexts(dir, "lesson.md", "assignment.md", "starter_code.py", "test_code.py", "solution_code.py")
	if err != nil {
		return Lesson{}, err
	}
	lesson.Content = texts["lesson.md"]
	lesson.Assignment = texts["assignment.md"]
	lesson.StartCode = texts["starter_code.py"]
	lesson.TestCode = texts["test_code.py"]
	lesson.SolutionCode = texts["solution_code.py"]
	return lesson, nil
}

func readFile(path string, v any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	return json.Unmarshal(content, v)
}

func readTexts(dir string, names ...string) (map[string]string, error) {
	texts := make(map[string]string)
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", name, err)
		}
		texts[name] = string(content)
	}
	return texts, nil
}

// Sync creates lessons that are missing from the database and refreshes
// the challenge content of existing ones.
func Sync(ctx context.Context, q database.Querier, lessons []Lesson) error {
	for _, lesson := range lessons {
		existing, err := q.GetLesson(ctx, lesson.ID)
		if errors.Is(err, sql.ErrNoRows) {
			if err := createLesson(ctx, q, lesson); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("looking up lesson %s: %w", lesson.ID, err)
		}
		if err := updateLesson(ctx, q, existing.ChallengeID, lesson); err != nil {
			return err
		}
	}
	return nil
}

func createLesson(ctx context.Context, q database.Querier, lesson Lesson) error {
	challenge, err := q.CreateChallenge(ctx, database.CreateChallengeParams{
		Assignment:   lesson.Assignment,
		StartCode:    lesson.StartCode,
		TestCode:     lesson.TestCode,
		SolutionCode: lesson.SolutionCode,
		Difficulty:   int32(lesson.Difficulty),
	})
	if err != nil {
		return fmt.Errorf("creating challenge for lesson %s: %w", lesson.ID, err)
	}
	topic, err := q.CreateTopic(ctx, lesson.Topics[0])
	if err != nil {
		return fmt.Errorf("creating topic %s: %w", lesson.Topics[0], err)
	}
	if err := q.CreateLesson(ctx, database.CreateLessonParams{
		ID:           lesson.ID,
		LessonNumber: int32(lesson.Number),
		TopicID:      topic.ID,
		ChallengeID:  challenge.ID,
	}); err != nil {
		return fmt.Errorf("creating lesson %s: %w", lesson.ID, err)
	}
	return linkTopics(ctx, q, challenge.ID, lesson.Topics)
}

func updateLesson(ctx context.Context, q database.Querier, challengeID uuid.UUID, lesson Lesson) error {
	if err := q.UpdateChallengeContent(ctx, database.UpdateChallengeContentParams{
		ID:           challengeID,
		Assignment:   lesson.Assignment,
		StartCode:    lesson.StartCode,
		TestCode:     lesson.TestCode,
		SolutionCode: lesson.SolutionCode,
		Difficulty:   int32(lesson.Difficulty),
	}); err != nil {
		return fmt.Errorf("updating challenge for lesson %s: %w", lesson.ID, err)
	}
	return linkTopics(ctx, q, challengeID, lesson.Topics)
}

func linkTopics(ctx context.Context, q database.Querier, challengeID uuid.UUID, names []string) error {
	for _, name := range names {
		topic, err := q.CreateTopic(ctx, name)
		if err != nil {
			return fmt.Errorf("creating topic %s: %w", name, err)
		}
		if err := q.EnsureChallengeTopic(ctx, database.EnsureChallengeTopicParams{
			ChallengeID: challengeID,
			TopicID:     topic.ID,
		}); err != nil {
			return fmt.Errorf("linking challenge to topic %s: %w", name, err)
		}
	}
	return nil
}
