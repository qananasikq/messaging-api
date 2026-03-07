package services

import (
	"context"
	"time"

	"messaging-api/internal/models"
	"messaging-api/internal/repositories"

	"github.com/google/uuid"
)

type DialogService struct {
	dialogRepo  *repositories.DialogRepo
	messageRepo *repositories.MessageRepo
	cache       *RedisCache
}

func NewDialogService(dr *repositories.DialogRepo, mr *repositories.MessageRepo, cache *RedisCache) *DialogService {
	return &DialogService{dialogRepo: dr, messageRepo: mr, cache: cache}
}

type CreateDialogInput struct {
	Name           *string
	ParticipantIDs []uuid.UUID
}

func (s *DialogService) Create(ctx context.Context, createdBy uuid.UUID, in CreateDialogInput) (models.Dialog, error) {

	found := false
	for _, id := range in.ParticipantIDs {
		if id == createdBy {
			found = true
			break
		}
	}
	if !found {
		in.ParticipantIDs = append(in.ParticipantIDs, createdBy)
	}

	in.ParticipantIDs = uniqueUUIDs(in.ParticipantIDs)
	if len(in.ParticipantIDs) < 2 || len(in.ParticipantIDs) > 50 {
		return models.Dialog{}, ErrValidation
	}

	dtype := models.DialogTypeGroup
	if len(uniqueUUIDs(in.ParticipantIDs)) == 2 && (in.Name == nil || *in.Name == "") {
		dtype = models.DialogTypeDirect
	}

	p := repositories.CreateDialogParams{
		ID:             uuid.New(),
		Type:           dtype,
		Name:           in.Name,
		CreatedBy:      createdBy,
		ParticipantIDs: in.ParticipantIDs,
	}
	d, err := s.dialogRepo.Create(ctx, p)
	if err != nil {
		return models.Dialog{}, err
	}
	parts, _ := s.dialogRepo.GetParticipants(ctx, d.ID)
	d.Participants = parts
	return d, nil
}

func (s *DialogService) ListMyDialogs(ctx context.Context, userID uuid.UUID) ([]models.Dialog, error) {
	ds, err := s.dialogRepo.ListDialogs(ctx, userID, 50)
	if err != nil {
		return nil, err
	}
	for i := range ds {
		parts, err := s.dialogRepo.GetParticipants(ctx, ds[i].ID)
		if err == nil {
			ds[i].Participants = parts
		}
		if n, ok, err := s.cache.GetUnread(ctx, userID, ds[i].ID); err == nil && ok {
			ds[i].UnreadCount = n
		} else {
			since, _ := s.dialogRepo.GetLastReadAt(ctx, ds[i].ID, userID)
			n2, err := s.messageRepo.CountUnreadSince(ctx, ds[i].ID, userID, since)
			if err == nil {
				ds[i].UnreadCount = n2
			}
		}
	}
	return ds, nil
}

func (s *DialogService) GetDialogDetail(ctx context.Context, dialogID, userID uuid.UUID) (models.Dialog, error) {
	if err := s.dialogRepo.EnsureParticipant(ctx, dialogID, userID); err != nil {
		if err == repositories.ErrForbidden {
			return models.Dialog{}, ErrForbidden
		}
		return models.Dialog{}, err
	}
	d, err := s.dialogRepo.GetDialog(ctx, dialogID)
	if err != nil {
		if err == repositories.ErrNotFound {
			return models.Dialog{}, ErrNotFound
		}
		return models.Dialog{}, err
	}
	parts, _ := s.dialogRepo.GetParticipants(ctx, dialogID)
	d.Participants = parts

	if n, ok, err := s.cache.GetUnread(ctx, userID, dialogID); err == nil && ok {
		d.UnreadCount = n
	} else {
		since, _ := s.dialogRepo.GetLastReadAt(ctx, dialogID, userID)
		n2, err := s.messageRepo.CountUnreadSince(ctx, dialogID, userID, since)
		if err == nil {
			d.UnreadCount = n2
		}
	}
	return d, nil
}

func (s *DialogService) MarkRead(ctx context.Context, dialogID, userID uuid.UUID) error {
	if err := s.dialogRepo.EnsureParticipant(ctx, dialogID, userID); err != nil {
		if err == repositories.ErrForbidden {
			return ErrForbidden
		}
		return err
	}
	now := time.Now().UTC()
	if err := s.dialogRepo.TouchRead(ctx, dialogID, userID, now); err != nil {
		return err
	}
	_ = s.cache.ResetUnread(ctx, userID, dialogID)
	return nil
}

func uniqueUUIDs(in []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(in))
	out := make([]uuid.UUID, 0, len(in))
	for _, v := range in {
		if v == uuid.Nil {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func (s *DialogService) Delete(ctx context.Context, dialogID, userID uuid.UUID) error {
	d, err := s.dialogRepo.GetDialog(ctx, dialogID)
	if err != nil {
		if err == repositories.ErrNotFound {
			return ErrNotFound
		}
		return err
	}
	if d.CreatedBy != userID {
		return ErrForbidden
	}
	if err := s.dialogRepo.Delete(ctx, dialogID); err != nil {
		if err == repositories.ErrNotFound {
			return ErrNotFound
		}
		return err
	}
	return nil
}
