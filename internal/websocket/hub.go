package websocket

import (
	"context"
	"log/slog"
	"sync"
)

type CloseReason string

const CloseReasonServerShutdown CloseReason = "server_shutdown"

type Hub struct {
	logger *slog.Logger

	mu        sync.RWMutex
	subsByDlg map[string]map[*Client]struct{}
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		logger:    logger,
		subsByDlg: make(map[string]map[*Client]struct{}),
	}
}

func (h *Hub) Run(ctx context.Context) {
	<-ctx.Done()
}

func (h *Hub) Subscribe(dialogID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set, ok := h.subsByDlg[dialogID]
	if !ok {
		set = make(map[*Client]struct{})
		h.subsByDlg[dialogID] = set
	}
	set[c] = struct{}{}
}

func (h *Hub) Unsubscribe(dialogID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.subsByDlg[dialogID]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.subsByDlg, dialogID)
		}
	}
}

func (h *Hub) Broadcast(dialogID string, payload []byte) {
	h.mu.RLock()
	set := h.subsByDlg[dialogID]
	clients := make([]*Client, 0, len(set))
	for c := range set {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		c.TrySend(payload)
	}
}

func (h *Hub) CloseAll(reason CloseReason) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, set := range h.subsByDlg {
		for c := range set {
			c.Close(reason)
		}
	}
	h.subsByDlg = make(map[string]map[*Client]struct{})
}
