package repository

import (
	"blockChainBrowser/server/internal/models"

	"gorm.io/gorm"
)

// UserAddressRepository 用户地址仓库接口
type UserAddressRepository interface {
	Create(userAddress *models.UserAddress) error
	GetByID(id uint) (*models.UserAddress, error)
	GetByUserID(userID uint) ([]models.UserAddress, error)
	GetByAddress(address string) (*models.UserAddress, error)
	Update(userAddress *models.UserAddress) error
	Delete(id uint) error
	GetActiveByUserID(userID uint) ([]models.UserAddress, error)
}

// userAddressRepository 用户地址仓库实现
type userAddressRepository struct {
	db *gorm.DB
}

// NewUserAddressRepository 创建用户地址仓库
func NewUserAddressRepository(db *gorm.DB) UserAddressRepository {
	return &userAddressRepository{db: db}
}

// Create 创建用户地址
func (r *userAddressRepository) Create(userAddress *models.UserAddress) error {
	return r.db.Create(userAddress).Error
}

// GetByID 根据ID获取用户地址
func (r *userAddressRepository) GetByID(id uint) (*models.UserAddress, error) {
	var userAddress models.UserAddress
	err := r.db.First(&userAddress, id).Error
	if err != nil {
		return nil, err
	}
	return &userAddress, nil
}

// GetByUserID 根据用户ID获取所有地址
func (r *userAddressRepository) GetByUserID(userID uint) ([]models.UserAddress, error) {
	var userAddresses []models.UserAddress
	err := r.db.Where("user_id = ?", userID).Find(&userAddresses).Error
	return userAddresses, err
}

// GetByAddress 根据地址获取用户地址
func (r *userAddressRepository) GetByAddress(address string) (*models.UserAddress, error) {
	var userAddress models.UserAddress
	err := r.db.Where("address = ?", address).First(&userAddress).Error
	if err != nil {
		return nil, err
	}
	return &userAddress, nil
}

// Update 更新用户地址
func (r *userAddressRepository) Update(userAddress *models.UserAddress) error {
	return r.db.Save(userAddress).Error
}

// Delete 删除用户地址
func (r *userAddressRepository) Delete(id uint) error {
	return r.db.Delete(&models.UserAddress{}, id).Error
}

// GetActiveByUserID 获取用户的所有活跃地址
func (r *userAddressRepository) GetActiveByUserID(userID uint) ([]models.UserAddress, error) {
	var userAddresses []models.UserAddress
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&userAddresses).Error
	return userAddresses, err
}
