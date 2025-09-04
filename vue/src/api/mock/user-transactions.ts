import type { 
  UserTransaction, 
  UserTransactionListResponse, 
  ExportTransactionResponse,
  UserTransactionStatsResponse 
} from '@/types'

/**
 * 模拟创建用户交易接口
 */
export const handleMockCreateUserTransaction = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const mockResponse: UserTransaction = {
        id: Math.floor(Math.random() * 1000) + 1,
        user_id: 1,
        chain: data.chain,
        symbol: data.symbol,
        from_address: data.from_address,
        to_address: data.to_address,
        amount: data.amount,
        fee: data.fee,
        gas_limit: data.gas_limit,
        gas_price: data.gas_price,
        nonce: data.nonce,
        status: 'draft',
        tx_hash: undefined,
        block_height: undefined,
        confirmations: 0,
        error_msg: undefined,
        remark: data.remark || '',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
      
      resolve({
        success: true,
        message: '创建交易成功',
        data: mockResponse,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取用户交易列表接口
 */
export const handleMockGetUserTransactions = (params: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const mockTransactions: UserTransaction[] = [
        {
          id: 1,
          user_id: 1,
          chain: 'eth',
          symbol: 'ETH',
          from_address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
          to_address: '0x8ba1f109551bD432803012645Hac136c22C177e9',
          amount: '0.1',
          fee: '0.00042',
          gas_limit: 21000,
          gas_price: '20',
          nonce: 0,
          status: 'draft',
          tx_hash: undefined,
          block_height: undefined,
          confirmations: 0,
          error_msg: undefined,
          remark: '测试交易1',
          created_at: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(),
          updated_at: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString()
        },
        {
          id: 2,
          user_id: 1,
          chain: 'eth',
          symbol: 'ETH',
          from_address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
          to_address: '0x1234567890123456789012345678901234567890',
          amount: '0.05',
          fee: '0.00042',
          gas_limit: 21000,
          gas_price: '20',
          nonce: 1,
          status: 'unsigned',
          tx_hash: undefined,
          block_height: undefined,
          confirmations: 0,
          error_msg: undefined,
          remark: '测试交易2',
          created_at: new Date(Date.now() - 1000 * 60 * 60 * 12).toISOString(),
          updated_at: new Date(Date.now() - 1000 * 60 * 60 * 12).toISOString()
        },
        {
          id: 3,
          user_id: 1,
          chain: 'eth',
          symbol: 'ETH',
          from_address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
          to_address: '0x8ba1f109551bD432803012645Hac136c22C177e9',
          amount: '0.2',
          fee: '0.00042',
          gas_limit: 21000,
          gas_price: '20',
          nonce: 2,
          status: 'in_progress',
          tx_hash: undefined,
          block_height: undefined,
          confirmations: 0,
          error_msg: undefined,
          remark: '测试交易3',
          created_at: new Date(Date.now() - 1000 * 60 * 60 * 6).toISOString(),
          updated_at: new Date(Date.now() - 1000 * 60 * 60 * 6).toISOString()
        },
        {
          id: 4,
          user_id: 1,
          chain: 'eth',
          symbol: 'ETH',
          from_address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
          to_address: '0x1234567890123456789012345678901234567890',
          amount: '0.15',
          fee: '0.00042',
          gas_limit: 21000,
          gas_price: '20',
          nonce: 3,
          status: 'in_progress',
          tx_hash: '0xabcdef1234567890abcdef1234567890abcdef12',
          block_height: undefined,
          confirmations: 0,
          error_msg: undefined,
          remark: '测试交易4',
          created_at: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
          updated_at: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString()
        },
        {
          id: 5,
          user_id: 1,
          chain: 'eth',
          symbol: 'ETH',
          from_address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
          to_address: '0x8ba1f109551bD432803012645Hac136c22C177e9',
          amount: '0.3',
          fee: '0.00042',
          gas_limit: 21000,
          gas_price: '20',
          nonce: 4,
          status: 'confirmed',
          tx_hash: '0x567890abcdef1234567890abcdef1234567890ab',
          block_height: 12345678,
          confirmations: 12,
          error_msg: undefined,
          remark: '测试交易5',
          created_at: new Date(Date.now() - 1000 * 60 * 60 * 24 * 3).toISOString(),
          updated_at: new Date(Date.now() - 1000 * 60 * 60 * 24 * 3).toISOString()
        }
      ]

      // 根据状态筛选
      let filteredTransactions = mockTransactions
      if (params.status) {
        filteredTransactions = mockTransactions.filter(tx => tx.status === params.status)
      }

      // 分页处理
      const page = params.page || 1
      const pageSize = params.page_size || 10
      const start = (page - 1) * pageSize
      const end = start + pageSize
      const paginatedTransactions = filteredTransactions.slice(start, end)

      const mockResponse: UserTransactionListResponse = {
        transactions: paginatedTransactions,
        total: filteredTransactions.length,
        page: page,
        page_size: pageSize,
        has_more: end < filteredTransactions.length
      }

      resolve({
        success: true,
        message: '获取交易列表成功',
        data: mockResponse,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取用户交易统计接口
 */
export const handleMockGetUserTransactionStats = (): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const mockResponse: UserTransactionStatsResponse = {
        total_transactions: 25,
        draft_count: 2,
        unsigned_count: 3,
        in_progress_count: 1,
        packed_count: 1,
        confirmed_count: 16,
        failed_count: 0
      }

      resolve({
        success: true,
        message: '获取统计信息成功',
        data: mockResponse,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟根据ID获取用户交易接口
 */
export const handleMockGetUserTransactionById = (id: number): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const mockTransaction: UserTransaction = {
        id: id,
        user_id: 1,
        chain: 'eth',
        symbol: 'ETH',
        from_address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
        to_address: '0x8ba1f109551bD432803012645Hac136c22C177e9',
        amount: '0.1',
        fee: '0.00042',
        gas_limit: 21000,
        gas_price: '20',
        nonce: 0,
        status: 'draft',
        tx_hash: undefined,
        block_height: undefined,
        confirmations: 0,
        error_msg: undefined,
        remark: '测试交易',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }

      resolve({
        success: true,
        message: '获取交易成功',
        data: mockTransaction,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟更新用户交易接口
 */
export const handleMockUpdateUserTransaction = (id: number, data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const mockTransaction: UserTransaction = {
        id: id,
        user_id: 1,
        chain: 'eth',
        symbol: 'ETH',
        from_address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
        to_address: '0x8ba1f109551bD432803012645Hac136c22C177e9',
        amount: '0.1',
        fee: '0.00042',
        gas_limit: 21000,
        gas_price: '20',
        nonce: 0,
        status: data.status || 'draft',
        tx_hash: data.tx_hash,
        block_height: data.block_height,
        confirmations: data.confirmations || 0,
        error_msg: data.error_msg,
        remark: data.remark || '测试交易',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }

      resolve({
        success: true,
        message: '更新交易成功',
        data: mockTransaction,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟删除用户交易接口
 */
export const handleMockDeleteUserTransaction = (id: number): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        success: true,
        message: '删除交易成功',
        data: null,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟导出交易接口
 */
export const handleMockExportTransaction = (id: number): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const mockResponse: ExportTransactionResponse = {
        unsigned_tx: `{
          "chain": "eth",
          "symbol": "ETH",
          "from": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
          "to": "0x8ba1f109551bD432803012645Hac136c22C177e9",
          "amount": "0.1",
          "fee": "0.00042",
          "gasLimit": 21000,
          "gasPrice": "20",
          "nonce": 0
        }`,
        chain: 'eth',
        symbol: 'ETH',
        from_address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
        to_address: '0x8ba1f109551bD432803012645Hac136c22C177e9',
        amount: '0.1',
        fee: '0.00042',
        gas_limit: 21000,
        gas_price: '20',
        nonce: 0
      }

      resolve({
        success: true,
        message: '导出交易成功',
        data: mockResponse,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟导入签名接口
 */
export const handleMockImportSignature = (id: number, data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const mockTransaction: UserTransaction = {
        id: id,
        user_id: 1,
        chain: 'eth',
        symbol: 'ETH',
        from_address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
        to_address: '0x8ba1f109551bD432803012645Hac136c22C177e9',
        amount: '0.1',
        fee: '0.00042',
        gas_limit: 21000,
        gas_price: '20',
        nonce: 0,
        status: 'in_progress',
        tx_hash: undefined,
        block_height: undefined,
        confirmations: 0,
        error_msg: undefined,
        remark: '测试交易',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }

      resolve({
        success: true,
        message: '导入签名成功',
        data: mockTransaction,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟发送交易接口
 */
export const handleMockSendTransaction = (id: number): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const mockTransaction: UserTransaction = {
        id: id,
        user_id: 1,
        chain: 'eth',
        symbol: 'ETH',
        from_address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
        to_address: '0x8ba1f109551bD432803012645Hac136c22C177e9',
        amount: '0.1',
        fee: '0.00042',
        gas_limit: 21000,
        gas_price: '20',
        nonce: 0,
        status: 'in_progress',
        tx_hash: '0xabcdef1234567890abcdef1234567890abcdef12',
        block_height: undefined,
        confirmations: 0,
        error_msg: undefined,
        remark: '测试交易',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }

      resolve({
        success: true,
        message: '发送交易成功',
        data: mockTransaction,
        timestamp: Date.now()
      })
    }, 300)
  })
}
