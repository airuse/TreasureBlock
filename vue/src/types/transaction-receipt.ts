// 交易凭证类型定义

export interface TransactionReceipt {
  // 交易基本信息
  tx_hash: string
  tx_type: number
  status: number
  gas_used: number
  effective_gas_price: string
  blob_gas_used: number
  blob_gas_price: string
  block_hash: string
  block_number: number
  transaction_index: number
  chain: string
  
  // 交易输入数据（用于前端解析）
  input_data?: string
  
  // 日志数据（用于前端解析）
  logs_data?: string
  
  // 合约地址
  contract_address?: string
  
  // 时间信息
  created_at: string
  updated_at: string
  
  // 解析配置数据（用于前端解析）
  parser_configs?: ParserConfigInfo[]
}

export interface ParserConfigInfo {
  function_signature: string
  function_name: string
  function_description: string
  display_format: string
  parser_type?: string
  param_config?: ParamConfigInfo[]
  parser_rules?: Record<string, any>
  
  // 日志解析相关字段
  logs_parser_type?: string
  event_signature?: string
  event_name?: string
  event_description?: string
  logs_param_config?: LogsParamConfigInfo[]
  logs_parser_rules?: Record<string, any>
  logs_display_format?: string
}

export interface ParamConfigInfo {
  name: string
  type: string
  offset: number
  length: number
  description: string
}

export interface LogsParamConfigInfo {
  name: string
  type: string
  topic_index?: number
  data_index?: number
  description: string
}
