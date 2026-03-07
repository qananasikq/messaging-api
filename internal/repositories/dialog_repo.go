package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"messaging-api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/ defined globally.

type CreateDialogParams struct {
	ID             uuid.UUID
	Type           models.DialogType
	Name           *string
	CreatedBy      uuid.UUID
	ParticipantIDs []uuid.UUID
}



type DialogRepo struct {
	db *pgxpool.Pool
}

func NewDialogRepo(db *pgxpool.Pool) *DialogRepo {
	return &DialogRepo{db: db}
}

// Create — создаёт диалог + участников + записи о прочтении в одной транзакции
func (r *DialogRepo) Create(ctx context.Context, p CreateDialogParams) (models.Dialog, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Dialog{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) // безопасный rollback даже при панике

	// 1. Сам диалог
	const qInsertDialog = `
		INSERT INTO dialogs (id, type, name, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`
	var createdAt time.Time
	err = tx.QueryRow(ctx, qInsertDialog, p.ID, p.Type, p.Name, p.CreatedBy).Scan(&createdAt)
	if err != nil {
		return models.Dialog{}, fmt.Errorf("insert dialog: %w", err)
	}

	// 2. Участники (ON CONFLICT DO NOTHING — на случай дублей)
	if err = r.batchInsert(ctx, tx,
		"dialog_participants", "(dialog_id, user_id)",
		p.ID, p.ParticipantIDs); err != nil {
		return models.Dialog{}, err
	}

	// 3. Записи о прочтении (тоже idempotent)
	if err = r.batchInsert(ctx, tx,
		"dialog_reads", "(dialog_id, user_id, last_read_at)",
		p.ID, p.ParticipantIDs, "now()"); err != nil {
		return models.Dialog{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return models.Dialog{}, fmt.Errorf("commit: %w", err)
	}

	return models.Dialog{
		ID:        p.ID,
		Type:      p.Type,
		Name:      p.Name,
		CreatedBy: p.CreatedBy,
		CreatedAt: createdAt,
	}, nil
}

// Вспомогательная функция для батч-вставок
func (r *DialogRepo) batchInsert(ctx context.Context, tx pgx.Tx, table, columns string, dialogID uuid.UUID, userIDs []uuid.UUID, extraValues ...any) error {
	const qTpl = `INSERT INTO %s %s VALUES ($1, $2%s) ON CONFLICT DO NOTHING`

	var q string
	if len(extraValues) > 0 {
		q = fmt.Sprintf(qTpl, table, columns, ", $3")
	} else {
		q = fmt.Sprintf(qTpl, table, columns, "")
	}

	b := &pgx.Batch{}
	for _, uid := range userIDs {
		if len(extraValues) > 0 {
			b.Queue(q, dialogID, uid, extraValues[0])
		} else {
			b.Queue(q, dialogID, uid)
		}
	}

	br := tx.SendBatch(ctx, b)
	defer br.Close()

	for range userIDs {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert into %s: %w", table, err)
		}
	}

	return nil
}

func (r *DialogRepo) EnsureParticipant(ctx context.Context, dialogID, userID uuid.UUID) error {
	const q = `SELECT 1 FROM dialog_participants WHERE dialog_id = $1 AND user_id = $2`
	var dummy int
	err := r.db.QueryRow(ctx, q, dialogID, userID).Scan(&dummy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return fmt.Errorf("check participant: %w", err)
	}
	return nil
}

func (r *DialogRepo) GetDialog(ctx context.Context, dialogID uuid.UUID) (models.Dialog, error) {
	const q = `
		SELECT id, type, name, created_by, created_at
		FROM dialogs
		WHERE id = $1`

	var d models.Dialog
	err := r.db.QueryRow(ctx, q, dialogID).Scan(
		&d.ID, &d.Type, &d.Name, &d.CreatedBy, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Dialog{}, ErrNotFound
		}
		return models.Dialog{}, fmt.Errorf("get dialog: %w", err)
	}
	return d, nil
}

func (r *DialogRepo) GetParticipants(ctx context.Context, dialogID uuid.UUID) ([]models.UserMini, error) {
	const q = `
		SELECT u.id, u.username
		FROM dialog_participants dp
		JOIN users u ON u.id = dp.user_id
		WHERE dp.dialog_id = $1
		ORDER BY u.username`

	rows, err := r.db.Query(ctx, q, dialogID)
	if err != nil {
		return nil, fmt.Errorf("query participants: %w", err)
	}
	defer rows.Close()

	var users []models.UserMini
	for rows.Next() {
		var u models.UserMini
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			return nil, fmt.Errorf("scan participant: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *DialogRepo) ListDialogs(ctx context.Context, userID uuid.UUID, limit int) ([]models.Dialog, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	const q = `
		WITH my_dialogs AS (
			SELECT d.id, d.type, d.name, d.created_by, d.created_at
			FROM dialogs d
			JOIN dialog_participants dp ON dp.dialog_id = d.id
			WHERE dp.user_id = $1
		)
		SELECT
			d.id, d.type, d.name, d.created_by, d.created_at,
			lm.id, lm.sender_id, lm.content, lm.created_at
		FROM my_dialogs d
		LEFT JOIN LATERAL (
			SELECT m.id, m.sender_id, m.content, m.created_at
			FROM messages m
			WHERE m.dialog_id = d.id
			ORDER BY m.created_at DESC, m.id DESC
			LIMIT 1
		) lm ON true
		ORDER BY COALESCE(lm.created_at, d.created_at) DESC
		LIMIT $2`

	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list dialogs query: %w", err)
	}
	defer rows.Close()

	var dialogs []models.Dialog
	for rows.Next() {
		var d models.Dialog
		var lastMsgID *uuid.UUID
		var senderID *uuid.UUID
		var content *string
		var msgTime *time.Time

		err = rows.Scan(
			&d.ID, &d.Type, &d.Name, &d.CreatedBy, &d.CreatedAt,
			&lastMsgID, &senderID, &content, &msgTime,
		)
		if err != nil {
			return nil, fmt.Errorf("scan dialog row: %w", err)
		}

		if lastMsgID != nil {
			d.LastMessage = &models.MessageMini{
				ID:        *lastMsgID,
				SenderID:  *senderID,
				Content:   *content,
				CreatedAt: *msgTime,
			}
		}

		dialogs = append(dialogs, d)
	}

	return dialogs, rows.Err()
}

func (r *DialogRepo) TouchRead(ctx context.Context, dialogID, userID uuid.UUID, readAt time.Time) error {
	const q = `
		INSERT INTO dialog_reads (dialog_id, user_id, last_read_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (dialog_id, user_id)
		DO UPDATE SET last_read_at = EXCLUDED.last_read_at`

	_, err := r.db.Exec(ctx, q, dialogID, userID, readAt)
	if err != nil {
		return fmt.Errorf("update last read: %w", err)
	}
	return nil
}

func (r *DialogRepo) GetLastReadAt(ctx context.Context, dialogID, userID uuid.UUID) (time.Time, error) {
	const q = `SELECT last_read_at FROM dialog_reads WHERE dialog_id = $1 AND user_id = $2`

	var t time.Time
	err := r.db.QueryRow(ctx, q, dialogID, userID).Scan(&t)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Unix(0, 0).UTC(), nil
		}
		return time.Time{}, fmt.Errorf("get last read time: %w", err)
	}

	return t.UTC(), nil
}

func (r *DialogRepo) Delete(ctx context.Context, dialogID uuid.UUID) error {
	const q = `DELETE FROM dialogs WHERE id = $1`

	cmd, err := r.db.Exec(ctx, q, dialogID)
	if err != nil {
		return fmt.Errorf("delete dialog: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
