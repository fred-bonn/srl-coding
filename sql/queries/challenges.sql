-- name: CreateChallenge :one
INSERT INTO challenges (
    assignment,
    start_code,
    test_code,
    solution_code,
    difficulty
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetChallenge :one
SELECT * FROM challenges WHERE id = $1;

-- name: UpdateChallengeContent :exec
UPDATE challenges SET
    assignment = $2,
    start_code = $3,
    test_code = $4,
    solution_code = $5,
    difficulty = $6,
    updated_at = now()
WHERE id = $1;

-- name: ListChallengesForTopic :many
SELECT c.* FROM challenges c
JOIN challenge_topics ct ON ct.challenge_id = c.id
WHERE ct.topic_id = $1;

-- name: ListUnseenChallengesForTopic :many
SELECT c.* FROM challenges c
JOIN challenge_topics ct ON ct.challenge_id = c.id
LEFT JOIN user_reviews ur ON ur.challenge_id = c.id AND ur.user_id = $1
WHERE ct.topic_id = $2 AND ur.id IS NULL;
