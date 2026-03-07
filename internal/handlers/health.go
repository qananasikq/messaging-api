package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) readyz(c *gin.Context) {
	if h.readyCheck == nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.readyCheck(ctx); err != nil {
		h.logger.Warn("readiness check failed", "err", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"ok": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
