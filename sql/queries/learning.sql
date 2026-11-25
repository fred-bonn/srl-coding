-- name: CompleteLesson :exec
INSERT INTO user_lessons (user_id, lesson_id, completed_at) VALUES ($1, $2, $3);

-- name: EnsureUserTopic :exec
INSERT INTO user_topics (user_id, topic_id, due_at) VALUES ($1, $2, $3)
ON CONFLICT (user_id, topic_id) DO NOTHING;

-- name: ListDueUserTopics :many
SELECT t.id, t.name FROM user_topics ut
JOIN topics t ON t.id = ut.topic_id
WHERE ut.user_id = $1 AND ut.due_at <= $2
ORDER BY ut.due_at, t.name;

-- name: SetUserTopicDue :exec
UPDATE user_topics SET due_at = $3, updated_at = now()
WHERE user_id = $1 AND topic_id = $2;

-- name: CreateAttempt :one
INSERT INTO user_reviews (user_id, challenge_id, code, started_at) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAttempt :one
SELECT * FROM user_reviews WHERE id = $1;

-- name: SaveAttemptCode :execrows
UPDATE user_reviews SET code = $2 WHERE id = $1 AND correct IS NULL;

-- name: CompleteAttempt :execrows
UPDATE user_reviews
SET correct = $2, duration_seconds = $3, completed_at = $4
WHERE id = $1 AND correct IS NULL;

-- name: ListOpenAttempts :many
SELECT r.id, r.code, r.started_at, c.id AS challenge_id, c.assignment
FROM user_reviews r
JOIN challenges c ON c.id = r.challenge_id
WHERE r.user_id = $1 AND r.correct IS NULL
ORDER BY r.started_at;

-- name: ListCompletedTopicReviews :many
SELECT r.correct, r.completed_at
FROM user_reviews r
JOIN challenge_topics ct ON ct.challenge_id = r.challenge_id
WHERE r.user_id = $1 AND ct.topic_id = $2 AND r.correct IS NOT NULL
ORDER BY r.completed_at;

-- name: GetOpenAttemptByChallenge :one
SELECT * FROM user_reviews
WHERE user_id = $1 AND challenge_id = $2 AND correct IS NULL
ORDER BY started_at
LIMIT 1;

-- name: HasCompletedLessonToday :one
SELECT EXISTS (
    SELECT 1 FROM user_lessons
    WHERE user_id = $1 AND completed_at >= $2
) AS completed;
