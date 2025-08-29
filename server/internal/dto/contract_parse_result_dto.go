package dto

import "blockChainBrowser/server/internal/models"

type ContractParseResultResponse struct {
	ID              uint   `json:"id"`
	TxHash          string `json:"tx_hash"`
	ContractAddress string `json:"contract_address"`
	Chain           string `json:"chain"`
	BlockNumber     uint64 `json:"block_number"`
	LogIndex        uint   `json:"log_index"`
	EventSignature  string `json:"event_signature"`
	EventName       string `json:"event_name"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	AmountWei       string `json:"amount_wei"`
	TokenDecimals   uint16 `json:"token_decimals"`
	TokenSymbol     string `json:"token_symbol"`
}

func NewContractParseResultResponse(m *models.ContractParseResult) *ContractParseResultResponse {
	return &ContractParseResultResponse{
		ID:              m.ID,
		TxHash:          m.TxHash,
		ContractAddress: m.ContractAddress,
		Chain:           m.Chain,
		BlockNumber:     m.BlockNumber,
		LogIndex:        m.LogIndex,
		EventSignature:  m.EventSignature,
		EventName:       m.EventName,
		FromAddress:     m.FromAddress,
		ToAddress:       m.ToAddress,
		AmountWei:       m.AmountWei,
		TokenDecimals:   m.TokenDecimals,
		TokenSymbol:     m.TokenSymbol,
	}
}
