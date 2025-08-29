// 导入接口响应数据
import apiData from '../../../ApiDatas/transactions/transactions-v1.json'

/**
 * 模拟获取交易列表接口
 */
export const handleMockGetTransactions = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/transactions'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取交易详情接口
 */
export const handleMockGetTransaction = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/transactions/{hash}'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 200)
  })
}

/**
 * 模拟搜索交易接口
 */
export const handleMockSearchTransactions = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/transactions/search'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 400)
  })
}

/**
 * 模拟获取交易解析结果接口
 */
export const handleMockGetParsedTransaction = (hash: string): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 参考 handleMockSearchTransactions，从 Apifox 导出的 JSON 读取示例
      try {
        const response = (apiData as any).paths['/transactions/parsed/{hash}'].get.responses['200'].content['application/json'].example
        resolve({ success: true, data: response })
      } catch (e) {
        // 兼容数据缺失时的最小返回
        resolve({ success: true, data: [] })
      }
    }, 200)
  })
}
