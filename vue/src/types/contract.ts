// ==================== 合约业务实体类型定义 ====================

// 基础合约类型 - 匹配实际后端数据结构
export interface Contract {
  id: number
  address: string // 修复：实际后端返回的是 address
  chain_name: string
  contract_type: string
  name: string // 修复：实际后端返回的是 name
  symbol: string // 修复：实际后端返回的是 symbol
  decimals: number
  total_supply: string
  is_erc20: boolean
  interfaces: string | string[]
  methods: string | string[]
  events: string | string[]
  metadata: string | Record<string, string>
  status: number // 修复：实际后端返回的是数字状态
  verified: boolean
  creator: string
  creation_tx: string
  creation_block: number
  contract_logo?: string // 新增：合约Logo图片(Base64编码)
  c_time: string
  m_time: string
}

// 合约列表项类型
export interface ContractListItem {
  id: number
  chain_name: string
  contract_address: string
  contract_name: string
  contract_symbol: string
  contract_type: string
  decimals: number
  description: string
  logo_url: string
  status: string
  created_at: string
}

// 合约详情类型
export interface ContractDetail extends Contract {
  // 继承基础合约类型的所有字段
  // 可以在这里添加详情页面特有的字段
}

// 合约创建/更新请求类型
export interface ContractCreateRequest {
  chain_name: string
  contract_address: string
  contract_name: string
  contract_symbol: string
  contract_type: string
  decimals: number
  description?: string
  logo_url?: string
  website_url?: string
  explorer_url?: string
}

// 合约状态更新请求类型
export interface ContractStatusUpdateRequest {
  status: string
}

// 合约筛选参数类型
export interface ContractFilterParams {
  chain_name?: string
  contract_type?: string
  status?: string
  is_verified?: boolean
}

// 合约搜索参数类型
export interface ContractSearchParams {
  query: string
  chain_name?: string
  contract_type?: string
  status?: string
}
