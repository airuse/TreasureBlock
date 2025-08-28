package handlers

import (
	"net/http"
	"strconv"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/middleware"
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// EarningsHandler 收益处理器
type EarningsHandler struct {
	earningsService services.EarningsService
	logger          *logrus.Logger
}

// NewEarningsHandler 创建收益处理器
func NewEarningsHandler(earningsService services.EarningsService) *EarningsHandler {
	return &EarningsHandler{
		earningsService: earningsService,
		logger:          logrus.New(),
	}
}

// GetUserBalance 获取用户余额
func (h *EarningsHandler) GetUserBalance(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.SimpleErrorResponse(c, http.StatusUnauthorized, "用户认证失败")
		return
	}

	balance, err := h.earningsService.GetUserBalance(c.Request.Context(), uint64(userID))
	if err != nil {
		h.logger.Errorf("Failed to get user balance: %v", err)
		utils.SimpleErrorResponse(c, http.StatusInternalServerError, "获取余额失败")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取余额成功", balance)
}

// GetUserEarningsRecords 获取用户收益记录列表
func (h *EarningsHandler) GetUserEarningsRecords(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.SimpleErrorResponse(c, http.StatusUnauthorized, "用户认证失败")
		return
	}

	var req dto.EarningsRecordListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.SimpleErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	records, total, err := h.earningsService.GetUserEarningsRecords(c.Request.Context(), uint64(userID), &req)
	if err != nil {
		h.logger.Errorf("Failed to get user earnings records: %v", err)
		utils.SimpleErrorResponse(c, http.StatusInternalServerError, "获取收益记录失败")
		return
	}

	response := gin.H{
		"records": records,
		"pagination": gin.H{
			"page":      req.Page,
			"page_size": req.PageSize,
			"total":     total,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "获取收益记录成功", response)
}

// GetUserEarningsStats 获取用户收益统计
func (h *EarningsHandler) GetUserEarningsStats(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.SimpleErrorResponse(c, http.StatusUnauthorized, "用户认证失败")
		return
	}

	stats, err := h.earningsService.GetUserEarningsStats(c.Request.Context(), uint64(userID))
	if err != nil {
		h.logger.Errorf("Failed to get user earnings stats: %v", err)
		utils.SimpleErrorResponse(c, http.StatusInternalServerError, "获取收益统计失败")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取收益统计成功", stats)
}

// GetEarningsTrend 获取收益趋势数据
func (h *EarningsHandler) GetEarningsTrend(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.SimpleErrorResponse(c, http.StatusUnauthorized, "用户认证失败")
		return
	}

	// 获取查询参数
	hoursStr := c.DefaultQuery("hours", "2")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours < 1 || hours > 24 {
		hours = 2 // 默认2小时
	}

	trendData, err := h.earningsService.GetEarningsTrend(c.Request.Context(), uint64(userID), hours)
	if err != nil {
		h.logger.Errorf("Failed to get earnings trend: %v", err)
		utils.SimpleErrorResponse(c, http.StatusInternalServerError, "获取收益趋势失败")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取收益趋势成功", trendData)
}

// TransferTCoins 转账T币
func (h *EarningsHandler) TransferTCoins(c *gin.Context) {
	fromUserID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.SimpleErrorResponse(c, http.StatusUnauthorized, "用户认证失败")
		return
	}

	var req dto.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SimpleErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 验证转账参数
	if req.Amount <= 0 {
		utils.SimpleErrorResponse(c, http.StatusBadRequest, "转账金额必须大于0")
		return
	}

	if uint64(fromUserID) == req.ToUserID {
		utils.SimpleErrorResponse(c, http.StatusBadRequest, "不能转账给自己")
		return
	}

	result, err := h.earningsService.TransferTCoins(c.Request.Context(), uint64(fromUserID), req.ToUserID, req.Amount, req.Description)
	if err != nil {
		h.logger.Errorf("Failed to transfer T-coins: %v", err)
		if err.Error() == "insufficient balance" {
			utils.SimpleErrorResponse(c, http.StatusBadRequest, "余额不足")
			return
		}
		utils.SimpleErrorResponse(c, http.StatusInternalServerError, "转账失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "转账成功", result)
}

// GetEarningsRecordDetail 获取收益记录详情
func (h *EarningsHandler) GetEarningsRecordDetail(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.SimpleErrorResponse(c, http.StatusUnauthorized, "用户认证失败")
		return
	}

	recordIDStr := c.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 64)
	if err != nil {
		utils.SimpleErrorResponse(c, http.StatusBadRequest, "无效的记录ID")
		return
	}

	// 这里可以添加获取单个记录详情的逻辑
	// 目前简化处理，返回基础信息
	utils.SuccessResponse(c, http.StatusOK, "获取记录详情成功", gin.H{
		"record_id": recordID,
		"user_id":   userID,
		"message":   "功能开发中",
	})
}
