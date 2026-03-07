-- +goose Up

CREATE INDEX IF NOT EXISTS idx_messages_dialog_created_at_desc
  ON messages (dialog_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_messages_dialog_id_desc
  ON messages (dialog_id, id DESC);

CREATE INDEX IF NOT EXISTS idx_dialog_participants_user
  ON dialog_participants (user_id, dialog_id);

CREATE INDEX IF NOT EXISTS idx_dialog_reads_user
  ON dialog_reads (user_id, dialog_id);

-- +goose Down
DROP INDEX IF EXISTS idx_dialog_reads_user;
DROP INDEX IF EXISTS idx_dialog_participants_user;
DROP INDEX IF EXISTS idx_messages_dialog_id_desc;
DROP INDEX IF EXISTS idx_messages_dialog_created_at_desc;
