// 币种配置基础类型
export interface CoinConfig {
  id: number
  contract_addr: string
  chain_name: string
  symbol: string
  coin_type: number
  precision: number
  decimals: number
  name: string
  logo_url: string
  website_url?: string
  explorer_url?: string
  description?: string
  market_cap_rank?: number
  is_stablecoin?: boolean
  is_verified: boolean
  status: number
  created_at?: string
  updated_at?: string
}

// 币种配置列表项类型
export interface CoinConfigListItem extends CoinConfig {
  // 继承基础类型，列表展示需要的字段
}

// 币种配置详情类型
export interface CoinConfigDetail extends CoinConfig {
  description?: string
  // 其他详情字段
}

// 解析配置类型
export interface ParserConfig {
  id: number
  contract_address: string
  function_name: string
  function_signature: string
  function_description: string
  display_format: string
  param_config: any
  parser_rules: any
  priority: number
  is_active: boolean
  // 日志解析配置字段
  logs_parser_type?: string
  event_signature?: string
  event_name?: string
  event_description?: string
  logs_param_config?: any
  logs_parser_rules?: any
  logs_display_format?: string
  created_at?: string
  updated_at?: string
}

// 币种配置维护响应类型
export interface CoinConfigMaintenanceResponse {
  coin_config: CoinConfig | null
  parser_configs: ParserConfig[]
  contract_address: string
}
