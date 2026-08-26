-- typebook's own table. Not part of CrydenSync's schema — a real app
-- using CrydenSync owns its own domain tables and just references
-- user_id as a foreign key, same as any consumer would.

CREATE TABLE IF NOT EXISTS notes (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      TEXT NOT NULL DEFAULT '',
    body       TEXT NOT NULL DEFAULT '',
    color      TEXT NOT NULL DEFAULT 'default',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes(user_id, updated_at DESC);