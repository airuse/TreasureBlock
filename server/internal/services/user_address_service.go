package services

import (
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"errors"
	"strings"
)

// UserAddressService 用户地址服务接口
type UserAddressService interface {
	CreateAddress(userID uint, req *dto.CreateUserAddressRequest) (*dto.UserAddressResponse, error)
	GetUserAddresses(userID uint) ([]dto.UserAddressResponse, error)
	UpdateAddress(userID uint, addressID uint, req *dto.UpdateUserAddressRequest) (*dto.UserAddressResponse, error)
	DeleteAddress(userID uint, addressID uint) error
	GetAddressByID(userID uint, addressID uint) (*dto.UserAddressResponse, error)
}

// userAddressService 用户地址服务实现
type userAddressService struct {
	userAddressRepo repository.UserAddressRepository
}

// NewUserAddressService 创建用户地址服务
func NewUserAddressService(userAddressRepo repository.UserAddressRepository) UserAddressService {
	return &userAddressService{
		userAddressRepo: userAddressRepo,
	}
}

// CreateAddress 创建用户地址
func (s *userAddressService) CreateAddress(userID uint, req *dto.CreateUserAddressRequest) (*dto.UserAddressResponse, error) {
	// 验证地址格式
	if !s.isValidAddress(req.Address) {
		return nil, errors.New("无效的地址格式")
	}

	// 检查地址是否已存在
	existingAddress, err := s.userAddressRepo.GetByAddress(req.Address)
	if err == nil && existingAddress != nil {
		return nil, errors.New("该地址已存在")
	}

	// 创建新地址
	userAddress := &models.UserAddress{
		UserID:           userID,
		Address:          req.Address,
		Label:            req.Label,
		Type:             req.Type,
		Balance:          0,
		TransactionCount: 0,
		IsActive:         true,
	}

	if err := s.userAddressRepo.Create(userAddress); err != nil {
		return nil, err
	}

	return s.convertToResponse(userAddress), nil
}

// GetUserAddresses 获取用户的所有地址
func (s *userAddressService) GetUserAddresses(userID uint) ([]dto.UserAddressResponse, error) {
	addresses, err := s.userAddressRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	var responses []dto.UserAddressResponse
	for _, addr := range addresses {
		responses = append(responses, *s.convertToResponse(&addr))
	}

	return responses, nil
}

// UpdateAddress 更新用户地址
func (s *userAddressService) UpdateAddress(userID uint, addressID uint, req *dto.UpdateUserAddressRequest) (*dto.UserAddressResponse, error) {
	// 获取地址
	address, err := s.userAddressRepo.GetByID(addressID)
	if err != nil {
		return nil, errors.New("地址不存在")
	}

	// 验证权限
	if address.UserID != userID {
		return nil, errors.New("无权限操作此地址")
	}

	// 更新字段
	if req.Label != nil {
		address.Label = *req.Label
	}
	if req.Type != nil {
		address.Type = *req.Type
	}
	if req.IsActive != nil {
		address.IsActive = *req.IsActive
	}

	// 保存更新
	if err := s.userAddressRepo.Update(address); err != nil {
		return nil, err
	}

	return s.convertToResponse(address), nil
}

// DeleteAddress 删除用户地址
func (s *userAddressService) DeleteAddress(userID uint, addressID uint) error {
	// 获取地址
	address, err := s.userAddressRepo.GetByID(addressID)
	if err != nil {
		return errors.New("地址不存在")
	}

	// 验证权限
	if address.UserID != userID {
		return errors.New("无权限操作此地址")
	}

	return s.userAddressRepo.Delete(addressID)
}

// GetAddressByID 根据ID获取地址
func (s *userAddressService) GetAddressByID(userID uint, addressID uint) (*dto.UserAddressResponse, error) {
	address, err := s.userAddressRepo.GetByID(addressID)
	if err != nil {
		return nil, errors.New("地址不存在")
	}

	// 验证权限
	if address.UserID != userID {
		return nil, errors.New("无权限查看此地址")
	}

	return s.convertToResponse(address), nil
}

// isValidAddress 验证地址格式
func (s *userAddressService) isValidAddress(address string) bool {
	// 简单的以太坊地址验证
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	if len(address) != 42 {
		return false
	}
	// 可以添加更多验证逻辑
	return true
}

// convertToResponse 转换为响应DTO
func (s *userAddressService) convertToResponse(address *models.UserAddress) *dto.UserAddressResponse {
	return &dto.UserAddressResponse{
		ID:               address.ID,
		Address:          address.Address,
		Label:            address.Label,
		Type:             address.Type,
		Balance:          address.Balance,
		TransactionCount: address.TransactionCount,
		IsActive:         address.IsActive,
		CreatedAt:        address.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        address.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
