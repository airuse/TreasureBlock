// ==================== Sol 业务实体类型 ====================

// 交易详情项
export interface SolTxDetailItem {
  id: number
  tx_id: string
  slot: number
  block_id?: number
  blockhash: string
  recent_blockhash: string
  version: string
  fee: number
  compute_units: number
  status: string
  account_keys: string
  pre_balances: string
  post_balances: string
  pre_token_balances: string
  post_token_balances: string
  logs: string
  instructions: string
  inner_instructions: string
  loaded_addresses: string
  rewards: string
  events: string
  raw_transaction: string
  raw_meta: string
  ctime: string
  mtime: string
}

// 指令响应项
export interface SolInstructionItem {
  id: number
  tx_id: string
  block_id?: number
  slot: number
  instruction_index: number
  program_id: string
  accounts: string
  data: string
  parsed_data: string
  instruction_type: string
  is_inner: boolean
  stack_height: number
  ctime: string
}

// 事件响应项
export interface SolEventItem {
  id: number
  tx_id: string
  block_id?: number
  slot: number
  event_index: number
  event_type: string
  program_id: string
  from_address: string
  to_address: string
  amount: string
  mint: string
  decimals: number
  is_inner: boolean
  asset_type: string
  extra_data: string
  ctime: string
}

// ==================== 请求类型 ====================

export interface ListSolTxDetailsRequest {
  slot?: number
  page?: number
  page_size?: number
}


