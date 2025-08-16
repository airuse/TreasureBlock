package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
)

// ContractHandler 合约处理器
type ContractHandler struct {
	contractService services.ContractService
}

// NewContractHandler 创建合约处理器
func NewContractHandler(contractService services.ContractService) *ContractHandler {
	return &ContractHandler{
		contractService: contractService,
	}
}

// convertToJSONString 将切片或map转换为JSON字符串
func (h *ContractHandler) convertToJSONString(data interface{}) string {
	if data == nil {
		return ""
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(jsonData)
}

// CreateOrUpdateContract 创建或更新合约
func (h *ContractHandler) CreateOrUpdateContract(c *gin.Context) {
	var contractInfo dto.ContractInfo
	if err := c.ShouldBindJSON(&contractInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	contract, err := h.contractService.CreateOrUpdateContract(c.Request.Context(), &models.Contract{
		Address:      contractInfo.Address,
		ChainName:    contractInfo.ChainName,
		ContractType: contractInfo.ContractType,
		Name:         contractInfo.Name,
		Symbol:       contractInfo.Symbol,
		Decimals:     contractInfo.Decimals,
		TotalSupply:  contractInfo.TotalSupply,
		IsERC20:      contractInfo.IsERC20,
		// 添加缺失的字段
		Interfaces:    h.convertToJSONString(contractInfo.Interfaces),
		Methods:       h.convertToJSONString(contractInfo.Methods),
		Events:        h.convertToJSONString(contractInfo.Events),
		Metadata:      h.convertToJSONString(contractInfo.Metadata),
		Status:        contractInfo.Status,        // 使用DTO中的状态
		Verified:      contractInfo.Verified,      // 使用DTO中的验证状态
		Creator:       contractInfo.Creator,       // 使用DTO中的创建者
		CreationTx:    contractInfo.CreationTx,    // 使用DTO中的创建交易
		CreationBlock: contractInfo.CreationBlock, // 使用DTO中的创建区块
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create/update contract: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    contract,
		"message": "Contract created/updated successfully",
	})
}

// GetContractByAddress 根据地址获取合约
func (h *ContractHandler) GetContractByAddress(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Contract address is required",
		})
		return
	}

	contract, err := h.contractService.GetContractByAddress(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Contract not found: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    contract,
	})
}

// GetContractsByChain 根据链名称获取合约列表
func (h *ContractHandler) GetContractsByChain(c *gin.Context) {
	chainName := c.Param("chainName")
	if chainName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Chain name is required",
		})
		return
	}

	contracts, err := h.contractService.GetContractsByChain(c.Request.Context(), chainName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get contracts: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    contracts,
		"count":   len(contracts),
	})
}

// GetContractsByType 根据合约类型获取合约列表
func (h *ContractHandler) GetContractsByType(c *gin.Context) {
	contractType := c.Param("type")
	if contractType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Contract type is required",
		})
		return
	}

	contracts, err := h.contractService.GetContractsByType(c.Request.Context(), contractType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get contracts: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    contracts,
		"count":   len(contracts),
	})
}

// GetERC20Tokens 获取所有ERC-20代币合约
func (h *ContractHandler) GetERC20Tokens(c *gin.Context) {
	contracts, err := h.contractService.GetERC20Tokens(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get ERC-20 tokens: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    contracts,
		"count":   len(contracts),
	})
}

// GetAllContracts 获取所有合约
func (h *ContractHandler) GetAllContracts(c *gin.Context) {
	contracts, err := h.contractService.GetAllContracts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get contracts: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    contracts,
		"count":   len(contracts),
	})
}

// UpdateContractStatus 更新合约状态
func (h *ContractHandler) UpdateContractStatus(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Contract address is required",
		})
		return
	}

	statusStr := c.Param("status")
	status, err := strconv.ParseInt(statusStr, 10, 8)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid status value",
		})
		return
	}

	err = h.contractService.UpdateContractStatus(c.Request.Context(), address, int8(status))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update contract status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Contract status updated successfully",
	})
}

// VerifyContract 验证合约
func (h *ContractHandler) VerifyContract(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Contract address is required",
		})
		return
	}

	err := h.contractService.VerifyContract(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to verify contract: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Contract verified successfully",
	})
}

// DeleteContract 删除合约
func (h *ContractHandler) DeleteContract(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Contract address is required",
		})
		return
	}

	err := h.contractService.DeleteContract(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete contract: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Contract deleted successfully",
	})
}
