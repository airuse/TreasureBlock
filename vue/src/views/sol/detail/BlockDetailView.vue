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
                    <span class="text-gray-500">费用: </span>
                    <span class="font-medium">{{ formatSolFee(tx.fee) }} SOL</span>
                  </div>
                  <div>
                    <span class="text-gray-500">计算单元: </span>
                    <span class="text-gray-600">{{ tx.compute_units?.toLocaleString() || 'N/A' }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">版本: </span>
                    <span class="text-gray-600">{{ tx.version || 'N/A' }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">区块哈希: </span>
                    <span class="font-mono text-blue-600 cursor-pointer hover:text-blue-800" title="点击复制" @click="copyToClipboard(tx.blockhash, $event)">
                      {{ formatAddress(tx.blockhash) }}
                    </span>
                  </div>
                  <!-- 新增：签名者、价值变化、指令概要 -->
                  <div class="md:col-span-2">
                    <span class="text-gray-500">签名者: </span>
                    <span class="font-mono text-blue-600">{{ getSignerListText(tx.account_keys) }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">价值变化: </span>
                    <span class="text-gray-900 font-medium">{{ computeTotalValueText(tx.account_keys, tx.pre_balances, tx.post_balances) }}</span>
                  </div>
                  <div class="md:col-span-2">
                    <span class="text-gray-500">指令: </span>
                    <span class="text-gray-700">{{ summarizeInstructionsInline(tx.instructions) }}</span>
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

            <!-- 展开的交易详情信息 -->
            <div v-if="expandedTransactions[tx.tx_id || tx.hash]" class="mt-2 pt-2 border-t border-gray-200">
              <!-- 加载状态 -->
              <div v-if="loadingArtifacts[tx.tx_id || tx.hash]" class="text-center py-3">
                <div class="inline-flex items-center px-4 py-2 text-sm text-gray-600">
                  <svg class="animate-spin -ml-1 mr-3 h-4 w-4 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  加载交易详情中...
                </div>
              </div>
              
              <!-- 交易详情内容 -->
              <div v-else class="space-y-2">
                <h4 class="text-sm font-medium text-gray-900 border-b border-gray-200 pb-2">Sol 交易详情</h4>
                
                <!-- 交易基本信息 -->
                <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
                  <div>
                    <span class="text-gray-500">状态: </span>
                    <span :class="getStatusClass(tx.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full ml-2">
                      {{ getStatusText(tx.status) }}
                    </span>
                  </div>
                  <div>
                    <span class="text-gray-500">Slot: </span>
                    <span class="text-gray-600">{{ tx.slot?.toLocaleString() || 'N/A' }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">费用: </span>
                    <span class="text-gray-600">{{ formatSolFee(tx.fee) }} SOL</span>
                  </div>
                  <div>
                    <span class="text-gray-500">计算单元: </span>
                    <span class="text-gray-600">{{ tx.compute_units?.toLocaleString() || 'N/A' }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">版本: </span>
                    <span class="text-gray-600">{{ tx.version || 'N/A' }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">区块ID: </span>
                    <span class="text-gray-600">{{ tx.block_id || 'N/A' }}</span>
                  </div>
                </div>

                <!-- 交易费用详情 - Sol 专用 -->
                <div class="bg-gray-50 p-2 rounded-lg">
                  <h5 class="text-sm font-medium text-gray-900 mb-1">交易费用详情</h5>
                  
                  <!-- 主要费用信息 -->
                  <div class="grid grid-cols-1 md:grid-cols-3 gap-2 text-sm mb-3">
                    <div>
                      <span class="text-gray-500">交易手续费: </span>
                      <span class="text-gray-900 font-medium">{{ formatSolFee(tx.fee) }} SOL</span>
                    </div>
                    <div>
                      <span class="text-gray-500">计算单元: </span>
                      <span class="text-gray-900 font-medium">{{ tx.compute_units?.toLocaleString() || 'N/A' }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">版本: </span>
                      <span class="text-gray-600">{{ tx.version || 'N/A' }}</span>
                    </div>
                  </div>

                  <!-- Sol 特有信息 -->
                  <div class="border-t border-gray-200 pt-2">
                    <h6 class="text-sm font-medium text-gray-700 mb-2">Sol 交易信息</h6>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm">
                      <div>
                        <span class="text-gray-500">最近区块哈希: </span>
                        <span class="font-mono text-blue-600 cursor-pointer hover:text-blue-800" @click="copyToClipboard(tx.recent_blockhash, $event)">
                          {{ formatAddress(tx.recent_blockhash) }}
                        </span>
                      </div>
                      <div>
                        <span class="text-gray-500">区块哈希: </span>
                        <span class="font-mono text-blue-600 cursor-pointer hover:text-blue-800" @click="copyToClipboard(tx.blockhash, $event)">
                          {{ formatAddress(tx.blockhash) }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 交易属性 - Sol 特有信息 -->
                <div class="bg-gray-50 p-2 rounded-lg border-t border-gray-200">
                  <h5 class="text-sm font-medium text-gray-900 mb-1">交易属性</h5>
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm">
                    <div>
                      <span class="text-gray-500">版本: </span>
                      <span class="text-gray-600">{{ tx.version || 'N/A' }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">Slot: </span>
                      <span class="text-gray-600">{{ tx.slot?.toLocaleString() || 'N/A' }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">计算单元: </span>
                      <span class="text-gray-600">{{ tx.compute_units?.toLocaleString() || 'N/A' }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">状态: </span>
                      <span :class="getStatusClass(tx.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                        {{ getStatusText(tx.status) }}
                      </span>
                    </div>
                  </div>
                  
                  <!-- Sol 账户信息 -->
                  <div class="border-t border-gray-200 pt-2 mt-2">
                    <h6 class="text-sm font-medium text-gray-700 mb-2">账户信息</h6>
                    <div class="space-y-2">
                      <div v-if="tx.account_keys">
                        <span class="text-gray-500">账户密钥: </span>
                        <div class="mt-1 max-h-32 overflow-y-auto bg-white p-2 rounded border text-xs font-mono">
                          {{ formatAccountKeys(tx.account_keys) }}
                        </div>
                      </div>
                      <div v-if="tx.pre_balances">
                        <span class="text-gray-500">交易前余额: </span>
                        <div class="mt-1 max-h-32 overflow-y-auto bg-white p-2 rounded border text-xs font-mono">
                          {{ formatBalances(tx.pre_balances) }}
                        </div>
                      </div>
                      <div v-if="tx.post_balances">
                        <span class="text-gray-500">交易后余额: </span>
                        <div class="mt-1 max-h-32 overflow-y-auto bg-white p-2 rounded border text-xs font-mono">
                          {{ formatBalances(tx.post_balances) }}
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Sol 指令与事件（来自 getArtifactsByTxId API） -->
                  <div class="border-t border-gray-200 pt-2 mt-2">
                    <h6 class="text-sm font-medium text-gray-700 mb-2">Sol 指令与事件</h6>

                    <!-- 事件列表 -->
                    <div v-if="artifactsByTxId[tx.tx_id || tx.hash]" class="space-y-2">
                      <div v-if="(artifactsByTxId[tx.tx_id || tx.hash].events || []).length > 0" class="bg-white p-2 rounded border">
                        <h6 class="text-sm font-medium text-gray-700 mb-1">事件 ({{ artifactsByTxId[tx.tx_id || tx.hash].events.length }})</h6>
                        <div class="overflow-x-auto max-w-full">
                          <table class="w-full text-xs" style="min-width: 500px;">
                            <thead>
                              <tr class="text-left text-gray-500">
                                <th class="py-1 pr-4 w-8">#</th>
                                <th class="py-1 pr-4 w-20">类型</th>
                                <th class="py-1 pr-4 w-32">Program</th>
                                <th class="py-1 pr-4 min-w-[200px]">From</th>
                                <th class="py-1 pr-4 min-w-[200px]">To</th>
                                <th class="py-1 pr-4 w-20">Amount</th>
                              </tr>
                            </thead>
                            <tbody>
                              <tr v-for="(ev, idx) in artifactsByTxId[tx.tx_id || tx.hash].events" :key="idx" class="text-gray-700">
                                <td class="py-1 pr-4">{{ ev.event_index }}</td>
                                <td class="py-1 pr-4">{{ ev.event_type }}</td>
                                <td class="py-1 pr-4 font-mono text-xs break-all">{{ ev.program_id }}</td>
                                <td class="py-1 pr-4 font-mono text-xs break-all" :title="ev.from_address">{{ ev.from_address || '-' }}</td>
                                <td class="py-1 pr-4 font-mono text-xs break-all" :title="ev.to_address">{{ ev.to_address || '-' }}</td>
                                <td class="py-1 pr-4">{{ ev.amount }}</td>
                              </tr>
                            </tbody>
                          </table>
                        </div>
                      </div>

                      <!-- 指令列表 -->
                      <div v-if="(artifactsByTxId[tx.tx_id || tx.hash].instructions || []).length > 0" class="bg-white p-2 rounded border">
                        <h6 class="text-sm font-medium text-gray-700 mb-1">指令 ({{ artifactsByTxId[tx.tx_id || tx.hash].instructions.length }})</h6>
                        <div class="overflow-x-auto max-w-full">
                          <table class="w-full text-xs" style="min-width: 800px;">
                            <thead>
                              <tr class="text-left text-gray-500">
                                <th class="py-1 pr-4 w-16">Idx</th>
                                <th class="py-1 pr-4 w-32">Program</th>
                                <th class="py-1 pr-4 w-24">Type</th>
                                <th class="py-1 pr-4 w-24">Category</th>
                                <th class="py-1 pr-4 min-w-[200px]">Action</th>
                                <th class="py-1 pr-4 min-w-[200px]">Accounts</th>
                              </tr>
                            </thead>
                            <tbody>
                              <tr v-for="(ins, idx) in artifactsByTxId[tx.tx_id || tx.hash].instructions" :key="idx" class="text-gray-700">
                                <td class="py-1 pr-4">{{ ins.instruction_index }}</td>
                                <td class="py-1 pr-4 font-mono text-xs break-all">{{ ins.program_id }}</td>
                                <td class="py-1 pr-4">
                                  <div v-if="ins.program_id === 'ComputeBudget111111111111111111111111111111' && ins.data">
                                    <div class="font-medium text-blue-600">
                                      {{ parseComputeBudgetInstruction(ins.data)?.typeCn || ins.instruction_type || '-' }}
                                    </div>
                                    <div class="text-xs text-gray-500">
                                      {{ parseComputeBudgetInstruction(ins.data)?.type || '-' }}
                                    </div>
                                  </div>
                                  <div v-else>
                                    {{ ins.instruction_type || '-' }}
                                  </div>
                                </td>
                                <td class="py-1 pr-4">
                                  <span :class="getProgramTypeClass(getProgramType(ins.program_id).type)" class="px-2 py-1 text-xs font-semibold rounded-full">
                                    {{ getProgramType(ins.program_id).category }}
                                  </span>
                                </td>
                                <td class="py-1 pr-4 text-xs">
                                  <div v-if="ins.program_id === 'ComputeBudget111111111111111111111111111111' && ins.data">
                                    <div class="font-medium text-gray-900">
                                      {{ parseComputeBudgetInstruction(ins.data)?.actionCn || '-' }}
                                    </div>
                                    <div class="text-gray-500">
                                      {{ parseComputeBudgetInstruction(ins.data)?.action || '-' }}
                                    </div>
                                  </div>
                                  <div v-else class="text-gray-500">
                                    -
                                  </div>
                                </td>
                                <td class="py-1 pr-4 font-mono text-xs break-all" :title="ins.accounts">{{ ins.accounts }}</td>
                              </tr>
                            </tbody>
                          </table>
                        </div>
                      </div>

                      <div v-if="(artifactsByTxId[tx.tx_id || tx.hash].events || []).length === 0 && (artifactsByTxId[tx.tx_id || tx.hash].instructions || []).length === 0" class="text-sm text-gray-500">
                        暂无指令与事件
                      </div>
                    </div>
                  </div>

                  <!-- 注意：Sol 不需要解析合约转账，因为 Sol 使用不同的代币系统 -->

                  <!-- 指令数据 -->
                  <div class="border-t border-gray-200 pt-1 mt-1">
                    <h6 class="text-sm font-medium text-gray-700 mb-1">指令数据</h6>
                    <div class="bg-white p-1 rounded border overflow-x-auto max-w-full">
                      <pre class="text-xs text-gray-700 whitespace-pre-wrap break-all max-w-full">{{ formatSolInstructions(tx.instructions) }}</pre>
                    </div>
                  </div>

                  <!-- 解析后的指令详情 -->
                  <div v-if="artifactsByTxId[tx.tx_id || tx.hash] && (artifactsByTxId[tx.tx_id || tx.hash].instructions || []).length > 0" class="border-t border-gray-200 pt-2 mt-2">
                    <h6 class="text-sm font-medium text-gray-700 mb-2">指令详情解析</h6>
                    <div class="space-y-3">
                      <div v-for="(ins, idx) in artifactsByTxId[tx.tx_id || tx.hash].instructions" :key="idx" class="bg-gray-50 p-3 rounded border">
                        <div class="flex items-center justify-between mb-2">
                          <h7 class="text-sm font-medium text-gray-900">
                            #{{ ins.instruction_index }} - {{ getProgramType(ins.program_id).name }}
                            <span v-if="ins.program_id === 'ComputeBudget111111111111111111111111111111' && ins.data" class="text-blue-600">
                              : {{ parseComputeBudgetInstruction(ins.data)?.typeCn || '未知指令' }}
                            </span>
                          </h7>
                          <span :class="getProgramTypeClass(getProgramType(ins.program_id).type)" class="px-2 py-1 text-xs font-semibold rounded-full">
                            {{ getProgramType(ins.program_id).category }}
                          </span>
                        </div>
                        
                        <!-- 动作描述 -->
                        <div v-if="ins.program_id === 'ComputeBudget111111111111111111111111111111' && ins.data" class="mb-3">
                          <div class="text-sm text-gray-700">
                            <span class="font-medium">动作: </span>
                            <span class="text-blue-600">{{ parseComputeBudgetInstruction(ins.data)?.actionCn || '-' }}</span>
                          </div>
                          <div class="text-xs text-gray-500">
                            <span class="font-medium">Action: </span>
                            {{ parseComputeBudgetInstruction(ins.data)?.action || '-' }}
                          </div>
                        </div>
                        
                        <!-- 交互程序 -->
                        <div class="mb-3">
                          <div class="text-sm text-gray-700">
                            <span class="font-medium">交互程序: </span>
                            <span class="font-mono text-blue-600">{{ ins.program_id }}</span>
                          </div>
                          <div class="text-xs text-gray-500">
                            <span class="font-medium">Interact With: </span>
                            {{ getProgramType(ins.program_id).name }} - {{ ins.program_id }}
                          </div>
                        </div>
                        
                        <!-- 指令数据 JSON 结构 -->
                        <div v-if="ins.program_id === 'ComputeBudget111111111111111111111111111111' && ins.data && parseComputeBudgetInstruction(ins.data)" class="bg-white p-2 rounded border">
                          <div class="text-sm font-medium text-gray-700 mb-1">指令数据 (Instruction Data)</div>
                          <div class="text-xs font-mono text-gray-600">
                            <div class="space-y-1">
                              <div v-for="(value, key) in (parseComputeBudgetInstruction(ins.data)?.data || {})" :key="key" class="flex items-start">
                                <span class="text-blue-600 mr-2">"{{ key }}":</span>
                                <div class="flex-1">
                                  <span class="text-gray-500">{"</span>
                                  <span class="text-green-600">type</span>
                                  <span class="text-gray-500">": "</span>
                                  <span class="text-purple-600">{{ (value as any)?.type || 'unknown' }}</span>
                                  <span class="text-gray-500">", "</span>
                                  <span class="text-green-600">data</span>
                                  <span class="text-gray-500">": </span>
                                  <span class="text-orange-600">{{ (value as any)?.data || 'null' }}</span>
                                  <span class="text-gray-500">}</span>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>
                        
                        <!-- 原始数据 -->
                        <div v-if="ins.data" class="bg-white p-2 rounded border">
                          <div class="text-sm font-medium text-gray-700 mb-1">原始数据 (Raw Data)</div>
                          <div class="text-xs font-mono text-gray-600 break-all">{{ ins.data }}</div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 交易日志 -->
                <div v-if="tx.logs" class="bg-gray-50 p-2 rounded-lg">
                  <h5 class="text-sm font-medium text-gray-900 mb-1">交易日志</h5>
                  <div class="bg-white p-1 rounded border overflow-x-auto max-w-full">
                    <pre class="text-xs text-gray-700 whitespace-pre-wrap break-all max-w-full">{{ formatSolLogs(tx.logs) }}</pre>
                  </div>
                </div>
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
import { sol as solApi } from '@/api'
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
// sol 指令/事件（后端 artifacts）
const artifactsByTxId = ref<Record<string, { events: any[]; instructions: any[] }>>({})
const loadingArtifacts = ref<Record<string, boolean>>({})

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

// 格式化 Sol 费用（从 lamports 转换为 SOL）
const formatSolFee = (fee: number | string) => {
  if (!fee) return '0'
  
  let num: number
  if (typeof fee === 'string') {
    num = parseFloat(fee)
  } else {
    num = fee
  }
  
  if (isNaN(num) || num === 0) return '0'
  
  // Sol 费用是 lamports，需要转换为 SOL (1 SOL = 10^9 lamports)
  return formatNumber.weiToEth(num)
}

// 格式化账户密钥
const formatAccountKeys = (accountKeys: string) => {
  try {
    const keys = JSON.parse(accountKeys)
    if (Array.isArray(keys)) {
      return keys.map((key, index) => 
        `${index}: ${key.pubkey} (${key.signer ? '签名者' : '非签名者'}, ${key.writable ? '可写' : '只读'})`
      ).join('\n')
    }
    return accountKeys
  } catch (error) {
    return accountKeys
  }
}

// 解析签名者列表（简短展示）
const getSignerListText = (accountKeys: string | any) => {
  try {
    const keys = typeof accountKeys === 'string' ? JSON.parse(accountKeys) : accountKeys
    if (!Array.isArray(keys)) return '-'
    const signers: string[] = keys.filter((k: any) => k && k.signer).map((k: any) => k.pubkey)
    if (signers.length === 0) return '-'
    // 只展示前2个，更多用 +N 表示
    const head = signers.slice(0, 2)
    const suffix = signers.length > 2 ? ` +${signers.length - 2}` : ''
    return [...head].join(', ') + suffix
  } catch {
    return '-'
  }
}

// 计算主签名者的价值变化（lamports -> SOL）
const computeTotalValueText = (accountKeys: string | any, preBalances: string | any, postBalances: string | any) => {
  try {
    const keys = typeof accountKeys === 'string' ? JSON.parse(accountKeys) : accountKeys
    const pre = typeof preBalances === 'string' ? JSON.parse(preBalances) : preBalances
    const post = typeof postBalances === 'string' ? JSON.parse(postBalances) : postBalances
    if (!Array.isArray(keys) || !Array.isArray(pre) || !Array.isArray(post) || pre.length !== post.length) return 'N/A'
    const mainSignerIndex = keys.findIndex((k: any) => k && k.signer)
    if (mainSignerIndex < 0 || mainSignerIndex >= pre.length) return 'N/A'
    const delta = new BigNumber(post[mainSignerIndex]).minus(new BigNumber(pre[mainSignerIndex]))
    const sol = formatNumber.weiToEth(delta.abs().toString())
    const sign = delta.isGreaterThan(0) ? '+' : delta.isLessThan(0) ? '-' : ''
    return `${sign}${sol} SOL`
  } catch {
    return 'N/A'
  }
}

// 已知的Solana程序类型映射
const PROGRAM_TYPES = {
  // 系统程序
  '11111111111111111111111111111111': { name: 'System Program', type: 'system', category: '系统程序' },
  'Vote111111111111111111111111111111111111111': { name: 'Vote Program', type: 'system', category: '系统程序' },
  'ComputeBudget111111111111111111111111111111': { name: 'Compute Budget Program', type: 'system', category: '系统程序' },
  'Config1111111111111111111111111111111111111': { name: 'Config Program', type: 'system', category: '系统程序' },
  'Stake11111111111111111111111111111111111111': { name: 'Stake Program', type: 'system', category: '系统程序' },
  'AddressLookupTab1e1111111111111111111111111': { name: 'Address Lookup Table Program', type: 'system', category: '系统程序' },
  
  // SPL代币程序
  'TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA': { name: 'SPL Token Program', type: 'spl-token', category: '代币程序' },
  'ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL': { name: 'Associated Token Program', type: 'spl-token', category: '代币程序' },
  'MemoSq4gqABAXKb96qnH8TysKcWfC85B2q2': { name: 'Memo Program', type: 'memo', category: '备忘录程序' },
  
  // 已知的第三方程序（可根据需要扩展）
  'dRiftyHA39MWEi3m9aunc5MzRF1JYuBsbn6VPcn33UH': { name: 'Drift Protocol', type: 'defi', category: 'DeFi协议' },
  'JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4': { name: 'Jupiter', type: 'defi', category: 'DeFi协议' },
  '9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM': { name: 'Raydium', type: 'defi', category: 'DeFi协议' },
  'So1endDq2YkqhipRh3WViPa8hdiSpxWy6z3Z6tMCpAo': { name: 'Solend', type: 'defi', category: 'DeFi协议' },
}

// 解析 ComputeBudget 程序指令数据（自动检测 hex/base58，浏览器环境友好）
const parseComputeBudgetInstruction = (data: string) => {
  try {
    if (!data) return null

    // 工具: hex/base58 判定与解码
    const isHexLike = (s: string): boolean => {
      const t = s.startsWith('0x') ? s.slice(2) : s
      return t.length % 2 === 0 && /^[0-9a-fA-F]+$/.test(t)
    }

    const hexToBytes = (hex: string): Uint8Array => {
      const t = hex.startsWith('0x') ? hex.slice(2) : hex
      const out = new Uint8Array(t.length / 2)
      for (let i = 0; i < out.length; i++) {
        out[i] = parseInt(t.substr(i * 2, 2), 16)
      }
      return out
    }

    // Base58 (Bitcoin/Solana alphabet)
    const decodeBase58 = (s: string): Uint8Array => {
      const alphabet = '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz'
      const base = BigInt(58)
      let num = BigInt(0)
      for (const ch of s) {
        const idx = alphabet.indexOf(ch)
        if (idx === -1) throw new Error('Invalid base58 character')
        num = num * base + BigInt(idx)
      }
      // Count leading zeros
      let leadingZeros = 0
      for (const ch of s) {
        if (ch === '1') leadingZeros++
        else break
      }
      // Convert BigInt to bytes (big-endian), then add leading zeros
      const bytesBE: number[] = []
      while (num > 0) {
        bytesBE.push(Number(num % BigInt(256)))
        num = num / BigInt(256)
      }
      const bytes = new Uint8Array(leadingZeros + bytesBE.length)
      // prepend zeros, then reversed remainder
      for (let i = 0; i < leadingZeros; i++) bytes[i] = 0
      for (let i = 0; i < bytesBE.length; i++) bytes[leadingZeros + i] = bytesBE[bytesBE.length - 1 - i]
      return bytes
    }

    const bytes: Uint8Array = isHexLike(data) ? hexToBytes(data) : decodeBase58(data)
    if (!bytes || bytes.length === 0) return null

    const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength)
    const discriminator = view.getUint8(0)

    switch (discriminator) {
      case 2: { // SetComputeUnitLimit
        if (bytes.length >= 5) {
          const units = view.getUint32(1, true)
          return {
            type: 'SetComputeUnitLimit',
            typeCn: '设置计算单元限制',
            action: `设置 ${units.toLocaleString()} 计算单位`,
            actionCn: `设置 ${units.toLocaleString()} 计算单位`,
            data: {
              discriminator: { type: 'u8', data: discriminator },
              units: { type: 'u32', data: units }
            }
          }
        }
        break
      }
      case 3: { // SetComputeUnitPrice
        if (bytes.length >= 9) {
          const microLamports = view.getBigUint64(1, true)
          const lamports = Number(microLamports) / 1_000_000
          return {
            type: 'SetComputeUnitPrice',
            typeCn: '设置计算单元价格',
            action: `设置 ${lamports.toFixed(6)} lamports 每个计算单元`,
            actionCn: `设置 ${lamports.toFixed(6)} lamports 每个计算单元`,
            data: {
              discriminator: { type: 'u8', data: discriminator },
              microLamports: { type: 'u64', data: Number(microLamports) }
            }
          }
        }
        break
      }
      default:
        return {
          type: 'Unknown',
          typeCn: '未知指令',
          action: `未知指令类型 (${discriminator})`,
          actionCn: `未知指令类型 (${discriminator})`,
          data: {
            discriminator: { type: 'u8', data: discriminator }
          }
        }
    }
  } catch (error) {
    console.error('解析 ComputeBudget 指令失败:', error, data)
    return null
  }

  return null
}

// 获取程序类型信息
const getProgramType = (programId: string) => {
  const info = (PROGRAM_TYPES as any)[programId]
  if (info) return info
  
  // 如果不在已知列表中，根据地址特征判断
  if (programId.length === 44) {
    return { name: 'Custom Program', type: 'custom', category: '个人程序' }
  }
  
  return { name: 'Unknown Program', type: 'unknown', category: '未知程序' }
}

// 获取程序类型样式类
const getProgramTypeClass = (type: string) => {
  const typeClasses: Record<string, string> = {
    'system': 'bg-blue-100 text-blue-800',
    'spl-token': 'bg-green-100 text-green-800',
    'memo': 'bg-gray-100 text-gray-800',
    'defi': 'bg-purple-100 text-purple-800',
    'custom': 'bg-orange-100 text-orange-800',
    'unknown': 'bg-gray-100 text-gray-600'
  }
  return typeClasses[type] || typeClasses['unknown']
}

// 指令简要汇总（Program + type/数据长度）
const summarizeInstructionsInline = (instructions: string | any) => {
  try {
    const list = typeof instructions === 'string' ? JSON.parse(instructions) : instructions
    if (!Array.isArray(list) || list.length === 0) return '无'
    const items = list.slice(0, 3).map((ins: any) => {
      const prog = ins.program_id || 'unknown'
      const progInfo = getProgramType(prog)
      const t = ins.instruction_type || ins.type || ''
      const dataLen = (ins.data || '').length
      
      // 显示程序类型和指令类型
      const typeInfo = progInfo.type !== 'unknown' ? `[${progInfo.category}]` : ''
      const instructionInfo = t ? `:${t}` : ''
      const dataInfo = dataLen ? `(${dataLen})` : ''
      
      return `${progInfo.name}${typeInfo}${instructionInfo}${dataInfo}`
    })
    const more = list.length > 3 ? `, +${list.length - 3} more` : ''
    return items.join(', ') + more
  } catch {
    return '无'
  }
}

// 格式化余额数组
const formatBalances = (balances: string) => {
  try {
    const balanceArray = JSON.parse(balances)
    if (Array.isArray(balanceArray)) {
      return balanceArray.map((balance, index) => 
        `${index}: ${formatSolFee(balance)} SOL`
      ).join('\n')
    }
    return balances
  } catch (error) {
    return balances
  }
}

const formatHash = (hash: string) => {
  if (!hash) return 'N/A'
  return `${hash.substring(0, 10)}...${hash.substring(hash.length - 10)}`
}

const getStatusClass = (status: string | number) => {
  const statusStr = typeof status === 'string' ? status.toLowerCase() : status.toString()
  switch (statusStr) {
    case 'success':
    case '1':
      return 'bg-green-100 text-green-800'
    case 'failed':
    case '2':
      return 'bg-red-100 text-red-800'
    case 'pending':
    case '0':
      return 'bg-gray-100 text-gray-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

const getStatusText = (status: string | number) => {
  const statusStr = typeof status === 'string' ? status.toLowerCase() : status.toString()
  switch (statusStr) {
    case 'success':
    case '1':
      return 'Success'
    case 'failed':
    case '2':
      return 'Failed'
    case 'pending':
    case '0':
      return 'Pending'
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
    console.log('🔍 开始加载 Sol 交易数据:', {
      slot: parseInt(blockHeight.value),
      page: currentPage.value,
      page_size: pageSize.value
    })
    
    // 改为调用 Sol 专属分页接口
    const resp = await solApi.listTxDetails({
      slot: parseInt(blockHeight.value),
      page: currentPage.value,
      page_size: pageSize.value,
    })
    
    console.log('📊 Sol API 响应:', resp)
    
    if (resp && (resp as any).success === true) {
      const list = (resp as any).data || []
      const meta = (resp as any).meta || {}
      console.log('✅ 成功获取交易数据:', { list: list.length, meta })
      transactions.value = list
      totalCount.value = meta.total || list.length || 0
      totalPages.value = Math.max(1, Math.ceil(totalCount.value / pageSize.value))
    } else {
      console.error('❌ API 响应失败:', resp)
      throw new Error((resp as any)?.message || '获取交易信息失败')
    }
  } catch (error) {
    console.error('❌ 加载交易失败:', error)
    console.error('错误详情:', {
      message: error instanceof Error ? error.message : 'Unknown error',
      stack: error instanceof Error ? error.stack : undefined,
      response: error
    })
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
  
  // 如果展开交易详情
  if (!isExpanded) {
    console.log('🔍 展开交易详情:', txHash)
    
    // 加载 sol 指令/事件（优先加载，无需鉴权）
    if (!artifactsByTxId.value[txHash] && !loadingArtifacts.value[txHash]) {
      await loadArtifacts(txHash)
    }
    
    // 对于 Sol，不需要加载交易凭证和解析结果，因为信息已经包含在 /api/v1/sol/tx/detail 中
    // Sol 使用不同的数据结构和解析方式
  }
}

// 注意：Sol 不需要加载交易凭证，因为信息已经包含在 /api/v1/sol/tx/detail 中

// 注意：Sol 不需要加载交易解析结果，因为信息已经包含在 /api/v1/sol/tx/detail 中

// 加载 sol artifacts（events + instructions）
const loadArtifacts = async (txHash: string) => {
  try {
    loadingArtifacts.value[txHash] = true
    console.log('🔍 开始加载 Sol artifacts:', txHash)
    
    const resp = await solApi.getArtifactsByTxId(txHash)
    console.log('📊 Sol artifacts API 响应:', resp)
    
    if (resp && (resp as any).success === true) {
      artifactsByTxId.value[txHash] = (resp as any).data || { events: [], instructions: [] }
      console.log('✅ 成功加载 Sol artifacts:', artifactsByTxId.value[txHash])
    } else {
      console.warn('❌ Sol artifacts API 响应失败:', resp)
      artifactsByTxId.value[txHash] = { events: [], instructions: [] }
    }
  } catch (e) {
    console.error('❌ 加载 Sol artifacts 失败:', e)
    artifactsByTxId.value[txHash] = { events: [], instructions: [] }
  } finally {
    loadingArtifacts.value[txHash] = false
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

// 格式化 Sol 日志
const formatSolLogs = (logs: string) => {
  try {
    const logArray = JSON.parse(logs)
    if (Array.isArray(logArray)) {
      return logArray.join('\n')
    }
    return logs
  } catch (error) {
    return logs || 'No logs data'
  }
}

// 格式化 Sol 指令
const formatSolInstructions = (instructions: string) => {
  try {
    const instructionArray = JSON.parse(instructions)
    if (Array.isArray(instructionArray)) {
      return instructionArray.map((inst, index) => 
        `指令 ${index}:\n` +
        `  程序ID: ${inst.program_id}\n` +
        `  数据: ${inst.data}\n` +
        `  账户: ${inst.accounts ? inst.accounts.join(', ') : '无'}\n` +
        `  类型: ${inst.type || '未知'}\n` +
        `  内部指令: ${inst.is_inner ? '是' : '否'}\n`
      ).join('\n')
    }
    return instructions
  } catch (error) {
    return instructions || 'No instructions data'
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


// 解析交易输入数据（使用parser_configs）- 保留用于其他链
const parseInputDataWithConfig = (inputData: string, txHash?: string) => {
  if (!inputData || inputData === '0x') return '0x (No input data)'
  
  // 对于 Sol，直接返回原始数据，因为 Sol 使用不同的指令格式
  return inputData
}

// 解析交易日志（使用parser_configs）- 保留用于其他链
const parseLogsDataWithConfig = (logsData: string, txHash?: string) => {
  if (!logsData) return 'No logs data'
  
  // 对于 Sol，使用专门的日志格式化函数
  return formatSolLogs(logsData)
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
    
    // 对于 Sol，不需要计算节省费用，因为 Sol 使用不同的费用模型
    return 'N/A (Sol 使用不同的费用模型)'
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
