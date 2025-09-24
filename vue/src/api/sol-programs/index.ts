import request from '../request'
import type { ApiResponse, PaginatedResponse } from '../types'
import type { SolProgram, ListProgramsRequest } from '@/types'

// 重新导出类型
export type { SolProgram, ListProgramsRequest }
import { 
  handleMockCreateProgram,
  handleMockUpdateProgram,
  handleMockDeleteProgram,
  handleMockGetProgram,
  handleMockListPrograms
} from '../mock/sol-programs'

export function createProgram(data: SolProgram): Promise<ApiResponse<SolProgram>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - createProgram')
    return handleMockCreateProgram(data)
  }
  return request({ url: '/api/v1/sol/programs', method: 'POST', data })
}

export function updateProgram(id: number, data: Partial<SolProgram>): Promise<ApiResponse<SolProgram>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - updateProgram')
    return handleMockUpdateProgram(id, data)
  }
  return request({ url: `/api/v1/sol/programs/${id}`, method: 'PUT', data })
}

export function deleteProgram(id: number): Promise<ApiResponse<null>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - deleteProgram')
    return handleMockDeleteProgram(id)
  }
  return request({ url: `/api/v1/sol/programs/${id}`, method: 'DELETE' })
}

export function getProgram(id: number): Promise<ApiResponse<SolProgram>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getProgram')
    return handleMockGetProgram(id)
  }
  return request({ url: `/api/v1/sol/programs/${id}`, method: 'GET' })
}

export function listPrograms(params: ListProgramsRequest): Promise<PaginatedResponse<SolProgram[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - listPrograms')
    return handleMockListPrograms(params)
  }
  return request({ url: '/api/v1/sol/programs', method: 'GET', params })
}

