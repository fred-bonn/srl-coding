-- +goose Up
CREATE TABLE lessons (
    id TEXT PRIMARY KEY,
    lesson_number INT NOT NULL UNIQUE,
    topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    challenge_id UUID NOT NULL REFERENCES challenges(id)
);

DROP TABLE user_lessons;
CREATE TABLE user_lessons (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id TEXT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, lesson_id)
);

-- +goose Down
DROP TABLE user_lessons;
CREATE TABLE user_lessons (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_number INT NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, lesson_number)
);
DROP TABLE lessons;
