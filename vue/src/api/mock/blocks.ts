// 导入接口响应数据
import apiData from '../../../ApiDatas/blocks/blocks-v1.json'

/**
 * 模拟获取区块列表接口
 */
export const handleMockGetBlocks = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/blocks'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取区块详情接口
 */
export const handleMockGetBlock = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/blocks/{height}'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 200)
  })
}

/**
 * 模拟搜索区块接口
 */
export const handleMockSearchBlocks = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/blocks/search'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 400)
  })
}
