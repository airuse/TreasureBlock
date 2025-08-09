package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateStruct 验证结构体
func ValidateStruct(obj interface{}) error {
	return validate.Struct(obj)
}

// ValidateAndBind 验证并绑定JSON请求
func ValidateAndBind(c *gin.Context, obj interface{}) error {
	// 1. 绑定JSON
	if err := c.ShouldBindJSON(obj); err != nil {
		return err
	}

	// 2. 验证结构体
	return ValidateStruct(obj)
}

// HandleValidationError 处理验证错误
func HandleValidationError(c *gin.Context, err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errors []string
		for _, e := range validationErrors {
			errors = append(errors, formatValidationError(e))
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation failed",
			"details": errors,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
	}
}

// formatValidationError 格式化验证错误
func formatValidationError(e validator.FieldError) string {
	field := strings.ToLower(e.Field())
	switch e.Tag() {
	case "required":
		return field + " is required"
	case "min":
		return field + " must be at least " + e.Param() + " characters long"
	case "max":
		return field + " must be at most " + e.Param() + " characters long"
	case "gte":
		return field + " must be greater than or equal to " + e.Param()
	case "lte":
		return field + " must be less than or equal to " + e.Param()
	case "email":
		return field + " must be a valid email address"
	case "url":
		return field + " must be a valid URL"
	case "omitempty":
		return field + " cannot be empty when provided"
	default:
		return field + " validation failed: " + e.Tag()
	}
}
