-- +goose Up
CREATE TABLE challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    description TEXT NOT NULL,
    start_code TEXT NOT NULL,
    test_code TEXT NOT NULL,
    solution_code TEXT NOT NULL
);

-- +goose Down
DROP TABLE challenges;