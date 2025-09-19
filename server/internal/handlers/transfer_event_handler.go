package handlers

import (
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TransferEventHandler struct{ svc services.TransferEventService }

func NewTransferEventHandler(svc services.TransferEventService) *TransferEventHandler {
	return &TransferEventHandler{svc: svc}
}

// CreateBatch 批量创建转账事件
func (h *TransferEventHandler) CreateBatch(c *gin.Context) {
	var body struct {
		Events []models.TransferEvent `json:"events"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid payload", "details": err.Error()})
		return
	}
	// 需要指针切片
	events := make([]*models.TransferEvent, 0, len(body.Events))
	for i := range body.Events {
		events = append(events, &body.Events[i])
	}
	if err := h.svc.CreateEvents(events); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "count": len(events)})
}
