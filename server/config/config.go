package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	Log         LogConfig         `yaml:"log"`
	WebSocket   WebSocketConfig   `yaml:"websocket"`
	CORS        CORSConfig        `yaml:"cors"`
	API         APIConfig         `yaml:"api"`
	Blockchain  BlockchainConfig  `yaml:"blockchain"`
	Cache       CacheConfig       `yaml:"cache"`
	Security    SecurityConfig    `yaml:"security"`
	DataCleanup DataCleanupConfig `yaml:"data_cleanup"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	TLSEnabled     bool          `yaml:"tls_enabled"`
	TLSPort        int           `yaml:"tls_port"`
	CertFile       string        `yaml:"cert_file"`
	KeyFile        string        `yaml:"key_file"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	MaxConnections int           `yaml:"max_connections"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `yaml:"driver"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"dbname"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
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

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	Enabled         bool          `yaml:"enabled"`
	Path            string        `yaml:"path"`
	PingInterval    time.Duration `yaml:"ping_interval"`
	PongWait        time.Duration `yaml:"pong_wait"`
	WriteWait       time.Duration `yaml:"write_wait"`
	ReadBufferSize  int           `yaml:"read_buffer_size"`
	WriteBufferSize int           `yaml:"write_buffer_size"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

// APIConfig API配置
type APIConfig struct {
	Version   string          `yaml:"version"`
	Prefix    string          `yaml:"prefix"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
	Burst             int  `yaml:"burst"`
}

// BlockchainConfig 区块链配置
type BlockchainConfig struct {
	Chains                  map[string]ChainConfig `yaml:"chains"`
	VerificationTimeout     time.Duration          `yaml:"verification_timeout"`      // 全局区块验证超时时间
	DefaultVerificationTime time.Duration          `yaml:"default_verification_time"` // 默认验证时间（如果未配置）
	// 动态验证超时配置
	DynamicVerification     bool          `yaml:"dynamic_verification"`      // 是否启用动态验证超时
	BalanceTransactionCount int           `yaml:"balance_transaction_count"` // 平衡交易数量（默认200条）
	BalanceBlockSize        int64         `yaml:"balance_block_size"`        // 平衡区块大小（默认128KB）
	MinVerificationTime     time.Duration `yaml:"min_verification_time"`     // 最小验证时间
	MaxVerificationTime     time.Duration `yaml:"max_verification_time"`     // 最大验证时间
}

// ChainConfig 链配置
type ChainConfig struct {
	Name     string   `yaml:"name"`
	ChainID  int      `yaml:"chain_id"`
	Symbol   string   `yaml:"symbol"`
	Decimals int      `yaml:"decimals"`
	Enabled  bool     `yaml:"enabled"`
	RPCURL   string   `yaml:"rpc_url"`   // RPC节点URL（兼容单个）
	RPCURLs  []string `yaml:"rpc_urls"`  // 多个RPC节点URL（用于JSON-RPC）
	RESTURLs []string `yaml:"rest_urls"` // 多个REST API URL（用于地址查询）
	Username string   `yaml:"username"`  // RPC用户名（如果需要）
	Password string   `yaml:"password"`  // RPC密码（如果需要）
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled bool        `yaml:"enabled"`
	Driver  string      `yaml:"driver"`
	Redis   RedisConfig `yaml:"redis"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret     string        `yaml:"jwt_secret"`
	JWTExpiration time.Duration `yaml:"jwt_expiration"`
	BcryptCost    int           `yaml:"bcrypt_cost"`
}

// DataCleanupConfig 数据清理配置
type DataCleanupConfig struct {
	ETH *ChainCleanupConfig `yaml:"eth"`
	BTC *ChainCleanupConfig `yaml:"btc"`
}

// ChainCleanupConfig 单链清理配置
type ChainCleanupConfig struct {
	MaxBlocks        int64 `yaml:"max_blocks"`        // 最大保留区块数
	CleanupThreshold int64 `yaml:"cleanup_threshold"` // 清理阈值
	BatchSize        int   `yaml:"batch_size"`        // 批量删除大小
	Interval         int   `yaml:"interval"`          // 清理间隔（分钟）
}

var AppConfig *Config

// Load 加载配置
func Load() error {
	// 首先尝试加载YAML配置文件
	if err := loadYAMLConfig(); err != nil {
		logrus.Warn("Failed to load YAML config, falling back to environment variables")
		// 如果YAML加载失败，回退到环境变量
		if err := loadEnvConfig(); err != nil {
			return fmt.Errorf("failed to load any configuration: %w", err)
		}
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(AppConfig.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// 设置日志格式
	switch AppConfig.Log.Format {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	}

	// 配置日志输出（控制台与文件可同时启用）
	var writers []io.Writer
	if AppConfig.Log.Output == "stdout" || AppConfig.Log.Output == "" {
		writers = append(writers, os.Stdout)
	} else if AppConfig.Log.Output == "stderr" {
		writers = append(writers, os.Stderr)
	}
	if AppConfig.Log.File.Enabled {
		logDir := AppConfig.Log.File.Path
		if logDir == "" {
			logDir = "."
		}
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
		logFile := filepath.Join(logDir, "server.log")
		writers = append(writers, &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    AppConfig.Log.File.MaxSize,
			MaxBackups: AppConfig.Log.File.MaxBackups,
			MaxAge:     AppConfig.Log.File.MaxAge,
			Compress:   true,
		})
	}
	if len(writers) > 0 {
		logrus.SetOutput(io.MultiWriter(writers...))
	} else {
		logrus.SetOutput(os.Stdout)
	}

	// 同步 Gin 的输出到相同的 writer
	gin.DisableConsoleColor()
	if len(writers) > 0 {
		mw := io.MultiWriter(writers...)
		gin.DefaultWriter = mw
		gin.DefaultErrorWriter = mw
	}

	// 将标准库 log 的输出重定向到 logrus，使其也写入相同的目标
	log.SetOutput(logrus.StandardLogger().Writer())
	log.SetFlags(0)

	return nil
}

// loadYAMLConfig 加载YAML配置文件
func loadYAMLConfig() error {
	configPath := getConfigPath()

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
			Host:           getEnv("SERVER_HOST", "localhost"),
			Port:           getEnvAsInt("SERVER_PORT", 8080),
			TLSEnabled:     getEnvAsBool("TLS_ENABLED", false),
			TLSPort:        getEnvAsInt("TLS_PORT", 8443),
			CertFile:       getEnv("TLS_CERT_FILE", ""),
			KeyFile:        getEnv("TLS_KEY_FILE", ""),
			ReadTimeout:    getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:   getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			MaxConnections: getEnvAsInt("SERVER_MAX_CONNECTIONS", 1000),
		},
		Database: DatabaseConfig{
			Driver:          getEnv("DB_DRIVER", "sqlite"),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 3306),
			Username:        getEnv("DB_USERNAME", ""),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "blockchain.db"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 3600*time.Second),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		WebSocket: WebSocketConfig{
			Enabled:         getEnvAsBool("WS_ENABLED", true),
			Path:            getEnv("WS_PATH", "/ws"),
			PingInterval:    getEnvAsDuration("WS_PING_INTERVAL", 30*time.Second),
			PongWait:        getEnvAsDuration("WS_PONG_WAIT", 60*time.Second),
			WriteWait:       getEnvAsDuration("WS_WRITE_WAIT", 10*time.Second),
			ReadBufferSize:  getEnvAsInt("WS_READ_BUFFER_SIZE", 1024),
			WriteBufferSize: getEnvAsInt("WS_WRITE_BUFFER_SIZE", 1024),
		},
		CORS: CORSConfig{
			AllowedOrigins:   []string{getEnv("CORS_ALLOW_ORIGIN", "*")},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getEnvAsInt("CORS_MAX_AGE", 86400),
		},
		API: APIConfig{
			Version: getEnv("API_VERSION", "v1"),
			Prefix:  getEnv("API_PREFIX", "/api"),
			RateLimit: RateLimitConfig{
				Enabled:           getEnvAsBool("API_RATE_LIMIT_ENABLED", true),
				RequestsPerMinute: getEnvAsInt("API_RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
				Burst:             getEnvAsInt("API_RATE_LIMIT_BURST", 20),
			},
		},
		Blockchain: BlockchainConfig{
			VerificationTimeout:     getEnvAsDuration("BLOCKCHAIN_VERIFICATION_TIMEOUT", 30*time.Second),
			DefaultVerificationTime: getEnvAsDuration("BLOCKCHAIN_DEFAULT_VERIFICATION_TIME", 10*time.Second),
			// 动态验证超时默认配置
			DynamicVerification:     getEnvAsBool("BLOCKCHAIN_DYNAMIC_VERIFICATION", true),
			BalanceTransactionCount: getEnvAsInt("BLOCKCHAIN_BALANCE_TRANSACTION_COUNT", 200),
			BalanceBlockSize:        getEnvAsInt64("BLOCKCHAIN_BALANCE_BLOCK_SIZE", 128*1024), // 128KB
			MinVerificationTime:     getEnvAsDuration("BLOCKCHAIN_MIN_VERIFICATION_TIME", 5*time.Second),
			MaxVerificationTime:     getEnvAsDuration("BLOCKCHAIN_MAX_VERIFICATION_TIME", 30*time.Second),
		},
		Security: SecurityConfig{
			JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
			JWTExpiration: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
			BcryptCost:    getEnvAsInt("BCRYPT_COST", 12),
		},
	}

	logrus.Info("Configuration loaded from environment variables")
	return nil
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
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

// getEnvAsInt64 获取环境变量并转换为int64
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
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
