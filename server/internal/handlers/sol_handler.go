package handlers

import (
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SolHandler struct {
	svc services.SolService
}

func NewSolHandler(svc services.SolService) *SolHandler {
	return &SolHandler{svc: svc}
}

// CreateTxDetail 创建单笔Sol交易明细
func (h *SolHandler) CreateTxDetail(c *gin.Context) {
	var req dto.SolTxDetailCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid payload",
			"details": err.Error(),
		})
		return
	}

	response, err := h.svc.SaveTxDetail(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Solana transaction detail created successfully",
		"data":    response,
	})
}

// CreateTxDetailBatch 批量创建Sol交易明细
func (h *SolHandler) CreateTxDetailBatch(c *gin.Context) {
	var req dto.BatchSolDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid payload",
			"details": err.Error(),
		})
		return
	}

	response, err := h.svc.SaveTxDetailBatch(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	status := http.StatusCreated
	if !response.Success {
		status = http.StatusPartialContent
	}

	c.JSON(status, gin.H{
		"success": response.Success,
		"message": response.Message,
		"data":    response,
	})
}

// GetTxDetail 获取交易详情
func (h *SolHandler) GetTxDetail(c *gin.Context) {
	txID := c.Param("txId")
	if txID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "transaction ID is required",
		})
		return
	}

	response, err := h.svc.GetTxDetail(c.Request.Context(), txID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetTxEvents 获取交易事件
func (h *SolHandler) GetTxEvents(c *gin.Context) {
	txID := c.Param("txId")
	if txID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "transaction ID is required",
		})
		return
	}

	events, err := h.svc.GetTxEvents(c.Request.Context(), txID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    events,
		"count":   len(events),
	})
}

// GetTxInstructions 获取交易指令
func (h *SolHandler) GetTxInstructions(c *gin.Context) {
	txID := c.Param("txId")
	if txID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "transaction ID is required",
		})
		return
	}

	instructions, err := h.svc.GetTxInstructions(c.Request.Context(), txID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    instructions,
		"count":   len(instructions),
	})
}

// GetTxsBySlot 根据slot获取交易列表
func (h *SolHandler) GetTxsBySlot(c *gin.Context) {
	slotStr := c.Param("slot")
	slot, err := strconv.ParseUint(slotStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid slot number",
		})
		return
	}

	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	transactions, total, err := h.svc.GetTxsBySlot(c.Request.Context(), slot, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    transactions,
		"meta": gin.H{
			"slot":  slot,
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetSlotStats 获取slot统计信息
func (h *SolHandler) GetSlotStats(c *gin.Context) {
	slotStr := c.Param("slot")
	slot, err := strconv.ParseUint(slotStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid slot number",
		})
		return
	}

	stats, err := h.svc.GetSlotStats(c.Request.Context(), slot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// ListTxDetails 分页查询sol_tx_detail
func (h *SolHandler) ListTxDetails(c *gin.Context) {
	var slotPtr *uint64
	if s := c.Query("slot"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			slotPtr = &v
		}
	}
	page := 1
	if p := c.DefaultQuery("page", "1"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	pageSize := 20
	if ps := c.DefaultQuery("page_size", "20"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}
	list, total, err := h.svc.ListTxDetails(c.Request.Context(), slotPtr, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

// GetArtifactsByTxID 通过txid查询指令与事件
func (h *SolHandler) GetArtifactsByTxID(c *gin.Context) {
	txID := c.Param("txId")
	if txID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "txId is required"})
		return
	}
	data, err := h.svc.GetArtifactsByTxID(c.Request.Context(), txID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// Program maintenance endpoints

// CreateProgram 创建程序
func (h *SolHandler) CreateProgram(c *gin.Context) {
	var p models.SolProgram
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err := h.svc.CreateProgram(c.Request.Context(), &p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": p})
}

// UpdateProgram 更新程序
func (h *SolHandler) UpdateProgram(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}
	p, err := h.svc.GetProgramByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}
	var req models.SolProgram
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	// apply editable fields
	p.Name = req.Name
	p.Alias = req.Alias
	p.Category = req.Category
	p.Type = req.Type
	p.IsSystem = req.IsSystem
	p.Version = req.Version
	p.Status = req.Status
	p.Description = req.Description
	p.InstructionRules = req.InstructionRules
	p.EventRules = req.EventRules
	p.SampleData = req.SampleData
	if err := h.svc.UpdateProgram(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": p})
}

// DeleteProgram 删除程序
func (h *SolHandler) DeleteProgram(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}
	if err := h.svc.DeleteProgram(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetProgram 获取程序详情
func (h *SolHandler) GetProgram(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}
	p, err := h.svc.GetProgramByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": p})
}

// ListPrograms 列表
func (h *SolHandler) ListPrograms(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	keyword := c.DefaultQuery("keyword", "")
	list, total, err := h.svc.ListPrograms(c.Request.Context(), page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}
