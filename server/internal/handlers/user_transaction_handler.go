package handlers

import (
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserTransactionHandler 用户交易处理器
type UserTransactionHandler struct {
	userTxService services.UserTransactionService
}

// NewUserTransactionHandler 创建用户交易处理器实例
func NewUserTransactionHandler() *UserTransactionHandler {
	return &UserTransactionHandler{
		userTxService: services.NewUserTransactionService(),
	}
}

// CreateTransaction 创建用户交易
func (h *UserTransactionHandler) CreateTransaction(c *gin.Context) {
	var req dto.CreateUserTransactionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 获取用户ID（从JWT中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权", "用户ID不存在")
		return
	}

	// 验证交易相关字段
	if err := h.validateTransactionFields(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数验证失败", err.Error())
		return
	}

	// 创建交易
	response, err := h.userTxService.CreateTransaction(c.Request.Context(), uint64(userID.(uint)), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建交易失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "创建交易成功", response)
}

// validateTransactionFields 验证交易相关字段
func (h *UserTransactionHandler) validateTransactionFields(req *dto.CreateUserTransactionRequest) error {
	// 智能判断交易类型
	if req.TransactionType == "" {
		// 根据代币符号和合约地址智能判断
		if req.Symbol != "ETH" && req.Symbol != "BTC" {
			// 如果不是原生代币，自动设置为代币交易
			req.TransactionType = "token"
		} else {
			req.TransactionType = "coin"
		}
	}

	// 如果选择了代币交易类型，验证相关字段
	if req.TransactionType == "token" {
		// 代币交易必须指定合约操作类型
		if req.ContractOperationType == "" {
			// 根据用户填写的情况智能设置默认操作类型
			if req.ToAddress != "" && req.Amount != "" {
				req.ContractOperationType = "transfer" // 有接收地址和金额，默认为转账
			} else if req.ToAddress != "" && req.Amount == "" {
				req.ContractOperationType = "balanceOf" // 只有接收地址，默认为查询余额
			} else {
				return fmt.Errorf("代币交易需要指定合约操作类型或填写完整的交易信息")
			}
		}

		// 根据操作类型验证必需字段
		switch req.ContractOperationType {
		case "transfer":
			if req.FromAddress == "" || req.ToAddress == "" || req.Amount == "" {
				return fmt.Errorf("转账操作需要发送地址、接收地址和金额")
			}
		case "approve":
			if req.FromAddress == "" || req.ToAddress == "" || req.Amount == "" {
				return fmt.Errorf("授权操作需要授权者地址、被授权者地址和授权额度")
			}
		case "transferFrom":
			if req.FromAddress == "" || req.ToAddress == "" || req.Amount == "" {
				return fmt.Errorf("授权转账操作需要发送地址、接收地址和金额")
			}
		case "balanceOf":
			if req.FromAddress == "" {
				return fmt.Errorf("查询余额操作需要查询地址")
			}
			// 查询余额不需要金额
			req.Amount = "0"
		default:
			return fmt.Errorf("不支持的合约操作类型: %s", req.ContractOperationType)
		}
	} else {
		// 原生代币转账的验证（ETH、BTC等）
		if req.FromAddress == "" || req.ToAddress == "" || req.Amount == "" {
			return fmt.Errorf("原生代币转账需要发送地址、接收地址和金额")
		}
		// 验证金额格式（必须是有效的整数）
		if err := h.validateAmountFormat(req.Amount); err != nil {
			return fmt.Errorf("金额格式错误: %v", err)
		}
		// 手续费可以为0，但必须提供
		if req.Fee == "" {
			req.Fee = "0"
		}
		// 验证手续费格式（必须是有效的整数）
		if err := h.validateAmountFormat(req.Fee); err != nil {
			return fmt.Errorf("手续费格式错误: %v", err)
		}
	}

	return nil
}

// validateAmountFormat 验证金额格式（必须是有效的整数）
func (h *UserTransactionHandler) validateAmountFormat(amount string) error {
	if amount == "" {
		return fmt.Errorf("金额不能为空")
	}

	// 尝试将字符串转换为大整数
	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return fmt.Errorf("金额必须是有效的整数格式")
	}

	// 检查金额是否为正数
	if amountBig.Sign() < 0 {
		return fmt.Errorf("金额不能为负数")
	}

	// 检查金额是否超过最大值（65位十进制数）
	maxAmount := new(big.Int)
	maxAmount.SetString("99999999999999999999999999999999999999999999999999999999999999999", 10)
	if amountBig.Cmp(maxAmount) > 0 {
		return fmt.Errorf("金额超过最大值")
	}

	return nil
}

// GetTransactionByID 根据ID获取用户交易
func (h *UserTransactionHandler) GetTransactionByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误", "无效的交易ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权", "用户ID不存在")
		return
	}

	// 获取交易
	response, err := h.userTxService.GetTransactionByID(c.Request.Context(), uint(id), uint64(userID.(uint)))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "获取交易失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取交易成功", response)
}

// GetUserTransactions 获取用户交易列表
func (h *UserTransactionHandler) GetUserTransactions(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	chain := c.Query("chain")

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权", "用户ID不存在")
		return
	}

	// 根据是否指定链类型选择不同的查询方法
	var response *dto.UserTransactionListResponse
	var err error

	if chain != "" {
		// 按链类型查询
		response, err = h.userTxService.GetUserTransactionsByChain(c.Request.Context(), uint64(userID.(uint)), chain, page, pageSize, status)
	} else {
		// 查询所有链的交易
		response, err = h.userTxService.GetUserTransactions(c.Request.Context(), uint64(userID.(uint)), page, pageSize, status)
	}

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取交易列表失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取交易列表成功", response)
}

// UpdateTransaction 更新用户交易
func (h *UserTransactionHandler) UpdateTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误", "无效的交易ID")
		return
	}

	var req dto.UpdateUserTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权", "用户ID不存在")
		return
	}

	// 验证更新请求的字段
	if err := h.validateUpdateTransactionFields(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数验证失败", err.Error())
		return
	}

	// 更新交易
	response, err := h.userTxService.UpdateTransaction(c.Request.Context(), uint(id), uint64(userID.(uint)), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新交易失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "更新交易成功", response)
}

// validateUpdateTransactionFields 验证更新交易字段
func (h *UserTransactionHandler) validateUpdateTransactionFields(req *dto.UpdateUserTransactionRequest) error {
	// 验证金额格式（如果提供了金额）
	if req.Amount != nil {
		if err := h.validateAmountFormat(*req.Amount); err != nil {
			return fmt.Errorf("金额格式错误: %v", err)
		}
	}

	// 验证手续费格式（如果提供了手续费）
	if req.Fee != nil {
		if err := h.validateAmountFormat(*req.Fee); err != nil {
			return fmt.Errorf("手续费格式错误: %v", err)
		}
	}

	// 验证代币交易相关字段
	if req.TransactionType != nil && *req.TransactionType == "token" {
		if req.ContractOperationType != nil {
			validOps := []string{"transfer", "approve", "transferFrom", "balanceOf"}
			isValid := false
			for _, op := range validOps {
				if *req.ContractOperationType == op {
					isValid = true
					break
				}
			}
			if !isValid {
				return fmt.Errorf("不支持的合约操作类型: %s", *req.ContractOperationType)
			}
		}
	}

	// 验证BTC交易相关字段
	if len(req.BTCTxIn) > 0 {
		for i, txIn := range req.BTCTxIn {
			if txIn.TxID == "" {
				return fmt.Errorf("BTC TxIn[%d] 缺少 txid", i)
			}
		}
	}

	if len(req.BTCTxOut) > 0 {
		for i, txOut := range req.BTCTxOut {
			if txOut.Address == "" {
				return fmt.Errorf("BTC TxOut[%d] 缺少 address", i)
			}
			if txOut.ValueSatoshi == 0 {
				return fmt.Errorf("BTC TxOut[%d] 金额不能为0", i)
			}
		}
	}

	return nil
}

// DeleteTransaction 删除用户交易
func (h *UserTransactionHandler) DeleteTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误", "无效的交易ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权", "用户ID不存在")
		return
	}

	// 删除交易
	if err := h.userTxService.DeleteTransaction(c.Request.Context(), uint(id), uint64(userID.(uint))); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除交易失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "删除交易成功", nil)
}

// ExportTransaction 导出交易
func (h *UserTransactionHandler) ExportTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误", "无效的交易ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权", "用户ID不存在")
		return
	}

	// 解析请求体，获取费率设置
	var req dto.ExportTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 导出交易
	response, err := h.userTxService.ExportTransaction(c.Request.Context(), uint(id), uint64(userID.(uint)), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "导出交易失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "导出交易成功", response)
}

// ImportSignature 导入签名
func (h *UserTransactionHandler) ImportSignature(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误", "无效的交易ID")
		return
	}

	var req dto.ImportSignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权", "用户ID不存在")
		return
	}

	// 导入签名
	response, err := h.userTxService.ImportSignature(c.Request.Context(), uint(id), uint64(userID.(uint)), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "导入签名失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "导入签名成功", response)
}

// SendTransaction 发送交易
func (h *UserTransactionHandler) SendTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误", "无效的交易ID")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权", "用户ID不存在")
		return
	}

	// 发送交易
	response, err := h.userTxService.SendTransaction(c.Request.Context(), uint(id), uint64(userID.(uint)))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "发送交易失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "发送交易成功", response)
}

// GetUserTransactionStats 获取用户交易统计
func (h *UserTransactionHandler) GetUserTransactionStats(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权", "用户ID不存在")
		return
	}

	// 获取链类型参数（可选）
	chain := c.Query("chain")

	// 获取统计信息
	response, err := h.userTxService.GetUserTransactionStats(c.Request.Context(), uint64(userID.(uint)), chain)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取统计信息失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取统计信息成功", response)
}
