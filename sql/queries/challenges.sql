-- name: CreateChallenge :one
INSERT INTO challenges (
    start_code,
    test_code,
    solution_code
) VALUES (
    $1, $2, $3
)
RETURNING *;
