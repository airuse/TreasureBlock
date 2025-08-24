// 导入所有模块
import * as blocks from './blocks'
import * as transactions from './transactions'
import * as addresses from './addresses'
import * as stats from './stats'
import * as auth from './auth'
import * as user from './user'
import * as coinconfig from './coinconfig'

// 统一导出所有API模块
export {
  blocks,
  transactions,
  addresses,
  stats,
  auth,
  user,
  coinconfig
}

// 默认导出所有模块
export default {
  blocks,
  transactions,
  addresses,
  stats,
  auth,
  user,
  coinconfig
}
