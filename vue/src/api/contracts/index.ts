import request from '../request'
import { 
  handleMockGetContracts,
  handleMockGetContractByAddress,
  handleMockGetContractsByChain,
  handleMockGetContractsByType,
  handleMockGetERC20Tokens,
  handleMockCreateOrUpdateContract,
  handleMockUpdateContractStatus,
  handleMockVerifyContract,
  handleMockDeleteContract
} from '../mock/contracts'
import type { Contract } from '@/types'
import type { ApiResponse, PaginatedResponse, PaginationRequest, SortRequest, SearchRequest } from '../types'

// ==================== API相关类型定义 ====================

// 获取合约请求参数类型 - 继承通用类型
interface GetContractsRequest extends PaginationRequest, SortRequest {
  chainName?: string
  contractType?: string
  status?: string
  search?: string
}

// 根据链名称获取合约请求参数类型
interface GetContractsByChainRequest extends PaginationRequest, SortRequest {
  contractType?: string
  status?: string
}

// 根据合约类型获取合约请求参数类型
interface GetContractsByTypeRequest extends PaginationRequest, SortRequest {
  chainName?: string
  status?: string
}

// 获取ERC-20代币请求参数类型
interface GetERC20TokensRequest extends PaginationRequest, SortRequest {
  chainName?: string
  status?: string
}

// ==================== API函数实现 ====================

/**
 * 获取所有合约
 */
export function getContracts(data: GetContractsRequest): Promise<PaginatedResponse<Contract>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getContracts')
    return handleMockGetContracts(data)
  }
  
  console.log('🌐 使用真实API - getContracts')
  return request({
    url: '/api/v1/contracts',
    method: 'GET',
    params: data
  })
}

/**
 * 根据地址获取合约
 */
export function getContractByAddress(address: string): Promise<ApiResponse<Contract>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getContractByAddress')
    return handleMockGetContractByAddress({ address })
  }
  
  console.log('🌐 使用真实API - getContractByAddress')
  return request({
    url: `/api/v1/contracts/address/${address}`,
    method: 'GET'
  })
}

/**
 * 根据链名称获取合约
 */
export function getContractsByChain(chainName: string, data: GetContractsByChainRequest): Promise<PaginatedResponse<Contract>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getContractsByChain')
    return handleMockGetContractsByChain({ chainName, ...data })
  }
  
  console.log('🌐 使用真实API - getContractsByChain')
  return request({
    url: `/api/v1/contracts/chain/${chainName}`,
    method: 'GET',
    params: data
  })
}

/**
 * 根据合约类型获取合约
 */
export function getContractsByType(type: string, data: GetContractsByTypeRequest): Promise<PaginatedResponse<Contract>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getContractsByType')
    return handleMockGetContractsByType({ type, ...data })
  }
  
  console.log('🌐 使用真实API - getContractsByType')
  return request({
    url: `/api/v1/contracts/type/${type}`,
    method: 'GET',
    params: data
  })
}

/**
 * 获取所有ERC-20代币合约
 */
export function getERC20Tokens(data: GetERC20TokensRequest): Promise<PaginatedResponse<Contract>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getERC20Tokens')
    return handleMockGetERC20Tokens(data)
  }
  
  console.log('🌐 使用真实API - getERC20Tokens')
  return request({
    url: '/api/v1/contracts/erc20',
    method: 'GET',
    params: data
  })
}

/**
 * 创建或更新合约
 */
export function createOrUpdateContract(data: Partial<Contract>): Promise<ApiResponse<Contract>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - createOrUpdateContract')
    return handleMockCreateOrUpdateContract(data)
  }
  
  console.log('🌐 使用真实API - createOrUpdateContract')
  return request({
    url: '/api/v1/contracts',
    method: 'POST',
    data
  })
}

/**
 * 更新合约状态
 */
export function updateContractStatus(address: string, status: string): Promise<ApiResponse<Contract>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - updateContractStatus')
    return handleMockUpdateContractStatus({ address, status })
  }
  
  console.log('🌐 使用真实API - updateContractStatus')
  return request({
    url: `/api/v1/contracts/${address}/status/${status}`,
    method: 'PUT'
  })
}

/**
 * 验证合约
 */
export function verifyContract(address: string): Promise<ApiResponse<Contract>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - verifyContract')
    return handleMockVerifyContract({ address })
  }
  
  console.log('🌐 使用真实API - verifyContract')
  return request({
    url: `/api/v1/contracts/${address}/verify`,
    method: 'PUT'
  })
}

/**
 * 删除合约
 */
export function deleteContract(address: string): Promise<ApiResponse<void>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - deleteContract')
    return handleMockDeleteContract({ address })
  }
  
  console.log('🌐 使用真实API - deleteContract')
  return request({
    url: `/api/v1/contracts/${address}`,
    method: 'DELETE'
  })
}
