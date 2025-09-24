import { createRouter, createWebHistory } from 'vue-router'

// ETH页面
import ETHHomeView from '../views/eth/HomeView.vue'
import ETHBlocksView from '../views/eth/BlocksView.vue'
import ETHBlockDetailView from '../views/eth/detail/BlockDetailView.vue'
import ETHContractDetailView from '../views/eth/detail/ContractDetailView.vue'
import ETHAddressesView from '../views/eth/AddressesView.vue'
import ETHStatisticsView from '../views/eth/StatisticsView.vue'
import ETHSettingsView from '../views/eth/SettingsView.vue'

// BTC页面
import BTCHomeView from '../views/btc/HomeView.vue'
import BTCBlocksView from '../views/btc/BlocksView.vue'
import BTCStatsView from '../views/btc/StatsView.vue'
import BTCBlockDetailView from '../views/btc/detail/BlockDetailView.vue'

// BSC页面
import BSCHomeView from '../views/bsc/HomeView.vue'
import BSCBlocksView from '../views/bsc/BlocksView.vue'
import BSCBlockDetailView from '../views/bsc/detail/BlockDetailView.vue'
import BSCContractDetailView from '../views/bsc/detail/ContractDetailView.vue'
import BSCAddressesView from '../views/bsc/AddressesView.vue'
import BSCStatisticsView from '../views/bsc/StatisticsView.vue'
import BSCSettingsView from '../views/bsc/SettingsView.vue'

// SOL页面
import SOLHomeView from '../views/sol/HomeView.vue'
import SOLBlocksView from '../views/sol/BlocksView.vue'
import SOLBlockDetailView from '../views/sol/detail/BlockDetailView.vue'
import SOLContractDetailView from '../views/sol/detail/ContractDetailView.vue'
import SOLAddressesView from '../views/sol/AddressesView.vue'
import SOLProgramView from '../views/sol/ProgramView.vue'
import SOLStatisticsView from '../views/sol/StatisticsView.vue'
import SOLSettingsView from '../views/sol/SettingsView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // 默认重定向到ETH
    {
      path: '/',
      redirect: '/eth'
    },
    // ETH路由
    {
      path: '/eth',
      name: 'eth-home',
      component: ETHHomeView
    },
    {
      path: '/eth/blocks',
      name: 'eth-blocks',
      component: ETHBlocksView
    },
    {
      path: '/eth/blocks/:height',
      name: 'eth-block-detail',
      component: ETHBlockDetailView
    },
    {
      path: '/eth/addresses',
      name: 'eth-addresses',
      component: ETHAddressesView
    },
    {
      path: '/eth/address-transactions',
      name: 'eth-address-transactions',
      component: () => import('../views/eth/detail/AddressTransactionsView.vue')
    },
    {
      path: '/eth/addresses/:address',
      name: 'eth-contract-detail',
      component: ETHContractDetailView
    },
    {
      path: '/eth/statistics',
      name: 'eth-statistics',
      component: ETHStatisticsView
    },
    {
      path: '/eth/settings',
      name: 'eth-settings',
      component: ETHSettingsView
    },
    // 个人中心 - ETH
    {
      path: '/eth/personal',
      name: 'eth-personal',
      redirect: '/eth/personal/earnings'
    },
    {
      path: '/eth/personal/earnings',
      name: 'eth-personal-earnings',
      component: () => import('../views/eth/personal/EarningsView.vue')
    },
    {
      path: '/eth/personal/addresses',
      name: 'eth-personal-addresses',
      component: () => import('../views/eth/personal/AddressesView.vue')
    },
    {
      path: '/eth/personal/transactions',
      name: 'eth-personal-transactions',
      component: () => import('../views/eth/personal/TransactionsView.vue')
    },
    // BTC路由
    {
      path: '/btc',
      name: 'btc-home',
      component: BTCHomeView
    },
    {
      path: '/btc/blocks',
      name: 'btc-blocks',
      component: BTCBlocksView
    },
    {
      path: '/btc/blocks/:height',
      name: 'btc-block-detail',
      component: BTCBlockDetailView
    },
    {
      path: '/btc/address-transactions',
      name: 'btc-address-transactions',
      component: () => import('../views/btc/detail/AddressTransactionsView.vue')
    },
    {
      path: '/btc/statistics',
      name: 'btc-statistics',
      component: BTCStatsView
    },
    // 个人中心 - BTC
    {
      path: '/btc/personal',
      name: 'btc-personal',
      redirect: '/btc/personal/earnings'
    },
    {
      path: '/btc/personal/earnings',
      name: 'btc-personal-earnings',
      component: () => import('../views/btc/personal/EarningsView.vue')
    },
    {
      path: '/btc/personal/addresses',
      name: 'btc-personal-addresses',
      component: () => import('../views/btc/personal/AddressesView.vue')
    },
    {
      path: '/btc/personal/transactions',
      name: 'btc-personal-transactions',
      component: () => import('../views/btc/personal/TransactionsView.vue')
    },
    // BSC路由
    {
      path: '/bsc',
      name: 'bsc-home',
      component: BSCHomeView
    },
    {
      path: '/bsc/blocks',
      name: 'bsc-blocks',
      component: BSCBlocksView
    },
    {
      path: '/bsc/blocks/:height',
      name: 'bsc-block-detail',
      component: BSCBlockDetailView
    },
    {
      path: '/bsc/addresses',
      name: 'bsc-addresses',
      component: BSCAddressesView
    },
    {
      path: '/bsc/address-transactions',
      name: 'bsc-address-transactions',
      component: () => import('../views/bsc/detail/AddressTransactionsView.vue')
    },
    {
      path: '/bsc/addresses/:address',
      name: 'bsc-contract-detail',
      component: BSCContractDetailView
    },
    {
      path: '/bsc/statistics',
      name: 'bsc-statistics',
      component: BSCStatisticsView
    },
    {
      path: '/bsc/settings',
      name: 'bsc-settings',
      component: BSCSettingsView
    },
    // 个人中心 - BSC
    {
      path: '/bsc/personal',
      name: 'bsc-personal',
      redirect: '/bsc/personal/earnings'
    },
    {
      path: '/bsc/personal/earnings',
      name: 'bsc-personal-earnings',
      component: () => import('../views/bsc/personal/EarningsView.vue')
    },
    {
      path: '/bsc/personal/addresses',
      name: 'bsc-personal-addresses',
      component: () => import('../views/bsc/personal/AddressesView.vue')
    },
    {
      path: '/bsc/personal/transactions',
      name: 'bsc-personal-transactions',
      component: () => import('../views/bsc/personal/TransactionsView.vue')
    },
    // SOL路由
    {
      path: '/sol',
      name: 'sol-home',
      component: SOLHomeView
    },
    {
      path: '/sol/blocks',
      name: 'sol-blocks',
      component: SOLBlocksView
    },
    {
      path: '/sol/blocks/:height',
      name: 'sol-block-detail',
      component: SOLBlockDetailView
    },
    {
      path: '/sol/programs',
      name: 'sol-programs',
      component: SOLProgramView
    },
    {
      path: '/sol/addresses',
      name: 'sol-addresses',
      component: SOLAddressesView
    },
    {
      path: '/sol/address-transactions',
      name: 'sol-address-transactions',
      component: () => import('../views/sol/detail/AddressTransactionsView.vue')
    },
    // 保留合约详情但入口改到 programs 页，这里保持兼容
    {
      path: '/sol/addresses/:address',
      name: 'sol-contract-detail',
      component: SOLContractDetailView
    },
    {
      path: '/sol/statistics',
      name: 'sol-statistics',
      component: SOLStatisticsView
    },
    {
      path: '/sol/settings',
      name: 'sol-settings',
      component: SOLSettingsView
    },
    // 个人中心 - SOL
    {
      path: '/sol/personal',
      name: 'sol-personal',
      redirect: '/sol/personal/earnings'
    },
    {
      path: '/sol/personal/earnings',
      name: 'sol-personal-earnings',
      component: () => import('../views/sol/personal/EarningsView.vue')
    },
    {
      path: '/sol/personal/addresses',
      name: 'sol-personal-addresses',
      component: () => import('../views/sol/personal/AddressesView.vue')
    },
    {
      path: '/sol/personal/transactions',
      name: 'sol-personal-transactions',
      component: () => import('../views/sol/personal/TransactionsView.vue')
    }
  ]
})

// 全局前置守卫 - 检查路由有效性
router.beforeEach((to, from, next) => {
  const path = to.path
  
  // 检查是否是BTC的无效路由
  if (path.startsWith('/btc/addresses') || path.startsWith('/btc/settings') || path.startsWith('/btc/statistics')) {
    console.log(`检测到无效的BTC路由: ${path}，重定向到BTC首页`)
    next('/btc')
    return
  }
  
  // 检查是否是ETH的无效路由
  if (path.startsWith('/eth/settings') && !path.includes('/personal/')) {
    // 检查settings页面是否存在（这里可以根据需要调整）
    next()
    return
  }
  
  // 检查是否是ETH的统计页面（暂时屏蔽）
  if (path.startsWith('/eth/statistics')) {
    console.log(`检测到暂时屏蔽的ETH统计路由: ${path}，重定向到ETH首页`)
    next('/eth')
    return
  }
  
  // 检查是否是BSC的无效路由
  if (path.startsWith('/bsc/settings') && !path.includes('/personal/')) {
    // 检查settings页面是否存在（这里可以根据需要调整）
    next()
    return
  }
  
  // 检查是否是BSC的统计页面（暂时屏蔽）
  if (path.startsWith('/bsc/statistics')) {
    console.log(`检测到暂时屏蔽的BSC统计路由: ${path}，重定向到BSC首页`)
    next('/bsc')
    return
  }
  
  // 检查是否是SOL的无效路由
  if (path.startsWith('/sol/settings') && !path.includes('/personal/')) {
    // 检查settings页面是否存在（这里可以根据需要调整）
    next()
    return
  }
  
  // 检查是否是SOL的统计页面（暂时屏蔽）
  if (path.startsWith('/sol/statistics')) {
    console.log(`检测到暂时屏蔽的SOL统计路由: ${path}，重定向到SOL首页`)
    next('/sol')
    return
  }
  
  // 其他路由正常通过
  next()
})

export default router
