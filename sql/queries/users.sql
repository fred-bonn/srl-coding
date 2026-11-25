-- name: CreateUser :one
INSERT INTO users (email, password_hash) VALUES ($1, $2)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateAuthSession :exec
INSERT INTO auth_sessions (token, user_id, expires_at) VALUES ($1, $2, $3);

-- name: GetUserBySessionToken :one
SELECT u.* FROM users u
JOIN auth_sessions s ON s.user_id = u.id
WHERE s.token = $1 AND s.expires_at > now();
