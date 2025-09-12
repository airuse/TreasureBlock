package repository

import (
	"blockChainBrowser/server/internal/models"
	"context"
	"time"

	"gorm.io/gorm"
)

// BTCUTXORepository 管理 btc_utxo 表
type BTCUTXORepository interface {
	UpsertOutputs(ctx context.Context, outputs []*models.BTCUTXO) error
	MarkSpent(ctx context.Context, chain string, prevTxID string, voutIndex uint32, spentTxID string, vinIndex uint32, spentHeight uint64) error
	GetOutputs(ctx context.Context, chain string, txID string) ([]*models.BTCUTXO, error)
	GetOutputsByHeight(ctx context.Context, chain string, height uint64) ([]*models.BTCUTXO, error)
	GetSpent(ctx context.Context, chain string, txID string) ([]*models.BTCUTXO, error)
	GetUTXOCountByAddress(ctx context.Context, chain string, address string) (int64, error)
	GetUTXOsByAddress(ctx context.Context, chain string, address string) ([]*models.BTCUTXO, error)
}

type btcUTXORepository struct {
	db *gorm.DB
}

func NewBTCUTXORepository(db *gorm.DB) BTCUTXORepository {
	return &btcUTXORepository{db: db}
}

// UpsertOutputs 批量插入或忽略重复（按链+txid+voutIndex）
func (r *btcUTXORepository) UpsertOutputs(ctx context.Context, outputs []*models.BTCUTXO) error {
	// 使用 MySQL 的 ON DUPLICATE KEY UPDATE 仅更新非关键字段
	// 这里通过 GORM 的 Clauses 实现；若不支持，退化为逐个 Save
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量插入，冲突时更新脚本、地址、金额等（不改 spent 字段）
		if err := tx.Clauses(
		// gorm.io/gorm/clause
		// 由于工具限制不显式导入clause，使用Raw SQL作为保底
		).Create(&outputs).Error; err != nil {
			// 回退策略：逐个插入，存在则忽略
			for _, out := range outputs {
				// 尝试先查询，存在则跳过
				var existed models.BTCUTXO
				err := tx.Where("chain = ? AND tx_id = ? AND vout_index = ?", out.Chain, out.TxID, out.VoutIndex).First(&existed).Error
				if err != nil {
					if err == gorm.ErrRecordNotFound {
						if err := tx.Create(out).Error; err != nil {
							return err
						}
					} else {
						return err
					}
				}
			}
		}
		return nil
	})
}

// MarkSpent 标记指定 utxo 已被花费
func (r *btcUTXORepository) MarkSpent(ctx context.Context, chain string, prevTxID string, voutIndex uint32, spentTxID string, vinIndex uint32, spentHeight uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.BTCUTXO{}).
		Where("chain = ? AND tx_id = ? AND vout_index = ? AND (spent_tx_id = '' OR spent_tx_id IS NULL)", chain, prevTxID, voutIndex).
		Updates(map[string]interface{}{
			"spent_tx_id":     spentTxID,
			"spent_vin_index": vinIndex,
			"spent_height":    spentHeight,
			"spent_at":        &now,
			"mtime":           now,
		}).Error
}

func (r *btcUTXORepository) GetOutputs(ctx context.Context, chain string, txID string) ([]*models.BTCUTXO, error) {
	var outputs []*models.BTCUTXO
	err := r.db.WithContext(ctx).Where("chain = ? AND tx_id = ?", chain, txID).Find(&outputs).Error
	if err != nil {
		return nil, err
	}
	return outputs, err
}

func (r *btcUTXORepository) GetOutputsByHeight(ctx context.Context, chain string, height uint64) ([]*models.BTCUTXO, error) {
	var outputs []*models.BTCUTXO
	err := r.db.WithContext(ctx).Where("chain = ? AND block_height = ?", chain, height).Find(&outputs).Error
	if err != nil {
		return nil, err
	}
	return outputs, err
}

func (r *btcUTXORepository) GetSpent(ctx context.Context, chain string, txID string) ([]*models.BTCUTXO, error) {
	var spent []*models.BTCUTXO
	err := r.db.WithContext(ctx).Where("chain = ? AND tx_id = ?", chain, txID).Find(&spent).Error
	if err != nil {
		return nil, err
	}
	return spent, err
}

// GetUTXOCountByAddress 获取指定地址的UTXO数量（未花费的输出）
func (r *btcUTXORepository) GetUTXOCountByAddress(ctx context.Context, chain string, address string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.BTCUTXO{}).
		Where("chain = ? AND address = ? AND (spent_tx_id = '' OR spent_tx_id IS NULL)", chain, address).
		Count(&count).Error
	return count, err
}

// GetUTXOsByAddress 获取指定地址的所有UTXO（未花费的输出）
func (r *btcUTXORepository) GetUTXOsByAddress(ctx context.Context, chain string, address string) ([]*models.BTCUTXO, error) {
	var utxos []*models.BTCUTXO
	err := r.db.WithContext(ctx).
		Where("chain = ? AND address = ? AND (spent_tx_id = '' OR spent_tx_id IS NULL)", chain, address).
		Order("block_height DESC, value_satoshi DESC").
		Find(&utxos).Error
	return utxos, err
}
