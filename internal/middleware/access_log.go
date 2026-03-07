package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		lat := time.Since(start)
		status := c.Writer.Status()
		rid, _ := c.Get(CtxRequestIDKey)

		logger.Info("http",
			"rid", rid,
			"method", method,
			"path", path,
			"status", status,
			"lat_ms", float64(lat.Milliseconds()),
			"client_ip", c.ClientIP(),
		)
	}
}
