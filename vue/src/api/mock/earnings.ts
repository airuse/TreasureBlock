import apiData from '../../../ApiDatas/earnings/earnings-v1.json'

/**
 * 模拟获取用户余额接口
 */
export const handleMockGetUserBalance = (): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = (apiData as any).paths['/earnings/balance'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取收益记录列表接口
 */
export const handleMockGetUserEarningsRecords = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = (apiData as any).paths['/earnings/records'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取收益记录详情接口
 */
export const handleMockGetEarningsRecordDetail = (id: number): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = (apiData as any).paths['/earnings/records/{id}'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取收益统计接口
 */
export const handleMockGetUserEarningsStats = (): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = (apiData as any).paths['/earnings/stats'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟转账T币接口
 */
export const handleMockTransferTCoins = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = (apiData as any).paths['/earnings/transfer'].post.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取收益趋势接口
 */
export const handleMockGetEarningsTrend = (hours: number = 2): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 生成模拟的趋势数据
      const now = new Date()
      const trendData = []
      
      for (let i = hours - 1; i >= 0; i--) {
        const time = new Date(now.getTime() - i * 60 * 60 * 1000)
        const hour = time.getHours()
        const minute = time.getMinutes()
        
        trendData.push({
          timestamp: `${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`,
          amount: Math.floor(Math.random() * 200) + 50, // 50-250之间的随机数
          block_height: 23201390 + i,
          transaction_count: Math.floor(Math.random() * 150) + 100, // 100-250之间的随机数
          source_chain: 'eth'
        })
      }
      
      resolve({
        success: true,
        message: '获取收益趋势成功',
        data: trendData,
        timestamp: Date.now()
      })
    }, 300)
  })
}
