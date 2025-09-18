package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"blockChainBrowser/client/scanner/pkg"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Config 扫块工具配置
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Blockchain BlockchainConfig `yaml:"blockchain"`
	Log        LogConfig        `yaml:"log"`
	Database   DatabaseConfig   `yaml:"database"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host        string    `yaml:"host"`
	Port        int       `yaml:"port"`
	Protocol    string    `yaml:"protocol"`
	Environment string    `yaml:"environment"` // 环境标识：development, production
	APIKey      string    `yaml:"api_key"`
	SecretKey   string    `yaml:"secret_key"`
	TLS         TLSConfig `yaml:"tls"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	SkipVerify bool   `yaml:"skip_verify"`
	CACert     string `yaml:"ca_cert"`
}

// BlockchainConfig 区块链配置
type BlockchainConfig struct {
	Chains map[string]ChainConfig `yaml:"chains"`
}

// ChainConfig 链配置
type ChainConfig struct {
	ChainID         int      `yaml:"chain_id"`
	Name            string   `yaml:"name"`
	Symbol          string   `yaml:"symbol"`
	Decimals        int      `yaml:"decimals"`
	Enabled         bool     `yaml:"enabled"`
	RPCURL          string   `yaml:"rpc_url"`
	ExplorerAPIURLs []string `yaml:"explorer_api_urls"` // 支持多个外部API节点
	APIKey          string   `yaml:"api_key"`
	Username        string   `yaml:"username"`
	Password        string   `yaml:"password"`
	TokenAddresses  []string `yaml:"token_addresses"` // 本地配置的币种合约地址列表（备用）
	// 每个链的独立扫描配置
	Scan ChainScanConfig `yaml:"scan"`
}

// ChainScanConfig 链扫描配置
type ChainScanConfig struct {
	Enabled       bool          `yaml:"enabled"`
	Interval      time.Duration `yaml:"interval"`
	Confirmations int           `yaml:"confirmations"`
	AutoStart     bool          `yaml:"auto_start"`
	SaveToFile    bool          `yaml:"save_to_file"`
	OutputDir     string        `yaml:"output_dir"`
	// 链特定的扫描配置
	Priority        int           `yaml:"priority"`         // 扫描优先级，数字越小优先级越高
	MaxConcurrent   int           `yaml:"max_concurrent"`   // 最大并发扫描数
	BlockTimeout    time.Duration `yaml:"block_timeout"`    // 单个区块扫描超时时间
	RescanInterval  time.Duration `yaml:"rescan_interval"`  // 重新扫描间隔（用于处理分叉）
	EnableMempool   bool          `yaml:"enable_mempool"`   // 是否启用内存池监控
	MempoolInterval time.Duration `yaml:"mempool_interval"` // 内存池检查间隔
	// 批量上传配置
	BatchUpload  bool          `yaml:"batch_upload"`  // 是否启用批量上传（推荐启用）
	BatchSize    int           `yaml:"batch_size"`    // 批量上传大小，默认1000
	BatchTimeout time.Duration `yaml:"batch_timeout"` // 批量上传超时时间

	// 预取与循环批处理参数
	PrefetchWindow  int `yaml:"prefetch_window"`   // 预取窗口大小（ensureTxPrefetch窗口）
	HeightsPerCycle int `yaml:"heights_per_cycle"` // 每个扫描周期处理的区块数量
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string     `yaml:"level"`
	Format string     `yaml:"format"`
	Output string     `yaml:"output"`
	File   FileConfig `yaml:"file"`
}

// FileConfig 文件配置
type FileConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	Enabled  bool   `yaml:"enabled"`
}

var AppConfig *Config
var ScannerAPIInstance *pkg.ScannerAPI // 全局API实例

// Load 加载配置
func Load(configPath string) error {
	// 首先尝试加载YAML配置文件
	if err := loadYAMLConfig(configPath); err != nil {
		logrus.Warn("Failed to load YAML config, falling back to environment variables")
		// 如果YAML加载失败，回退到环境变量
		if err := loadEnvConfig(); err != nil {
			return fmt.Errorf("failed to load any configuration: %w", err)
		}
	}

	// 为每个链设置默认扫描配置
	setDefaultChainScanConfigs()

	// 关于配置信息中的扫块配置，由于是多客户端统一扫块，因此需要将客户端的扫块配置取中央服务器配置信息！
	// 首先调用api接口 获取配置信息
	if err := loadServerConfig(); err != nil {
		logrus.Warnf("Failed to load server config: %v, using local config", err)
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(AppConfig.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	return nil
}

// loadYAMLConfig 加载YAML配置文件
func loadYAMLConfig(configPath string) error {
	configPath = getConfigPath(configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	AppConfig = &Config{}
	if err := yaml.Unmarshal(data, AppConfig); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	logrus.Info("Configuration loaded from YAML file")
	return nil
}

// loadEnvConfig 加载环境变量配置（回退方案）
func loadEnvConfig() error {
	// 加载环境变量文件
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using system environment variables")
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Host:     getEnv("SERVER_HOST", "localhost"),
			Port:     getEnvAsInt("SERVER_PORT", 8080),
			Protocol: getEnv("SERVER_PROTOCOL", "http"),
			APIKey:   getEnv("SERVER_API_KEY", ""),
			TLS: TLSConfig{
				Enabled:    getEnvAsBool("SERVER_TLS_ENABLED", false),
				SkipVerify: getEnvAsBool("SERVER_TLS_SKIP_VERIFY", true),
				CACert:     getEnv("SERVER_TLS_CA_CERT", ""),
			},
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "sqlite"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 3306),
			Username: getEnv("DB_USERNAME", ""),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "scanner.db"),
			Enabled:  getEnvAsBool("DB_ENABLED", false),
		},
	}

	logrus.Info("Configuration loaded from environment variables")
	return nil
}

// getConfigPath 获取配置文件路径
func getConfigPath(configPath string) string {
	if configPath != "" {
		return configPath
	}
	// 按优先级查找配置文件
	paths := []string{
		"config/config.yaml",
		"config/config.yml",
		"./config/config.yaml",
		"./config/config.yml",
		"../config/config.yaml",
		"../config/config.yml",
		"config.yaml",
		"config.yml",
		"./config.yaml",
		"./config.yml",
		"../config.yaml",
		"../config.yml",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "config/config.yaml" // 默认路径
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsUint64 获取环境变量并转换为uint64
func getEnvAsUint64(key string, defaultValue uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		if uintValue, err := strconv.ParseUint(value, 10, 64); err == nil {
			return uintValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsDuration 获取环境变量并转换为时间间隔
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// loadServerConfig 从服务器加载配置
func loadServerConfig() error {
	// 构建服务器URL
	serverURL := fmt.Sprintf("%s://%s:%d",
		AppConfig.Server.Protocol,
		AppConfig.Server.Host,
		AppConfig.Server.Port)

	// 创建API实例
	api := pkg.NewScannerAPI(serverURL, AppConfig.Server.APIKey, AppConfig.Server.SecretKey, AppConfig.Server.Environment, logrus.StandardLogger())
	ScannerAPIInstance = api // 保存全局API实例

	logrus.Infof("Initializing scanner API with server: %s", serverURL)

	// 健康检查
	if err := api.HealthCheck(); err != nil {
		return fmt.Errorf("server health check failed: %w", err)
	}
	logrus.Info("Server health check passed")

	// 为每个启用的链加载配置
	for chainKey, chainConfig := range AppConfig.Blockchain.Chains {
		if !chainConfig.Enabled {
			continue
		}

		logrus.Infof("Loading server config for chain: %s", chainKey)

		// 获取扫块配置
		scanConfig, err := api.GetScanConfig(chainKey)
		if err != nil {
			logrus.Warnf("Failed to load scan config for chain %s: %v", chainKey, err)
			continue
		}

		// 更新本地配置
		chainConfig.Scan.Confirmations = scanConfig.Confirmations

		// 获取RPC配置（可选）
		// rpcConfig, err := api.GetRPCConfig(chainKey)
		// if err != nil {
		// 	logrus.Warnf("Failed to load RPC config for chain %s: %v", chainKey, err)
		// } else {
		// 	// 更新RPC配置
		// 	chainConfig.RPCURL = rpcConfig.URL
		// 	chainConfig.APIKey = rpcConfig.APIKey
		// 	chainConfig.Username = rpcConfig.Username
		// 	chainConfig.Password = rpcConfig.Password
		// 	logrus.Infof("Updated RPC config for chain %s", chainKey)
		// }

		// 获取币种配置（如果是以太坊链或Solana链）
		if chainKey == "eth" || chainKey == "sol" || chainKey == "bsc" {
			// if err := loadCoinConfigsFromAPI(api, &chainConfig); err != nil {
			// 	logrus.Warnf("Failed to load coin configs for chain %s: %v", chainKey, err)
			// } else {
			// 	logrus.Infof("Successfully loaded coin configs for chain %s", chainKey)
			// }
		}

		// 更新链配置到全局配置中
		AppConfig.Blockchain.Chains[chainKey] = chainConfig

		logrus.Infof("Successfully loaded server config for chain: %s", chainKey)
	}

	return nil
}

// setDefaultChainScanConfigs 为每个链设置默认扫描配置
func setDefaultChainScanConfigs() {
	for chainKey, chainConfig := range AppConfig.Blockchain.Chains {
		// 如果链没有扫描配置，设置默认值
		if chainConfig.Scan.Confirmations == 0 {
			chainConfig.Scan.Confirmations = 6 // 默认确认数
		}
		if chainConfig.Scan.OutputDir == "" {
			chainConfig.Scan.OutputDir = filepath.Join("./output", chainKey)
		}
		if chainConfig.Scan.MaxConcurrent == 0 {
			chainConfig.Scan.MaxConcurrent = 3 // 默认并发数
		}
		if chainConfig.Scan.BlockTimeout == 0 {
			chainConfig.Scan.BlockTimeout = 30 * time.Second // 默认超时时间
		}
		if chainConfig.Scan.RescanInterval == 0 {
			chainConfig.Scan.RescanInterval = 60 * time.Second // 默认重新扫描间隔
		}
		if chainConfig.Scan.MempoolInterval == 0 {
			chainConfig.Scan.MempoolInterval = 10 * time.Second // 默认内存池检查间隔
		}
		// 批量上传配置默认值
		if chainConfig.Scan.BatchSize == 0 {
			chainConfig.Scan.BatchSize = 1000 // 默认批量大小
		}
		if chainConfig.Scan.BatchTimeout == 0 {
			chainConfig.Scan.BatchTimeout = 30 * time.Second // 默认批量上传超时时间
		}

		// 更新配置
		AppConfig.Blockchain.Chains[chainKey] = chainConfig
	}
}

// loadCoinConfigsFromAPI 从后端API加载币种配置
func loadCoinConfigsFromAPI(api *pkg.ScannerAPI, chainConfig *ChainConfig) error {
	// 获取所有币种配置
	coinConfigs, err := api.GetAllCoinConfigs()
	if err != nil {
		return fmt.Errorf("failed to get coin configs from API: %w", err)
	}

	// 提取以太坊链和Solana链的合约地址
	var contractAddresses []string
	for _, config := range coinConfigs {
		if (config.ChainName == "eth" || config.ChainName == "sol") && config.Status == 1 && config.ContractAddr != "" {
			contractAddresses = append(contractAddresses, config.ContractAddr)
		}
	}

	// 获取所有合约配置
	constants, err := api.GetAllContracts()
	if err != nil {
		return fmt.Errorf("failed to get all contracts from API: %w", err)
	}
	for _, constant := range constants {
		if (constant.ChainName == "eth" || constant.ChainName == "sol") && constant.Status == 1 && constant.Address != "" {
			contractAddresses = append(contractAddresses, constant.Address)
		}
	}

	// 将API获取的地址添加到本地配置中（去重）
	existingAddresses := make(map[string]bool)
	for _, addr := range chainConfig.TokenAddresses {
		existingAddresses[strings.ToLower(addr)] = true
	}

	for _, addr := range contractAddresses {
		if !existingAddresses[strings.ToLower(addr)] {
			chainConfig.TokenAddresses = append(chainConfig.TokenAddresses, addr)
			existingAddresses[strings.ToLower(addr)] = true
		}
	}

	logrus.Infof("Loaded %d coin configs from API, total token addresses: %d",
		len(coinConfigs), len(chainConfig.TokenAddresses))

	return nil
}

// GetScannerAPI 获取扫块器API实例
func GetScannerAPI() *pkg.ScannerAPI {
	return ScannerAPIInstance
}
