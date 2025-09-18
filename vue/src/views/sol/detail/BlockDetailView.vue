<template>
  <div class="space-y-6">
    <!-- 页面标题和返回按钮 -->
    <div class="flex items-center space-x-4">
      <router-link 
        to="/sol/blocks" 
        class="inline-flex items-center px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
        </svg>
        返回区块列表
      </router-link>
      <h1 class="text-2xl font-bold text-gray-900">区块详情 #{{ blockHeight }}</h1>
    </div>

    <!-- 全局轻提示：复制成功（跟随点击位置） -->
    <div v-if="showToast" class="fixed z-50 bg-gray-900 text-white text-sm px-3 py-2 rounded shadow pointer-events-none" :style="toastStyle">
      {{ toastMessage || '已复制到剪贴板' }}
    </div>

    <!-- 加载状态 -->
    <div v-if="isLoading" class="card">
      <div class="text-center py-8">
        <div class="inline-flex items-center px-4 py-2 text-sm text-gray-600">
          <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          加载区块信息中...
        </div>
      </div>
    </div>

    <!-- 区块信息 -->
    <div v-else-if="block" class="space-y-3">
      <!-- 区块基本信息 -->
      <div class="card">
        <h2 class="text-lg font-medium text-gray-900 mb-2">区块信息</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
          <div>
            <label class="block text-sm font-medium text-gray-500">区块高度</label>
            <p class="mt-1 text-sm text-gray-900">#{{ block.height?.toLocaleString() }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">时间戳</label>
            <p class="mt-1 text-sm text-gray-900">{{ formatTimestamp(block.timestamp) }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">交易数量</label>
            <p class="mt-1 text-sm text-gray-900">{{ block.transaction_count?.toLocaleString() || block.transactions?.toLocaleString() || 'N/A' }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">区块大小</label>
            <p class="mt-1 text-sm text-gray-900">{{ formatBytes(block.size || block.stripped_size) }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">Gas使用</label>
            <p class="mt-1 text-sm text-gray-900">{{ formatGas(block.gas_used || block.gasUsed, block.gas_limit || block.gasLimit) }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">矿工地址</label>
            <p class="mt-1 text-sm text-gray-900 font-mono cursor-pointer hover:text-blue-600" @click="copyToClipboard(block.miner || block.miner_address, $event)">
              {{ block.miner || block.miner_address || 'N/A' }}
            </p>
          </div>
          <div>
            <span class="text-gray-500">区块奖励</span>
            <p class="mt-1 text-sm text-gray-900">
              <span class="font-medium">{{ formatMinerTip(block.miner_tip_eth) }} SOL</span>
              <span v-if="block.burned_eth && parseFloat(block.burned_eth) > 0" class="text-sm text-gray-500 ml-2">
                (燃烧: {{ formatBurnedEth(block.burned_eth) }} SOL)
              </span>
            </p>
          </div>
        </div>
      </div>

      <!-- 交易列表 -->
      <div class="card">
        <div class="flex justify-between items-center mb-2">
          <h2 class="text-lg font-medium text-gray-900">交易列表</h2>
          <div class="text-sm text-gray-500">
            共 {{ totalCount }} 笔交易 (第 {{ currentPage }}/{{ totalPages }} 页)
          </div>
        </div>


        <!-- 交易加载状态 -->
        <div v-if="loadingTransactions" class="text-center py-8">
          <div class="inline-flex items-center px-4 py-2 text-sm text-gray-600">
            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            加载交易中...
          </div>
        </div>

        <!-- 交易列表 -->
        <div v-else-if="transactions.length > 0" class="space-y-1">
          <div v-for="tx in transactions" :key="tx.id" class="bg-gray-50 p-2 rounded-lg">
            <!-- 交易基本信息 -->
            <div class="flex items-center justify-between">
              <div class="flex-1">
                <div class="flex items-center space-x-3 mb-1">
                  <span class="font-mono text-sm text-gray-600 cursor-pointer hover:text-blue-600" title="点击复制" @click="copyToClipboard(tx.tx_id || tx.hash, $event)">
                    {{ tx.tx_id || tx.hash || 'N/A' }}
                  </span>
                  <span class="text-sm text-gray-500">{{ formatTimestamp(tx.ctime || tx.timestamp) }}</span>
                </div>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm">
                  <div>
                    <span class="text-gray-500">从: </span>
                    <span class="font-mono text-blue-600 cursor-pointer hover:text-blue-800" title="点击复制" @click="copyToClipboard(tx.address_from || tx.from, $event)">
                      {{ tx.address_from || tx.from || 'N/A' }}
                    </span>
                  </div>
                  <div>
                    <span class="text-gray-500">到: </span>
                    <span class="font-mono text-blue-600 cursor-pointer hover:text-blue-800" title="点击复制" @click="copyToClipboard(tx.address_to || tx.to, $event)">
                      {{ tx.address_to || tx.to || 'N/A' }}
                    </span>
                    <span v-if="tx.is_token && tx.token_name" class="text-sm text-blue-600 ml-1">({{ tx.token_name }})</span>
                  </div>
                  <div>
                    <span class="text-gray-500">金额: </span>
                    <span class="font-medium">{{ formatAmount(tx.amount || tx.value) }} SOL</span>
                  </div>
                  <div>
                    <span class="text-gray-500">Gas: </span>
                    <span class="text-gray-600">{{ tx.gas_used?.toLocaleString() || tx.gasUsed?.toLocaleString() || 'N/A' }}</span>
                  </div>
                </div>
              </div>
              <div class="flex items-center space-x-2">
                <span :class="getStatusClass(tx.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                  {{ getStatusText(tx.status) }}
                </span>
                <button 
                  @click="toggleTransactionExpansion(tx.tx_id || tx.hash)"
                  class="inline-flex items-center px-2 py-1 text-xs font-medium text-gray-600 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                >
                  <svg v-if="!expandedTransactions[tx.tx_id || tx.hash]" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
                  </svg>
                  <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7"></path>
                  </svg>
                </button>
              </div>
            </div>

            <!-- 展开的交易凭证信息 -->
            <div v-if="expandedTransactions[tx.tx_id || tx.hash]" class="mt-2 pt-2 border-t border-gray-200">
              <!-- 未登录用户提示 -->
              <div v-if="!authStore.isAuthenticated" class="text-center py-3 text-gray-500">
                请登录后查看交易凭证信息
              </div>
              
              <!-- 已登录用户显示凭证信息 -->
              <div v-else-if="loadingReceipts[tx.tx_id || tx.hash]" class="text-center py-3">
                <div class="inline-flex items-center px-4 py-2 text-sm text-gray-600">
                  <svg class="animate-spin -ml-1 mr-3 h-4 w-4 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  加载凭证信息中...
                </div>
              </div>
              
              <div v-else-if="transactionReceipts[tx.tx_id || tx.hash]" class="space-y-2">
                <h4 class="text-sm font-medium text-gray-900 border-b border-gray-200 pb-2">交易详情</h4>
                
                <!-- 交易状态和区块信息 -->
                <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
                  <div>
                    <span class="text-gray-500">状态: </span>
                    <span :class="getReceiptStatusClass(transactionReceipts[tx.tx_id || tx.hash].status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full ml-2">
                      {{ getReceiptStatusText(transactionReceipts[tx.tx_id || tx.hash].status) }}
                    </span>
                  </div>
                  <div>
                    <span class="text-gray-500">区块: </span>
                    <span class="text-gray-600">{{ transactionReceipts[tx.tx_id || tx.hash].block_number?.toLocaleString() || 'N/A' }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">区块内位置: </span>
                    <span class="text-gray-600">{{ transactionReceipts[tx.tx_id || tx.hash].transaction_index || 'N/A' }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">时间戳: </span>
                    <span class="text-gray-600">{{ formatTimestamp(tx.ctime || tx.timestamp) }}</span>
                  </div>
                </div>

                <!-- 交易费用详情 - 像 Etherscan.io 一样完整 -->
                <div class="bg-gray-50 p-2 rounded-lg">
                  <h5 class="text-sm font-medium text-gray-900 mb-1">交易费用详情</h5>
                  
                  <!-- 主要费用信息 -->
                  <div class="grid grid-cols-1 md:grid-cols-3 gap-2 text-sm mb-3">
                    <div>
                      <span class="text-gray-500">交易手续费: </span>
                      <span class="text-gray-900 font-medium">{{ formatTransactionFeeFromReceipt(transactionReceipts[tx.tx_id || tx.hash]) }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">Gas 价格: </span>
                      <span class="text-gray-900 font-medium">{{ formatGasPriceFromReceipt(transactionReceipts[tx.tx_id || tx.hash]) }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">Gas Limit & Gas 使用: </span>
                      <span class="text-gray-600">{{ tx.gas_limit || tx.gasLimit || 'N/A' }} | {{ transactionReceipts[tx.tx_id || tx.hash].gas_used || 'N/A' }} ({{ tx.gas_limit && transactionReceipts[tx.tx_id || tx.hash].gas_used ? ((transactionReceipts[tx.tx_id || tx.hash].gas_used / tx.gas_limit) * 100).toFixed(2) : 'N/A' }}%)</span>
                    </div>
                  </div>

                  <!-- Gas 使用 - 已合并到上面，这里移除 -->

                  <!-- EIP-1559 费用详情 -->
                  <div v-if="(tx.type || tx.tx_type) === 2" class="border-t border-gray-200 pt-2">
                    <h6 class="text-sm font-medium text-gray-700 mb-2">EIP-1559 费用详情</h6>
                    <div class="grid grid-cols-1 md:grid-cols-3 gap-2 text-sm">
                      <div>
                        <span class="text-gray-500">基础费: </span>
                        <span class="text-gray-600">{{ formatBaseFee(block.base_fee) }}</span>
                      </div>
                      <div>
                        <span class="text-gray-500">最高费用: </span>
                        <span class="text-gray-600">{{ formatGasPrice(tx.max_fee_per_gas || tx.maxFeePerGas) }}</span>
                      </div>
                      <div>
                        <span class="text-gray-500">最高小费: </span>
                        <span class="text-gray-600">{{ formatGasPrice(tx.max_priority_fee_per_gas || tx.maxPriorityFeePerGas) }}</span>
                      </div>
                    </div>
                  </div>

                  <!-- 燃烧和节省费用 -->
                  <div v-if="block.base_fee" class="border-t border-gray-200 pt-2">
                    <h6 class="text-sm font-medium text-gray-700 mb-2">燃烧和节省费用</h6>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm">
                      <div>
                        <span class="text-gray-500">燃烧费: </span>
                        <span class="text-red-600 font-medium">{{ formatBurnedFee(block.base_fee, transactionReceipts[tx.tx_id || tx.hash].gas_used) }}</span>
                      </div>
                      <div>
                        <span class="text-gray-500">节省费用: </span>
                        <span class="text-green-600 font-medium">{{ formatSavedFee(tx, block.base_fee, transactionReceipts[tx.tx_id || tx.hash].gas_used) }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 交易属性 - 更详细的信息 -->
                <div class="bg-gray-50 p-2 rounded-lg border-t border-gray-200">
                  <h5 class="text-sm font-medium text-gray-900 mb-1">交易属性</h5>
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm">
                    <div>
                      <span class="text-gray-500">交易类型: </span>
                      <span class="text-gray-600">{{ getTransactionTypeText(tx.type || tx.tx_type) }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">Nonce: </span>
                      <span class="text-gray-600">{{ tx.nonce !== undefined && tx.nonce !== null ? tx.nonce : 'N/A' }}</span>
                    </div>
                  </div>
                  
                  <!-- 合约相关信息 -->
                  <div v-if="tx.is_token" class="border-t border-gray-200 pt-1 mt-1">
                    <h6 class="text-sm font-medium text-gray-700 mb-1">合约信息</h6>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm">
                                          <div>
                        <span class="text-gray-500">代币交易: </span>
                        <span class="text-green-600 font-medium">ERC-20</span>
                    </div>
                      <div v-if="tx.contract_addr && tx.contract_addr !== '0x0000000000000000000000000000000000000000'">
                        <span class="text-gray-500">合约地址: </span>
                        <span class="font-mono text-blue-600 cursor-pointer hover:text-blue-800" @click="copyToClipboard(tx.contract_addr, $event)">
                          {{ tx.contract_addr }}
                        </span>
                      </div>
                      <div v-if="tx.token_name">
                        <span class="text-gray-500">代币名称: </span>
                        <span class="text-blue-600 font-medium">{{ tx.token_name }}</span>
                      </div>
                      <div v-if="tx.token_symbol">
                        <span class="text-gray-500">代币符号: </span>
                        <span class="text-gray-900 font-medium">{{ tx.token_symbol }}</span>
                      </div>
                      <div v-if="tx.token_decimals">
                        <span class="text-gray-500">代币精度: </span>
                        <span class="text-gray-900 font-medium">{{ tx.token_decimals }} 位小数</span>
                      </div>
                      <div v-if="tx.token_market_cap_rank">
                        <span class="text-gray-500">市值排名: </span>
                        <span class="text-gray-900 font-medium">#{{ tx.token_market_cap_rank }}</span>
                      </div>
                      <div v-if="tx.token_is_stablecoin">
                        <span class="text-gray-500">代币类型: </span>
                        <span class="text-green-600 font-medium">稳定币</span>
                      </div>
                      <div v-if="tx.token_is_verified">
                        <span class="text-gray-500">验证状态: </span>
                        <span class="text-green-600 font-medium">已验证</span>
                      </div>
                      <div v-if="tx.token_description">
                        <span class="text-gray-500">代币描述: </span>
                        <span class="text-gray-900">{{ tx.token_description }}</span>
                      </div>
                      <div v-if="tx.token_website">
                        <span class="text-gray-500">官方网站: </span>
                        <a :href="tx.token_website" target="_blank" class="text-blue-600 hover:text-blue-800 underline">
                          {{ tx.token_website }}
                        </a>
                      </div>
                      <div v-if="tx.token_explorer">
                        <span class="text-gray-500">浏览器链接: </span>
                        <a :href="tx.token_explorer" target="_blank" class="text-blue-600 hover:text-blue-800 underline">
                          {{ tx.token_explorer }}
                        </a>
                      </div>
                    </div>
                  </div>

                  <!-- 解析合约转账（优先后端预解析） -->
                  <div v-if="tx.is_token && parsedResults[(tx.tx_id || tx.hash)] && parsedResults[(tx.tx_id || tx.hash)].length > 0" class="border-t border-gray-200 pt-1 mt-1">
                    <h6 class="text-sm font-medium text-gray-700 mb-1">解析合约</h6>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm">
                      <div>
                        <span class="text-gray-500">From: </span>
                        <span class="font-mono text-blue-600 cursor-pointer hover:text-blue-800" @click="copyToClipboard(parsedResults[(tx.tx_id || tx.hash)][0]?.from_address || '', $event)">
                          {{ parsedResults[(tx.tx_id || tx.hash)][0]?.from_address }}
                        </span>
                      </div>
                      <div>
                        <span class="text-gray-500">To: </span>
                        <span class="font-mono text-blue-600 cursor-pointer hover:text-blue-800" @click="copyToClipboard(parsedResults[(tx.tx_id || tx.hash)][0]?.to_address || '', $event)">
                          {{ parsedResults[(tx.tx_id || tx.hash)][0]?.to_address }}
                        </span>
                      </div>
                      <div class="md:col-span-2">
                        <span class="text-gray-500">Amount: </span>
                        <span class="text-gray-900 font-medium">
                          {{ formatNumber.wei(new BigNumber(parsedResults[(tx.tx_id || tx.hash)][0]?.amount_wei || '0').dividedBy(new BigNumber(10).pow(tx.token_decimals)).toString()) }}
                          {{ parsedResults[(tx.tx_id || tx.hash)][0]?.token_symbol || tx.token_symbol || 'SOL' }}
                        </span>
                      </div>
                    </div>
                  </div>

                  <!-- 输入数据 -->
                  <div class="border-t border-gray-200 pt-1 mt-1">
                    <h6 class="text-sm font-medium text-gray-700 mb-1">输入数据</h6>
                    <div class="bg-white p-1 rounded border overflow-x-auto max-w-full">
                      <pre class="text-xs text-gray-700 whitespace-pre-wrap break-all max-w-full">{{ parseInputDataWithConfig(transactionReceipts[tx.tx_id || tx.hash]?.input_data || tx.hex || tx.input || tx.data, tx.tx_id || tx.hash) }}</pre>
                    </div>
                  </div>
                </div>

                <!-- 交易日志 -->
                <div v-if="transactionReceipts[tx.tx_id || tx.hash]?.logs_data" class="bg-gray-50 p-2 rounded-lg">
                  <h5 class="text-sm font-medium text-gray-900 mb-1">交易日志</h5>
                  <div class="bg-white p-1 rounded border overflow-x-auto max-w-full">
                    <pre class="text-xs text-gray-700 whitespace-pre-wrap break-all max-w-full">{{ formatLogsData(transactionReceipts[tx.tx_id || tx.hash].logs_data) }}</pre>
                  </div>
                </div>
              </div>

              <div v-else class="text-center py-4 text-gray-500 text-sm">
                暂无凭证信息
              </div>
            </div>
          </div>

          <!-- 分页控件 -->
          <div v-if="totalPages > 1" class="mt-6 flex justify-center">
            <nav class="flex items-center space-x-2">
              <button 
                @click="changePage(currentPage - 1)" 
                :disabled="currentPage <= 1"
                class="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                上一页
              </button>
              
              <div class="flex items-center space-x-1">
                <span v-for="page in visiblePages" :key="page" 
                      @click="changePage(page)"
                      :class="[
                        'px-3 py-2 text-sm font-medium rounded-md cursor-pointer',
                        page === currentPage 
                          ? 'bg-blue-600 text-white' 
                          : 'text-gray-500 bg-white border border-gray-300 hover:bg-gray-50'
                      ]"
                >
                  {{ page }}
                </span>
              </div>
              
              <button 
                @click="changePage(currentPage + 1)" 
                :disabled="currentPage >= totalPages"
                class="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                下一页
              </button>
            </nav>
          </div>
        </div>

        <!-- 无交易状态 -->
        <div v-else class="text-center py-8 text-gray-500">
          该区块暂无交易
        </div>
      </div>
    </div>

    <!-- 错误状态 -->
    <div v-else class="card">
      <div class="text-center py-8">
        <div class="text-red-600 mb-2">
          <svg class="w-12 h-12 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"></path>
          </svg>
        </div>
        <h3 class="text-lg font-medium text-gray-900 mb-2">加载失败</h3>
        <p class="text-gray-500 mb-4">{{ errorMessage || '无法加载区块信息' }}</p>
        <button 
          @click="loadBlockData" 
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
        >
          重试
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { blocks as blocksApi } from '@/api'
import { transactions as transactionsApi } from '@/api'
import BigNumber from 'bignumber.js'

// 路由参数
const route = useRoute()
const blockHeight = computed(() => route.params.height as string)

// 认证store
const authStore = useAuthStore()

// 响应式数据
const block = ref<any>(null)
const transactions = ref<any[]>([])
const isLoading = ref(true)
const loadingTransactions = ref(true)
const errorMessage = ref('')

// 复制提示（跟随点击位置）
const showToast = ref(false)
const toastMessage = ref('')
const toastX = ref<number | null>(null)
const toastY = ref<number | null>(null)
const toastStyle = computed(() => {
  if (toastX.value !== null && toastY.value !== null) {
    return { top: `${toastY.value}px`, left: `${toastX.value}px` }
  }
  return { top: '16px', right: '16px' }
})
let toastTimer: any = null

// 分页相关数据
const currentPage = ref(1)
const pageSize = ref(20)
const totalCount = ref(0)
const totalPages = ref(1)

// 交易展开相关数据
const expandedTransactions = ref<Record<string, boolean>>({})
const loadingReceipts = ref<Record<string, boolean>>({})
const transactionReceipts = ref<Record<string, any>>({})
// 解析结果缓存与加载状态
const parsedResults = ref<Record<string, any[]>>({})
const loadingParsed = ref<Record<string, boolean>>({})

// 计算属性
const isFilteredByBlock = computed(() => {
  // 检查交易是否按区块筛选
  if (transactions.value.length === 0) return false
  
  // 如果第一个交易有区块高度字段，说明是按区块筛选的
  const firstTx = transactions.value[0]
  return !!(firstTx.blockHeight || firstTx.block_number || firstTx.block_height)
})

// 分页计算属性
const visiblePages = computed(() => {
  const pages = []
  const start = Math.max(1, currentPage.value - 2)
  const end = Math.min(totalPages.value, currentPage.value + 2)
  
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  return pages
})

// 统一的数字格式化工具 - 使用BigNumber确保精确计算
const formatNumber = {
  // 格式化SOL金额（9位精度，去掉末尾0）
  eth: (value: string | number | undefined): string => {
    if (value === undefined || value === null) return '0'
    
    try {
      // 使用BigNumber确保精确计算
      const eth = new BigNumber(value)
      if (eth.isZero()) return '0'
      
      // 使用BigNumber的toFixed确保18位精度，然后去掉末尾的0
      let result = eth.toFixed(18)
      result = result.replace(/0+$/, '') // 只去掉末尾的0
      if (result.endsWith('.')) {
        result = result.slice(0, -1) // 去掉末尾的小数点
      }
      
      return result
    } catch (error) {
      console.error('SOL格式化错误:', error, value)
      return '0'
    }
  },
  
  // 格式化Gwei金额（9位精度，去掉末尾0）
  gwei: (value: string | number | undefined): string => {
    if (value === undefined || value === null) return '0'
    
    try {
      // 使用BigNumber确保精确计算
      const wei = new BigNumber(value)
      if (wei.isZero()) return '0'
      
      // 转换为Gwei单位：wei / 10^9
      const gwei = wei.dividedBy(new BigNumber(10).pow(9))
      
      // 使用BigNumber的toFixed确保9位精度，然后去掉末尾的0
      let result = gwei.toFixed(9)
      result = result.replace(/0+$/, '') // 只去掉末尾的0
      if (result.endsWith('.')) {
        result = result.slice(0, -1) // 去掉末尾的小数点
      }
      
      return result
    } catch (error) {
      console.error('Gwei格式化错误:', error, value)
      return '0'
    }
  },
  
  // 格式化Wei金额（保持原始精度，去掉末尾0）
  wei: (value: string | number | undefined): string => {
    if (value === undefined || value === null) return '0'
    
    try {
      // 使用BigNumber确保精确计算
      const wei = new BigNumber(value)
      if (wei.isZero()) return '0'
      
      // 直接转换为字符串，然后去掉末尾的0
      let result = wei.toString()
      if (result.includes('.')) {
        result = result.replace(/0+$/, '') // 只去掉末尾的0
        if (result.endsWith('.')) {
          result = result.slice(0, -1) // 去掉末尾的小数点
        }
      }
      
      return result
    } catch (error) {
      console.error('Wei格式化错误:', error, value)
      return '0'
    }
  },
  
  // 从Lamports转换为SOL（用于交易金额）
  weiToEth: (value: string | number | undefined): string => {
    if (value === undefined || value === null) return '0'
    
    try {
      // 使用BigNumber确保精确计算
      const wei = new BigNumber(value)
      if (wei.isZero()) return '0'
      
      // 转换为SOL单位：lamports / 10^9
      const sol = wei.dividedBy(new BigNumber(10).pow(9))
      
      // 使用BigNumber的toFixed确保9位精度，然后去掉末尾的0
      let result = sol.toFixed(9)
      result = result.replace(/0+$/, '') // 只去掉末尾的0
      if (result.endsWith('.')) {
        result = result.slice(0, -1) // 去掉末尾的小数点
      }
      
      return result
    } catch (error) {
      console.error('Lamports到SOL转换错误:', error, value)
      return '0'
    }
  }
}

// 格式化函数
const formatTimestamp = (timestamp: string | number) => {
  if (!timestamp) return 'N/A'
  
  let date: Date
  if (typeof timestamp === 'string') {
    // 处理ISO格式字符串
    date = new Date(timestamp)
  } else {
    // 处理Unix时间戳
    date = new Date(timestamp * 1000)
  }
  
  // 检查日期是否有效
  if (isNaN(date.getTime())) {
    return 'Invalid Date'
  }
  
  return date.toLocaleString()
}

// 格式化矿工奖励（MinerTipEth）- 使用统一格式化工具
const formatMinerTip = (minerTip: string | number | undefined): string => {
  return formatNumber.eth(minerTip)
}

// 格式化燃烧费用 - 使用统一格式化工具
const formatBurnedEth = (burnedEth: string | number | undefined): string => {
  return formatNumber.eth(burnedEth)
}

const formatAddress = (address: string) => {
  if (!address) return 'N/A'
  return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`
}

const formatBytes = (bytes: number) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toString() + ' ' + sizes[i]
}

const formatGas = (used: number, limit: number) => {
  if (!used || !limit) return 'N/A'
  const percentage = ((used / limit) * 100).toString()
  return `${used.toLocaleString()} / ${limit.toLocaleString()} (${percentage}%)`
}

// 格式化交易金额 - 使用统一格式化工具
const formatAmount = (amount: number | string) => {
  if (!amount) return '0'
  
  // 如果amount是字符串，先转换为数字
  let num: number
  if (typeof amount === 'string') {
    num = parseFloat(amount)
  } else {
    num = amount
  }
  
  if (isNaN(num) || num === 0) return '0'
  
  // 交易金额是Lamports，需要转换为SOL
  return formatNumber.weiToEth(num)
}

const formatHash = (hash: string) => {
  if (!hash) return 'N/A'
  return `${hash.substring(0, 10)}...${hash.substring(hash.length - 10)}`
}

const getStatusClass = (status: number) => {
  switch (status) {
    case 0:
      return 'bg-gray-100 text-gray-800'
    case 1:
      return 'bg-green-100 text-green-800'
    case 2:
      return 'bg-red-100 text-red-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

const getStatusText = (status: number) => {
  switch (status) {
    case 0:
      return 'Pending'
    case 1:
      return 'Success'
    case 2:
      return 'Failed'
    default:
      return 'Unknown'
  }
}

// 加载区块数据
const loadBlockData = async () => {
  try {
    isLoading.value = true
    errorMessage.value = ''
    
    // 根据登录状态调用不同的API
    if (authStore.isAuthenticated) {
      // 已登录用户：调用 /v1/ 下的API
      console.log('🔐 已登录用户，调用 /v1/ API 获取区块详情')
      const response = await blocksApi.getBlock({ 
        height: parseInt(blockHeight.value), 
        chain: 'sol' 
      })
      
      if (response && response.success === true) {
        console.log('📊 后端返回区块数据:', response.data)
        block.value = response.data
      } else {
        throw new Error(response?.message || '获取区块信息失败')
      }
    } else {
      // 未登录用户：调用 /no-auth/ 下的API（有限制）
      console.log('👤 未登录用户，调用 /no-auth/ API 获取区块详情（有限制）')
      const response = await blocksApi.getBlockPublic({ 
        height: parseInt(blockHeight.value), 
        chain: 'sol' 
      })
      
      if (response && response.success === true) {
        console.log('📊 后端返回区块数据:', response.data)
        block.value = response.data
      } else {
        throw new Error(response?.message || '获取区块信息失败')
      }
    }
  } catch (error) {
    console.error('Failed to load block:', error)
    errorMessage.value = error instanceof Error ? error.message : '加载区块信息失败'
  } finally {
    isLoading.value = false
  }
}

// 加载交易数据
const loadTransactions = async () => {
  try {
    loadingTransactions.value = true
    
    // 根据登录状态调用不同的API
    if (authStore.isAuthenticated) {
      // 已登录用户：调用 /v1/ 下的API
      console.log('🔐 已登录用户，调用 /v1/ API 获取区块交易')
      const response = await blocksApi.getBlockTransactions({
        height: parseInt(blockHeight.value),
        chain: 'sol',
        page: currentPage.value,
        page_size: pageSize.value
      })
      
      if (response && response.success === true) {
        console.log('📊 后端返回交易数据:', response.data)
        
        // 新API直接返回交易数据，不需要过滤
        const responseData = response.data as any
        console.log('🔍 解析API返回数据:', responseData)
        
        if (responseData?.transactions && Array.isArray(responseData.transactions)) {
          transactions.value = responseData.transactions
          
          // 尝试多种可能的字段名
          totalCount.value = responseData.total_count || responseData.total || responseData.totalCount || responseData.totalTransactions || responseData.transaction_count || 0
          
          // 如果总数还是0，但有交易数据，说明可能是单页返回所有数据
          if (totalCount.value === 0 && transactions.value.length > 0) {
            // 尝试从区块信息中获取交易总数
            if (block.value && block.value.transaction_count) {
              totalCount.value = block.value.transaction_count
              console.log('📊 从区块信息获取交易总数:', totalCount.value)
            } else if (block.value && block.value.transactions) {
              totalCount.value = block.value.transactions
              console.log('📊 从区块信息获取交易总数:', totalCount.value)
            } else {
              totalCount.value = transactions.value.length
              console.log('⚠️ 后端未返回总数，使用当前页交易数量作为总数')
            }
          }
          
          totalPages.value = Math.max(1, Math.ceil(totalCount.value / pageSize.value))
          console.log('✅ 成功加载区块交易:', transactions.value.length, '笔交易，总计:', totalCount.value, '页数:', totalPages.value)
        } else {
          console.warn('API返回数据格式异常:', responseData)
          transactions.value = []
          totalCount.value = 0
          totalPages.value = 1
        }
      } else {
        throw new Error(response?.message || '获取交易信息失败')
      }
    } else {
      // 未登录用户：调用 /no-auth/ 下的API（有限制）
      console.log('👤 未登录用户，调用 /no-auth/ API 获取区块交易（有限制）')
      const response = await blocksApi.getBlockTransactionsPublic({
        height: parseInt(blockHeight.value),
        chain: 'sol',
        page: currentPage.value,
        page_size: pageSize.value
      })
      
      if (response && response.success === true) {
        console.log('📊 后端返回交易数据:', response.data)
        
        // 新API直接返回交易数据，不需要过滤
        const responseData = response.data as any
        console.log('🔍 解析API返回数据:', responseData)
        
        if (responseData?.transactions && Array.isArray(responseData.transactions)) {
          transactions.value = responseData.transactions
          
          // 尝试多种可能的字段名
          totalCount.value = responseData.total_count || responseData.total || responseData.totalCount || responseData.totalTransactions || responseData.transaction_count || 0
          
          // 如果总数还是0，但有交易数据，说明可能是单页返回所有数据
          if (totalCount.value === 0 && transactions.value.length > 0) {
            // 尝试从区块信息中获取交易总数
            if (block.value && block.value.transaction_count) {
              totalCount.value = block.value.transaction_count
              console.log('📊 从区块信息获取交易总数:', totalCount.value)
            } else if (block.value && block.value.transactions) {
              totalCount.value = block.value.transactions
              console.log('📊 从区块信息获取交易总数:', totalCount.value)
            } else {
              totalCount.value = transactions.value.length
              console.log('⚠️ 后端未返回总数，使用当前页交易数量作为总数')
            }
          }
          
          totalPages.value = Math.max(1, Math.ceil(totalCount.value / pageSize.value))
          console.log('✅ 成功加载区块交易:', transactions.value.length, '笔交易，总计:', totalCount.value, '页数:', totalPages.value)
        } else {
          console.warn('API返回数据格式异常:', responseData)
          transactions.value = []
          totalCount.value = 0
          totalPages.value = 1
        }
      } else {
        throw new Error(response?.message || '获取交易信息失败')
      }
    }
  } catch (error) {
    console.error('Failed to load transactions:', error)
    transactions.value = []
    totalCount.value = 0
    totalPages.value = 1
  } finally {
    loadingTransactions.value = false
  }
}

// 复制到剪贴板（支持传入点击事件以定位提示位置）
const copyToClipboard = async (text: string, e?: MouseEvent) => {
  try {
    await navigator.clipboard.writeText(text)
    // 计算提示位置（相对视口，稍微偏移）
    if (e) {
      const offset = 12
      toastX.value = Math.min(window.innerWidth - 16, e.clientX + offset)
      toastY.value = Math.min(window.innerHeight - 16, e.clientY + offset)
    } else {
      toastX.value = null
      toastY.value = null
    }
    toastMessage.value = '已复制到剪贴板'
    showToast.value = true
    if (toastTimer) clearTimeout(toastTimer)
    toastTimer = setTimeout(() => {
      showToast.value = false
      toastTimer = null
    }, 1200)
  } catch (err) {
    console.error('复制失败:', err)
  }
}

// 分页切换
const changePage = async (page: number) => {
  if (page < 1 || page > totalPages.value) return
  
  currentPage.value = page
  await loadTransactions()
}

// 切换交易展开状态
const toggleTransactionExpansion = async (txHash: string) => {
  if (!txHash) return
  
  const isExpanded = expandedTransactions.value[txHash]
  expandedTransactions.value[txHash] = !isExpanded
  
  // 如果展开且还没有加载凭证，且用户已登录，则加载
  if (!isExpanded && !transactionReceipts.value[txHash] && authStore.isAuthenticated) {
    await loadTransactionReceipt(txHash)
  }

  // 同步尝试加载后端预解析结果（需要登录）
  if (!isExpanded && authStore.isAuthenticated && !parsedResults.value[txHash] && !loadingParsed.value[txHash]) {
    await loadParsedResults(txHash)
  }
}

// 加载交易凭证
const loadTransactionReceipt = async (txHash: string) => {
  if (!txHash || transactionReceipts.value[txHash]) return
  
  try {
    loadingReceipts.value[txHash] = true
    
    // 调用API获取凭证
    const response = await transactionsApi.getTransactionReceipt(txHash)
    
    if (response && response.success === true) {
      transactionReceipts.value[txHash] = response.data
      console.log('✅ 成功加载交易凭证:', txHash, response.data)
    } else {
      console.warn('获取交易凭证失败:', response?.message)
    }
  } catch (error) {
    console.error('Failed to load transaction receipt:', error)
  } finally {
    loadingReceipts.value[txHash] = false
  }
}

// 加载交易解析结果（后端预解析）
const loadParsedResults = async (txHash: string) => {
  try {
    loadingParsed.value[txHash] = true
    const resp = await transactionsApi.getParsedTransaction(txHash)
    if (resp && resp.success === true) {
      parsedResults.value[txHash] = resp.data || []
    }
  } catch (e) {
    console.warn('加载交易解析结果失败:', e)
  } finally {
    loadingParsed.value[txHash] = false
  }
}

// 凭证状态样式
const getReceiptStatusClass = (status: number) => {
  switch (status) {
    case 0:
      return 'bg-red-100 text-red-800'
    case 1:
      return 'bg-green-100 text-green-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

// 凭证状态文本
const getReceiptStatusText = (status: number) => {
  switch (status) {
    case 0:
      return 'Failed'
    case 1:
      return 'Success'
    default:
      return 'Unknown'
  }
}

// 格式化日志数据
const formatLogsData = (logsData: string) => {
  try {
    if (typeof logsData === 'string') {
      const parsed = JSON.parse(logsData)
      return JSON.stringify(parsed, null, 2)
    }
    return JSON.stringify(logsData, null, 2)
  } catch (error) {
    return logsData || 'Invalid logs data'
  }
}

// 格式化Gas价格 - 使用统一格式化工具
const formatGasPrice = (gasPrice: number | string) => {
  if (!gasPrice) return 'N/A'
  
  let price: number
  if (typeof gasPrice === 'string') {
    // 智能检测：如果以0x开头，按十六进制解析；否则按十进制解析
    if (gasPrice.startsWith('0x')) {
      price = parseInt(gasPrice, 16)
    } else {
      price = parseInt(gasPrice, 10)
    }
  } else {
    price = gasPrice
  }
  
  if (isNaN(price) || price === 0) return '0 Gwei'
  
  // 使用统一格式化工具，转换为Gwei并添加单位
  return `${formatNumber.gwei(price)} Gwei`
}



// 获取交易类型文本
const getTransactionTypeText = (type: number | string) => {
  if (!type) return 'Legacy'
  
  const txType = typeof type === 'string' ? parseInt(type, 16) : type
  
  switch (txType) {
    case 0:
      return 'Legacy'
    case 1:
      return '1(EIP-2930)'
    case 2:
      return '2(EIP-1559)'
    case 3:
      return '3(EIP-4844)'
    default:
      return `${txType}(Type ${txType})`
  }
}


// 解析交易输入数据（使用parser_configs）
const parseInputDataWithConfig = (inputData: string, txHash?: string) => {
  if (!inputData || inputData === '0x') return '0x (No input data)'
  
  // 优先使用后端返回的解析配置
  if (txHash && transactionReceipts.value[txHash]?.parser_configs) {
    const receipt = transactionReceipts.value[txHash]
    const signature = inputData.substring(0, 10)
    
    // 查找匹配的解析配置
    const matchedConfig = receipt.parser_configs.find((config: any) => 
      config.function_signature === signature
    )
    
    if (matchedConfig) {
      let result = `方法名：${signature}(${matchedConfig.function_description})\n`
      
      // 如果有参数配置，解析参数
      if (matchedConfig.param_config && matchedConfig.param_config.length > 0) {
        for (const param of matchedConfig.param_config) {
          if (param.offset !== undefined && param.length) {
            // 偏移量是以字节为单位，需要转换为十六进制字符位置
            // 每个字节 = 2个十六进制字符
            const startPos = param.offset * 2
            const endPos = startPos + param.length * 2
            
            console.log(`🔍 解析参数 ${param.name}:`, {
              offset: param.offset,
              length: param.length,
              startPos,
              endPos,
              inputDataLength: inputData.length
            })
            
            const paramValue = inputData.substring(startPos, endPos)
            console.log(`🔍 参数 ${param.name} 值:`, paramValue)
            
            // 根据参数类型格式化显示
            if (param.type === 'address') {
              // 地址类型：添加0x前缀
              result += `${param.name}: 0x${paramValue}\n`
            } else {
              // 数值类型：不添加0x前缀
              result += `${param.name}: ${paramValue}\n`
            }
          }
        }
      }
      
      result += `Raw Data: ${inputData}`
      return result
    }
  }
}

// 解析交易日志（使用parser_configs）
const parseLogsDataWithConfig = (logsData: string, txHash?: string) => {
  if (!logsData) return 'No logs data'
  
  try {
    const logs = JSON.parse(logsData)
    
    // 优先使用后端返回的解析配置
    if (txHash && transactionReceipts.value[txHash]?.parser_configs) {
      const receipt = transactionReceipts.value[txHash]
      
      // 查找event_log类型的解析配置
      const eventConfigs = receipt.parser_configs.filter((config: any) => 
        config.parser_type === 'event_log'
      )
      
      if (eventConfigs.length > 0) {
        let result = 'Parsed Logs:\n'
        
        for (const log of logs) {
          if (log.topics && log.topics.length > 0) {
            const eventSignature = log.topics[0]
            
            // 查找匹配的事件配置
            const matchedEvent = eventConfigs.find((config: any) => 
              config.function_signature === eventSignature
            )
            
            if (matchedEvent) {
              result += `Event: ${matchedEvent.display_format || matchedEvent.function_description}\n`
              result += `Address: ${log.address}\n`
              result += `Data: ${log.data}\n`
              result += `Topics: ${log.topics.join(', ')}\n\n`
            } else {
              result += `Unknown Event: ${eventSignature}\n`
              result += `Address: ${log.address}\n`
              result += `Data: ${log.data}\n\n`
            }
          }
        }
        
        return result
      }
    }
    
    // 降级到原来的格式化
    return formatLogsData(logsData)
  } catch (error) {
    return formatLogsData(logsData)
  }
}


// 格式化Base Fee - 使用统一格式化工具
const formatBaseFee = (baseFee: string | number | undefined): string => {
  if (!baseFee) return 'N/A'
  
  let fee: number
  if (typeof baseFee === 'string') {
    // 智能检测：如果以0x开头，按十六进制解析；否则按十进制解析
    if (baseFee.startsWith('0x')) {
      fee = parseInt(baseFee, 16)
    } else {
      fee = parseInt(baseFee, 10)
    }
  } else {
    fee = baseFee
  }
  
  if (isNaN(fee) || fee === 0) return '0 Gwei'
  
  // 使用统一格式化工具，转换为Gwei并添加单位
  return `${formatNumber.gwei(fee)} Gwei`
}

// 格式化燃烧费用 - 使用BigNumber确保精确计算
const formatBurnedFee = (baseFee: string | number | undefined, gasUsed: number): string => {
  if (!baseFee || !gasUsed) return '0 SOL'
  
  try {
    // 使用BigNumber确保精确计算
    let baseFeeBN: BigNumber
    if (typeof baseFee === 'string') {
      // 智能检测：如果以0x开头，按十六进制解析；否则按十进制解析
      if (baseFee.startsWith('0x')) {
        baseFeeBN = new BigNumber(parseInt(baseFee, 16))
      } else {
        baseFeeBN = new BigNumber(baseFee)
      }
    } else {
      baseFeeBN = new BigNumber(baseFee)
    }
    
    // 燃烧费用 = Base Fee * Gas Used
    const burnedWei = baseFeeBN.times(gasUsed)
    
    // 使用统一格式化工具，转换为ETH并添加单位
    return `${formatNumber.weiToEth(burnedWei.toString())} SOL`
  } catch (error) {
    console.error('燃烧费用计算错误:', error)
    return '0 SOL'
  }
}

// 格式化节省费用 - 使用BigNumber确保精确计算
const formatSavedFee = (tx: any, baseFee: string | number | undefined, gasUsed: number): string => {
  if (!baseFee || !gasUsed) return '0 SOL'
  
  try {
    // 使用BigNumber确保精确计算
    let baseFeeBN: BigNumber
    if (typeof baseFee === 'string') {
      // 智能检测：如果以0x开头，按十六进制解析；否则按十进制解析
      if (baseFee.startsWith('0x')) {
        baseFeeBN = new BigNumber(parseInt(baseFee, 16))
      } else {
        baseFeeBN = new BigNumber(baseFee)
      }
    } else {
      baseFeeBN = new BigNumber(baseFee)
    }
    
    const maxFee = tx.max_fee_per_gas || tx.maxFeePerGas
    let maxFeeBN: BigNumber
    if (typeof maxFee === 'string') {
      // 智能检测：如果以0x开头，按十六进制解析；否则按十进制解析
      if (maxFee.startsWith('0x')) {
        maxFeeBN = new BigNumber(maxFee)
      } else {
        maxFeeBN = new BigNumber(maxFee)
      }
    } else {
      maxFeeBN = new BigNumber(maxFee)
    }
    
    // 获取Effective Gas Price（从交易回执）
    const receipt = transactionReceipts.value[tx.tx_id || tx.hash]
    if (!receipt?.effective_gas_price) {
      // 如果没有回执，使用Base Fee + Priority Fee作为Effective Gas Price
      const priorityFee = tx.max_priority_fee_per_gas || tx.maxPriorityFeePerGas || 0
      const priorityFeeBN = new BigNumber(priorityFee)
      const effectiveGasPriceBN = baseFeeBN.plus(priorityFeeBN)
      
      // 节省费用 = (Max Fee - Effective Gas Price) * Gas Used
      const savedWei = maxFeeBN.minus(effectiveGasPriceBN).times(gasUsed)
      
      // 调试信息
      console.log('节省费用计算(无回执):', {
        txHash: tx.tx_id || tx.hash,
        maxFee: maxFeeBN.toString(),
        baseFee: baseFeeBN.toString(),
        priorityFee: priorityFeeBN.toString(),
        effectiveGasPrice: effectiveGasPriceBN.toString(),
        gasUsed,
        savedWei: savedWei.toString(),
        savedEth: formatNumber.weiToEth(savedWei.toString())
      })
      
      if (savedWei.isLessThan(0)) {
        return '0 SOL'
      }
      
      return `${formatNumber.weiToEth(savedWei.toString())} ETH`
    }
    
    // 使用回执中的Effective Gas Price
    let effectiveGasPriceBN: BigNumber
    if (typeof receipt.effective_gas_price === 'string') {
      if (receipt.effective_gas_price.startsWith('0x')) {
        effectiveGasPriceBN = new BigNumber(parseInt(receipt.effective_gas_price, 16))
      } else {
        effectiveGasPriceBN = new BigNumber(receipt.effective_gas_price)
      }
    } else {
      effectiveGasPriceBN = new BigNumber(receipt.effective_gas_price)
    }
    
    // 节省费用 = (Max Fee - Effective Gas Price) * Gas Used
    const savedWei = maxFeeBN.minus(effectiveGasPriceBN).times(gasUsed)
    
    // 调试信息
    console.log('节省费用计算(有回执):', {
      txHash: tx.tx_id || tx.hash,
      maxFee: maxFeeBN.toString(),
      effectiveGasPrice: effectiveGasPriceBN.toString(),
      gasUsed,
      savedWei: savedWei.toString(),
      savedEth: formatNumber.weiToEth(savedWei.toString())
    })
    
    if (savedWei.isLessThan(0)) {
      return '0 SOL'
    }
    
    return `${formatNumber.weiToEth(savedWei.toString())} ETH`
  } catch (error) {
    console.error('节省费用计算错误:', error)
    return '0 SOL'
  }
}



// 从交易回执获取Gas价格 - 使用统一格式化工具
const formatGasPriceFromReceipt = (receipt: any): string => {
  if (!receipt?.effective_gas_price) return 'N/A'
  
  let price: number
  if (typeof receipt.effective_gas_price === 'string') {
    // 智能检测：如果以0x开头，按十六进制解析；否则按十进制解析
    if (receipt.effective_gas_price.startsWith('0x')) {
      price = parseInt(receipt.effective_gas_price, 16)
    } else {
      price = parseInt(receipt.effective_gas_price, 10)
    }
  } else {
    price = receipt.effective_gas_price
  }
  
  if (isNaN(price) || price === 0) return '0 Gwei'
  
  // 使用统一格式化工具，转换为Gwei并添加单位
  return `${formatNumber.gwei(price)} Gwei`
}

// 从交易回执计算交易手续费 - 使用统一格式化工具
const formatTransactionFeeFromReceipt = (receipt: any): string => {
  if (!receipt?.effective_gas_price || !receipt?.gas_used) return 'N/A'
  
  let price: number
  if (typeof receipt.effective_gas_price === 'string') {
    // 智能检测：如果以0x开头，按十六进制解析；否则按十进制解析
    if (receipt.effective_gas_price.startsWith('0x')) {
      price = parseInt(receipt.effective_gas_price, 16)
    } else {
      price = parseInt(receipt.effective_gas_price, 10)
    }
  } else {
    price = receipt.effective_gas_price
  }
  
  const gasUsed = receipt.gas_used
  const fee = price * gasUsed
  
  if (fee === 0) return '0 SOL'
  
  // 使用统一格式化工具，转换为SOL并添加单位
  return `${formatNumber.weiToEth(fee)} SOL`
}

// 监听路由参数变化
onMounted(async () => {
  await loadBlockData()
  if (block.value) {
    await loadTransactions()
  }
})
</script>

<style scoped>
.card {
  @apply bg-white shadow-sm rounded-lg border border-gray-200 p-4;
}
</style>
