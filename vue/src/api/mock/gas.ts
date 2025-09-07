// 导入接口响应数据
import apiData from '../../../ApiDatas/gas/gas-v1.json'

/**
 * 模拟获取Gas费率接口
 */
export const handleMockGetGasRates = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/api/user/gas'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取所有链Gas费率接口
 */
export const handleMockGetAllGasRates = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/api/user/gas/all'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 400)
  })
}
