package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"messaging-api/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type postMessageReq struct {
	DialogID string `json:"dialog_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

func (h *Handler) postMessage(c *gin.Context) {
	var req postMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, ErrBindValidation)
		return
	}
	did, err := uuid.Parse(req.DialogID)
	if err != nil {
		writeError(c, ErrBindValidation)
		return
	}

	msg, err := h.messageSvc.Send(c.Request.Context(), mustUserID(c), services.SendMessageInput{
		DialogID: did,
		Content:  req.Content,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	evt := gin.H{"type": "message.created", "data": msg}
	b, _ := json.Marshal(evt)
	h.hub.Broadcast(did.String(), b)

	c.JSON(http.StatusCreated, msg)
}

func (h *Handler) listMessages(c *gin.Context) {
	did, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeError(c, ErrBindValidation)
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	cursor := c.Query("cursor")
	var curPtr *string
	if cursor != "" {
		curPtr = &cursor
	}

	res, err := h.messageSvc.ListMessages(c.Request.Context(), did, mustUserID(c), limit, curPtr)
	if err != nil {
		writeError(c, err)
		return
	}
	_ = h.dialogSvc.MarkRead(c.Request.Context(), did, mustUserID(c))

	c.JSON(http.StatusOK, gin.H{"messages": res.Messages, "next_cursor": res.NextCursor})
}

func (h *Handler) unreadCount(c *gin.Context) {
	did, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeError(c, ErrBindValidation)
		return
	}
	n, err := h.messageSvc.UnreadCount(c.Request.Context(), did, mustUserID(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"unread_count": n})
}
