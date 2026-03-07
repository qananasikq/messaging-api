package handlers

import (
	"messaging-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/healthz", h.healthz)
	r.GET("/readyz", h.readyz)

	r.POST("/users", h.createUser)
	r.POST("/auth/login", h.loginUser)
	auth := r.Group("/")
	auth.Use(middleware.Auth(h.jwt))

	auth.GET("/users/:id", h.getUser)

	auth.POST("/dialogs", h.createDialog)
	auth.GET("/dialogs", h.listDialogs)
	auth.GET("/dialogs/:id", h.getDialog)
	auth.DELETE("/dialogs/:id", h.deleteDialog)

	auth.POST("/messages", h.postMessage)
	auth.GET("/dialogs/:id/messages", h.listMessages)
	auth.GET("/dialogs/:id/unread_count", h.unreadCount)

	auth.GET("/ws", h.ws)
}
