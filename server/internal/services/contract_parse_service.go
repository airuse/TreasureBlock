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

// TxHashWithIndex represents a transaction hash with its sequential index
type TxHashWithIndex struct {
	Hash  string
	Index int
}

type ContractParseService interface {
	ParseAndStore(ctx context.Context, receipt *models.TransactionReceipt, tx *models.Transaction, parserConfigs []*models.ParserConfig) ([]*models.ContractParseResult, error)
	ParseAndStoreBatchAsync(ctx context.Context, receipts []*models.TransactionReceipt, txs map[string]*models.Transaction, parserConfigs map[string][]*models.ParserConfig)
	ParseAndStoreBatchByTxHashesAsync(ctx context.Context, txHashesWithIndex []TxHashWithIndex)
	GetByTxHash(ctx context.Context, hash string) ([]*models.ContractParseResult, error)
}

type contractParseService struct {
	repo             repository.ContractParseResultRepository
	receiptRepo      repository.TransactionReceiptRepository
	txRepo           repository.TransactionRepository
	parserConfigRepo repository.ParserConfigRepository
	coinConfigRepo   repository.CoinConfigRepository
	userAddressRepo  repository.UserAddressRepository
}

func NewContractParseService(
	repo repository.ContractParseResultRepository,
	receiptRepo repository.TransactionReceiptRepository,
	txRepo repository.TransactionRepository,
	parserConfigRepo repository.ParserConfigRepository,
	coinConfigRepo repository.CoinConfigRepository,
	userAddressRepo repository.UserAddressRepository,
) ContractParseService {
	return &contractParseService{
		repo:             repo,
		receiptRepo:      receiptRepo,
		txRepo:           txRepo,
		parserConfigRepo: parserConfigRepo,
		coinConfigRepo:   coinConfigRepo,
		userAddressRepo:  userAddressRepo,
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
		bgCtx := context.Background()
		// totalReceipts := len(receipts)
		// totalTxs := len(txs)
		pcKeys := 0
		totalPC := 0
		for _, v := range parserConfigs {
			pcKeys++
			totalPC += len(v)
		}
		// fmt.Printf("[ParseAndStoreBatchAsync] start: receipts=%d txs=%d parserConfigKeys=%d totalParserConfigs=%d\n", totalReceipts, totalTxs, pcKeys, totalPC)

		for i := 0; i < len(receipts); i++ {
			rc := receipts[i]
			if rc == nil || rc.TxHash == "" {
				// fmt.Printf("[ParseAndStoreBatchAsync] #%d skip: nil receipt or empty txHash\n", i)
				continue
			}
			pcs := parserConfigs[rc.TxHash]
			t := txs[rc.TxHash]
			// logsLen := len(rc.LogsData)
			// fmt.Printf("[ParseAndStoreBatchAsync] #%d txHash=%s logsLen=%d pcs=%d txPresent=%v\n", i, rc.TxHash, logsLen, len(pcs), t != nil)

			coinConfigs, err := s.coinConfigRepo.GetAll(bgCtx)
			if err != nil {
				// fmt.Printf("[ParseAndStoreBatchAsync] GetAll coinConfig error: %v\n", err)
				continue
			}
			results := s.parseReceipt(rc, t, pcs, coinConfigs)
			// fmt.Printf("[ParseAndStoreBatchAsync] parsed results=%d for txHash=%s\n", len(results), rc.TxHash)
			if len(results) > 0 {
				if err := s.repo.CreateBatch(bgCtx, results); err != nil {
					// fmt.Printf("[ParseAndStoreBatchAsync] CreateBatch error for txHash=%s: %v\n", rc.TxHash, err)
				}
			} else {
				// fmt.Printf("[ParseAndStoreBatchAsync] no results to store for txHash=%s\n", rc.TxHash)
			}
		}
		// fmt.Printf("[ParseAndStoreBatchAsync] done\n")
	}()
}

// New entry point: only provide tx hashes; service will fetch receipts/txs/configs and filter
func (s *contractParseService) ParseAndStoreBatchByTxHashesAsync(ctx context.Context, txHashesWithIndex []TxHashWithIndex) {
	go func() {
		bgCtx := context.Background()
		if len(txHashesWithIndex) == 0 {
			return
		}

		// fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] start: total=%d transactions\n", len(txHashesWithIndex))

		// Sort by index to ensure strict ordering regardless of input order
		sortedTxs := make([]TxHashWithIndex, len(txHashesWithIndex))
		copy(sortedTxs, txHashesWithIndex)

		// Simple bubble sort by Index (for small arrays this is fine)
		for i := 0; i < len(sortedTxs)-1; i++ {
			for j := 0; j < len(sortedTxs)-i-1; j++ {
				if sortedTxs[j].Index > sortedTxs[j+1].Index {
					sortedTxs[j], sortedTxs[j+1] = sortedTxs[j+1], sortedTxs[j]
				}
			}
		}

		// fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] sorted transactions by index\n")

		// Preload all parser configs and group by contract address
		allConfigs, err := s.parserConfigRepo.GetAllParserConfigs(bgCtx)
		if err != nil {
			// fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] failed to load parser configs: %v\n", err)
			return
		}
		grouped := make(map[string][]*models.ParserConfig)
		for _, pc := range allConfigs {
			grouped[pc.ContractAddress] = append(grouped[pc.ContractAddress], pc)
		}

		// Process transactions in strict index order
		for _, txItem := range sortedTxs {
			if txItem.Hash == "" {
				// fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] skip empty hash at index %d\n", txItem.Index)
				continue
			}

			// fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] processing index=%d hash=%s\n", txItem.Index, txItem.Hash)

			rc, err := s.receiptRepo.GetByTxHash(bgCtx, txItem.Hash)
			if err != nil || rc == nil {
				// fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] receipt not found for index=%d hash=%s\n", txItem.Index, txItem.Hash)
				continue
			}
			tx, err := s.txRepo.GetByHash(bgCtx, txItem.Hash)
			if err != nil || tx == nil {
				// fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] transaction not found for index=%d hash=%s\n", txItem.Index, txItem.Hash)
				continue
			}

			// Only parse when there is explicit contract config for AddressTo
			pcs := grouped[tx.AddressTo]
			if len(pcs) == 0 {
				// If no parser config for the specific address, try to use the default config (key "*")
				defaultPcs := grouped["*"]
				if len(defaultPcs) == 0 {
					// fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] no parser config for index=%d hash=%s addressTo=%s\n", txItem.Index, txItem.Hash, tx.AddressTo)
					continue
				}
				pcs = defaultPcs
			}

			coinConfigs, err := s.coinConfigRepo.GetAll(bgCtx)
			if err != nil {
				// fmt.Printf("[ParseAndStoreBatchByTxHashesAsync] failed to load coin configs for index=%d: %v\n", txItem.Index, err)
				continue
			}

			_ = s.parseReceipt(rc, tx, pcs, coinConfigs)
		}
	}()
}

func (s *contractParseService) GetByTxHash(ctx context.Context, hash string) ([]*models.ContractParseResult, error) {
	return s.repo.GetByTxHash(ctx, hash)
}

// EventUpdate represents a single balance update event with strict ordering
type EventUpdate struct {
	Height      uint64
	TxIndex     uint64
	EventType   string // "transfer", "transferFrom", "approve", "balanceOf"
	UserAddress *models.UserAddress
	DeltaWei    *big.Int // nil for absolute updates
	AbsoluteWei string   // empty for delta updates
	SpenderAddr string   // for approve events
}

// helper: check if event should be processed (snapshot + idempotency)
func (s *contractParseService) shouldProcessEvent(height uint64, txIdx uint64, txHash string, logIdx uint, a *models.UserAddress) bool {
	if a == nil {
		return false
	}
	// Idempotency: skip if already parsed
	existing, err := s.repo.GetByTxHashAndLogIndex(context.Background(), txHash, logIdx)
	if err == nil && existing != nil {
		return false
	}
	// Gate by BalanceHeight: if stored balance snapshot height is newer than this event, skip
	if a.BalanceHeight > 0 && a.BalanceHeight > height {
		return false
	}
	return true
}

// helper: apply balance update (does not modify BalanceHeight)
func (s *contractParseService) applyBalanceUpdate(update *EventUpdate, txHash string, logIdx uint) error {
	if update == nil || update.UserAddress == nil {
		return nil
	}
	if !s.shouldProcessEvent(update.Height, update.TxIndex, txHash, logIdx, update.UserAddress) {
		return nil
	}
	a := update.UserAddress
	if update.EventType == "balanceOf" {
		val := update.AbsoluteWei
		a.ContractBalance = &val
	} else if update.DeltaWei != nil {
		cur := new(big.Int)
		if a.ContractBalance != nil {
			_, _ = cur.SetString(*a.ContractBalance, 10)
		}
		cur.Add(cur, update.DeltaWei)
		if cur.Sign() < 0 {
			cur.SetInt64(0)
		}
		val := cur.String()
		a.ContractBalance = &val
	} else if update.EventType == "approve" && update.SpenderAddr != "" {
		// 需要更新授权余额
	}
	return s.userAddressRepo.Update(a)
}

// parseReceipt 仅基于 logs_parser_rules 进行快速解析
func (s *contractParseService) parseReceipt(receipt *models.TransactionReceipt, tx *models.Transaction, parserConfigs []*models.ParserConfig, coinConfigs []*models.CoinConfig) []*models.ContractParseResult {
	if receipt == nil || receipt.LogsData == "" {
		return nil
	}
	var logs []map[string]interface{}
	if err := json.Unmarshal([]byte(receipt.LogsData), &logs); err != nil {
		return nil
	}
	// 预计算原始日志hash（用于幂等）
	rawHash := sha256.Sum256([]byte(receipt.LogsData))
	rawHashHex := "0x" + hex.EncodeToString(rawHash[:])

	// 交易上下文
	chain := receipt.Chain
	contractAddr := emptyIfNil(receipt.ContractAddress)

	// 建立 event_signature -> rules 的快速映射
	sigToRule := map[string]*models.ParserConfig{}
	for _, c := range parserConfigs {
		if c != nil && c.EventSignature != "" && (c.LogsParserRules.ExtractAmount != "" || c.LogsParserRules.ExtractFromAddress != "" || c.LogsParserRules.ExtractToAddress != "" || c.LogsParserRules.ExtractOwnerAddress != "" || c.LogsParserRules.ExtractSpenderAddress != "") {
			sig := strings.ToLower(c.EventSignature)
			sigToRule[sig] = c
		}
	}
	var out []*models.ContractParseResult

	for idx, log := range logs {
		topics, _ := log["topics"].([]interface{})
		data, _ := log["data"].(string)
		if len(topics) == 0 {
			continue
		}
		topic0, _ := topics[0].(string)
		sig := strings.ToLower(topic0)
		rule := sigToRule[sig]
		if rule == nil {
			continue
		}

		fromAddr := extractAddressByRule(rule.LogsParserRules.ExtractFromAddress, topics, data)
		toAddr := extractAddressByRule(rule.LogsParserRules.ExtractToAddress, topics, data)
		amountWei := extractAmountByRule(rule.LogsParserRules.ExtractAmount, data)
		ownerAddr := extractAddressByRule(rule.LogsParserRules.ExtractOwnerAddress, topics, data)
		spenderAddr := extractAddressByRule(rule.LogsParserRules.ExtractSpenderAddress, topics, data)

		// fmt.Printf("[parseReceipt] #%d matched sig=%s from=%s to=%s amountWei=%s owner=%s spender=%s\n", idx, sig, fromAddr, toAddr, amountWei, ownerAddr, spenderAddr)

		coinConfig, err := GetCoinConfigByContractAddress(coinConfigs, contractAddr)
		if err != nil {
			// fmt.Printf("[parseReceipt] #%d get coinConfig error: %v\n", idx, err)
			continue
		}

		// update user_address snapshots with strict ordering
		if receipt.Status == 1 && tx != nil {
			txIdx := uint64(receipt.TransactionIndex)
			logIdx := uint(idx)

			if rule.EventName == "Transfer" {
				amt := new(big.Int)
				if _, ok := amt.SetString(amountWei, 10); ok && amt.Sign() > 0 {
					// Process from address (subtract)
					if fromAddr != "" {
						fromAddress, _ := s.userAddressRepo.GetByContractAddress(fromAddr, coinConfig.ID)
						if fromAddress != nil {
							update := &EventUpdate{
								Height:      tx.Height,
								TxIndex:     txIdx,
								EventType:   "transfer",
								UserAddress: fromAddress,
								DeltaWei:    new(big.Int).Neg(amt),
							}
							_ = s.applyBalanceUpdate(update, receipt.TxHash, logIdx)
						}
					}
					// Process to address (add)
					if toAddr != "" {
						toAddress, _ := s.userAddressRepo.GetByContractAddress(toAddr, coinConfig.ID)
						if toAddress != nil {
							update := &EventUpdate{
								Height:      tx.Height,
								TxIndex:     txIdx,
								EventType:   "transfer",
								UserAddress: toAddress,
								DeltaWei:    new(big.Int).Set(amt),
							}
							_ = s.applyBalanceUpdate(update, receipt.TxHash, logIdx)
						}
					}
				}
			} else if rule.EventName == "TransferFrom" {
				amt := new(big.Int)
				if _, ok := amt.SetString(amountWei, 10); ok && amt.Sign() > 0 {
					// Process from address (subtract)
					if fromAddr != "" {
						fromAddress, _ := s.userAddressRepo.GetByContractApproveAddress(fromAddr, coinConfig.ID)
						if fromAddress != nil {
							update := &EventUpdate{
								Height:      tx.Height,
								TxIndex:     txIdx,
								EventType:   "transferFrom",
								UserAddress: fromAddress,
								DeltaWei:    new(big.Int).Neg(amt),
							}
							_ = s.applyBalanceUpdate(update, receipt.TxHash, logIdx)
						}
					}
					// Process to address (add)
					if toAddr != "" {
						toAddress, _ := s.userAddressRepo.GetByContractApproveAddress(toAddr, coinConfig.ID)
						if toAddress != nil {
							update := &EventUpdate{
								Height:      tx.Height,
								TxIndex:     txIdx,
								EventType:   "transferFrom",
								UserAddress: toAddress,
								DeltaWei:    new(big.Int).Set(amt),
							}
							_ = s.applyBalanceUpdate(update, receipt.TxHash, logIdx)
						}
					}
				}
			} else if rule.EventName == "Approve" {
				if ownerAddr != "" && spenderAddr != "" {
					owner, _ := s.userAddressRepo.GetByContractAddress(ownerAddr, coinConfig.ID)
					if owner != nil {
						update := &EventUpdate{
							Height:      tx.Height,
							TxIndex:     txIdx,
							EventType:   "approve",
							UserAddress: owner,
							SpenderAddr: spenderAddr,
						}
						_ = s.applyBalanceUpdate(update, receipt.TxHash, logIdx)
					}
				}
			} else if rule.EventName == "BalanceOf" {
				if ownerAddr != "" && amountWei != "0" {
					owner, _ := s.userAddressRepo.GetByContractAddress(ownerAddr, coinConfig.ID)
					if owner != nil {
						update := &EventUpdate{
							Height:      tx.Height,
							TxIndex:     txIdx,
							EventType:   "balanceOf",
							UserAddress: owner,
							AbsoluteWei: amountWei,
						}
						_ = s.applyBalanceUpdate(update, receipt.TxHash, logIdx)
					}
				}
			}
		}

		// result
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
		res.OwnerAddress = ownerAddr
		res.SpenderAddress = spenderAddr
		res.AmountWei = amountWei
		res.TokenDecimals = 0
		if tx != nil && tx.TokenDecimals != 0 {
			res.TokenDecimals = uint16(tx.TokenDecimals)
		}
		res.TokenSymbol = coinConfig.Symbol
		res.RawLogsHash = rawHashHex
		res.ParsedJSON = string(bj)
		out = append(out, res)
		// fmt.Printf("[parseReceipt] #%d produce result txHash=%s logIndex=%d symbol=%s decimals=%d\n", idx, res.TxHash, res.LogIndex, res.TokenSymbol, res.TokenDecimals)
	}
	// fmt.Printf("[parseReceipt] end txHash=%s produced=%d\n", emptyIfNil(receipt.TxHash), len(out))
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

func emptyIfNil(s string) string {
	return s
}
func GetCoinConfigByContractAddress(coinConfigs []*models.CoinConfig, contractAddress string) (*models.CoinConfig, error) {
	for _, c := range coinConfigs {
		if c.ContractAddr == contractAddress {
			return c, nil
		}
	}
	return nil, fmt.Errorf("coinConfig not found for contractAddress=%s", contractAddress)
}
