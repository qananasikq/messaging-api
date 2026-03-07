package repositories

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"messaging-api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepo struct {
	db *pgxpool.Pool
}

func NewMessageRepo(db *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{db: db}
}

type CreateMessageParams struct {
	ID       uuid.UUID
	DialogID uuid.UUID
	SenderID uuid.UUID
	Content  string
}

func (r *MessageRepo) Create(ctx context.Context, p CreateMessageParams) (models.Message, error) {
	const q = `
		INSERT INTO messages (id, dialog_id, sender_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	var createdAt time.Time
	err := r.db.QueryRow(ctx, q, p.ID, p.DialogID, p.SenderID, p.Content).Scan(&createdAt)
	if err != nil {
		return models.Message{}, fmt.Errorf("insert message: %w", err)
	}

	return models.Message{
		ID:        p.ID,
		DialogID:  p.DialogID,
		SenderID:  p.SenderID,
		Content:   p.Content,
		CreatedAt: createdAt,
	}, nil
}

// Cursor helpers — base64url(RFC3339Nano|uuid)
func EncodeCursor(t time.Time, id uuid.UUID) string {
	// всегда в UTC, чтобы не было сюрпризов с зонами
	raw := t.UTC().Format(time.RFC3339Nano) + "|" + id.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(cursor string) (time.Time, uuid.UUID, error) {
	if cursor == "" {
		return time.Time{}, uuid.Nil, nil
	}

	b, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, uuid.Nil, ErrInvalidCursor
	}

	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, ErrInvalidCursor
	}

	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, ErrInvalidCursor
	}

	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, ErrInvalidCursor
	}

	return ts.UTC(), id, nil
}

type ListMessagesParams struct {
	DialogID uuid.UUID
	Limit    int
	Cursor   *string
}

type ListMessagesResult struct {
	Messages   []models.Message
	NextCursor *string
}

func (r *MessageRepo) ListByDialogDesc(ctx context.Context, p ListMessagesParams) (ListMessagesResult, error) {
	limit := clamp(p.Limit, 1, 200, 50)

	args := []any{p.DialogID, limit}
	where := ""

	if p.Cursor != nil && *p.Cursor != "" {
		ts, id, err := DecodeCursor(*p.Cursor)
		if err != nil {
			return ListMessagesResult{}, err
		}
		where = "AND (m.created_at < $3 OR (m.created_at = $3 AND m.id < $4))"
		args = append(args, ts, id)
	}

	const baseQuery = `
		SELECT id, dialog_id, sender_id, content, created_at
		FROM messages
		WHERE dialog_id = $1 %s
		ORDER BY created_at DESC, id DESC
		LIMIT $2`

	query := fmt.Sprintf(baseQuery, where)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return ListMessagesResult{}, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	messages := make([]models.Message, 0, limit)
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.DialogID, &m.SenderID, &m.Content, &m.CreatedAt); err != nil {
			return ListMessagesResult{}, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return ListMessagesResult{}, fmt.Errorf("rows error: %w", err)
	}

	var nextCursor *string
	if len(messages) == limit {
		last := messages[len(messages)-1]
		c := EncodeCursor(last.CreatedAt, last.ID)
		nextCursor = &c
	}

	return ListMessagesResult{
		Messages:   messages,
		NextCursor: nextCursor,
	}, nil
}

// маленький хелпер, чтобы не писать if-ы в каждом месте
func clamp(v, min, max, def int) int {
	if v < min || v > max {
		return def
	}
	return v
}

func (r *MessageRepo) CountUnreadSince(ctx context.Context, dialogID, userID uuid.UUID, since time.Time) (int64, error) {
	const q = `
		SELECT COUNT(*)
		FROM messages
		WHERE dialog_id = $1
		  AND created_at > $2
		  AND sender_id <> $3`

	var count int64
	err := r.db.QueryRow(ctx, q, dialogID, since, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unread messages: %w", err)
	}

	return count, nil
}

func (r *MessageRepo) GetLatestMessageTime(ctx context.Context, dialogID uuid.UUID) (time.Time, error) {
	const q = `
		SELECT COALESCE(MAX(created_at), '1970-01-01T00:00:00Z')
		FROM messages
		WHERE dialog_id = $1`

	var t time.Time
	err := r.db.QueryRow(ctx, q, dialogID).Scan(&t)
	if err != nil {
		return time.Time{}, fmt.Errorf("get latest message time: %w", err)
	}

	return t.UTC(), nil
}
