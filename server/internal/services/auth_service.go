package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req *dto.RegisterRequest) (*models.User, error)
	Login(req *dto.LoginRequest) (*dto.LoginResponse, error)
	CreateAPIKey(userID uint, req *dto.CreateAPIKeyRequest) (*dto.CreateAPIKeyResponse, error)
	GetAPIKeys(userID uint) ([]*dto.APIKeyResponse, error)
	UpdateAPIKey(userID uint, keyID uint, req *dto.UpdateAPIKeyRequest) (*dto.APIKeyResponse, error)
	DeleteAPIKey(userID uint, keyID uint) error
	GetAccessToken(req *dto.GetAccessTokenRequest) (*dto.GetAccessTokenResponse, error)
	ValidateAccessToken(tokenString string) (*jwt.Token, error)
	RefreshToken(userID uint) (*dto.LoginResponse, error)
	GetUserProfile(userID uint) (*dto.UserProfileResponse, error)
	ChangePassword(userID uint, req *dto.ChangePasswordRequest) error
	GetUsageStats(userID uint, apiKeyID uint) (*dto.APIUsageStatsResponse, error)
}

type authService struct {
	userRepo          repository.UserRepository
	apiKeyRepo        repository.APIKeyRepository
	requestLogRepo    repository.RequestLogRepository
	permissionService *PermissionService // 添加权限服务
	jwtSecret         string
	jwtExpiration     time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	apiKeyRepo repository.APIKeyRepository,
	requestLogRepo repository.RequestLogRepository,
	permissionService *PermissionService, // 添加权限服务参数
	jwtSecret string,
	jwtExpiration time.Duration,
) AuthService {
	service := &authService{
		userRepo:          userRepo,
		apiKeyRepo:        apiKeyRepo,
		requestLogRepo:    requestLogRepo,
		permissionService: permissionService, // 设置权限服务
		jwtSecret:         jwtSecret,
		jwtExpiration:     jwtExpiration,
	}

	// 初始化默认角色和权限
	go func() {
		if err := permissionService.InitializeDefaultRoles(); err != nil {
			log.Printf("初始化默认角色和权限失败: %v", err)
		}
	}()

	return service
}

// Register 用户注册
func (s *authService) Register(req *dto.RegisterRequest) (*models.User, error) {
	// 验证用户名
	if req.Username == "" || len(req.Username) < 3 {
		return nil, errors.New("用户名至少需要3个字符")
	}

	// 验证邮箱格式
	if req.Email == "" {
		return nil, errors.New("邮箱不能为空")
	}

	// 验证密码强度
	if len(req.Password) < 6 {
		return nil, errors.New("密码至少需要6个字符")
	}

	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	exists, err = s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if exists {
		return nil, errors.New("邮箱已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 检查是否是第一个用户
	isFirstUser, err := s.userRepo.IsFirstUser()
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user", // 默认为普通用户
		Status:   1,      // 默认启用
	}

	// 如果是第一个用户，设置为管理员
	if isFirstUser {
		user.Role = "administrator"
		log.Printf("第一个用户 %s 自动设置为管理员", req.Username)
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 如果是第一个用户，确保角色权限系统已初始化
	if isFirstUser {
		if err := s.permissionService.SetFirstUserAsAdmin(); err != nil {
			log.Printf("设置第一个用户为管理员失败: %v", err)
		}
	}

	return user, nil
}

// Login 用户登录
func (s *authService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("用户账户已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 生成JWT令牌
	token, expiresAt, err := s.generateJWT(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	// 更新最后登录时间（如果需要的话，可以在用户模型中添加LastLogin字段）
	// 这里暂时不更新，因为当前用户模型没有LastLogin字段

	return &dto.LoginResponse{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

func (s *authService) CreateAPIKey(userID uint, req *dto.CreateAPIKeyRequest) (*dto.CreateAPIKeyResponse, error) {
	// 生成API Key和Secret Key
	apiKey, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("生成API密钥失败: %w", err)
	}

	secretKey, err := s.generateSecretKey()
	if err != nil {
		return nil, fmt.Errorf("生成Secret密钥失败: %w", err)
	}

	// 加密Secret Key存储
	hashedSecretKey, err := bcrypt.GenerateFromPassword([]byte(secretKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("加密Secret密钥失败: %w", err)
	}

	// 设置默认限流值
	rateLimit := req.RateLimit
	if rateLimit == 0 {
		rateLimit = 1000 // 默认每小时1000次请求
	}

	// 创建API密钥记录
	key := &models.APIKey{
		UserID:      userID,
		Name:        req.Name,
		APIKey:      apiKey,
		SecretKey:   string(hashedSecretKey),
		Permissions: s.convertPermissionsToString(req.Permissions),
		RateLimit:   rateLimit,
		IsActive:    true,
		ExpiresAt:   s.parseExpiresAt(req.ExpiresAt),
	}

	err = s.apiKeyRepo.Create(key)
	if err != nil {
		return nil, fmt.Errorf("创建API密钥失败: %w", err)
	}

	return &dto.CreateAPIKeyResponse{
		ID:        key.ID,
		Name:      key.Name,
		APIKey:    key.APIKey,
		SecretKey: secretKey, // 只在创建时返回
		ExpiresAt: key.ExpiresAt,
		CreatedAt: key.CreatedAt,
	}, nil
}

func (s *authService) GetAPIKeys(userID uint) ([]*dto.APIKeyResponse, error) {
	keys, err := s.apiKeyRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取API密钥列表失败: %w", err)
	}

	var responses []*dto.APIKeyResponse
	for _, key := range keys {
		responses = append(responses, &dto.APIKeyResponse{
			ID:          key.ID,
			Name:        key.Name,
			APIKey:      key.APIKey,
			SecretKey:   key.SecretKey, // 添加SecretKey字段
			Permissions: s.parsePermissionsFromString(key.Permissions),
			RateLimit:   key.RateLimit, // 添加RateLimit字段
			IsActive:    key.IsActive,
			ExpiresAt:   key.ExpiresAt,
			LastUsedAt:  key.LastUsedAt,
			UsageCount:  key.UsageCount,
			CreatedAt:   key.CreatedAt,
			UpdatedAt:   key.UpdatedAt,
		})
	}

	return responses, nil
}

func (s *authService) UpdateAPIKey(userID uint, keyID uint, req *dto.UpdateAPIKeyRequest) (*dto.APIKeyResponse, error) {
	// 获取API密钥
	key, err := s.apiKeyRepo.GetByID(keyID)
	if err != nil {
		return nil, err
	}

	// 检查所有权
	if key.UserID != userID {
		return nil, errors.New("无权限修改此API密钥")
	}

	// 更新字段
	if req.Name != nil {
		key.Name = *req.Name
	}
	if req.Permissions != nil {
		key.Permissions = s.convertPermissionsToString(*req.Permissions)
	}
	if req.RateLimit != nil {
		key.RateLimit = *req.RateLimit
	}
	if req.ExpiresAt != nil {
		key.ExpiresAt = s.parseExpiresAt(*req.ExpiresAt)
	}
	if req.IsActive != nil {
		key.IsActive = *req.IsActive
	}

	// 保存更改
	err = s.apiKeyRepo.Update(key)
	if err != nil {
		return nil, fmt.Errorf("更新API密钥失败: %w", err)
	}

	return &dto.APIKeyResponse{
		ID:          key.ID,
		Name:        key.Name,
		APIKey:      key.APIKey,
		SecretKey:   key.SecretKey, // 添加SecretKey字段
		Permissions: s.parsePermissionsFromString(key.Permissions),
		RateLimit:   key.RateLimit, // 添加RateLimit字段
		IsActive:    key.IsActive,
		ExpiresAt:   key.ExpiresAt,
		LastUsedAt:  key.LastUsedAt,
		UsageCount:  key.UsageCount,
		CreatedAt:   key.CreatedAt,
		UpdatedAt:   key.UpdatedAt,
	}, nil
}

// 辅助方法：将权限数组转换为JSON字符串
func (s *authService) convertPermissionsToString(permissions []string) string {
	if len(permissions) == 0 {
		return "[]"
	}
	data, err := json.Marshal(permissions)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// 辅助方法：将IP白名单数组转换为JSON字符串
func (s *authService) convertIPWhitelistToString(ipWhitelist []string) string {
	if len(ipWhitelist) == 0 {
		return "[]"
	}
	data, err := json.Marshal(ipWhitelist)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// 辅助方法：解析过期时间
func (s *authService) parseExpiresAt(expiresAt string) *time.Time {
	if expiresAt == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", expiresAt)
	if err != nil {
		return nil
	}
	return &t
}

// 辅助方法：将JSON字符串解析为权限数组
func (s *authService) parsePermissionsFromString(permissions string) []string {
	if permissions == "" || permissions == "[]" {
		return []string{}
	}
	var perms []string
	err := json.Unmarshal([]byte(permissions), &perms)
	if err != nil {
		return []string{}
	}
	return perms
}

// 辅助方法：将JSON字符串解析为IP白名单数组
func (s *authService) parseIPWhitelistFromString(ipWhitelist string) []string {
	if ipWhitelist == "" || ipWhitelist == "[]" {
		return []string{}
	}
	var ips []string
	err := json.Unmarshal([]byte(ipWhitelist), &ips)
	if err != nil {
		return []string{}
	}
	return ips
}

func (s *authService) DeleteAPIKey(userID uint, keyID uint) error {
	// 获取API密钥
	key, err := s.apiKeyRepo.GetByID(keyID)
	if err != nil {
		return err
	}

	// 检查所有权
	if key.UserID != userID {
		return errors.New("无权限删除此API密钥")
	}

	// 删除密钥
	return s.apiKeyRepo.Delete(keyID)
}

func (s *authService) GetAccessToken(req *dto.GetAccessTokenRequest) (*dto.GetAccessTokenResponse, error) {
	// 验证API密钥
	apiKey, err := s.apiKeyRepo.GetByAPIKey(req.APIKey)
	if err != nil {
		return nil, errors.New("API密钥或Secret密钥错误")
	}

	// 检查API密钥是否有效
	if !apiKey.IsActive {
		return nil, errors.New("API密钥无效或已过期")
	}

	// 验证Secret密钥
	if apiKey.SecretKey != req.SecretKey {
		return nil, errors.New("API密钥或Secret密钥错误")
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(apiKey.UserID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("用户账户已被禁用")
	}

	// 生成访问令牌
	token, expiresAt, err := s.generateJWT(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	// 更新API密钥使用统计
	now := time.Now()
	apiKey.LastUsedAt = &now
	apiKey.UsageCount++
	if err := s.apiKeyRepo.Update(apiKey); err != nil {
		// 不影响令牌生成，只记录错误
		log.Printf("更新API密钥使用统计失败: %v", err)
	}

	return &dto.GetAccessTokenResponse{
		AccessToken: token,
		ExpiresAt:   expiresAt.Unix(),
		TokenType:   "Bearer",
	}, nil
}

func (s *authService) ValidateAccessToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("令牌解析失败: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("令牌无效")
	}

	return token, nil
}

func (s *authService) GetUserProfile(userID uint) (*dto.UserProfileResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	return &dto.UserProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Status:   user.Status,
	}, nil
}

func (s *authService) ChangePassword(userID uint, req *dto.ChangePasswordRequest) error {
	// 获取用户
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// 验证当前密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword))
	if err != nil {
		return errors.New("当前密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("新密码加密失败: %w", err)
	}

	// 更新密码
	user.Password = string(hashedPassword)
	return s.userRepo.Update(user)
}

func (s *authService) GetUsageStats(userID uint, apiKeyID uint) (*dto.APIUsageStatsResponse, error) {
	stats, err := s.requestLogRepo.GetUsageStats(userID, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("获取使用统计失败: %w", err)
	}

	return &dto.APIUsageStatsResponse{
		TotalRequests:    stats.TotalRequests,
		TodayRequests:    stats.TodayRequests,
		ThisHourRequests: stats.ThisHourRequests,
		AvgResponseTime:  stats.AvgResponseTime,
	}, nil
}

// 生成JWT令牌
func (s *authService) generateJWT(userID uint, username string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.jwtExpiration)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
		"type":     "login_token",
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	return tokenString, expiresAt, err
}

// 生成API Key
func (s *authService) generateAPIKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "ak_" + hex.EncodeToString(bytes), nil
}

// 生成Secret Key
func (s *authService) generateSecretKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sk_" + hex.EncodeToString(bytes), nil
}

// 生成令牌哈希
func (s *authService) generateTokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// 从JWT令牌中提取用户ID
func ExtractUserIDFromToken(token *jwt.Token) (uint, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("无效的令牌声明")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("无效的用户ID")
	}

	return uint(userIDFloat), nil
}

// 从JWT令牌中提取API Key ID
func ExtractAPIKeyIDFromToken(token *jwt.Token) (uint, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("无效的令牌声明")
	}

	apiKeyIDFloat, ok := claims["api_key_id"].(float64)
	if !ok {
		return 0, errors.New("无效的API密钥ID")
	}

	return uint(apiKeyIDFloat), nil
}

// RefreshToken 刷新JWT令牌
func (s *authService) RefreshToken(userID uint) (*dto.LoginResponse, error) {
	// 获取用户信息
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 检查用户是否激活
	if user.Status != 1 {
		return nil, errors.New("用户账户已被禁用")
	}

	// 生成新的JWT令牌
	token, expiresAt, err := s.generateJWT(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	// 不更新LastLogin（模型无此字段）

	return &dto.LoginResponse{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}
