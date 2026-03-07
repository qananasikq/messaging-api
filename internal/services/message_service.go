package services

import (
	"context"
	"fmt"
	"strings"

	"messaging-api/internal/models"
	"messaging-api/internal/repositories"

	"github.com/google/uuid"
)

type MessageService struct {
	msgRepo    *repositories.MessageRepo
	dialogRepo *repositories.DialogRepo
	cache      *RedisCache
}

func NewMessageService(mr *repositories.MessageRepo, dr *repositories.DialogRepo, cache *RedisCache) *MessageService {
	return &MessageService{msgRepo: mr, dialogRepo: dr, cache: cache}
}

type SendMessageInput struct {
	DialogID uuid.UUID
	Content  string
}

func (s *MessageService) Send(ctx context.Context, senderID uuid.UUID, in SendMessageInput) (models.Message, error) {
	in.Content = strings.TrimSpace(in.Content)
	if in.DialogID == uuid.Nil || in.Content == "" || len(in.Content) > 4000 {
		return models.Message{}, ErrValidation
	}

	if err := s.dialogRepo.EnsureParticipant(ctx, in.DialogID, senderID); err != nil {
		if err == repositories.ErrForbidden {
			return models.Message{}, ErrForbidden
		}
		return models.Message{}, err
	}

	msg, err := s.msgRepo.Create(ctx, repositories.CreateMessageParams{
		ID:       uuid.New(),
		DialogID: in.DialogID,
		SenderID: senderID,
		Content:  in.Content,
	})
	if err != nil {
		return models.Message{}, fmt.Errorf("create message: %w", err)
	}

	parts, err := s.dialogRepo.GetParticipants(ctx, in.DialogID)
	if err == nil {
		unreadTargets := make([]uuid.UUID, 0, len(parts))
		for _, p := range parts {
			if p.ID != senderID {
				unreadTargets = append(unreadTargets, p.ID)
			}
		}
		_ = s.cache.IncrUnreadForUsers(ctx, in.DialogID, unreadTargets)
	}
	_ = s.cache.PushLastMessage(ctx, in.DialogID, msg)

	return msg, nil
}

func (s *MessageService) ListMessages(ctx context.Context, dialogID, userID uuid.UUID, limit int, cursor *string) (repositories.ListMessagesResult, error) {
	if err := s.dialogRepo.EnsureParticipant(ctx, dialogID, userID); err != nil {
		if err == repositories.ErrForbidden {
			return repositories.ListMessagesResult{}, ErrForbidden
		}
		return repositories.ListMessagesResult{}, err
	}

	if cursor == nil || *cursor == "" {
		if msgs, ok, err := s.cache.GetLastMessages(ctx, dialogID, limit); err == nil && ok {
			res := repositories.ListMessagesResult{Messages: msgs}
			if len(msgs) == limit && len(msgs) > 0 {
				last := msgs[len(msgs)-1]
				c := repositories.EncodeCursor(last.CreatedAt, last.ID)
				res.NextCursor = &c
			}
			return res, nil
		}
	}

	res, err := s.msgRepo.ListByDialogDesc(ctx, repositories.ListMessagesParams{
		DialogID: dialogID,
		Limit:    limit,
		Cursor:   cursor,
	})
	if err != nil {
		if err == repositories.ErrInvalidCursor {
			return repositories.ListMessagesResult{}, ErrValidation
		}
		return repositories.ListMessagesResult{}, err
	}
	return res, nil
}

func (s *MessageService) UnreadCount(ctx context.Context, dialogID, userID uuid.UUID) (int64, error) {
	if err := s.dialogRepo.EnsureParticipant(ctx, dialogID, userID); err != nil {
		if err == repositories.ErrForbidden {
			return 0, ErrForbidden
		}
		return 0, err
	}
	if n, ok, err := s.cache.GetUnread(ctx, userID, dialogID); err == nil && ok {
		return n, nil
	}
	since, _ := s.dialogRepo.GetLastReadAt(ctx, dialogID, userID)
	n, err := s.msgRepo.CountUnreadSince(ctx, dialogID, userID, since)
	if err != nil {
		return 0, err
	}
	return n, nil
}
