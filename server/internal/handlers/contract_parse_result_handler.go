package handlers

import (
	"net/http"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
)

type ContractParseResultHandler struct {
	service services.ContractParseService
}

func NewContractParseResultHandler(service services.ContractParseService) *ContractParseResultHandler {
	return &ContractParseResultHandler{service: service}
}

// GetByTxHash 查询某交易的解析结果
func (h *ContractParseResultHandler) GetByTxHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "hash is required"})
		return
	}
	results, err := h.service.GetByTxHash(c.Request.Context(), hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	out := make([]*dto.ContractParseResultResponse, 0, len(results))
	for _, r := range results {
		out = append(out, dto.NewContractParseResultResponse(r))
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": out})
}
