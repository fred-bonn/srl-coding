-- name: CreateUser :one
INSERT INTO users (email, password_hash) VALUES ($1, $2)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateAuthSession :exec
INSERT INTO auth_sessions (token, user_id, expires_at) VALUES ($1, $2, $3);
