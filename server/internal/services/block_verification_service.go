package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"blockChainBrowser/server/config"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"blockChainBrowser/server/internal/utils"

	common "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
)

// BlockVerificationService 区块验证服务接口
type BlockVerificationService interface {
	VerifyBlock(ctx context.Context, blockID uint64) (*BlockVerificationResult, error)
	GetLastVerifiedBlockHeight(ctx context.Context, chain string) (uint64, error)
	UpdateBlockVerificationStatus(ctx context.Context, blockID uint64, isVerified bool, reason string) error
	HandleTimeoutBlocks(ctx context.Context, chain string, height uint64) error
	GetBlockByID(ctx context.Context, blockID uint64) (*models.Block, error)
}

// BlockVerificationResult 区块验证结果
type BlockVerificationResult struct {
	BlockID      uint64 `json:"block_id"`
	IsValid      bool   `json:"is_valid"`
	Reason       string `json:"reason"`
	Details      string `json:"details"`
	Transactions int    `json:"transactions"`
	Receipts     int    `json:"receipts"`
}

// blockVerificationService 区块验证服务实现
type blockVerificationService struct {
	blockRepo      repository.BlockRepository
	txRepo         repository.TransactionRepository
	receiptRepo    repository.TransactionReceiptRepository
	coinConfigRepo repository.CoinConfigRepository
}

// NewBlockVerificationService 创建区块验证服务
func NewBlockVerificationService(
	blockRepo repository.BlockRepository,
	txRepo repository.TransactionRepository,
	receiptRepo repository.TransactionReceiptRepository,
	coinConfigRepo repository.CoinConfigRepository,
) BlockVerificationService {
	return &blockVerificationService{
		blockRepo:      blockRepo,
		txRepo:         txRepo,
		receiptRepo:    receiptRepo,
		coinConfigRepo: coinConfigRepo,
	}
}

// VerifyBlock 验证区块
func (s *blockVerificationService) VerifyBlock(ctx context.Context, blockID uint64) (*BlockVerificationResult, error) {
	// 获取区块信息
	block, err := s.blockRepo.GetByID(ctx, blockID)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	// 获取区块的所有交易
	transactions, err := s.txRepo.GetByBlockID(ctx, blockID)
	if err != nil {
		return nil, fmt.Errorf("failed to get block transactions: %w", err)
	}
	// 获取区块的所有交易凭证
	if block.Chain == "eth" {
		receipts, err := s.receiptRepo.GetByBlockNumber(ctx, block.Height, block.Chain)
		if err != nil {
			return nil, fmt.Errorf("failed to get block receipts: %w", err)
		}
		result := &BlockVerificationResult{
			BlockID:      blockID,
			Transactions: len(transactions),
			Receipts:     len(receipts),
		}

		// 执行验证
		if err := s.performETHVerification(block, transactions, receipts); err != nil {
			result.IsValid = false
			result.Reason = "验证失败"
			result.Details = err.Error()

			// 验证失败时，更新哈希后缀并标记为失败
			if err := s.handleVerificationFailure(ctx, block, err.Error()); err != nil {
				return nil, fmt.Errorf("failed to handle verification failure: %w", err)
			}

			return result, nil
		}

		result.IsValid = true
		result.Reason = "验证通过"
		result.Details = "所有验证项均通过"

		return result, nil
	} else if block.Chain == "bsc" {
		receipts, err := s.receiptRepo.GetByBlockNumber(ctx, block.Height, block.Chain)
		if err != nil {
			return nil, fmt.Errorf("failed to get block receipts: %w", err)
		}
		result := &BlockVerificationResult{
			BlockID:      blockID,
			Transactions: len(transactions),
			Receipts:     len(receipts),
		}

		// 执行验证
		if err := s.performETHVerification(block, transactions, receipts); err != nil {
			result.IsValid = false
			result.Reason = "验证失败"
			result.Details = err.Error()

			// 验证失败时，更新哈希后缀并标记为失败
			if err := s.handleVerificationFailure(ctx, block, err.Error()); err != nil {
				return nil, fmt.Errorf("failed to handle verification failure: %w", err)
			}

			return result, nil
		}

		result.IsValid = true
		result.Reason = "验证通过"
		result.Details = "所有验证项均通过"

		return result, nil
	} else if block.Chain == "btc" {
		result := &BlockVerificationResult{
			BlockID:      blockID,
			Transactions: len(transactions),
			Receipts:     0,
		}

		// 执行验证
		if err := s.performBTCVerification(block, transactions); err != nil {
			result.IsValid = false
			result.Reason = "验证失败"
			result.Details = err.Error()

			// 验证失败时，更新哈希后缀并标记为失败
			if err := s.handleVerificationFailure(ctx, block, err.Error()); err != nil {
				return nil, fmt.Errorf("failed to handle verification failure: %w", err)
			}

			return result, nil
		}

		result.IsValid = true
		result.Reason = "验证通过"
		result.Details = "所有验证项均通过"
		return result, nil
	} else {
		return nil, fmt.Errorf("不支持的链类型: %s", block.Chain)
	}
}

// handleVerificationFailure 处理验证失败的情况
func (s *blockVerificationService) handleVerificationFailure(ctx context.Context, block *models.Block, reason string) error {
	// 标记为验证失败
	if err := s.blockRepo.UpdateVerificationStatus(ctx, uint64(block.ID), 2, reason); err != nil {
		return fmt.Errorf("failed to update verification status: %w", err)
	}

	// 尝试更新哈希，如果失败则重试
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		// 生成递增的哈希后缀
		newHash, err := s.generateIncrementalHash(ctx, block.Hash, block.Height, block.Chain)
		if err != nil {
			return fmt.Errorf("failed to generate incremental hash: %w", err)
		}

		// 更新哈希为失败状态（带递增后缀）
		if err := s.blockRepo.UpdateBlockHash(ctx, uint64(block.ID), newHash); err != nil {
			// 如果是唯一性约束冲突，重试
			if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062") {
				// 等待一小段时间后重试
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return fmt.Errorf("failed to update block hash after %d retries: %w", i+1, err)
		}

		// 更新成功，退出循环
		break
	}

	// 逻辑删除该区块下的所有交易和凭证
	if err := s.logicalDeleteBlockData(ctx, uint64(block.ID)); err != nil {
		return fmt.Errorf("failed to logically delete block data: %w", err)
	}

	return nil
}

// performVerification 执行具体的验证逻辑
func (s *blockVerificationService) performETHVerification(
	block *models.Block,
	transactions []*models.Transaction,
	receipts []*models.TransactionReceipt,
) error {
	// 1. 验证交易数量
	if len(transactions) != int(block.TransactionCount) {
		return fmt.Errorf("交易数量不匹配: 期望 %d, 实际 %d", block.TransactionCount, len(transactions))
	}

	// 2. 验证交易顺序和索引
	if err := s.verifyTransactionOrder(transactions); err != nil {
		return fmt.Errorf("交易顺序验证失败: %w", err)
	}

	return s.verifyEthereumBlock(block, transactions, receipts)
}

func (s *blockVerificationService) performBTCVerification(
	block *models.Block,
	transactions []*models.Transaction,
) error {
	// return s.verifyBitcoinBlock(block, transactions)
	return nil
}

// verifyTransactionOrder 验证交易顺序
func (s *blockVerificationService) verifyTransactionOrder(transactions []*models.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	// 按区块索引排序
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].BlockIndex < transactions[j].BlockIndex
	})

	// 验证索引连续性
	for i, tx := range transactions {
		if uint(i) != tx.BlockIndex {
			return fmt.Errorf("交易索引不连续: 期望 %d, 实际 %d", i, tx.BlockIndex)
		}
	}

	return nil
}

// verifyEthereumBlock 验证以太坊区块
func (s *blockVerificationService) verifyEthereumBlock(
	block *models.Block,
	transactions []*models.Transaction,
	receipts []*models.TransactionReceipt,
) error {
	// 0. 可选：用RPC校验交易列表（只调一次RPC）
	if err := s.verifyEthTransactionsWithRPC(block, transactions); err != nil {
		return fmt.Errorf("交易列表校验失败: %w", err)
	}

	// 2. 验证收据根哈希（本地重建）
	if err := s.verifyReceiptsRoot(block, receipts); err != nil {
		return fmt.Errorf("收据根验证失败: %w", err)
	}

	return nil
}

// verifyEthTransactionsWithRPC 使用eth_getBlockByHash校验交易哈希序列（只调一次RPC）
func (s *blockVerificationService) verifyEthTransactionsWithRPC(block *models.Block, txs []*models.Transaction) error {
	if strings.ToLower(block.Chain) != "eth" {
		return nil
	}

	// 从配置文件读取链配置，若无配置则跳过
	if _, exists := config.AppConfig.Blockchain.Chains[strings.ToLower(block.Chain)]; !exists {
		return nil
	}

	// 使用ETH故障转移管理器
	fo, err := utils.NewEthFailoverFromChain(strings.ToLower(block.Chain))
	if err != nil {
		return fmt.Errorf("初始化ETH故障转移失败: %w", err)
	}
	defer fo.Close()

	ctx := context.Background()
	b, err := fo.BlockByHash(ctx, common.HexToHash(block.Hash))
	if err != nil {
		return fmt.Errorf("获取区块失败: %w", err)
	}
	ethTxs := b.Transactions()
	if len(ethTxs) != len(txs) {
		return fmt.Errorf("交易数量不一致: 链上%d vs 本地%d", len(ethTxs), len(txs))
	}
	// 按BlockIndex排序本地交易
	sort.Slice(txs, func(i, j int) bool { return txs[i].BlockIndex < txs[j].BlockIndex })
	for i := 0; i < len(ethTxs); i++ {
		chainHash := strings.TrimPrefix(ethTxs[i].Hash().Hex(), "0x")
		localHash := strings.TrimPrefix(txs[i].TxID, "0x")
		if !strings.EqualFold(chainHash, localHash) {
			return fmt.Errorf("交易哈希不一致@%d: 链上%s vs 本地%s", i, chainHash, localHash)
		}
	}
	return nil
}

// verifyBitcoinBlock 验证比特币区块
func (s *blockVerificationService) verifyBitcoinBlock(
	block *models.Block,
	transactions []*models.Transaction,
) error {
	// 0. 可选：用RPC校验区块信息（只调一次RPC）
	if err := s.verifyBtcBlockWithRPC(block, transactions); err != nil {
		return fmt.Errorf("区块RPC校验失败: %w", err)
	}

	// 1. 验证默克尔根
	if err := s.verifyMerkleRoot(block, transactions); err != nil {
		return fmt.Errorf("默克尔根验证失败: %w", err)
	}

	// 2. 验证交易数量
	if len(transactions) == 0 {
		return fmt.Errorf("比特币区块必须包含至少一笔交易（coinbase）")
	}

	// 3. 验证第一笔交易是否为coinbase
	firstTx := transactions[0]
	if firstTx.BlockIndex != 0 {
		return fmt.Errorf("第一笔交易索引必须为0")
	}

	return nil
}

// verifyBtcBlockWithRPC 使用Bitcoin RPC校验区块信息
func (s *blockVerificationService) verifyBtcBlockWithRPC(block *models.Block, txs []*models.Transaction) error {
	if strings.ToLower(block.Chain) != "btc" {
		return nil
	}

	// 从配置获取可能的多个RPC地址
	chainConfig, exists := config.AppConfig.Blockchain.Chains[strings.ToLower(block.Chain)]
	if !exists {
		return nil
	}

	urls := make([]string, 0)
	if len(chainConfig.RPCURLs) > 0 {
		urls = append(urls, chainConfig.RPCURLs...)
	}
	if chainConfig.RPCURL != "" {
		urls = append(urls, chainConfig.RPCURL)
	}
	if len(urls) == 0 {
		return nil
	}

	// 构建RPC请求
	rpcRequest := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "block_verification",
		"method":  "getblock",
		"params":  []interface{}{block.Hash, 1},
	}

	jsonData, err := json.Marshal(rpcRequest)
	if err != nil {
		return fmt.Errorf("序列化RPC请求失败: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	var lastErr error
	var rpcResponse struct {
		Error *struct {
			Code    int
			Message string
		} `json:"error"`
		Result *struct {
			Hash              string   `json:"hash"`
			Height            int64    `json:"height"`
			Tx                []string `json:"tx"`
			Merkleroot        string   `json:"merkleroot"`
			Confirmations     int      `json:"confirmations"`
			Size              int      `json:"size"`
			Strippedsize      int      `json:"strippedsize"`
			Weight            int      `json:"weight"`
			Version           int      `json:"version"`
			VersionHex        string   `json:"versionHex"`
			Time              int64    `json:"time"`
			MedianTime        int64    `json:"mediantime"`
			Nonce             uint32   `json:"nonce"`
			Bits              string   `json:"bits"`
			Difficulty        float64  `json:"difficulty"`
			Chainwork         string   `json:"chainwork"`
			NTx               int      `json:"nTx"`
			Previousblockhash string   `json:"previousblockhash"`
		} `json:"result"`
	}

	for _, url := range urls {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("创建HTTP请求失败: %w", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if chainConfig.Username != "" && chainConfig.Password != "" {
			req.SetBasicAuth(chainConfig.Username, chainConfig.Password)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("发送RPC请求失败(%s): %w", url, err)
			continue
		}
		// Ensure body closed for each attempt
		func() {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("RPC响应状态错误(%s): %d", url, resp.StatusCode)
				return
			}
			// reset previous response
			rpcResponse.Error = nil
			rpcResponse.Result = nil
			if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
				lastErr = fmt.Errorf("解析RPC响应失败(%s): %w", url, err)
				return
			}
			lastErr = nil
		}()

		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return lastErr
	}

	if rpcResponse.Error != nil {
		return fmt.Errorf("RPC错误: %d - %s", rpcResponse.Error.Code, rpcResponse.Error.Message)
	}

	if rpcResponse.Result == nil {
		return fmt.Errorf("RPC响应中没有结果数据")
	}

	result := rpcResponse.Result

	// 验证区块高度
	if result.Height != int64(block.Height) {
		return fmt.Errorf("区块高度不匹配: 期望 %d, 实际 %d", block.Height, result.Height)
	}

	// 验证交易数量
	if result.NTx != len(txs) {
		return fmt.Errorf("交易数量不一致: 链上%d vs 本地%d", result.NTx, len(txs))
	}

	// 验证默克尔根
	if result.Merkleroot != block.MerkleRoot {
		return fmt.Errorf("默克尔根不匹配: 期望 %s, 实际 %s", block.MerkleRoot, result.Merkleroot)
	}

	// 验证交易哈希列表（按BlockIndex排序）
	sort.Slice(txs, func(i, j int) bool { return txs[i].BlockIndex < txs[j].BlockIndex })
	for i, txHash := range result.Tx {
		chainHash := strings.TrimPrefix(txHash, "0x")
		localHash := strings.TrimPrefix(txs[i].TxID, "0x")
		if !strings.EqualFold(chainHash, localHash) {
			return fmt.Errorf("交易哈希不一致@%d: 链上%s vs 本地%s", i, chainHash, localHash)
		}
	}

	return nil
}

// verifyReceiptsRoot 验证收据根哈希（以太坊）
func (s *blockVerificationService) verifyReceiptsRoot(block *models.Block, receipts []*models.TransactionReceipt) error {
	if strings.ToLower(block.Chain) != "eth" {
		return nil
	}
	if block.ReceiptsRoot == "" {
		return nil
	}
	// 空块直接跳过
	if block.TransactionCount == 0 {
		return nil
	}
	// 数量校验
	if len(receipts) != int(block.TransactionCount) {
		return fmt.Errorf("收据数量不匹配: 期望 %d, 实际 %d", block.TransactionCount, len(receipts))
	}

	// 排序校验
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].TransactionIndex < receipts[j].TransactionIndex })

	// 组装 go-ethereum types.Receipt 列表
	calculated := s.calculateReceiptsRoot(receipts)
	expected := strings.TrimPrefix(block.ReceiptsRoot, "0x")

	if calculated != expected {
		return fmt.Errorf("收据根不匹配: 期望 %s, 实际 %s", block.ReceiptsRoot, calculated)
	}
	return nil
}

// verifyMerkleRoot 验证默克尔根
func (s *blockVerificationService) verifyMerkleRoot(block *models.Block, transactions []*models.Transaction) error {
	if block.MerkleRoot == "" {
		return nil // 如果没有默克尔根，跳过验证
	}

	// 构建默克尔根
	calculatedRoot := s.calculateMerkleRoot(transactions)
	if calculatedRoot != block.MerkleRoot {
		return fmt.Errorf("默克尔根不匹配: 期望 %s, 实际 %s", block.MerkleRoot, calculatedRoot)
	}

	return nil
}

// calculateReceiptsRoot 基于本地存储的收据重建以太坊Receipts Trie并计算真实根
func (s *blockVerificationService) calculateReceiptsRoot(receipts []*models.TransactionReceipt) string {
	if len(receipts) == 0 {
		return ""
	}
	// 构造 go-ethereum receipts
	ethReceipts := make(ethtypes.Receipts, len(receipts))
	for i, r := range receipts {
		gr := &ethtypes.Receipt{}
		gr.Type = r.TxType
		if r.PostState != "" && r.PostState != "0x" {
			if bs, err := hex.DecodeString(strings.TrimPrefix(r.PostState, "0x")); err == nil {
				gr.PostState = bs
			}
		} else {
			if r.Status == 1 {
				gr.Status = ethtypes.ReceiptStatusSuccessful
			} else {
				gr.Status = ethtypes.ReceiptStatusFailed
			}
		}
		if b, err := hex.DecodeString(strings.TrimPrefix(r.Bloom, "0x")); err == nil {
			var bloom ethtypes.Bloom
			copy(bloom[:], b)
			gr.Bloom = bloom
		}
		// 处理 CumulativeGasUsed：现在数据库中是十进制数字
		var cumulativeGas uint64
		if r.CumulativeGasUsed == "" || r.CumulativeGasUsed == "0" {
			// 如果为空或0，根据 GasUsed 计算正确的值
			for j := 0; j <= i; j++ {
				cumulativeGas += receipts[j].GasUsed
			}
		} else {
			// 数据库中是十进制字符串，直接解析
			if val, err := strconv.ParseUint(r.CumulativeGasUsed, 10, 64); err == nil {
				cumulativeGas = val
			} else {
				// 解析失败，使用计算值
				for j := 0; j <= i; j++ {
					cumulativeGas += receipts[j].GasUsed
				}
			}
		}
		gr.CumulativeGasUsed = cumulativeGas

		gr.Logs = s.parseEthLogs(r.LogsData)
		gr.TxHash = common.HexToHash(r.TxHash)
		gr.TransactionIndex = uint(r.TransactionIndex)
		gr.BlockNumber = new(big.Int)
		gr.BlockNumber.SetUint64(r.BlockNumber)
		gr.BlockHash = common.HexToHash(r.BlockHash)
		gr.GasUsed = r.GasUsed

		// 处理 EffectiveGasPrice：支持16进制字符串
		gr.EffectiveGasPrice = new(big.Int)
		if strings.HasPrefix(r.EffectiveGasPrice, "0x") {
			// 16进制字符串，去掉0x前缀后解析
			if _, ok := gr.EffectiveGasPrice.SetString(strings.TrimPrefix(r.EffectiveGasPrice, "0x"), 16); !ok {
				gr.EffectiveGasPrice.SetInt64(0)
			}
		} else {
			// 十进制字符串
			if _, ok := gr.EffectiveGasPrice.SetString(r.EffectiveGasPrice, 10); !ok {
				gr.EffectiveGasPrice.SetInt64(0)
			}
		}
		gr.BlobGasUsed = r.BlobGasUsed
		gr.BlobGasPrice = new(big.Int)
		gr.BlobGasPrice.SetString(r.BlobGasPrice, 10)
		gr.ContractAddress = common.HexToAddress(r.ContractAddress)

		ethReceipts[i] = gr
	}
	// 使用栈trie哈希器（无需底层db）
	hasher := trie.NewStackTrie(nil)
	rootHash := ethtypes.DeriveSha(ethReceipts, hasher)
	return strings.TrimPrefix(rootHash.Hex(), "0x")
}

// parseEthLogs 将我们存储的日志JSON转为go-ethereum的logs（尽量容错）
func (s *blockVerificationService) parseEthLogs(logsJSON string) []*ethtypes.Log {
	if len(logsJSON) == 0 {
		return nil
	}

	type topicAlias []string
	type rawLog struct {
		Address     string     `json:"address"`
		Topics      topicAlias `json:"topics"`
		Data        string     `json:"data"`
		BlockNumber string     `json:"blockNumber"` // 改为string，处理16进制
		TxHash      string     `json:"transactionHash"`
		TxIndex     string     `json:"transactionIndex"` // 改为string，处理16进制
		BlockHash   string     `json:"blockHash"`
		Index       string     `json:"logIndex"` // 改为string，处理16进制
		Removed     bool       `json:"removed"`
	}

	var raws []rawLog
	if err := json.Unmarshal([]byte(logsJSON), &raws); err != nil {
		return nil
	}

	out := make([]*ethtypes.Log, 0, len(raws))
	for _, rl := range raws {
		lg := &ethtypes.Log{}
		lg.Address = common.HexToAddress(rl.Address)
		// topics
		for _, t := range rl.Topics {
			lg.Topics = append(lg.Topics, common.HexToHash(t))
		}
		// data
		if rl.Data != "" {
			if b, err := hex.DecodeString(strings.TrimPrefix(rl.Data, "0x")); err == nil {
				lg.Data = b
			}
		}
		// 处理BlockNumber：16进制字符串转uint64
		if rl.BlockNumber != "" {
			if strings.HasPrefix(rl.BlockNumber, "0x") {
				if val, err := strconv.ParseUint(strings.TrimPrefix(rl.BlockNumber, "0x"), 16, 64); err == nil {
					lg.BlockNumber = val
				}
			} else {
				if val, err := strconv.ParseUint(rl.BlockNumber, 10, 64); err == nil {
					lg.BlockNumber = val
				}
			}
		}
		lg.TxHash = common.HexToHash(rl.TxHash)
		// 处理TxIndex：16进制字符串转uint
		if rl.TxIndex != "" {
			if strings.HasPrefix(rl.TxIndex, "0x") {
				if val, err := strconv.ParseUint(strings.TrimPrefix(rl.TxIndex, "0x"), 16, 32); err == nil {
					lg.TxIndex = uint(val)
				}
			} else {
				if val, err := strconv.ParseUint(rl.TxIndex, 10, 32); err == nil {
					lg.TxIndex = uint(val)
				}
			}
		}
		lg.BlockHash = common.HexToHash(rl.BlockHash)
		// 处理Index：16进制字符串转uint
		if rl.Index != "" {
			if strings.HasPrefix(rl.Index, "0x") {
				if val, err := strconv.ParseUint(strings.TrimPrefix(rl.Index, "0x"), 16, 32); err == nil {
					lg.Index = uint(val)
				}
			} else {
				if val, err := strconv.ParseUint(rl.Index, 10, 32); err == nil {
					lg.Index = uint(val)
				}
			}
		}
		lg.Removed = rl.Removed
		out = append(out, lg)
	}

	return out
}

// calculateMerkleRoot 计算默克尔根
func (s *blockVerificationService) calculateMerkleRoot(transactions []*models.Transaction) string {
	if len(transactions) == 0 {
		return ""
	}

	// 按区块索引排序
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].BlockIndex < transactions[j].BlockIndex
	})

	// 构建交易哈希列表
	txHashes := make([][]byte, len(transactions))
	for i, tx := range transactions {
		hash, _ := hex.DecodeString(tx.TxID)
		txHashes[i] = hash
	}

	// 计算默克尔根
	root := s.calculateMerkleRootFromHashes(txHashes)
	return hex.EncodeToString(root)
}

// calculateMerkleRootFromHashes 从哈希列表计算默克尔根
func (s *blockVerificationService) calculateMerkleRootFromHashes(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return nil
	}

	if len(hashes) == 1 {
		return hashes[0]
	}

	// 如果哈希数量为奇数，复制最后一个
	if len(hashes)%2 == 1 {
		hashes = append(hashes, hashes[len(hashes)-1])
	}

	// 计算下一层哈希
	nextLevel := make([][]byte, len(hashes)/2)
	for i := 0; i < len(hashes); i += 2 {
		combined := append(hashes[i], hashes[i+1]...)
		hash := sha256.Sum256(combined)
		nextLevel[i/2] = hash[:]
	}

	// 递归计算
	return s.calculateMerkleRootFromHashes(nextLevel)
}

// GetLastVerifiedBlockHeight 获取最后一个验证通过的区块高度
func (s *blockVerificationService) GetLastVerifiedBlockHeight(ctx context.Context, chain string) (uint64, error) {
	block, err := s.blockRepo.GetLastVerifiedBlock(ctx, chain)
	if err != nil {
		return 0, err
	}
	return block.Height, nil
}

// UpdateBlockVerificationStatus 更新区块验证状态
func (s *blockVerificationService) UpdateBlockVerificationStatus(ctx context.Context, blockID uint64, isVerified bool, reason string) error {
	var status uint8
	if isVerified {
		status = 1
	} else {
		status = 2
	}

	return s.blockRepo.UpdateVerificationStatus(ctx, blockID, status, reason)
}

// HandleTimeoutBlocks 处理超时的区块，标记为失败并逻辑删除
func (s *blockVerificationService) HandleTimeoutBlocks(ctx context.Context, chain string, height uint64) error {
	// 获取所有超时的区块（未验证且超过验证截止时间）
	timeoutBlocks, err := s.blockRepo.GetTimeoutBlocks(ctx, chain, height)
	if err != nil {
		return fmt.Errorf("failed to get timeout blocks: %w", err)
	}

	// 处理每个超时的区块
	for _, block := range timeoutBlocks {
		// 生成递增的后缀，避免重复
		newHash, err := s.generateIncrementalHash(ctx, block.Hash, block.Height, block.Chain)
		if err != nil {
			return fmt.Errorf("failed to generate incremental hash for block %d: %w", block.ID, err)
		}

		// 标记为验证失败
		if err := s.blockRepo.UpdateVerificationStatus(ctx, uint64(block.ID), 2, "验证超时"); err != nil {
			return fmt.Errorf("failed to update verification status for block %d: %w", block.ID, err)
		}

		// 更新哈希为失败状态（带递增后缀）
		if err := s.blockRepo.UpdateBlockHash(ctx, uint64(block.ID), newHash); err != nil {
			return fmt.Errorf("failed to update block hash for block %d: %w", block.ID, err)
		}

		// 逻辑删除区块下的所有交易和凭证
		if err := s.logicalDeleteBlockData(ctx, uint64(block.ID)); err != nil {
			return fmt.Errorf("failed to logically delete block data for block %d: %w", block.ID, err)
		}

		// 逻辑删除区块（设置deleted_at）
		if err := s.blockRepo.LogicalDelete(ctx, uint64(block.ID)); err != nil {
			return fmt.Errorf("failed to logically delete block %d: %w", block.ID, err)
		}
	}

	return nil
}

// generateIncrementalHash 生成递增的哈希后缀，避免重复
func (s *blockVerificationService) generateIncrementalHash(ctx context.Context, originalHash string, height uint64, chain string) (string, error) {
	// 基础后缀格式：_failed_<timestamp>_<counter>
	baseSuffix := "_failed"

	// 添加时间戳确保唯一性
	timestamp := time.Now().Unix()

	// 查找当前高度下已存在的失败区块数量
	failedCount, err := s.blockRepo.GetFailedBlockCountByHeight(ctx, height, chain)
	if err != nil {
		return "", fmt.Errorf("failed to get failed block count: %w", err)
	}

	// 生成递增后缀：_failed_<timestamp>_<count+1>
	suffix := fmt.Sprintf("%s_%d_%d", baseSuffix, timestamp, failedCount+1)

	// 组合新哈希：原始哈希 + 后缀
	newHash := originalHash + suffix

	return newHash, nil
}

// GetBlockByID 根据区块ID获取区块信息
func (s *blockVerificationService) GetBlockByID(ctx context.Context, blockID uint64) (*models.Block, error) {
	return s.blockRepo.GetByID(ctx, blockID)
}

// logicalDeleteBlockData 逻辑删除区块下的所有交易和凭证
func (s *blockVerificationService) logicalDeleteBlockData(ctx context.Context, blockID uint64) error {
	// 逻辑删除该区块下的所有交易
	if err := s.txRepo.LogicalDeleteByBlockID(ctx, blockID); err != nil {
		return fmt.Errorf("failed to logically delete transactions: %w", err)
	}

	// 逻辑删除该区块下的所有凭证
	if err := s.receiptRepo.LogicalDeleteByBlockID(ctx, blockID); err != nil {
		return fmt.Errorf("failed to logically delete receipts: %w", err)
	}

	return nil
}
