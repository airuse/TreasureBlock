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

	// 将数组/对象转换为JSON字符串存储到数据库
	interfacesStr := h.convertToJSONString(contractInfo.Interfaces)
	methodsStr := h.convertToJSONString(contractInfo.Methods)
	eventsStr := h.convertToJSONString(contractInfo.Events)
	metadataStr := h.convertToJSONString(contractInfo.Metadata)

	contract, err := h.contractService.CreateOrUpdateContract(c.Request.Context(), &models.Contract{
		Address:      contractInfo.Address,
		ChainName:    contractInfo.ChainName,
		ContractType: contractInfo.ContractType,
		Name:         contractInfo.Name,
		Symbol:       contractInfo.Symbol,
		Decimals:     contractInfo.Decimals,
		TotalSupply:  contractInfo.TotalSupply,
		IsERC20:      contractInfo.IsERC20,
		// 转换为JSON字符串存储
		Interfaces:    interfacesStr,
		Methods:       methodsStr,
		Events:        eventsStr,
		Metadata:      metadataStr,
		Status:        contractInfo.Status,        // 使用DTO中的状态
		Verified:      contractInfo.Verified,      // 使用DTO中的验证状态
		Creator:       contractInfo.Creator,       // 使用DTO中的创建者
		CreationTx:    contractInfo.CreationTx,    // 使用DTO中的创建交易
		CreationBlock: contractInfo.CreationBlock, // 使用DTO中的创建区块
		ContractLogo:  contractInfo.ContractLogo,  // 使用DTO中的合约Logo
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

	// 将JSON字符串转换回数组/对象返回给前端
	responseData := gin.H{
		"id":             contract.ID,
		"address":        contract.Address,
		"chain_name":     contract.ChainName,
		"contract_type":  contract.ContractType,
		"name":           contract.Name,
		"symbol":         contract.Symbol,
		"decimals":       contract.Decimals,
		"total_supply":   contract.TotalSupply,
		"is_erc20":       contract.IsERC20,
		"status":         contract.Status,
		"verified":       contract.Verified,
		"creator":        contract.Creator,
		"creation_tx":    contract.CreationTx,
		"creation_block": contract.CreationBlock,
		"contract_logo":  contract.ContractLogo,
		"ctime":          contract.CTime,
		"mtime":          contract.MTime,
	}

	// 解析JSON字符串为数组/对象
	if contract.Interfaces != "" {
		var interfaces interface{}
		if err := json.Unmarshal([]byte(contract.Interfaces), &interfaces); err == nil {
			responseData["interfaces"] = interfaces
		} else {
			responseData["interfaces"] = []string{}
		}
	} else {
		responseData["interfaces"] = []string{}
	}

	if contract.Methods != "" {
		var methods interface{}
		if err := json.Unmarshal([]byte(contract.Methods), &methods); err == nil {
			responseData["methods"] = methods
		} else {
			responseData["methods"] = []string{}
		}
	} else {
		responseData["methods"] = []string{}
	}

	if contract.Events != "" {
		var events interface{}
		if err := json.Unmarshal([]byte(contract.Events), &events); err == nil {
			responseData["events"] = events
		} else {
			responseData["events"] = []string{}
		}
	} else {
		responseData["events"] = []string{}
	}

	if contract.Metadata != "" {
		var metadata interface{}
		if err := json.Unmarshal([]byte(contract.Metadata), &metadata); err == nil {
			responseData["metadata"] = metadata
		} else {
			responseData["metadata"] = map[string]string{}
		}
	} else {
		responseData["metadata"] = map[string]string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responseData,
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

// GetAllContracts 获取所有合约（支持过滤和分页）
func (h *ContractHandler) GetAllContracts(c *gin.Context) {
	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "25")
	chainName := c.Query("chainName")
	contractType := c.Query("contractType")
	status := c.Query("status")
	search := c.Query("search")

	// 转换分页参数
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 25
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if chainName != "" {
		filters["chainName"] = chainName
	}
	if contractType != "" {
		filters["contractType"] = contractType
	}
	if status != "" {
		filters["status"] = status
	}
	if search != "" {
		filters["search"] = search
	}

	// 使用过滤服务获取合约
	contracts, total, err := h.contractService.GetContractsWithFilters(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get contracts: " + err.Error(),
		})
		return
	}

	// 转换每个合约的数据格式
	var responseContracts []gin.H
	for _, contract := range contracts {
		contractData := gin.H{
			"id":             contract.ID,
			"address":        contract.Address,
			"chain_name":     contract.ChainName,
			"contract_type":  contract.ContractType,
			"name":           contract.Name,
			"symbol":         contract.Symbol,
			"decimals":       contract.Decimals,
			"total_supply":   contract.TotalSupply,
			"is_erc20":       contract.IsERC20,
			"status":         contract.Status,
			"verified":       contract.Verified,
			"creator":        contract.Creator,
			"creation_tx":    contract.CreationTx,
			"creation_block": contract.CreationBlock,
			"contract_logo":  contract.ContractLogo,
			"ctime":          contract.CTime,
			"mtime":          contract.MTime,
		}

		// 解析JSON字符串为数组/对象
		if contract.Interfaces != "" {
			var interfaces interface{}
			if err := json.Unmarshal([]byte(contract.Interfaces), &interfaces); err == nil {
				contractData["interfaces"] = interfaces
			} else {
				contractData["interfaces"] = []string{}
			}
		} else {
			contractData["interfaces"] = []string{}
		}

		if contract.Methods != "" {
			var methods interface{}
			if err := json.Unmarshal([]byte(contract.Methods), &methods); err == nil {
				contractData["methods"] = methods
			} else {
				contractData["methods"] = []string{}
			}
		} else {
			contractData["methods"] = []string{}
		}

		if contract.Events != "" {
			var events interface{}
			if err := json.Unmarshal([]byte(contract.Events), &events); err == nil {
				contractData["events"] = events
			} else {
				contractData["events"] = []string{}
			}
		} else {
			contractData["events"] = []string{}
		}

		if contract.Metadata != "" {
			var metadata interface{}
			if err := json.Unmarshal([]byte(contract.Metadata), &metadata); err == nil {
				contractData["metadata"] = metadata
			} else {
				contractData["metadata"] = map[string]string{}
			}
		} else {
			contractData["metadata"] = map[string]string{}
		}

		responseContracts = append(responseContracts, contractData)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responseContracts,
		"count":   total, // 使用过滤后的总数
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
