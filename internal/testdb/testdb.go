// Package testdb provides an in-memory database.Querier for tests.
package testdb

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/fred-bonn/srlcode/internal/database"
)

type DB struct {
	Users       map[uuid.UUID]database.User
	Sessions    map[string]database.AuthSession
	Topics      map[uuid.UUID]database.Topic
	Challenges  map[uuid.UUID]database.Challenge
	Lessons     map[string]database.Lesson
	Attempts    map[uuid.UUID]database.UserReview
	TopicLinks  map[uuid.UUID][]uuid.UUID
	UserLessons []database.UserLesson
	UserTopics  []database.UserTopic
}

var _ database.Querier = (*DB)(nil)

func New() *DB {
	return &DB{
		Users:      map[uuid.UUID]database.User{},
		Sessions:   map[string]database.AuthSession{},
		Topics:     map[uuid.UUID]database.Topic{},
		Challenges: map[uuid.UUID]database.Challenge{},
		Lessons:    map[string]database.Lesson{},
		Attempts:   map[uuid.UUID]database.UserReview{},
		TopicLinks: map[uuid.UUID][]uuid.UUID{},
	}
}

func (d *DB) CompleteAttempt(ctx context.Context, arg database.CompleteAttemptParams) (int64, error) {
	attempt, ok := d.Attempts[arg.ID]
	if !ok || attempt.Correct.Valid {
		return 0, nil
	}
	attempt.Correct = arg.Correct
	attempt.DurationSeconds = arg.DurationSeconds
	attempt.CompletedAt = arg.CompletedAt
	d.Attempts[arg.ID] = attempt
	return 1, nil
}

func (d *DB) CompleteLesson(ctx context.Context, arg database.CompleteLessonParams) error {
	d.UserLessons = append(d.UserLessons, database.UserLesson{
		UserID:      arg.UserID,
		LessonID:    arg.LessonID,
		CompletedAt: arg.CompletedAt,
	})
	return nil
}

func (d *DB) CreateAttempt(ctx context.Context, arg database.CreateAttemptParams) (database.UserReview, error) {
	attempt := database.UserReview{
		ID:          uuid.New(),
		UserID:      arg.UserID,
		ChallengeID: arg.ChallengeID,
		Code:        arg.Code,
		StartedAt:   arg.StartedAt,
	}
	d.Attempts[attempt.ID] = attempt
	return attempt, nil
}

func (d *DB) CreateAuthSession(ctx context.Context, arg database.CreateAuthSessionParams) error {
	d.Sessions[arg.Token] = database.AuthSession{
		Token:     arg.Token,
		UserID:    arg.UserID,
		CreatedAt: time.Now(),
		ExpiresAt: arg.ExpiresAt,
	}
	return nil
}

func (d *DB) CreateChallenge(ctx context.Context, arg database.CreateChallengeParams) (database.Challenge, error) {
	challenge := database.Challenge{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Assignment:   arg.Assignment,
		StartCode:    arg.StartCode,
		TestCode:     arg.TestCode,
		SolutionCode: arg.SolutionCode,
		Difficulty:   arg.Difficulty,
	}
	d.Challenges[challenge.ID] = challenge
	return challenge, nil
}

func (d *DB) CreateChallengeTopic(ctx context.Context, arg database.CreateChallengeTopicParams) error {
	d.TopicLinks[arg.ChallengeID] = append(d.TopicLinks[arg.ChallengeID], arg.TopicID)
	return nil
}

func (d *DB) CreateLesson(ctx context.Context, arg database.CreateLessonParams) error {
	d.Lessons[arg.ID] = database.Lesson{
		ID:           arg.ID,
		LessonNumber: arg.LessonNumber,
		TopicID:      arg.TopicID,
		ChallengeID:  arg.ChallengeID,
	}
	return nil
}

func (d *DB) CreateTopic(ctx context.Context, name string) (database.Topic, error) {
	if topic, err := d.GetTopicByName(ctx, name); err == nil {
		return topic, nil
	}
	topic := database.Topic{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	}
	d.Topics[topic.ID] = topic
	return topic, nil
}

func (d *DB) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	user := database.User{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Email:        arg.Email,
		PasswordHash: arg.PasswordHash,
	}
	d.Users[user.ID] = user
	return user, nil
}

func (d *DB) EnsureChallengeTopic(ctx context.Context, arg database.EnsureChallengeTopicParams) error {
	for _, topicID := range d.TopicLinks[arg.ChallengeID] {
		if topicID == arg.TopicID {
			return nil
		}
	}
	return d.CreateChallengeTopic(ctx, database.CreateChallengeTopicParams(arg))
}

func (d *DB) EnsureUserTopic(ctx context.Context, arg database.EnsureUserTopicParams) error {
	for _, ut := range d.UserTopics {
		if ut.UserID == arg.UserID && ut.TopicID == arg.TopicID {
			return nil
		}
	}
	d.UserTopics = append(d.UserTopics, database.UserTopic{
		UserID:    arg.UserID,
		TopicID:   arg.TopicID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DueAt:     arg.DueAt,
	})
	return nil
}

func (d *DB) GetAttempt(ctx context.Context, id uuid.UUID) (database.UserReview, error) {
	attempt, ok := d.Attempts[id]
	if !ok {
		return database.UserReview{}, sql.ErrNoRows
	}
	return attempt, nil
}

func (d *DB) GetChallenge(ctx context.Context, id uuid.UUID) (database.Challenge, error) {
	challenge, ok := d.Challenges[id]
	if !ok {
		return database.Challenge{}, sql.ErrNoRows
	}
	return challenge, nil
}

func (d *DB) GetCurrentLesson(ctx context.Context, userID uuid.UUID) (database.Lesson, error) {
	var current database.Lesson
	found := false
	for _, lesson := range d.Lessons {
		if d.hasCompletedLesson(userID, lesson.ID) {
			continue
		}
		if !found || lesson.LessonNumber < current.LessonNumber {
			current, found = lesson, true
		}
	}
	if !found {
		return database.Lesson{}, sql.ErrNoRows
	}
	return current, nil
}

func (d *DB) GetLesson(ctx context.Context, id string) (database.Lesson, error) {
	lesson, ok := d.Lessons[id]
	if !ok {
		return database.Lesson{}, sql.ErrNoRows
	}
	return lesson, nil
}

func (d *DB) GetOpenAttemptByChallenge(ctx context.Context, arg database.GetOpenAttemptByChallengeParams) (database.UserReview, error) {
	var open database.UserReview
	found := false
	for _, attempt := range d.Attempts {
		if attempt.UserID != arg.UserID || attempt.ChallengeID != arg.ChallengeID || attempt.Correct.Valid {
			continue
		}
		if !found || attempt.StartedAt.Before(open.StartedAt) {
			open, found = attempt, true
		}
	}
	if !found {
		return database.UserReview{}, sql.ErrNoRows
	}
	return open, nil
}

func (d *DB) GetTopicByName(ctx context.Context, name string) (database.Topic, error) {
	for _, topic := range d.Topics {
		if topic.Name == name {
			return topic, nil
		}
	}
	return database.Topic{}, sql.ErrNoRows
}

func (d *DB) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	for _, user := range d.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return database.User{}, sql.ErrNoRows
}

func (d *DB) GetUserBySessionToken(ctx context.Context, token string) (database.User, error) {
	session, ok := d.Sessions[token]
	if !ok || session.ExpiresAt.Before(time.Now()) {
		return database.User{}, sql.ErrNoRows
	}
	user, ok := d.Users[session.UserID]
	if !ok {
		return database.User{}, sql.ErrNoRows
	}
	return user, nil
}

func (d *DB) HasCompletedLessonToday(ctx context.Context, arg database.HasCompletedLessonTodayParams) (bool, error) {
	for _, ul := range d.UserLessons {
		if ul.UserID == arg.UserID && !ul.CompletedAt.Before(arg.CompletedAt) {
			return true, nil
		}
	}
	return false, nil
}

func (d *DB) ListChallengeTopics(ctx context.Context, challengeID uuid.UUID) ([]uuid.UUID, error) {
	return d.TopicLinks[challengeID], nil
}

func (d *DB) ListCompletedTopicReviews(ctx context.Context, arg database.ListCompletedTopicReviewsParams) ([]database.ListCompletedTopicReviewsRow, error) {
	var reviews []database.ListCompletedTopicReviewsRow
	for _, attempt := range d.Attempts {
		if attempt.UserID != arg.UserID || !attempt.Correct.Valid || !attempt.CompletedAt.Valid {
			continue
		}
		if !d.challengeHasTopic(attempt.ChallengeID, arg.TopicID) {
			continue
		}
		reviews = append(reviews, database.ListCompletedTopicReviewsRow{
			Correct:     attempt.Correct,
			CompletedAt: attempt.CompletedAt,
		})
	}
	sort.Slice(reviews, func(i, j int) bool { return reviews[i].CompletedAt.Time.Before(reviews[j].CompletedAt.Time) })
	return reviews, nil
}

func (d *DB) ListDueUserTopics(ctx context.Context, arg database.ListDueUserTopicsParams) ([]database.ListDueUserTopicsRow, error) {
	var due []database.ListDueUserTopicsRow
	for _, ut := range d.UserTopics {
		if ut.UserID != arg.UserID || ut.DueAt.After(arg.DueAt) {
			continue
		}
		topic, ok := d.Topics[ut.TopicID]
		if !ok {
			continue
		}
		due = append(due, database.ListDueUserTopicsRow{ID: topic.ID, Name: topic.Name})
	}
	sort.Slice(due, func(i, j int) bool { return due[i].Name < due[j].Name })
	return due, nil
}

func (d *DB) ListOpenAttempts(ctx context.Context, userID uuid.UUID) ([]database.ListOpenAttemptsRow, error) {
	var open []database.ListOpenAttemptsRow
	for _, attempt := range d.Attempts {
		if attempt.UserID != userID || attempt.Correct.Valid {
			continue
		}
		challenge, ok := d.Challenges[attempt.ChallengeID]
		if !ok {
			continue
		}
		open = append(open, database.ListOpenAttemptsRow{
			ID:          attempt.ID,
			Code:        attempt.Code,
			StartedAt:   attempt.StartedAt,
			ChallengeID: challenge.ID,
			Assignment:  challenge.Assignment,
		})
	}
	sort.Slice(open, func(i, j int) bool { return open[i].StartedAt.Before(open[j].StartedAt) })
	return open, nil
}

func (d *DB) SaveAttemptCode(ctx context.Context, arg database.SaveAttemptCodeParams) (int64, error) {
	attempt, ok := d.Attempts[arg.ID]
	if !ok || attempt.Correct.Valid {
		return 0, nil
	}
	attempt.Code = arg.Code
	d.Attempts[arg.ID] = attempt
	return 1, nil
}

func (d *DB) SetUserTopicDue(ctx context.Context, arg database.SetUserTopicDueParams) error {
	for i, ut := range d.UserTopics {
		if ut.UserID == arg.UserID && ut.TopicID == arg.TopicID {
			ut.DueAt = arg.DueAt
			ut.UpdatedAt = time.Now()
			d.UserTopics[i] = ut
			return nil
		}
	}
	return nil
}

func (d *DB) UpdateChallengeContent(ctx context.Context, arg database.UpdateChallengeContentParams) error {
	challenge, ok := d.Challenges[arg.ID]
	if !ok {
		return sql.ErrNoRows
	}
	challenge.Assignment = arg.Assignment
	challenge.StartCode = arg.StartCode
	challenge.TestCode = arg.TestCode
	challenge.SolutionCode = arg.SolutionCode
	challenge.Difficulty = arg.Difficulty
	challenge.UpdatedAt = time.Now()
	d.Challenges[arg.ID] = challenge
	return nil
}

func (d *DB) hasCompletedLesson(userID uuid.UUID, lessonID string) bool {
	for _, ul := range d.UserLessons {
		if ul.UserID == userID && ul.LessonID == lessonID {
			return true
		}
	}
	return false
}

func (d *DB) challengeHasTopic(challengeID, topicID uuid.UUID) bool {
	for _, linked := range d.TopicLinks[challengeID] {
		if linked == topicID {
			return true
		}
	}
	return false
}
