package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
)

type ContractParseService interface {
	ParseAndStore(ctx context.Context, receipt *models.TransactionReceipt, tx *models.Transaction, parserConfigs []*models.ParserConfig) ([]*models.ContractParseResult, error)
	ParseAndStoreBatchAsync(ctx context.Context, receipts []*models.TransactionReceipt, txs map[string]*models.Transaction, parserConfigs map[string][]*models.ParserConfig)
	ParseAndStoreBatchByTxHashesAsync(ctx context.Context, txHashes []string)
	GetByTxHash(ctx context.Context, hash string) ([]*models.ContractParseResult, error)
}

type contractParseService struct {
	repo             repository.ContractParseResultRepository
	receiptRepo      repository.TransactionReceiptRepository
	txRepo           repository.TransactionRepository
	parserConfigRepo repository.ParserConfigRepository
	coinConfigRepo   repository.CoinConfigRepository
}

func NewContractParseService(
	repo repository.ContractParseResultRepository,
	receiptRepo repository.TransactionReceiptRepository,
	txRepo repository.TransactionRepository,
	parserConfigRepo repository.ParserConfigRepository,
	coinConfigRepo repository.CoinConfigRepository,
) ContractParseService {
	return &contractParseService{
		repo:             repo,
		receiptRepo:      receiptRepo,
		txRepo:           txRepo,
		parserConfigRepo: parserConfigRepo,
		coinConfigRepo:   coinConfigRepo,
	}
}

func (s *contractParseService) ParseAndStore(ctx context.Context, receipt *models.TransactionReceipt, tx *models.Transaction, parserConfigs []*models.ParserConfig) ([]*models.ContractParseResult, error) {
	coinConfigs, err := s.coinConfigRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	results := s.parseReceipt(receipt, tx, parserConfigs, coinConfigs)
	for _, r := range results {
		_ = s.repo.Create(ctx, r)
	}
	return results, nil
}

func (s *contractParseService) ParseAndStoreBatchAsync(ctx context.Context, receipts []*models.TransactionReceipt, txs map[string]*models.Transaction, parserConfigs map[string][]*models.ParserConfig) {
	go func() {
		// do NOT use request-scoped ctx in background goroutine
		bgCtx := context.Background()
		// 调试日志：入参统计
		totalReceipts := len(receipts)
		totalTxs := len(txs)
		pcKeys := 0
		totalPC := 0
		for _, v := range parserConfigs {
			pcKeys++
			totalPC += len(v)
		}
		fmt.Printf("[ParseAndStoreBatchAsync] start: receipts=%d txs=%d parserConfigKeys=%d totalParserConfigs=%d\n", totalReceipts, totalTxs, pcKeys, totalPC)

		for i, rc := range receipts {
			if rc == nil || rc.TxHash == "" {
				fmt.Printf("[ParseAndStoreBatchAsync] #%d skip: nil receipt or empty txHash\n", i)
				continue
			}
			pcs := parserConfigs[rc.TxHash]
			t := txs[rc.TxHash]
			logsLen := len(rc.LogsData)
			fmt.Printf("[ParseAndStoreBatchAsync] #%d txHash=%s logsLen=%d pcs=%d txPresent=%v\n", i, rc.TxHash, logsLen, len(pcs), t != nil)

			coinConfigs, err := s.coinConfigRepo.GetAll(bgCtx)
			if err != nil {
				fmt.Printf("[ParseAndStoreBatchAsync] GetAll coinConfig error: %v\n", err)
				continue
			}
			results := s.parseReceipt(rc, t, pcs, coinConfigs)
			fmt.Printf("[ParseAndStoreBatchAsync] parsed results=%d for txHash=%s\n", len(results), rc.TxHash)
			if len(results) > 0 {
				if err := s.repo.CreateBatch(bgCtx, results); err != nil {
					fmt.Printf("[ParseAndStoreBatchAsync] CreateBatch error for txHash=%s: %v\n", rc.TxHash, err)
				}
			} else {
				fmt.Printf("[ParseAndStoreBatchAsync] no results to store for txHash=%s\n", rc.TxHash)
			}
		}
		fmt.Printf("[ParseAndStoreBatchAsync] done\n")
	}()
}

// New entry point: only provide tx hashes; service will fetch receipts/txs/configs and filter
func (s *contractParseService) ParseAndStoreBatchByTxHashesAsync(ctx context.Context, txHashes []string) {
	go func() {
		// do NOT use request-scoped ctx in background goroutine
		bgCtx := context.Background()
		if len(txHashes) == 0 {
			fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] empty input\n")
			return
		}
		// preload all parser configs and group by contract address
		allConfigs, err := s.parserConfigRepo.GetAllParserConfigs(bgCtx)
		if err != nil {
			fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] GetAllParserConfigs error: %v\n", err)
			return
		}
		grouped := make(map[string][]*models.ParserConfig)
		for _, pc := range allConfigs {
			grouped[pc.ContractAddress] = append(grouped[pc.ContractAddress], pc)
		}
		fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] start hashes=%d groupedKeys=%d\n", len(txHashes), len(grouped))

		for _, h := range txHashes {
			if h == "" {
				continue
			}

			rc, err := s.receiptRepo.GetByTxHash(bgCtx, h)
			if err != nil || rc == nil {
				fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] skip tx=%s no receipt: %v\n", h, err)
				continue
			}
			tx, err := s.txRepo.GetByHash(bgCtx, h)
			if err != nil || tx == nil {
				fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] skip tx=%s no tx record: %v\n", h, err)
				continue
			}
			// Only parse when there is explicit contract config for AddressTo
			pcs := grouped[tx.AddressTo]
			if len(pcs) == 0 {
				fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] skip tx=%s no parser config for address=%s\n", h, tx.AddressTo)
				// If no parser config for the specific address, try to use the default config (key "*")
				defaultPcs := grouped["*"]
				if len(defaultPcs) == 0 {
					fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] skip tx=%s no parser config for address=%s and no default config\n", h, tx.AddressTo)
					continue
				}
				pcs = defaultPcs
			}
			coinConfigs, err := s.coinConfigRepo.GetAll(bgCtx)
			if err != nil {
				fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] GetAll coinConfig error: %v\n", err)
				continue
			}
			results := s.parseReceipt(rc, tx, pcs, coinConfigs)
			if len(results) > 0 {
				if err := s.repo.CreateBatch(bgCtx, results); err != nil {
					fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] CreateBatch error for tx=%s: %v\n", h, err)
				}
			} else {
				fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] no results for tx=%s\n", h)
			}
		}
		fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] done\n")
	}()
}

func (s *contractParseService) GetByTxHash(ctx context.Context, hash string) ([]*models.ContractParseResult, error) {
	return s.repo.GetByTxHash(ctx, hash)
}

// parseReceipt 仅基于 logs_parser_rules 进行快速解析
func (s *contractParseService) parseReceipt(receipt *models.TransactionReceipt, tx *models.Transaction, parserConfigs []*models.ParserConfig, coinConfigs []*models.CoinConfig) []*models.ContractParseResult {
	if receipt == nil || receipt.LogsData == "" {
		return nil
	}
	var logs []map[string]interface{}
	if err := json.Unmarshal([]byte(receipt.LogsData), &logs); err != nil {
		fmt.Printf("[parseReceipt] unmarshal logs failed txHash=%s error=%v\n", emptyIfNil(receipt.TxHash), err)
		return nil
	}
	fmt.Printf("[parseReceipt] begin txHash=%s logs=%d parserConfigs=%d\n", emptyIfNil(receipt.TxHash), len(logs), len(parserConfigs))

	// 预计算原始日志hash（用于幂等）
	rawHash := sha256.Sum256([]byte(receipt.LogsData))
	rawHashHex := "0x" + hex.EncodeToString(rawHash[:])

	// 交易上下文
	chain := receipt.Chain
	contractAddr := emptyIfNil(receipt.ContractAddress)

	// 建立 event_signature -> rules 的快速映射
	sigToRule := map[string]*models.ParserConfig{}
	for _, c := range parserConfigs {
		if c != nil && c.EventSignature != "" && (c.LogsParserRules.ExtractAmount != "" || c.LogsParserRules.ExtractFromAddress != "" || c.LogsParserRules.ExtractToAddress != "") {
			sig := strings.ToLower(c.EventSignature)
			sigToRule[sig] = c
		}
	}
	fmt.Printf("[parseReceipt] built sigToRule size=%d\n", len(sigToRule))

	var out []*models.ContractParseResult

	for idx, log := range logs {
		topics, _ := log["topics"].([]interface{})
		data, _ := log["data"].(string)
		if len(topics) == 0 {
			fmt.Printf("[parseReceipt] #%d skip: no topics\n", idx)
			continue
		}
		topic0, _ := topics[0].(string)
		sig := strings.ToLower(topic0)
		rule := sigToRule[sig]
		if rule == nil {
			fmt.Printf("[parseReceipt] #%d no rule match for sig=%s\n", idx, sig)
			continue
		}

		// extract fields by rules
		fromAddr := extractAddressByRule(rule.LogsParserRules.ExtractFromAddress, topics, data)
		toAddr := extractAddressByRule(rule.LogsParserRules.ExtractToAddress, topics, data)
		amountWei := extractAmountByRule(rule.LogsParserRules.ExtractAmount, data)
		fmt.Printf("[parseReceipt] #%d matched sig=%s from=%s to=%s amountWei=%s\n", idx, sig, fromAddr, toAddr, amountWei)

		// decimals 优先用交易里 token_decimals（若有），否则 18/0 由前端展示层处理
		var decimals uint16
		if tx != nil && tx.TokenDecimals != 0 {
			decimals = uint16(tx.TokenDecimals)
		}

		// 封装结果
		bj, _ := json.Marshal(log)
		res := &models.ContractParseResult{}
		res.TxHash = receipt.TxHash
		res.ContractAddress = contractAddr
		res.Chain = chain
		res.BlockNumber = receipt.BlockNumber
		res.LogIndex = uint(idx)
		res.EventSignature = sig
		res.EventName = rule.EventName
		res.FromAddress = fromAddr
		res.ToAddress = toAddr
		res.AmountWei = amountWei
		res.TokenDecimals = decimals
		res.TokenSymbol = txSymbol(coinConfigs, tx)
		res.RawLogsHash = rawHashHex
		res.ParsedJSON = string(bj)
		out = append(out, res)
		fmt.Printf("[parseReceipt] #%d produce result txHash=%s logIndex=%d symbol=%s decimals=%d\n", idx, res.TxHash, res.LogIndex, res.TokenSymbol, res.TokenDecimals)
	}
	fmt.Printf("[parseReceipt] end txHash=%s produced=%d\n", emptyIfNil(receipt.TxHash), len(out))
	return out
}

func extractAddressByRule(rule string, topics []interface{}, data string) string {
	if rule == "" || len(topics) == 0 {
		return ""
	}
	low := strings.ToLower(rule)
	if strings.HasPrefix(low, "topics[") {
		// topics[i]
		i := indexFromBracket(low)
		if i > 0 && i < len(topics) {
			if s, ok := topics[i].(string); ok && len(s) >= 66 {
				return "0x" + strings.ToLower(s[len(s)-40:])
			}
		}
	} else if strings.HasPrefix(low, "data[") {
		// data[i] -> 32 bytes slot
		i := indexFromBracket(low)
		slot := getDataSlot(data, i)
		if slot != "" && len(slot) == 64 {
			return "0x" + strings.ToLower(slot[24:])
		}
	}
	return ""
}

func extractAmountByRule(rule string, data string) string {
	if rule == "" || data == "" {
		return "0"
	}
	low := strings.ToLower(rule)
	if strings.HasPrefix(low, "data[") {
		i := indexFromBracket(low)
		slot := getDataSlot(data, i)
		if slot != "" {
			return hexToDecimalString(slot)
		}
	}
	return "0"
}

// hexToDecimalString converts a hex string (with or without 0x) to a base-10 string.
func hexToDecimalString(hexStr string) string {
	if hexStr == "" {
		return "0"
	}
	s := strings.TrimPrefix(strings.ToLower(hexStr), "0x")
	for len(s) > 0 && s[0] == '0' {
		s = s[1:]
	}
	if s == "" {
		return "0"
	}
	n := new(big.Int)
	if _, ok := n.SetString(s, 16); !ok {
		return "0"
	}
	return n.String()
}

func indexFromBracket(expr string) int {
	// expr like topics[1] or data[0]
	l := strings.Index(expr, "[")
	r := strings.Index(expr, "]")
	if l >= 0 && r > l+1 {
		var n int
		_, _ = fmtSscanf(expr[l+1:r], "%d", &n)
		return n
	}
	return -1
}

// lightweight sscanf for ints
func fmtSscanf(s, f string, p *int) (int, error) {
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	*p = n
	return 1, nil
}

func getDataSlot(data string, idx int) string {
	if data == "" || idx < 0 {
		return ""
	}
	d := strings.TrimPrefix(strings.ToLower(data), "0x")
	start := idx * 64
	end := start + 64
	if start < 0 || end > len(d) {
		return ""
	}
	return d[start:end]
}

func removeHexPrefix(s string) string {
	return strings.TrimPrefix(strings.ToLower(s), "0x")
}

func emptyIfNil(s string) string {
	return s
}
func txSymbol(coinConfigs []*models.CoinConfig, tx *models.Transaction) string {
	for _, c := range coinConfigs {
		if c.ContractAddr == tx.AddressTo {
			return c.Symbol
		}
	}
	return ""
}
