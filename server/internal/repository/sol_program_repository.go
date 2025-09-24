package repository

import (
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

// SolProgramRepository 程序维护仓库接口
type SolProgramRepository interface {
	Create(ctx context.Context, program *models.SolProgram) error
	Update(ctx context.Context, program *models.SolProgram) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.SolProgram, error)
	GetByProgramID(ctx context.Context, programID string) (*models.SolProgram, error)
	List(ctx context.Context, page, pageSize int, keyword string) ([]*models.SolProgram, int64, error)
	GetAll(ctx context.Context) ([]*models.SolProgram, error)
}

type solProgramRepository struct {
	db *gorm.DB
}

func NewSolProgramRepository(db *gorm.DB) SolProgramRepository {
	return &solProgramRepository{db: db}
}

func (r *solProgramRepository) Create(ctx context.Context, program *models.SolProgram) error {
	return r.db.WithContext(ctx).Create(program).Error
}

func (r *solProgramRepository) Update(ctx context.Context, program *models.SolProgram) error {
	return r.db.WithContext(ctx).Save(program).Error
}

func (r *solProgramRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.SolProgram{}, id).Error
}

func (r *solProgramRepository) GetByID(ctx context.Context, id uint) (*models.SolProgram, error) {
	var program models.SolProgram
	if err := r.db.WithContext(ctx).First(&program, id).Error; err != nil {
		return nil, err
	}
	return &program, nil
}

func (r *solProgramRepository) GetByProgramID(ctx context.Context, programID string) (*models.SolProgram, error) {
	var program models.SolProgram
	if err := r.db.WithContext(ctx).Where("program_id = ?", programID).First(&program).Error; err != nil {
		return nil, err
	}
	return &program, nil
}

func (r *solProgramRepository) List(ctx context.Context, page, pageSize int, keyword string) ([]*models.SolProgram, int64, error) {
	var programs []*models.SolProgram
	var total int64
	q := r.db.WithContext(ctx).Model(&models.SolProgram{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("program_id LIKE ? OR name LIKE ? OR alias LIKE ?", like, like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&programs).Error; err != nil {
		return nil, 0, err
	}
	return programs, total, nil
}

func (r *solProgramRepository) GetAll(ctx context.Context) ([]*models.SolProgram, error) {
	var programs []*models.SolProgram
	if err := r.db.WithContext(ctx).Order("id DESC").Find(&programs).Error; err != nil {
		return nil, err
	}
	return programs, nil
}
