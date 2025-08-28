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
  
  // 其他路由正常通过
  next()
})

export default router
