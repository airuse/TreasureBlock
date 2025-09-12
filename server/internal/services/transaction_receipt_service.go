package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
)

// TransactionReceiptService 交易凭证服务接口
type TransactionReceiptService interface {
	CreateTransactionReceipt(ctx context.Context, receiptData map[string]interface{}) error
	CreateTransactionReceiptsBatch(ctx context.Context, receiptsData []map[string]interface{}) error
	GetTransactionReceiptByHash(ctx context.Context, txHash string) (*models.TransactionReceipt, error)
	GetTransactionReceiptsByBlockHash(ctx context.Context, blockHash string) ([]*models.TransactionReceipt, error)
	GetTransactionReceiptsByBlockNumber(ctx context.Context, blockNumber uint64, chain string) ([]*models.TransactionReceipt, error)
}

// transactionReceiptService 交易凭证服务实现
type transactionReceiptService struct {
	receiptRepo repository.TransactionReceiptRepository
}

// NewTransactionReceiptService 创建新的交易凭证服务
func NewTransactionReceiptService(receiptRepo repository.TransactionReceiptRepository) TransactionReceiptService {
	return &transactionReceiptService{
		receiptRepo: receiptRepo,
	}
}

// CreateTransactionReceipt 创建交易凭证
func (s *transactionReceiptService) CreateTransactionReceipt(ctx context.Context, receiptData map[string]interface{}) error {
	// 检查必需字段
	txHash, ok := receiptData["tx_hash"].(string)
	if !ok || txHash == "" {
		return fmt.Errorf("tx_hash is required")
	}

	// 检查凭证是否已存在

	// exists, err := s.receiptRepo.Exists(ctx, txHash)
	// if err != nil {
	// 	return fmt.Errorf("failed to check receipt existence: %w", err)
	// }
	// if exists {
	// 	return nil // 已存在，跳过创建
	// }

	// 转换数据为模型
	receipt := &models.TransactionReceipt{
		TxHash: txHash,
		Chain:  getStringValue(receiptData, "chain"),
	}

	// 设置可选字段
	if txType, ok := receiptData["tx_type"].(uint8); ok {
		receipt.TxType = txType
	}
	if postState, ok := receiptData["post_state"].(string); ok {
		receipt.PostState = postState
	}
	if status, ok := receiptData["status"].(uint64); ok {
		receipt.Status = status
	}
	// 处理新增字段：EffectiveGasPrice, BlobGasUsed, BlobGasPrice
	if effectiveGasPrice, ok := receiptData["effective_gas_price"].(string); ok {
		receipt.EffectiveGasPrice = effectiveGasPrice
	}
	if blobGasUsed, ok := receiptData["blob_gas_used"].(uint64); ok {
		receipt.BlobGasUsed = blobGasUsed
	}
	if blobGasPrice, ok := receiptData["blob_gas_price"].(string); ok {
		receipt.BlobGasPrice = blobGasPrice
	}
	if bloom, ok := receiptData["bloom"].(string); ok {
		receipt.Bloom = bloom
	}
	if contractAddress, ok := receiptData["contract_address"].(string); ok {
		receipt.ContractAddress = contractAddress
	}
	if gasUsed, ok := receiptData["gas_used"].(uint64); ok {
		receipt.GasUsed = gasUsed
	}
	if blockHash, ok := receiptData["block_hash"].(string); ok {
		receipt.BlockHash = blockHash
	}
	if blockNumber, ok := receiptData["block_number"].(uint64); ok {
		receipt.BlockNumber = blockNumber
	}
	if transactionIndex, ok := receiptData["transaction_index"].(uint); ok {
		receipt.TransactionIndex = transactionIndex
	}
	if cumulativeGasUsed, ok := receiptData["cumulative_gas_used"].(uint64); ok {
		receipt.CumulativeGasUsed = strconv.FormatUint(cumulativeGasUsed, 10)
	}
	if blockID, ok := receiptData["block_id"].(uint64); ok {
		receipt.BlockID = blockID
	}

	// 处理日志数据 - 转换为JSON字符串
	if logsData, ok := receiptData["logs_data"]; ok {
		if logsJSON, err := json.Marshal(logsData); err == nil {
			receipt.LogsData = string(logsJSON)
		} else {
			// 日志序列化失败，记录错误但继续处理
			return fmt.Errorf("failed to marshal logs data for tx %s: %w", txHash, err)
		}
	}

	// 创建凭证
	if err := s.receiptRepo.Create(ctx, receipt); err != nil {
		return fmt.Errorf("failed to create transaction receipt: %w", err)
	}

	return nil
}

// CreateTransactionReceiptsBatch 批量创建交易凭证
func (s *transactionReceiptService) CreateTransactionReceiptsBatch(ctx context.Context, receiptsData []map[string]interface{}) error {
	if len(receiptsData) == 0 {
		return nil
	}
	receipts := make([]*models.TransactionReceipt, 0, len(receiptsData))
	for _, receiptData := range receiptsData {
		// 复用单条校验逻辑的关键字段校验
		txHash, ok := receiptData["tx_hash"].(string)
		if !ok || txHash == "" {
			return fmt.Errorf("tx_hash is required")
		}
		// 已存在则跳过（为避免逐条 Exists 查询的 N+1，可在上层过滤或在此处简单跳过）

		// exists, err := s.receiptRepo.Exists(ctx, txHash)
		// if err != nil {
		// 	return fmt.Errorf("failed to check receipt existence: %w", err)
		// }
		// if exists {
		// 	continue
		// }

		receipt := &models.TransactionReceipt{
			TxHash: txHash,
			Chain:  getStringValue(receiptData, "chain"),
		}
		if txType, ok := receiptData["tx_type"].(uint8); ok {
			receipt.TxType = txType
		}
		if postState, ok := receiptData["post_state"].(string); ok {
			receipt.PostState = postState
		}
		if status, ok := receiptData["status"].(uint64); ok {
			receipt.Status = status
		}
		if effectiveGasPrice, ok := receiptData["effective_gas_price"].(string); ok {
			receipt.EffectiveGasPrice = effectiveGasPrice
		}
		if blobGasUsed, ok := receiptData["blob_gas_used"].(uint64); ok {
			receipt.BlobGasUsed = blobGasUsed
		}
		if blobGasPrice, ok := receiptData["blob_gas_price"].(string); ok {
			receipt.BlobGasPrice = blobGasPrice
		}
		if bloom, ok := receiptData["bloom"].(string); ok {
			receipt.Bloom = bloom
		}
		if contractAddress, ok := receiptData["contract_address"].(string); ok {
			receipt.ContractAddress = contractAddress
		}
		if gasUsed, ok := receiptData["gas_used"].(uint64); ok {
			receipt.GasUsed = gasUsed
		}
		if blockHash, ok := receiptData["block_hash"].(string); ok {
			receipt.BlockHash = blockHash
		}
		if blockNumber, ok := receiptData["block_number"].(uint64); ok {
			receipt.BlockNumber = blockNumber
		}
		if transactionIndex, ok := receiptData["transaction_index"].(uint); ok {
			receipt.TransactionIndex = transactionIndex
		}
		if cumulativeGasUsed, ok := receiptData["cumulative_gas_used"].(uint64); ok {
			receipt.CumulativeGasUsed = strconv.FormatUint(cumulativeGasUsed, 10)
		}
		if blockID, ok := receiptData["block_id"].(uint64); ok {
			receipt.BlockID = blockID
		}
		if logsData, ok := receiptData["logs_data"]; ok {
			if logsJSON, err := json.Marshal(logsData); err == nil {
				receipt.LogsData = string(logsJSON)
			} else {
				return fmt.Errorf("failed to marshal logs data for tx %s: %w", txHash, err)
			}
		}
		receipts = append(receipts, receipt)
	}
	if len(receipts) == 0 {
		return nil
	}
	return s.receiptRepo.CreateBatch(ctx, receipts)
}

// GetTransactionReceiptByHash 根据交易哈希获取凭证
func (s *transactionReceiptService) GetTransactionReceiptByHash(ctx context.Context, txHash string) (*models.TransactionReceipt, error) {
	return s.receiptRepo.GetByTxHash(ctx, txHash)
}

// GetTransactionReceiptsByBlockHash 根据区块哈希获取所有凭证
func (s *transactionReceiptService) GetTransactionReceiptsByBlockHash(ctx context.Context, blockHash string) ([]*models.TransactionReceipt, error) {
	return s.receiptRepo.GetByBlockHash(ctx, blockHash)
}

// GetTransactionReceiptsByBlockNumber 根据区块号获取所有凭证
func (s *transactionReceiptService) GetTransactionReceiptsByBlockNumber(ctx context.Context, blockNumber uint64, chain string) ([]*models.TransactionReceipt, error) {
	return s.receiptRepo.GetByBlockNumber(ctx, blockNumber, chain)
}

// getStringValue 安全地从map获取字符串值
func getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}
