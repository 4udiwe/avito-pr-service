-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE pr_status (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);
    
INSERT INTO pr_status (id, name) VALUES (0, 'OPEN'), (1, 'MERGED');

CREATE TABLE team (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE user (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    team_id UUID REFERENCES team(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE pr (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES user(id) ON DELETE RESTRICT,
    status_id INT NOT NULL REFERENCES pr_status(id) DEFAULT 0,
    need_more_reviewers BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now(),
    merged_at TIMESTAMPTZ
);

CREATE TABLE pr_reviewer (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pr_id TEXT NOT NULL REFERENCES pr(id) ON DELETE CASCADE,
    reviewer_id TEXT NOT NULL REFERENCES user(id) ON DELETE RESTRICT,
    assigned_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (pr_id, reviewer_id)
);

CREATE INDEX idx_user_team_id ON user(team_id);
CREATE INDEX idx_pr_author_id ON pr(author_id);
CREATE INDEX idx_pr_reviewer_pr_id ON pr_reviewer(pr_id);
CREATE INDEX idx_pr_reviewer_reviewer_id ON pr_reviewer(reviewer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_pr_reviewer_reviewer_id;
DROP INDEX IF EXISTS idx_pr_reviewer_pr_id;
DROP INDEX IF EXISTS idx_pr_author_id;
DROP INDEX IF EXISTS idx_user_team_id;

DROP TABLE IF EXISTS pr_reviewer;
DROP TABLE IF EXISTS pr;
DROP TABLE IF EXISTS user;
DROP TABLE IF EXISTS team;
DROP TABLE IF EXISTS pr_status;
-- +goose StatementEnd
