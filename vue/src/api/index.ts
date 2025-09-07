// 导入所有模块
import * as blocks from './blocks'
import * as transactions from './transactions'
import * as stats from './stats'
import * as auth from './auth'
import * as user from './user'
import * as coinconfig from './coinconfig'
import * as contracts from './contracts'
import * as parserConfigs from './parser-configs'
import * as home from './home'
import * as earnings from './earnings'
import * as personalAddresses from './personal-addresses'
import * as noAuth from './no-auth'
import * as userTransactions from './user-transactions'
import * as gas from './gas'

// 统一导出所有API模块
export {
  blocks,
  transactions,
  stats,
  auth,
  user,
  coinconfig,
  contracts,
  parserConfigs,
  home,
  earnings,
  personalAddresses,
  noAuth,
  userTransactions,
  gas
}

// 默认导出所有模块
export default {
  blocks,
  transactions,
  stats,
  auth,
  user,
  coinconfig,
  contracts,
  parserConfigs,
  home,
  earnings,
  personalAddresses,
  noAuth,
  userTransactions,
  gas
}
