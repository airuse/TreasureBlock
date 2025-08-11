package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SimpleSuccessResponse 简单成功响应（保持向后兼容）
func SimpleSuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// ErrorResponse 错误响应（支持详细错误信息）
func ErrorResponse(c *gin.Context, statusCode int, message string, details ...interface{}) {
	response := Response{
		Success: false,
		Error:   message,
	}
	
	// 如果有详细信息，将其作为数据返回
	if len(details) > 0 && details[0] != nil {
		response.Data = details[0]
	}
	
	c.JSON(statusCode, response)
}

// SimpleErrorResponse 简单错误响应（保持向后兼容）
func SimpleErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   message,
	})
}

// BadRequestResponse 400错误响应
func BadRequestResponse(c *gin.Context, message string) {
	SimpleErrorResponse(c, http.StatusBadRequest, message)
}

// NotFoundResponse 404错误响应
func NotFoundResponse(c *gin.Context, message string) {
	SimpleErrorResponse(c, http.StatusNotFound, message)
}

// InternalServerErrorResponse 500错误响应
func InternalServerErrorResponse(c *gin.Context, message string) {
	SimpleErrorResponse(c, http.StatusInternalServerError, message)
}

// ValidationErrorResponse 验证错误响应
func ValidationErrorResponse(c *gin.Context, message string) {
	SimpleErrorResponse(c, http.StatusUnprocessableEntity, message)
}
