import apiData from '../../../ApiDatas/contracts/contracts-v1.json'

/**
 * 模拟获取所有合约接口
 */
export const handleMockGetContracts = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/contracts'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟根据地址获取合约接口
 */
export const handleMockGetContractByAddress = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/contracts/address/{address}'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟根据链名称获取合约接口
 */
export const handleMockGetContractsByChain = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/contracts/chain/{chainName}'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟根据合约类型获取合约接口
 */
export const handleMockGetContractsByType = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/contracts/type/{type}'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟获取所有ERC-20代币合约接口
 */
export const handleMockGetERC20Tokens = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/contracts/erc20'].get.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟创建或更新合约接口
 */
export const handleMockCreateOrUpdateContract = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/contracts'].post.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟更新合约状态接口
 */
export const handleMockUpdateContractStatus = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/contracts/{address}/status/{status}'].put.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟验证合约接口
 */
export const handleMockVerifyContract = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/contracts/{address}/verify'].put.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}

/**
 * 模拟删除合约接口
 */
export const handleMockDeleteContract = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/contracts/{address}'].delete.responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}
