-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    passwordHash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX users_login_hash_idx on users USING HASH(login);
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS users;