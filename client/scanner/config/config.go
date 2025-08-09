package config

import (
	"fmt"
	"os"
	"strconv"
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
	Scan       ScanConfig       `yaml:"scan"`
	Log        LogConfig        `yaml:"log"`
	Database   DatabaseConfig   `yaml:"database"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"`
	APIKey   string `yaml:"api_key"`
}

// BlockchainConfig 区块链配置
type BlockchainConfig struct {
	Chains map[string]ChainConfig `yaml:"chains"`
}

// ChainConfig 链配置
type ChainConfig struct {
	Name           string `yaml:"name"`
	Symbol         string `yaml:"symbol"`
	Decimals       int    `yaml:"decimals"`
	Enabled        bool   `yaml:"enabled"`
	RPCURL         string `yaml:"rpc_url"`
	ExplorerAPIURL string `yaml:"explorer_api_url"`
	APIKey         string `yaml:"api_key"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
}

// ScanConfig 扫描配置
type ScanConfig struct {
	Enabled          bool          `yaml:"enabled"`
	Interval         time.Duration `yaml:"interval"`
	BatchSize        int           `yaml:"batch_size"`
	MaxRetries       int           `yaml:"max_retries"`
	RetryDelay       time.Duration `yaml:"retry_delay"`
	Confirmations    int           `yaml:"confirmations"`
	StartBlockHeight uint64        `yaml:"start_block_height"`
	AutoStart        bool          `yaml:"auto_start"`
	SaveToFile       bool          `yaml:"save_to_file"`
	OutputDir        string        `yaml:"output_dir"`
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
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		Scan: ScanConfig{
			Enabled:          getEnvAsBool("SCAN_ENABLED", true),
			Interval:         getEnvAsDuration("SCAN_INTERVAL", 10*time.Second),
			BatchSize:        getEnvAsInt("SCAN_BATCH_SIZE", 100),
			MaxRetries:       getEnvAsInt("SCAN_MAX_RETRIES", 3),
			RetryDelay:       getEnvAsDuration("SCAN_RETRY_DELAY", 5*time.Second),
			Confirmations:    getEnvAsInt("SCAN_CONFIRMATIONS", 6),
			StartBlockHeight: getEnvAsUint64("SCAN_START_BLOCK_HEIGHT", 0),
			AutoStart:        getEnvAsBool("SCAN_AUTO_START", true),
			SaveToFile:       getEnvAsBool("SCAN_SAVE_TO_FILE", false),
			OutputDir:        getEnv("SCAN_OUTPUT_DIR", "./output"),
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
	api := pkg.NewScannerAPI(serverURL, AppConfig.Server.APIKey, logrus.StandardLogger())

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
		AppConfig.Scan.Interval = scanConfig.ScanInterval
		AppConfig.Scan.BatchSize = scanConfig.BatchSize
		AppConfig.Scan.Confirmations = scanConfig.Confirmations
		AppConfig.Scan.StartBlockHeight = scanConfig.StartBlockHeight
		AppConfig.Scan.MaxRetries = scanConfig.MaxRetries
		AppConfig.Scan.RetryDelay = scanConfig.RetryDelay

		// 获取RPC配置（可选）
		rpcConfig, err := api.GetRPCConfig(chainKey)
		if err != nil {
			logrus.Warnf("Failed to load RPC config for chain %s: %v", chainKey, err)
		} else {
			// 更新RPC配置
			chain := AppConfig.Blockchain.Chains[chainKey]
			chain.RPCURL = rpcConfig.URL
			chain.APIKey = rpcConfig.APIKey
			chain.Username = rpcConfig.Username
			chain.Password = rpcConfig.Password
			AppConfig.Blockchain.Chains[chainKey] = chain
			logrus.Infof("Updated RPC config for chain %s", chainKey)
		}

		logrus.Infof("Successfully loaded server config for chain: %s", chainKey)
	}

	return nil
}

// GetScannerAPI 获取扫块器API实例
func GetScannerAPI() *pkg.ScannerAPI {
	serverURL := fmt.Sprintf("%s://%s:%d",
		AppConfig.Server.Protocol,
		AppConfig.Server.Host,
		AppConfig.Server.Port)

	return pkg.NewScannerAPI(serverURL, AppConfig.Server.APIKey, logrus.StandardLogger())
}
