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
import BTCAddressView from '../views/btc/AddressView.vue'
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
      path: '/btc/addresses',
      name: 'btc-addresses',
      component: BTCAddressView
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

export default router
