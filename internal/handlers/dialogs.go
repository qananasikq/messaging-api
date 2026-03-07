package handlers

import (
	"net/http"

	"messaging-api/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createDialogReq struct {
	Name           *string  `json:"name"`
	ParticipantIDs []string `json:"participant_ids" binding:"required"`
}

func (h *Handler) createDialog(c *gin.Context) {
	var req createDialogReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, ErrBindValidation)
		return
	}
	ids := make([]uuid.UUID, 0, len(req.ParticipantIDs))
	for _, s := range req.ParticipantIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			writeError(c, ErrBindValidation)
			return
		}
		ids = append(ids, id)
	}

	d, err := h.dialogSvc.Create(c.Request.Context(), mustUserID(c), services.CreateDialogInput{
		Name:           req.Name,
		ParticipantIDs: ids,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *Handler) listDialogs(c *gin.Context) {
	ds, err := h.dialogSvc.ListMyDialogs(c.Request.Context(), mustUserID(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"dialogs": ds})
}

func (h *Handler) getDialog(c *gin.Context) {
	did, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeError(c, ErrBindValidation)
		return
	}
	d, err := h.dialogSvc.GetDialogDetail(c.Request.Context(), did, mustUserID(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *Handler) deleteDialog(c *gin.Context) {
	did, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeError(c, ErrBindValidation)
		return
	}
	if err := h.dialogSvc.Delete(c.Request.Context(), did, mustUserID(c)); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
