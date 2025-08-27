package repository

import (
	"blockChainBrowser/server/internal/models"

	"gorm.io/gorm"
)

// PermissionRepository 权限仓库接口
type PermissionRepository interface {
	CreatePermission(permission *models.Permission) error
	GetPermissionByID(id uint) (*models.Permission, error)
	GetPermissionByName(name string) (*models.Permission, error)
	GetAllPermissions() ([]models.Permission, error)
	GetPermissionsByRole(roleName string) ([]models.Permission, error)
	UpdatePermission(permission *models.Permission) error
	DeletePermission(id uint) error
	GetPermissionsByResource(resource string) ([]models.Permission, error)
}

// permissionRepository 权限仓库实现
type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建权限仓库实例
func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

// CreatePermission 创建权限
func (r *permissionRepository) CreatePermission(permission *models.Permission) error {
	return r.db.Create(permission).Error
}

// GetPermissionByID 根据ID获取权限
func (r *permissionRepository) GetPermissionByID(id uint) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.First(&permission, id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetPermissionByName 根据名称获取权限
func (r *permissionRepository) GetPermissionByName(name string) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.Where("name = ?", name).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetAllPermissions 获取所有权限
func (r *permissionRepository) GetAllPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Find(&permissions).Error
	return permissions, err
}

// GetPermissionsByRole 根据角色获取权限
func (r *permissionRepository) GetPermissionsByRole(roleName string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN roles ON role_permissions.role_id = roles.id").
		Where("roles.name = ?", roleName).
		Find(&permissions).Error
	return permissions, err
}

// UpdatePermission 更新权限
func (r *permissionRepository) UpdatePermission(permission *models.Permission) error {
	return r.db.Save(permission).Error
}

// DeletePermission 删除权限
func (r *permissionRepository) DeletePermission(id uint) error {
	return r.db.Delete(&models.Permission{}, id).Error
}

// GetPermissionsByResource 根据资源获取权限
func (r *permissionRepository) GetPermissionsByResource(resource string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Where("resource = ?", resource).Find(&permissions).Error
	return permissions, err
}
