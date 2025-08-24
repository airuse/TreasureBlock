import type { CoinConfigMaintenanceResponse } from '@/types/coinconfig'

/**
 * 模拟获取币种配置维护信息接口
 */
export const handleMockGetCoinConfigMaintenance = (contractAddress: string): Promise<CoinConfigMaintenanceResponse> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 模拟数据
      const mockData: CoinConfigMaintenanceResponse = {
        coin_config: null, // 新合约，没有币种配置
        parser_configs: [], // 新合约，没有解析配置
        contract_address: contractAddress
      }
      
      resolve(mockData)
    }, 300)
  })
}

/**
 * 模拟创建币种配置接口
 */
export const handleMockCreateCoinConfig = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        success: true,
        message: '币种配置创建成功',
        data: {
          id: Math.floor(Math.random() * 1000) + 1,
          ...data,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }
      })
    }, 300)
  })
}
