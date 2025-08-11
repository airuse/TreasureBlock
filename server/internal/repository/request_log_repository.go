package repository

import (
	"time"

	"blockChainBrowser/server/internal/models"

	"gorm.io/gorm"
)

type RequestLogRepository interface {
	Create(log *models.RequestLog) error
	GetUsageStats(userID uint, apiKeyID uint) (*UsageStats, error)
	GetHourlyRequestCount(userID uint, apiKeyID uint, startTime time.Time) (int64, error)
	CleanupOldLogs(days int) error
}

type UsageStats struct {
	TotalRequests   int64
	TodayRequests   int64
	ThisHourRequests int64
	AvgResponseTime float64
}

type requestLogRepository struct {
	db *gorm.DB
}

func NewRequestLogRepository(db *gorm.DB) RequestLogRepository {
	return &requestLogRepository{db: db}
}

func (r *requestLogRepository) Create(log *models.RequestLog) error {
	return r.db.Create(log).Error
}

func (r *requestLogRepository) GetUsageStats(userID uint, apiKeyID uint) (*UsageStats, error) {
	var stats UsageStats
	
	// 总请求数
	err := r.db.Model(&models.RequestLog{}).
		Where("user_id = ? AND api_key_id = ?", userID, apiKeyID).
		Count(&stats.TotalRequests).Error
	if err != nil {
		return nil, err
	}
	
	// 今日请求数
	today := time.Now().Truncate(24 * time.Hour)
	err = r.db.Model(&models.RequestLog{}).
		Where("user_id = ? AND api_key_id = ? AND created_at >= ?", userID, apiKeyID, today).
		Count(&stats.TodayRequests).Error
	if err != nil {
		return nil, err
	}
	
	// 当前小时请求数
	thisHour := time.Now().Truncate(time.Hour)
	err = r.db.Model(&models.RequestLog{}).
		Where("user_id = ? AND api_key_id = ? AND created_at >= ?", userID, apiKeyID, thisHour).
		Count(&stats.ThisHourRequests).Error
	if err != nil {
		return nil, err
	}
	
	// 平均响应时间
	var avgDuration float64
	err = r.db.Model(&models.RequestLog{}).
		Where("user_id = ? AND api_key_id = ?", userID, apiKeyID).
		Select("COALESCE(AVG(duration), 0)").
		Scan(&avgDuration).Error
	if err != nil {
		return nil, err
	}
	stats.AvgResponseTime = avgDuration
	
	return &stats, nil
}

func (r *requestLogRepository) GetHourlyRequestCount(userID uint, apiKeyID uint, startTime time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.RequestLog{}).
		Where("user_id = ? AND api_key_id = ? AND created_at >= ?", userID, apiKeyID, startTime).
		Count(&count).Error
	return count, err
}

func (r *requestLogRepository) CleanupOldLogs(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	return r.db.Where("created_at < ?", cutoffTime).Delete(&models.RequestLog{}).Error
}
