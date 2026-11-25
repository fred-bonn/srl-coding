-- name: CreateTopic :one
INSERT INTO topics (name) VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = excluded.name
RETURNING *;

-- name: CreateChallengeTopic :exec
INSERT INTO challenge_topics (challenge_id, topic_id) VALUES ($1, $2);

-- name: EnsureChallengeTopic :exec
INSERT INTO challenge_topics (challenge_id, topic_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListChallengeTopics :many
SELECT topic_id FROM challenge_topics WHERE challenge_id = $1;

-- name: GetTopicByName :one
SELECT * FROM topics WHERE name = $1;

-- name: GetTopic :one
SELECT * FROM topics WHERE id = $1;
