package handlers

import (
	"context"
	"encoding/json"
	wshub "messaging-api/internal/websocket"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type wsHello struct {
	DialogID string `json:"dialog_id"`
}

func (h *Handler) ws(c *gin.Context) {
	// Принятие WebSocket-соединения
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to establish WebSocket connection"})
		return
	}

	// Контекст запроса используется как контекст для клиента
	ctx := c.Request.Context()

	client := wshub.NewClient(conn)
	conn.SetReadLimit(4 << 10) // ограничение размера сообщения

	rctx, rcancel := context.WithTimeout(ctx, 10*time.Second)
	_, data, err := conn.Read(rctx)
	rcancel()
	if err != nil {
		client.Close(wshub.CloseReasonServerShutdown)
		return
	}

	var hello wsHello
	if err := json.Unmarshal(data, &hello); err != nil {
		_ = conn.Close(websocket.StatusPolicyViolation, "invalid hello")
		return
	}

	did, err := uuid.Parse(hello.DialogID)
	if err != nil {
		_ = conn.Close(websocket.StatusPolicyViolation, "invalid dialog ID")
		return
	}

	h.hub.Subscribe(did.String(), client)

	go client.ReadPump(ctx)
	go client.WritePump(ctx)

	<-ctx.Done()

	h.hub.Unsubscribe(did.String(), client)
}
