<template>
  <div class="space-y-6">
    <!-- 页面头部 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="text-2xl font-bold text-gray-900">交易历史</h1>
            <p class="text-sm text-gray-500">查看和管理您的交易记录</p>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-3 h-3 bg-orange-500 rounded-full"></div>
            <span class="text-sm text-gray-600">BTC 网络</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 实时费率信息 -->
    <div v-if="feeLevels" class="bg-white shadow rounded-lg">
      <div class="px-4 py-3">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-lg leading-6 font-medium text-gray-900">实时费率信息</h3>
          <div class="text-sm text-gray-500">
            最后更新: {{ formatTime(new Date(feeLevels.normal.last_updated * 1000)) }}
          </div>
        </div>
        <div class="flex flex-col lg:flex-row gap-3">
          <!-- 左侧：费率信息 -->
          <div class="lg:w-80 flex-shrink-0">
            <div class="space-y-1.5">
              <!-- 慢速费率 -->
              <div class="border border-gray-200 rounded-lg p-2.5">
                <div class="flex items-center justify-between mb-1">
                  <h4 class="text-sm font-medium text-gray-900">慢速</h4>
                  <span class="text-xs text-gray-500">20% 分位</span>
                </div>
                <div class="text-sm text-gray-600">
                  <span class="font-mono">{{ feeLevels.slow.max_priority_fee }} sat/vB</span>
                </div>
              </div>
              
              <!-- 普通费率 -->
              <div class="border border-orange-200 bg-orange-50 rounded-lg p-2.5">
                <div class="flex items-center justify-between mb-1">
                  <h4 class="text-sm font-medium text-orange-900">普通</h4>
                  <span class="text-xs text-orange-600">50% 分位</span>
                </div>
                <div class="text-sm text-orange-800">
                  <span class="font-mono">{{ feeLevels.normal.max_priority_fee }} sat/vB</span>
                </div>
              </div>
              
              <!-- 快速费率 -->
              <div class="border border-gray-200 rounded-lg p-2.5">
                <div class="flex items-center justify-between mb-1">
                  <h4 class="text-sm font-medium text-gray-900">快速</h4>
                  <span class="text-xs text-gray-500">80% 分位</span>
                </div>
                <div class="text-sm text-gray-600">
                  <span class="font-mono">{{ feeLevels.fast.max_priority_fee }} sat/vB</span>
                </div>
              </div>
            </div>
          </div>
          
          <!-- 右侧：趋势图 -->
          <div class="flex-1 min-w-0">
            <!-- 费率趋势图 -->
            <div class="space-y-4">
              <!-- 费率图表 -->
              <div class="relative">
                <div class="text-sm font-medium text-gray-700 mb-2">费率趋势</div>
                <div class="h-32">
                  <canvas ref="feeChartCanvas" class="w-full h-full cursor-crosshair"></canvas>
                </div>
                <!-- 费率工具提示 -->
                <div 
                  ref="feeTooltip" 
                  class="absolute bg-gray-800 text-white text-xs px-2 py-1 rounded shadow-lg pointer-events-none opacity-0 transition-opacity duration-200 z-10"
                  style="transform: translate(-50%, -100%); margin-top: -8px;"
                >
                  <div class="font-medium">费率</div>
                  <div class="text-gray-300">Value: <span class="text-white font-mono" id="tooltip-fee-value">0</span> sat/vB</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 交易概览 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">交易概览</h3>
        <div class="grid grid-cols-1 md:grid-cols-5 gap-6">
          <div class="text-center">
            <div class="text-2xl font-bold text-gray-600">{{ totalTransactions }}</div>
            <div class="text-sm text-gray-500">总交易</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-yellow-600">{{ unsignedCount }}</div>
            <div class="text-sm text-gray-500">未签名</div>
          </div>
          <div class="text-center">
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-orange-600">{{ inProgressCount }}</div>
            <div class="text-sm text-gray-500">在途</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-green-600">{{ confirmedCount }}</div>
            <div class="text-sm text-gray-500">已确认</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 交易列表 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg leading-6 font-medium text-gray-900">交易历史</h3>
          <div class="flex space-x-2">
            <select v-model="selectedStatus" class="border border-gray-300 rounded-md px-3 py-2 text-sm">
              <option value="">全部状态</option>
              <option value="draft">草稿</option>
              <option value="unsigned">未签名</option>
              <option value="in_progress">在途</option>
              <option value="packed">已打包</option>
              <option value="confirmed">已确认</option>
              <option value="failed">失败</option>
            </select>
            <button
              @click="openCreateModal"
              class="px-4 py-2 bg-orange-600 text-white text-sm font-medium rounded-md hover:bg-orange-700 transition-colors"
            >
              新建交易
            </button>
          </div>
        </div>
        
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易哈希</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">发送地址</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">接收地址</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">金额 (BTC)</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">费率 (sat/vB)</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">创建时间</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="tx in filteredTransactions" :key="tx.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono">
                    {{ tx.tx_hash ? tx.tx_hash.substring(0, 10) + '...' + tx.tx_hash.substring(tx.tx_hash.length - 8) : '未生成' }}
                  </code>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono">
                    {{ tx.from_address.substring(0, 10) }}...{{ tx.from_address.substring(tx.from_address.length - 8) }}
                  </code>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono">
                    {{ tx.to_address.substring(0, 10) }}...{{ tx.to_address.substring(tx.to_address.length - 8) }}
                  </code>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {{ formatBtcAmount(tx.amount) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ getTransactionFeeRate(tx) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="getStatusClass(tx.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ getStatusText(tx.status) }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ formatTime(tx.created_at) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                  <button
                    v-if="tx.status === 'draft'"
                    @click="editTransaction(tx)"
                    class="text-indigo-600 hover:text-indigo-900"
                  >
                    编辑
                  </button>
                  <button
                    v-if="tx.status === 'draft' || tx.status === 'unsigned'"
                    @click="handleExportTransaction(tx)"
                    class="text-blue-600 hover:text-blue-900"
                  >
                    导出交易
                  </button>
                  <button
                    v-if="tx.status === 'unsigned'"
                    @click="openImportSignatureModal(tx)"
                    class="text-teal-600 hover:text-teal-900"
                  >
                    导入签名
                  </button>
                  <button
                    v-if="tx.status === 'in_progress'"
                    @click="sendTransactionToChain(tx.id)"
                    class="text-green-600 hover:text-green-900"
                  >
                    发送交易
                  </button>
                  <button
                    v-if="tx.status === 'in_progress' || tx.status === 'packed'"
                    @click="viewTransaction(tx)"
                    class="text-purple-600 hover:text-purple-900"
                  >
                    查看详情
                  </button>
                  <button
                    v-if="tx.status === 'confirmed'"
                    @click="viewTransaction(tx)"
                    class="text-gray-600 hover:text-gray-900"
                  >
                    查看详情
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 分页 -->
        <div class="flex items-center justify-between mt-4">
          <div class="text-sm text-gray-700">
            显示第 {{ (currentPage - 1) * pageSize + 1 }} 到 {{ Math.min(currentPage * pageSize, totalItems) }} 条，共 {{ totalItems }} 条记录
          </div>
          <div class="flex space-x-2">
            <button
              @click="prevPage"
              :disabled="currentPage === 1"
              class="px-3 py-2 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              上一页
            </button>
            <button
              @click="nextPage"
              :disabled="currentPage >= totalPages"
              class="px-3 py-2 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              下一页
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 新建交易模态框 -->
    <teleport to="body">
    <div v-if="showCreateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[1000]">
      <div class="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-medium text-gray-900">
            {{ isEditMode ? '编辑交易（BTC - UTXO）' : '新建交易（BTC - UTXO）' }}
          </h3>
        </div>
        <div class="px-6 py-4 space-y-6">
          <!-- 发送地址选择 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">发送地址</label>
            <select
              v-model="createForm.fromAddress"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
            >
              <option value="">请选择一个BTC地址</option>
              <option v-for="addr in btcAddresses" :key="addr.id" :value="addr.address">
                {{ addr.label ? addr.label + ' - ' : '' }}{{ addr.address }}
                {{ addr.balance ? '（余额 ' + addr.balance + '）' : '' }}
              </option>
            </select>
          </div>

          <!-- UTXO选择 -->
          <div v-if="createForm.fromAddress" class="space-y-2">
            <div class="flex items-center justify-between">
              <div class="text-sm font-medium text-gray-700">选择要解锁的UTXO</div>
              <div class="text-xs text-gray-500">
                已选 {{ numInputs }} 个 | 输入总额：{{ toBtc(selectedInputsTotalSats).toFixed(8) }} BTC
              </div>
            </div>
            <div class="text-xs text-orange-600 bg-orange-50 border border-orange-200 rounded p-2">
              <div class="flex items-center">
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 19.5c-.77.833.192 2.5 1.732 2.5z" />
                </svg>
                状态为"打包中"的UTXO不可选择，因为它们正在被其他交易使用
              </div>
            </div>
            <div class="overflow-x-auto overflow-y-auto max-h-48 border border-gray-200 rounded-md">
              <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                  <tr>
                    <th class="px-4 py-2 text-left text-xs font-medium text-gray-500">选择</th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-gray-500">TXID:VOUT</th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-gray-500">
                      <button class="inline-flex items-center space-x-1 hover:text-gray-700" @click="toggleUtxoSort('value_satoshi')">
                        <span>金额 (BTC)</span>
                        <span class="text-[10px]" :class="{ 'text-gray-900': utxoSort.key === 'value_satoshi', 'text-gray-300': utxoSort.key !== 'value_satoshi' }">
                          {{ utxoSort.key === 'value_satoshi' ? (utxoSort.order === 'asc' ? '▲' : '▼') : '▲' }}
                        </span>
                      </button>
                    </th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-gray-500">
                      <button class="inline-flex items-center space-x-1 hover:text-gray-700" @click="toggleUtxoSort('block_height')">
                        <span>区块高度</span>
                        <span class="text-[10px]" :class="{ 'text-gray-900': utxoSort.key === 'block_height', 'text-gray-300': utxoSort.key !== 'block_height' }">
                          {{ utxoSort.key === 'block_height' ? (utxoSort.order === 'asc' ? '▲' : '▼') : '▲' }}
                        </span>
                      </button>
                    </th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-gray-500">状态</th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-gray-500">Coinbase</th>
                  </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                  <tr v-for="u in sortedUtxos" :key="u.id" class="hover:bg-gray-50" :class="{ 'opacity-50 bg-gray-100': u.status === 'spent' }">
                    <td class="px-4 py-2">
                      <input 
                        type="checkbox" 
                        :checked="selectedUtxoIds.has(u.id)" 
                        :disabled="u.status === 'spent'"
                        @change="toggleUtxo(u)" 
                        :class="{ 'cursor-not-allowed': u.status === 'spent' }"
                      />
                    </td>
                    <td class="px-4 py-2 text-xs font-mono">
                      {{ u.tx_id.substring(0, 10) }}...{{ u.tx_id.substring(u.tx_id.length - 8) }}:{{ u.vout_index }}
                    </td>
                    <td class="px-4 py-2 text-sm">
                      {{ toBtc(u.value_satoshi).toFixed(8) }}
                    </td>
                    <td class="px-4 py-2 text-sm">{{ u.block_height }}</td>
                    <td class="px-4 py-2 text-sm">
                      <span v-if="u.status === 'spent'" class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-orange-100 text-orange-800">
                        打包中
                      </span>
                      <span v-else-if="u.spent_tx_id" class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800">
                        已花费
                      </span>
                      <span v-else class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                        可用
                      </span>
                    </td>
                    <td class="px-4 py-2 text-sm">{{ u.is_coinbase ? '是' : '否' }}</td>
                  </tr>
                  <tr v-if="utxos.length === 0">
                    <td colspan="6" class="px-4 py-6 text-center text-sm text-gray-500">暂无UTXO</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- 输出编辑器 -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <div class="text-sm font-medium text-gray-700">交易输出</div>
              <div class="flex items-center space-x-4">
                <label class="inline-flex items-center text-sm text-gray-700 select-none">
                  <input type="checkbox" v-model="autoChange" class="mr-2" /> 自动找零
                </label>
                <button @click="addOutput" class="text-sm text-orange-600 hover:text-orange-700">+ 添加输出</button>
              </div>
            </div>
            <div class="space-y-3 min-h-40 max-h-64 overflow-y-auto">
              <div v-for="(out, idx) in createForm.outputs" :key="idx" class="grid grid-cols-1 md:grid-cols-10 gap-3 items-center">
                <div class="md:col-span-6 relative">
                  <input
                    v-model.trim="out.toAddress"
                    type="text"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                    placeholder="接收地址（支持自动补全：输入我的地址前缀或备注）"
                    @focus="showSuggestIndex = idx"
                    @input="showSuggestIndex = idx"
                    @blur="onAddressBlur()"
                  />
                  <div
                    v-if="showSuggestIndex === idx && addressSuggestions(out.toAddress).length > 0"
                    class="absolute z-50 mt-1 w-full bg-white border border-gray-200 rounded-md shadow-lg max-h-56 overflow-auto"
                  >
                    <div
                      v-for="addr in addressSuggestions(out.toAddress)"
                      :key="addr.id"
                      class="px-3 py-2 hover:bg-orange-50 cursor-pointer text-sm"
                      @mousedown.prevent="selectSuggestion(idx, addr.address)"
                    >
                      <div class="flex items-center justify-between">
                        <span class="font-mono truncate">{{ addr.address }}</span>
                        <span v-if="addr.label" class="text-xs text-gray-500 ml-2 truncate">{{ addr.label }}</span>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="md:col-span-3">
                  <input
                    v-model.number="out.amountBtc"
                    type="number"
                    step="0.00000001"
                    min="0"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                    placeholder="金额 (BTC)"
                  />
                </div>
                <div class="md:col-span-1 flex justify-end">
                  <button @click="removeOutput(idx)" class="px-2 py-2 text-red-600 hover:text-red-700" title="删除输出">
                    删除
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- 手续费与汇总 -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 items-start">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">手续费档位</label>
              <select
                v-model="createForm.feeTier"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
              >
                <option value="slow">慢速 (20%) - {{ feeLevels?.slow?.max_priority_fee || '-' }} sat/vB</option>
                <option value="normal">普通 (50%) - {{ feeLevels?.normal?.max_priority_fee || '-' }} sat/vB</option>
                <option value="fast">快速 (80%) - {{ feeLevels?.fast?.max_priority_fee || '-' }} sat/vB</option>
              </select>
              <div class="text-xs text-gray-500 mt-1">基于分位建议值估算交易费。</div>
            </div>
            <div class="md:col-span-2">
              <div class="bg-gray-50 border border-gray-200 rounded-md p-3 text-sm space-y-1">
                <div class="flex justify-between"><span>输入总额</span><span>{{ toBtc(selectedInputsTotalSats).toFixed(8) }} BTC</span></div>
                <div class="flex justify-between"><span>输出总额</span><span>{{ toBtc(outputsTotalSats).toFixed(8) }} BTC</span></div>
                <div class="flex justify-between"><span>预计费率</span><span>{{ satPerVb }} sat/vB</span></div>
                <div class="flex justify-between"><span>预计大小</span><span>{{ estimatedVBytes }} vB</span></div>
                <div class="flex justify-between"><span>预计手续费</span><span>{{ toBtc(estimatedFeeSats).toFixed(8) }} BTC ({{ estimatedFeeSats }} sats)</span></div>
                <div class="flex justify-between font-medium" :class="{ 'text-red-600': !isBalanceEnough, 'text-green-700': isBalanceEnough }">
                  <span>找零</span><span>{{ autoChange ? toBtc(Math.max(changeSats, 0)).toFixed(8) + ' BTC' : '关闭' }}</span>
                </div>
                <div v-if="!isBalanceEnough" class="text-xs text-red-600 mt-1">余额不足：请选择更多UTXO或减少输出金额。</div>
              </div>
            </div>
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="closeCreateModal"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            取消
          </button>
          <button
            @click="submitCreateTransaction"
            :disabled="!isCreateFormValid"
            class="px-4 py-2 bg-orange-600 text-white rounded-md hover:bg-orange-700 disabled:opacity-50"
          >
            {{ isEditMode ? '更新' : '创建' }}
          </button>
        </div>
      </div>
    </div>
    </teleport>

    <!-- 导出费率选择模态框 -->
    <teleport to="body">
    <div v-if="showExportFeeModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[1000]">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-medium text-gray-900">选择导出费率</h3>
        </div>
        <div class="px-6 py-4 space-y-4">
          <!-- 费率模式选择 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">费率模式</label>
            <div class="flex space-x-4">
              <label class="flex items-center">
                <input
                  type="radio"
                  v-model="exportFeeMode"
                  value="auto"
                  class="mr-2 text-orange-600"
                />
                <span class="text-sm text-gray-700">自动模式</span>
              </label>
              <label class="flex items-center">
                <input
                  type="radio"
                  v-model="exportFeeMode"
                  value="manual"
                  class="mr-2 text-orange-600"
                />
                <span class="text-sm text-gray-700">手动模式</span>
              </label>
            </div>
          </div>

          <!-- 自动模式 -->
          <div v-if="exportFeeMode === 'auto'">
            <label class="block text-sm font-medium text-gray-700 mb-2">费率档位</label>
            <select
              v-model="exportFeeTier"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
            >
              <option value="slow">慢速 (20%) - {{ feeLevels?.slow?.max_priority_fee || '-' }} sat/vB</option>
              <option value="normal">普通 (50%) - {{ feeLevels?.normal?.max_priority_fee || '-' }} sat/vB</option>
              <option value="fast">快速 (80%) - {{ feeLevels?.fast?.max_priority_fee || '-' }} sat/vB</option>
            </select>
          </div>

          <!-- 手动模式 -->
          <div v-if="exportFeeMode === 'manual'">
            <label class="block text-sm font-medium text-gray-700 mb-1">手续费率 (sat/vB)</label>
            <input
              v-model.number="exportManualFeeRate"
              type="number"
              step="0.1"
              min="0.1"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
              placeholder="输入手续费率"
            />
            <div class="text-xs text-gray-500 mt-1">建议范围：1-100 sat/vB</div>
          </div>

          <!-- 费率信息显示 -->
          <div v-if="exportingTransaction" class="bg-gray-50 border border-gray-200 rounded-md p-3 text-sm space-y-1">
            <div class="flex justify-between">
              <span>当前费率</span>
              <span>{{ getCurrentTransactionFeeRate(exportingTransaction).toFixed(1) }} sat/vB</span>
            </div>
            <div class="flex justify-between">
              <span>新费率</span>
              <span>{{ exportFeeRate.toFixed(1) }} sat/vB</span>
            </div>
            <div class="flex justify-between font-medium" :class="{
              'text-green-600': exportFeeRate < getCurrentTransactionFeeRate(exportingTransaction),
              'text-red-600': exportFeeRate > getCurrentTransactionFeeRate(exportingTransaction),
              'text-gray-600': Math.abs(exportFeeRate - getCurrentTransactionFeeRate(exportingTransaction)) < 0.1
            }">
              <span>费率变化</span>
              <span>{{ (exportFeeRate - getCurrentTransactionFeeRate(exportingTransaction)).toFixed(1) }} sat/vB</span>
            </div>
          </div>

          <!-- 费率验证提示 -->
          <div v-if="!exportFeeValidation.isValid" class="bg-red-50 border border-red-200 rounded-md p-3 text-sm">
            <div class="flex items-start">
              <div class="flex-shrink-0">
                <svg class="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                </svg>
              </div>
              <div class="ml-3">
                <p class="text-red-800 font-medium">费率过高</p>
                <p class="text-red-700 mt-1">{{ exportFeeValidation.message }}</p>
              </div>
            </div>
          </div>

          <!-- 找零用完警告 -->
          <div v-if="exportFeeValidation.warning" class="bg-yellow-50 border border-yellow-200 rounded-md p-3 text-sm">
            <div class="flex items-start">
              <div class="flex-shrink-0">
                <svg class="h-5 w-5 text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
                </svg>
              </div>
              <div class="ml-3">
                <p class="text-yellow-800 font-medium">找零用完</p>
                <p class="text-yellow-700 mt-1">{{ exportFeeValidation.warning }}</p>
              </div>
            </div>
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="showExportFeeModal = false"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            取消
          </button>
          <button
            @click="confirmExportFee"
            :disabled="!exportFeeValidation.isValid"
            class="px-4 py-2 bg-orange-600 text-white rounded-md hover:bg-orange-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            确认导出
          </button>
        </div>
      </div>
    </div>
    </teleport>

    <!-- 导出QR码模态框 -->
    <teleport to="body">
    <div v-if="showExportQRModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[1000]">
      <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-medium text-gray-900">交易QR码</h3>
            <button
              @click="closeExportQRModal"
              class="text-gray-400 hover:text-gray-600 transition-colors"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
        
        <div class="px-6 py-4">
          <div class="text-center">
            <div class="mb-4">
              <h4 class="text-md font-medium text-gray-900 mb-2">BTC交易数据</h4>
              <p class="text-sm text-gray-600">请使用扫码程序扫描此QR码进行签名</p>
            </div>
            
            <div class="flex justify-center mb-4">
              <div v-if="exportQRCodeDataURL" class="bg-white p-4 rounded-lg border-2 border-gray-200">
                <img :src="exportQRCodeDataURL" alt="交易QR码" class="max-w-full h-auto" />
              </div>
              <div v-else class="bg-gray-100 p-8 rounded-lg">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-orange-600 mx-auto"></div>
                <p class="text-sm text-gray-500 mt-2">生成中...</p>
              </div>
            </div>
            
            <!-- 交易信息摘要 -->
            <div v-if="exportQRTransaction" class="bg-gray-50 border border-gray-200 rounded-md p-3 text-sm space-y-1 text-left">
              <div class="flex justify-between">
                <span class="font-medium">交易ID:</span>
                <span class="font-mono">{{ exportQRTransaction.id }}</span>
              </div>
              <div class="flex justify-between">
                <span class="font-medium">链类型:</span>
                <span class="font-mono">BTC</span>
              </div>
              <div class="flex justify-between">
                <span class="font-medium">发送地址:</span>
                <span class="font-mono text-xs">{{ exportQRTransaction.from_address.substring(0, 10) }}...{{ exportQRTransaction.from_address.substring(exportQRTransaction.from_address.length - 8) }}</span>
              </div>
              <div class="flex justify-between">
                <span class="font-medium">金额:</span>
                <span class="font-mono">{{ formatBtcAmount(exportQRTransaction.amount) }} BTC</span>
              </div>
            </div>
          </div>
        </div>
        
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="closeExportQRModal"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            关闭
          </button>
          <button
            @click="downloadExportQRCode"
            :disabled="!exportQRCodeDataURL"
            class="px-4 py-2 bg-orange-600 text-white rounded-md hover:bg-orange-700 disabled:opacity-50"
          >
            下载QR码
          </button>
        </div>
      </div>
    </div>
    </teleport>

    <!-- 导入签名模态框 -->
    <teleport to="body">
    <div v-if="showImportSignatureModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[1000]">
      <div class="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-medium text-gray-900">导入签名数据</h3>
        </div>
        
        <div class="px-6 py-4">
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <!-- 左侧：交易基本信息 -->
            <div class="space-y-4">
              <h4 class="text-md font-medium text-gray-900">交易信息</h4>
              
              <!-- 显示选中交易的详细信息 -->
              <div v-if="selectedImportTransaction" class="bg-gray-50 rounded-lg p-4 space-y-3">
                <div class="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span class="text-gray-500">交易ID:</span>
                    <span class="font-mono">{{ selectedImportTransaction.id }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">状态:</span>
                    <span :class="getStatusClass(selectedImportTransaction.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                      {{ getStatusText(selectedImportTransaction.status) }}
                    </span>
                  </div>
                  <div>
                    <span class="text-gray-500">链类型:</span>
                    <span class="font-medium">{{ selectedImportTransaction.chain.toUpperCase() }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">币种:</span>
                    <span class="font-medium">{{ selectedImportTransaction.symbol }}</span>
                  </div>
                  <div class="col-span-2">
                    <span class="text-gray-500">发送地址:</span>
                    <code class="ml-2 text-xs font-mono bg-gray-200 px-2 py-1 rounded">{{ selectedImportTransaction.from_address }}</code>
                  </div>
                  <div class="col-span-2">
                    <span class="text-gray-500">接收地址:</span>
                    <code class="ml-2 text-xs font-mono bg-gray-200 px-2 py-1 rounded">{{ selectedImportTransaction.to_address }}</code>
                  </div>
                  <div>
                    <span class="text-gray-500">金额:</span>
                    <span class="font-medium">{{ formatBtcAmount(selectedImportTransaction.amount) }} {{ selectedImportTransaction.symbol }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">费率:</span>
                    <span class="font-medium">{{ getTransactionFeeRate(selectedImportTransaction) }} sat/vB</span>
                  </div>
                </div>
              </div>
            </div>
            
            <!-- 右侧：签名数据输入 -->
            <div class="space-y-4">
              <h4 class="text-md font-medium text-gray-900">签名数据</h4>
              
              <!-- 签名数据 -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">签名JSON数据</label>
                <textarea
                  v-model="importSignature"
                  rows="8"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500 font-mono text-sm"
                  placeholder='请粘贴从离线程序导出的签名数据，例如：
{"id":72,"signer":"0200000001ee8503075eb27aca508cda579dba06b6173dd6c22d6c95544ccdde158048a26e000000006a47304402204c9d4a2aba2ae1be503f48c80176a099b877c2f2fc16b1d418b9bf7173f0722402206168d4b535ec6ff6cd8ceba50fd02ea8b3653827699b68d3b64c535be10e4baf01210232d7aba6ebc83b114d91a228add5f0ad3ff65cae831af35a0ed273bac21349beffffffff02a0860100000000001976a914474aeb78ace0adea1bea973cffd1140cf9459f3788ac4d920000000000001976a914c338f892f9479e0e6bb1b78378f24b7f662cd80f88ac00000000"}'
                ></textarea>
              </div>
              
              <!-- 操作提示 -->
              <div class="bg-orange-50 border border-orange-200 rounded-md p-3">
                <div class="flex">
                  <div class="flex-shrink-0">
                    <svg class="h-5 w-5 text-orange-400" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                    </svg>
                  </div>
                  <div class="ml-3">
                    <p class="text-sm text-orange-800">
                      支持导入签名数据：完整的签名交易字符串或包含id和signer字段的JSON格式。导入后可以选择发送上链。
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div class="px-6 py-4 border-t border-gray-200 flex justify-between">
          <button
            @click="closeImportSignatureModal"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            取消
          </button>
          <div class="flex space-x-3">
            <button
              @click="importSignatureOnly"
              :disabled="!importSignature.trim() || !selectedImportTransaction || isImporting"
              class="px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 disabled:opacity-50 flex items-center"
            >
              <svg v-if="isImporting" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              {{ isImporting ? '导入中...' : '仅导入签名' }}
            </button>
            <button
              @click="importAndSendTransaction"
              :disabled="!importSignature.trim() || !selectedImportTransaction || isImporting || isSending"
              class="px-4 py-2 bg-orange-600 text-white rounded-md hover:bg-orange-700 disabled:opacity-50 flex items-center"
            >
              <svg v-if="isSending" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              {{ isSending ? '发送中...' : '导入并发送上链' }}
            </button>
          </div>
        </div>
      </div>
    </div>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import type { PersonalTransaction } from '@/types'
import { useChainWebSocket } from '@/composables/useWebSocket'
import type { FeeLevels } from '@/types'
import { getBtcCachedGasRates } from '@/api/no-auth'
import { getPersonalAddresses, getAddressUTXOs } from '@/api/personal-addresses'
import type { PersonalAddressItem, BTCUTXO } from '@/types/personal-address'
import { 
  getUserTransactions, 
  createUserTransaction, 
  updateUserTransaction,
  getUserTransactionById,
  getUserTransactionStats,
  exportTransaction,
  importSignature as importSignatureAPI,
  sendTransaction as sendTransactionAPI
} from '@/api/user-transactions'
import type { 
  UserTransaction, 
  CreateUserTransactionRequest, 
  UpdateUserTransactionRequest,
  UserTransactionStatsResponse,
  BTCTxIn,
  BTCTxOut
} from '@/types/user-transaction'

// 响应式数据
const showCreateModal = ref(false)
const isEditMode = ref(false)
const editingTransaction = ref<UserTransaction | null>(null)
const selectedStatus = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const totalItems = ref(0)
const totalPages = ref(0)

// 导出费率选择相关
const showExportFeeModal = ref(false)
const exportingTransaction = ref<UserTransaction | null>(null)
const exportFeeMode = ref<'auto' | 'manual'>('auto')
const exportFeeTier = ref<'slow' | 'normal' | 'fast'>('normal')
const exportManualFeeRate = ref(2.5)

// QR码显示相关
const showExportQRModal = ref(false)
const exportQRCodeDataURL = ref<string>('')
const exportQRTransaction = ref<UserTransaction | null>(null)

// 导入签名相关状态
const showImportSignatureModal = ref(false)
const selectedImportTransaction = ref<UserTransaction | null>(null)
const importSignature = ref('')
const isImporting = ref(false)
const isSending = ref(false)

// 新建交易表单（UTXO模型）
const createForm = ref<{
  fromAddress: string
  outputs: Array<{ toAddress: string; amountBtc: number | null }>
  feeTier: 'slow' | 'normal' | 'fast'
}>(
  { fromAddress: '', outputs: [{ toAddress: '', amountBtc: null }], feeTier: 'normal' }
)

// 自动找零
const autoChange = ref<boolean>(true)

// 地址与UTXO数据
const btcAddresses = ref<PersonalAddressItem[]>([])
const utxos = ref<BTCUTXO[]>([])
const selectedUtxoIds = ref<Set<number>>(new Set())

// WebSocket相关
const { subscribeChainEvent } = useChainWebSocket('btc')
// 收集本组件的取消订阅函数，避免重复回调
const wsUnsubscribes: Array<() => void> = []

// 费率数据
const feeLevels = ref<FeeLevels | null>(null)
const networkCongestion = ref<string>('normal')

// 费率历史数据存储（用于折线图）
const feeHistory = ref<Array<{
  timestamp: number
  fee: number
}>>([])

// 图表相关
const feeChartCanvas = ref<HTMLCanvasElement | null>(null)
const feeTooltip = ref<HTMLDivElement | null>(null)

// 交易统计
const totalTransactions = ref(0)
const unsignedCount = ref(0)
const inProgressCount = ref(0)
const confirmedCount = ref(0)

// 交易列表
const transactionsList = ref<UserTransaction[]>([])

// fee与大小估算
const satPerVb = computed<number>(() => {
  const tier = createForm.value.feeTier
  const f = feeLevels.value
  if (!f) return 0
  if (tier === 'slow') return Number(f.slow?.max_priority_fee || 0)
  if (tier === 'fast') return Number(f.fast?.max_priority_fee || 0)
  return Number(f.normal?.max_priority_fee || 0)
})

const numInputs = computed(() => selectedUtxoIds.value.size)
const validOutputsCount = computed<number>(() => {
  return createForm.value.outputs.filter(o => o.toAddress && Number(o.amountBtc || 0) > 0).length
})

const estimatedVBytesBase = computed<number>(() => {
  // 基础估算（不含找零输出）：10 + 148*inputs + 34*outputs
  return 10 + 148 * numInputs.value + 34 * validOutputsCount.value
})

// 资金计算
const toSats = (btc: number) => Math.round(btc * 1e8)
const toBtc = (sats: number) => sats / 1e8

// 格式化BTC金额显示
const formatBtcAmount = (amount: string | number) => {
  if (!amount) return '0.00000000'
  
  // 如果amount是字符串，先转换为数字
  const numAmount = typeof amount === 'string' ? parseFloat(amount) : amount
  
  // BTC交易在后端存储的是聪单位，需要转换为BTC显示
  // 对于BTC链，总是将存储的金额视为聪单位进行转换
  return toBtc(numAmount).toFixed(8)
}

// 获取交易费率显示
const getTransactionFeeRate = (tx: UserTransaction) => {
  // 尝试从BTC交易数据中获取费率
  if (tx.btc_tx_in_json && tx.btc_tx_out_json) {
    try {
      const txIn = JSON.parse(tx.btc_tx_in_json)
      const txOut = JSON.parse(tx.btc_tx_out_json)
      
      // 计算交易大小（估算）
      const estimatedVBytes = 10 + 148 * txIn.length + 34 * txOut.length
      
      // 计算费率 = 手续费 / 交易大小
      const feeSats = Number(tx.fee || 0)
      if (feeSats > 0 && estimatedVBytes > 0) {
        const feeRate = feeSats / estimatedVBytes
        return feeRate.toFixed(1)
      }
    } catch (e) {
      console.error('解析BTC交易数据失败:', e)
    }
  }
  
  // 如果没有BTC数据，显示默认值
  return '-'
}

const selectedInputsTotalSats = computed(() => {
  if (selectedUtxoIds.value.size === 0) return 0
  return utxos.value
    .filter(u => selectedUtxoIds.value.has(u.id))
    .reduce((sum, u) => sum + Number(u.value_satoshi || 0), 0)
})

const outputsTotalSats = computed(() => {
  return createForm.value.outputs.reduce((sum, o) => {
    const v = Number(o.amountBtc || 0)
    return sum + (v > 0 ? toSats(v) : 0)
  }, 0)
})

const estimatedFeeSatsBase = computed<number>(() => Math.round(satPerVb.value * estimatedVBytesBase.value))
const changeSatsBase = computed<number>(() => selectedInputsTotalSats.value - outputsTotalSats.value - estimatedFeeSatsBase.value)
const willHaveChange = computed<boolean>(() => changeSatsBase.value > 0)

const estimatedVBytes = computed<number>(() => estimatedVBytesBase.value + (autoChange.value && willHaveChange.value ? 34 : 0))
const estimatedFeeSats = computed<number>(() => Math.round(satPerVb.value * estimatedVBytes.value))
const changeSats = computed<number>(() => selectedInputsTotalSats.value - outputsTotalSats.value - estimatedFeeSats.value)
const isBalanceEnough = computed<boolean>(() => {
  if (autoChange.value) return changeSats.value >= 0
  // 不自动找零时，必须刚好用尽输入（允许极小误差）
  return Math.abs(changeSats.value) < 2 // 允许2 sats舍入误差
})

// 导出费率计算
const exportFeeRate = computed<number>(() => {
  if (exportFeeMode.value === 'manual') {
    return exportManualFeeRate.value
  }
  
  const f = feeLevels.value
  if (!f) return 0
  if (exportFeeTier.value === 'slow') return Number(f.slow?.max_priority_fee || 0)
  if (exportFeeTier.value === 'fast') return Number(f.fast?.max_priority_fee || 0)
  return Number(f.normal?.max_priority_fee || 0)
})

// 费率验证
const exportFeeValidation = computed(() => {
  if (!exportingTransaction.value) return { isValid: true, message: '', warning: '' }
  
  const tx = exportingTransaction.value
  const txIn = JSON.parse(tx.btc_tx_in_json || '[]')
  const txOut = JSON.parse(tx.btc_tx_out_json || '[]')
  const estimatedVBytes = 10 + 148 * txIn.length + 34 * txOut.length
  const newFeeSats = Math.round(exportFeeRate.value * estimatedVBytes)
  
  const currentFeeSats = Number(tx.fee || 0)
  const feeDifference = newFeeSats - currentFeeSats
  
  // 计算当前找零金额
  const currentChangeSats = txOut
    .filter((out: any) => out.address === tx.from_address)
    .reduce((sum: number, out: any) => sum + Number(out.value_satoshi || 0), 0)
  
  // 验证费率是否过高（允许找零被完全用完）
  if (feeDifference > currentChangeSats) {
    const maxFeeSats = currentFeeSats + currentChangeSats
    const maxFeeRate = maxFeeSats / estimatedVBytes
    return {
      isValid: false,
      message: `当前费率已经超过了解锁总金额！最大手续费不能超过 ${maxFeeSats} sats (${maxFeeRate.toFixed(1)} sat/vB)`,
      warning: ''
    }
  }
  
  // 检查是否找零被完全用完
  const newChangeSats = Math.max(0, currentChangeSats - feeDifference)
  if (newChangeSats === 0 && currentChangeSats > 0) {
    return {
      isValid: true,
      message: '',
      warning: '⚠️ 当前费率将完全用完找零金额，找零输出将被删除'
    }
  }
  
  return { isValid: true, message: '', warning: '' }
})

// 表单校验
const isCreateFormValid = computed(() => {
  const a = createForm.value
  const hasFrom = !!a.fromAddress
  const hasAtLeastOneUtxo = selectedUtxoIds.value.size > 0
  const hasValidOutputs = a.outputs.some(o => o.toAddress && Number(o.amountBtc || 0) > 0)
  return hasFrom && hasAtLeastOneUtxo && hasValidOutputs && isBalanceEnough.value
})

// 计算属性
const filteredTransactions = computed(() => {
  if (!selectedStatus.value) {
    return transactionsList.value
  }
  return transactionsList.value.filter(tx => tx.status === selectedStatus.value)
})

// 获取状态样式
const getStatusClass = (status: string) => {
  switch (status) {
    case 'draft': return 'bg-gray-100 text-gray-800'
    case 'unsigned': return 'bg-gray-100 text-gray-800'
    case 'in_progress': return 'bg-yellow-100 text-yellow-800'
    case 'packed': return 'bg-orange-100 text-orange-800'
    case 'confirmed': return 'bg-green-100 text-green-800'
    case 'failed': return 'bg-red-100 text-red-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

// 获取状态文本
const getStatusText = (status: string) => {
  switch (status) {
    case 'draft': return '草稿'
    case 'unsigned': return '未签名'
    case 'in_progress': return '在途'
    case 'packed': return '已打包'
    case 'confirmed': return '已确认'
    case 'failed': return '失败'
    default: return '未知'
  }
}

// 格式化时间
const formatTime = (timestamp: Date | string | undefined) => {
  if (!timestamp) return '未知时间'
  const date = typeof timestamp === 'string' ? new Date(timestamp) : timestamp
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 导出交易
const handleExportTransaction = async (tx: UserTransaction) => {
  exportingTransaction.value = tx
  showExportFeeModal.value = true
}

// 发送交易
const sendTransactionToChain = async (txId: number) => {
  try {
    const response = await sendTransactionAPI(txId)
    return response
  } catch (error) {
    console.error('发送交易失败:', error)
    throw error
  }
}

// 打开导入签名模态框
const openImportSignatureModal = (tx: UserTransaction) => {
  selectedImportTransaction.value = tx
  importSignature.value = ''
  showImportSignatureModal.value = true
}

// 关闭导入签名模态框
const closeImportSignatureModal = () => {
  showImportSignatureModal.value = false
  selectedImportTransaction.value = null
  importSignature.value = ''
  isImporting.value = false
  isSending.value = false
}

// 仅导入签名
const importSignatureOnly = async () => {
  if (!selectedImportTransaction.value || !importSignature.value.trim()) return
  
  try {
    isImporting.value = true
    
    // 解析签名数据
    const signatureData = parseSignatureData(importSignature.value)
    if (!signatureData) {
      alert('签名数据格式错误，请检查数据格式')
      return
    }
    
    // 验证ID是否匹配
    if (signatureData.id !== undefined && signatureData.id !== selectedImportTransaction.value.id) {
      alert(`签名数据ID(${signatureData.id})与所选交易ID(${selectedImportTransaction.value.id})不匹配`)
      return
    }
    
    // 调用导入签名API
    const response = await importSignatureAPI(selectedImportTransaction.value.id, { 
      id: selectedImportTransaction.value.id, 
      signed_tx: signatureData.signedTx,
      v: signatureData.v,
      r: signatureData.r,
      s: signatureData.s
    })
    
    if ((response as any)?.success) {
      alert('导入签名成功！')
      loadTransactions()
      loadTransactionStats()
      closeImportSignatureModal()
    } else {
      alert('导入签名失败: ' + ((response as any)?.message || '未知错误'))
    }
  } catch (error) {
    console.error('导入签名失败:', error)
    alert('导入签名失败，请重试')
  } finally {
    isImporting.value = false
  }
}

// 导入并发送交易
const importAndSendTransaction = async () => {
  if (!selectedImportTransaction.value || !importSignature.value.trim()) return
  
  try {
    isSending.value = true
    
    // 先导入签名
    const signatureData = parseSignatureData(importSignature.value)
    if (!signatureData) {
      alert('签名数据格式错误，请检查数据格式')
      return
    }
    
    // 验证ID是否匹配
    if (signatureData.id !== undefined && signatureData.id !== selectedImportTransaction.value.id) {
      alert(`签名数据ID(${signatureData.id})与所选交易ID(${selectedImportTransaction.value.id})不匹配`)
      return
    }
    
    // 导入签名
    const importResponse = await importSignatureAPI(selectedImportTransaction.value.id, { 
      id: selectedImportTransaction.value.id, 
      signed_tx: signatureData.signedTx,
      v: signatureData.v,
      r: signatureData.r,
      s: signatureData.s
    })
    
    if (!(importResponse as any)?.success) {
      alert('导入签名失败: ' + ((importResponse as any)?.message || '未知错误'))
      return
    }
    
    // 发送交易上链
    const sendResponse = await sendTransactionToChain(selectedImportTransaction.value.id)
    
    if ((sendResponse as any)?.success) {
      alert('交易发送成功！')
      loadTransactions()
      loadTransactionStats()
      closeImportSignatureModal()
    } else {
      alert('发送交易失败: ' + ((sendResponse as any)?.message || '未知错误'))
    }
  } catch (error) {
    console.error('导入并发送交易失败:', error)
    alert('导入并发送交易失败，请重试')
  } finally {
    isSending.value = false
  }
}

// 解析签名数据
const parseSignatureData = (signatureText: string) => {
  try {
    // 尝试解析JSON格式的签名数据
    const data = JSON.parse(signatureText)
    
    // 检查是否包含BTC签名字段
    if (data.id && data.signer) {
      return {
        id: typeof data.id === 'string' ? parseInt(data.id, 10) : data.id,
        signedTx: data.signer,
        v: null,
        r: null,
        s: null
      }
    }
    
    // 如果只是签名交易字符串
    if (typeof data === 'string' || data.signedTx) {
      return {
        signedTx: data.signedTx || data,
        v: null,
        r: null,
        s: null
      }
    }
    
    return null
  } catch (error) {
    // 如果不是JSON格式，假设是直接的签名交易字符串
    if (signatureText.startsWith('0') && signatureText.length > 100) {
      return {
        signedTx: signatureText,
        v: null,
        r: null,
        s: null
      }
    }
    
    console.error('解析签名数据失败:', error)
    return null
  }
}

// 确认导出费率并执行导出
const confirmExportFee = async () => {
  if (!exportingTransaction.value) return
  
  try {
    // 获取当前交易的费率信息
    const currentTx = exportingTransaction.value
    const currentFeeRate = getCurrentTransactionFeeRate(currentTx)
    const newFeeRate = exportFeeRate.value
    
    let updatedTx = currentTx
    
    // 如果费率有变化，需要更新交易
    if (Math.abs(currentFeeRate - newFeeRate) > 0.1) {
      try {
        await updateTransactionFeeRate(currentTx, newFeeRate)
        
        // 更新后重新获取最新的交易数据
        const updatedTxRes = await getUserTransactionById(currentTx.id)
        if ((updatedTxRes as any)?.success && (updatedTxRes as any)?.data) {
          updatedTx = (updatedTxRes as any).data
          console.log('已获取更新后的交易数据:', updatedTx)
        } else {
          console.warn('获取更新后的交易数据失败，使用原始数据')
        }
      } catch (error) {
        // 显示用户友好的错误信息
        const errorMessage = error instanceof Error ? error.message : '费率更新失败'
        alert(`❌ ${errorMessage}`)
        return
      }
    }
    
    // 执行导出
    const res = await exportTransaction(updatedTx.id)
    if ((res as any)?.success && (res as any)?.data) {
      console.log('导出交易成功:', (res as any).data)
      showExportFeeModal.value = false
      
      // 刷新交易列表和统计
      await loadTransactions()
      await loadTransactionStats()
      
      // 显示QR码，使用更新后的交易数据
      await showExportQRCode(updatedTx, (res as any).data)
    } else {
      console.error('导出交易失败:', (res as any)?.message || '未知错误')
      alert('❌ 导出交易失败: ' + ((res as any)?.message || '未知错误'))
    }
  } catch (error) {
    console.error('导出交易失败:', error)
    alert('❌ 导出交易失败: ' + (error instanceof Error ? error.message : '未知错误'))
  }
}

// 获取当前交易的费率
const getCurrentTransactionFeeRate = (tx: UserTransaction): number => {
  if (tx.btc_tx_in_json && tx.btc_tx_out_json) {
    try {
      const txIn = JSON.parse(tx.btc_tx_in_json)
      const txOut = JSON.parse(tx.btc_tx_out_json)
      const estimatedVBytes = 10 + 148 * txIn.length + 34 * txOut.length
      const feeSats = Number(tx.fee || 0)
      if (feeSats > 0 && estimatedVBytes > 0) {
        return feeSats / estimatedVBytes
      }
    } catch (e) {
      console.error('解析BTC交易数据失败:', e)
    }
  }
  return 0
}

// 更新交易费率
const updateTransactionFeeRate = async (tx: UserTransaction, newFeeRate: number) => {
  // 计算新的手续费
  const txIn = JSON.parse(tx.btc_tx_in_json || '[]')
  const txOut = JSON.parse(tx.btc_tx_out_json || '[]')
  const estimatedVBytes = 10 + 148 * txIn.length + 34 * txOut.length
  const newFeeSats = Math.round(newFeeRate * estimatedVBytes)
  
  // 计算找零调整
  const currentFeeSats = Number(tx.fee || 0)
  const feeDifference = newFeeSats - currentFeeSats
  
  // 计算当前找零金额
  const currentChangeSats = txOut
    .filter((out: any) => out.address === tx.from_address)
    .reduce((sum: number, out: any) => sum + Number(out.value_satoshi || 0), 0)
  
  // 验证费率是否过高（允许找零被完全用完）
  if (feeDifference > currentChangeSats) {
    const maxFeeSats = currentFeeSats + currentChangeSats
    const maxFeeRate = maxFeeSats / estimatedVBytes
    throw new Error(`当前费率已经超过了解锁总金额！最大手续费不能超过 ${maxFeeSats} sats (${maxFeeRate.toFixed(1)} sat/vB)`)
  }
  
  console.log('费率更新详情:', {
    currentFeeRate: currentFeeSats / estimatedVBytes,
    newFeeRate,
    currentFeeSats,
    newFeeSats,
    feeDifference,
    estimatedVBytes,
    currentChangeSats
  })
  
  // 更新交易数据
  const updateRequest: UpdateUserTransactionRequest = {
    fee: newFeeSats.toString(),
    // 如果费率降低，增加找零；如果费率提高，减少找零
    btc_tx_out: txOut
      .map((out: any) => {
        if (out.address === tx.from_address) {
          // 这是找零输出，调整金额
          const currentChangeSats = Number(out.value_satoshi || 0)
          const newChangeSats = Math.max(0, currentChangeSats - feeDifference)
          console.log('找零调整:', {
            address: out.address,
            currentChangeSats,
            newChangeSats,
            feeDifference
          })
          return {
            ...out,
            value_satoshi: newChangeSats
          }
        }
        return out
      })
      .filter((out: any) => {
        // 过滤掉金额为0的找零输出
        if (out.address === tx.from_address && Number(out.value_satoshi || 0) === 0) {
          console.log('删除金额为0的找零输出:', out.address)
          return false
        }
        return true
      })
  }
  
  const res = await updateUserTransaction(tx.id, updateRequest)
  if (!(res as any)?.success) {
    throw new Error((res as any)?.message || '更新交易费率失败')
  }
  
  console.log('交易费率更新成功')
}

// 显示导出QR码
const showExportQRCode = async (tx: UserTransaction, exportData: any) => {
  try {
    exportQRTransaction.value = tx
    showExportQRModal.value = true
    exportQRCodeDataURL.value = '' // 重置QR码
    
    // 构建扫码程序需要的数据结构（异步，包含 PrevOuts）
    const qrData = await buildExportQRData(tx, exportData)
    
    // 生成QR码
    await generateExportQRCode(qrData)
  } catch (error) {
    console.error('显示QR码失败:', error)
    alert('生成QR码失败，请重试')
  }
}

// 创建导出QR码数据（包含 PrevOuts）
const buildExportQRData = async (tx: UserTransaction, exportData: any) => {
  // 解析BTC交易数据
  let txIn: any[] = []
  let txOut: any[] = []
  
  if (tx.btc_tx_in_json) {
    try { txIn = JSON.parse(tx.btc_tx_in_json) } catch (e) { console.error('解析TxIn失败:', e) }
  }
  if (tx.btc_tx_out_json) {
    try { txOut = JSON.parse(tx.btc_tx_out_json) } catch (e) { console.error('解析TxOut失败:', e) }
  }

  // 准备 PrevOuts：从地址UTXO中按 txid:vout 匹配
  let prevOuts: Array<{ txid: string; vout: number; value_satoshi: number; script_pubkey_hex: string }> = []
  try {
    const utxoRes = await getAddressUTXOs(tx.from_address)
    if ((utxoRes as any)?.success && (utxoRes as any)?.data) {
      const addrUtxos = (utxoRes as any).data as any[]
      prevOuts = txIn.map((input: any) => {
        const match = addrUtxos.find(u => u.tx_id === input.txid && Number(u.vout_index) === Number(input.vout))
        return {
          txid: input.txid,
          vout: Number(input.vout),
          value_satoshi: match ? Number(match.value_satoshi || 0) : 0,
          // 使用后端DTO中的脚本字段：script_pub_key
          script_pubkey_hex: match && match.script_pub_key ? match.script_pub_key : ''
        }
      })
    } else {
      // 无法获取UTXO时，仍生成占位PrevOuts，value为0，脚本为空
      prevOuts = txIn.map((input: any) => ({ txid: input.txid, vout: Number(input.vout), value_satoshi: 0, script_pubkey_hex: '' }))
    }
  } catch (e) {
    console.warn('获取地址UTXO失败，PrevOuts将为空或占位:', e)
    prevOuts = txIn.map((input: any) => ({ txid: input.txid, vout: Number(input.vout), value_satoshi: 0, script_pubkey_hex: '' }))
  }

  // 构建扫码程序需要的数据结构
  const qrData = {
    id: tx.id,
    type: 'btc',
    address: tx.from_address,
    MsgTx: {
      Version: tx.btc_version || 2,
      TxIn: txIn.map(input => ({
        txid: input.txid,
        vout: input.vout,
        sequence: input.sequence || 0xffffffff
      })),
      TxOut: txOut.map(output => ({
        value_satoshi: output.value_satoshi,
        address: output.address
      })),
      LockTime: tx.btc_lock_time || 0,
      PrevOuts: prevOuts
    }
  }
  return qrData
}

// 生成导出QR码
const generateExportQRCode = async (qrData: any) => {
  try {
    // 动态导入QRCode库
    const QRCode = await import('qrcode')
    
    // 将数据转换为JSON字符串
    const qrDataString = JSON.stringify(qrData, null, 0)
    
    console.log('准备生成导出QR码:', {
      dataLength: qrDataString.length,
      dataPreview: qrDataString.substring(0, 1800) + '...'
    })
    
    // 生成QR码配置
    const qrOptions = {
      type: 'image/png' as const,
      quality: 1.0,
      margin: 4,
      color: { dark: '#000000', light: '#FFFFFF' },
      width: 1024,
      errorCorrectionLevel: 'H' as const,
      scale: 8
    }
    
    // 生成QR码数据URL
    const qrDataURL = await QRCode.toDataURL(qrDataString, qrOptions)
    exportQRCodeDataURL.value = qrDataURL
    
    console.log('导出QR码生成完成:', { dataLength: qrDataString.length, qrSize: qrOptions.width })
    
  } catch (error) {
    console.error('生成导出QR码失败:', error)
    exportQRCodeDataURL.value = ''
    alert('QR码生成失败，请重试')
  }
}

// 关闭QR码模态框
const closeExportQRModal = async () => {
  showExportQRModal.value = false
  exportQRCodeDataURL.value = ''
  exportQRTransaction.value = null
  
  // 刷新交易列表和统计
  await loadTransactions()
  await loadTransactionStats()
}

// 下载导出QR码
const downloadExportQRCode = () => {
  if (!exportQRCodeDataURL.value || !exportQRTransaction.value) return
  
  const tx = exportQRTransaction.value
  const link = document.createElement('a')
  link.href = exportQRCodeDataURL.value
  link.download = `export_transaction_${tx.id}_btc_qr.png`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  
  // 显示成功提示
  const toast = document.createElement('div')
  toast.className = 'fixed top-4 right-4 bg-green-500 text-white px-4 py-2 rounded-md shadow-lg z-50 transition-opacity duration-300'
  toast.textContent = 'QR码已下载！'
  document.body.appendChild(toast)
  
  setTimeout(() => {
    toast.style.opacity = '0'
    setTimeout(() => {
      document.body.removeChild(toast)
    }, 300)
  }, 3000)
}

// 编辑交易
const editTransaction = async (tx: UserTransaction) => {
  console.log('编辑交易:', tx)
  
  // 如果是BTC交易，打开新建交易模态框并预填充数据
  if (tx.chain === 'btc') {
    await openEditModal(tx)
  } else {
    // 其他链类型的编辑逻辑
    console.log('暂不支持编辑此链类型的交易')
  }
}

// 查看交易
const viewTransaction = (tx: UserTransaction) => {
  // TODO: 实现查看交易详情功能
  console.log('查看交易:', tx)
}

// 新建交易 - 打开/关闭/提交
const openCreateModal = async () => {
  // 重置状态
  isEditMode.value = false
  editingTransaction.value = null
  
  // 重置表单
  createForm.value = { fromAddress: '', outputs: [{ toAddress: '', amountBtc: null }], feeTier: 'normal' }
  utxos.value = []
  selectedUtxoIds.value = new Set()

  try {
    const res = await getPersonalAddresses('btc')
    if ((res as any)?.success && (res as any)?.data) {
      btcAddresses.value = (res as any).data
    } else {
      btcAddresses.value = []
    }
  } catch {
    btcAddresses.value = []
  }

  showCreateModal.value = true
}

// 编辑交易 - 打开模态框并预填充数据
const openEditModal = async (tx: UserTransaction) => {
  try {
    // 设置编辑状态
    isEditMode.value = true
    editingTransaction.value = tx
    
    // 加载BTC地址列表
    const res = await getPersonalAddresses('btc')
    if ((res as any)?.success && (res as any)?.data) {
      btcAddresses.value = (res as any).data
    } else {
      btcAddresses.value = []
    }

    // 解析BTC交易数据
    let btcTxIn: BTCTxIn[] = []
    let btcTxOut: BTCTxOut[] = []
    
    if (tx.btc_tx_in_json) {
      try {
        btcTxIn = JSON.parse(tx.btc_tx_in_json)
      } catch (e) {
        console.error('解析BTC TxIn失败:', e)
      }
    }
    
    if (tx.btc_tx_out_json) {
      try {
        btcTxOut = JSON.parse(tx.btc_tx_out_json)
      } catch (e) {
        console.error('解析BTC TxOut失败:', e)
      }
    }

    // 预填充表单数据
    // 过滤掉找零地址（输出地址与发送地址相同的记录）
    const filteredOutputs = btcTxOut.filter(out => out.address !== tx.from_address)
    
    createForm.value = {
      fromAddress: tx.from_address,
      outputs: filteredOutputs.map(out => ({
        toAddress: out.address || '',
        amountBtc: toBtc(out.value_satoshi)
      })),
      feeTier: 'normal' // 默认费率档位
    }
    
    // 如果过滤后没有输出，添加一个空输出
    if (createForm.value.outputs.length === 0) {
      createForm.value.outputs = [{ toAddress: '', amountBtc: null }]
    }

    // 加载发送地址的UTXO
    if (tx.from_address) {
      try {
        const utxoRes = await getAddressUTXOs(tx.from_address)
        if ((utxoRes as any)?.success && (utxoRes as any)?.data) {
          utxos.value = (utxoRes as any).data as BTCUTXO[]
          
          // 根据TxIn预选择UTXO
          const selectedIds = new Set<number>()
          btcTxIn.forEach(txIn => {
            const utxo = utxos.value.find(u => u.tx_id === txIn.txid && u.vout_index === txIn.vout)
            if (utxo) {
              selectedIds.add(utxo.id)
            }
          })
          selectedUtxoIds.value = selectedIds
        }
      } catch (e) {
        console.error('加载UTXO失败:', e)
        utxos.value = []
      }
    }

    showCreateModal.value = true
  } catch (error) {
    console.error('打开编辑模态框失败:', error)
    alert('加载交易数据失败，请重试')
  }
}

const closeCreateModal = () => {
  showCreateModal.value = false
  isEditMode.value = false
  editingTransaction.value = null
}

const submitCreateTransaction = async () => {
  if (!isCreateFormValid.value) return
  
  try {
    // 构建BTC交易数据
    const selectedUtxos = utxos.value.filter(u => selectedUtxoIds.value.has(u.id))
    const btcTxIn: BTCTxIn[] = selectedUtxos.map(u => ({
      txid: u.tx_id,
      vout: u.vout_index,
      sequence: 0xffffffff // 默认序列号
    }))
    
    const btcTxOut: BTCTxOut[] = createForm.value.outputs
      .filter(o => o.toAddress && Number(o.amountBtc || 0) > 0)
      .map(o => ({
        value_satoshi: toSats(Number(o.amountBtc || 0)),
        address: o.toAddress
      }))
    
    // 如果有找零，添加找零输出
    if (autoChange.value && changeSats.value > 0) {
      btcTxOut.push({
        value_satoshi: changeSats.value,
        address: createForm.value.fromAddress // 找零回到发送地址
      })
    }
    
    if (isEditMode.value && editingTransaction.value) {
      // 编辑模式：更新现有交易
      const updateRequest: UpdateUserTransactionRequest = {
        from_address: createForm.value.fromAddress,
        to_address: createForm.value.outputs[0]?.toAddress || '',
        amount: outputsTotalSats.value.toString(),
        fee: estimatedFeeSats.value.toString(),
        remark: `BTC交易 - ${selectedUtxos.length}个输入，${btcTxOut.length}个输出`,
        
        // BTC特有字段
        btc_version: 2,
        btc_lock_time: 0,
        btc_tx_in: btcTxIn,
        btc_tx_out: btcTxOut
      }
      
      const res = await updateUserTransaction(editingTransaction.value.id, updateRequest)
      if ((res as any)?.success && (res as any)?.data) {
        // 更新成功，刷新列表
        await loadTransactions()
        await loadTransactionStats()
        closeCreateModal()
        console.log('更新交易成功:', (res as any).data)
      } else {
        console.error('更新交易失败:', (res as any)?.message || '未知错误')
      }
    } else {
      // 创建模式：新建交易
      const createRequest: CreateUserTransactionRequest = {
        chain: 'btc',
        symbol: 'BTC',
        from_address: createForm.value.fromAddress,
        to_address: createForm.value.outputs[0]?.toAddress || '',
        amount: outputsTotalSats.value.toString(), // outputsTotalSats已经是聪单位，不需要再次转换
        fee: estimatedFeeSats.value.toString(),
        remark: `BTC交易 - ${selectedUtxos.length}个输入，${btcTxOut.length}个输出`,
        
        // BTC特有字段
        btc_version: 2,
        btc_lock_time: 0,
        btc_tx_in: btcTxIn,
        btc_tx_out: btcTxOut
      }
      
      const res = await createUserTransaction(createRequest)
      if ((res as any)?.success && (res as any)?.data) {
        // 创建成功，刷新列表
        await loadTransactions()
        await loadTransactionStats()
        closeCreateModal()
        console.log('新建交易成功:', (res as any).data)
      } else {
        console.error('新建交易失败:', (res as any)?.message || '未知错误')
      }
    }
  } catch (error) {
    console.error(isEditMode.value ? '更新交易失败:' : '新建交易失败:', error)
  }
}

// 监听fromAddress变化，加载其UTXO
watch(() => createForm.value.fromAddress, async (addr) => {
  utxos.value = []
  selectedUtxoIds.value = new Set()
  if (!addr) return
  try {
    const res = await getAddressUTXOs(addr)
    if ((res as any)?.success && (res as any)?.data) {
      utxos.value = (res as any).data as BTCUTXO[]
    } else {
      utxos.value = []
    }
  } catch {
    utxos.value = []
  }
})

// 分页方法
const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    loadTransactions()
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    loadTransactions()
  }
}

// 加载交易数据
const loadTransactions = async () => {
  try {
    const res = await getUserTransactions({
      page: currentPage.value,
      page_size: pageSize.value,
      status: selectedStatus.value || undefined,
      chain: 'btc'
    })
    
    if ((res as any)?.success && (res as any)?.data) {
      const data = (res as any).data
      transactionsList.value = data.transactions || []
      totalItems.value = data.total || 0
      totalPages.value = Math.ceil(totalItems.value / pageSize.value)
    } else {
      console.error('加载交易列表失败:', (res as any)?.message || '未知错误')
      transactionsList.value = []
      totalItems.value = 0
      totalPages.value = 0
    }
  } catch (error) {
    console.error('加载交易列表失败:', error)
    transactionsList.value = []
    totalItems.value = 0
    totalPages.value = 0
  }
}

// 加载交易统计
const loadTransactionStats = async () => {
  try {
    const res = await getUserTransactionStats()
    if ((res as any)?.success && (res as any)?.data) {
      const stats = (res as any).data as UserTransactionStatsResponse
      totalTransactions.value = stats.total_transactions
      unsignedCount.value = stats.unsigned_count
      inProgressCount.value = stats.in_progress_count
      confirmedCount.value = stats.confirmed_count
    } else {
      console.error('加载交易统计失败:', (res as any)?.message || '未知错误')
    }
  } catch (error) {
    console.error('加载交易统计失败:', error)
  }
}

// 监听状态筛选变化
watch(selectedStatus, () => {
  currentPage.value = 1
  loadTransactions()
})

// 添加费率历史数据
const addFeeHistory = (feeData: FeeLevels) => {
  const now = Date.now()
  const historyItem = {
    timestamp: now,
    fee: parseFloat(feeData.normal.max_priority_fee || '0')
  }
  
  // 添加到历史数据
  feeHistory.value.push(historyItem)
  
  // 只保留最近20条记录
  if (feeHistory.value.length > 20) {
    feeHistory.value = feeHistory.value.slice(-20)
  }
  
  // 更新图表
  updateChart()
}

// 费率鼠标移动事件处理
const handleFeeMouseMove = (event: MouseEvent) => {
  if (!feeChartCanvas.value || !feeTooltip.value || feeHistory.value.length === 0) return
  
  const canvas = feeChartCanvas.value
  const rect = canvas.getBoundingClientRect()
  const x = event.clientX - rect.left
  const y = event.clientY - rect.top
  
  // 计算数据点索引
  const padding = { top: 10, right: 10, bottom: 20, left: 40 }
  const chartWidth = rect.width - padding.left - padding.right
  const dataIndex = Math.round(((x - padding.left) / chartWidth) * (feeHistory.value.length - 1))
  
  // 确保索引在有效范围内
  if (dataIndex >= 0 && dataIndex < feeHistory.value.length) {
    const data = feeHistory.value[dataIndex]
    
    // 更新工具提示内容
    const feeElement = document.getElementById('tooltip-fee-value')
    if (feeElement) feeElement.textContent = data.fee.toFixed(1)
    
    // 计算相对于父容器的位置
    const parentRect = feeTooltip.value.parentElement?.getBoundingClientRect()
    
    if (parentRect) {
      const relativeX = event.clientX - parentRect.left
      const relativeY = event.clientY - parentRect.top
      
      feeTooltip.value.style.left = relativeX + 'px'
      feeTooltip.value.style.top = (relativeY - 10) + 'px'
      feeTooltip.value.style.opacity = '1'
    }
  }
}

// 费率鼠标离开事件处理
const handleFeeMouseLeave = () => {
  if (feeTooltip.value) {
    feeTooltip.value.style.opacity = '0'
  }
}

// 绘制费率图表
const drawFeeChart = (canvas: HTMLCanvasElement, data: number[], color: string, title: string) => {
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  
  // 设置canvas尺寸
  const rect = canvas.getBoundingClientRect()
  canvas.width = rect.width * window.devicePixelRatio
  canvas.height = rect.height * window.devicePixelRatio
  ctx.scale(window.devicePixelRatio, window.devicePixelRatio)
  
  // 清空画布
  ctx.clearRect(0, 0, rect.width, rect.height)
  
  if (data.length === 0) return
  
  // 移除之前的鼠标事件监听器
  canvas.removeEventListener('mousemove', handleFeeMouseMove)
  canvas.removeEventListener('mouseleave', handleFeeMouseLeave)
  
  // 添加新的鼠标事件监听器
  canvas.addEventListener('mousemove', handleFeeMouseMove)
  canvas.addEventListener('mouseleave', handleFeeMouseLeave)
  
  // 计算数据范围（增加自适应 padding，使小幅波动也能看见）
  const rawMin = Math.min(...data)
  const rawMax = Math.max(...data)
  let range = rawMax - rawMin
  if (range < 0.5) range = 0.5 // 最小范围，避免看起来成一条直线
  const pad = range * 0.1
  const minValue = rawMin - pad
  const maxValue = rawMax + pad
  
  // 设置边距
  const padding = { top: 10, right: 10, bottom: 20, left: 50 }
  const chartWidth = rect.width - padding.left - padding.right
  const chartHeight = rect.height - padding.top - padding.bottom
  
  // 绘制背景网格
  ctx.strokeStyle = '#f3f4f6'
  ctx.lineWidth = 1
  for (let i = 0; i <= 4; i++) {
    const y = padding.top + (chartHeight / 4) * i
    ctx.beginPath()
    ctx.moveTo(padding.left, y)
    ctx.lineTo(padding.left + chartWidth, y)
    ctx.stroke()
  }
  
  // 绘制折线
  ctx.strokeStyle = color
  ctx.lineWidth = 2
  ctx.beginPath()
  
  data.forEach((value, index) => {
    const x = data.length === 1 
      ? padding.left + chartWidth / 2
      : padding.left + (chartWidth / (data.length - 1)) * index
    const y = padding.top + chartHeight - ((value - minValue) / (maxValue - minValue)) * chartHeight
    
    if (index === 0) {
      ctx.moveTo(x, y)
    } else {
      ctx.lineTo(x, y)
    }
  })
  
  ctx.stroke()
  
  // 绘制数据点
  ctx.fillStyle = color
  data.forEach((value, index) => {
    const x = data.length === 1 
      ? padding.left + chartWidth / 2
      : padding.left + (chartWidth / (data.length - 1)) * index
    const y = padding.top + chartHeight - ((value - minValue) / (maxValue - minValue)) * chartHeight
    
    ctx.beginPath()
    ctx.arc(x, y, 2, 0, 2 * Math.PI)
    ctx.fill()
  })
  
  // 绘制Y轴标签（1位小数 + 单位）
  ctx.fillStyle = '#6b7280'
  ctx.font = '10px sans-serif'
  ctx.textAlign = 'right'
  for (let i = 0; i <= 4; i++) {
    const value = minValue + ((maxValue - minValue) / 4) * (4 - i)
    const y = padding.top + (chartHeight / 4) * i
    ctx.fillText(value.toFixed(1), padding.left - 5, y + 3)
  }
  // Y轴单位
  ctx.textAlign = 'left'
  ctx.fillText('sat/vB', 4, padding.top + 10)
}

// 更新折线图
const updateChart = () => {
  if (feeHistory.value.length === 0) return
  
  // 绘制费率图表
  if (feeChartCanvas.value) {
    const feeData = feeHistory.value.map(item => item.fee)
    drawFeeChart(
      feeChartCanvas.value, 
      feeData, 
      '#f97316', // 橙色
      'BTC费率'
    )
  }
}

// WebSocket监听
const setupWebSocketListeners = () => {
  // 监听费率更新
  const unsubNetwork = subscribeChainEvent('network', (message) => {
    if (message.action === 'fee_update' && message.data) {
      console.log('收到BTC费率更新:', message.data)
      feeLevels.value = message.data as unknown as FeeLevels
      
      // 添加历史数据
      addFeeHistory(message.data as unknown as FeeLevels)
      
      if (feeLevels.value?.normal?.network_congestion) {
        networkCongestion.value = feeLevels.value.normal.network_congestion
      }
    }
  })
  wsUnsubscribes.push(unsubNetwork)
}

onMounted(async () => {
  // 并行加载数据
  await Promise.all([
    loadTransactions(),
    loadTransactionStats()
  ])
  
  // 优先加载一次缓存的BTC费率，页面初次打开即可展示
  getBtcCachedGasRates()
    .then((res) => {
      if ((res as any)?.success && (res as any)?.data) {
        feeLevels.value = (res as any).data as FeeLevels
        addFeeHistory((res as any).data as FeeLevels)
        if (feeLevels.value?.normal?.network_congestion) {
          networkCongestion.value = feeLevels.value.normal.network_congestion
        }
      }
    })
    .catch(() => {})

  setupWebSocketListeners()
  
  // 监听窗口大小变化，重新绘制图表
  window.addEventListener('resize', updateChart)
  
  // 确保DOM完全渲染后再次更新图表
  setTimeout(() => {
    updateChart()
  }, 100)
})

onUnmounted(() => {
  // 组件卸载时取消订阅，避免重复注册导致一次数据多次回调
  wsUnsubscribes.forEach(unsub => { try { unsub() } catch {} })
  wsUnsubscribes.length = 0
  
  // 移除窗口大小变化监听
  window.removeEventListener('resize', updateChart)
})

// -------- 新建交易辅助方法（UTXO/输出管理） --------
const toggleUtxo = (u: BTCUTXO) => {
  // 如果UTXO状态为spent，不允许选择
  if (u.status === 'spent') {
    return
  }
  
  const s = new Set(selectedUtxoIds.value)
  if (s.has(u.id)) s.delete(u.id)
  else s.add(u.id)
  selectedUtxoIds.value = s
}

const addOutput = () => {
  createForm.value.outputs.push({ toAddress: '', amountBtc: null })
}

const removeOutput = (idx: number) => {
  if (createForm.value.outputs.length <= 1) return
  createForm.value.outputs.splice(idx, 1)
}

// -------- 地址自动补全（基于我的地址本） --------
const showSuggestIndex = ref<number | null>(null)

const addressSuggestions = (query: string): PersonalAddressItem[] => {
  const q = (query || '').toLowerCase()
  const source = btcAddresses.value
  if (!q) return source.slice(0, 20)
  return source.filter(a => {
    const addr = (a.address || '').toLowerCase()
    const label = (a.label || '').toLowerCase()
    return addr.includes(q) || label.includes(q)
  }).slice(0, 20)
}

const selectSuggestion = (idx: number, address: string) => {
  if (!createForm.value.outputs[idx]) return
  createForm.value.outputs[idx].toAddress = address
  showSuggestIndex.value = null
}

const onAddressBlur = () => {
  setTimeout(() => { showSuggestIndex.value = null }, 120)
}

// -------- UTXO 排序 --------
const utxoSort = ref<{ key: 'value_satoshi' | 'block_height'; order: 'asc' | 'desc' }>({ key: 'block_height', order: 'desc' })

const toggleUtxoSort = (key: 'value_satoshi' | 'block_height') => {
  if (utxoSort.value.key === key) {
    utxoSort.value.order = utxoSort.value.order === 'asc' ? 'desc' : 'asc'
  } else {
    utxoSort.value.key = key
    utxoSort.value.order = 'desc'
  }
}

const sortedUtxos = computed(() => {
  const list = [...utxos.value]
  const k = utxoSort.value.key
  const o = utxoSort.value.order
  list.sort((a, b) => {
    const av = Number((a as any)[k] || 0)
    const bv = Number((b as any)[k] || 0)
    return o === 'asc' ? av - bv : bv - av
  })
  return list
})
</script>
