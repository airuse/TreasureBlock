<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">系统合约地址管理</h1>
      <div class="flex items-center space-x-4">
        <div class="text-sm text-gray-500">
          共 {{ totalAddresses.toLocaleString() }} 个系统地址
        </div>
        <button 
          v-if="isAdmin"
          @click="showAddAddressModal = true"
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
          </svg>
          添加合约地址
        </button>
      </div>
    </div>

    <!-- 搜索和筛选 -->
    <div class="card">
      <div class="flex flex-col sm:flex-row gap-4">
        <div class="flex-1">
          <label class="block text-sm font-medium text-gray-700 mb-2">搜索合约地址</label>
          <div class="relative">
            <input 
              v-model="searchQuery" 
              type="text" 
              placeholder="输入合约地址哈希或名称..."
              class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <svg class="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
              </svg>
            </div>
          </div>
        </div>
        <div class="sm:w-48">
          <label class="block text-sm font-medium text-gray-700 mb-2">合约类型</label>
          <select 
            v-model="typeFilter" 
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">全部类型</option>
            <option value="erc20">ERC-20 代币</option>
            <option value="erc721">ERC-721 NFT</option>
            <option value="erc1155">ERC-1155 多代币</option>
            <option value="defi">DeFi 协议</option>
            <option value="dex">DEX 交易所</option>
            <option value="lending">借贷协议</option>
            <option value="other">其他合约</option>
          </select>
        </div>
        <div class="sm:w-48">
          <label class="block text-sm font-medium text-gray-700 mb-2">状态</label>
          <select 
            v-model="statusFilter" 
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">全部状态</option>
            <option value="active">活跃</option>
            <option value="inactive">非活跃</option>
            <option value="paused">暂停</option>
          </select>
        </div>
        <div class="sm:w-48">
          <label class="block text-sm font-medium text-gray-700 mb-2">每页显示</label>
          <select 
            v-model="pageSize" 
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="10">10</option>
            <option value="25">25</option>
            <option value="50">50</option>
            <option value="100">100</option>
          </select>
        </div>
      </div>
    </div>

    <!-- 编辑合约模态框 -->
    <div v-if="showEditModal" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div class="relative top-16 mx-auto p-5 border w-11/12 max-w-5xl shadow-xl rounded-xl bg-white">
        <div class="flex justify-between items-center mb-4 pb-3 border-b">
          <h3 class="text-lg font-semibold text-gray-900">编辑合约信息</h3>
          <button 
            @click="closeEditModal"
            class="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
            </svg>
          </button>
        </div>
        
        <form @submit.prevent="saveEdit" class="max-h-[65vh] overflow-y-auto pr-2">
          <!-- 基本信息 -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3 mb-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">合约地址</label>
              <input 
                v-model="editingAddress.hash" 
                type="text" 
                readonly
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded bg-gray-50 text-gray-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">合约名称</label>
              <input 
                v-model="editingAddress.name" 
                type="text" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">合约符号</label>
              <input 
                v-model="editingAddress.symbol" 
                type="text" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">合约类型</label>
              <select 
                v-model="editingAddress.type" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="erc20">ERC-20 代币</option>
                <option value="erc721">ERC-721 NFT</option>
                <option value="erc1155">ERC-1155 多代币</option>
                <option value="defi">DeFi 协议</option>
                <option value="dex">DEX 交易所</option>
                <option value="lending">借贷协议</option>
                <option value="other">其他合约</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">精度</label>
              <input 
                v-model="editingAddress.decimals" 
                type="number" 
                min="0" 
                max="18"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">状态</label>
              <select 
                v-model="editingAddress.status" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="active">活跃</option>
                <option value="inactive">非活跃</option>
                <option value="paused">暂停</option>
              </select>
            </div>
          </div>

          <!-- 合约详细信息 -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">接口 (Interfaces)</label>
              <div class="space-y-2">
                <div class="flex space-x-2">
                  <input 
                    v-model="newInterface" 
                    type="text" 
                    placeholder="输入接口名称"
                    class="flex-1 px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                    @keyup.enter="addInterface"
                  />
                  <button 
                    @click="addInterface"
                    type="button"
                    class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  >
                    +
                  </button>
                </div>
                <div class="flex flex-wrap gap-1">
                  <span 
                    v-for="(item, index) in editingAddress.interfacesList" 
                    :key="index"
                    class="inline-flex items-center px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded"
                  >
                    {{ item }}
                    <button 
                      @click="removeInterface(index)"
                      type="button"
                      class="ml-1 text-blue-600 hover:text-blue-800"
                    >
                      ×
                    </button>
                  </span>
                </div>
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">方法 (Methods)</label>
              <div class="space-y-2">
                <div class="flex space-x-2">
                  <input 
                    v-model="newMethod" 
                    type="text" 
                    placeholder="输入方法名称"
                    class="flex-1 px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                    @keyup.enter="addMethod"
                  />
                  <button 
                    @click="addMethod"
                    type="button"
                    class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  >
                    +
                  </button>
                </div>
                <div class="flex flex-wrap gap-1">
                  <span 
                    v-for="(item, index) in editingAddress.methodsList" 
                    :key="index"
                    class="inline-flex items-center px-2 py-1 text-xs bg-green-100 text-green-800 rounded"
                  >
                    {{ item }}
                    <button 
                      @click="removeMethod(index)"
                      type="button"
                      class="ml-1 text-green-600 hover:text-green-800"
                    >
                      ×
                    </button>
                  </span>
                </div>
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">事件 (Events)</label>
              <div class="space-y-2">
                <div class="flex space-x-2">
                  <input 
                    v-model="newEvent" 
                    type="text" 
                    placeholder="输入事件名称"
                    class="flex-1 px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                    @keyup.enter="addEvent"
                  />
                  <button 
                    @click="addEvent"
                    type="button"
                    class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  >
                    +
                  </button>
                </div>
                <div class="flex flex-wrap gap-1">
                  <span 
                    v-for="(item, index) in editingAddress.eventsList" 
                    :key="index"
                    class="inline-flex items-center px-2 py-1 text-xs bg-purple-100 text-purple-800 rounded"
                  >
                    {{ item }}
                    <button 
                      @click="removeEvent(index)"
                      type="button"
                      class="ml-1 text-purple-600 hover:text-purple-800"
                    >
                      ×
                    </button>
                  </span>
                </div>
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">元数据 (Metadata)</label>
              <div class="space-y-2">
                <div class="flex space-x-2">
                  <input 
                    v-model="newMetadataKey" 
                    type="text" 
                    placeholder="键名"
                    class="flex-1 px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                  />
                  <input 
                    v-model="newMetadataValue" 
                    type="text" 
                    placeholder="值"
                    class="flex-1 px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                    @keyup.enter="addMetadata"
                  />
                  <button 
                    @click="addMetadata"
                    type="button"
                    class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  >
                    +
                  </button>
                </div>
                <div class="flex flex-wrap gap-1">
                  <span 
                    v-for="(key, index) in Object.keys(editingAddress.metadataObj || {})" 
                    :key="index"
                    class="inline-flex items-center px-2 py-1 text-xs bg-yellow-100 text-yellow-800 rounded"
                  >
                    {{ key }}: {{ editingAddress.metadataObj?.[key] }}
                    <button 
                      @click="removeMetadata(key)"
                      type="button"
                      class="ml-1 text-yellow-600 hover:text-yellow-800"
                    >
                      ×
                    </button>
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- 其他信息 -->
          <div class="grid grid-cols-1 md:grid-cols-4 gap-3 mb-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">总供应量</label>
              <input 
                v-model="editingAddress.totalSupply" 
                type="text" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">是否ERC-20</label>
              <select 
                v-model="editingAddress.isErc20" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option :value="true">是</option>
                <option :value="false">否</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">是否已验证</label>
              <select 
                v-model="editingAddress.verified" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option :value="true">已验证</option>
                <option :value="false">未验证</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">创建区块</label>
              <input 
                v-model="editingAddress.creationBlock" 
                type="number" 
                placeholder="12345678"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </div>

          <!-- 创建信息 -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">创建者地址</label>
              <input 
                v-model="editingAddress.creator" 
                type="text" 
                placeholder="0x..."
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">创建交易哈希</label>
              <input 
                v-model="editingAddress.creationTx" 
                type="text" 
                placeholder="0x..."
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </div>

          <!-- 合约Logo和描述 -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">合约Logo</label>
              <div class="flex items-center space-x-2">
                <img 
                  v-if="editingAddress.contractLogo" 
                  :src="editingAddress.contractLogo" 
                  alt="合约Logo" 
                  class="w-10 h-10 rounded object-cover border"
                />
                <input 
                  type="file" 
                  @change="handleLogoUpload" 
                  accept="image/*"
                  class="block w-full text-xs text-gray-500 file:mr-2 file:py-1 file:px-2 file:rounded file:border-0 file:text-xs file:font-medium file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
                />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">描述</label>
              <textarea 
                v-model="editingAddress.description" 
                rows="2"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              ></textarea>
            </div>
          </div>
        </form>

        <!-- 底部按钮 -->
        <div class="flex justify-end space-x-3 pt-4 border-t">
          <button 
            type="button"
            @click="closeEditModal"
            class="px-4 py-2 text-xs font-medium text-gray-700 bg-gray-100 border border-gray-300 rounded hover:bg-gray-200 transition-colors"
          >
            取消
          </button>
          <button 
            @click="saveEdit"
            class="px-4 py-2 text-xs font-medium text-white bg-blue-600 border border-transparent rounded hover:bg-blue-700 transition-colors"
          >
            保存
          </button>
        </div>
      </div>
    </div>

    <!-- 维护币种信息模态框 -->
    <div v-if="showCoinConfigModal" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div class="relative top-16 mx-auto p-5 border w-11/12 max-w-4xl shadow-xl rounded-xl bg-white">
        <div class="flex justify-between items-center mb-4 pb-3 border-b">
          <h3 class="text-lg font-semibold text-gray-900">维护币种信息</h3>
          <div class="flex items-center space-x-2">
            <!-- 关闭按钮 -->
            <button 
              @click="closeCoinConfigModal"
              class="text-gray-400 hover:text-gray-600 transition-colors"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
              </svg>
            </button>
          </div>
        </div>
        
        <form @submit.prevent="saveCoinConfig" class="max-h-[65vh] overflow-y-auto pr-2">
          <!-- 币种基本信息 -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3 mb-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">合约地址</label>
              <input 
                v-model="coinConfigData.contract_address" 
                type="text" 
                readonly
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded bg-gray-50 text-gray-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">币种名称 *</label>
              <input 
                v-model="coinConfigData.name" 
                type="text" 
                required
                placeholder="输入币种名称"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">币种符号 *</label>
              <input 
                v-model="coinConfigData.symbol" 
                type="text" 
                required
                placeholder="输入币种符号"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">币种类型</label>
              <select 
                v-model="coinConfigData.coin_type" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="0">原生币</option>
                <option value="1">ERC-20 代币</option>
                <option value="2">ERC-223 代币</option>
                <option value="3">ERC-721 NFT</option>
                <option value="4">ERC-1155 多代币</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">精度</label>
              <input 
                v-model="coinConfigData.precision" 
                type="number" 
                min="0" 
                max="18"
                placeholder="18"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">精度别名</label>
              <input 
                v-model="coinConfigData.decimals" 
                type="number" 
                min="0" 
                max="18"
                placeholder="18"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">市值排名</label>
              <input 
                v-model="coinConfigData.market_cap_rank" 
                type="number" 
                min="0"
                placeholder="0"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">是否为稳定币</label>
              <select 
                v-model="coinConfigData.is_stablecoin" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option :value="false">否</option>
                <option :value="true">是</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">状态</label>
              <select 
                v-model="coinConfigData.status" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="1">启用</option>
                <option value="0">禁用</option>
              </select>
            </div>
          </div>

          <!-- 其他信息 -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Logo URL</label>
              <input 
                v-model="coinConfigData.logo_url" 
                type="text" 
                placeholder="输入Logo URL"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">是否已验证</label>
              <select 
                v-model="coinConfigData.is_verified" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option :value="true">已验证</option>
                <option :value="false">未验证</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">官方网站</label>
              <input 
                v-model="coinConfigData.website_url" 
                type="text" 
                placeholder="输入官方网站URL"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">区块浏览器地址</label>
              <input 
                v-model="coinConfigData.explorer_url" 
                type="text" 
                placeholder="输入区块浏览器地址"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div class="md:col-span-2">
              <label class="block text-xs font-medium text-gray-700 mb-1">币种描述</label>
              <textarea 
                v-model="coinConfigData.description" 
                rows="2"
                placeholder="输入币种描述"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              ></textarea>
            </div>
          </div>

          <!-- 解析配置 -->
          <div class="mb-4">
            <h4 class="text-sm font-medium text-gray-700 mb-2">解析配置</h4>
            <div class="space-y-3">
              <div 
                v-for="(config, index) in coinConfigData.parser_configs" 
                :key="index"
                class="border border-gray-200 rounded p-3"
              >
                <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-2">
                  <div>
                    <label class="block text-xs font-medium text-gray-700 mb-1">函数名称</label>
                    <input 
                      v-model="config.function_name" 
                      type="text" 
                      placeholder="transfer"
                      class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                    />
                  </div>
                  <div>
                    <label class="block text-xs font-medium text-gray-700 mb-1">函数签名</label>
                    <input 
                      v-model="config.function_signature" 
                      type="text" 
                      placeholder="0xa9059cbb"
                      class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                    />
                  </div>
                </div>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-2">
                  <div>
                    <label class="block text-xs font-medium text-gray-700 mb-1">函数描述</label>
                    <input 
                      v-model="config.function_description" 
                      type="text" 
                      placeholder="转账函数"
                      class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                    />
                  </div>
                  <div>
                    <label class="block text-xs font-medium text-gray-700 mb-1">显示格式</label>
                    <input 
                      v-model="config.display_format" 
                      type="text" 
                      placeholder="转账 {amount} {symbol} 到 {to}"
                      class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                    />
                  </div>
                </div>
                
                <!-- 参数配置 -->
                <div class="mb-2">
                  <label class="block text-xs font-medium text-gray-700 mb-1">参数配置</label>
                  <div class="space-y-2">
                    <div 
                      v-for="(param, paramIndex) in config.param_config" 
                      :key="paramIndex"
                      class="grid grid-cols-1 md:grid-cols-5 gap-2 p-2 bg-gray-50 rounded"
                    >
                      <input 
                        v-model="param.name" 
                        type="text" 
                        placeholder="参数名"
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                      <select 
                        v-model="param.type" 
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      >
                        <option value="address">地址</option>
                        <option value="uint256">大整数</option>
                        <option value="bytes">字节</option>
                        <option value="string">字符串</option>
                      </select>
                      <input 
                        v-model="param.offset" 
                        type="number" 
                        placeholder="偏移量"
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                      <input 
                        v-model="param.length" 
                        type="number" 
                        placeholder="长度"
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                      <div class="flex items-center space-x-2">
                        <input 
                          v-model="param.description" 
                          type="text" 
                          placeholder="参数描述"
                          class="flex-1 px-2 py-1 text-xs border border-gray-300 rounded"
                        />
                        <button 
                          @click="removeParamConfig(index, paramIndex)"
                          type="button"
                          class="text-red-600 hover:text-red-800 text-xs px-2 py-1"
                          title="删除参数"
                        >
                          ×
                        </button>
                      </div>
                    </div>
                    <button 
                      @click="addParamConfig(index)"
                      type="button"
                      class="text-xs text-blue-600 hover:text-blue-800"
                    >
                      + 添加参数
                    </button>
                  </div>
                </div>
                
                <!-- 解析规则 -->
                <div class="mb-2">
                  <label class="block text-xs font-medium text-gray-700 mb-1">解析规则</label>
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                    <input 
                      v-model="config.parser_rules.extract_to_address" 
                      type="text" 
                      placeholder="提取收款地址规则"
                      class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                    />
                    <input 
                      v-model="config.parser_rules.extract_amount" 
                      type="text" 
                      placeholder="提取金额规则"
                      class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                    />
                    <input 
                      v-model="config.parser_rules.amount_unit" 
                      type="text" 
                      placeholder="金额单位"
                      class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                    />
                    <input 
                      v-model="config.parser_rules.extract_data" 
                      type="text" 
                      placeholder="提取其他数据规则"
                      class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                    />
                  </div>
                </div>

                <!-- 日志解析配置 -->
                <div class="mb-2 border-t pt-2">
                  <label class="block text-xs font-medium text-gray-700 mb-1">日志解析配置</label>
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-2">
                    <div>
                      <label class="block text-xs font-medium text-gray-500 mb-1">日志解析类型</label>
                      <select 
                        v-model="config.logs_parser_type" 
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      >
                        <option value="input_data">输入数据</option>
                        <option value="event_log">事件日志</option>
                        <option value="both">两者</option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-gray-500 mb-1">事件签名</label>
                      <input 
                        v-model="config.event_signature" 
                        type="text" 
                        placeholder="0x..."
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-gray-500 mb-1">事件名称</label>
                      <input 
                        v-model="config.event_name" 
                        type="text" 
                        placeholder="Transfer"
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-gray-500 mb-1">事件描述</label>
                      <input 
                        v-model="config.event_description" 
                        type="text" 
                        placeholder="转账事件"
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                    </div>
                    <div>
                      <label class="block text-xs font-medium text-gray-500 mb-1">日志显示格式</label>
                      <input 
                        v-model="config.logs_display_format" 
                        type="text" 
                        placeholder="转账 {amount} {symbol} 从 {from} 到 {to}"
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                    </div>
                  </div>

                  <!-- 日志参数配置 -->
                  <div class="mb-2">
                    <label class="block text-xs font-medium text-gray-500 mb-1">日志参数配置</label>
                    <div class="space-y-2">
                      <div 
                        v-for="(logParam, logParamIndex) in config.logs_param_config" 
                        :key="logParamIndex"
                        class="grid grid-cols-1 md:grid-cols-6 gap-2 p-2 bg-blue-50 rounded"
                      >
                        <input 
                          v-model="logParam.name" 
                          type="text" 
                          placeholder="参数名"
                          class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                        />
                        <select 
                          v-model="logParam.type" 
                          class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                        >
                          <option value="address">地址</option>
                          <option value="uint256">大整数</option>
                          <option value="bytes">字节</option>
                          <option value="string">字符串</option>
                        </select>
                        <input 
                          v-model="logParam.topic_index" 
                          type="number" 
                          placeholder="Topic索引"
                          class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                        />
                        <input 
                          v-model="logParam.data_index" 
                          type="number" 
                          placeholder="Data索引"
                          class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                        />
                        <input 
                          v-model="logParam.description" 
                          type="text" 
                          placeholder="参数描述"
                          class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                        />
                        <button 
                          @click="removeLogsParamConfig(index, logParamIndex)"
                          type="button"
                          class="text-red-600 hover:text-red-800 text-xs px-2 py-1"
                          title="删除日志参数"
                        >
                          ×
                        </button>
                      </div>
                      <button 
                        @click="addLogsParamConfig(index)"
                        type="button"
                        class="text-xs text-blue-600 hover:text-blue-800"
                      >
                        + 添加日志参数
                      </button>
                    </div>
                  </div>

                  <!-- 日志解析规则 -->
                  <div class="mb-2">
                    <label class="block text-xs font-medium text-gray-500 mb-1">日志解析规则</label>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                      <input 
                        v-model="config.logs_parser_rules.extract_from_address" 
                        type="text" 
                        placeholder="提取发送地址规则"
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                      <input 
                        v-model="config.logs_parser_rules.extract_to_address" 
                        type="text" 
                        placeholder="提取收款地址规则"
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                      <input 
                        v-model="config.logs_parser_rules.extract_amount" 
                        type="text" 
                        placeholder="提取金额规则"
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                      <input 
                        v-model="config.logs_parser_rules.amount_unit" 
                        type="text" 
                        placeholder="金额单位"
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                      <input 
                        v-model="config.logs_parser_rules.extract_owner_address" 
                        type="text" 
                        placeholder="提取所有者地址规则"
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                      <input 
                        v-model="config.logs_parser_rules.extract_spender_address" 
                        type="text" 
                        placeholder="提取授权地址规则"
                        class="block w-full px-2 py-1 text-xs border border-gray-300 rounded"
                      />
                    </div>
                  </div>
                </div>

                <div class="flex items-center space-x-2">
                  <label class="flex items-center">
                    <input 
                      v-model="config.is_active" 
                      type="checkbox" 
                      class="mr-2"
                    />
                    <span class="text-xs text-gray-700">启用</span>
                  </label>
                  <button 
                    @click="removeParserConfig(index)"
                    type="button"
                    class="ml-auto text-red-600 hover:text-red-800 text-xs"
                  >
                    删除
                  </button>
                </div>
              </div>
              <button 
                @click="addParserConfig"
                type="button"
                class="w-full px-3 py-2 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                + 添加解析配置
              </button>
            </div>
          </div>
        </form>

        <!-- 底部按钮 -->
        <div class="flex justify-end space-x-3 pt-4 border-t">
          <button 
            type="button"
            @click="closeCoinConfigModal"
            class="px-4 py-2 text-xs font-medium text-gray-700 bg-gray-100 border border-gray-300 rounded hover:bg-gray-200 transition-colors"
          >
            取消
          </button>
          <button 
            @click="saveCoinConfig"
            class="px-4 py-2 text-xs font-medium text-white bg-blue-600 border border-transparent rounded hover:bg-blue-700 transition-colors"
          >
            保存
          </button>
        </div>
      </div>
    </div>

    <!-- 添加合约地址模态框 -->
    <div v-if="showAddAddressModal" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div class="relative top-16 mx-auto p-5 border w-11/12 max-w-5xl shadow-xl rounded-xl bg-white">
        <div class="flex justify-between items-center mb-4 pb-3 border-b">
          <h3 class="text-lg font-semibold text-gray-900">添加新合约地址</h3>
          <button 
            @click="closeAddModal"
            class="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
            </svg>
          </button>
        </div>
        
        <form @submit.prevent="saveAdd" class="max-h-[65vh] overflow-y-auto pr-2">
          <!-- 基本信息 -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3 mb-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">合约地址 *</label>
              <input 
                v-model="newAddress.hash" 
                type="text" 
                required
                placeholder="0x..."
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">合约名称</label>
              <input 
                v-model="newAddress.name" 
                type="text" 
                placeholder="输入合约名称"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">合约符号</label>
              <input 
                v-model="newAddress.symbol" 
                type="text" 
                placeholder="输入合约符号"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">合约类型</label>
              <select 
                v-model="newAddress.type" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="erc20">ERC-20 代币</option>
                <option value="erc721">ERC-721 NFT</option>
                <option value="erc1155">ERC-1155 多代币</option>
                <option value="defi">DeFi 协议</option>
                <option value="dex">DEX 交易所</option>
                <option value="lending">借贷协议</option>
                <option value="other">其他合约</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">精度</label>
              <input 
                v-model="newAddress.decimals" 
                type="number" 
                min="0" 
                max="18"
                placeholder="18"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">状态</label>
              <select 
                v-model="newAddress.status" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="active">活跃</option>
                <option value="inactive">非活跃</option>
                <option value="paused">暂停</option>
              </select>
            </div>
          </div>

          <!-- 合约详细信息 -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">接口 (Interfaces)</label>
              <div class="space-y-2">
                <div class="flex space-x-2">
                  <input 
                    v-model="newInterface" 
                    type="text" 
                    placeholder="输入接口名称"
                    class="flex-1 px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                    @keyup.enter="addNewInterface"
                  />
                  <button 
                    @click="addNewInterface"
                    type="button"
                    class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  >
                    +
                  </button>
                </div>
                <div class="flex flex-wrap gap-1">
                  <span 
                    v-for="(item, index) in newAddress.interfacesList" 
                    :key="index"
                    class="inline-flex items-center px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded"
                  >
                    {{ item }}
                    <button 
                      @click="removeNewInterface(index)"
                      type="button"
                      class="ml-1 text-blue-600 hover:text-blue-800"
                    >
                      ×
                    </button>
                  </span>
                </div>
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">方法 (Methods)</label>
              <div class="space-y-2">
                <div class="flex space-x-2">
                  <input 
                    v-model="newMethod" 
                    type="text" 
                    placeholder="输入方法名称"
                    class="flex-1 px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                    @keyup.enter="addNewMethod"
                  />
                  <button 
                    @click="addNewMethod"
                    type="button"
                    class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  >
                    +
                  </button>
                </div>
                <div class="flex flex-wrap gap-1">
                  <span 
                    v-for="(item, index) in newAddress.methodsList" 
                    :key="index"
                    class="inline-flex items-center px-2 py-1 text-xs bg-green-100 text-green-800 rounded"
                  >
                    {{ item }}
                    <button 
                      @click="removeNewMethod(index)"
                      type="button"
                      class="ml-1 text-green-600 hover:text-green-800"
                    >
                      ×
                    </button>
                  </span>
                </div>
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">事件 (Events)</label>
              <div class="space-y-2">
                <div class="flex space-x-2">
                  <input 
                    v-model="newEvent" 
                    type="text" 
                    placeholder="输入事件名称"
                    class="flex-1 px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                    @keyup.enter="addNewEvent"
                  />
                  <button 
                    @click="addNewEvent"
                    type="button"
                    class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  >
                    +
                  </button>
                </div>
                <div class="flex flex-wrap gap-1">
                  <span 
                    v-for="(item, index) in newAddress.eventsList" 
                    :key="index"
                    class="inline-flex items-center px-2 py-1 text-xs bg-purple-100 text-purple-800 rounded"
                  >
                    {{ item }}
                    <button 
                      @click="removeNewEvent(index)"
                      type="button"
                      class="ml-1 text-purple-600 hover:text-purple-800"
                    >
                      ×
                    </button>
                  </span>
                </div>
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">元数据 (Metadata)</label>
              <div class="space-y-2">
                <div class="flex space-x-2">
                  <input 
                    v-model="newMetadataKey" 
                    type="text" 
                    placeholder="键名"
                    class="flex-1 px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                  />
                  <input 
                    v-model="newMetadataValue" 
                    type="text" 
                    placeholder="值"
                    class="flex-1 px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                    @keyup.enter="addNewMetadata"
                  />
                  <button 
                    @click="addNewMetadata"
                    type="button"
                    class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  >
                    +
                  </button>
                </div>
                <div class="flex flex-wrap gap-1">
                  <span 
                    v-for="(key, index) in Object.keys(newAddress.metadataObj || {})" 
                    :key="index"
                    class="inline-flex items-center px-2 py-1 text-xs bg-yellow-100 text-yellow-800 rounded"
                  >
                    {{ key }}: {{ newAddress.metadataObj?.[key] }}
                    <button 
                      @click="removeNewMetadata(key)"
                      type="button"
                      class="ml-1 text-yellow-600 hover:text-yellow-800"
                    >
                      ×
                    </button>
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- 其他信息 -->
          <div class="grid grid-cols-1 md:grid-cols-4 gap-3 mb-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">总供应量</label>
              <input 
                v-model="newAddress.totalSupply" 
                type="text" 
                placeholder="1000000000000000000000000000000"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">是否ERC-20</label>
              <select 
                v-model="newAddress.isErc20" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option :value="true">是</option>
                <option :value="false">否</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">是否已验证</label>
              <select 
                v-model="newAddress.verified" 
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-blue-500"
              >
                <option :value="true">已验证</option>
                <option :value="false">未验证</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">创建区块</label>
              <input 
                v-model="newAddress.creationBlock" 
                type="number" 
                placeholder="12345678"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </div>

          <!-- 创建信息 -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">创建者地址</label>
              <input 
                v-model="newAddress.creator" 
                type="text" 
                placeholder="0x..."
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">创建交易哈希</label>
              <input 
                v-model="newAddress.creationTx" 
                type="text" 
                placeholder="0x..."
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </div>

          <!-- 合约Logo和描述 -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">合约Logo</label>
              <div class="flex items-center space-x-2">
                <img 
                  v-if="newAddress.contractLogo" 
                  :src="newAddress.contractLogo" 
                  alt="合约Logo" 
                  class="w-10 h-10 rounded object-cover border"
                />
                <input 
                  type="file" 
                  @change="handleNewLogoUpload" 
                  accept="image/*"
                  class="block w-full text-xs text-gray-500 file:mr-2 file:py-1 file:px-2 file:rounded file:border-0 file:text-xs file:font-medium file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
                />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">描述</label>
              <textarea 
                v-model="newAddress.description" 
                rows="2"
                placeholder="输入合约描述"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              ></textarea>
            </div>
          </div>
        </form>

        <!-- 底部按钮 -->
        <div class="flex justify-end space-x-3 pt-4 border-t">
          <button 
            type="button"
            @click="closeAddModal"
            class="px-4 py-2 text-xs font-medium text-white bg-gray-100 border border-gray-300 rounded hover:bg-gray-200 transition-colors"
          >
            取消
          </button>
          <button 
            @click="saveAdd"
            class="px-4 py-2 text-xs font-medium text-white bg-blue-600 border border-transparent rounded hover:bg-blue-700 transition-colors"
          >
            添加
          </button>
        </div>
      </div>
    </div>

    <!-- 地址列表 -->
    <div class="card">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">图标</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">合约地址</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">合约类型</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">合约名称</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">合约符号</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">精度</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="address in addresses" :key="address.hash" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap">
                <div v-if="address.contractLogo" class="flex items-center justify-center">
                  <div class="w-10 h-10 rounded-full overflow-hidden border-2 border-gray-200">
                    <img 
                      :src="address.contractLogo" 
                      :alt="address.name || '合约图标'"
                      class="w-full h-full object-cover"
                      @error="handleImageError"
                    />
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center space-x-2">
                  <router-link :to="`/eth/addresses/${address.hash}`" class="text-blue-600 hover:text-blue-700 font-mono text-sm">
                    {{ address.hash }}
                  </router-link>
                  <button 
                    @click="copyToClipboard(address.hash)"
                    class="text-gray-400 hover:text-gray-600"
                    title="复制地址"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 002 2v8a2 2 0 002 2z"></path>
                    </svg>
                  </button>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getTypeClass(address.type)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                  {{ getTypeText(address.type) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                <div class="flex items-center space-x-2">
                  <span class="font-medium">{{ address.name || '未命名合约' }}</span>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                <span class="text-gray-600">{{ address.symbol || '-' }}</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                <span class="text-gray-600">{{ address.decimals !== undefined ? address.decimals : '-' }}</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getStatusClass(address.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                  {{ getStatusText(address.status) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                <div class="flex items-center space-x-2">
                  <button 
                    v-if="isAdmin"
                    @click="editAddress(address)"
                    class="text-blue-600 hover:text-blue-800 text-xs"
                  >
                    编辑
                  </button>
                  <button 
                    v-if="isAdmin && address.isErc20"
                    @click="maintainCoinConfig(address)"
                    class="text-green-600 hover:text-green-800 text-xs"
                  >
                    维护币种信息
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页 -->
      <div class="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
        <div class="flex-1 flex justify-between sm:hidden">
          <button 
            @click="previousPage" 
            :disabled="currentPage === 1"
            class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            上一页
          </button>
          <button 
            @click="nextPage" 
            :disabled="currentPage >= totalPages"
            class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            下一页
          </button>
        </div>
        <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
          <div>
            <p class="text-sm text-gray-700">
              显示第 <span class="font-medium">{{ (currentPage - 1) * pageSize + 1 }}</span> 到 
              <span class="font-medium">{{ Math.min(currentPage * pageSize, totalAddresses) }}</span> 条，
              共 <span class="font-medium">{{ totalAddresses.toLocaleString() }}</span> 条记录
            </p>
          </div>
          <div>
            <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
              <button 
                @click="previousPage" 
                :disabled="currentPage === 1"
                class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <svg class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
                </svg>
              </button>
              
              <button 
                v-for="page in visiblePages" 
                :key="page"
                @click="goToPage(page)"
                :class="[
                  page === currentPage 
                    ? 'z-10 bg-blue-50 border-blue-500 text-blue-600' 
                    : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50',
                  'relative inline-flex items-center px-4 py-2 border text-sm font-medium'
                ]"
              >
                {{ page }}
              </button>
              
              <button 
                @click="nextPage" 
                :disabled="currentPage >= totalPages"
                class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <svg class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
              </button>
            </nav>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { getCoinConfigMaintenance, createCoinConfig } from '@/api/coinconfig'
import { batchSaveParserConfigs, type ParserConfig } from '@/api/parser-configs'
import { showSuccess, showError } from '@/composables/useToast'
import request from '@/api/request'

// 定义实际后端响应的类型（因为与标准PaginatedResponse不同）
interface ContractsResponse {
  success: boolean
  message?: string
  data: any[]
  count: number // 实际后端返回的是 count 而不是 pagination.total
}

// 响应式数据
const searchQuery = ref('')
const typeFilter = ref('')
const statusFilter = ref('')
const pageSize = ref(25)
const currentPage = ref(1)
const totalAddresses = ref(0)
const showAddAddressModal = ref(false)
const showEditModal = ref(false)
const showCoinConfigModal = ref(false)

// 新增输入变量
const newInterface = ref('')
const newMethod = ref('')
const newEvent = ref('')
const newMetadataKey = ref('')
const newMetadataValue = ref('')

const editingAddress = ref<Address>({
  hash: '',
  type: '',
  name: '',
  symbol: '',
  status: 'active',
  transactionCount: 0,
  lastActivity: 0,
  description: '',
  decimals: 0,
  totalSupply: '',
  contractLogo: '',
  interfaces: '',
  methods: '',
  events: '',
  metadata: '',
  interfacesList: [],
  methodsList: [],
  eventsList: [],
  metadataObj: {},
  isErc20: false,
  verified: false,
  creator: '',
  creationTx: '',
  creationBlock: 0
})

// 新增合约地址
const newAddress = ref<Address>({
  hash: '',
  type: 'erc20',
  name: '',
  symbol: '',
  status: 'active',
  transactionCount: 0,
  lastActivity: 0,
  description: '',
  decimals: 18,
  totalSupply: '',
  contractLogo: '',
  interfaces: '',
  methods: '',
  events: '',
  metadata: '',
  interfacesList: [],
  methodsList: [],
  eventsList: [],
  metadataObj: {},
  isErc20: true,
  verified: false,
  creator: '',
  creationTx: '',
  creationBlock: 0
})

// 币种配置数据
const coinConfigData = ref({
  contract_address: '',
  name: '',
  symbol: '',
  coin_type: 1,
  precision: 18,
  decimals: 18,
  status: 1,
  logo_url: '',
  is_verified: false,
  market_cap_rank: 0,
  is_stablecoin: false,
  website_url: '',
  explorer_url: '',
  description: '',
  parser_configs: [] as any[]
})

// 认证store
const authStore = useAuthStore()

// 导入类型
import type { Contract } from '@/types'

// 定义系统合约地址类型（兼容原有接口）
interface Address {
  hash: string
  type: string
  name: string | null
  symbol: string | null
  status: string
  transactionCount: number
  lastActivity: number
  description: string | null
  decimals?: number
  totalSupply?: string
  contractLogo?: string
  // 新增合约详细信息字段
  interfaces?: string | string[]
  methods?: string | string[]
  events?: string | string[]
  metadata?: string | Record<string, string>
  // 新增列表形式字段，用于UI显示
  interfacesList?: string[]
  methodsList?: string[]
  eventsList?: string[]
  metadataObj?: Record<string, string>
  isErc20?: boolean
  verified?: boolean
  creator?: string
  creationTx?: string
  creationBlock?: number
}

const addresses = ref<Address[]>([])

// 计算属性
const totalPages = computed(() => Math.ceil(totalAddresses.value / pageSize.value))

const visiblePages = computed(() => {
  const pages = []
  const start = Math.max(1, currentPage.value - 2)
  const end = Math.min(totalPages.value, currentPage.value + 2)
  
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  return pages
})

// 格式化函数
const formatAddress = (address: string) => {
  return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`
}

const formatTimestamp = (timestamp: number) => {
  return new Date(timestamp * 1000).toLocaleString()
}



const getTypeClass = (type: string) => {
  switch (type) {
    case 'contract':
      return 'bg-purple-100 text-purple-800'
    case 'wallet':
      return 'bg-blue-100 text-blue-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

const getTypeText = (type: string) => {
  switch (type) {
    case 'erc20':
      return 'ERC-20 代币'
    case 'erc721':
      return 'ERC-721 NFT'
    case 'erc1155':
      return 'ERC-1155 多代币'
    case 'defi':
      return 'DeFi 协议'
    case 'dex':
      return 'DEX 交易所'
    case 'lending':
      return '借贷协议'
    case 'other':
      return '其他合约'
    default:
      return '未知'
  }
}

const getStatusClass = (status: string) => {
  switch (status) {
    case 'active':
      return 'bg-green-100 text-green-800'
    case 'inactive':
      return 'bg-gray-100 text-gray-800'
    case 'paused':
      return 'bg-yellow-100 text-yellow-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'active':
      return '活跃'
    case 'inactive':
      return '非活跃'
    case 'paused':
      return '暂停'
    default:
      return '未知'
  }
}

// 复制到剪贴板
const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text).then(() => {
    showSuccess('地址已复制到剪贴板')
    console.log('地址已复制到剪贴板:', text)
  }).catch(err => {
    console.error('复制失败:', err)
    showError('复制失败，请手动复制')
  })
}

// 编辑地址 - 重新查询最新数据
const editAddress = async (address: Address) => {
  console.log('编辑地址:', address)
  
  try {
    // 显示加载状态
    showEditModal.value = true
    
    // 根据登录状态调用不同的API
    let response
    if (authStore.isAuthenticated) {
      // 已登录用户：调用 /v1/ 下的API
      const { contracts } = await import('@/api')
      response = await contracts.getContractByAddress(address.hash)
    } else {
      // 未登录用户：调用 /no-auth/ 下的API
      const { noAuth } = await import('@/api')
      response = await noAuth.getContracts({ 
        chainName: 'eth',
        search: address.hash,
        page: 1,
        page_size: 1
      })
      // 转换响应格式以匹配getContractByAddress的格式
      if (response.success && response.data && response.data.length > 0) {
        response = {
          success: true,
          data: response.data[0]
        }
      }
    }
    
          if (response.success && response.data) {
        // 确保latestContract是单个合约对象
        const latestContract = Array.isArray(response.data) ? response.data[0] : response.data
        console.log('获取到最新合约数据:', latestContract)
      
      // 解析最新数据
      let interfacesList: string[] = []
      let methodsList: string[] = []
      let eventsList: string[] = []
      let metadataObj: Record<string, string> = {}
      
      try {
        if (latestContract.interfaces) {
          // 如果已经是数组格式，直接使用
          if (Array.isArray(latestContract.interfaces)) {
            interfacesList = latestContract.interfaces
          } else if (typeof latestContract.interfaces === 'string') {
            // 尝试解析JSON字符串
            interfacesList = JSON.parse(latestContract.interfaces)
          }
        }
      } catch (e) {
        console.warn('解析interfaces失败:', e, '原始数据:', latestContract.interfaces)
        interfacesList = []
      }
      
      try {
        if (latestContract.methods) {
          if (Array.isArray(latestContract.methods)) {
            methodsList = latestContract.methods
          } else if (typeof latestContract.methods === 'string') {
            methodsList = JSON.parse(latestContract.methods)
          }
        }
      } catch (e) {
        console.warn('解析methods失败:', e, '原始数据:', latestContract.methods)
        methodsList = []
      }
      
      try {
        if (latestContract.events) {
          if (Array.isArray(latestContract.events)) {
            eventsList = latestContract.events
          } else if (typeof latestContract.events === 'string') {
            eventsList = JSON.parse(latestContract.events)
          }
        }
      } catch (e) {
        console.warn('解析events失败:', e, '原始数据:', latestContract.events)
        eventsList = []
      }
      
      try {
        if (latestContract.metadata) {
          if (typeof latestContract.metadata === 'object' && latestContract.metadata !== null) {
            metadataObj = latestContract.metadata
          } else {
            metadataObj = JSON.parse(latestContract.metadata)
          }
        }
      } catch (e) {
        console.warn('解析metadata失败:', e, '原始数据:', latestContract.metadata)
        metadataObj = {}
      }
      
      console.log('解析后的最新数据:', {
        interfacesList,
        methodsList,
        eventsList,
        metadataObj
      })
      
      // 使用最新数据填充编辑表单
      editingAddress.value = {
        hash: latestContract.address || address.hash,
        type: latestContract.contract_type?.toLowerCase() || address.type,
        name: latestContract.name || address.name,
        symbol: latestContract.symbol || address.symbol,
        status: latestContract.status === 1 ? 'active' : 'inactive',
        transactionCount: address.transactionCount || 0,
        lastActivity: address.lastActivity || 0,
        description: address.description || '',
        decimals: latestContract.decimals || address.decimals || 0,
        totalSupply: latestContract.total_supply || address.totalSupply || '',
        contractLogo: latestContract.contract_logo || address.contractLogo || '',
        interfaces: latestContract.interfaces || address.interfaces || '',
        methods: latestContract.methods || address.methods || '',
        events: latestContract.events || address.events || '',
        metadata: latestContract.metadata || address.metadata || '',
        interfacesList,
        methodsList,
        eventsList,
        metadataObj,
        isErc20: latestContract.is_erc20 || address.isErc20 || false,
        verified: latestContract.verified || address.verified || false,
        creator: latestContract.creator || address.creator || '',
        creationTx: latestContract.creation_tx || address.creationTx || '',
        creationBlock: latestContract.creation_block || address.creationBlock || 0
      }
      
      
    } else {
      console.warn('获取最新合约数据失败，使用本地数据')
      // 如果获取失败，回退到使用本地数据
      await editAddressWithLocalData(address)
    }
  } catch (error) {
    console.error('编辑地址时出错:', error)
    showError('获取合约信息失败，使用本地数据')
    // 如果出错，回退到使用本地数据
    await editAddressWithLocalData(address)
  }
}

// 使用本地数据编辑地址（回退方案）
const editAddressWithLocalData = async (address: Address) => {
  console.log('使用本地数据编辑地址:', address)
  
  // 解析现有数据
  let interfacesList: string[] = []
  let methodsList: string[] = []
  let eventsList: string[] = []
  let metadataObj: Record<string, string> = {}
  
  try {
    if (address.interfaces) {
      // 如果已经是数组格式，直接使用
      if (Array.isArray(address.interfaces)) {
        interfacesList = address.interfaces
      } else if (typeof address.interfaces === 'string') {
        // 尝试解析JSON字符串
        interfacesList = JSON.parse(address.interfaces)
      }
    }
  } catch (e) {
    console.warn('解析interfaces失败:', e, '原始数据:', address.interfaces)
    interfacesList = []
  }
  
  try {
    if (address.methods) {
      if (Array.isArray(address.methods)) {
        methodsList = address.methods
      } else {
        methodsList = JSON.parse(address.methods)
      }
    }
  } catch (e) {
    console.warn('解析methods失败:', e, '原始数据:', address.methods)
    methodsList = []
  }
  
  try {
    if (address.events) {
      if (Array.isArray(address.events)) {
        eventsList = address.events
      } else {
        eventsList = JSON.parse(address.events)
      }
    }
  } catch (e) {
    console.warn('解析events失败:', e, '原始数据:', address.events)
    eventsList = []
  }
  
  try {
    if (address.metadata) {
      if (typeof address.metadata === 'object' && address.metadata !== null) {
        metadataObj = address.metadata
      } else {
        metadataObj = JSON.parse(address.metadata)
      }
    }
  } catch (e) {
    console.warn('解析metadata失败:', e, '原始数据:', address.metadata)
    metadataObj = {}
  }
  
  editingAddress.value = { 
    ...address,
    interfacesList,
    methodsList,
    eventsList,
    metadataObj
  }
}

// 关闭编辑弹窗并清除缓存
const closeEditModal = () => {
  showEditModal.value = false
  
  // 清除所有输入变量
  newInterface.value = ''
  newMethod.value = ''
  newEvent.value = ''
  newMetadataKey.value = ''
  newMetadataValue.value = ''
  
  // 重置编辑地址数据
  editingAddress.value = {
    hash: '',
    type: '',
    name: '',
    symbol: '',
    status: 'active',
    transactionCount: 0,
    lastActivity: 0,
    description: '',
    decimals: 0,
    totalSupply: '',
    contractLogo: '',
    interfaces: '',
    methods: '',
    events: '',
    metadata: '',
    interfacesList: [],
    methodsList: [],
    eventsList: [],
    metadataObj: {},
    isErc20: false,
    verified: false,
    creator: '',
    creationTx: '',
    creationBlock: 0
  }
}

// 关闭添加弹窗并清除缓存
const closeAddModal = () => {
  showAddAddressModal.value = false
  
  // 清除所有输入变量
  newInterface.value = ''
  newMethod.value = ''
  newEvent.value = ''
  newMetadataKey.value = ''
  newMetadataValue.value = ''
  
  // 重置新增地址数据
  newAddress.value = {
    hash: '',
    type: 'erc20',
    name: '',
    symbol: '',
    status: 'active',
    transactionCount: 0,
    lastActivity: 0,
    description: '',
    decimals: 18,
    totalSupply: '',
    contractLogo: '',
    interfaces: '',
    methods: '',
    events: '',
    metadata: '',
    interfacesList: [],
    methodsList: [],
    eventsList: [],
    metadataObj: {},
    isErc20: true,
    verified: false,
    creator: '',
    creationTx: '',
    creationBlock: 0
  }
}

// 处理Logo上传
const handleLogoUpload = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    const file = target.files[0]
    const reader = new FileReader()
    
    reader.onload = (e) => {
      if (e.target?.result) {
        editingAddress.value.contractLogo = e.target.result as string
      }
    }
    
    reader.readAsDataURL(file)
  }
}

// 处理图片加载错误
const handleImageError = (event: Event) => {
  const target = event.target as HTMLImageElement
  // 隐藏错误的图片，不显示任何替代内容
  target.style.display = 'none'
}

// 保存编辑
const saveEdit = async () => {
  try {
    console.log('保存编辑:', editingAddress.value)
    
    // 转换前端数据为后端API需要的格式 - 直接传递数组/对象，避免JSON字符串转换
    const contractData = {
      address: editingAddress.value.hash,
      chain_name: 'eth', // 默认以太坊
      contract_type: editingAddress.value.type.toUpperCase(),
      name: editingAddress.value.name || undefined,
      symbol: editingAddress.value.symbol || undefined,
      decimals: editingAddress.value.decimals || 0,
      total_supply: editingAddress.value.totalSupply || '',
      is_erc20: editingAddress.value.isErc20 || false,
      interfaces: editingAddress.value.interfacesList || [], // 直接传递数组
      methods: editingAddress.value.methodsList || [], // 直接传递数组
      events: editingAddress.value.eventsList || [], // 直接传递数组
      metadata: editingAddress.value.metadataObj || {}, // 直接传递对象
      status: editingAddress.value.status === 'active' ? 1 : 0,
      verified: editingAddress.value.verified || false,
      creator: editingAddress.value.creator || '',
      creation_tx: editingAddress.value.creationTx || '',
      creation_block: editingAddress.value.creationBlock || 0,
      contract_logo: editingAddress.value.contractLogo || ''
    }
    
    console.log('保存前的数据:', {
      interfacesList: editingAddress.value.interfacesList,
      methodsList: editingAddress.value.methodsList,
      eventsList: editingAddress.value.eventsList,
      metadataObj: editingAddress.value.metadataObj
    })
    
    console.log('发送到后端的数据:', {
      interfaces: contractData.interfaces,
      methods: contractData.methods,
      events: contractData.events,
      metadata: contractData.metadata
    })
    
    // 根据登录状态调用不同的API
    let response
    if (authStore.isAuthenticated) {
      // 已登录用户：调用 /v1/ 下的API
      const { contracts } = await import('@/api')
      response = await contracts.createOrUpdateContract(contractData)
    } else {
      // 未登录用户无法编辑合约
      showError('游客模式下无法编辑合约，请先登录')
      return
    }
    
    if (response.success) {
      console.log('保存成功:', response.data)
      
      // 更新本地数据
      const index = addresses.value.findIndex(addr => addr.hash === editingAddress.value.hash)
      if (index !== -1) {
        addresses.value[index] = { ...editingAddress.value }
      }
      
      closeEditModal()
      showSuccess('合约信息保存成功！')
    } else {
      console.error('保存失败:', response.message)
      showError(`保存失败: ${response.message || '未知错误'}`)
    }
  } catch (error) {
    console.error('保存失败:', error)
    showError(`保存失败: ${error instanceof Error ? error.message : '未知错误'}`)
  }
}

// 处理新增Logo上传
const handleNewLogoUpload = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    const file = target.files[0]
    const reader = new FileReader()
    
    reader.onload = (e) => {
      if (e.target?.result) {
        newAddress.value.contractLogo = e.target.result as string
      }
    }
    
    reader.readAsDataURL(file)
  }
}

// 接口管理函数
const addInterface = () => {
  if (newInterface.value.trim()) {
    if (!editingAddress.value.interfacesList) {
      editingAddress.value.interfacesList = []
    }
    editingAddress.value.interfacesList.push(newInterface.value.trim())
    newInterface.value = ''
  }
}

const removeInterface = (index: number) => {
  if (editingAddress.value.interfacesList) {
    editingAddress.value.interfacesList.splice(index, 1)
  }
}

// 方法管理函数
const addMethod = () => {
  if (newMethod.value.trim()) {
    if (!editingAddress.value.methodsList) {
      editingAddress.value.methodsList = []
    }
    editingAddress.value.methodsList.push(newMethod.value.trim())
    newMethod.value = ''
  }
}

const removeMethod = (index: number) => {
  if (editingAddress.value.methodsList) {
    editingAddress.value.methodsList.splice(index, 1)
  }
}

// 事件管理函数
const addEvent = () => {
  if (newEvent.value.trim()) {
    if (!editingAddress.value.eventsList) {
      editingAddress.value.eventsList = []
    }
    editingAddress.value.eventsList.push(newEvent.value.trim())
    newEvent.value = ''
  }
}

const removeEvent = (index: number) => {
  if (editingAddress.value.eventsList) {
    editingAddress.value.eventsList.splice(index, 1)
  }
}

// 元数据管理函数
const addMetadata = () => {
  if (newMetadataKey.value.trim() && newMetadataValue.value.trim()) {
    if (!editingAddress.value.metadataObj) {
      editingAddress.value.metadataObj = {}
    }
    editingAddress.value.metadataObj[newMetadataKey.value.trim()] = newMetadataValue.value.trim()
    newMetadataKey.value = ''
    newMetadataValue.value = ''
  }
}

const removeMetadata = (key: string) => {
  if (editingAddress.value.metadataObj) {
    delete editingAddress.value.metadataObj[key]
  }
}

// 新建模态框的接口管理函数
const addNewInterface = () => {
  if (newInterface.value.trim()) {
    if (!newAddress.value.interfacesList) {
      newAddress.value.interfacesList = []
    }
    newAddress.value.interfacesList.push(newInterface.value.trim())
    newInterface.value = ''
  }
}

const removeNewInterface = (index: number) => {
  if (newAddress.value.interfacesList) {
    newAddress.value.interfacesList.splice(index, 1)
  }
}

// 新建模态框的方法管理函数
const addNewMethod = () => {
  if (newMethod.value.trim()) {
    if (!newAddress.value.methodsList) {
      newAddress.value.methodsList = []
    }
    newAddress.value.methodsList.push(newMethod.value.trim())
    newMethod.value = ''
  }
}

const removeNewMethod = (index: number) => {
  if (newAddress.value.methodsList) {
    newAddress.value.methodsList.splice(index, 1)
  }
}

// 新建模态框的事件管理函数
const addNewEvent = () => {
  if (newEvent.value.trim()) {
    if (!newAddress.value.eventsList) {
      newAddress.value.eventsList = []
    }
    newAddress.value.eventsList.push(newEvent.value.trim())
    newEvent.value = ''
  }
}

const removeNewEvent = (index: number) => {
  if (newAddress.value.eventsList) {
    newAddress.value.eventsList.splice(index, 1)
  }
}

// 新建模态框的元数据管理函数
const addNewMetadata = () => {
  if (newMetadataKey.value.trim() && newMetadataValue.value.trim()) {
    if (!newAddress.value.metadataObj) {
      newAddress.value.metadataObj = {}
    }
    newAddress.value.metadataObj[newMetadataKey.value.trim()] = newMetadataValue.value.trim()
    newMetadataKey.value = ''
    newMetadataValue.value = ''
  }
}

const removeNewMetadata = (key: string) => {
  if (newAddress.value.metadataObj) {
    delete newAddress.value.metadataObj[key]
  }
}

// 刷新币种配置数据
const refreshCoinConfigData = async () => {
  try {
    console.log('🔄 手动刷新币种配置数据')
    if (coinConfigData.value.contract_address) {
      await maintainCoinConfig({ hash: coinConfigData.value.contract_address } as Address)
      showSuccess('数据已刷新')
    }
  } catch (error) {
    console.error('刷新币种配置数据失败:', error)
    showError('刷新数据失败')
  }
}

// 维护币种信息
const maintainCoinConfig = async (address: Address) => {
  try {
    console.log('🔄 维护币种信息 - 重新获取最新数据:', address.hash)
    
    // 重置币种配置数据
    coinConfigData.value = {
      contract_address: address.hash,
      name: address.name || '',
      symbol: address.symbol || '',
      coin_type: 1,
      precision: address.decimals || 18,
      decimals: address.decimals || 18,
      status: 1,
      logo_url: address.contractLogo || '',
      is_verified: address.verified || false,
      market_cap_rank: 0,
      is_stablecoin: false,
      website_url: '',
      explorer_url: '',
      description: '',
      parser_configs: []
    }
    
    // 使用标准API模块调用后端接口，确保获取最新数据
    // 添加时间戳参数避免浏览器缓存
    const timestamp = Date.now()
    console.log(`📡 发起API请求，时间戳: ${timestamp}`)
    
    // 强制刷新：清除可能的前端缓存数据
    coinConfigData.value.parser_configs = []
    
    const response = await getCoinConfigMaintenance(address.hash)
    
    if (response.success) {
      console.log('获取币种配置成功:', response.data)
      
      if (response.data.coin_config) {
        // 如果存在币种配置，填充数据
        const config = response.data.coin_config
        coinConfigData.value = {
          contract_address: config.contract_addr || address.hash,
          name: config.name || address.name || '',
          symbol: config.symbol || address.symbol || '',
          coin_type: config.coin_type || 1,
          precision: config.precision || address.decimals || 18,
          decimals: config.decimals || address.decimals || 18,
          status: config.status || 1,
          logo_url: config.logo_url || address.contractLogo || '',
          is_verified: config.is_verified || address.verified || false,
          market_cap_rank: config.market_cap_rank || 0,
          is_stablecoin: config.is_stablecoin || false,
          website_url: config.website_url || '',
          explorer_url: config.explorer_url || '',
          description: config.description || '',
          parser_configs: response.data.parser_configs || []
        }
        console.log('填充币种配置数据:', coinConfigData.value)
      } else {
        console.log('没有现有币种配置，使用默认值')
      }
      
      // 处理解析配置数据，确保包含日志解析配置字段
      if (response.data.parser_configs && response.data.parser_configs.length > 0) {
        coinConfigData.value.parser_configs = response.data.parser_configs.map((config: any) => ({
          id: config.id, // 保留原始ID，用于区分创建还是更新
          function_name: config.function_name || '',
          function_signature: config.function_signature || '',
          function_description: config.function_description || '',
          display_format: config.display_format || '',
          is_active: config.is_active !== undefined ? config.is_active : true,
          param_config: config.param_config || [],
          parser_rules: config.parser_rules || {
            extract_to_address: '',
            extract_amount: '',
            amount_unit: '',
            extract_data: ''
          },
          // 日志解析配置字段
          logs_parser_type: config.logs_parser_type || 'input_data',
          event_signature: config.event_signature || '',
          event_name: config.event_name || '',
          event_description: config.event_description || '',
          logs_param_config: config.logs_param_config || [],
          logs_parser_rules: config.logs_parser_rules || {
            extract_from_address: '',
            extract_to_address: '',
            extract_amount: '',
            amount_unit: '',
            extract_owner_address: '',
            extract_spender_address: ''
          },
          logs_display_format: config.logs_display_format || ''
        }))
      }
      
      console.log('解析配置数据:', response.data.parser_configs)
      console.log('最终币种配置数据:', coinConfigData.value)
      
      showCoinConfigModal.value = true
    } else {
      showError(`获取币种配置失败: ${response.error || '未知错误'}`)
    }
  } catch (error) {
    console.error('维护币种信息失败:', error)
    showError(`维护币种信息失败: ${error instanceof Error ? error.message : '未知错误'}`)
  }
}

// 关闭币种配置模态框
const closeCoinConfigModal = () => {
  showCoinConfigModal.value = false
  // 重置数据
  coinConfigData.value = {
    contract_address: '',
    name: '',
    symbol: '',
    coin_type: 1,
    precision: 18,
    decimals: 18,
    status: 1,
    logo_url: '',
    is_verified: false,
    market_cap_rank: 0,
    is_stablecoin: false,
    website_url: '',
    explorer_url: '',
    description: '',
    parser_configs: []
  }
}

// 添加解析配置
const addParserConfig = () => {
  coinConfigData.value.parser_configs.push({
    function_name: '',
    function_signature: '',
    function_description: '',
    display_format: '',
    is_active: true,
    param_config: [],
    parser_rules: {
      extract_to_address: '',
      extract_amount: '',
      amount_unit: '',
      extract_data: ''
    },
    // 新增日志解析配置字段
    logs_parser_type: 'input_data',
    event_signature: '',
    event_name: '',
    event_description: '',
    logs_param_config: [],
    logs_parser_rules: {
      extract_from_address: '',
      extract_to_address: '',
      extract_amount: '',
      amount_unit: '',
      extract_owner_address: '',
      extract_spender_address: ''
    },
    logs_display_format: ''
  })
}

// 添加参数配置
const addParamConfig = (configIndex: number) => {
  if (!coinConfigData.value.parser_configs[configIndex].param_config) {
    coinConfigData.value.parser_configs[configIndex].param_config = []
  }
  coinConfigData.value.parser_configs[configIndex].param_config.push({
    name: '',
    type: 'uint256',
    offset: 0,
    length: 32,
    description: ''
  })
}

// 添加日志参数配置
const addLogsParamConfig = (configIndex: number) => {
  if (!coinConfigData.value.parser_configs[configIndex].logs_param_config) {
    coinConfigData.value.parser_configs[configIndex].logs_param_config = []
  }
  coinConfigData.value.parser_configs[configIndex].logs_param_config.push({
    name: '',
    type: 'uint256',
    topic_index: 0,
    data_index: 0,
    description: ''
  })
}

// 删除参数配置
const removeParamConfig = (configIndex: number, paramIndex: number) => {
  if (coinConfigData.value.parser_configs[configIndex].param_config) {
    coinConfigData.value.parser_configs[configIndex].param_config.splice(paramIndex, 1)
  }
}

// 删除日志参数配置
const removeLogsParamConfig = (configIndex: number, paramIndex: number) => {
  if (coinConfigData.value.parser_configs[configIndex].logs_param_config) {
    coinConfigData.value.parser_configs[configIndex].logs_param_config.splice(paramIndex, 1)
  }
}

// 删除解析配置
const removeParserConfig = (index: number) => {
  coinConfigData.value.parser_configs.splice(index, 1)
}

// 保存解析配置
const saveParserConfigs = async (contractAddress: string, parserConfigs: any[]) => {
  // 这里可以调用后端API保存解析配置
  // 暂时使用console.log模拟
  console.log('保存解析配置:', { contractAddress, parserConfigs })
  
  // TODO: 实现真实的解析配置保存API调用
  // const response = await saveParserConfigsAPI(contractAddress, parserConfigs)
  // return response
}

// 保存币种配置
const saveCoinConfig = async () => {
  try {
    console.log('保存币种配置:', coinConfigData.value)
    
    // 构建保存数据
    const saveData = {
      contract_addr: coinConfigData.value.contract_address,
      chain_name: 'eth',
      symbol: coinConfigData.value.symbol,
      coin_type: coinConfigData.value.coin_type,
      precision: coinConfigData.value.precision,
      decimals: coinConfigData.value.precision,
      name: coinConfigData.value.name,
      logo_url: coinConfigData.value.logo_url,
      is_verified: coinConfigData.value.is_verified,
      status: coinConfigData.value.status
    }
    
    // 使用标准API模块调用保存接口
    const response = await createCoinConfig(saveData)
    
    if (response.success) {
      // 如果币种配置保存成功，还需要保存解析配置
      if (coinConfigData.value.parser_configs.length > 0) {
        try {
          // 准备解析配置数据
          const parserConfigs: ParserConfig[] = coinConfigData.value.parser_configs.map(config => {
            const result = {
              id: config.id, // 如果有ID说明是更新，没有ID说明是新建
              contract_address: coinConfigData.value.contract_address,
              parser_type: 'input_data' as const, // 默认类型，可以根据需要调整
              function_signature: config.function_signature || '',
              function_name: config.function_name || '',
              function_description: config.function_description || '',
              param_config: config.param_config || [],
              parser_rules: config.parser_rules || {},
              display_format: config.display_format || '',
              is_active: config.is_active !== undefined ? config.is_active : true,
              priority: config.priority || 0,
              // 日志解析配置字段
              logs_parser_type: config.logs_parser_type || 'input_data',
              event_signature: config.event_signature || '',
              event_name: config.event_name || '',
              event_description: config.event_description || '',
              logs_param_config: config.logs_param_config || [],
              logs_parser_rules: config.logs_parser_rules || {},
              logs_display_format: config.logs_display_format || ''
            }
            
            // 添加调试日志，显示每个配置是创建还是更新
            if (config.id) {
              console.log(`🔄 准备更新解析配置 ID: ${config.id}, 函数: ${config.function_name}`)
            } else {
              console.log(`🆕 准备创建新解析配置, 函数: ${config.function_name}`)
            }
            
            return result
          })
          
          console.log('准备保存的解析配置:', parserConfigs)
          
          // 调用批量保存解析配置API
          const parserResponse = await batchSaveParserConfigs(coinConfigData.value.contract_address, parserConfigs)
          
          if (parserResponse.success) {
            showSuccess('币种配置和解析配置保存成功！')
          } else {
            showSuccess('币种配置保存成功！解析配置保存失败，请稍后重试。')
          }
        } catch (parserError) {
          console.warn('币种配置保存成功，但解析配置保存失败:', parserError)
          showSuccess('币种配置保存成功！解析配置保存失败，请稍后重试。')
        }
      } else {
        showSuccess('币种配置保存成功！')
      }
      closeCoinConfigModal()
    } else {
      showError(`保存失败: ${response.error || '未知错误'}`)
    }
  } catch (error) {
    console.error('保存币种配置失败:', error)
    showError(`保存失败: ${error instanceof Error ? error.message : '未知错误'}`)
  }
}

// 保存新增
const saveAdd = async () => {
  try {
    console.log('保存新增:', newAddress.value)
    
    // 转换前端数据为后端API需要的格式
    const contractData = {
      address: newAddress.value.hash,
      chain_name: 'eth', // 默认以太坊
      contract_type: newAddress.value.type.toUpperCase(),
      name: newAddress.value.name || undefined,
      symbol: newAddress.value.symbol || undefined,
      decimals: newAddress.value.decimals || 18,
      total_supply: newAddress.value.totalSupply || '',
      is_erc20: newAddress.value.isErc20 || true,
      interfaces: newAddress.value.interfacesList || [], // 直接传递数组
      methods: newAddress.value.methodsList || [], // 直接传递数组
      events: newAddress.value.eventsList || [], // 直接传递数组
      metadata: newAddress.value.metadataObj || {}, // 直接传递对象
      status: newAddress.value.status === 'active' ? 1 : 0,
      verified: newAddress.value.verified || false,
      creator: newAddress.value.creator || '',
      creation_tx: newAddress.value.creationTx || '',
      creation_block: newAddress.value.creationBlock || 0,
      contract_logo: newAddress.value.contractLogo || ''
    }
    
    // 根据登录状态调用不同的API
    let response
    if (authStore.isAuthenticated) {
      // 已登录用户：调用 /v1/ 下的API
      const { contracts } = await import('@/api')
      response = await contracts.createOrUpdateContract(contractData)
    } else {
      // 未登录用户无法添加合约
      showError('游客模式下无法添加合约，请先登录')
      return
    }
    
    if (response.success) {
      console.log('添加成功:', response.data)
      
      // 添加到本地数据
      addresses.value.unshift({
        ...newAddress.value,
        transactionCount: 0,
        lastActivity: Date.now() / 1000
      })
      
      // 更新总数
      totalAddresses.value++
      
      // 重置表单
      newAddress.value = {
        hash: '',
        type: 'erc20',
        name: '',
        symbol: '',
        status: 'active',
        transactionCount: 0,
        lastActivity: 0,
        description: '',
        decimals: 18,
        totalSupply: '',
        contractLogo: '',
        interfaces: '',
        methods: '',
        events: '',
        metadata: '',
        interfacesList: [],
        methodsList: [],
        eventsList: [],
        metadataObj: {},
        isErc20: true,
        verified: false,
        creator: '',
        creationTx: '',
        creationBlock: 0
      }
      
      closeAddModal()
      showSuccess('新合约添加成功！')
    } else {
      console.error('添加失败:', response.message)
      showError(`添加失败: ${response.message || '未知错误'}`)
    }
  } catch (error) {
    console.error('保存失败:', error)
    showError(`保存失败: ${error instanceof Error ? error.message : '未知错误'}`)
  }
}

// 切换地址状态
const toggleAddressStatus = (address: Address) => {
  console.log('切换地址状态:', address)
  // 这里可以实现状态切换逻辑
}

// 数据加载
const loadData = async () => {
  try {
    const params = {
      page: currentPage.value,
      page_size: pageSize.value,
      chainName: 'eth', // 默认以太坊
      contractType: typeFilter.value || undefined,
      status: statusFilter.value || undefined,
      search: searchQuery.value || undefined
    }
    
    // 根据登录状态调用不同的API
    let response
    if (authStore.isAuthenticated) {
      // 已登录用户：调用 /v1/ 下的API
      const { contracts } = await import('@/api')
      response = await contracts.getContracts(params) as unknown as ContractsResponse
    } else {
      // 未登录用户：调用 /no-auth/ 下的API
      const { noAuth } = await import('@/api')
      response = await noAuth.getContracts(params) as unknown as ContractsResponse
    }
    
    if (response.success) {
      // 转换API数据为页面需要的格式
      addresses.value = response.data.map(contract => ({
        hash: contract.address, // 使用 address 字段
        type: contract.contract_type?.toLowerCase() || 'unknown',
        name: contract.name || '未命名合约', // 使用 name 字段
        symbol: contract.symbol, // 使用 symbol 字段
        status: contract.status === 1 ? 'active' : 'inactive', // 转换状态值
        transactionCount: 0, // 保留字段但不显示
        lastActivity: new Date(contract.m_time || contract.c_time).getTime() / 1000, // 保留字段但不显示
        description: contract.metadata ? (typeof contract.metadata === 'string' ? JSON.parse(contract.metadata).description : contract.metadata.description) : null, // 从 metadata 中提取描述
        decimals: contract.decimals, // 新增：精度
        totalSupply: contract.total_supply, // 新增：总供应量
        contractLogo: contract.contract_logo, // 新增：合约Logo
        // 新增：合约详细信息字段
        interfaces: contract.interfaces,
        methods: contract.methods,
        events: contract.events,
        metadata: contract.metadata,
        isErc20: contract.is_erc20,
        verified: contract.verified,
        creator: contract.creator,
        creationTx: contract.creation_tx,
        creationBlock: contract.creation_block
      }))
      
      // 使用 count 字段，这是实际后端返回的格式
      totalAddresses.value = response.count || 0
    } else {
      console.error('获取合约数据失败:', response.message)
      showError(`获取合约数据失败: ${response.message || '未知错误'}`)
      addresses.value = []
      totalAddresses.value = 0
    }
  } catch (error) {
    console.error('加载合约数据出错:', error)
    showError(`加载合约数据出错: ${error instanceof Error ? error.message : '未知错误'}`)
    addresses.value = []
    totalAddresses.value = 0
  }
}

// 分页方法
const previousPage = async () => {
  if (currentPage.value > 1) {
    currentPage.value--
    await loadData()
  }
}

const nextPage = async () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    await loadData()
  }
}

const goToPage = async (page: number) => {
  currentPage.value = page
  await loadData()
}

// 监听搜索查询
watch(searchQuery, (newQuery) => {
  currentPage.value = 1
  loadData()
})

// 监听类型筛选
watch(typeFilter, async () => {
  currentPage.value = 1
  await loadData()
})

// 监听状态筛选
watch(statusFilter, async () => {
  currentPage.value = 1
  await loadData()
})

// 监听页面大小变化
watch(pageSize, async () => {
  currentPage.value = 1
  await loadData()
})

// 是否管理员（用于权限控制）
const isAdmin = ref(false)

// 获取用户资料并设置角色
const fetchUserProfile = async () => {
  try {
    const res: any = await request.get('/api/user/profile')
    if (res?.success && res.data) {
      const role = res.data.role || res.data.Role || ''
      isAdmin.value = role.toLowerCase() === 'administrator'
    }
  } catch (e) {
    isAdmin.value = false
  }
}

// 组件挂载时加载数据
onMounted(async () => {
  await fetchUserProfile()
  await loadData()
})
</script> 