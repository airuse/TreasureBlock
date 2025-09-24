import request from '../request'
import type { ApiResponse } from '../types'
import type { ListSolTxDetailsRequest, SolTxDetailItem, SolInstructionItem, SolEventItem } from '@/types'
import { handleMockListTxDetails, handleMockGetArtifactsByTxId } from '../mock/sol'

// ==================== API函数实现 ====================

/**
 * 分页查询 Sol 交易详情
 */
export function listTxDetails(params: ListSolTxDetailsRequest): Promise<ApiResponse<SolTxDetailItem[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - listTxDetails')
    return handleMockListTxDetails(params)
  }

  return request({
    url: '/api/v1/sol/tx/detail',
    method: 'GET',
    params,
  })
}

/**
 * 通过 txId 查询 artifacts（events + instructions）
 */
export function getArtifactsByTxId(txId: string): Promise<ApiResponse<{ events: SolEventItem[]; instructions: SolInstructionItem[] }>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getArtifactsByTxId')
    return handleMockGetArtifactsByTxId(txId)
  }

  return request({
    url: `/api/v1/sol/tx/${txId}/artifacts`,
    method: 'GET',
  })
}

