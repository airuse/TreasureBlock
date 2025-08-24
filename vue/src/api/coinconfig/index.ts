import request from '../request'
import { handleMockGetCoinConfigMaintenance, handleMockCreateCoinConfig } from '../mock/coinconfig'
import type { CoinConfig, CoinConfigMaintenanceResponse } from '@/types/coinconfig'
import type { ApiResponse, PaginatedResponse, PaginationRequest } from '../types'

// ==================== API相关类型定义 ====================

// 创建币种配置请求类型
interface CreateCoinConfigRequest {
  contract_addr: string
  chain_name: string
  symbol: string
  coin_type: number
  precision: number
  decimals: number
  name: string
  logo_url: string
  is_verified: boolean
  status: number
}

// 币种配置列表请求类型
interface ListCoinConfigsRequest extends PaginationRequest {
  chain?: string
  symbol?: string
  status?: number
}

// ==================== API函数实现 ====================

/**
 * 获取币种配置维护信息（包含解析配置）
 */
export function getCoinConfigMaintenance(contractAddress: string): Promise<ApiResponse<CoinConfigMaintenanceResponse>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getCoinConfigMaintenance')
    return handleMockGetCoinConfigMaintenance(contractAddress).then(data => ({
      success: true,
      data
    }))
  }
  
  console.log('🌐 使用真实API - getCoinConfigMaintenance')
  // 添加时间戳参数避免浏览器缓存，确保每次都是新请求
  const timestamp = Date.now()
  return request({
    url: `/api/v1/coin-configs/maintenance/${contractAddress}`,
    method: 'GET',
    params: {
      _t: timestamp // 时间戳参数，避免缓存
    }
  })
}

/**
 * 创建币种配置
 */
export function createCoinConfig(data: CreateCoinConfigRequest): Promise<ApiResponse<CoinConfig>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - createCoinConfig')
    return handleMockCreateCoinConfig(data)
  }
  
  console.log('🌐 使用真实API - createCoinConfig')
  return request({
    url: '/api/v1/coin-configs',
    method: 'POST',
    data
  })
}

/**
 * 获取币种配置列表
 */
export function listCoinConfigs(params: ListCoinConfigsRequest): Promise<PaginatedResponse<CoinConfig>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - listCoinConfigs')
    return Promise.resolve({
      success: true,
      data: [],
      pagination: {
        total: 0,
        page: 1,
        page_size: 20,
        total_pages: 0
      }
    })
  }
  
  console.log('🌐 使用真实API - listCoinConfigs')
  return request({
    url: '/api/v1/coin-configs',
    method: 'GET',
    params
  })
}

/**
 * 根据合约地址获取币种配置
 */
export function getCoinConfigByContractAddress(contractAddress: string): Promise<ApiResponse<CoinConfig | null>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getCoinConfigByContractAddress')
    return Promise.resolve({
      success: true,
      data: null
    })
  }
  
  console.log('🌐 使用真实API - getCoinConfigByContractAddress')
  return request({
    url: `/api/v1/coin-configs/contract/${contractAddress}`,
    method: 'GET'
  })
}
