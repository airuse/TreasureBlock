package handlers

import (
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GasHandler Gas费率处理器
type GasHandler struct {
	feeScheduler *services.FeeScheduler
}

// NewGasHandler 创建Gas费率处理器
func NewGasHandler(feeScheduler *services.FeeScheduler) *GasHandler {
	return &GasHandler{
		feeScheduler: feeScheduler,
	}
}

// GetGasRates 获取Gas费率信息
// @Summary 获取Gas费率信息
// @Description 获取最新的Gas费率信息，包括慢速、普通、快速三个等级
// @Tags Gas费率
// @Produce json
// @Param chain query string false "链类型" Enums(eth,btc) default(eth)
// @Success 200 {object} utils.Response{data=services.FeeLevels} "获取成功"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/user/gas [get]
func (h *GasHandler) GetGasRates(c *gin.Context) {
	// 获取链类型参数，默认为eth
	chain := c.DefaultQuery("chain", "eth")

	// 验证链类型
	if chain != "eth" && chain != "btc" && chain != "bsc" && chain != "sol" {
		utils.ErrorResponse(c, http.StatusBadRequest, "不支持的链类型，仅支持eth、btc、bsc和sol", nil)
		return
	}

	// 获取缓存的费率数据
	feeData := h.feeScheduler.GetLastFeeData(chain)
	if feeData == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "暂无费率数据，请稍后重试", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取Gas费率成功", feeData)
}

// GetAllGasRates 获取所有链的Gas费率信息
// @Summary 获取所有链的Gas费率信息
// @Description 获取所有链的最新Gas费率信息
// @Tags Gas费率
// @Produce json
// @Success 200 {object} utils.Response{data=map[string]services.FeeLevels} "获取成功"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/user/gas/all [get]
func (h *GasHandler) GetAllGasRates(c *gin.Context) {
	// 获取所有链的费率数据
	allFeeData := h.feeScheduler.GetAllLastFeeData()
	if len(allFeeData) == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "暂无费率数据，请稍后重试", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取所有Gas费率成功", allFeeData)
}

// GetBTCGasRatesCached 获取BTC缓存费率信息（无鉴权初始加载用）
// @Summary 获取BTC缓存费率信息（无鉴权）
// @Description 读取调度器内存缓存的BTC费率，适合页面初次打开时快速展示
// @Tags Gas费率
// @Produce json
// @Success 200 {object} utils.Response{data=services.FeeLevels} "获取成功"
// @Failure 404 {object} utils.Response "暂无费率数据"
// @Router /api/no-auth/gas/btc [get]
func (h *GasHandler) GetBTCGasRatesCached(c *gin.Context) {
	feeData := h.feeScheduler.GetLastFeeData("btc")
	if feeData == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "暂无BTC费率数据，请稍后重试", nil)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "获取BTC缓存费率成功", feeData)
}

// GetSOLGasRatesCached 获取SOL缓存费率信息（无鉴权初始加载用）
// @Summary 获取SOL缓存费率信息（无鉴权）
// @Description 读取调度器内存缓存的SOL费率，适合页面初次打开时快速展示
// @Tags Gas费率
// @Produce json
// @Success 200 {object} utils.Response{data=services.FeeLevels} "获取成功"
// @Failure 404 {object} utils.Response "暂无费率数据"
// @Router /api/no-auth/gas/sol [get]
func (h *GasHandler) GetSOLGasRatesCached(c *gin.Context) {
	feeData := h.feeScheduler.GetLastFeeData("sol")
	if feeData == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "暂无SOL费率数据，请稍后重试", nil)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "获取SOL缓存费率成功", feeData)
}
