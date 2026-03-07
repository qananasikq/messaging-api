package handlers

import (
	"context"
	"log/slog"

	"messaging-api/internal/services"
	wshub "messaging-api/internal/websocket"
	jwtpkg "messaging-api/pkg/jwt"
)

type Handler struct {
	logger *slog.Logger

	userSvc    *services.UserService
	dialogSvc  *services.DialogService
	messageSvc *services.MessageService

	jwt *jwtpkg.JWT
	hub *wshub.Hub

	readyCheck func(ctx context.Context) error
}

type Deps struct {
	Logger     *slog.Logger
	UserSvc    *services.UserService
	DialogSvc  *services.DialogService
	MessageSvc *services.MessageService
	JWT        *jwtpkg.JWT
	WSHub      *wshub.Hub
	ReadyCheck func(ctx context.Context) error
}

func NewHandler(d Deps) *Handler {
	return &Handler{
		logger:     d.Logger,
		userSvc:    d.UserSvc,
		dialogSvc:  d.DialogSvc,
		messageSvc: d.MessageSvc,
		jwt:        d.JWT,
		hub:        d.WSHub,
		readyCheck: d.ReadyCheck,
	}
}
