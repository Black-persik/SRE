-- +goose Up
CREATE TYPE response_status AS ENUM (
    'waiting', 'rejected', 'accepted'
    );

CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT uuidv7(),
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ default now()
);

CREATE TABLE IF NOT EXISTS responses (
    response_id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    company TEXT NOT NULL,
    vacancy TEXT NOT NULL,
    status response_status NOT NULL DEFAULT 'waiting',
    date TIMESTAMPTZ DEFAULT now()

);

-- +goose Down
DROP TABLE IF EXISTS responses;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS response_status;