package handlers

import (
	"net/http"
	"strconv"

	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// BlockVerificationHandler 区块验证处理器
type BlockVerificationHandler struct {
	verificationService services.BlockVerificationService
}

// NewBlockVerificationHandler 创建区块验证处理器
func NewBlockVerificationHandler(verificationService services.BlockVerificationService) *BlockVerificationHandler {
	return &BlockVerificationHandler{
		verificationService: verificationService,
	}
}

// VerifyBlock 验证区块接口
func (h *BlockVerificationHandler) VerifyBlock(c *gin.Context) {
	blockIDStr := c.Param("blockID")
	if blockIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "block ID is required",
		})
		return
	}

	blockID, err := strconv.ParseUint(blockIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid block ID format",
		})
		return
	}

	// 执行区块验证
	result, err := h.verificationService.VerifyBlock(c.Request.Context(), blockID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "BLOCK_VERIFICATION_FAILED",
			"details": err.Error(),
		})
		return
	}

	if !result.IsValid {
		// 如果验证不通过，则返回错误，让客户端结束
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "BLOCK_VERIFICATION_FAILED",
			"details": result.Details,
		})
		return
	}

	// 验证通过需要吧数据库 block 表的 verification_status 更新为 1
	h.verificationService.UpdateBlockVerificationStatus(c.Request.Context(), blockID, true, "验证通过")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"block_id":                    result.BlockID,
			"is_valid":                    result.IsValid,
			"reason":                      result.Reason,
			"details":                     result.Details,
			"transactions":                result.Transactions,
			"receipts":                    result.Receipts,
			"verification_status_updated": true,
		},
		"message": "区块验证完成",
	})
}

// GetLastVerifiedBlockHeight 获取最后一个验证通过的区块高度
func (h *BlockVerificationHandler) GetLastVerifiedBlockHeight(c *gin.Context) {
	chain := c.Query("chain")
	if chain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "chain parameter is required",
		})
		return
	}

	height, err := h.verificationService.GetLastVerifiedBlockHeight(c.Request.Context(), chain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "未找到验证通过的区块",
			"details": err.Error(),
		})
		return
	}

	// 获取最后一个验证通过的区块高度，然后判断是否已经超时，如果超时则将此高度hash后缀增加_loser,然后吧deleted_at逻辑删除掉
	if err := h.verificationService.HandleTimeoutBlocks(c.Request.Context(), chain, height+1); err != nil {
		// 记录错误但不影响正常流程，继续执行
		logrus.Errorf("HandleTimeoutBlocks error: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"chain":  chain,
			"height": height,
		},
		"message": "获取成功",
	})
}
