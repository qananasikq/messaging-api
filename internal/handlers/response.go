package handlers

import (
	"net/http"

	"messaging-api/internal/services"

	"github.com/gin-gonic/gin"
)

func writeError(c *gin.Context, err error) {
	switch err {
	case services.ErrValidation:
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation error", "code": "VALIDATION"})
	case services.ErrUnauthorized:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "code": "UNAUTHORIZED"})
	case services.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "code": "FORBIDDEN"})
	case services.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "not found", "code": "NOT_FOUND"})
	case services.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{"error": "conflict", "code": "CONFLICT"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error", "code": "INTERNAL"})
	}
}
