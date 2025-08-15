package scanners

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"blockChainBrowser/client/scanner/config"
	"blockChainBrowser/client/scanner/internal/models"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// BitcoinScanner 比特币扫块器 - 使用标准RPC调用
type BitcoinScanner struct {
	config *config.ChainConfig
	// HTTP客户端
	httpClient *http.Client
	// 多节点支持
	currentNodeIndex int // 当前使用的外部节点索引
}

// 比特币RPC请求结构
type BitcoinRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// 比特币RPC响应结构
type BitcoinRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewBitcoinScanner 创建新的比特币扫块器
func NewBitcoinScanner(cfg *config.ChainConfig) *BitcoinScanner {
	return &BitcoinScanner{
		config:           cfg,
		currentNodeIndex: 0,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// callRPC 调用RPC接口
func (bs *BitcoinScanner) callRPC(url string, method string, params []interface{}) ([]byte, error) {
	// 判断是本地节点还是外部API
	if url == bs.config.RPCURL {
		// 本地节点RPC调用
		request := BitcoinRPCRequest{
			JSONRPC: "1.0",
			ID:      "scanner",
			Method:  method,
			Params:  params,
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal RPC request: %w", err)
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		// 使用bitcoind RPC认证
		if bs.config.Username != "" && bs.config.Password != "" {
			req.SetBasicAuth(bs.config.Username, bs.config.Password)
		}

		resp, err := bs.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to call local node: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("local node returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		return body, nil
	} else {
		// 外部API HTTP调用
		switch method {
		case "getblockcount":
			return bs.callExternalAPI(url, "blocks/tip/height")
		case "getblockhash":
			height := params[0].(uint64)
			return bs.callExternalAPI(url, fmt.Sprintf("block-height/%d", height))
		case "getblock":
			hash := params[0].(string)
			return bs.callExternalAPI(url, fmt.Sprintf("block/%s", hash))
		default:
			return nil, fmt.Errorf("unsupported method for external API: %s", method)
		}
	}
}

// callExternalAPI 调用外部API
func (bs *BitcoinScanner) callExternalAPI(url string, endpoint string) ([]byte, error) {
	fullURL := fmt.Sprintf("%s/%s", url, endpoint)
	resp, err := bs.httpClient.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call external API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// callWithFailoverString 通用的故障转移调用方法（返回string）
func (bs *BitcoinScanner) callWithFailoverString(operation string, clientCall func(string) (string, error)) (string, error) {
	// 首先尝试本地节点
	if bs.config.RPCURL != "" {
		result, err := clientCall(bs.config.RPCURL)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[BTC Scanner] Local node failed for %s: %v, trying external APIs\n", operation, err)
	}

	// 如果本地节点失败或不存在，尝试外部节点
	if len(bs.config.ExplorerAPIURLs) > 0 {
		// 从当前节点开始尝试
		startIndex := bs.currentNodeIndex
		for i := 0; i < len(bs.config.ExplorerAPIURLs); i++ {
			currentIndex := (startIndex + i) % len(bs.config.ExplorerAPIURLs)
			apiURL := bs.config.ExplorerAPIURLs[currentIndex]

			result, err := clientCall(apiURL)
			if err == nil {
				// 成功获取，更新当前节点索引
				bs.currentNodeIndex = currentIndex
				return result, nil
			}

			fmt.Printf("[BTC Scanner] External API node %d failed for %s: %v\n", currentIndex, operation, err)
		}
	}

	return "", fmt.Errorf("failed to %s: all nodes failed", operation)
}

// callWithFailoverUint64 通用的故障转移调用方法（返回uint64）
func (bs *BitcoinScanner) callWithFailoverUint64(operation string, clientCall func(string) (uint64, error)) (uint64, error) {
	// 首先尝试本地节点
	if bs.config.RPCURL != "" {
		result, err := clientCall(bs.config.RPCURL)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[BTC Scanner] Local node failed for %s: %v, trying external APIs\n", operation, err)
	}

	// 如果本地节点失败或不存在，尝试外部节点
	if len(bs.config.ExplorerAPIURLs) > 0 {
		// 从当前节点开始尝试
		startIndex := bs.currentNodeIndex
		for i := 0; i < len(bs.config.ExplorerAPIURLs); i++ {
			currentIndex := (startIndex + i) % len(bs.config.ExplorerAPIURLs)
			apiURL := bs.config.ExplorerAPIURLs[currentIndex]

			result, err := clientCall(apiURL)
			if err == nil {
				// 成功获取，更新当前节点索引
				bs.currentNodeIndex = currentIndex
				return result, nil
			}

			fmt.Printf("[BTC Scanner] External API node %d failed for %s: %v\n", currentIndex, operation, err)
		}
	}

	return 0, fmt.Errorf("failed to %s: all nodes failed", operation)
}

// callWithFailoverBlock 通用的故障转移调用方法（返回*models.Block）
func (bs *BitcoinScanner) callWithFailoverBlock(operation string, clientCall func(string) (*models.Block, error)) (*models.Block, error) {
	// 首先尝试本地节点
	if bs.config.RPCURL != "" {
		result, err := clientCall(bs.config.RPCURL)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[BTC Scanner] Local node failed for %s: %v, trying external APIs\n", operation, err)
	}

	// 如果本地节点失败或不存在，尝试外部节点
	if len(bs.config.ExplorerAPIURLs) > 0 {
		// 从当前节点开始尝试
		startIndex := bs.currentNodeIndex
		for i := 0; i < len(bs.config.ExplorerAPIURLs); i++ {
			currentIndex := (startIndex + i) % len(bs.config.ExplorerAPIURLs)
			apiURL := bs.config.ExplorerAPIURLs[currentIndex]

			result, err := clientCall(apiURL)
			if err == nil {
				// 成功获取，更新当前节点索引
				bs.currentNodeIndex = currentIndex
				return result, nil
			}

			fmt.Printf("[BTC Scanner] External API node %d failed for %s: %v\n", currentIndex, operation, err)
		}
	}

	return nil, fmt.Errorf("failed to %s: all nodes failed", operation)
}

// callWithFailoverTransactions 通用的故障转移调用方法（返回[]map[string]interface{}）
func (bs *BitcoinScanner) callWithFailoverTransactions(operation string, clientCall func(string) ([]map[string]interface{}, error)) ([]map[string]interface{}, error) {
	// 首先尝试本地节点
	if bs.config.RPCURL != "" {
		result, err := clientCall(bs.config.RPCURL)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[BTC Scanner] Local node failed for %s: %v, trying external APIs\n", operation, err)
	}

	// 如果本地节点失败或不存在，尝试外部节点
	if len(bs.config.ExplorerAPIURLs) > 0 {
		// 从当前节点开始尝试
		startIndex := bs.currentNodeIndex
		for i := 0; i < len(bs.config.ExplorerAPIURLs); i++ {
			currentIndex := (startIndex + i) % len(bs.config.ExplorerAPIURLs)
			apiURL := bs.config.ExplorerAPIURLs[currentIndex]

			result, err := clientCall(apiURL)
			if err == nil {
				// 成功获取，更新当前节点索引
				bs.currentNodeIndex = currentIndex
				return result, nil
			}

			fmt.Printf("[BTC Scanner] External API node %d failed for %s: %v\n", currentIndex, operation, err)
		}
	}

	return nil, fmt.Errorf("failed to %s: all nodes failed", operation)
}

// GetLatestBlockHeight 获取最新区块高度
func (bs *BitcoinScanner) GetLatestBlockHeight() (uint64, error) {
	result, err := bs.callWithFailoverUint64("get latest block height", func(url string) (uint64, error) {
		return bs.getBlockHeightFromURL(url)
	})

	if err == nil {
		fmt.Printf("[BTC Scanner] Latest block height: %d\n", result)
	}
	return result, err
}

// getBlockHeightFromURL 从指定URL获取区块高度
func (bs *BitcoinScanner) getBlockHeightFromURL(url string) (uint64, error) {
	if url == bs.config.RPCURL {
		// 本地节点RPC调用
		request := BitcoinRPCRequest{
			JSONRPC: "1.0",
			ID:      "scanner",
			Method:  "getblockcount",
			Params:  []interface{}{},
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal RPC request: %w", err)
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return 0, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		// 使用bitcoind RPC认证
		if bs.config.Username != "" && bs.config.Password != "" {
			req.SetBasicAuth(bs.config.Username, bs.config.Password)
		}

		resp, err := bs.httpClient.Do(req)
		if err != nil {
			return 0, fmt.Errorf("failed to call local node: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("local node returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, fmt.Errorf("failed to read response body: %w", err)
		}

		var rpcResp BitcoinRPCResponse
		if err := json.Unmarshal(body, &rpcResp); err != nil {
			return 0, fmt.Errorf("failed to parse RPC response: %w", err)
		}

		if rpcResp.Error != nil {
			return 0, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
		}

		if height, ok := rpcResp.Result.(float64); ok {
			return uint64(height), nil
		}

		return 0, fmt.Errorf("unexpected result type for block height")
	} else {
		// 外部API调用
		apiURL := fmt.Sprintf("%s/blocks/tip/height", url)
		resp, err := bs.httpClient.Get(apiURL)
		if err != nil {
			return 0, fmt.Errorf("failed to call external API: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("external API returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, fmt.Errorf("failed to read response body: %w", err)
		}

		var heightStr string
		if err := json.Unmarshal(body, &heightStr); err != nil {
			return 0, fmt.Errorf("failed to decode API response: %w", err)
		}

		height, err := strconv.ParseUint(heightStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse height from API: %w", err)
		}

		return height, nil
	}
}

// GetBlockByHeight 根据高度获取区块
func (bs *BitcoinScanner) GetBlockByHeight(height uint64) (*models.Block, error) {
	fmt.Printf("[BTC Scanner] Scanning block at height: %d\n", height)

	result, err := bs.callWithFailoverBlock("get block by height", func(url string) (*models.Block, error) {
		return bs.getBlockByHeightFromURL(url, height)
	})

	if err == nil {
		fmt.Printf("[BTC Scanner] Successfully scanned block %d (hash: %s) with %d transactions\n",
			result.Height, result.Hash[:16]+"...", result.TransactionCount)
	}
	return result, err
}

// getBlockByHeightFromURL 从指定URL根据高度获取区块
func (bs *BitcoinScanner) getBlockByHeightFromURL(url string, height uint64) (*models.Block, error) {
	if url == bs.config.RPCURL {
		// 本地节点：先获取哈希，再获取区块
		response, err := bs.callRPC(url, "getblockhash", []interface{}{height})
		if err != nil {
			return nil, fmt.Errorf("failed to get block hash for height %d: %w", height, err)
		}

		var rpcResp BitcoinRPCResponse
		if err := json.Unmarshal(response, &rpcResp); err != nil {
			return nil, fmt.Errorf("failed to parse RPC response: %w", err)
		}

		if rpcResp.Error != nil {
			return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
		}

		blockHash, ok := rpcResp.Result.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected result type for block hash")
		}

		// 获取区块详情
		return bs.GetBlockByHash(blockHash)
	} else {
		// 外部API：直接获取区块
		response, err := bs.callRPC(url, "getblock", []interface{}{height})
		if err != nil {
			return nil, fmt.Errorf("failed to get block for height %d: %w", height, err)
		}

		var blockData map[string]interface{}
		if err := json.Unmarshal(response, &blockData); err != nil {
			return nil, fmt.Errorf("failed to decode block from API: %w", err)
		}

		return bs.parseExternalBlockData(blockData)
	}
}

// GetBlockByHash 根据哈希获取区块
func (bs *BitcoinScanner) GetBlockByHash(blockHash string) (*models.Block, error) {
	fmt.Printf("[BTC Scanner] Scanning block with hash: %s\n", blockHash)

	result, err := bs.callWithFailoverBlock("get block by hash", func(url string) (*models.Block, error) {
		return bs.getBlockByHashFromURL(url, blockHash)
	})

	if err == nil {
		fmt.Printf("[BTC Scanner] Successfully scanned block %d (hash: %s) with %d transactions\n",
			result.Height, result.Hash[:16]+"...", result.TransactionCount)
	}
	return result, err
}

// getBlockByHashFromURL 从指定URL根据哈希获取区块
func (bs *BitcoinScanner) getBlockByHashFromURL(url string, blockHash string) (*models.Block, error) {
	response, err := bs.callRPC(url, "getblock", []interface{}{blockHash, 2}) // 2表示详细模式
	if err != nil {
		return nil, fmt.Errorf("failed to get block %s: %w", blockHash, err)
	}

	if url == bs.config.RPCURL {
		// 本地节点RPC响应
		var rpcResp BitcoinRPCResponse
		if err := json.Unmarshal(response, &rpcResp); err != nil {
			return nil, fmt.Errorf("failed to parse RPC response: %w", err)
		}

		if rpcResp.Error != nil {
			return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
		}

		blockData, ok := rpcResp.Result.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected result type for block data")
		}

		return bs.parseLocalBlockData(blockData)
	} else {
		// 外部API响应
		var blockData map[string]interface{}
		if err := json.Unmarshal(response, &blockData); err != nil {
			return nil, fmt.Errorf("failed to decode block from API: %w", err)
		}

		return bs.parseExternalBlockData(blockData)
	}
}

// GetBlockTransactions 获取区块交易
func (bs *BitcoinScanner) GetBlockTransactions(blockHash string) ([]map[string]interface{}, error) {
	fmt.Printf("[BTC Scanner] Getting transactions for block: %s\n", blockHash)

	result, err := bs.callWithFailoverTransactions("get block transactions", func(url string) ([]map[string]interface{}, error) {
		return bs.getBlockTransactionsFromURL(url, blockHash)
	})

	if err == nil {
		fmt.Printf("[BTC Scanner] Retrieved %d transactions from block %s\n", len(result), blockHash[:16]+"...")
	}
	return result, err
}

// GetBlockTransactionsFromBlock 直接从区块中获取交易信息（避免哈希不一致问题）
func (bs *BitcoinScanner) GetBlockTransactionsFromBlock(block *models.Block) ([]map[string]interface{}, error) {
	// 通过区块高度重新获取交易信息，避免哈希不一致问题
	result, err := bs.callWithFailoverTransactions("get block transactions by height", func(url string) ([]map[string]interface{}, error) {
		return bs.getBlockTransactionsFromURL(url, block.Hash)
	})

	if err == nil {
		fmt.Printf("[BTC Scanner] Retrieved %d transactions from block %d\n", len(result), block.Height)
	}
	return result, err
}

// getBlockTransactionsFromURL 从指定URL获取区块交易（使用btcd包改进）
func (bs *BitcoinScanner) getBlockTransactionsFromURL(url string, blockHash string) ([]map[string]interface{}, error) {
	if url == bs.config.RPCURL {
		// 本地节点：获取区块详细信息（包含交易）
		response, err := bs.callRPC(url, "getblock", []interface{}{blockHash, 2}) // 2表示详细模式，包含交易信息
		if err != nil {
			return nil, fmt.Errorf("failed to get block transactions: %w", err)
		}

		var rpcResp BitcoinRPCResponse
		if err := json.Unmarshal(response, &rpcResp); err != nil {
			return nil, fmt.Errorf("failed to parse RPC response: %w", err)
		}

		if rpcResp.Error != nil {
			return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
		}

		blockData, ok := rpcResp.Result.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected result type for block data")
		}

		// 提取交易信息并使用btcd包验证
		if txs, ok := blockData["tx"].([]interface{}); ok {
			transactions := make([]map[string]interface{}, len(txs))
			for i, tx := range txs {
				if txMap, ok := tx.(map[string]interface{}); ok {
					// 使用btcd包验证和解析交易
					validatedTx, err := bs.ParseTransaction(txMap)
					if err != nil {
						fmt.Printf("[BTC Scanner] Warning: Invalid transaction at index %d: %v\n", i, err)
						// 继续处理其他交易，不因为单个交易失败而中断
						transactions[i] = txMap
					} else {
						transactions[i] = validatedTx
					}
				}
			}
			return transactions, nil
		}

		return nil, fmt.Errorf("no transactions found in block")
	} else {
		// 外部API：获取区块交易
		response, err := bs.callExternalAPI(url, fmt.Sprintf("block/%s/txs", blockHash))
		if err != nil {
			return nil, fmt.Errorf("failed to get block transactions from API: %w", err)
		}

		var transactions []map[string]interface{}
		if err := json.Unmarshal(response, &transactions); err != nil {
			return nil, fmt.Errorf("failed to decode transactions from API: %w", err)
		}

		// 使用btcd包验证所有交易
		validatedTransactions := make([]map[string]interface{}, len(transactions))
		for i, tx := range transactions {
			validatedTx, err := bs.ParseTransaction(tx)
			if err != nil {
				fmt.Printf("[BTC Scanner] Warning: Invalid transaction at index %d: %v\n", i, err)
				// 继续处理其他交易，不因为单个交易失败而中断
				validatedTransactions[i] = tx
			} else {
				validatedTransactions[i] = validatedTx
			}
		}

		return validatedTransactions, nil
	}
}

// parseLocalBlockData 解析本地节点返回的区块数据（使用btcd包改进）
func (bs *BitcoinScanner) parseLocalBlockData(blockData map[string]interface{}) (*models.Block, error) {
	block := &models.Block{
		Chain: "btc",
	}

	// 设置基本字段
	if hash, ok := blockData["hash"].(string); ok {
		// 使用btcd验证哈希格式
		if _, err := chainhash.NewHashFromStr(hash); err == nil {
			block.Hash = hash
		} else {
			return nil, fmt.Errorf("invalid block hash format: %s", hash)
		}
	}

	if height, ok := blockData["height"].(float64); ok {
		block.Height = uint64(height)
	}
	if version, ok := blockData["version"].(float64); ok {
		block.Version = int(version)
	}
	if timestamp, ok := blockData["time"].(float64); ok {
		block.Timestamp = time.Unix(int64(timestamp), 0)
	}
	if size, ok := blockData["size"].(float64); ok {
		block.Size = uint64(size)
	}
	if weight, ok := blockData["weight"].(float64); ok {
		block.Weight = uint64(weight)
	}
	if txCount, ok := blockData["nTx"].(float64); ok {
		block.TransactionCount = int(txCount)
	}
	if difficulty, ok := blockData["difficulty"].(float64); ok {
		block.Difficulty = difficulty
	}
	if nonce, ok := blockData["nonce"].(float64); ok {
		block.Nonce = uint64(nonce)
	}
	if bits, ok := blockData["bits"].(string); ok {
		block.Bits = bits
	}

	// 设置其他字段
	if prevHash, ok := blockData["previousblockhash"].(string); ok {
		// 验证前一个区块哈希
		if _, err := chainhash.NewHashFromStr(prevHash); err == nil {
			block.PreviousHash = prevHash
		}
	}
	if merkleRoot, ok := blockData["merkleroot"].(string); ok {
		// 验证默克尔根哈希
		if _, err := chainhash.NewHashFromStr(merkleRoot); err == nil {
			block.MerkleRoot = merkleRoot
		}
	}

	return block, nil
}

// parseExternalBlockData 解析外部API返回的区块数据（使用btcd包改进）
func (bs *BitcoinScanner) parseExternalBlockData(blockData map[string]interface{}) (*models.Block, error) {
	block := &models.Block{
		Chain: "btc",
	}

	// 设置基本字段
	if hash, ok := blockData["id"].(string); ok {
		// 使用btcd验证哈希格式
		if _, err := chainhash.NewHashFromStr(hash); err == nil {
			block.Hash = hash
		} else {
			return nil, fmt.Errorf("invalid block hash format: %s", hash)
		}
	}

	if heightVal, ok := blockData["height"].(float64); ok {
		block.Height = uint64(heightVal)
	}
	if version, ok := blockData["version"].(float64); ok {
		block.Version = int(version)
	}
	if timestamp, ok := blockData["timestamp"].(float64); ok {
		block.Timestamp = time.Unix(int64(timestamp), 0)
	}
	if size, ok := blockData["size"].(float64); ok {
		block.Size = uint64(size)
	}
	if weight, ok := blockData["weight"].(float64); ok {
		block.Weight = uint64(weight)
	}
	if txCount, ok := blockData["tx_count"].(float64); ok {
		block.TransactionCount = int(txCount)
	}
	if difficulty, ok := blockData["difficulty"].(float64); ok {
		block.Difficulty = difficulty
	}
	if nonce, ok := blockData["nonce"].(float64); ok {
		block.Nonce = uint64(nonce)
	}
	if bits, ok := blockData["bits"].(string); ok {
		block.Bits = bits
	}
	if fee, ok := blockData["fee"].(float64); ok {
		block.Fee = fee
	}

	// 设置其他字段
	if prevHash, ok := blockData["previousblockhash"].(string); ok {
		// 验证前一个区块哈希
		if _, err := chainhash.NewHashFromStr(prevHash); err == nil {
			block.PreviousHash = prevHash
		}
	}
	if merkleRoot, ok := blockData["merkle_root"].(string); ok {
		// 验证默克尔根哈希
		if _, err := chainhash.NewHashFromStr(merkleRoot); err == nil {
			block.MerkleRoot = merkleRoot
		}
	}

	return block, nil
}

// ValidateBlock 验证区块（使用btcd包改进）
func (bs *BitcoinScanner) ValidateBlock(block *models.Block) error {
	// 基本验证
	if block.Hash == "" {
		return fmt.Errorf("block hash is empty")
	}

	if block.Height == 0 {
		return fmt.Errorf("block height is zero")
	}

	if block.Timestamp.IsZero() {
		return fmt.Errorf("block timestamp is zero")
	}

	// 使用btcd验证哈希格式
	if _, err := chainhash.NewHashFromStr(block.Hash); err != nil {
		return fmt.Errorf("invalid hash format: %s, error: %w", block.Hash, err)
	}

	// 验证前一个区块哈希（如果存在）
	if block.PreviousHash != "" {
		if _, err := chainhash.NewHashFromStr(block.PreviousHash); err != nil {
			return fmt.Errorf("invalid previous hash format: %s, error: %w", block.PreviousHash, err)
		}
	}

	// 验证默克尔根哈希（如果存在）
	if block.MerkleRoot != "" {
		if _, err := chainhash.NewHashFromStr(block.MerkleRoot); err != nil {
			return fmt.Errorf("invalid merkle root format: %s, error: %w", block.MerkleRoot, err)
		}
	}

	return nil
}

// ValidateBitcoinAddress 验证比特币地址（使用btcd包）
func (bs *BitcoinScanner) ValidateBitcoinAddress(address string) error {
	// 尝试解析地址
	addr, err := btcutil.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		// 尝试测试网络
		addr, err = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
		if err != nil {
			// 尝试回归测试网络
			addr, err = btcutil.DecodeAddress(address, &chaincfg.RegressionNetParams)
			if err != nil {
				return fmt.Errorf("invalid bitcoin address: %s, error: %w", address, err)
			}
		}
	}

	// 验证地址是否有效
	if !addr.IsForNet(&chaincfg.MainNetParams) &&
		!addr.IsForNet(&chaincfg.TestNet3Params) &&
		!addr.IsForNet(&chaincfg.RegressionNetParams) {
		return fmt.Errorf("address %s is not valid for any supported network", address)
	}

	return nil
}

// ParseTransaction 解析交易数据（使用btcd包改进）
func (bs *BitcoinScanner) ParseTransaction(txData map[string]interface{}) (map[string]interface{}, error) {
	// 验证交易哈希
	if txHash, ok := txData["txid"].(string); ok {
		if _, err := chainhash.NewHashFromStr(txHash); err != nil {
			return nil, fmt.Errorf("invalid transaction hash: %s, error: %w", txHash, err)
		}
	}

	// 验证输入地址
	if vin, ok := txData["vin"].([]interface{}); ok {
		for i, input := range vin {
			if inputMap, ok := input.(map[string]interface{}); ok {
				if prevTx, ok := inputMap["txid"].(string); ok && prevTx != "" {
					if _, err := chainhash.NewHashFromStr(prevTx); err != nil {
						return nil, fmt.Errorf("invalid input txid at index %d: %s, error: %w", i, prevTx, err)
					}
				}
			}
		}
	}

	// 验证输出地址
	if vout, ok := txData["vout"].([]interface{}); ok {
		for i, output := range vout {
			if outputMap, ok := output.(map[string]interface{}); ok {
				if scriptPubKey, ok := outputMap["scriptPubKey"].(map[string]interface{}); ok {
					if addresses, ok := scriptPubKey["addresses"].([]interface{}); ok {
						for j, addr := range addresses {
							if addrStr, ok := addr.(string); ok {
								if err := bs.ValidateBitcoinAddress(addrStr); err != nil {
									return nil, fmt.Errorf("invalid output address at index %d, address %d: %s, error: %w", i, j, addrStr, err)
								}
							}
						}
					}
				}
			}
		}
	}

	return txData, nil
}

// CalculateBlockStats 计算区块统计信息（使用btcd包改进）
func (bs *BitcoinScanner) CalculateBlockStats(block *models.Block, transactions []map[string]interface{}) {
	// 计算总金额和总费用
	var totalAmount float64
	var totalFee float64
	var inputCount, outputCount int

	for _, tx := range transactions {
		// 计算输入数量
		if vin, ok := tx["vin"].([]interface{}); ok {
			inputCount += len(vin)
		}

		// 计算输出数量和总金额
		if vout, ok := tx["vout"].([]interface{}); ok {
			outputCount += len(vout)
			for _, output := range vout {
				if outputMap, ok := output.(map[string]interface{}); ok {
					if value, ok := outputMap["value"].(float64); ok {
						totalAmount += value
					}
				}
			}
		}

		// 计算交易费用（如果有fee字段）
		if fee, ok := tx["fee"].(float64); ok {
			totalFee += fee
		}
	}

	// 设置区块统计信息
	block.TotalAmount = totalAmount
	block.Fee = totalFee
	block.Confirmations = 1

	// 记录详细的统计信息
	fmt.Printf("[BTC Scanner] Block %d stats: Inputs: %d, Outputs: %d, Total amount: %.8f BTC, Total fee: %.8f BTC\n",
		block.Height, inputCount, outputCount, totalAmount, totalFee)
}

// CalculateTransactionFee 计算交易费用（使用btcd包改进）
func (bs *BitcoinScanner) CalculateTransactionFee(txData map[string]interface{}) (float64, error) {
	var totalInputValue, totalOutputValue float64

	// 计算输入总价值
	if vin, ok := txData["vin"].([]interface{}); ok {
		for _, input := range vin {
			if inputMap, ok := input.(map[string]interface{}); ok {
				if prevOut, ok := inputMap["prevout"].(map[string]interface{}); ok {
					if value, ok := prevOut["value"].(float64); ok {
						totalInputValue += value
					}
				}
			}
		}
	}

	// 计算输出总价值
	if vout, ok := txData["vout"].([]interface{}); ok {
		for _, output := range vout {
			if outputMap, ok := output.(map[string]interface{}); ok {
				if value, ok := outputMap["value"].(float64); ok {
					totalOutputValue += value
				}
			}
		}
	}

	// 计算费用 = 输入 - 输出
	fee := totalInputValue - totalOutputValue
	if fee < 0 {
		return 0, fmt.Errorf("invalid transaction: output value exceeds input value")
	}

	return fee, nil
}

// GetNetworkInfo 获取网络信息（使用btcd包）
func (bs *BitcoinScanner) GetNetworkInfo() (map[string]interface{}, error) {
	// 尝试从本地节点获取网络信息
	if bs.config.RPCURL != "" {
		request := BitcoinRPCRequest{
			JSONRPC: "1.0",
			ID:      "scanner",
			Method:  "getblockchaininfo",
			Params:  []interface{}{},
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal RPC request: %w", err)
		}

		req, err := http.NewRequest("POST", bs.config.RPCURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		if bs.config.Username != "" && bs.config.Password != "" {
			req.SetBasicAuth(bs.config.Username, bs.config.Password)
		}

		resp, err := bs.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get network info: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("local node returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var rpcResp BitcoinRPCResponse
		if err := json.Unmarshal(body, &rpcResp); err != nil {
			return nil, fmt.Errorf("failed to parse RPC response: %w", err)
		}

		if rpcResp.Error != nil {
			return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
		}

		if networkInfo, ok := rpcResp.Result.(map[string]interface{}); ok {
			return networkInfo, nil
		}

		return nil, fmt.Errorf("unexpected result type for network info")
	}

	return nil, fmt.Errorf("no local node configured for network info")
}
