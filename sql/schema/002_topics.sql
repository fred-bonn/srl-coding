-- +goose Up
CREATE TABLE topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE challenge_topics (
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    PRIMARY KEY (challenge_id, topic_id)
);

-- +goose Down
DROP TABLE challenge_topics;
DROP TABLE topics;