package scanners

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"blockChainBrowser/client/scanner/config"
	"blockChainBrowser/client/scanner/internal/failover"
	"blockChainBrowser/client/scanner/internal/models"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
)

// BitcoinScanner 比特币扫块器 - 使用标准RPC调用
type BitcoinScanner struct {
	config *config.ChainConfig
	// HTTP客户端
	httpClient *http.Client
	// 多节点支持
	currentNodeIndex int // 当前使用的外部节点索引
	// 每块级别的 prevout 缓存：txid -> voutIndex -> scriptPubKey hex
	prevoutCache map[string]map[int]string
	// 全局复用的故障转移管理器
	failoverManager *failover.BTCFailoverManager
}

// newBTCFailover 创建BTC故障转移管理器
func (bs *BitcoinScanner) newBTCFailover() *failover.BTCFailoverManager {
	return failover.NewBTCFailoverManager(bs.config.RPCURL, bs.config.ExplorerAPIURLs)
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
	bs := &BitcoinScanner{
		config:           cfg,
		currentNodeIndex: 0,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		prevoutCache: make(map[string]map[int]string),
	}
	// 初始化全局故障转移管理器（包含本地与外部节点）
	bs.failoverManager = failover.NewBTCFailoverManager(bs.config.RPCURL, bs.config.ExplorerAPIURLs)
	return bs
}

// callRPC 调用RPC接口
func (bs *BitcoinScanner) callRPC(url string, method string, params []interface{}) ([]byte, error) {
	// 统一按bitcoind JSON-RPC调用
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

	resp, err := bs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call RPC: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RPC returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// GetLatestBlockHeight 获取最新区块高度
func (bs *BitcoinScanner) GetLatestBlockHeight() (uint64, error) {
	fm := bs.failoverManager
	result, err := fm.CallWithFailoverUint64("get latest block height", func(baseURL string) (uint64, error) {
		return bs.getBlockHeightFromURL(baseURL)
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
		// 外部节点同样按 JSON-RPC 处理
		body, err := bs.callRPC(url, "getblockcount", []interface{}{})
		if err != nil {
			return 0, fmt.Errorf("failed to call external RPC: %w", err)
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
	}
}

// GetBlockByHeight 根据高度获取区块
func (bs *BitcoinScanner) GetBlockByHeight(height uint64) (*models.Block, error) {
	fmt.Printf("[BTC Scanner] Scanning block at height: %d\n", height)

	fm := bs.failoverManager
	resultInterface, err := fm.Execute("get block by height", func(baseURL string) (interface{}, error) {
		return bs.getBlockByHeightFromURL(baseURL, height)
	})
	if err != nil {
		return nil, err
	}
	if result, ok := resultInterface.(*models.Block); ok {
		fmt.Printf("[BTC Scanner] Successfully scanned block %d (hash: %s) with %d transactions\n",
			result.Height, result.Hash[:16]+"...", result.TransactionCount)
		return result, nil
	}
	return nil, fmt.Errorf("unexpected result type for block")
}

// getBlockByHeightFromURL 从指定URL根据高度获取区块
func (bs *BitcoinScanner) getBlockByHeightFromURL(url string, height uint64) (*models.Block, error) {
	// 统一：先通过高度获取哈希，再获取区块详情
	// getblockhash
	hashRespBody, err := bs.callRPC(url, "getblockhash", []interface{}{height})
	if err != nil {
		return nil, fmt.Errorf("failed to get block hash for height %d: %w", height, err)
	}
	var hashRPC BitcoinRPCResponse
	if err := json.Unmarshal(hashRespBody, &hashRPC); err != nil {
		return nil, fmt.Errorf("failed to parse RPC response: %w", err)
	}
	if hashRPC.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", hashRPC.Error.Message)
	}
	blockHash, ok := hashRPC.Result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type for block hash")
	}

	// getblock with verbosity=2
	blockRespBody, err := bs.callRPC(url, "getblock", []interface{}{blockHash, 2})
	if err != nil {
		return nil, fmt.Errorf("failed to get block %s: %w", blockHash, err)
	}
	var blockRPC BitcoinRPCResponse
	if err := json.Unmarshal(blockRespBody, &blockRPC); err != nil {
		return nil, fmt.Errorf("failed to parse RPC response: %w", err)
	}
	if blockRPC.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", blockRPC.Error.Message)
	}
	blockData, ok := blockRPC.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type for block data")
	}
	return bs.parseLocalBlockData(blockData)
}

// GetBlockByHash 根据哈希获取区块
func (bs *BitcoinScanner) GetBlockByHash(blockHash string) (*models.Block, error) {
	fmt.Printf("[BTC Scanner] Scanning block with hash: %s\n", blockHash)

	fm := bs.failoverManager
	resultInterface, err := fm.Execute("get block by hash", func(baseURL string) (interface{}, error) {
		return bs.getBlockByHashFromURL(baseURL, blockHash)
	})
	if err != nil {
		return nil, err
	}
	if result, ok := resultInterface.(*models.Block); ok {
		fmt.Printf("[BTC Scanner] Successfully scanned block %d (hash: %s) with %d transactions\n",
			result.Height, result.Hash[:16]+"...", result.TransactionCount)
		return result, nil
	}
	return nil, fmt.Errorf("unexpected result type for block")
}

// getBlockByHashFromURL 从指定URL根据哈希获取区块
func (bs *BitcoinScanner) getBlockByHashFromURL(url string, blockHash string) (*models.Block, error) {
	response, err := bs.callRPC(url, "getblock", []interface{}{blockHash, 2}) // 2表示详细模式
	if err != nil {
		return nil, fmt.Errorf("failed to get block %s: %w", blockHash, err)
	}

	// 统一解析 JSON-RPC envelope
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
}

// GetBlockTransactions 获取区块交易
func (bs *BitcoinScanner) GetBlockTransactions(blockHash string) ([]map[string]interface{}, error) {
	fmt.Printf("[BTC Scanner] Getting transactions for block: %s\n", blockHash)

	fm := bs.failoverManager
	result, err := fm.CallWithFailoverMaps("get block transactions", func(baseURL string) ([]map[string]interface{}, error) {
		return bs.getBlockTransactionsFromURL(baseURL, blockHash)
	})

	if err == nil {
		fmt.Printf("[BTC Scanner] Retrieved %d transactions from block %s\n", len(result), blockHash[:16]+"...")
	}
	return result, err
}

// GetBlockTransactionsFromBlock 直接从区块中获取交易信息（避免哈希不一致问题）
func (bs *BitcoinScanner) GetBlockTransactionsFromBlock(block *models.Block) ([]map[string]interface{}, error) {
	fm := bs.failoverManager
	result, err := fm.CallWithFailoverMaps("get block transactions by height", func(baseURL string) ([]map[string]interface{}, error) {
		return bs.getBlockTransactionsFromURL(baseURL, block.Hash)
	})

	return result, err
}

// getBlockTransactionsFromURL 从指定URL获取区块交易（使用btcd包改进）
func (bs *BitcoinScanner) getBlockTransactionsFromURL(url string, blockHash string) ([]map[string]interface{}, error) {
	// 优先尝试 verbosity=3 获取完整 prevout 信息（支持 Taproot）
	response, err := bs.callRPC(url, "getblock", []interface{}{blockHash, 3})
	if err != nil {
		fmt.Printf("[BTC Scanner] getblock verbosity=3 失败，回退到 verbosity=2: %v\n", err)
		// 回退到 verbosity=2
		response, err = bs.callRPC(url, "getblock", []interface{}{blockHash, 2})
		if err != nil {
			return nil, fmt.Errorf("failed to get block transactions: %w", err)
		}
	}

	var rpcResp BitcoinRPCResponse
	if err := json.Unmarshal(response, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse RPC response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	if blockData, ok := rpcResp.Result.(map[string]interface{}); ok {
		if txs, ok := blockData["tx"].([]interface{}); ok {
			transactions := make([]map[string]interface{}, len(txs))
			for i, tx := range txs {
				if txMap, ok := tx.(map[string]interface{}); ok {
					validatedTx, err := bs.ParseTransaction(txMap)
					if err != nil {
						fmt.Printf("[BTC Scanner] Warning: Invalid transaction at index %d: %v\n", i, err)
						transactions[i] = txMap
					} else {
						transactions[i] = validatedTx
					}
				}
			}
			return transactions, nil
		}
	}

	return nil, fmt.Errorf("no transactions found in block")
}

// parseLocalBlockData 解析节点返回的区块数据（使用btcd包改进）
func (bs *BitcoinScanner) parseLocalBlockData(blockData map[string]interface{}) (*models.Block, error) {
	block := &models.Block{
		Chain:   "btc",
		ChainID: bs.config.ChainID,
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

	// 基本完整性校验（避免上报被后端驳回）
	if block.Hash == "" || block.MerkleRoot == "" {
		return nil, fmt.Errorf("incomplete block: missing hash or merkle root")
	}
	if block.Height > 0 && block.PreviousHash == "" {
		return nil, fmt.Errorf("incomplete block: missing previous hash for non-genesis block")
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

// extractAddressesFromScriptPubKey 使用 btcd 专业工具链解析 scriptPubKey 地址
func (bs *BitcoinScanner) extractAddressesFromScriptPubKey(scriptPubKeyHex string) ([]string, error) {
	if scriptPubKeyHex == "" {
		return nil, nil
	}
	// 解码十六进制脚本
	script, err := hex.DecodeString(scriptPubKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode scriptPubKey hex: %w", err)
	}

	// 使用 txscript.ExtractPkScriptAddrs 解析地址
	// 尝试主网参数
	_, addrs, _, err := txscript.ExtractPkScriptAddrs(script, &chaincfg.MainNetParams)
	if err != nil {
		// 尝试测试网参数
		_, addrs, _, err = txscript.ExtractPkScriptAddrs(script, &chaincfg.TestNet3Params)
		if err != nil {
			// 尝试回归测试网参数
			_, addrs, _, err = txscript.ExtractPkScriptAddrs(script, &chaincfg.RegressionNetParams)
			if err != nil {
				return nil, fmt.Errorf("failed to extract addresses from scriptPubKey: %w", err)
			}
		}
	}

	// 转换为字符串地址
	addresses := make([]string, len(addrs))
	for i, addr := range addrs {
		addresses[i] = addr.EncodeAddress()
	}

	return addresses, nil
}

// extractAddressesFromRPCData 从 RPC 返回的数据中快速提取地址（性能优化）
func (bs *BitcoinScanner) extractAddressesFromRPCData(spk map[string]interface{}) []string {
	var addresses []string

	// 优先使用 RPC 已经解析好的地址（最快）
	if addrStr, ok := spk["address"].(string); ok && addrStr != "" {
		addresses = append(addresses, addrStr)
	}

	if addressesList, ok := spk["addresses"].([]interface{}); ok && len(addressesList) > 0 {
		for _, a := range addressesList {
			if s, ok := a.(string); ok && s != "" {
				addresses = append(addresses, s)
			}
		}
	}

	return addresses
}

// extractAddressesFromWitness 从 witness 数据解析地址
func (bs *BitcoinScanner) extractAddressesFromWitness(txinwitness []interface{}) ([]string, error) {
	if len(txinwitness) == 0 {
		return nil, nil
	}

	var addresses []string

	// 尝试解析 P2WPKH (Pay-to-Witness-Public-Key-Hash)
	if len(txinwitness) == 2 {
		// 第一个元素是签名，第二个是公钥
		if pubkeyHex, ok := txinwitness[1].(string); ok && len(pubkeyHex) == 66 { // 33字节压缩公钥
			// 解析公钥
			pubkeyBytes, err := hex.DecodeString(pubkeyHex)
			if err == nil {
				pubkey, err := btcec.ParsePubKey(pubkeyBytes)
				if err == nil {
					// 计算公钥哈希
					hash160 := btcutil.Hash160(pubkey.SerializeCompressed())

					// 尝试不同网络参数
					networks := []*chaincfg.Params{&chaincfg.MainNetParams, &chaincfg.TestNet3Params, &chaincfg.RegressionNetParams}
					for _, params := range networks {
						addr, err := btcutil.NewAddressWitnessPubKeyHash(hash160, params)
						if err == nil {
							addresses = append(addresses, addr.EncodeAddress())
							break
						}
					}
				}
			}
		}
	}

	// 尝试解析 P2WSH (Pay-to-Witness-Script-Hash)
	if len(txinwitness) > 2 {
		// 最后一个元素是 redeem script
		if redeemScriptHex, ok := txinwitness[len(txinwitness)-1].(string); ok {
			redeemScript, err := hex.DecodeString(redeemScriptHex)
			if err == nil {
				// 计算脚本哈希
				sha := sha256.Sum256(redeemScript)
				scriptHash := sha[:]

				// 尝试不同网络参数
				networks := []*chaincfg.Params{&chaincfg.MainNetParams, &chaincfg.TestNet3Params, &chaincfg.RegressionNetParams}
				for _, params := range networks {
					addr, err := btcutil.NewAddressWitnessScriptHash(scriptHash, params)
					if err == nil {
						addresses = append(addresses, addr.EncodeAddress())
						break
					}
				}
			}
		}
	}

	return addresses, nil
}

// ParseTransaction 解析交易数据（使用btcd包专业工具链）
func (bs *BitcoinScanner) ParseTransaction(txData map[string]interface{}) (map[string]interface{}, error) {

	// 验证交易哈希
	if txHash, ok := txData["txid"].(string); ok {
		if _, err := chainhash.NewHashFromStr(txHash); err != nil {
			return nil, fmt.Errorf("invalid transaction hash: %s, error: %w", txHash, err)
		}
		// 统一字段名：为上层上传逻辑提供 hash 键
		if _, exists := txData["hash"]; !exists {
			txData["hash"] = txHash
		}
	}

	// 使用优化的方法解析输出地址（vout）
	firstOutputAddr := ""
	outputAddrs := make([]string, 0)
	if vout, ok := txData["vout"].([]interface{}); ok {
		for _, output := range vout {
			if outputMap, ok := output.(map[string]interface{}); ok {
				if spk, ok := outputMap["scriptPubKey"].(map[string]interface{}); ok {
					// 1. 优先使用 RPC 已经解析好的地址（最快）
					addrs := bs.extractAddressesFromRPCData(spk)

					// 2. 如果 RPC 没有提供地址，尝试从 hex 解析
					if len(addrs) == 0 {
						if scriptHex, ok := spk["hex"].(string); ok && scriptHex != "" {
							var err error
							addrs, err = bs.extractAddressesFromScriptPubKey(scriptHex)
							if err != nil {
								fmt.Printf("[BTC Scanner] Warning: Failed to parse output scriptPubKey: %v\n", err)
							}
						}
					}

					// 添加到输出地址列表
					for _, addr := range addrs {
						if firstOutputAddr == "" {
							firstOutputAddr = addr
						}
						outputAddrs = append(outputAddrs, addr)
					}
				}
			}
		}
	}

	// 使用 btcd 专业工具链解析输入地址（vin）
	firstInputAddr := ""
	allInputAddrs := make([]string, 0)
	if vin, ok := txData["vin"].([]interface{}); ok {
		for i, input := range vin {
			if inputMap, ok := input.(map[string]interface{}); ok {
				// 验证 txid
				if prevTx, ok := inputMap["txid"].(string); ok && prevTx != "" {
					if _, err := chainhash.NewHashFromStr(prevTx); err != nil {
						return nil, fmt.Errorf("invalid input txid at index %d: %s, error: %w", i, prevTx, err)
					}
				}

				var currentInputAddrs []string
				var err error

				// 1. 优先使用 RPC 返回的现成地址信息（最快）
				if prevout, ok := inputMap["prevout"].(map[string]interface{}); ok {
					if spk, ok := prevout["scriptPubKey"].(map[string]interface{}); ok {
						// 首先尝试使用 RPC 已经解析好的地址
						currentInputAddrs = bs.extractAddressesFromRPCData(spk)

						// 如果 RPC 没有提供地址，尝试从 hex 解析
						if len(currentInputAddrs) == 0 {
							if scriptHex, ok := spk["hex"].(string); ok && scriptHex != "" {
								currentInputAddrs, err = bs.extractAddressesFromScriptPubKey(scriptHex)
								if err != nil {
									fmt.Printf("[BTC Scanner] Warning: Failed to parse prevout scriptPubKey for input %d: %v\n", i, err)
								}
							}
						}
					}
				}

				// 2. 如果没有 prevout 或解析失败，尝试 witness 解析
				if len(currentInputAddrs) == 0 {
					if txinwitness, ok := inputMap["txinwitness"].([]interface{}); ok {
						currentInputAddrs, err = bs.extractAddressesFromWitness(txinwitness)
						if err != nil {
							fmt.Printf("[BTC Scanner] Warning: Failed to parse witness for input %d: %v\n", i, err)
						}
					}
				}

				// 3. 不再调用 getrawtransaction：若 prevout 与 witness 均不可用则跳过，避免额外 RPC

				// 添加到总列表
				for _, addr := range currentInputAddrs {
					if firstInputAddr == "" {
						firstInputAddr = addr
					}
					allInputAddrs = append(allInputAddrs, addr)
				}
			}
		}
	}

	// 设置 from/to 字段
	if _, ok := txData["from"]; !ok && firstInputAddr != "" {
		txData["from"] = firstInputAddr
	}
	if _, ok := txData["to"]; !ok && firstOutputAddr != "" {
		txData["to"] = firstOutputAddr
	}

	// 附加所有输入/输出地址的JSON（给后端 transaction.address_froms/address_tos）
	if len(allInputAddrs) > 0 {
		if b, err := json.Marshal(allInputAddrs); err == nil {
			txData["address_froms"] = string(b)
		}
	}
	if len(outputAddrs) > 0 {
		if b, err := json.Marshal(outputAddrs); err == nil {
			txData["address_tos"] = string(b)
		}
	}

	// 计算 BTC 交易费用（在转换JSON之前）
	fee, totalInputValue, _, err := bs.CalculateTransactionFee(txData)
	if err != nil {
		fmt.Printf("[BTC Scanner] Warning: Failed to calculate transaction fee: %v\n", err)
		fee = 0
		totalInputValue = 0
	}
	txData["fee"] = fee

	txData["value"] = totalInputValue * 1e8

	// 存储 BTC 原始交易数据 vin 和 vout
	if vin, ok := txData["vin"].([]interface{}); ok {
		if vinJSON, err := json.Marshal(vin); err == nil {
			txData["vin"] = string(vinJSON)
		}
	}
	if vout, ok := txData["vout"].([]interface{}); ok {
		if voutJSON, err := json.Marshal(vout); err == nil {
			txData["vout"] = string(voutJSON)
		}
	}

	// BTC 无 gas 概念：为上层统一接口默认补全 gas 字段（字符串零）
	if _, ok := txData["gasLimit"]; !ok {
		txData["gasLimit"] = 0
	}
	if _, ok := txData["gasPrice"]; !ok {
		txData["gasPrice"] = "0"
	}
	if _, ok := txData["gasUsed"]; !ok {
		txData["gasUsed"] = 0
	}

	return txData, nil
}

// ensureJSONArrayField ensures tx[key] is a JSON array ([]interface{}).
// If it's a JSON-encoded string, it will be unmarshaled back into an array.
func (bs *BitcoinScanner) ensureJSONArrayField(tx map[string]interface{}, key string) []interface{} {
	if tx == nil {
		return nil
	}
	if arr, ok := tx[key].([]interface{}); ok {
		return arr
	}
	if s, ok := tx[key].(string); ok && s != "" {
		var out []interface{}
		if err := json.Unmarshal([]byte(s), &out); err == nil {
			tx[key] = out
			return out
		}
	}
	return nil
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

		// 累加交易费用（已在 ParseTransaction 中计算）
		if fee, ok := tx["fee"].(float64); ok {
			totalFee += fee
		}
	}

	// 解析第一笔交易，它是矿工的信息！
	if len(transactions) > 0 {
		coinbaseTx := transactions[0]
		// 恢复 vin/vout 为数组
		vinArr := bs.ensureJSONArrayField(coinbaseTx, "vin")
		voutArr := bs.ensureJSONArrayField(coinbaseTx, "vout")

		// 检查是否为 coinbase 交易（依据 vin[0].coinbase 字段）
		if len(vinArr) > 0 {
			if firstInput, ok := vinArr[0].(map[string]interface{}); ok {
				if _, isCoinbase := firstInput["coinbase"]; isCoinbase {
					fmt.Printf("[BTC Scanner] 确认是 coinbase 交易\n")

					// 解析矿工地址（第一个输出地址）
					if len(voutArr) > 0 {
						if firstOutput, ok := voutArr[0].(map[string]interface{}); ok {
							if spk, ok := firstOutput["scriptPubKey"].(map[string]interface{}); ok {
								if minerAddr, ok := spk["address"].(string); ok && minerAddr != "" {
									fmt.Printf("[BTC Scanner] 矿工地址: %s\n", minerAddr)
									block.Miner = minerAddr
									if reward, ok := firstOutput["value"].(float64); ok {
										block.BaseFee = big.NewInt(int64(reward * 1e8))
									}
								}
							}
						}
					}
				}
			}
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
func (bs *BitcoinScanner) CalculateTransactionFee(txData map[string]interface{}) (float64, float64, float64, error) {
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
		return 0, 0, 0, fmt.Errorf("invalid transaction: output value exceeds input value")
	}

	return fee, totalInputValue, totalOutputValue, nil
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
