import apiData from '../../../ApiDatas/home/home-v1.json'

/**
 * 模拟获取首页统计数据接口
 */
export const handleMockGetHomeStats = (chain: string): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档数据中获取示例数据
      const exampleData = apiData.paths['/api/v1/home/stats'].get.responses['200'].content['application/json'].example
      
      // 根据链类型过滤数据
      const filteredData = {
        ...exampleData.data,
        latestBlocks: exampleData.data.latestBlocks.filter((block: any) => block.chain === chain),
        latestTransactions: exampleData.data.latestTransactions.filter((tx: any) => tx.chain === chain)
      }
      
      // 构建正确的响应结构
      const response = {
        success: true,
        data: filteredData,
        message: '成功获取首页统计数据',
        timestamp: Date.now()
      }
      
      console.log('🔧 Mock数据响应:', response)
      resolve(response)
    }, 300)
  })
}
