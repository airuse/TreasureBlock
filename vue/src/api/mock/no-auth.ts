import apiData from '../../../ApiDatas/no-auth/no-auth-v1.json'

// ==================== Mock数据处理函数 ====================

/**
 * 模拟获取区块列表（游客模式）
 */
export const handleMockNoAuthGetBlocks = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/api/no-auth/blocks'].get.responses['200'].content['application/json'].example
      
      // 根据page_size限制数据量
      const maxSize = Math.min(data.page_size || 25, 100)
      const limitedData = response.data.slice(0, maxSize)
      
      resolve({
        ...response,
        data: limitedData,
        pagination: {
          ...response.pagination,
          page_size: maxSize,
          total: Math.min(response.pagination.total, 100)
        },
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟搜索区块（游客模式）
 */
export const handleMockNoAuthSearchBlocks = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/api/no-auth/blocks/search'].get.responses['200'].content['application/json'].example
      
      // 根据page_size限制数据量，最多20个
      const maxSize = Math.min(data.page_size || 20, 20)
      const limitedData = response.data.slice(0, maxSize)
      
      resolve({
        ...response,
        data: limitedData,
        pagination: {
          ...response.pagination,
          page_size: maxSize,
          total: Math.min(response.pagination.total, 20)
        },
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取首页统计数据（游客模式）
 */
export const handleMockNoAuthGetHomeStats = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/api/no-auth/home/stats'].get.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取合约列表（游客模式）
 */
export const handleMockNoAuthGetContracts = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/api/no-auth/contracts'].get.responses['200'].content['application/json'].example
      
      // 根据page_size限制数据量，最多100个
      const maxSize = Math.min(data.page_size || 25, 100)
      const limitedData = response.data.slice(0, maxSize)
      
      resolve({
        ...response,
        data: limitedData,
        count: Math.min(response.count, 100),
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取区块详情（游客模式）
 */
export const handleMockNoAuthGetBlockByHeight = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 模拟区块详情数据
      const blockData = {
        height: data.height,
        hash: `0x${Math.random().toString(16).substr(2, 64)}`,
        timestamp: Math.floor(Date.now() / 1000),
        transactions: Math.floor(Math.random() * 200) + 50,
        size: Math.floor(Math.random() * 200000) + 50000,
        gasUsed: Math.floor(Math.random() * 30000000) + 15000000,
        gasLimit: 45000000,
        miner: `0x${Math.random().toString(16).substr(2, 40)}`,
        miner_tip_eth: (Math.random() * 0.02 + 0.01).toFixed(6)
      }
      
      resolve({
        success: true,
        data: blockData,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取交易列表（游客模式）
 */
export const handleMockNoAuthGetTransactions = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 模拟交易数据
      const transactions = Array.from({ length: Math.min(data.page_size || 25, 1000) }, (_, i) => ({
        hash: `0x${Math.random().toString(16).substr(2, 64)}`,
        timestamp: Math.floor(Date.now() / 1000) - i * 12,
        amount: (Math.random() * 10).toFixed(6),
        from: `0x${Math.random().toString(16).substr(2, 40)}`,
        to: `0x${Math.random().toString(16).substr(2, 40)}`,
        gas_price: Math.floor(Math.random() * 100000000000) + 20000000000,
        gas_used: Math.floor(Math.random() * 21000) + 21000,
        height: Math.floor(Math.random() * 1000) + 23203000
      }))
      
      resolve({
        success: true,
        data: transactions,
        pagination: {
          page: data.page || 1,
          page_size: Math.min(data.page_size || 25, 1000),
          total: Math.min(1000, transactions.length)
        },
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟根据区块高度获取交易列表（游客模式）
 */
export const handleMockNoAuthGetTransactionsByBlockHeight = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 模拟区块交易数据
      const transactions = Array.from({ length: Math.min(data.page_size || 25, 50) }, (_, i) => ({
        hash: `0x${Math.random().toString(16).substr(2, 64)}`,
        timestamp: Math.floor(Date.now() / 1000),
        amount: (Math.random() * 10).toFixed(6),
        from: `0x${Math.random().toString(16).substr(2, 40)}`,
        to: `0x${Math.random().toString(16).substr(2, 40)}`,
        gas_price: Math.floor(Math.random() * 100000000000) + 20000000000,
        gas_used: Math.floor(Math.random() * 21000) + 21000,
        height: data.blockHeight
      }))
      
      resolve({
        success: true,
        data: transactions,
        timestamp: Date.now()
      })
    }, 300)
  })
}
