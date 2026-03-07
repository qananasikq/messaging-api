package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func TokenAuth(token string) gin.HandlerFunc {
	if token == "" {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		if c.GetHeader("X-Access-Token") == token {
			c.Next()
			return
		}

		ah := c.GetHeader("Authorization")
		if strings.HasPrefix(ah, "Bearer ") && strings.TrimPrefix(ah, "Bearer ") == token {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}
