package handlers

import (
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SolHandler struct{ svc services.SolService }

func NewSolHandler(svc services.SolService) *SolHandler { return &SolHandler{svc: svc} }

// CreateTxDetail 创建单笔Sol交易明细以及其指令（一次写入）
func (h *SolHandler) CreateTxDetail(c *gin.Context) {
	var req struct {
		Detail       models.SolTxDetail      `json:"detail"`
		Instructions []models.SolInstruction `json:"instructions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid payload", "details": err.Error()})
		return
	}
	// 转为指针切片
	ins := make([]*models.SolInstruction, 0, len(req.Instructions))
	for i := range req.Instructions {
		ins = append(ins, &req.Instructions[i])
	}
	if err := h.svc.SaveTxDetail(&req.Detail, ins); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true})
}
