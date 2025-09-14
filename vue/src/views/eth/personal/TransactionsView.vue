<template>
  <div class="space-y-6">
    <!-- 页面头部 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="text-2xl font-bold text-gray-900">交易历史</h1>
            <p class="mt-1 text-sm text-gray-500">查看和管理您的交易记录</p>
          </div>
          <div class="flex items-center space-x-4">
            <!-- 网络状态 -->
          <div class="flex items-center space-x-2">
            <div class="w-3 h-3 bg-blue-500 rounded-full"></div>
            <span class="text-sm text-gray-600">ETH 网络</span>
          </div>
            <!-- 网络拥堵状态 -->
            <div v-if="networkCongestion" class="flex items-center space-x-2">
              <div :class="[
                'w-2 h-2 rounded-full',
                networkCongestion === 'high' ? 'bg-red-500' : 
                networkCongestion === 'medium' ? 'bg-yellow-500' : 'bg-green-500'
              ]"></div>
              <span class="text-xs text-gray-500">
                {{ networkCongestion === 'high' ? '高拥堵' : 
                   networkCongestion === 'medium' ? '中等拥堵' : '低拥堵' }}
              </span>
            </div>
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
            最后更新: {{ formatTime(feeLevels.normal.last_updated) }}
          </div>
        </div>
        <div class="flex flex-col lg:flex-row gap-3">
          <!-- 左侧：费率信息 -->
          <div class="lg:w-80 flex-shrink-0">
            <div class="space-y-1.5">
              <!-- Base Fee 显示在最上方 -->
              <div class="border border-gray-200 rounded-lg p-2.5 bg-gray-50">
                <div class="flex items-center justify-between mb-1">
                  <h4 class="text-sm font-medium text-gray-900">Base Fee</h4>
                  <span class="text-xs text-gray-500">基础费用</span>
                </div>
                <div class="text-sm text-gray-700">
                  <span class="font-mono">{{ formatFeeWithPrecision(feeLevels.normal.base_fee) }} Gwei</span>
                </div>
              </div>
              
              <!-- 慢速费率 -->
              <div class="border border-gray-200 rounded-lg p-2.5">
                <div class="flex items-center justify-between mb-1">
                  <h4 class="text-sm font-medium text-gray-900">慢速</h4>
                  <span class="text-xs text-gray-500">0.5x 倍率</span>
                </div>
                <div class="text-sm text-gray-600">
                  Priority: <span class="font-mono">{{ formatFeeWithPrecision(feeLevels.slow.max_priority_fee) }} Gwei</span>
                </div>
              </div>
              
              <!-- 普通费率 -->
              <div class="border border-blue-200 bg-blue-50 rounded-lg p-2.5">
                <div class="flex items-center justify-between mb-1">
                  <h4 class="text-sm font-medium text-blue-900">普通</h4>
                  <span class="text-xs text-blue-600">1.0x 倍率</span>
                </div>
                <div class="text-sm text-blue-800">
                  Priority: <span class="font-mono">{{ formatFeeWithPrecision(feeLevels.normal.max_priority_fee) }} Gwei</span>
                </div>
              </div>
              
              <!-- 快速费率 -->
              <div class="border border-gray-200 rounded-lg p-2.5">
                <div class="flex items-center justify-between mb-1">
                  <h4 class="text-sm font-medium text-gray-900">快速</h4>
                  <span class="text-xs text-gray-500">2.0x 倍率</span>
                </div>
                <div class="text-sm text-gray-600">
                  Priority: <span class="font-mono">{{ formatFeeWithPrecision(feeLevels.fast.max_priority_fee) }} Gwei</span>
                </div>
              </div>
            </div>
          </div>
          
          <!-- 右侧：趋势图 -->
          <div class="flex-1 min-w-0">
            <!-- 两个独立的折线图 -->
            <div class="space-y-4">
              <!-- Base Fee 图表 -->
              <div class="relative">
                <div class="text-sm font-medium text-gray-700 mb-2">Base Fee 趋势</div>
                <div class="h-32">
                  <canvas ref="baseFeeChartCanvas" class="w-full h-full cursor-crosshair"></canvas>
                </div>
                <!-- Base Fee 工具提示 -->
                <div 
                  ref="baseFeeTooltip" 
                  class="absolute bg-gray-800 text-white text-xs px-2 py-1 rounded shadow-lg pointer-events-none opacity-0 transition-opacity duration-200 z-10"
                  style="transform: translate(-50%, -100%); margin-top: -8px;"
                >
                  <div class="font-medium">Base Fee</div>
                  <div class="text-gray-300">Value: <span class="text-white font-mono" id="tooltip-base-fee-value">0</span> Gwei</div>
                </div>
              </div>
              
              <!-- Max Priority Fee 图表 -->
              <div class="relative">
                <div class="text-sm font-medium text-gray-700 mb-2">Max Priority Fee 趋势</div>
                <div class="h-32">
                  <canvas ref="priorityFeeChartCanvas" class="w-full h-full cursor-crosshair"></canvas>
                </div>
                <!-- Priority Fee 工具提示 -->
                <div 
                  ref="priorityFeeTooltip" 
                  class="absolute bg-gray-800 text-white text-xs px-2 py-1 rounded shadow-lg pointer-events-none opacity-0 transition-opacity duration-200 z-10"
                  style="transform: translate(-50%, -100%); margin-top: -8px;"
                >
                  <div class="font-medium">Max Priority Fee</div>
                  <div class="text-gray-300">Value: <span class="text-white font-mono" id="tooltip-priority-fee-value">0</span> Gwei</div>
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
        <div class="grid grid-cols-1 md:grid-cols-4 lg:grid-cols-8 gap-4">
          <div class="text-center">
            <div class="text-2xl font-bold text-gray-600">{{ totalTransactions }}</div>
            <div class="text-sm text-gray-500">总交易</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-gray-500">{{ draftCount }}</div>
            <div class="text-sm text-gray-500">草稿</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-yellow-600">{{ unsignedCount }}</div>
            <div class="text-sm text-gray-500">未签名</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-orange-600">{{ inProgressCount }}</div>
            <div class="text-sm text-gray-500">在途</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-purple-600">{{ packedCount }}</div>
            <div class="text-sm text-gray-500">已打包</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-green-600">{{ confirmedCount }}</div>
            <div class="text-sm text-gray-500">已确认</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-red-600">{{ failedCount }}</div>
            <div class="text-sm text-gray-500">失败</div>
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
              class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 transition-colors"
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
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易类型</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">发送地址</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">接收地址</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">金额</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">创建时间</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="tx in filteredTransactions" :key="tx.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code v-if="tx.tx_hash" 
                        class="bg-gray-100 px-2 py-1 rounded text-xs font-mono cursor-pointer hover:bg-gray-200 transition-colors"
                        :title="tx.tx_hash"
                        @click="copyToClipboard(tx.tx_hash)">
                    {{ tx.tx_hash.substring(0, 10) + '...' + tx.tx_hash.substring(tx.tx_hash.length - 8) }}
                  </code>
                  <span v-else class="text-gray-400">未生成</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <div class="flex flex-col">
                    <span class="font-medium">{{ getTransactionTypeText(tx) }}</span>
                    <span v-if="tx.transaction_type === 'token' && tx.token_contract_address" 
                          class="text-xs text-gray-500 font-mono cursor-pointer hover:text-gray-700 transition-colors"
                          :title="tx.token_contract_address"
                          @click="copyToClipboard(tx.token_contract_address)">
                      {{ tx.token_contract_address.substring(0, 8) }}...{{ tx.token_contract_address.substring(tx.token_contract_address.length - 6) }}
                    </span>
                    <span v-if="tx.contract_operation_type" class="text-xs text-blue-600">
                      {{ getContractOperationText(tx.contract_operation_type) }}
                    </span>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono cursor-pointer hover:bg-gray-200 transition-colors" 
                        :title="tx.from_address"
                        @click="copyToClipboard(tx.from_address)">
                    {{ tx.from_address.substring(0, 10) }}...{{ tx.from_address.substring(tx.from_address.length - 8) }}
                  </code>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono cursor-pointer hover:bg-gray-200 transition-colors" 
                        :title="tx.to_address"
                        @click="copyToClipboard(tx.to_address)">
                    {{ tx.to_address.substring(0, 10) }}...{{ tx.to_address.substring(tx.to_address.length - 8) }}
                  </code>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  <div class="flex flex-col">
                    <span>{{ formatTokenAmount(tx.amount, tx.symbol, tx.token_decimals) }} {{ tx.symbol }}</span>
                    <span v-if="tx.transaction_type === 'token' && tx.token_name" class="text-xs text-gray-500">
                      {{ tx.token_name }}
                    </span>
                  </div>
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
                    v-if="tx.status === 'draft' || tx.status === 'unsigned'"
                    @click="editTransaction(tx)"
                    class="text-indigo-600 hover:text-indigo-900"
                  >
                    编辑
                  </button>
                  <button
                    v-if="tx.status === 'draft' || tx.status === 'unsigned'"
                    @click="exportTransaction(tx)"
                    class="text-blue-600 hover:text-blue-900"
                  >
                    导出交易
                  </button>
                  <button
                    v-if="tx.status === 'unsigned'"
                    @click="openImportModal(tx)"
                    class="text-teal-600 hover:text-teal-900"
                  >
                    导入签名
                  </button>
                  <button
                    v-if="tx.status === 'in_progress'"
                    @click="exportTransaction(tx)"
                    class="text-blue-600 hover:text-blue-900"
                  >
                    导出交易
                  </button>
                  <button
                    v-if="tx.status === 'in_progress'"
                    @click="openImportModal(tx)"
                    class="text-teal-600 hover:text-teal-900"
                  >
                    导入签名
                  </button>
                  <button
                    v-if="tx.status === 'packed' || tx.status === 'confirmed' || tx.status === 'failed'"
                    @click="viewTransaction(tx)"
                    class="text-purple-600 hover:text-purple-900"
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
    <CreateTransactionModal
      :show="showCreateModal"
      :isEditMode="isEditMode"
      :transaction="selectedTransaction"
      @close="handleModalClose"
      @created="handleTransactionCreated"
      @updated="handleTransactionUpdated"
    />


    <!-- QR码预览模态框 -->
    <div v-if="showQRModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-medium text-gray-900">交易QR码</h3>
            <button
              @click="showQRModal = false"
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
              <h4 class="text-md font-medium text-gray-900 mb-2">交易数据</h4>
            </div>
            
            <div class="flex justify-center mb-4">
              <div v-if="qrCodeDataURL" class="bg-white p-4 rounded-lg border-2 border-gray-200">
                <img :src="qrCodeDataURL" alt="交易QR码" class="max-w-full h-auto" />
              </div>
              <div v-else class="bg-gray-100 p-8 rounded-lg">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                <p class="text-sm text-gray-500 mt-2">生成中...</p>
              </div>
            </div>
            
            
          </div>
        </div>
        
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="showQRModal = false"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            关闭
          </button>
          <button
            @click="downloadQRCode"
            :disabled="!qrCodeDataURL"
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            下载QR码
          </button>
        </div>
      </div>
    </div>

    <!-- 费率设置模态框 -->
    <div v-if="showFeeModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-medium text-gray-900">设置交易费率</h3>
        </div>
        
        <div class="px-6 py-4">
          <div class="space-y-4">
            <!-- 手续费模式选择 -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">手续费模式</label>
              <div class="flex space-x-4">
                <label class="flex items-center">
                  <input
                    type="radio"
                    v-model="feeMode"
                    value="auto"
                    class="mr-2 text-blue-600"
                  />
                  <span class="text-sm text-gray-700">自动模式</span>
                </label>
                <label class="flex items-center">
                  <input
                    type="radio"
                    v-model="feeMode"
                    value="manual"
                    class="mr-2 text-blue-600"
                  />
                  <span class="text-sm text-gray-700">手动模式</span>
                </label>
              </div>
            </div>

            <!-- 自动模式 -->
            <div v-if="feeMode === 'auto'" class="space-y-3">
              <div class="grid grid-cols-4 gap-3">
                <label class="flex items-center p-3 border border-gray-200 rounded-lg cursor-pointer hover:border-blue-300">
                  <input
                    type="radio"
                    v-model="autoFeeSpeed"
                    value="ultra-low"
                    class="mr-2 text-blue-600"
                  />
                  <div>
                    <div class="font-medium text-gray-900">极低</div>
                    <div class="text-xs text-gray-500">
                      {{ feeLevels ? formatFeeForDisplay((BigInt(feeLevels.normal.base_fee) + BigInt(1)).toString()) + ' Gwei' : '1 Wei + 20 Gwei' }}
                    </div>
                  </div>
                </label>
                
                <label class="flex items-center p-3 border border-gray-200 rounded-lg cursor-pointer hover:border-blue-300">
                  <input
                    type="radio"
                    v-model="autoFeeSpeed"
                    value="slow"
                    class="mr-2 text-blue-600"
                  />
                  <div>
                    <div class="font-medium text-gray-900">慢速</div>
                    <div class="text-xs text-gray-500">
                      {{ feeLevels ? formatFeeForDisplay(feeLevels.slow.max_fee) + ' Gwei' : autoFeeRates.slow + ' Gwei' }}
                    </div>
                  </div>
                </label>
                
                <label class="flex items-center p-3 border border-blue-200 bg-blue-50 rounded-lg cursor-pointer hover:border-blue-300">
                  <input
                    type="radio"
                    v-model="autoFeeSpeed"
                    value="normal"
                    class="mr-2 text-blue-600"
                  />
                  <div>
                    <div class="font-medium text-blue-900">普通</div>
                    <div class="text-xs text-blue-600">
                      {{ feeLevels ? formatFeeForDisplay(feeLevels.normal.max_fee) + ' Gwei' : autoFeeRates.normal + ' Gwei' }}
                    </div>
                  </div>
                </label>
                
                <label class="flex items-center p-3 border border-gray-200 rounded-lg cursor-pointer hover:border-blue-300">
                  <input
                    type="radio"
                    v-model="autoFeeSpeed"
                    value="fast"
                    class="mr-2 text-blue-600"
                  />
                  <div>
                    <div class="font-medium text-gray-900">快速</div>
                    <div class="text-xs text-gray-500">
                      {{ feeLevels ? formatFeeForDisplay(feeLevels.fast.max_fee) + ' Gwei' : autoFeeRates.fast + ' Gwei' }}
                    </div>
                  </div>
                </label>
              </div>
              
              <!-- 实时费率提示 -->
              <div v-if="feeLevels" class="bg-blue-50 border border-blue-200 rounded-md p-3">
                <div class="flex">
                  <div class="flex-shrink-0">
                    <svg class="h-5 w-5 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                    </svg>
                  </div>
                  <div class="ml-3">
                    <p class="text-sm text-blue-800">
                      使用实时费率数据，网络拥堵状态: 
                      <span :class="[
                        'font-medium',
                        networkCongestion === 'high' ? 'text-red-600' : 
                        networkCongestion === 'medium' ? 'text-yellow-600' : 'text-green-600'
                      ]">
                        {{ networkCongestion === 'high' ? '高拥堵' : 
                           networkCongestion === 'medium' ? '中等拥堵' : '低拥堵' }}
                      </span>
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <!-- 手动模式 -->
            <div v-if="feeMode === 'manual'" class="space-y-4">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">矿工费 (Gwei)</label>
                  <input
                    v-model="manualFee.maxPriorityFeePerGas"
                    type="number"
                    step="0.1"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="1.5"
                  />
                </div>
                
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">最大手续费 (Gwei)</label>
                  <input
                    v-model="manualFee.maxFeePerGas"
                    type="number"
                    step="0.1"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="20"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="showFeeModal = false"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            取消
          </button>
          <button
            @click="confirmFeeAndExport"
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            确认并导出
          </button>
        </div>
      </div>
    </div>

    <!-- 导入签名模态框 -->
    <div v-if="showImportModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-medium text-gray-900">导入签名数据</h3>
        </div>
        
        <div class="px-6 py-4">
          <div class="space-y-4">
            <!-- 选择交易 -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">选择要导入签名的交易</label>
              <select v-model="selectedImportTransactionId" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="">请选择交易</option>
                <option v-for="tx in transactionsList.filter(t => t.status === 'unsigned' || t.status === 'in_progress')" :key="tx.id" :value="tx.id">
                  ID: {{ tx.id }} - {{ tx.from_address.substring(0, 10) }}... → {{ tx.to_address.substring(0, 10) }}... ({{ tx.amount }} {{ tx.symbol }})
                </option>
              </select>
            </div>
            
            <!-- 签名数据 -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">签名数据</label>
              <textarea
                v-model="importSignature"
                rows="6"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="请粘贴从离线程序导出的签名数据..."
              ></textarea>
            </div>
            
            <div class="bg-blue-50 border border-blue-200 rounded-md p-3">
              <div class="flex">
                <div class="flex-shrink-0">
                  <svg class="h-5 w-5 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                  </svg>
                </div>
                <div class="ml-3">
                  <p class="text-sm text-blue-800">
                    支持导入签名数据：完整的签名交易字符串或包含v,r,s字段的JSON格式。导入后交易状态将变为"未发送"
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="showImportModal = false"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            取消
          </button>
          <button
            @click="importSignatureData"
            :disabled="!importSignature.trim() || !selectedImportTransactionId || isImporting"
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 flex items-center"
          >
            <svg v-if="isImporting" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            {{ isImporting ? '导入中...' : '导入签名' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import type { UserTransaction, UserTransactionStatsResponse } from '@/types'
import CreateTransactionModal from '@/components/eth/personal/CreateTransactionModal.vue'
import { getUserTransactions, getUserTransactionStats, exportTransaction as exportTransactionAPI, importSignature as importSignatureAPI } from '@/api/user-transactions'
import { getGasRates } from '@/api/gas'
import { useChainWebSocket } from '@/composables/useWebSocket'
import { formatTokenAmount } from '@/utils/amountFormatter'
import { convertWeiToGwei, formatFeeForDisplay } from '@/utils/unitConverter'
import type { FeeLevels } from '@/types'
import type { TransactionStatusUpdate } from '@/utils/websocket'

// 响应式数据
const showCreateModal = ref(false)
const showImportModal = ref(false)
const showFeeModal = ref(false) // 费率设置模态框
const selectedStatus = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const totalItems = ref(0)
const totalPages = ref(0)
const importSignature = ref('')
const selectedTransaction = ref<UserTransaction | null>(null)
const selectedImportTransactionId = ref<number | ''>('')
const isEditMode = ref(false) // 是否为编辑模式

// 费率设置相关
const feeMode = ref<'auto' | 'manual'>('auto')
const autoFeeSpeed = ref<'ultra-low' | 'slow' | 'normal' | 'fast'>('normal')
const autoFeeRates = {
  'ultra-low': 0.000000001, // 1 Wei
  slow: 1.5,
  normal: 2.0,
  fast: 2.5
}
const manualFee = ref({
  maxPriorityFeePerGas: '1.5',
  maxFeePerGas: '20'
})

// QR码相关状态
const showQRModal = ref(false)
const qrCodeDataURL = ref<string>('')
const selectedQRTransaction = ref<UserTransaction | null>(null)

// 导入签名相关状态
const isImporting = ref(false)

// 交易统计
const totalTransactions = ref(0)
const unsignedCount = ref(0)
const inProgressCount = ref(0)
const confirmedCount = ref(0)
const draftCount = ref(0)
const packedCount = ref(0)
const failedCount = ref(0)

// 交易列表
const transactionsList = ref<UserTransaction[]>([])

// WebSocket相关
const { subscribeChainEvent } = useChainWebSocket('eth')
// 收集本组件的取消订阅函数，避免重复回调
const wsUnsubscribes: Array<() => void> = []

// 费率数据
const feeLevels = ref<FeeLevels | null>(null)
const networkCongestion = ref<string>('normal')

// 费率历史数据存储（用于折线图）
const feeHistory = ref<Array<{
  timestamp: number
  baseFee: number
  maxPriorityFee: number
  maxFee: number
}>>([])

// 图表相关
const baseFeeChartCanvas = ref<HTMLCanvasElement | null>(null)
const priorityFeeChartCanvas = ref<HTMLCanvasElement | null>(null)
const baseFeeTooltip = ref<HTMLDivElement | null>(null)
const priorityFeeTooltip = ref<HTMLDivElement | null>(null)
let chartInstance: any = null

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
    case 'unsigned': return 'bg-yellow-100 text-yellow-800'
    case 'in_progress': return 'bg-orange-100 text-orange-800'
    case 'packed': return 'bg-purple-100 text-purple-800'
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

// 获取交易类型文本
const getTransactionTypeText = (tx: UserTransaction) => {
  // 如果是查询余额操作，显示为"查询余额"
  if (tx.contract_operation_type === 'balanceOf') {
    return `${tx.symbol} 查询余额`
  }
  
  if (tx.transaction_type === 'coin' || tx.transaction_type === 'native') {
    return 'ETH 转账'
  } else if (tx.transaction_type === 'token') {
    return `${tx.symbol} 代币转账`
  } else if (tx.symbol === 'ETH') {
    return 'ETH 转账'
  } else {
    return `${tx.symbol} 代币转账`
  }
}

// 获取合约操作类型文本
const getContractOperationText = (type: string) => {
  switch (type) {
    case 'transfer': return '转账'
    case 'approve': return '授权'
    case 'transferFrom': return '代币转移'
    case 'mint': return '铸造'
    case 'burn': return '销毁'
    case 'setApprovalForAll': return '设置授权'
    case 'transferOwnership': return '转让所有权'
    default: return type
  }
}

// 格式化费率，保留9位小数（Gwei精度）
const formatFeeWithPrecision = (feeInWei: string | number) => {
  if (!feeInWei) return '0.000000000'
  
  // 转换为Gwei（除以1e9）
  const feeInGwei = typeof feeInWei === 'string' ? parseFloat(feeInWei) / 1e9 : feeInWei / 1e9
  
  // 保留9位小数
  return feeInGwei.toFixed(9)
}

// 格式化时间
const formatTime = (timestamp: string | number | undefined) => {
  if (!timestamp) return '未知时间'
  
  let date: Date
  if (typeof timestamp === 'number') {
    // 判断是秒还是毫秒时间戳
    // 如果时间戳小于 1e12，认为是秒时间戳，需要转换为毫秒
    if (timestamp < 1e12) {
      date = new Date(timestamp * 1000)
    } else {
      date = new Date(timestamp)
    }
  } else if (typeof timestamp === 'string') {
    // 如果是字符串，尝试解析
    date = new Date(timestamp)
  } else {
    return '未知时间'
  }
  
  // 检查日期是否有效
  if (isNaN(date.getTime())) {
    return '无效时间'
  }
  
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}


// 复制到剪贴板
const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    // 使用更友好的提示方式
    const toast = document.createElement('div')
    toast.className = 'fixed top-4 right-4 bg-green-500 text-white px-4 py-2 rounded-md shadow-lg z-50 transition-opacity duration-300'
    toast.textContent = '地址已复制到剪贴板！'
    document.body.appendChild(toast)
    
    // 3秒后自动消失
    setTimeout(() => {
      toast.style.opacity = '0'
      setTimeout(() => {
        document.body.removeChild(toast)
      }, 300)
    }, 3000)
  } catch (err) {
    console.error('复制失败:', err)
    // 降级方案：使用传统方法
    const textArea = document.createElement('textarea')
    textArea.value = text
    document.body.appendChild(textArea)
    textArea.select()
    try {
      document.execCommand('copy')
      alert('地址已复制到剪贴板！')
    } catch (fallbackErr) {
      alert('复制失败，请手动复制：' + text)
    }
    document.body.removeChild(textArea)
  }
}

// 导出交易 - 先显示费率设置模态框
const exportTransaction = (tx: UserTransaction) => {
  selectedTransaction.value = tx
  showFeeModal.value = true
}

// 确认费率并导出交易
const confirmFeeAndExport = async () => {
  if (!selectedTransaction.value) return
  
  try {
    // 准备费率数据
    let feeData: any = {}
    console.log('🔍 前端费率设置调试信息:')
    console.log('  feeMode.value:', feeMode.value)
    console.log('  feeLevels.value:', feeLevels.value)
    console.log('  autoFeeSpeed.value:', autoFeeSpeed.value)
    console.log('  manualFee.value:', manualFee.value)
    
    if (feeMode.value === 'auto') {
      // 使用实时费率数据
      if (feeLevels.value) {
        if (autoFeeSpeed.value === 'ultra-low') {
          // 极低模式：Priority = 1 Wei, Max = Base + 1 Wei
          const baseFeeWei = feeLevels.value.normal.base_fee
          const maxFeeWei = (BigInt(baseFeeWei) + BigInt(1)).toString()
          feeData = {
            maxPriorityFeePerGas: '1', // 1 Wei
            maxFeePerGas: maxFeeWei    // Base + 1 Wei
          }
          console.log('  ✅ 使用极低费率数据 (Wei):', feeData)
        } else {
          const selectedFee = feeLevels.value[autoFeeSpeed.value]
          console.log('  selectedFee:', selectedFee)
          // 实时费率数据已经是Wei单位，直接使用
          feeData = {
            maxPriorityFeePerGas: selectedFee.max_priority_fee,
            maxFeePerGas: selectedFee.max_fee
          }
          console.log('  ✅ 使用实时费率数据 (Wei):', feeData)
        }
      } else {
        // 降级到默认费率，转换为Wei
        if (autoFeeSpeed.value === 'ultra-low') {
          // 极低模式：Priority = 1 Wei, Max = 20 Gwei + 1 Wei
          const baseFeeWei = (20 * 1e9).toString() // 20 Gwei
          const maxFeeWei = (BigInt(baseFeeWei) + BigInt(1)).toString()
          feeData = {
            maxPriorityFeePerGas: '1', // 1 Wei
            maxFeePerGas: maxFeeWei    // 20 Gwei + 1 Wei
          }
          console.log('  ⚠️ 使用极低默认费率数据 (Wei):', feeData)
        } else {
          const gasPrice = autoFeeRates[autoFeeSpeed.value]
          feeData = {
            maxPriorityFeePerGas: (gasPrice * 1e9).toString(), // 转换为Wei
            maxFeePerGas: (gasPrice * 1.5 * 1e9).toString() // 转换为Wei
          }
          console.log('  ⚠️ 使用默认费率数据 (Wei):', feeData)
        }
      }
    } else {
      // 手动模式，将Gwei转换为Wei
      const priorityFeeWei = (parseFloat(manualFee.value.maxPriorityFeePerGas) * 1e9).toString()
      const maxFeeWei = (parseFloat(manualFee.value.maxFeePerGas) * 1e9).toString()
      feeData = {
        maxPriorityFeePerGas: priorityFeeWei,
        maxFeePerGas: maxFeeWei
      }
      console.log('  ✅ 使用手动费率数据 (Wei):', feeData)
    }
    
    // 调用导出API，传递费率数据
    const response = await exportTransactionAPI(selectedTransaction.value.id, feeData)
    if (response.success) {
      // 成功导出后，更新本地状态为未签名
      selectedTransaction.value.status = 'unsigned'
      
      // 刷新列表与统计，确保计数正确
      loadTransactions()
      loadTransactionStats()

      // 关闭费率设置模态框
      showFeeModal.value = false
      
      // 显示QR码预览模态框
      selectedQRTransaction.value = selectedTransaction.value
      showQRModal.value = true
      qrCodeDataURL.value = '' // 重置QR码
      
      // 异步生成QR码
      generateQRCode(response.data, selectedTransaction.value)
      
      
    } else {
      alert('导出交易失败: ' + response.message)
    }
  } catch (error) {
    console.error('导出交易失败:', error)
    alert('导出交易失败，请重试')
  }
}

// 生成QR码（用于预览）
const generateQRCode = async (transactionData: any, tx: UserTransaction) => {
  try {
    // 动态导入QRCode库
    const QRCode = await import('qrcode')
    
    // 创建精简的交易数据结构，只包含签名必需的核心字段
    const minimalTxData = createMinimalTransactionData(tx, transactionData)
    
    // 将精简数据转换为JSON字符串
    const transactionJson = JSON.stringify(minimalTxData, null, 0) // 不格式化，减少字符数
    
    console.log('准备生成QR码:', {
      dataLength: transactionJson.length,
      dataPreview: transactionJson.substring(0, 100) + '...'
    })
    
    
    
    
    
    
    // 生成QR码配置 - 大幅提高分辨率和质量
    const qrOptions = {
      type: 'image/png' as const,
      quality: 1.0, // 最高质量
      margin: 4, // 进一步增加边距，提高识别率
      color: {
        dark: '#000000',
        light: '#FFFFFF'
      },
      width: 2048, // 大幅增加尺寸，从512提升到2048
      errorCorrectionLevel: 'H' as const, // 使用最高错误纠正级别
      scale: 16, // 增加缩放比例，提高清晰度
      rendererOpts: {
        quality: 1.0, // 渲染质量
        precision: 'high' // 高精度渲染
      }
    }
    
    // 生成QR码数据URL - 使用多种策略
    let qrDataURL: string
    try {
      // 策略1: 尝试生成SVG格式的QR码，然后转换为PNG
      console.log('尝试生成SVG格式QR码...')
      const svgString = await QRCode.toString(transactionJson, {
        type: 'svg',
        width: 2048,
        margin: 12,
        color: {
          dark: '#000000',
          light: '#FFFFFF'
        },
        errorCorrectionLevel: 'H'
      })
      
      // 将SVG转换为PNG
      const canvas = document.createElement('canvas')
      const ctx = canvas.getContext('2d')
      const img = new Image()
      
      qrDataURL = await new Promise<string>((resolve, reject) => {
        img.onload = () => {
          canvas.width = 2048
          canvas.height = 2048
          ctx?.drawImage(img, 0, 0, 2048, 2048)
          const pngDataURL = canvas.toDataURL('image/png', 1.0)
          console.log('SVG转PNG成功，QR码尺寸:', 2048)
          resolve(pngDataURL)
        }
        img.onerror = () => {
          console.warn('SVG转PNG失败，尝试标准PNG生成...')
          // 回退到标准PNG生成
          QRCode.toDataURL(transactionJson, qrOptions).then(resolve).catch(reject)
        }
        img.src = 'data:image/svg+xml;base64,' + btoa(svgString)
      })
      
    } catch (error) {
      console.warn('SVG生成失败，尝试标准PNG生成:', error)
      try {
        qrDataURL = await QRCode.toDataURL(transactionJson, qrOptions)
      } catch (fallbackError) {
        console.warn('高分辨率PNG生成失败，尝试标准分辨率:', fallbackError)
        // 如果高分辨率失败，使用标准分辨率
        const fallbackOptions = {
          type: 'image/png' as const,
          quality: 1.0,
          margin: 8,
          color: {
            dark: '#000000',
            light: '#FFFFFF'
          },
          width: 1024,
          errorCorrectionLevel: 'H' as const,
          scale: 8
        }
        qrDataURL = await QRCode.toDataURL(transactionJson, fallbackOptions)
      }
    }
    
    qrCodeDataURL.value = qrDataURL
    
    console.log('QR码生成完成:', {
      dataLength: transactionJson.length,
      qrSize: qrOptions.width,
      errorCorrection: qrOptions.errorCorrectionLevel,
      margin: qrOptions.margin,
      actualSize: qrDataURL.length
    })
    

  } catch (error) {
    console.error('生成QR码失败:', error)
    qrCodeDataURL.value = ''
    alert('QR码生成失败，请重试')
  }
}

// 创建精简的交易数据结构
const createMinimalTransactionData = (tx: UserTransaction, fullData: any) => {
  // 优先使用后端保存的完整数据，确保数据一致性
  const minimalData: any = {
    // 交易标识
    id: tx.id,
    
    // 链信息 - 优先使用后端保存的chainId
    chainId: fullData.chain_id || (tx.chain === 'eth' ? '1' : tx.chain),
    type: tx.chain, // 添加类型字段：eth 或 btc
    
    // 交易核心字段
    nonce: fullData.nonce || tx.nonce || 0, // 优先使用API返回的nonce
    from: tx.from_address, // 添加from字段用于签名程序自动匹配私钥
    to: tx.transaction_type === 'token' && tx.token_contract_address ? tx.token_contract_address : tx.to_address,
    value: tx.transaction_type === 'token' ? '0x0' : convertToHexString(tx.amount || '0'), // 代币转账value为0，ETH转账使用整数金额的十六进制格式
    data: fullData.tx_data || generateContractData(tx, fullData), // 优先使用后端保存的tx_data
    
    // EIP-1559费率字段 - 转换为Gwei单位供签名程序使用
    maxPriorityFeePerGas: convertWeiToGwei(fullData.max_priority_fee_per_gas || tx.max_priority_fee_per_gas || '2000000000'),
    maxFeePerGas: convertWeiToGwei(fullData.max_fee_per_gas || tx.max_fee_per_gas || '30000000000')
  }

  // 将后端估算的GasLimit透传给签名器（数字类型）
  if (fullData.gas_limit || tx.gas_limit) {
    try {
      const gas = fullData.gas_limit ?? tx.gas_limit
      minimalData.gas = typeof gas === 'string' ? parseInt(gas, 10) : Number(gas)
    } catch (e) {
      // 忽略解析失败，保持未设置
    }
  }
  
  // 添加AccessList - 直接使用后端计算好的数据
  if (fullData.access_list && fullData.access_list !== '[]') {
    try {
      minimalData.accessList = JSON.parse(fullData.access_list)
    } catch (error) {
      console.warn('解析AccessList失败:', error)
      // 如果解析失败，不添加AccessList
    }
  }
  
  return minimalData
}


// 转换金额为十六进制格式
const convertToHexString = (amount: string) => {
  if (!amount || amount === '0') return '0x0'
  
  // 如果已经包含0x前缀，直接返回
  if (amount.startsWith('0x')) {
    return amount
  }
  
  // 检查是否是小数，如果是小数，先转换为整数
  let intAmount: bigint
  try {
    if (amount.includes('.')) {
      // 如果是小数，先转换为整数（假设是ETH，使用18位精度）
      const numAmount = parseFloat(amount)
      const weiAmount = Math.floor(numAmount * Math.pow(10, 18))
      intAmount = BigInt(weiAmount.toString())
    } else {
      // 如果已经是整数格式，直接转换
      intAmount = BigInt(amount)
    }
  } catch (error) {
    console.error(`无法转换金额为BigInt: ${amount}`, error)
    return '0x0'
  }
  
  // 转换为十六进制字符串
  const hexString = intAmount.toString(16)
  return '0x' + hexString
}

// 根据操作类型生成合约调用数据
const generateContractData = (tx: UserTransaction, fullData: any) => {
  // 如果有完整的data，优先使用
  if (fullData.data && fullData.data !== '0x') {
    return fullData.data
  }
  
  // 如果是代币交易，根据操作类型生成data
  if (tx.transaction_type === 'token' && tx.token_contract_address) {
    switch (tx.contract_operation_type) {
      case 'balanceOf':
        // balanceOf(address) 函数调用
        return generateBalanceOfData(tx.from_address)
        
      case 'transfer':
        // transfer(address,uint256) 函数调用
        return generateTransferData(tx.to_address, tx.amount)
        
      case 'approve':
        // approve(address,uint256) 函数调用
        return generateApproveData(tx.to_address, tx.amount)
        
      case 'transferFrom':
        // transferFrom(address,address,uint256) 函数调用
        return generateTransferFromData(tx.from_address, tx.to_address, tx.amount)
        
      default:
        return '0x'
    }
  }
  
  // ETH转账，data为空
  return '0x'
}

// 生成balanceOf函数调用数据
const generateBalanceOfData = (address: string) => {
  // balanceOf(address) 函数选择器: 0x70a08231
  const functionSelector = '0x70a08231'
  // 地址参数（32字节，右对齐）
  const addressParam = address.slice(2).padStart(64, '0')
  return functionSelector + addressParam
}

// 生成transfer函数调用数据
const generateTransferData = (toAddress: string, amount: string) => {
  // transfer(address,uint256) 函数选择器: 0xa9059cbb
  const functionSelector = '0xa9059cbb'
  // 接收地址参数（32字节，右对齐）
  const toParam = toAddress.slice(2).padStart(64, '0')
  // 金额参数（32字节，直接使用整数金额的十六进制）
  const amountHex = convertToHexString(amount)
  // 确保去掉0x前缀
  const amountParam = amountHex.startsWith('0x') ? amountHex.slice(2) : amountHex
  const paddedAmountParam = amountParam.padStart(64, '0')
  return functionSelector + toParam + paddedAmountParam
}

// 生成approve函数调用数据
const generateApproveData = (spenderAddress: string, amount: string) => {
  // approve(address,uint256) 函数选择器: 0x095ea7b3
  const functionSelector = '0x095ea7b3'
  // 被授权者地址参数（32字节，右对齐）
  const spenderParam = spenderAddress.slice(2).padStart(64, '0')
  // 授权金额参数（32字节，直接使用整数金额的十六进制）
  const amountHex = convertToHexString(amount)
  const amountParam = amountHex.slice(2).padStart(64, '0')
  return functionSelector + spenderParam + amountParam
}

// 生成transferFrom函数调用数据
const generateTransferFromData = (fromAddress: string, toAddress: string, amount: string) => {
  // transferFrom(address,address,uint256) 函数选择器: 0x23b872dd
  const functionSelector = '0x23b872dd'
  // 发送者地址参数（32字节，右对齐）
  const fromParam = fromAddress.slice(2).padStart(64, '0')
  // 接收者地址参数（32字节，右对齐）
  const toParam = toAddress.slice(2).padStart(64, '0')
  // 金额参数（32字节，直接使用整数金额的十六进制）
  const amountHex = convertToHexString(amount)
  const amountParam = amountHex.slice(2).padStart(64, '0')
  return functionSelector + fromParam + toParam + amountParam
}


// 下载QR码
const downloadQRCode = () => {
  if (!qrCodeDataURL.value || !selectedQRTransaction.value) return
  
  const tx = selectedQRTransaction.value
  const link = document.createElement('a')
  link.href = qrCodeDataURL.value
  link.download = `transaction_${tx.id}_${tx.chain}_${tx.symbol}_qr.png`
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


// 查看交易
const viewTransaction = (tx: UserTransaction) => {
  // 显示交易详情
  
  let details = `交易详情:
  
状态: ${getStatusText(tx.status)}
链类型: ${tx.chain.toUpperCase()}
币种: ${tx.symbol}
${tx.contract_operation_type === 'balanceOf' ? '查询地址' : '发送地址'}: ${tx.from_address}
${tx.contract_operation_type === 'balanceOf' ? '' : `接收地址: ${tx.to_address}
金额: ${formatTokenAmount(tx.amount, tx.symbol, tx.token_decimals)} ${tx.symbol}`}
Gas限制: ${tx.gas_limit || '未设置'}
Gas价格: ${tx.gas_price || '未设置'} Gwei
Nonce: ${tx.nonce || '自动获取'}
交易哈希: ${tx.tx_hash || '未生成'}
区块高度: ${tx.block_height || '未确认'}
确认数: ${tx.confirmations || 0}
备注: ${tx.remark || '无'}
创建时间: ${formatTime(tx.created_at)}
更新时间: ${formatTime(tx.updated_at)}`

  // 添加错误信息（如果有）
  if (tx.error_msg) {
    details += `

=== 错误信息 ===
错误详情: ${tx.error_msg}`
  }

  // 添加ERC-20相关信息
  if (tx.transaction_type === 'token') {
    details += `

=== ERC-20 代币信息 ===
交易类型: 代币转账
合约操作: ${getContractOperationText(tx.contract_operation_type || '')}
代币合约地址: ${tx.token_contract_address || '未设置'}
代币名称: ${tx.token_name || '未设置'}
代币精度: ${tx.token_decimals || '未设置'}`
  } else {
    details += `

=== 交易类型 ===
交易类型: ETH转账`
  }
  
  alert(details)
}

// 编辑交易
const editTransaction = (tx: UserTransaction) => {
  selectedTransaction.value = tx
  isEditMode.value = true
  showCreateModal.value = true // 使用新建交易模态框进行编辑
}

// 导入签名数据
const importSignatureData = async () => {
  // 防止重复提交
  if (isImporting.value) return
  
  try {
    if (!selectedImportTransactionId.value) {
      alert('请选择要导入签名的交易')
      return
    }
    
    // 设置加载状态
    isImporting.value = true
    
    const id = selectedImportTransactionId.value as number
    
    // 解析签名数据
    const signatureData = parseSignatureData(importSignature.value)
    if (!signatureData) {
      alert('签名数据格式错误，请检查数据格式')
      return
    }
    
    // 验证ID是否匹配
    if (signatureData.id !== undefined && signatureData.id !== id) {
      alert(`签名数据ID(${signatureData.id})与所选交易ID(${id})不匹配`)
      return
    }
    
    // 调用导入签名API
    const response = await importSignatureAPI(id, { 
      id, 
      signed_tx: signatureData.signedTx,
      v: signatureData.v,
      r: signatureData.r,
      s: signatureData.s
    })
    
    if (response.success) {
      alert('导入签名成功！')
      loadTransactions()
      loadTransactionStats()
      showImportModal.value = false
      importSignature.value = ''
      selectedImportTransactionId.value = ''
    } else {
      alert('导入签名失败: ' + response.message)
    }
  } catch (error) {
    console.error('导入签名失败:', error)
    alert('导入签名失败，请重试')
  } finally {
    // 无论成功还是失败，都要重置加载状态
    isImporting.value = false
  }
}

// 解析签名数据
const parseSignatureData = (signatureText: string) => {
  try {
    // 尝试解析JSON格式的签名数据
    const data = JSON.parse(signatureText)
    
    // 检查是否包含签名字段
    if (data.v && data.r && data.s) {
      return {
        signedTx: data.signedTx || signatureText, // 如果有完整的签名交易，使用它
        v: data.v,
        r: data.r,
        s: data.s
      }
    }
    
    // 支持格式：{"id":2,"signer":"0x..."}
    if ((typeof data.id === 'number' || typeof data.id === 'string') && typeof data.signer === 'string') {
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
    if (signatureText.startsWith('0x') && signatureText.length > 100) {
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

// 从操作列打开导入签名模态框并预选交易
const openImportModal = (tx: UserTransaction) => {
  selectedImportTransactionId.value = tx.id
  showImportModal.value = true
}

// 处理交易创建成功
const handleTransactionCreated = (transaction: any) => {
  // 刷新交易列表和统计
  loadTransactions()
  loadTransactionStats()
  isEditMode.value = false // 关闭编辑模式
  selectedTransaction.value = null // 清除选中的交易
}


// 打开创建交易模态框
const openCreateModal = () => {
  // 重置所有状态
  isEditMode.value = false
  selectedTransaction.value = null
  showCreateModal.value = true
}

// 处理模态框关闭
const handleModalClose = () => {
  showCreateModal.value = false
  isEditMode.value = false
  selectedTransaction.value = null
}

// 处理交易更新
const handleTransactionUpdated = (transaction: any) => {
  // 刷新交易列表和统计
  loadTransactions()
  loadTransactionStats()
  isEditMode.value = false // 关闭编辑模式
  selectedTransaction.value = null // 清除选中的交易
}

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
    const response = await getUserTransactions({
      page: currentPage.value,
      page_size: pageSize.value,
      status: selectedStatus.value,
      chain: 'eth' // 只加载ETH链的交易
    })
    
    if (response.success) {
      transactionsList.value = response.data.transactions
      totalItems.value = response.data.total
      totalPages.value = Math.ceil(totalItems.value / pageSize.value)
    }
  } catch (error) {
    console.error('加载交易数据失败:', error)
  }
}

// 加载交易统计
const loadTransactionStats = async () => {
  try {
    const response = await getUserTransactionStats({ chain: 'eth' })
    
    if (response.success) {
      const stats = response.data
      totalTransactions.value = stats.total_transactions
      draftCount.value = stats.draft_count
      unsignedCount.value = stats.unsigned_count
      inProgressCount.value = stats.in_progress_count
      packedCount.value = stats.packed_count
      confirmedCount.value = stats.confirmed_count
      failedCount.value = stats.failed_count
    }
  } catch (error) {
    console.error('加载交易统计失败:', error)
  }
}

// 添加费率历史数据
const addFeeHistory = (feeData: FeeLevels) => {
  const now = Date.now()
  const historyItem = {
    timestamp: now,
    baseFee: parseFloat(feeData.normal.base_fee || '0') / 1e9, // 转换为Gwei
    maxPriorityFee: parseFloat(feeData.normal.max_priority_fee) / 1e9,
    maxFee: parseFloat(feeData.normal.max_fee) / 1e9
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

// Base Fee 鼠标移动事件处理
const handleBaseFeeMouseMove = (event: MouseEvent) => {
  if (!baseFeeChartCanvas.value || !baseFeeTooltip.value || feeHistory.value.length === 0) return
  
  const canvas = baseFeeChartCanvas.value
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
    const baseFeeElement = document.getElementById('tooltip-base-fee-value')
    if (baseFeeElement) baseFeeElement.textContent = data.baseFee.toFixed(9)
    
    // 计算相对于父容器的位置
    const parentRect = baseFeeTooltip.value.parentElement?.getBoundingClientRect()
    
    if (parentRect) {
      const relativeX = event.clientX - parentRect.left
      const relativeY = event.clientY - parentRect.top
      
      baseFeeTooltip.value.style.left = relativeX + 'px'
      baseFeeTooltip.value.style.top = (relativeY - 10) + 'px'
      baseFeeTooltip.value.style.opacity = '1'
    }
  }
}

// Priority Fee 鼠标移动事件处理
const handlePriorityFeeMouseMove = (event: MouseEvent) => {
  if (!priorityFeeChartCanvas.value || !priorityFeeTooltip.value || feeHistory.value.length === 0) return
  
  const canvas = priorityFeeChartCanvas.value
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
    const priorityFeeElement = document.getElementById('tooltip-priority-fee-value')
    if (priorityFeeElement) priorityFeeElement.textContent = data.maxPriorityFee.toFixed(9)
    
    // 计算相对于父容器的位置
    const parentRect = priorityFeeTooltip.value.parentElement?.getBoundingClientRect()
    
    if (parentRect) {
      const relativeX = event.clientX - parentRect.left
      const relativeY = event.clientY - parentRect.top
      
      priorityFeeTooltip.value.style.left = relativeX + 'px'
      priorityFeeTooltip.value.style.top = (relativeY - 10) + 'px'
      priorityFeeTooltip.value.style.opacity = '1'
    }
  }
}

// Base Fee 鼠标离开事件处理
const handleBaseFeeMouseLeave = () => {
  if (baseFeeTooltip.value) {
    baseFeeTooltip.value.style.opacity = '0'
  }
}

// Priority Fee 鼠标离开事件处理
const handlePriorityFeeMouseLeave = () => {
  if (priorityFeeTooltip.value) {
    priorityFeeTooltip.value.style.opacity = '0'
  }
}

// 绘制单个图表的通用函数
const drawSingleChart = (canvas: HTMLCanvasElement, data: number[], color: string, title: string, mouseMoveHandler: (event: MouseEvent) => void, mouseLeaveHandler: () => void) => {
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
  canvas.removeEventListener('mousemove', mouseMoveHandler)
  canvas.removeEventListener('mouseleave', mouseLeaveHandler)
  
  // 添加新的鼠标事件监听器
  canvas.addEventListener('mousemove', mouseMoveHandler)
  canvas.addEventListener('mouseleave', mouseLeaveHandler)
  
  // 计算数据范围
  const minValue = Math.min(...data)
  const maxValue = Math.max(...data)
  const valueRange = maxValue - minValue || 1
  
  // 设置边距
  const padding = { top: 10, right: 10, bottom: 20, left: 40 }
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
    const y = padding.top + chartHeight - ((value - minValue) / valueRange) * chartHeight
    
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
    const y = padding.top + chartHeight - ((value - minValue) / valueRange) * chartHeight
    
    ctx.beginPath()
    ctx.arc(x, y, 2, 0, 2 * Math.PI)
    ctx.fill()
  })
  
  // 绘制Y轴标签
  ctx.fillStyle = '#6b7280'
  ctx.font = '10px sans-serif'
  ctx.textAlign = 'right'
  for (let i = 0; i <= 4; i++) {
    const value = minValue + (valueRange / 4) * (4 - i)
    const y = padding.top + (chartHeight / 4) * i
    ctx.fillText(value.toFixed(6), padding.left - 5, y + 3)
  }
}

// 更新折线图
const updateChart = () => {
  if (feeHistory.value.length === 0) return
  
  // 绘制 Base Fee 图表
  if (baseFeeChartCanvas.value) {
    const baseFeeData = feeHistory.value.map(item => item.baseFee)
    drawSingleChart(
      baseFeeChartCanvas.value, 
      baseFeeData, 
      '#6b7280', 
      'Base Fee',
      handleBaseFeeMouseMove,
      handleBaseFeeMouseLeave
    )
  }
  
  // 绘制 Max Priority Fee 图表
  if (priorityFeeChartCanvas.value) {
    const priorityFeeData = feeHistory.value.map(item => item.maxPriorityFee)
    drawSingleChart(
      priorityFeeChartCanvas.value, 
      priorityFeeData, 
      '#3b82f6', 
      'Max Priority Fee',
      handlePriorityFeeMouseMove,
      handlePriorityFeeMouseLeave
    )
  }
}

// 加载Gas费率数据
const loadGasRates = async () => {
  try {
    // console.log('🔄 加载Gas费率数据...')
    const response = await getGasRates({ chain: 'eth' })
    
    if (response.success) {
      // console.log('✅ Gas费率数据加载成功:', response.data)
      feeLevels.value = response.data
      
      // 添加历史数据
      addFeeHistory(response.data)
      
      // 更新网络拥堵状态
      if (response.data?.normal?.network_congestion) {
        networkCongestion.value = response.data.normal.network_congestion
      }
      
      // 立即更新图表显示
      updateChart()
    } else {
      console.warn('⚠️ Gas费率数据加载失败:', response.message)
    }
  } catch (error) {
    console.error('❌ 加载Gas费率数据失败:', error)
  }
}

// 监听状态筛选变化
watch(selectedStatus, () => {
  currentPage.value = 1
  loadTransactions()
})


// WebSocket监听
const setupWebSocketListeners = () => {
  // 监听费率更新
  const unsubNetwork = subscribeChainEvent('network', (message) => {
    if (message.action === 'fee_update' && message.data) {
      // console.log('收到费率更新:', message.data)
      feeLevels.value = message.data as unknown as FeeLevels
      
      // 添加历史数据
      addFeeHistory(message.data as unknown as FeeLevels)
      
      if (feeLevels.value?.normal?.network_congestion) {
        networkCongestion.value = feeLevels.value.normal.network_congestion
      }
    }
  })
  wsUnsubscribes.push(unsubNetwork)

  // 监听交易状态更新
  const unsubTx = subscribeChainEvent('transaction', (message) => {
    if (message.action === 'status_update' && message.data) {
      // console.log('收到交易状态更新:', message.data)
      const statusUpdate = message.data as unknown as TransactionStatusUpdate
      
      // 更新本地交易列表中的对应交易
      const txIndex = transactionsList.value.findIndex(tx => tx.id === statusUpdate.id)
      if (txIndex !== -1) {
        const tx = transactionsList.value[txIndex]
        tx.status = statusUpdate.status
        if (statusUpdate.tx_hash) tx.tx_hash = statusUpdate.tx_hash
        if (statusUpdate.block_height) tx.block_height = statusUpdate.block_height
        if (statusUpdate.confirmations) tx.confirmations = statusUpdate.confirmations
        if (statusUpdate.error_msg) tx.error_msg = statusUpdate.error_msg
        tx.updated_at = statusUpdate.updated_at
        
        // 触发响应式更新
        transactionsList.value = [...transactionsList.value]
        
        // 刷新统计信息
        loadTransactionStats()
        
      }
    }
  })
  wsUnsubscribes.push(unsubTx)
}

// 监听模态框状态变化
watch(showCreateModal, (newVal) => {
  if (!newVal) {
    // 模态框关闭时重置编辑状态
    isEditMode.value = false
    selectedTransaction.value = null
  }
})

onMounted(async () => {
  // 先加载Gas费率数据，确保页面打开时立即显示费率信息
  await loadGasRates()
  
  // 然后加载其他数据
  loadTransactions()
  loadTransactionStats()
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
  wsUnsubscribes.forEach(unsub => { try { unsub() } catch {}
  })
  wsUnsubscribes.length = 0
  
  // 移除窗口大小变化监听
  window.removeEventListener('resize', updateChart)
})
</script>

