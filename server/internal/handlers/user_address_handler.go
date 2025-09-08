package handlers

import (
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserAddressHandler 用户地址处理器
type UserAddressHandler struct {
	userAddressService services.UserAddressService
}

// NewUserAddressHandler 创建用户地址处理器
func NewUserAddressHandler(userAddressService services.UserAddressService) *UserAddressHandler {
	return &UserAddressHandler{
		userAddressService: userAddressService,
	}
}

// CreateAddress 创建用户地址
// @Summary 创建用户地址
// @Description 创建新的用户地址
// @Tags 用户地址管理
// @Accept json
// @Produce json
// @Param request body dto.CreateUserAddressRequest true "地址信息"
// @Success 200 {object} utils.Response{data=dto.UserAddressResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/user/addresses [post]
func (h *UserAddressHandler) CreateAddress(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	var req dto.CreateUserAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}
	address, err := h.userAddressService.CreateAddress(userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "地址创建成功", address)
}

// GetUserAddresses 获取用户地址列表
// @Summary 获取用户地址列表
// @Description 获取当前用户的所有地址
// @Tags 用户地址管理
// @Produce json
// @Success 200 {object} utils.Response{data=[]dto.UserAddressResponse}
// @Failure 401 {object} utils.Response
// @Router /api/user/addresses [get]
func (h *UserAddressHandler) GetUserAddresses(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	addresses, err := h.userAddressService.GetUserAddresses(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取地址列表失败")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取地址列表成功", addresses)
}

// UpdateAddress 更新用户地址
// @Summary 更新用户地址
// @Description 更新指定地址的信息
// @Tags 用户地址管理
// @Accept json
// @Produce json
// @Param id path int true "地址ID"
// @Param request body dto.UpdateUserAddressRequest true "更新信息"
// @Success 200 {object} utils.Response{data=dto.UserAddressResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/user/addresses/{id} [put]
func (h *UserAddressHandler) UpdateAddress(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	addressIDStr := c.Param("id")
	addressID, err := strconv.ParseUint(addressIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的地址ID")
		return
	}

	var req dto.UpdateUserAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	address, err := h.userAddressService.UpdateAddress(userID, uint(addressID), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "地址更新成功", address)
}

// DeleteAddress 删除用户地址
// @Summary 删除用户地址
// @Description 删除指定的用户地址
// @Tags 用户地址管理
// @Produce json
// @Param id path int true "地址ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/user/addresses/{id} [delete]
func (h *UserAddressHandler) DeleteAddress(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	addressIDStr := c.Param("id")
	addressID, err := strconv.ParseUint(addressIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的地址ID")
		return
	}

	if err := h.userAddressService.DeleteAddress(userID, uint(addressID)); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "地址删除成功", nil)
}

// GetAddressByID 根据ID获取地址
// @Summary 获取地址详情
// @Description 根据ID获取指定地址的详细信息
// @Tags 用户地址管理
// @Produce json
// @Param id path int true "地址ID"
// @Success 200 {object} utils.Response{data=dto.UserAddressResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/user/addresses/{id} [get]
func (h *UserAddressHandler) GetAddressByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	addressIDStr := c.Param("id")
	addressID, err := strconv.ParseUint(addressIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的地址ID")
		return
	}

	address, err := h.userAddressService.GetAddressByID(userID, uint(addressID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取地址详情成功", address)
}

// GetAddressTransactions 获取地址相关的交易列表
// @Summary 获取地址交易列表
// @Description 获取指定地址相关的所有交易
// @Tags 用户地址管理
// @Produce json
// @Param address query string true "地址"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页大小，默认20，最大100" default(20)
// @Param chain query string false "链类型（eth/btc）"
// @Success 200 {object} utils.Response{data=dto.AddressTransactionsResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/user/addresses/transactions [get]
func (h *UserAddressHandler) GetAddressTransactions(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	var req dto.GetAddressTransactionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 获取交易列表
	transactions, err := h.userAddressService.GetAddressTransactions(userID, req.Address, req.Page, req.PageSize, req.Chain)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取交易列表成功", transactions)
}

// GetAuthorizedAddresses 根据发送地址查询授权关系
// @Summary 查询授权关系
// @Description 根据发送地址查询可操作的代币持有者地址
// @Tags 用户地址管理
// @Produce json
// @Param spender_address query string true "发送地址（被授权地址）"
// @Success 200 {object} utils.Response{data=[]dto.UserAddressResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/user/addresses/authorized [get]
func (h *UserAddressHandler) GetAuthorizedAddresses(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	spenderAddress := c.Query("spender_address")
	if spenderAddress == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "spender_address参数不能为空")
		return
	}

	// 查询授权关系
	addresses, err := h.userAddressService.GetAddressesByAuthorizedAddress(spenderAddress)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询授权关系失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "查询授权关系成功", addresses)
}

// RefreshBalance 刷新地址及授权余额
func (h *UserAddressHandler) RefreshBalance(c *gin.Context) {
	userID := c.GetUint("user_id")
	var id uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		c.JSON(400, gin.H{"success": false, "message": "无效的地址ID"})
		return
	}
	resp, err := h.userAddressService.RefreshAddressBalances(userID, id)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": resp})
}
