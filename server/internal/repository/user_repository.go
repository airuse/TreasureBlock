package repository

import (
	"errors"
	"time"

	"blockChainBrowser/server/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	UpdateLastLogin(userID uint, loginTime time.Time) error
	Delete(id uint) error
	IsUsernameExists(username string) (bool, error)
	IsEmailExists(email string) (bool, error)
	// 扩展方法
	ExistsByUsername(username string) (bool, error)
	ExistsByEmail(email string) (bool, error)
	IsFirstUser() (bool, error)
	GetByRole(role string) (*models.User, error)
	GetFirstUser() (*models.User, error)
	UpdateUser(user *models.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) UpdateLastLogin(userID uint, loginTime time.Time) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("last_login", loginTime).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) IsUsernameExists(username string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

func (r *userRepository) IsEmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// GetByRole 根据角色获取用户
func (r *userRepository) GetByRole(role string) (*models.User, error) {
	var user models.User
	err := r.db.Where("role = ?", role).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetFirstUser 获取第一个用户
func (r *userRepository) GetFirstUser() (*models.User, error) {
	var user models.User
	err := r.db.Order("id ASC").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (r *userRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// IsFirstUser 检查是否是第一个用户
func (r *userRepository) IsFirstUser() (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count == 0, err
}
