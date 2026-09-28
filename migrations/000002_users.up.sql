CREATE TABLE chat.users (
    id BIGSERIAL PRIMARY KEY,
    login VARCHAR(32) NOT NULL UNIQUE CHECK (login ~ '^[a-z0-9_]{3,32}$'),
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
