package database

import (
	"fmt"
	"log"
	"time"

	"blockChainBrowser/server/config"
	"blockChainBrowser/server/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init() error {
	var err error

	switch config.AppConfig.Database.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.AppConfig.Database.Username,
			config.AppConfig.Database.Password,
			config.AppConfig.Database.Host,
			config.AppConfig.Database.Port,
			config.AppConfig.Database.DBName,
		)

		// 使用自定义 GORM 日志器，输出到与 logrus 相同的 writer
		stdLogger := log.New(logrus.StandardLogger().Writer(), "", 0)
		gormLogger := logger.New(
			stdLogger,
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      true,
				Colorful:                  false,
			},
		)

		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})

		if err != nil {
			return fmt.Errorf("failed to connect to MySQL database: %w", err)
		}

		// 配置连接池
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}

		sqlDB.SetMaxOpenConns(config.AppConfig.Database.MaxOpenConns)
		sqlDB.SetMaxIdleConns(config.AppConfig.Database.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(config.AppConfig.Database.ConnMaxLifetime)

	default:
		return fmt.Errorf("unsupported database driver: %s", config.AppConfig.Database.Driver)
	}

	// 自动迁移数据库表
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Printf("Database initialized successfully with driver: %s", config.AppConfig.Database.Driver)
	return nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	return DB.AutoMigrate(
		&models.Block{},
		&models.Transaction{},
		&models.Address{},
		&models.User{},
		&models.APIKey{},
		&models.AccessToken{},
		&models.RequestLog{},
		&models.UserAddress{},
		&models.BaseConfig{},
		&models.CoinConfig{},
		&models.TransactionReceipt{},
		&models.ParserConfig{},
		&models.Contract{},
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.UserRole{},
		&models.UserBalance{},
		&models.EarningsRecord{},
	)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
