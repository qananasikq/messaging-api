package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type bucket struct {
	mu       sync.Mutex
	tokens   float64
	last     time.Time
	rate     float64
	burst    float64
	lastSeen time.Time
}

func (b *bucket) allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens = min(b.burst, b.tokens+elapsed*b.rate)
		b.last = now
	}
	b.lastSeen = now

	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func RateLimitPerUser(rate, burst int, cleanupPeriod time.Duration) gin.HandlerFunc {
	var m sync.Map

	go func() {
		t := time.NewTicker(cleanupPeriod)
		defer t.Stop()
		for range t.C {
			now := time.Now()
			m.Range(func(k, v any) bool {
				b := v.(*bucket)
				b.mu.Lock()
				idle := now.Sub(b.lastSeen) > 10*cleanupPeriod
				b.mu.Unlock()
				if idle {
					m.Delete(k)
				}
				return true
			})
		}
	}()

	return func(c *gin.Context) {
		// Не лимитируем WebSocket-подключения
		if c.Request.URL.Path == "/ws" {
			c.Next()
			return
		}

		key := c.ClientIP()
		if v, ok := c.Get(CtxUserIDKey); ok {
			if uid, ok2 := v.(uuid.UUID); ok2 {
				key = uid.String()
			}
		}

		now := time.Now()
		v, _ := m.LoadOrStore(key, &bucket{
			tokens:   float64(burst),
			last:     now,
			lastSeen: now,
			rate:     float64(rate),
			burst:    float64(burst),
		})
		b := v.(*bucket)
		if !b.allow(now) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"code":  "RATE_LIMIT",
			})
			return
		}
		c.Next()
	}
}
