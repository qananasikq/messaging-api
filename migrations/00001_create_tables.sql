-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY,
  username text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dialogs (
  id uuid PRIMARY KEY,
  type text NOT NULL CHECK (type IN ('direct','group')),
  name text NULL,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dialog_participants (
  dialog_id uuid NOT NULL REFERENCES dialogs(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (dialog_id, user_id)
);

CREATE TABLE IF NOT EXISTS dialog_reads (
  dialog_id uuid NOT NULL REFERENCES dialogs(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  last_read_at timestamptz NOT NULL DEFAULT 'epoch',
  PRIMARY KEY (dialog_id, user_id)
);

CREATE TABLE IF NOT EXISTS messages (
  id uuid PRIMARY KEY,
  dialog_id uuid NOT NULL REFERENCES dialogs(id) ON DELETE CASCADE,
  sender_id uuid NOT NULL REFERENCES users(id),
  content text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS dialog_reads;
DROP TABLE IF EXISTS dialog_participants;
DROP TABLE IF EXISTS dialogs;
DROP TABLE IF EXISTS users;
