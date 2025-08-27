package services

import (
	"errors"
	"log"

	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
)

// PermissionService 权限服务
type PermissionService struct {
	userRepo       repository.UserRepository
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
) *PermissionService {
	return &PermissionService{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// CheckPermission 检查用户是否有指定权限
func (s *PermissionService) CheckPermission(userID uint, resource, action string) (bool, error) {
	// 获取用户角色
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, err
	}

	// 如果是管理员，拥有所有权限
	if user.Role == "administrator" {
		return true, nil
	}

	// 获取用户的所有权限
	permissions, err := s.getUserPermissions(userID)
	if err != nil {
		return false, err
	}

	// 检查是否有指定权限
	for _, perm := range permissions {
		if perm.Resource == resource && perm.Action == action {
			return true, nil
		}
	}

	return false, nil
}

// CheckMultiplePermissions 检查用户是否有多个权限中的任意一个
func (s *PermissionService) CheckMultiplePermissions(userID uint, permissions []PermissionCheck) (bool, error) {
	for _, perm := range permissions {
		hasPermission, err := s.CheckPermission(userID, perm.Resource, perm.Action)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}
	return false, nil
}

// GetUserPermissions 获取用户的所有权限
func (s *PermissionService) GetUserPermissions(userID uint) ([]models.Permission, error) {
	return s.getUserPermissions(userID)
}

// getUserPermissions 内部方法：获取用户权限
func (s *PermissionService) getUserPermissions(userID uint) ([]models.Permission, error) {
	// 这里实现获取用户权限的逻辑
	// 可以通过用户角色来获取对应的权限
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 根据用户角色返回对应权限
	switch user.Role {
	case "administrator":
		// 管理员拥有所有权限
		return s.permissionRepo.GetAllPermissions()
	case "user":
		// 普通用户只有读取权限
		return s.permissionRepo.GetPermissionsByRole("user")
	default:
		return []models.Permission{}, nil
	}
}

// InitializeDefaultRoles 初始化默认角色和权限
func (s *PermissionService) InitializeDefaultRoles() error {
	// 检查是否已经初始化
	roles, err := s.roleRepo.GetAllRoles()
	if err != nil {
		return err
	}

	if len(roles) > 0 {
		log.Println("角色和权限已经初始化，跳过")
		return nil
	}

	log.Println("开始初始化默认角色和权限...")

	// 创建默认权限
	permissions := []models.Permission{
		// 合约管理权限
		{Name: "contract:read", DisplayName: "查看合约", Description: "查看合约信息", Resource: "contract", Action: "read"},
		{Name: "contract:create", DisplayName: "创建合约", Description: "创建新合约", Resource: "contract", Action: "create"},
		{Name: "contract:update", DisplayName: "更新合约", Description: "更新合约信息", Resource: "contract", Action: "update"},
		{Name: "contract:delete", DisplayName: "删除合约", Description: "删除合约", Resource: "contract", Action: "delete"},

		// 币种配置权限
		{Name: "coinconfig:read", DisplayName: "查看币种配置", Description: "查看币种配置信息", Resource: "coinconfig", Action: "read"},
		{Name: "coinconfig:create", DisplayName: "创建币种配置", Description: "创建新币种配置", Resource: "coinconfig", Action: "create"},
		{Name: "coinconfig:update", DisplayName: "更新币种配置", Description: "更新币种配置信息", Resource: "coinconfig", Action: "update"},
		{Name: "coinconfig:delete", DisplayName: "删除币种配置", Description: "删除币种配置", Resource: "coinconfig", Action: "delete"},

		// 区块管理权限
		{Name: "block:read", DisplayName: "查看区块", Description: "查看区块信息", Resource: "block", Action: "read"},
		{Name: "block:create", DisplayName: "创建区块", Description: "创建新区块", Resource: "block", Action: "create"},
		{Name: "block:update", DisplayName: "更新区块", Description: "更新区块信息", Resource: "block", Action: "update"},

		// 交易管理权限
		{Name: "transaction:read", DisplayName: "查看交易", Description: "查看交易信息", Resource: "transaction", Action: "read"},
		{Name: "transaction:create", DisplayName: "创建交易", Description: "创建新交易", Resource: "transaction", Action: "create"},
		{Name: "transaction:update", DisplayName: "更新交易", Description: "更新交易信息", Resource: "transaction", Action: "update"},
	}

	// 保存权限
	for i := range permissions {
		if err := s.permissionRepo.CreatePermission(&permissions[i]); err != nil {
			return err
		}
	}

	// 创建默认角色
	userRole := models.Role{
		Name:        "user",
		DisplayName: "普通用户",
		Description: "普通用户，只能查看信息",
	}

	adminRole := models.Role{
		Name:        "administrator",
		DisplayName: "系统管理员",
		Description: "系统管理员，拥有所有权限",
	}

	// 保存角色
	if err := s.roleRepo.CreateRole(&userRole); err != nil {
		return err
	}

	if err := s.roleRepo.CreateRole(&adminRole); err != nil {
		return err
	}

	// 为角色分配权限
	// 普通用户只有读取权限
	userPermissions := []models.Permission{}
	for _, perm := range permissions {
		if perm.Action == "read" {
			userPermissions = append(userPermissions, perm)
		}
	}
	userRole.Permissions = userPermissions

	// 管理员拥有所有权限
	adminRole.Permissions = permissions

	// 更新角色权限
	if err := s.roleRepo.UpdateRole(&userRole); err != nil {
		return err
	}

	if err := s.roleRepo.UpdateRole(&adminRole); err != nil {
		return err
	}

	log.Println("默认角色和权限初始化完成")
	return nil
}

// SetFirstUserAsAdmin 设置第一个用户为管理员
func (s *PermissionService) SetFirstUserAsAdmin() error {
	// 检查是否已经有管理员
	adminUser, err := s.userRepo.GetByRole("administrator")
	if err == nil && adminUser != nil {
		log.Println("已经存在管理员用户，跳过设置")
		return nil
	}

	// 获取第一个用户
	firstUser, err := s.userRepo.GetFirstUser()
	if err != nil {
		return err
	}

	if firstUser == nil {
		return errors.New("没有找到用户")
	}

	// 设置为管理员
	firstUser.Role = "administrator"
	if err := s.userRepo.UpdateUser(firstUser); err != nil {
		return err
	}

	log.Printf("用户 %s 已设置为管理员", firstUser.Username)
	return nil
}

// PermissionCheck 权限检查结构
type PermissionCheck struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// HasPermission 检查用户是否有指定权限的便捷方法
func (s *PermissionService) HasPermission(userID uint, resource, action string) bool {
	hasPermission, err := s.CheckPermission(userID, resource, action)
	if err != nil {
		log.Printf("权限检查失败: %v", err)
		return false
	}
	return hasPermission
}

// CanEditContract 检查用户是否可以编辑合约
func (s *PermissionService) CanEditContract(userID uint) bool {
	return s.HasPermission(userID, "contract", "update")
}

// CanEditCoinConfig 检查用户是否可以编辑币种配置
func (s *PermissionService) CanEditCoinConfig(userID uint) bool {
	return s.HasPermission(userID, "coinconfig", "update")
}

// CanCreateContract 检查用户是否可以创建合约
func (s *PermissionService) CanCreateContract(userID uint) bool {
	return s.HasPermission(userID, "contract", "create")
}

// CanDeleteContract 检查用户是否可以删除合约
func (s *PermissionService) CanDeleteContract(userID uint) bool {
	return s.HasPermission(userID, "contract", "delete")
}
