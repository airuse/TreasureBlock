// 导入接口响应数据
import apiData from '../../../ApiDatas/stats/stats-v1.json'

// 请求参数类型
interface GetNetworkStatsRequest {
  chain?: string
  period?: string
}

interface GetLatestBlocksRequest {
  limit?: number
  chain?: string
}

interface GetLatestTransactionsRequest {
  limit?: number
  chain?: string
}

// 响应类型 - 使用新的格式
interface ApiResponse<T> {
  success: boolean
  message?: string
  data: T
  error?: string
}

/**
 * 模拟获取网络统计接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockGetNetworkStats = (data: GetNetworkStatsRequest): Promise<ApiResponse<any>> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/stats/network'].get.responses['200'].content['application/json'].example
      
      resolve({
        success: true,
        data: response,
        message: 'Success'
      })
    }, 300) // 模拟网络延迟
  })
}

/**
 * 模拟获取最新区块接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockGetLatestBlocks = (data: GetLatestBlocksRequest): Promise<ApiResponse<any[]>> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/stats/latest-blocks'].get.responses['200'].content['application/json'].example
      
      resolve({
        success: true,
        data: response.data || response,
        message: 'Success'
      })
    }, 200) // 模拟网络延迟
  })
}

/**
 * 模拟获取最新交易接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockGetLatestTransactions = (data: GetLatestTransactionsRequest): Promise<ApiResponse<any[]>> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/stats/latest-transactions'].get.responses['200'].content['application/json'].example
      
      resolve({
        success: true,
        data: response.data || response,
        message: 'Success'
      })
    }, 250) // 模拟网络延迟
  })
}
