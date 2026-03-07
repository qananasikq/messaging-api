package handlers

import (
	"net/http"

	"messaging-api/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createUserReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) createUser(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, ErrBindValidation)
		return
	}
	res, err := h.userSvc.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":       res.User.ID,
		"username": res.User.Username,
		"token":    res.Token,
	})
}

func (h *Handler) getUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeError(c, ErrBindValidation)
		return
	}
	u, err := h.userSvc.Get(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

func mustUserID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(middleware.CtxUserIDKey)
	return v.(uuid.UUID)
}

type loginUserReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) loginUser(c *gin.Context) {
	var req loginUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, ErrBindValidation)
		return
	}
	res, err := h.userSvc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":       res.User.ID,
		"username": res.User.Username,
		"token":    res.Token,
	})
}
