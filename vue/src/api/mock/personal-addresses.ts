import apiData from '../../../ApiDatas/personal-addresses/personal-addresses-v1.json'

/**
 * 模拟创建个人地址接口
 */
export const handleMockCreatePersonalAddress = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/api/user/addresses'].post.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取个人地址列表接口
 */
export const handleMockGetPersonalAddresses = (): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/api/user/addresses'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取个人地址详情接口
 */
export const handleMockGetPersonalAddressById = (id: number): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/api/user/addresses/{id}'].get.responses['200'].content['application/json'].example
      // 更新ID以匹配请求
      response.data.id = id
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟更新个人地址接口
 */
export const handleMockUpdatePersonalAddress = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/api/user/addresses/{id}'].put.responses['200'].content['application/json'].example
      // 合并更新数据
      response.data = { ...response.data, ...data }
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟删除个人地址接口
 */
export const handleMockDeletePersonalAddress = (): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/api/user/addresses/{id}'].delete.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取地址交易列表接口
 */
export const handleMockGetAddressTransactions = (): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 模拟交易数据
      const mockTransactions = [
        {
          id: 1,
          tx_id: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
          height: 18945678,
          block_index: 0,
          address_from: "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
          address_to: "0x8ba1f109551bD432803012645Hac136c22C177e9",
          amount: "1000000000000000000",
          gas_limit: 21000,
          gas_price: "20000000000",
          gas_used: 21000,
          max_fee_per_gas: "25000000000",
          max_priority_fee_per_gas: "2000000000",
          effective_gas_price: "22000000000",
          fee: "0.000462",
          status: 1,
          confirm: 12,
          chain: "eth",
          symbol: "ETH",
          contract_addr: "",
          ctime: "2023-12-21T10:30:56Z",
          mtime: "2023-12-21T10:30:56Z"
        },
        {
          id: 2,
          tx_id: "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
          height: 18945677,
          block_index: 5,
          address_from: "0x8ba1f109551bD432803012645Hac136c22C177e9",
          address_to: "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
          amount: "500000000000000000",
          gas_limit: 21000,
          gas_price: "18000000000",
          gas_used: 21000,
          max_fee_per_gas: "20000000000",
          max_priority_fee_per_gas: "1500000000",
          effective_gas_price: "19500000000",
          fee: "0.0004095",
          status: 1,
          confirm: 25,
          chain: "eth",
          symbol: "ETH",
          contract_addr: "",
          ctime: "2023-12-21T09:15:30Z",
          mtime: "2023-12-21T09:15:30Z"
        }
      ]

      resolve({
        success: true,
        message: "获取交易列表成功",
        data: {
          transactions: mockTransactions,
          total: 2,
          page: 1,
          page_size: 20,
          has_more: false
        },
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟查询授权关系接口
 */
export const handleMockGetAuthorizedAddresses = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/api/user/addresses/authorized'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取用户所有在途交易地址接口
 */
export const handleMockGetUserAddressesByPending = (chain: string): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 模拟在途交易地址数据
      const mockPendingAddresses = [
        {
          id: 1,
          address: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
          amount: "100000000", // 1 BTC in satoshi
          fee: "510", // 手续费
          status: "in_progress",
          created_at: "2025-01-16 10:30:00",
          updated_at: "2025-01-16 10:30:00"
        },
        {
          id: 2,
          address: "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2",
          amount: "50000000", // 0.5 BTC in satoshi
          fee: "510", // 手续费
          status: "packed",
          created_at: "2025-01-16 09:15:30",
          updated_at: "2025-01-16 09:15:30"
        },
        {
          id: 3,
          address: "1PMycacnJaSqwwJqSqjqjqjqjqjqjqjqjqj",
          amount: "25000000", // 0.25 BTC in satoshi
          fee: "510", // 手续费
          status: "in_progress",
          created_at: "2025-01-16 08:45:15",
          updated_at: "2025-01-16 08:45:15"
        }
      ]

      // 根据链类型过滤数据
      const filteredAddresses = chain === 'btc' ? mockPendingAddresses : []

      resolve({
        success: true,
        message: "获取在途交易地址成功",
        data: filteredAddresses,
        timestamp: Date.now()
      })
    }, 300)
  })
}
