package services

import (
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type SolService interface {
	SaveTxDetail(ctx context.Context, req *dto.SolTxDetailCreateRequest) (*dto.SolTxDetailResponse, error)
	SaveTxDetailBatch(ctx context.Context, req *dto.BatchSolDataRequest) (*dto.BatchSolDataResponse, error)
	GetTxDetail(ctx context.Context, txID string) (*dto.SolTxDetailResponse, error)
	GetTxEvents(ctx context.Context, txID string) ([]dto.SolEventResponse, error)
	GetTxInstructions(ctx context.Context, txID string) ([]dto.SolInstructionResponse, error)
	GetTxsBySlot(ctx context.Context, slot uint64, page, pageSize int) ([]*dto.SolTxDetailResponse, int64, error)
	GetSlotStats(ctx context.Context, slot uint64) (*dto.SlotStatsResponse, error)
	ListTxDetails(ctx context.Context, slot *uint64, page, pageSize int) ([]*dto.SolTxDetailResponse, int64, error)
	GetArtifactsByTxID(ctx context.Context, txID string) (map[string]interface{}, error)
	// ParseUsingProgramRules 根据 SolProgram 规则解析交易明细，输出事件/指令/或扩展数据
	ParseUsingProgramRules(ctx context.Context, detail *models.SolTxDetail, programs map[string]*models.SolProgram) (*dto.SolTxDetailCreateRequest, map[string]interface{}, error)
	// Program maintenance
	CreateProgram(ctx context.Context, p *models.SolProgram) error
	UpdateProgram(ctx context.Context, p *models.SolProgram) error
	DeleteProgram(ctx context.Context, id uint) error
	GetProgramByID(ctx context.Context, id uint) (*models.SolProgram, error)
	GetProgramByProgramID(ctx context.Context, programID string) (*models.SolProgram, error)
	ListPrograms(ctx context.Context, page, pageSize int, keyword string) ([]*models.SolProgram, int64, error)
	GetAllPrograms(ctx context.Context) ([]*models.SolProgram, error)
	// Accessors for verification parse
	GetDetailsByBlockID(ctx context.Context, blockID uint64) ([]*models.SolTxDetail, error)
	SaveArtifacts(ctx context.Context, txID string, blockID *uint64, slot uint64, events []dto.SolEventRequest, instructions []dto.SolInstructionRequest, extras []models.SolParsedExtra) error
}

type solService struct {
	txDetailRepo    repository.SolTxDetailRepository
	eventRepo       repository.SolEventRepository
	instructionRepo repository.SolInstructionRepository
	programRepo     repository.SolProgramRepository
	parsedExtraRepo repository.SolParsedExtraRepository
}

func NewSolService(
	txDetailRepo repository.SolTxDetailRepository,
	eventRepo repository.SolEventRepository,
	instructionRepo repository.SolInstructionRepository,
	programRepo repository.SolProgramRepository,
	parsedExtraRepo repository.SolParsedExtraRepository,
) SolService {
	return &solService{
		txDetailRepo:    txDetailRepo,
		eventRepo:       eventRepo,
		instructionRepo: instructionRepo,
		programRepo:     programRepo,
		parsedExtraRepo: parsedExtraRepo,
	}
}

// SaveTxDetail 保存单个交易详情
func (s *solService) SaveTxDetail(ctx context.Context, req *dto.SolTxDetailCreateRequest) (*dto.SolTxDetailResponse, error) {
	// 转换DTO到Model
	detail := &models.SolTxDetail{
		TxID:              req.Detail.TxID,
		Slot:              req.Detail.Slot,
		BlockID:           req.Detail.BlockID,
		Blockhash:         req.Detail.Blockhash,
		RecentBlockhash:   req.Detail.RecentBlockhash,
		Version:           req.Detail.Version,
		Fee:               req.Detail.Fee,
		ComputeUnits:      req.Detail.ComputeUnits,
		Status:            req.Detail.Status,
		AccountKeys:       models.JSONText(req.Detail.AccountKeys),
		PreBalances:       models.JSONText(req.Detail.PreBalances),
		PostBalances:      models.JSONText(req.Detail.PostBalances),
		PreTokenBalances:  models.JSONText(req.Detail.PreTokenBalances),
		PostTokenBalances: models.JSONText(req.Detail.PostTokenBalances),
		Logs:              models.JSONText(req.Detail.Logs),
		Instructions:      models.JSONText(req.Detail.Instructions),
		InnerInstructions: models.JSONText(req.Detail.InnerInstructions),
		LoadedAddresses:   models.JSONText(req.Detail.LoadedAddresses),
		Rewards:           models.JSONText(req.Detail.Rewards),
		Events:            models.JSONText(req.Detail.Events),
		RawTransaction:    models.JSONText(req.Detail.RawTransaction),
		RawMeta:           models.JSONText(req.Detail.RawMeta),
		Ctime:             time.Now(),
		Mtime:             time.Now(),
	}

	// 设置默认值
	if detail.Status == "" {
		detail.Status = "success"
	}
	if detail.Version == "" {
		detail.Version = "legacy"
	}

	// 保存交易详情
	if err := s.txDetailRepo.Create(ctx, detail); err != nil {
		return nil, fmt.Errorf("failed to save transaction detail: %w", err)
	}

	// 保存事件
	if len(req.Events) > 0 {
		events := make([]*models.SolEvent, 0, len(req.Events))
		for _, eventReq := range req.Events {
			event := &models.SolEvent{
				TxID:        eventReq.TxID,
				BlockID:     eventReq.BlockID,
				Slot:        eventReq.Slot,
				EventIndex:  eventReq.EventIndex,
				EventType:   eventReq.EventType,
				ProgramID:   eventReq.ProgramID,
				FromAddress: eventReq.FromAddress,
				ToAddress:   eventReq.ToAddress,
				Amount:      eventReq.Amount,
				Mint:        eventReq.Mint,
				Decimals:    eventReq.Decimals,
				IsInner:     eventReq.IsInner,
				AssetType:   eventReq.AssetType,
				ExtraData:   models.JSONText(eventReq.ExtraData),
				Ctime:       time.Now(),
			}
			events = append(events, event)
		}
		if err := s.eventRepo.CreateBatch(ctx, events); err != nil {
			return nil, fmt.Errorf("failed to save events: %w", err)
		}
	}

	// 保存指令
	if len(req.Instructions) > 0 {
		instructions := make([]*models.SolInstruction, 0, len(req.Instructions))
		for _, instReq := range req.Instructions {
			instruction := &models.SolInstruction{
				TxID:             instReq.TxID,
				BlockID:          instReq.BlockID,
				Slot:             instReq.Slot,
				InstructionIndex: instReq.InstructionIndex,
				ProgramID:        instReq.ProgramID,
				Accounts:         models.JSONText(instReq.Accounts),
				Data:             instReq.Data,
				ParsedData:       models.JSONText(instReq.ParsedData),
				InstructionType:  instReq.InstructionType,
				IsInner:          instReq.IsInner,
				StackHeight:      instReq.StackHeight,
				Ctime:            time.Now(),
			}
			instructions = append(instructions, instruction)
		}
		if err := s.instructionRepo.CreateBatch(ctx, instructions); err != nil {
			return nil, fmt.Errorf("failed to save instructions: %w", err)
		}
	}

	// 转换Model到Response DTO
	return s.convertToResponse(detail), nil
}

// SaveTxDetailBatch 批量保存交易详情
func (s *solService) SaveTxDetailBatch(ctx context.Context, req *dto.BatchSolDataRequest) (*dto.BatchSolDataResponse, error) {
	response := &dto.BatchSolDataResponse{
		Success:   true,
		Processed: 0,
		Failed:    0,
		Errors:    []string{},
	}

	if len(req.Transactions) == 0 {
		response.Message = "no transactions to process"
		return response, nil
	}

	// 1) 组装三张表的批量数据
	details := make([]*models.SolTxDetail, 0, len(req.Transactions))
	events := make([]*models.SolEvent, 0)
	instructions := make([]*models.SolInstruction, 0)

	for i := range req.Transactions {
		t := &req.Transactions[i]

		// detail
		d := &models.SolTxDetail{
			TxID:              t.Detail.TxID,
			Slot:              t.Detail.Slot,
			BlockID:           t.Detail.BlockID,
			Blockhash:         t.Detail.Blockhash,
			RecentBlockhash:   t.Detail.RecentBlockhash,
			Version:           defaultString(t.Detail.Version, "legacy"),
			Fee:               t.Detail.Fee,
			ComputeUnits:      t.Detail.ComputeUnits,
			Status:            defaultString(t.Detail.Status, "success"),
			AccountKeys:       models.JSONText(t.Detail.AccountKeys),
			PreBalances:       models.JSONText(t.Detail.PreBalances),
			PostBalances:      models.JSONText(t.Detail.PostBalances),
			PreTokenBalances:  models.JSONText(t.Detail.PreTokenBalances),
			PostTokenBalances: models.JSONText(t.Detail.PostTokenBalances),
			Logs:              models.JSONText(t.Detail.Logs),
			Instructions:      models.JSONText(t.Detail.Instructions),
			InnerInstructions: models.JSONText(t.Detail.InnerInstructions),
			LoadedAddresses:   models.JSONText(t.Detail.LoadedAddresses),
			Rewards:           models.JSONText(t.Detail.Rewards),
			Events:            models.JSONText(t.Detail.Events),
			RawTransaction:    models.JSONText(t.Detail.RawTransaction),
			RawMeta:           models.JSONText(t.Detail.RawMeta),
			Ctime:             time.Now(),
			Mtime:             time.Now(),
		}
		details = append(details, d)

		// events
		for j := range t.Events {
			e := t.Events[j]
			events = append(events, &models.SolEvent{
				TxID:        e.TxID,
				BlockID:     e.BlockID,
				Slot:        e.Slot,
				EventIndex:  e.EventIndex,
				EventType:   e.EventType,
				ProgramID:   e.ProgramID,
				FromAddress: e.FromAddress,
				ToAddress:   e.ToAddress,
				Amount:      e.Amount,
				Mint:        e.Mint,
				Decimals:    e.Decimals,
				IsInner:     e.IsInner,
				AssetType:   e.AssetType,
				ExtraData:   models.JSONText(e.ExtraData),
				Ctime:       time.Now(),
			})
		}

		// instructions
		for k := range t.Instructions {
			in := t.Instructions[k]
			instructions = append(instructions, &models.SolInstruction{
				TxID:             in.TxID,
				BlockID:          in.BlockID,
				Slot:             in.Slot,
				InstructionIndex: in.InstructionIndex,
				ProgramID:        in.ProgramID,
				Accounts:         models.JSONText(in.Accounts),
				Data:             in.Data,
				ParsedData:       models.JSONText(in.ParsedData),
				InstructionType:  in.InstructionType,
				IsInner:          in.IsInner,
				StackHeight:      in.StackHeight,
				Ctime:            time.Now(),
			})
		}
	}

	// 2) 批量写入（三表分开批处理，避免单条循环）
	// 注意：若需要严格原子性，可引入事务（这里根据现有仓储接口逐个执行）
	if err := s.txDetailRepo.CreateBatch(ctx, details); err != nil {
		response.Success = false
		response.Message = "batch insert details failed"
		response.Errors = append(response.Errors, err.Error())
		return response, nil
	}

	if len(events) > 0 {
		if err := s.eventRepo.CreateBatch(ctx, events); err != nil {
			response.Success = false
			response.Message = "batch insert events partially failed"
			response.Errors = append(response.Errors, err.Error())
			// 不中断，尽量写入其他表
		}
	}

	if len(instructions) > 0 {
		if err := s.instructionRepo.CreateBatch(ctx, instructions); err != nil {
			response.Success = false
			response.Message = "batch insert instructions partially failed"
			response.Errors = append(response.Errors, err.Error())
		}
	}

	response.Processed = len(details)
	if response.Success && response.Message == "" {
		response.Message = fmt.Sprintf("Successfully processed %d transactions", response.Processed)
	}
	return response, nil
}

func defaultString(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// GetTxDetail 获取交易详情
func (s *solService) GetTxDetail(ctx context.Context, txID string) (*dto.SolTxDetailResponse, error) {
	detail, err := s.txDetailRepo.GetByTxID(ctx, txID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction detail: %w", err)
	}

	return s.convertToResponse(detail), nil
}

// GetTxEvents 获取交易事件
func (s *solService) GetTxEvents(ctx context.Context, txID string) ([]dto.SolEventResponse, error) {
	events, err := s.eventRepo.GetByTxID(ctx, txID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction events: %w", err)
	}

	responses := make([]dto.SolEventResponse, 0, len(events))
	for _, event := range events {
		responses = append(responses, dto.SolEventResponse{
			ID:          event.ID,
			TxID:        event.TxID,
			BlockID:     event.BlockID,
			Slot:        event.Slot,
			EventIndex:  event.EventIndex,
			EventType:   event.EventType,
			ProgramID:   event.ProgramID,
			FromAddress: event.FromAddress,
			ToAddress:   event.ToAddress,
			Amount:      event.Amount,
			Mint:        event.Mint,
			Decimals:    event.Decimals,
			IsInner:     event.IsInner,
			AssetType:   event.AssetType,
			ExtraData:   string(event.ExtraData),
			Ctime:       event.Ctime.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

// GetTxInstructions 获取交易指令
func (s *solService) GetTxInstructions(ctx context.Context, txID string) ([]dto.SolInstructionResponse, error) {
	instructions, err := s.instructionRepo.GetByTxID(ctx, txID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction instructions: %w", err)
	}

	responses := make([]dto.SolInstructionResponse, 0, len(instructions))
	for _, inst := range instructions {
		responses = append(responses, dto.SolInstructionResponse{
			ID:               inst.ID,
			TxID:             inst.TxID,
			BlockID:          inst.BlockID,
			Slot:             inst.Slot,
			InstructionIndex: inst.InstructionIndex,
			ProgramID:        inst.ProgramID,
			Accounts:         string(inst.Accounts),
			Data:             inst.Data,
			ParsedData:       string(inst.ParsedData),
			InstructionType:  inst.InstructionType,
			IsInner:          inst.IsInner,
			StackHeight:      inst.StackHeight,
			Ctime:            inst.Ctime.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

// ListTxDetails 分页查询sol_tx_detail，可按slot过滤
func (s *solService) ListTxDetails(ctx context.Context, slot *uint64, page, pageSize int) ([]*dto.SolTxDetailResponse, int64, error) {
	var (
		details []*models.SolTxDetail
		total   int64
		err     error
	)
	if slot != nil {
		details, total, err = s.txDetailRepo.GetBySlot(ctx, *slot, page, pageSize)
	} else {
		// 复用 List：GetBySlot 无slot时可新建List接口，这里简单调用GetBySlot(0)不合适，暂用GetBySlot分支要求slot
		details, total, err = s.txDetailRepo.GetBySlot(ctx, 0, page, pageSize)
	}
	if err != nil {
		return nil, 0, err
	}
	res := make([]*dto.SolTxDetailResponse, 0, len(details))
	for _, d := range details {
		res = append(res, s.convertToResponse(d))
	}
	return res, total, nil
}

// GetArtifactsByTxID 按txid查询指令与事件
func (s *solService) GetArtifactsByTxID(ctx context.Context, txID string) (map[string]interface{}, error) {
	events, err := s.eventRepo.GetByTxID(ctx, txID)
	if err != nil {
		return nil, err
	}
	insts, err := s.instructionRepo.GetByTxID(ctx, txID)
	if err != nil {
		return nil, err
	}
	evRes := make([]dto.SolEventResponse, 0, len(events))
	for _, e := range events {
		evRes = append(evRes, dto.SolEventResponse{
			ID:          e.ID,
			TxID:        e.TxID,
			BlockID:     e.BlockID,
			Slot:        e.Slot,
			EventIndex:  e.EventIndex,
			EventType:   e.EventType,
			ProgramID:   e.ProgramID,
			FromAddress: e.FromAddress,
			ToAddress:   e.ToAddress,
			Amount:      e.Amount,
			Mint:        e.Mint,
			Decimals:    e.Decimals,
			IsInner:     e.IsInner,
			AssetType:   e.AssetType,
			ExtraData:   string(e.ExtraData),
			Ctime:       e.Ctime.Format("2006-01-02 15:04:05"),
		})
	}
	inRes := make([]dto.SolInstructionResponse, 0, len(insts))
	for _, in := range insts {
		inRes = append(inRes, dto.SolInstructionResponse{
			ID:               in.ID,
			TxID:             in.TxID,
			BlockID:          in.BlockID,
			Slot:             in.Slot,
			InstructionIndex: in.InstructionIndex,
			ProgramID:        in.ProgramID,
			Accounts:         string(in.Accounts),
			Data:             in.Data,
			ParsedData:       string(in.ParsedData),
			InstructionType:  in.InstructionType,
			IsInner:          in.IsInner,
			StackHeight:      in.StackHeight,
			Ctime:            in.Ctime.Format("2006-01-02 15:04:05"),
		})
	}
	return map[string]interface{}{
		"events":       evRes,
		"instructions": inRes,
	}, nil
}

// ParseUsingProgramRules 根据 SolProgram 规则解析交易明细，输出事件/指令/或扩展数据
// 1) 读取 detail.Instructions / InnerInstructions 中 programId
// 2) 命中 programRepo 维护的规则，按规则映射生成 SolEventRequest 或 SolInstructionRequest
// 3) 返回可被 SaveTxDetail 使用的请求对象；如有无法归档的数据，放入 extra map
func (s *solService) ParseUsingProgramRules(ctx context.Context, detail *models.SolTxDetail, programs map[string]*models.SolProgram) (*dto.SolTxDetailCreateRequest, map[string]interface{}, error) {
	req := &dto.SolTxDetailCreateRequest{
		Detail: dto.SolTxDetailRequest{
			TxID:              detail.TxID,
			Slot:              detail.Slot,
			BlockID:           detail.BlockID,
			Blockhash:         detail.Blockhash,
			RecentBlockhash:   detail.RecentBlockhash,
			Version:           detail.Version,
			Fee:               detail.Fee,
			ComputeUnits:      detail.ComputeUnits,
			Status:            detail.Status,
			AccountKeys:       string(detail.AccountKeys),
			PreBalances:       string(detail.PreBalances),
			PostBalances:      string(detail.PostBalances),
			PreTokenBalances:  string(detail.PreTokenBalances),
			PostTokenBalances: string(detail.PostTokenBalances),
			Logs:              string(detail.Logs),
			Instructions:      string(detail.Instructions),
			InnerInstructions: string(detail.InnerInstructions),
			LoadedAddresses:   string(detail.LoadedAddresses),
			Rewards:           string(detail.Rewards),
			Events:            string(detail.Events),
			RawTransaction:    string(detail.RawTransaction),
			RawMeta:           string(detail.RawMeta),
		},
	}

	extra := map[string]interface{}{}
	var events []dto.SolEventRequest
	var instructions []dto.SolInstructionRequest

	// 解析主指令
	if detail.Instructions != "" {
		mainInstructions, err := s.parseInstructions(string(detail.Instructions), false)
		if err != nil {
			return req, extra, fmt.Errorf("failed to parse main instructions: %w", err)
		}

		for i, inst := range mainInstructions {
			// 根据程序规则解析指令
			parsedEvents, parsedInsts, extraData, err := s.parseInstructionWithRules(inst, programs, detail.TxID, detail.BlockID, detail.Slot, i, false)
			if err != nil {
				// 记录错误但不中断处理
				if extraData == nil {
					extraData = make(map[string]interface{})
				}
				extraData["parse_error"] = err.Error()
			}

			events = append(events, parsedEvents...)
			instructions = append(instructions, parsedInsts...)

			if len(extraData) > 0 {
				extra[fmt.Sprintf("instruction_%d", i)] = extraData
			}
		}
	}

	// 解析内部指令
	if detail.InnerInstructions != "" {
		innerInstructions, err := s.parseInnerInstructions(string(detail.InnerInstructions))
		if err != nil {
			return req, extra, fmt.Errorf("failed to parse inner instructions: %w", err)
		}

		for outerIdx, innerGroup := range innerInstructions {
			if innerInsts, ok := innerGroup["instructions"].([]interface{}); ok {
				for innerIdx, innerInst := range innerInsts {
					if instMap, ok := innerInst.(map[string]interface{}); ok {
						// 根据程序规则解析内部指令
						parsedEvents, parsedInsts, extraData, err := s.parseInstructionWithRules(instMap, programs, detail.TxID, detail.BlockID, detail.Slot, innerIdx, true)
						if err != nil {
							if extraData == nil {
								extraData = make(map[string]interface{})
							}
							extraData["parse_error"] = err.Error()
						}

						events = append(events, parsedEvents...)
						instructions = append(instructions, parsedInsts...)

						if len(extraData) > 0 {
							extra[fmt.Sprintf("inner_instruction_%d_%d", outerIdx, innerIdx)] = extraData
						}
					}
				}
			}
		}
	}

	req.Events = events
	req.Instructions = instructions

	return req, extra, nil
}

func (s *solService) CreateProgram(ctx context.Context, p *models.SolProgram) error {
	return s.programRepo.Create(ctx, p)
}

func (s *solService) UpdateProgram(ctx context.Context, p *models.SolProgram) error {
	return s.programRepo.Update(ctx, p)
}

func (s *solService) DeleteProgram(ctx context.Context, id uint) error {
	return s.programRepo.Delete(ctx, id)
}

func (s *solService) GetProgramByID(ctx context.Context, id uint) (*models.SolProgram, error) {
	return s.programRepo.GetByID(ctx, id)
}

func (s *solService) GetProgramByProgramID(ctx context.Context, programID string) (*models.SolProgram, error) {
	return s.programRepo.GetByProgramID(ctx, programID)
}

func (s *solService) ListPrograms(ctx context.Context, page, pageSize int, keyword string) ([]*models.SolProgram, int64, error) {
	return s.programRepo.List(ctx, page, pageSize, keyword)
}

func (s *solService) GetAllPrograms(ctx context.Context) ([]*models.SolProgram, error) {
	return s.programRepo.GetAll(ctx)
}

// Accessors for verification parse
func (s *solService) GetDetailsByBlockID(ctx context.Context, blockID uint64) ([]*models.SolTxDetail, error) {
	return s.txDetailRepo.GetByBlockID(ctx, blockID)
}

func (s *solService) SaveArtifacts(ctx context.Context, txID string, blockID *uint64, slot uint64, events []dto.SolEventRequest, instructions []dto.SolInstructionRequest, extras []models.SolParsedExtra) error {
	if len(events) > 0 {
		batch := make([]*models.SolEvent, 0, len(events))
		for _, e := range events {
			batch = append(batch, &models.SolEvent{
				TxID:        e.TxID,
				BlockID:     e.BlockID,
				Slot:        e.Slot,
				EventIndex:  e.EventIndex,
				EventType:   e.EventType,
				ProgramID:   e.ProgramID,
				FromAddress: e.FromAddress,
				ToAddress:   e.ToAddress,
				Amount:      e.Amount,
				Mint:        e.Mint,
				Decimals:    e.Decimals,
				IsInner:     e.IsInner,
				AssetType:   e.AssetType,
				ExtraData:   models.JSONText(e.ExtraData),
				Ctime:       time.Now(),
			})
		}
		if err := s.eventRepo.CreateBatch(ctx, batch); err != nil {
			return err
		}
	}
	if len(instructions) > 0 {
		batch := make([]*models.SolInstruction, 0, len(instructions))
		for _, in := range instructions {
			batch = append(batch, &models.SolInstruction{
				TxID:             in.TxID,
				BlockID:          in.BlockID,
				Slot:             in.Slot,
				InstructionIndex: in.InstructionIndex,
				ProgramID:        in.ProgramID,
				Accounts:         models.JSONText(in.Accounts),
				Data:             in.Data,
				ParsedData:       models.JSONText(in.ParsedData),
				InstructionType:  in.InstructionType,
				IsInner:          in.IsInner,
				StackHeight:      in.StackHeight,
				Ctime:            time.Now(),
			})
		}
		if err := s.instructionRepo.CreateBatch(ctx, batch); err != nil {
			return err
		}
	}
	if len(extras) > 0 {
		// fill common fields
		batch := make([]*models.SolParsedExtra, 0, len(extras))
		for i := range extras {
			ex := extras[i]
			ex.BlockID = blockID
			ex.Slot = slot
			ex.Ctime = time.Now()
			exCopy := ex
			batch = append(batch, &exCopy)
		}
		if err := s.parsedExtraRepo.CreateBatch(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

// GetTxsBySlot 根据slot获取交易列表
func (s *solService) GetTxsBySlot(ctx context.Context, slot uint64, page, pageSize int) ([]*dto.SolTxDetailResponse, int64, error) {
	details, total, err := s.txDetailRepo.GetBySlot(ctx, slot, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions by slot: %w", err)
	}

	responses := make([]*dto.SolTxDetailResponse, 0, len(details))
	for _, detail := range details {
		responses = append(responses, s.convertToResponse(detail))
	}

	return responses, total, nil
}

// GetSlotStats 获取slot统计信息
func (s *solService) GetSlotStats(ctx context.Context, slot uint64) (*dto.SlotStatsResponse, error) {
	// 获取交易列表
	details, _, err := s.txDetailRepo.GetBySlot(ctx, slot, 1, 1000) // 限制1000笔交易用于统计
	if err != nil {
		return nil, fmt.Errorf("failed to get slot transactions: %w", err)
	}

	stats := &dto.SlotStatsResponse{
		Slot:             slot,
		TransactionCount: len(details),
		TotalFees:        "0",
		SuccessfulTxs:    0,
		FailedTxs:        0,
		ComputeUnitsUsed: 0,
	}

	var totalFees uint64
	for _, detail := range details {
		totalFees += detail.Fee
		stats.ComputeUnitsUsed += detail.ComputeUnits

		if detail.Status == "success" {
			stats.SuccessfulTxs++
		} else {
			stats.FailedTxs++
		}
	}

	stats.TotalFees = fmt.Sprintf("%d", totalFees)
	return stats, nil
}

// convertToResponse 将Model转换为Response DTO
func (s *solService) convertToResponse(detail *models.SolTxDetail) *dto.SolTxDetailResponse {
	return &dto.SolTxDetailResponse{
		ID:                detail.ID,
		TxID:              detail.TxID,
		Slot:              detail.Slot,
		BlockID:           detail.BlockID,
		Blockhash:         detail.Blockhash,
		RecentBlockhash:   detail.RecentBlockhash,
		Version:           detail.Version,
		Fee:               detail.Fee,
		ComputeUnits:      detail.ComputeUnits,
		Status:            detail.Status,
		AccountKeys:       string(detail.AccountKeys),
		PreBalances:       string(detail.PreBalances),
		PostBalances:      string(detail.PostBalances),
		PreTokenBalances:  string(detail.PreTokenBalances),
		PostTokenBalances: string(detail.PostTokenBalances),
		Logs:              string(detail.Logs),
		Instructions:      string(detail.Instructions),
		InnerInstructions: string(detail.InnerInstructions),
		LoadedAddresses:   string(detail.LoadedAddresses),
		Rewards:           string(detail.Rewards),
		Events:            string(detail.Events),
		RawTransaction:    string(detail.RawTransaction),
		RawMeta:           string(detail.RawMeta),
		Ctime:             detail.Ctime.Format("2006-01-02 15:04:05"),
		Mtime:             detail.Mtime.Format("2006-01-02 15:04:05"),
	}
}

// parseInstructions 解析主指令JSON
func (s *solService) parseInstructions(instructionsJSON string, isInner bool) ([]map[string]interface{}, error) {
	var instructions []map[string]interface{}
	if err := json.Unmarshal([]byte(instructionsJSON), &instructions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal instructions: %w", err)
	}
	return instructions, nil
}

// parseInnerInstructions 解析内部指令JSON
func (s *solService) parseInnerInstructions(innerInstructionsJSON string) ([]map[string]interface{}, error) {
	var innerInstructions []map[string]interface{}
	if err := json.Unmarshal([]byte(innerInstructionsJSON), &innerInstructions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inner instructions: %w", err)
	}
	return innerInstructions, nil
}

// parseInstructionWithRules 根据程序规则解析单个指令
func (s *solService) parseInstructionWithRules(
	instruction map[string]interface{},
	programs map[string]*models.SolProgram,
	txID string,
	blockID *uint64,
	slot uint64,
	index int,
	isInner bool,
) ([]dto.SolEventRequest, []dto.SolInstructionRequest, map[string]interface{}, error) {

	var events []dto.SolEventRequest
	var instructions []dto.SolInstructionRequest
	extraData := make(map[string]interface{})

	// 提取基础信息
	programID, _ := instruction["programId"].(string)
	if programID == "" {
		return events, instructions, extraData, fmt.Errorf("missing programId in instruction")
	}

	accounts, _ := instruction["accounts"].([]interface{})
	data, _ := instruction["data"].(string)
	parsed, _ := instruction["parsed"]

	// 构建基础指令记录
	instReq := dto.SolInstructionRequest{
		TxID:             txID,
		BlockID:          blockID,
		Slot:             slot,
		InstructionIndex: index,
		ProgramID:        programID,
		Data:             data,
		IsInner:          isInner,
		StackHeight:      1,
	}

	// 序列化accounts
	if len(accounts) > 0 {
		accountsBytes, _ := json.Marshal(accounts)
		instReq.Accounts = string(accountsBytes)
	}

	// 序列化parsed data
	if parsed != nil {
		parsedBytes, _ := json.Marshal(parsed)
		instReq.ParsedData = string(parsedBytes)
	}

	// 根据程序类型解析
	program, exists := programs[programID]
	if !exists {
		// 未知程序，记录原始数据到扩展表
		extraData["unknown_program"] = map[string]interface{}{
			"program_id":   programID,
			"instruction":  instruction,
			"parse_status": "unknown_program",
		}
		instructions = append(instructions, instReq)
		return events, instructions, extraData, nil
	}

	// 根据程序类型和规则解析
	switch program.Type {
	case "system":
		parsedEvents, err := s.parseSystemInstruction(instruction, txID, blockID, slot, index, isInner)
		if err == nil {
			events = append(events, parsedEvents...)
		} else {
			extraData["system_parse_error"] = err.Error()
		}
		instReq.InstructionType = "system"

	case "token":
		parsedEvents, err := s.parseTokenInstruction(instruction, txID, blockID, slot, index, isInner)
		if err == nil {
			events = append(events, parsedEvents...)
		} else {
			extraData["token_parse_error"] = err.Error()
		}
		instReq.InstructionType = "token"

	case "spl-token":
		parsedEvents, err := s.parseSPLTokenInstruction(instruction, txID, blockID, slot, index, isInner)
		if err == nil {
			events = append(events, parsedEvents...)
		} else {
			extraData["spl_token_parse_error"] = err.Error()
		}
		instReq.InstructionType = "spl-token"

	default:
		// 使用自定义规则解析
		if program.InstructionRules != "" || program.EventRules != "" {
			parsedEvents, err := s.parseWithCustomRules(instruction, program, txID, blockID, slot, index, isInner)
			if err == nil {
				events = append(events, parsedEvents...)
			} else {
				extraData["custom_parse_error"] = err.Error()
			}
		}
		instReq.InstructionType = program.Type
	}

	instructions = append(instructions, instReq)
	return events, instructions, extraData, nil
}

// parseSystemInstruction 解析系统程序指令
func (s *solService) parseSystemInstruction(
	instruction map[string]interface{},
	txID string,
	blockID *uint64,
	slot uint64,
	index int,
	isInner bool,
) ([]dto.SolEventRequest, error) {

	var events []dto.SolEventRequest

	parsed, ok := instruction["parsed"].(map[string]interface{})
	if !ok {
		return events, fmt.Errorf("no parsed data for system instruction")
	}

	instType, _ := parsed["type"].(string)
	info, _ := parsed["info"].(map[string]interface{})

	switch instType {
	case "transfer":
		// SOL 转账
		event := dto.SolEventRequest{
			TxID:        txID,
			BlockID:     blockID,
			Slot:        slot,
			EventIndex:  index,
			EventType:   "transfer",
			ProgramID:   "11111111111111111111111111111112", // System Program
			FromAddress: getString(info, "source"),
			ToAddress:   getString(info, "destination"),
			Amount:      fmt.Sprintf("%v", info["lamports"]),
			Mint:        "So11111111111111111111111111111111111111112", // SOL mint
			Decimals:    9,
			IsInner:     isInner,
			AssetType:   "NATIVE",
		}
		events = append(events, event)

	case "createAccount":
		// 创建账户
		event := dto.SolEventRequest{
			TxID:        txID,
			BlockID:     blockID,
			Slot:        slot,
			EventIndex:  index,
			EventType:   "createAccount",
			ProgramID:   "11111111111111111111111111111112",
			FromAddress: getString(info, "source"),
			ToAddress:   getString(info, "newAccount"),
			Amount:      fmt.Sprintf("%v", info["lamports"]),
			Mint:        "So11111111111111111111111111111111111111112",
			Decimals:    9,
			IsInner:     isInner,
			AssetType:   "NATIVE",
		}
		extraDataBytes, _ := json.Marshal(map[string]interface{}{
			"space": info["space"],
			"owner": info["owner"],
		})
		event.ExtraData = string(extraDataBytes)
		events = append(events, event)
	}

	return events, nil
}

// parseTokenInstruction 解析Token程序指令
func (s *solService) parseTokenInstruction(
	instruction map[string]interface{},
	txID string,
	blockID *uint64,
	slot uint64,
	index int,
	isInner bool,
) ([]dto.SolEventRequest, error) {

	var events []dto.SolEventRequest

	parsed, ok := instruction["parsed"].(map[string]interface{})
	if !ok {
		return events, fmt.Errorf("no parsed data for token instruction")
	}

	instType, _ := parsed["type"].(string)
	info, _ := parsed["info"].(map[string]interface{})

	switch instType {
	case "transfer":
		// Token 转账
		event := dto.SolEventRequest{
			TxID:        txID,
			BlockID:     blockID,
			Slot:        slot,
			EventIndex:  index,
			EventType:   "transfer",
			ProgramID:   "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA", // SPL Token Program
			FromAddress: getString(info, "source"),
			ToAddress:   getString(info, "destination"),
			Amount:      getString(info, "amount"),
			Mint:        getString(info, "mint"),
			Decimals:    getInt(info, "decimals", 6), // 默认6位精度
			IsInner:     isInner,
			AssetType:   "TOKEN",
		}
		if authority, ok := info["authority"].(string); ok {
			extraDataBytes, _ := json.Marshal(map[string]interface{}{
				"authority": authority,
			})
			event.ExtraData = string(extraDataBytes)
		}
		events = append(events, event)

	case "mintTo":
		// Token 铸造
		event := dto.SolEventRequest{
			TxID:       txID,
			BlockID:    blockID,
			Slot:       slot,
			EventIndex: index,
			EventType:  "mint",
			ProgramID:  "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
			ToAddress:  getString(info, "account"),
			Amount:     getString(info, "amount"),
			Mint:       getString(info, "mint"),
			Decimals:   getInt(info, "decimals", 6),
			IsInner:    isInner,
			AssetType:  "TOKEN",
		}
		if authority, ok := info["mintAuthority"].(string); ok {
			extraDataBytes, _ := json.Marshal(map[string]interface{}{
				"mint_authority": authority,
			})
			event.ExtraData = string(extraDataBytes)
		}
		events = append(events, event)
	}

	return events, nil
}

// parseSPLTokenInstruction 解析SPL Token程序指令 (TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA)
func (s *solService) parseSPLTokenInstruction(
	instruction map[string]interface{},
	txID string,
	blockID *uint64,
	slot uint64,
	index int,
	isInner bool,
) ([]dto.SolEventRequest, error) {
	// SPL Token 和 Token 程序处理逻辑相同
	return s.parseTokenInstruction(instruction, txID, blockID, slot, index, isInner)
}

// parseWithCustomRules 使用自定义规则解析指令
func (s *solService) parseWithCustomRules(
	instruction map[string]interface{},
	program *models.SolProgram,
	txID string,
	blockID *uint64,
	slot uint64,
	index int,
	isInner bool,
) ([]dto.SolEventRequest, error) {

	var events []dto.SolEventRequest

	// 这里可以根据 program.InstructionRules 和 program.EventRules 中的JSON规则
	// 实现自定义解析逻辑，例如：
	// 1. 根据指令类型匹配规则
	// 2. 提取特定字段作为事件数据
	// 3. 应用转换规则

	// 现在提供一个基础实现，将parsed数据转换为通用事件
	parsed, ok := instruction["parsed"].(map[string]interface{})
	if !ok {
		return events, nil
	}

	instType, _ := parsed["type"].(string)
	info, _ := parsed["info"].(map[string]interface{})

	// 基于程序类型生成通用事件
	event := dto.SolEventRequest{
		TxID:       txID,
		BlockID:    blockID,
		Slot:       slot,
		EventIndex: index,
		EventType:  instType,
		ProgramID:  program.ProgramID,
		IsInner:    isInner,
		AssetType:  "CUSTOM",
	}

	// 尝试提取常见字段
	if from, ok := info["source"].(string); ok {
		event.FromAddress = from
	}
	if to, ok := info["destination"].(string); ok {
		event.ToAddress = to
	}
	if amount, ok := info["amount"]; ok {
		event.Amount = fmt.Sprintf("%v", amount)
	}
	if mint, ok := info["mint"].(string); ok {
		event.Mint = mint
	}

	// 将完整的info作为额外数据
	if len(info) > 0 {
		extraDataBytes, _ := json.Marshal(info)
		event.ExtraData = string(extraDataBytes)
	}

	events = append(events, event)
	return events, nil
}

// 辅助函数
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getInt(data map[string]interface{}, key string, defaultVal int) int {
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	if val, ok := data[key].(int); ok {
		return val
	}
	return defaultVal
}
