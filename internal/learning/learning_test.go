package learning

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/fred-bonn/srlcode/internal/database"
	"github.com/fred-bonn/srlcode/internal/lessons"
	"github.com/fred-bonn/srlcode/internal/srs"
	"github.com/fred-bonn/srlcode/internal/testdb"
)

type fixture struct {
	svc         *Service
	db          *testdb.DB
	user        database.User
	topicIDs    []uuid.UUID
	challengeOf map[uuid.UUID]uuid.UUID // topic → intro challenge
	clock       time.Time
}

func newFixture(t *testing.T, isTest bool) *fixture {
	t.Helper()
	db := testdb.New()
	ctx := context.Background()

	var ls []lessons.Lesson
	var topicIDs []uuid.UUID
	challengeOf := map[uuid.UUID]uuid.UUID{}
	for i := 1; i <= 3; i++ {
		digit := string(rune('0' + i))
		challenge, err := db.CreateChallenge(ctx, database.CreateChallengeParams{
			Assignment:   "assignment " + digit,
			StartCode:    "start " + digit,
			TestCode:     "test",
			SolutionCode: "solution",
			Difficulty:   1,
		})
		if err != nil {
			t.Fatalf("creating challenge: %v", err)
		}
		topic, err := db.CreateTopic(ctx, "topic"+digit)
		if err != nil {
			t.Fatalf("creating topic: %v", err)
		}
		topicIDs = append(topicIDs, topic.ID)
		challengeOf[topic.ID] = challenge.ID
		if err := db.EnsureChallengeTopic(ctx, database.EnsureChallengeTopicParams{
			ChallengeID: challenge.ID,
			TopicID:     topic.ID,
		}); err != nil {
			t.Fatalf("linking topic: %v", err)
		}
		id := "python-00" + digit + "-topic"
		if err := db.CreateLesson(ctx, database.CreateLessonParams{
			ID:           id,
			LessonNumber: int32(i),
			TopicID:      topic.ID,
			ChallengeID:  challenge.ID,
		}); err != nil {
			t.Fatalf("creating lesson: %v", err)
		}
		ls = append(ls, lessons.Lesson{ID: id, Number: i, Content: "content " + digit})
	}

	user, err := db.CreateUser(ctx, database.CreateUserParams{Email: "user@example.com", PasswordHash: "hash"})
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}

	f := &fixture{
		db:          db,
		user:        user,
		topicIDs:    topicIDs,
		challengeOf: challengeOf,
		clock:       time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC),
	}
	f.svc = New(db, srs.Fixed{}, ls, func(id uuid.UUID) bool { return isTest })
	f.svc.now = func() time.Time { return f.clock }
	return f
}

func (f *fixture) view(t *testing.T) TodayView {
	t.Helper()
	view, err := f.svc.Today(context.Background(), f.user.ID)
	if err != nil {
		t.Fatalf("Today: %v", err)
	}
	return view
}

// doDay finishes all of the day's work: the lesson plus every queued topic.
func (f *fixture) doDay(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	view := f.view(t)
	if view.Lesson != nil {
		if err := f.svc.CompleteLesson(ctx, f.user.ID); err != nil {
			t.Fatalf("completing lesson: %v", err)
		}
	}
	for len(view.Queue) > 0 {
		for _, item := range view.Queue {
			attempt, err := f.svc.StartAttempt(ctx, f.user.ID, f.challengeOf[uuid.MustParse(item.TopicID)])
			if err != nil {
				t.Fatalf("starting attempt for %s: %v", item.Name, err)
			}
			if err := f.svc.CompleteAttempt(ctx, f.user.ID, uuid.MustParse(attempt.ID), true, 60); err != nil {
				t.Fatalf("completing attempt for %s: %v", item.Name, err)
			}
		}
		view = f.view(t)
	}
}

func TestTodayIntroducesLessonTopic(t *testing.T) {
	f := newFixture(t, false)
	view := f.view(t)
	if view.Lesson == nil || view.Lesson.ID != "python-001-topic" {
		t.Fatalf("expected lesson 1, got %+v", view.Lesson)
	}
	if len(view.Queue) != 1 || view.Queue[0].TopicID != f.topicIDs[0].String() {
		t.Fatalf("expected lesson 1 topic in queue, got %+v", view.Queue)
	}
	if view.DoneForToday {
		t.Fatal("fresh user should have work")
	}
}

func TestRealUserDoneForToday(t *testing.T) {
	f := newFixture(t, false)
	f.doDay(t)

	view := f.view(t)
	if view.Lesson != nil {
		t.Fatalf("lesson should be done for today, got %+v", view.Lesson)
	}
	if len(view.Queue) != 0 {
		t.Fatalf("queue should be empty, got %+v", view.Queue)
	}
	if !view.DoneForToday {
		t.Fatal("expected done for today")
	}

	err := f.svc.CompleteLesson(context.Background(), f.user.ID)
	if !errors.Is(err, ErrLessonAlreadyDoneToday) {
		t.Fatalf("second lesson same day = %v, want ErrLessonAlreadyDoneToday", err)
	}
}

func TestRealUserNextDayNewLessonAndReviews(t *testing.T) {
	f := newFixture(t, false)
	f.doDay(t)
	f.clock = f.clock.AddDate(0, 0, 1)

	view := f.view(t)
	if view.Lesson == nil || view.Lesson.ID != "python-002-topic" {
		t.Fatalf("expected lesson 2 on the next day, got %+v", view.Lesson)
	}
	// The day-2 queue holds lesson 2's fresh topic and lesson 1's topic,
	// which is due for review after its one-day interval.
	queue := map[string]bool{}
	for _, item := range view.Queue {
		queue[item.TopicID] = true
	}
	if !queue[f.topicIDs[0].String()] || !queue[f.topicIDs[1].String()] {
		t.Fatalf("expected topics 1 and 2 in queue, got %+v", view.Queue)
	}
}

func TestTestUserJumpsBreakpoints(t *testing.T) {
	f := newFixture(t, true)

	f.doDay(t)
	view := f.view(t)
	if view.Lesson == nil || view.Lesson.ID != "python-002-topic" {
		t.Fatalf("test user should jump straight to lesson 2, got %+v", view.Lesson)
	}
	if view.DoneForToday {
		t.Fatal("jumped day should have work")
	}

	f.doDay(t)
	view = f.view(t)
	if view.Lesson == nil || view.Lesson.ID != "python-003-topic" {
		t.Fatalf("test user should jump straight to lesson 3, got %+v", view.Lesson)
	}
}

func TestTestUserCurriculumDone(t *testing.T) {
	f := newFixture(t, true)
	for i := 0; i < 3; i++ {
		f.doDay(t)
	}
	view := f.view(t)
	if view.Lesson != nil {
		t.Fatalf("curriculum should be done, got %+v", view.Lesson)
	}
	if !view.DoneForToday {
		t.Fatal("expected done for today after curriculum")
	}
}

func TestReviewSchedulingRecordsDueDate(t *testing.T) {
	f := newFixture(t, false)
	f.doDay(t)

	var due time.Time
	for _, ut := range f.db.UserTopics {
		if ut.TopicID == f.topicIDs[0] {
			due = ut.DueAt
		}
	}
	want := f.clock.Add(24 * time.Hour) // fixed ladder, streak of 1
	if !due.Equal(want) {
		t.Fatalf("topic due at %v, want %v", due, want)
	}
}

func TestStartAttemptIdempotent(t *testing.T) {
	f := newFixture(t, false)
	ctx := context.Background()
	view := f.view(t)
	challengeID := uuid.MustParse(view.Lesson.ChallengeID)

	first, err := f.svc.StartAttempt(ctx, f.user.ID, challengeID)
	if err != nil {
		t.Fatalf("first start: %v", err)
	}
	second, err := f.svc.StartAttempt(ctx, f.user.ID, challengeID)
	if err != nil {
		t.Fatalf("second start: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected same open attempt, got %v and %v", first.ID, second.ID)
	}
	if first.Code == "" {
		t.Fatal("attempt should start from the challenge's starter code")
	}
}

func TestStartAttemptNotAssigned(t *testing.T) {
	f := newFixture(t, false)
	ctx := context.Background()

	// A challenge that is neither the current lesson's nor linked to a
	// topic due today.
	challenge, err := f.db.CreateChallenge(ctx, database.CreateChallengeParams{
		Assignment: "other", StartCode: "x", TestCode: "x", SolutionCode: "x", Difficulty: 1,
	})
	if err != nil {
		t.Fatalf("creating challenge: %v", err)
	}

	_, err = f.svc.StartAttempt(ctx, f.user.ID, challenge.ID)
	if !errors.Is(err, ErrNotAssigned) {
		t.Fatalf("unassigned challenge = %v, want ErrNotAssigned", err)
	}
}

func TestSaveCodeAndResume(t *testing.T) {
	f := newFixture(t, false)
	ctx := context.Background()
	view := f.view(t)

	attempt, err := f.svc.StartAttempt(ctx, f.user.ID, uuid.MustParse(view.Lesson.ChallengeID))
	if err != nil {
		t.Fatalf("starting attempt: %v", err)
	}
	if err := f.svc.SaveCode(ctx, f.user.ID, uuid.MustParse(attempt.ID), "print('my solution')"); err != nil {
		t.Fatalf("saving code: %v", err)
	}

	resumed, err := f.svc.StartAttempt(ctx, f.user.ID, uuid.MustParse(view.Lesson.ChallengeID))
	if err != nil {
		t.Fatalf("resuming attempt: %v", err)
	}
	if resumed.Code != "print('my solution')" {
		t.Fatalf("resumed code = %q, want saved code", resumed.Code)
	}

	today := f.view(t)
	if len(today.OpenAttempts) != 1 || today.OpenAttempts[0].Code != "print('my solution')" {
		t.Fatalf("open attempts should show saved code, got %+v", today.OpenAttempts)
	}
}

func TestCompleteAttemptTwice(t *testing.T) {
	f := newFixture(t, false)
	ctx := context.Background()
	view := f.view(t)

	attempt, err := f.svc.StartAttempt(ctx, f.user.ID, uuid.MustParse(view.Lesson.ChallengeID))
	if err != nil {
		t.Fatalf("starting attempt: %v", err)
	}
	id := uuid.MustParse(attempt.ID)
	if err := f.svc.CompleteAttempt(ctx, f.user.ID, id, true, 60); err != nil {
		t.Fatalf("first complete: %v", err)
	}
	err = f.svc.CompleteAttempt(ctx, f.user.ID, id, true, 60)
	if !errors.Is(err, ErrAttemptClosed) {
		t.Fatalf("second complete = %v, want ErrAttemptClosed", err)
	}
}

func TestAttemptOwnership(t *testing.T) {
	f := newFixture(t, false)
	ctx := context.Background()
	view := f.view(t)

	attempt, err := f.svc.StartAttempt(ctx, f.user.ID, uuid.MustParse(view.Lesson.ChallengeID))
	if err != nil {
		t.Fatalf("starting attempt: %v", err)
	}

	other, err := f.db.CreateUser(ctx, database.CreateUserParams{Email: "other@example.com", PasswordHash: "hash"})
	if err != nil {
		t.Fatalf("creating other user: %v", err)
	}
	err = f.svc.SaveCode(ctx, other.ID, uuid.MustParse(attempt.ID), "print('hi')")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign save = %v, want ErrNotFound", err)
	}
	err = f.svc.CompleteAttempt(ctx, other.ID, uuid.MustParse(attempt.ID), true, 60)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign complete = %v, want ErrNotFound", err)
	}
}
