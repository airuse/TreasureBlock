package repository

import (
	"blockChainBrowser/server/internal/models"

	"gorm.io/gorm"
)

// RoleRepository 角色仓库接口
type RoleRepository interface {
	CreateRole(role *models.Role) error
	GetRoleByID(id uint) (*models.Role, error)
	GetRoleByName(name string) (*models.Role, error)
	GetAllRoles() ([]models.Role, error)
	UpdateRole(role *models.Role) error
	DeleteRole(id uint) error
	GetRolesByUserID(userID uint) ([]models.Role, error)
}

// roleRepository 角色仓库实现
type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色仓库实例
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

// CreateRole 创建角色
func (r *roleRepository) CreateRole(role *models.Role) error {
	return r.db.Create(role).Error
}

// GetRoleByID 根据ID获取角色
func (r *roleRepository) GetRoleByID(id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByName 根据名称获取角色
func (r *roleRepository) GetRoleByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetAllRoles 获取所有角色
func (r *roleRepository) GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

// UpdateRole 更新角色
func (r *roleRepository) UpdateRole(role *models.Role) error {
	return r.db.Save(role).Error
}

// DeleteRole 删除角色
func (r *roleRepository) DeleteRole(id uint) error {
	return r.db.Delete(&models.Role{}, id).Error
}

// GetRolesByUserID 获取用户的所有角色
func (r *roleRepository) GetRolesByUserID(userID uint) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Preload("Permissions").
		Find(&roles).Error
	return roles, err
}
