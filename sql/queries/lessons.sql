-- name: GetLesson :one
SELECT * FROM lessons WHERE id = $1;

-- name: CreateLesson :exec
INSERT INTO lessons (id, lesson_number, topic_id, challenge_id) VALUES ($1, $2, $3, $4);

-- name: GetCurrentLesson :one
SELECT l.* FROM lessons l
WHERE NOT EXISTS (
    SELECT 1 FROM user_lessons ul
    WHERE ul.user_id = $1 AND ul.lesson_id = l.id
)
ORDER BY l.lesson_number
LIMIT 1;

-- name: GetLessonByTopic :one
SELECT * FROM lessons WHERE topic_id = $1;
