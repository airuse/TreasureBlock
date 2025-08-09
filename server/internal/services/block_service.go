package services

import (
	"context"
	"fmt"

	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
)

// BlockService 区块服务接口
type BlockService interface {
	GetBlockByHash(ctx context.Context, hash string) (*models.Block, error)
	GetBlockByHeight(ctx context.Context, height uint64) (*models.Block, error)
	GetLatestBlock(ctx context.Context, chain string) (*models.Block, error)
	ListBlocks(ctx context.Context, page, pageSize int, chain string) ([]*models.Block, int64, error)
	CreateBlock(ctx context.Context, block *models.Block) error
	UpdateBlock(ctx context.Context, block *models.Block) error
	DeleteBlock(ctx context.Context, hash string) error
}

// blockService 区块服务实现
type blockService struct {
	blockRepo repository.BlockRepository
}

// NewBlockService 创建区块服务实例
func NewBlockService(blockRepo repository.BlockRepository) BlockService {
	return &blockService{
		blockRepo: blockRepo,
	}
}

// GetBlockByHash 根据哈希获取区块
func (s *blockService) GetBlockByHash(ctx context.Context, hash string) (*models.Block, error) {
	if hash == "" {
		return nil, fmt.Errorf("block hash cannot be empty")
	}

	block, err := s.blockRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash: %w", err)
	}

	return block, nil
}

// GetBlockByHeight 根据高度获取区块
func (s *blockService) GetBlockByHeight(ctx context.Context, height uint64) (*models.Block, error) {
	block, err := s.blockRepo.GetByHeight(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by height: %w", err)
	}

	return block, nil
}

// GetLatestBlock 获取最新区块
func (s *blockService) GetLatestBlock(ctx context.Context, chain string) (*models.Block, error) {
	block, err := s.blockRepo.GetLatest(ctx, chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	return block, nil
}

// ListBlocks 分页查询区块列表
func (s *blockService) ListBlocks(ctx context.Context, page, pageSize int, chain string) ([]*models.Block, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	blocks, total, err := s.blockRepo.List(ctx, offset, pageSize, chain)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list blocks: %w", err)
	}

	return blocks, total, nil
}

// CreateBlock 创建区块
func (s *blockService) CreateBlock(ctx context.Context, block *models.Block) error {
	if block == nil {
		return fmt.Errorf("block cannot be nil")
	}

	if block.Hash == "" {
		return fmt.Errorf("block hash cannot be empty")
	}

	if err := s.blockRepo.Create(ctx, block); err != nil {
		return fmt.Errorf("failed to create block: %w", err)
	}

	return nil
}

// UpdateBlock 更新区块
func (s *blockService) UpdateBlock(ctx context.Context, block *models.Block) error {
	if block == nil {
		return fmt.Errorf("block cannot be nil")
	}

	if err := s.blockRepo.Update(ctx, block); err != nil {
		return fmt.Errorf("failed to update block: %w", err)
	}

	return nil
}

// DeleteBlock 删除区块
func (s *blockService) DeleteBlock(ctx context.Context, hash string) error {
	if hash == "" {
		return fmt.Errorf("block hash cannot be empty")
	}

	if err := s.blockRepo.Delete(ctx, hash); err != nil {
		return fmt.Errorf("failed to delete block: %w", err)
	}

	return nil
}
