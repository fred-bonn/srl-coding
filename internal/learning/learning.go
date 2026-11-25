// Package learning runs the daily loop: the current lesson, the queue of
// due topics, and the attempt lifecycle. Test users live on a per-user
// clock offset that jumps a day at each breakpoint; real users wait for
// the calendar. All timestamps are recorded in user time.
package learning

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/fred-bonn/srlcode/internal/database"
	"github.com/fred-bonn/srlcode/internal/lessons"
	"github.com/fred-bonn/srlcode/internal/srs"
)

var (
	ErrNotAssigned            = errors.New("challenge is not part of today's work")
	ErrNotFound               = errors.New("not found")
	ErrAttemptClosed          = errors.New("attempt already completed")
	ErrLessonAlreadyDoneToday = errors.New("lesson already completed today")
	ErrCurriculumDone         = errors.New("all lessons completed")
)

type Service struct {
	db     database.Querier
	sched  srs.Scheduler
	isTest func(uuid.UUID) bool
	now    func() time.Time

	byID map[string]lessons.Lesson

	mu     sync.Mutex
	offset map[uuid.UUID]int
}

func New(db database.Querier, sched srs.Scheduler, ls []lessons.Lesson, isTest func(uuid.UUID) bool) *Service {
	byID := make(map[string]lessons.Lesson, len(ls))
	for _, lesson := range ls {
		byID[lesson.ID] = lesson
	}
	return &Service{db: db, sched: sched, isTest: isTest, now: time.Now, byID: byID, offset: map[uuid.UUID]int{}}
}

type LessonView struct {
	ID          string `json:"id"`
	Number      int    `json:"number"`
	Content     string `json:"content"`
	Assignment  string `json:"assignment"`
	ChallengeID string `json:"challenge_id"`
}

type QueueItem struct {
	TopicID string `json:"topic_id"`
	Name    string `json:"name"`
}

type AttemptView struct {
	ID          string `json:"id"`
	ChallengeID string `json:"challenge_id"`
	Assignment  string `json:"assignment"`
	Code        string `json:"code"`
}

type TodayView struct {
	Lesson       *LessonView   `json:"lesson"`
	Queue        []QueueItem   `json:"queue"`
	OpenAttempts []AttemptView `json:"open_attempts"`
	DoneForToday bool          `json:"done_for_today"`
}

func (s *Service) userNow(userID uuid.UUID) time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.now().Add(time.Duration(s.offset[userID]) * 24 * time.Hour)
}

func (s *Service) advanceDay(userID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.offset[userID]++
}

func dayStart(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return dayStart(t).AddDate(0, 0, 1)
}

func (s *Service) Today(ctx context.Context, userID uuid.UUID) (TodayView, error) {
	for {
		now := s.userNow(userID)
		view, work, curriculumDone, err := s.today(ctx, userID, now)
		if err != nil {
			return TodayView{}, err
		}
		view.DoneForToday = !work
		if work || !s.isTest(userID) || curriculumDone {
			return view, nil
		}
		// Breakpoint reached: the test user jumps to the next day.
		s.advanceDay(userID)
	}
}

func (s *Service) today(ctx context.Context, userID uuid.UUID, now time.Time) (TodayView, bool, bool, error) {
	lesson, lessonErr := s.db.GetCurrentLesson(ctx, userID)
	if lessonErr != nil && !errors.Is(lessonErr, sql.ErrNoRows) {
		return TodayView{}, false, false, fmt.Errorf("loading current lesson: %w", lessonErr)
	}
	curriculumDone := errors.Is(lessonErr, sql.ErrNoRows)
	completedToday, err := s.db.HasCompletedLessonToday(ctx, database.HasCompletedLessonTodayParams{
		UserID:      userID,
		CompletedAt: dayStart(now),
	})
	if err != nil {
		return TodayView{}, false, false, fmt.Errorf("checking lesson progress: %w", err)
	}

	view := TodayView{Queue: []QueueItem{}, OpenAttempts: []AttemptView{}}
	showLesson := lessonErr == nil && !completedToday
	if showLesson {
		content, ok := s.byID[lesson.ID]
		if !ok {
			return TodayView{}, false, false, fmt.Errorf("lesson %s missing from loaded lessons", lesson.ID)
		}
		if err := s.db.EnsureUserTopic(ctx, database.EnsureUserTopicParams{
			UserID:  userID,
			TopicID: lesson.TopicID,
			DueAt:   dayStart(now),
		}); err != nil {
			return TodayView{}, false, false, fmt.Errorf("introducing lesson topic: %w", err)
		}
		view.Lesson = &LessonView{
			ID:          content.ID,
			Number:      content.Number,
			Content:     content.Content,
			Assignment:  content.Assignment,
			ChallengeID: lesson.ChallengeID.String(),
		}
	}

	due, err := s.db.ListDueUserTopics(ctx, database.ListDueUserTopicsParams{
		UserID: userID,
		DueAt:  endOfDay(now),
	})
	if err != nil {
		return TodayView{}, false, false, fmt.Errorf("listing due topics: %w", err)
	}
	for _, topic := range due {
		view.Queue = append(view.Queue, QueueItem{TopicID: topic.ID.String(), Name: topic.Name})
	}

	open, err := s.db.ListOpenAttempts(ctx, userID)
	if err != nil {
		return TodayView{}, false, false, fmt.Errorf("listing open attempts: %w", err)
	}
	for _, attempt := range open {
		view.OpenAttempts = append(view.OpenAttempts, AttemptView{
			ID:          attempt.ID.String(),
			ChallengeID: attempt.ChallengeID.String(),
			Assignment:  attempt.Assignment,
			Code:        attempt.Code,
		})
	}

	work := view.Lesson != nil || len(view.Queue) > 0
	return view, work, curriculumDone, nil
}

func (s *Service) StartAttempt(ctx context.Context, userID, challengeID uuid.UUID) (AttemptView, error) {
	open, err := s.db.GetOpenAttemptByChallenge(ctx, database.GetOpenAttemptByChallengeParams{
		UserID:      userID,
		ChallengeID: challengeID,
	})
	if err == nil {
		return s.attemptView(ctx, open)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return AttemptView{}, fmt.Errorf("loading open attempt: %w", err)
	}

	now := s.userNow(userID)
	challenge, err := s.db.GetChallenge(ctx, challengeID)
	if errors.Is(err, sql.ErrNoRows) {
		return AttemptView{}, ErrNotFound
	}
	if err != nil {
		return AttemptView{}, fmt.Errorf("loading challenge: %w", err)
	}
	if err := s.checkAssigned(ctx, userID, challengeID, now); err != nil {
		return AttemptView{}, err
	}

	attempt, err := s.db.CreateAttempt(ctx, database.CreateAttemptParams{
		UserID:      userID,
		ChallengeID: challengeID,
		Code:        challenge.StartCode,
		StartedAt:   now,
	})
	if err != nil {
		return AttemptView{}, fmt.Errorf("creating attempt: %w", err)
	}
	return s.attemptView(ctx, attempt)
}

// checkAssigned rejects challenges that are neither the current lesson's
// nor linked to a topic that is due today.
func (s *Service) checkAssigned(ctx context.Context, userID, challengeID uuid.UUID, now time.Time) error {
	lesson, err := s.db.GetCurrentLesson(ctx, userID)
	if err == nil && lesson.ChallengeID == challengeID {
		return nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("loading current lesson: %w", err)
	}

	due, err := s.db.ListDueUserTopics(ctx, database.ListDueUserTopicsParams{
		UserID: userID,
		DueAt:  endOfDay(now),
	})
	if err != nil {
		return fmt.Errorf("listing due topics: %w", err)
	}
	dueIDs := make(map[uuid.UUID]struct{}, len(due))
	for _, topic := range due {
		dueIDs[topic.ID] = struct{}{}
	}
	topics, err := s.db.ListChallengeTopics(ctx, challengeID)
	if err != nil {
		return fmt.Errorf("listing challenge topics: %w", err)
	}
	for _, topicID := range topics {
		if _, ok := dueIDs[topicID]; ok {
			return nil
		}
	}
	return ErrNotAssigned
}

func (s *Service) attemptView(ctx context.Context, attempt database.UserReview) (AttemptView, error) {
	challenge, err := s.db.GetChallenge(ctx, attempt.ChallengeID)
	if err != nil {
		return AttemptView{}, fmt.Errorf("loading challenge: %w", err)
	}
	return AttemptView{
		ID:          attempt.ID.String(),
		ChallengeID: attempt.ChallengeID.String(),
		Assignment:  challenge.Assignment,
		Code:        attempt.Code,
	}, nil
}

func (s *Service) SaveCode(ctx context.Context, userID, attemptID uuid.UUID, code string) error {
	attempt, err := s.db.GetAttempt(ctx, attemptID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && attempt.UserID != userID) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("loading attempt: %w", err)
	}
	rows, err := s.db.SaveAttemptCode(ctx, database.SaveAttemptCodeParams{ID: attemptID, Code: code})
	if err != nil {
		return fmt.Errorf("saving code: %w", err)
	}
	if rows == 0 {
		return ErrAttemptClosed
	}
	return nil
}

func (s *Service) CompleteAttempt(ctx context.Context, userID, attemptID uuid.UUID, correct bool, durationSeconds int32) error {
	attempt, err := s.db.GetAttempt(ctx, attemptID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && attempt.UserID != userID) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("loading attempt: %w", err)
	}

	now := s.userNow(userID)
	rows, err := s.db.CompleteAttempt(ctx, database.CompleteAttemptParams{
		ID:              attemptID,
		Correct:         sql.NullBool{Bool: correct, Valid: true},
		DurationSeconds: sql.NullInt32{Int32: durationSeconds, Valid: true},
		CompletedAt:     sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("completing attempt: %w", err)
	}
	if rows == 0 {
		return ErrAttemptClosed
	}
	return s.scheduleTopics(ctx, userID, attempt.ChallengeID, now)
}

func (s *Service) scheduleTopics(ctx context.Context, userID, challengeID uuid.UUID, now time.Time) error {
	topicIDs, err := s.db.ListChallengeTopics(ctx, challengeID)
	if err != nil {
		return fmt.Errorf("listing challenge topics: %w", err)
	}
	for _, topicID := range topicIDs {
		rows, err := s.db.ListCompletedTopicReviews(ctx, database.ListCompletedTopicReviewsParams{
			UserID:  userID,
			TopicID: topicID,
		})
		if err != nil {
			return fmt.Errorf("listing review history: %w", err)
		}
		history := make([]srs.Review, 0, len(rows))
		for _, row := range rows {
			history = append(history, srs.Review{Correct: row.Correct.Bool, ReviewedAt: row.CompletedAt.Time})
		}
		if err := s.db.SetUserTopicDue(ctx, database.SetUserTopicDueParams{
			UserID:  userID,
			TopicID: topicID,
			DueAt:   s.sched.NextDue(history, now),
		}); err != nil {
			return fmt.Errorf("recording due date: %w", err)
		}
	}
	return nil
}

func (s *Service) CompleteLesson(ctx context.Context, userID uuid.UUID) error {
	now := s.userNow(userID)
	completedToday, err := s.db.HasCompletedLessonToday(ctx, database.HasCompletedLessonTodayParams{
		UserID:      userID,
		CompletedAt: dayStart(now),
	})
	if err != nil {
		return fmt.Errorf("checking lesson progress: %w", err)
	}
	if completedToday {
		return ErrLessonAlreadyDoneToday
	}
	lesson, err := s.db.GetCurrentLesson(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCurriculumDone
	}
	if err != nil {
		return fmt.Errorf("loading current lesson: %w", err)
	}
	return s.db.CompleteLesson(ctx, database.CompleteLessonParams{
		UserID:      userID,
		LessonID:    lesson.ID,
		CompletedAt: now,
	})
}
