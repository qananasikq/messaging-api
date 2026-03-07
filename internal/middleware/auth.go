package middleware

import (
	"net/http"
	"strings"

	jwtpkg "messaging-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	CtxUserIDKey   = "user_id"
	CtxUsernameKey = "username"
)

func Auth(j *jwtpkg.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		token := ""
		if strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimPrefix(h, "Bearer ")
		} else if q := strings.TrimSpace(c.Query("token")); q != "" {
			token = q
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token", "code": "UNAUTHORIZED"})
			return
		}
		claims, err := j.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "code": "UNAUTHORIZED"})
			return
		}
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "code": "UNAUTHORIZED"})
			return
		}
		c.Set(CtxUserIDKey, uid)
		c.Set(CtxUsernameKey, claims.Username)
		c.Next()
	}
}
