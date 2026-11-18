-- name: CreateChallenge :one
INSERT INTO challenges (
    description,
    start_code,
    test_code,
    solution_code,
    difficulty
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;
