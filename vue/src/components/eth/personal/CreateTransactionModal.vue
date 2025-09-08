<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[95vh] flex flex-col">
        <div class="px-6 py-4 border-b border-gray-200">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-medium text-gray-900">{{ isEditMode ? '编辑交易' : '新建交易' }}</h3>
            <button
              @click="$emit('close')"
              class="text-gray-400 hover:text-gray-600 transition-colors"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
        
        <form @submit.prevent="handleSubmit" class="flex flex-col flex-1 overflow-hidden">
          <div class="px-6 py-4 flex-1 overflow-y-auto">
            <div class="space-y-6">
            <!-- 链类型 - 固定为ETH -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">链类型</label>
              <div class="px-3 py-2 bg-gray-100 border border-gray-300 rounded-md text-gray-700">
                以太坊 (ETH)
              </div>
            </div>
            
            <!-- 交易类型选择 -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">交易类型</label>
              <div class="flex space-x-4">
                <label class="flex items-center">
                  <input
                    type="radio"
                    v-model="transactionType"
                    value="eth"
                    class="mr-2 text-blue-600"
                    @change="handleTransactionTypeChange"
                  />
                  <span class="text-sm text-gray-700">ETH转账</span>
                </label>
                <label class="flex items-center">
                  <input
                    type="radio"
                    v-model="transactionType"
                    value="erc20"
                    class="mr-2 text-blue-600"
                    @change="handleTransactionTypeChange"
                  />
                  <span class="text-sm text-gray-700">ERC-20代币</span>
                </label>
              </div>
            </div>

            <!-- 合约操作类型选择 (仅ERC-20时显示) -->
            <div v-if="transactionType === 'erc20'">
              <label class="block text-sm font-medium text-gray-700 mb-2">合约操作类型</label>
              <div class="flex space-x-4">
                <label class="flex items-center">
                  <input
                    type="radio"
                    v-model="contractOperationType"
                    value="transfer"
                    class="mr-2 text-blue-600"
                    @change="handleContractOperationTypeChange"
                  />
                  <span class="text-sm text-gray-700">转账</span>
                </label>
                <label class="flex items-center">
                  <input
                    type="radio"
                    v-model="contractOperationType"
                    value="approve"
                    class="mr-2 text-blue-600"
                    @change="handleContractOperationTypeChange"
                  />
                  <span class="text-sm text-gray-700">授权</span>
                </label>
                <label class="flex items-center">
                  <input
                    type="radio"
                    v-model="contractOperationType"
                    value="transferFrom"
                    class="mr-2 text-blue-600"
                    @change="handleContractOperationTypeChange"
                  />
                  <span class="text-sm text-gray-700">授权转账</span>
                </label>
                
              </div>
            </div>

            <!-- 代币选择 (仅ERC-20时显示) -->
            <div v-if="transactionType === 'erc20'">
              <label class="block text-sm font-medium text-gray-700 mb-2">选择代币</label>
              <div class="relative">
                <input
                  v-model="selectedTokenSearch"
                  type="text"
                  @focus="showTokenDropdown = true"
                  @blur="handleTokenDropdownBlur"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="搜索代币名称或合约地址"
                  required
                />
                <!-- 代币下拉选择 -->
                <div v-if="showTokenDropdown && filteredTokens.length > 0" class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-y-auto">
                  <div
                    v-for="token in filteredTokens"
                    :key="token.contract_address"
                    @click="selectToken(token)"
                    class="px-3 py-2 hover:bg-gray-100 cursor-pointer border-b border-gray-100 last:border-b-0"
                  >
                    <div class="flex items-center justify-between">
                      <div>
                        <div class="font-medium text-gray-900">{{ token.symbol }} - {{ token.name }}</div>
                        <div class="text-sm text-gray-500 font-mono">{{ token.contract_address }}</div>
                      </div>
                      <div class="text-right">
                        <div class="text-sm text-gray-600">精度: {{ token.decimals }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>



            <!-- 发送地址 - 智能下拉选择 -->
            <div v-if="shouldShowFromAddress">
              <label class="block text-sm font-medium text-gray-700 mb-2">
                {{ getFromAddressLabel() }}
              </label>
              <div class="relative">
                <input
                  v-model="form.from_address"
                  type="text"
                  @focus="showFromAddressDropdown = true"
                  @blur="handleFromAddressBlur"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  :placeholder="getFromAddressPlaceholder()"
                  required
                />
                <!-- 下拉选择 -->
                <div v-if="showFromAddressDropdown && filteredFromAddresses.length > 0" class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-y-auto">
                  <div
                    v-for="address in filteredFromAddresses"
                    :key="address.id"
                    @click="selectFromAddress(address)"
                    class="px-3 py-2 hover:bg-gray-100 cursor-pointer border-b border-gray-100 last:border-b-0"
                  >
                    <div class="flex items-center justify-between">
                      <div>
                        <div class="font-medium text-gray-900">{{ address.label }}</div>
                        <div class="text-sm text-gray-500 font-mono">{{ address.address }}</div>
                      </div>
                      <div class="text-right">
                        <div class="text-sm text-gray-600">{{ address.type }}</div>
                        <div class="text-xs text-gray-500">{{ formatBalance(address.balance) }} {{ form.symbol }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 授权地址选择 - 仅在transferFrom时显示（不再依赖是否已选择发送地址） -->
            <div v-if="contractOperationType === 'transferFrom'">
              <label class="block text-sm font-medium text-gray-700 mb-2">
                代币持有者地址
              </label>
              <div class="relative">
                <input
                  v-model="form.allowance_address"
                  type="text"
                  @focus="showAllowanceAddressDropdown = true"
                  @blur="handleAllowanceAddressBlur"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="选择或输入代币持有者地址"
                  required
                />
                <!-- 下拉选择 -->
                <div v-if="showAllowanceAddressDropdown && authorizedAddresses.length > 0" class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-y-auto">
                  <div
                    v-for="address in authorizedAddresses"
                    :key="address.address"
                    @click="selectAllowanceAddress(address)"
                    class="px-3 py-2 hover:bg-gray-100 cursor-pointer border-b border-gray-100 last:border-b-0"
                  >
                    <div class="flex items-center justify-between">
                      <div>
                        <div class="font-medium text-gray-900">{{ address.label }}</div>
                        <div class="text-sm text-gray-500 font-mono">{{ address.address }}</div>
                      </div>
                      <div class="text-right">
                        <div class="text-sm text-gray-600">{{ address.type }}</div>
                        <div class="text-xs text-gray-500">{{ formatBalance(address.balance) }} {{ form.symbol }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <p class="mt-1 text-sm text-gray-500">
                选择被授权地址 {{ form.from_address || '（请先选择发送地址）' }} 可以操作的代币持有者地址
              </p>
            </div>

            <!-- 接收地址 - 智能下拉选择 -->
            <div v-if="shouldShowToAddress">
              <label class="block text-sm font-medium text-gray-700 mb-2">
                {{ getToAddressLabel() }}
              </label>
              <div class="relative">
                <input
                  v-model="form.to_address"
                  type="text"
                  @focus="showToAddressDropdown = true"
                  @blur="handleToAddressBlur"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  :placeholder="getToAddressPlaceholder()"
                  required
                />
                <!-- 下拉选择 -->
                <div v-if="showToAddressDropdown && filteredToAddresses.length > 0" class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-y-auto">
                  <div
                    v-for="address in filteredToAddresses"
                    :key="address.id"
                    @click="selectToAddress(address)"
                    class="px-3 py-2 hover:bg-gray-100 cursor-pointer border-b border-gray-100 last:border-b-0"
                  >
                    <div class="flex items-center justify-between">
                      <div>
                        <div class="font-medium text-gray-900">{{ address.label }}</div>
                        <div class="text-sm text-gray-500 font-mono">{{ address.address }}</div>
                      </div>
                      <div class="text-right">
                        <div class="text-sm text-gray-600">{{ address.type }}</div>
                        <div class="text-xs text-gray-500">{{ formatBalance(address.balance) }} {{ form.symbol }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 交易金额 -->
            <div v-if="shouldShowAmount">
              <label class="block text-sm font-medium text-gray-700 mb-2">
                {{ getAmountLabel() }}
              </label>
              <input
                v-model="form.amount"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                :placeholder="getAmountPlaceholder()"
                required
              />
              <!-- 显示单位 -->
              <div class="mt-1 text-sm text-gray-500">
                <span v-if="transactionType === 'eth'">{{ formatToWei(form.amount) }} wei</span>
                <span v-else-if="selectedToken">{{ formatToTokenUnits(form.amount, selectedToken.decimals) }} {{ selectedToken.symbol }}</span>
                <span v-else>0 {{ form.symbol }}</span>
              </div>
            </div>

            </div>
          </div>

          <!-- 操作按钮 - 固定在底部 -->
          <div class="px-6 py-4 border-t border-gray-200 bg-white flex-shrink-0">
            <div class="flex justify-end">
              <button
                type="submit"
                :disabled="isSubmitting"
                class="px-6 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {{ isSubmitting ? (isEditMode ? '更新中...' : '创建中...') : (isEditMode ? '更新交易' : '创建交易') }}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, computed, onMounted } from 'vue'
import { createUserTransaction, updateUserTransaction } from '@/api/user-transactions'
import { getPersonalAddresses } from '@/api/personal-addresses'
import { listCoinConfigs } from '@/api/coinconfig'
import type { CreateUserTransactionRequest } from '@/types'
import type { PersonalAddressItem } from '@/types/personal-address'
import type { CoinConfig } from '@/types/coinconfig'

// ERC-20代币类型定义（基于CoinConfig）
interface ERC20Token {
  contract_address: string
  name: string
  symbol: string
  decimals: number
  balance?: string
  logo_url?: string
  is_verified?: boolean
  status?: number
}

// 下拉展示的授权地址最小结构
interface AuthorizedDropdownItem {
  address: string
  label: string
  type: string
  balance?: string
}

interface Props {
  show: boolean
  transaction?: any // 编辑模式下的交易数据
  isEditMode?: boolean // 是否为编辑模式
}

interface Emits {
  (e: 'close'): void
  (e: 'created', transaction: any): void
  (e: 'updated', transaction: any): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 表单数据
const form = ref<CreateUserTransactionRequest>({
  chain: 'eth',
  symbol: 'ETH',
  from_address: '',
  to_address: '',
  amount: '',
  fee: '0', // 创建时设为0，发送时再设置实际值
  gas_limit: undefined,
  gas_price: undefined,
  nonce: undefined, // Nonce自动获取
  remark: '',
  allowance_address: '' // 授权地址（代币持有者地址）
})

// 编辑模式下初始化表单数据
const initEditForm = () => {
  if (props.isEditMode && props.transaction) {
    const tx = props.transaction
    
    // 将整数金额转换为显示格式
    let displayAmount = ''
    if (tx.amount) {
      if (tx.transaction_type === 'token' && tx.token_decimals !== undefined) {
        // 代币：从最小单位转换为显示单位
        const intAmount = BigInt(tx.amount)
        const factor = BigInt(Math.pow(10, tx.token_decimals).toString())
        displayAmount = (Number(intAmount) / Number(factor)).toString()
      } else if (tx.symbol === 'ETH') {
        // ETH：从wei转换为ETH
        const intAmount = BigInt(tx.amount)
        const factor = BigInt('1000000000000000000') // 10^18
        displayAmount = (Number(intAmount) / Number(factor)).toString()
      } else {
        // 其他情况直接使用原始值
        displayAmount = tx.amount
      }
    }
    
    form.value = {
      chain: tx.chain || 'eth',
      symbol: tx.symbol || 'ETH',
      from_address: tx.from_address || '',
      to_address: tx.to_address || '',
      amount: displayAmount,
      fee: tx.fee || '0',
      gas_limit: tx.gas_limit,
      gas_price: tx.gas_price,
      nonce: tx.nonce,
      remark: tx.remark || '',
      transaction_type: tx.transaction_type,
      contract_operation_type: tx.contract_operation_type,
      token_contract_address: tx.token_contract_address,
      allowance_address: tx.allowance_address || ''
    }
    
    // 设置交易类型
    if (tx.transaction_type === 'token') {
      transactionType.value = 'erc20'
      // 设置代币相关信息
      if (tx.token_contract_address) {
        // 查找并设置选中的代币
        const token = tokens.value.find(t => t.contract_address === tx.token_contract_address)
        if (token) {
          selectedToken.value = token
        }
      }
      if (tx.contract_operation_type) {
        const op = tx.contract_operation_type as string
        contractOperationType.value = (op === 'transfer' || op === 'approve' || op === 'transferFrom') ? (op as any) : 'transfer'
      }
    }
  }
}

// 提交状态
const isSubmitting = ref(false)

// 交易类型
const transactionType = ref<'eth' | 'erc20'>('eth')

// 合约操作类型
const contractOperationType = ref<'transfer' | 'approve' | 'transferFrom'>('transfer')

// 地址相关
const addresses = ref<PersonalAddressItem[]>([])
const showFromAddressDropdown = ref(false)
const authorizedAddresses = ref<AuthorizedDropdownItem[]>([])
const showAllowanceAddressDropdown = ref(false)
const showToAddressDropdown = ref(false)

// 代币相关
const tokens = ref<ERC20Token[]>([])
const selectedTokenSearch = ref('')
const showTokenDropdown = ref(false)
const selectedToken = ref<ERC20Token | null>(null)

// 计算属性
const filteredFromAddresses = computed(() => {
  if (!form.value.from_address) return addresses.value
  return addresses.value.filter(addr => 
    addr.address.toLowerCase().includes(form.value.from_address.toLowerCase()) ||
    addr.label.toLowerCase().includes(form.value.from_address.toLowerCase())
  )
})

const filteredToAddresses = computed(() => {
  if (!form.value.to_address) return addresses.value
  return addresses.value.filter(addr => 
    addr.address.toLowerCase().includes(form.value.to_address.toLowerCase()) ||
    addr.label.toLowerCase().includes(form.value.to_address.toLowerCase())
  )
})

const isContractTransaction = computed(() => {
  // 检查是否为合约交易（接收地址是合约地址）
  const toAddress = addresses.value.find(addr => addr.address === form.value.to_address)
  return toAddress?.type === 'contract'
})

// 是否显示发送地址字段
const shouldShowFromAddress = computed(() => {
  if (transactionType.value === 'eth') return true
  // 查询余额时，发送地址字段作为查询地址显示
  return true
})

// 获取发送地址标签
const getFromAddressLabel = () => {
  if (transactionType.value === 'eth') return '发送地址'
  switch (contractOperationType.value) {
    case 'transfer': return '发送地址'
    case 'approve': return '授权者地址'
    case 'transferFrom': return '发送地址'
    default: return '发送地址'
  }
}

// 获取发送地址占位符
const getFromAddressPlaceholder = () => {
  if (transactionType.value === 'eth') return '选择或输入发送地址'
  switch (contractOperationType.value) {
    case 'transfer': return '选择或输入发送地址'
    case 'approve': return '选择或输入授权者地址'
    case 'transferFrom': return '选择或输入发送地址'
    default: return '选择或输入发送地址'
  }
}

// 是否显示接收地址字段
const shouldShowToAddress = computed(() => {
  if (transactionType.value === 'eth') return true
  return true
})

// 获取接收地址标签
const getToAddressLabel = () => {
  if (transactionType.value === 'eth') return '接收地址'
  switch (contractOperationType.value) {
    case 'transfer': return '接收地址'
    case 'approve': return '被授权者地址'
    case 'transferFrom': return '接收地址'
    default: return '接收地址'
  }
}

// 获取接收地址占位符
const getToAddressPlaceholder = () => {
  if (transactionType.value === 'eth') return '选择或输入接收地址'
  switch (contractOperationType.value) {
    case 'transfer': return '选择或输入接收地址'
    case 'approve': return '选择或输入被授权者地址'
    case 'transferFrom': return '选择或输入接收地址'
    default: return '选择或输入接收地址'
  }
}

// 是否显示金额字段
const shouldShowAmount = computed(() => {
  if (transactionType.value === 'eth') return true
  return true
})

// 获取金额标签
const getAmountLabel = () => {
  if (transactionType.value === 'eth') return '交易金额'
  switch (contractOperationType.value) {
    case 'transfer': return '转账金额'
    case 'approve': return '授权额度'
    case 'transferFrom': return '转账金额'
    default: return '交易金额'
  }
}

// 获取金额占位符
const getAmountPlaceholder = () => {
  if (transactionType.value === 'eth') return '0.0'
  switch (contractOperationType.value) {
    case 'transfer': return '0.0'
    case 'approve': return '0.0'
    case 'transferFrom': return '0.0'
    default: return '0.0'
  }
}

// 过滤代币列表
const filteredTokens = computed(() => {
  if (!selectedTokenSearch.value) return tokens.value
  return tokens.value.filter(token => 
    token.symbol.toLowerCase().includes(selectedTokenSearch.value.toLowerCase()) ||
    token.name.toLowerCase().includes(selectedTokenSearch.value.toLowerCase()) ||
    token.contract_address.toLowerCase().includes(selectedTokenSearch.value.toLowerCase())
  )
})

// 初始化ETH相关字段
const initEthFields = () => {
  form.value.gas_limit = 21000
  form.value.gas_price = '20'
  form.value.nonce = undefined // 自动获取
}

// 加载地址列表
const loadAddresses = async () => {
  try {
    const response = await getPersonalAddresses()
    if (response.success) {
      addresses.value = response.data || []
    }
  } catch (error) {
    console.error('加载地址列表失败:', error)
  }
}

// 加载代币列表
const loadTokens = async () => {
  try {
    // 调用币种配置API获取代币列表
    const response = await listCoinConfigs({
      chain: 'eth', // 只获取以太坊链的代币
      status: 1, // 只获取启用的代币
      page: 1,
      page_size: 100 // 获取足够多的代币
    })
    
    if (response.success && response.data) {
      // 检查响应数据结构，支持两种格式
      let coinConfigs: any[] = []
      
      if (Array.isArray(response.data)) {
        // 如果data直接是数组（PaginatedResponse<CoinConfig>格式）
        coinConfigs = response.data
      } else if (typeof response.data === 'object' && response.data !== null && 'coin_configs' in response.data) {
        // 如果data包含coin_configs数组（后端实际返回格式）
        const dataWithCoinConfigs = response.data as { coin_configs: any[] }
        if (Array.isArray(dataWithCoinConfigs.coin_configs)) {
          coinConfigs = dataWithCoinConfigs.coin_configs
        }
      }
      
      if (coinConfigs.length > 0) {
        // 将CoinConfig转换为ERC20Token格式
        tokens.value = coinConfigs.map(coin => ({
          contract_address: coin.contract_addr,
          name: coin.name,
          symbol: coin.symbol,
          decimals: coin.decimals,
          logo_url: coin.logo_url,
          is_verified: coin.is_verified,
          status: coin.status,
          balance: '0' // 余额暂时设为0，后续可以从用户代币余额API获取
        }))
      }
    }
  } catch (error) {
    console.error('加载代币列表失败:', error)
    // 如果API调用失败，使用一些常见的代币作为备用
    tokens.value = [
      {
        contract_address: '0xdAC17F958D2ee523a2206206994597C13D831ec7', // USDT
        name: 'Tether USD',
        symbol: 'USDT',
        decimals: 6,
        balance: '0'
      },
      {
        contract_address: '0xA0b86a33E6441b8c4C8C8C8C8C8C8C8C8C8C8C8', // USDC
        name: 'USD Coin',
        symbol: 'USDC',
        decimals: 6,
        balance: '0'
      }
    ]
  }
}

// 选择发送地址
const selectFromAddress = (address: PersonalAddressItem) => {
  form.value.from_address = address.address
  showFromAddressDropdown.value = false
  
  // 如果是transferFrom操作，查询授权关系
  if (contractOperationType.value === 'transferFrom') {
    loadAuthorizedAddresses(address.address)
  }
}

// 派生授权地址：从已加载的地址列表中找到与 from_address 匹配的地址，读取其 authorized_addresses
const loadAuthorizedAddresses = (spenderAddress: string) => {
  try {
    const owner = addresses.value.find(a => a.address.toLowerCase() === spenderAddress.toLowerCase())
    if (!owner || !owner.authorized_addresses) {
      authorizedAddresses.value = []
      return
    }
    const authMap = owner.authorized_addresses as Record<string, any>
    const authAddrs = Object.keys(authMap)
    // 将授权地址映射为下拉展示所需结构（若能在已加载地址中找到匹配项则带上label等，否则降级显示）
    const mapped = authAddrs.map(addr => {
      const found = addresses.value.find(a => a.address.toLowerCase() === addr.toLowerCase() && a.type === 'contract')
      if (found) {
        return { address: found.address, label: found.label || found.address, type: found.type, balance: found.balance }
      }
      return { address: addr, label: addr, type: 'wallet', balance: '0' }
    })
    authorizedAddresses.value = mapped
  } catch (e) {
    console.error('派生授权地址失败:', e)
    authorizedAddresses.value = []
  }
}

// 选择授权地址
const selectAllowanceAddress = (item: AuthorizedDropdownItem) => {
  form.value.allowance_address = item.address
  showAllowanceAddressDropdown.value = false
}

// 选择接收地址
const selectToAddress = (address: PersonalAddressItem) => {
  form.value.to_address = address.address
  showToAddressDropdown.value = false
}

// 处理发送地址输入框失去焦点
const handleFromAddressBlur = () => {
  // 延迟隐藏，让用户有时间点击下拉选项
  setTimeout(() => {
    showFromAddressDropdown.value = false
  }, 200)
}

// 监听发送地址变化，动态刷新授权地址下拉选项
watch(() => form.value.from_address, (newVal) => {
  if (contractOperationType.value === 'transferFrom' && newVal) {
    loadAuthorizedAddresses(newVal)
  } else {
    authorizedAddresses.value = []
  }
})

// 处理授权地址输入框失去焦点
const handleAllowanceAddressBlur = () => {
  // 延迟隐藏，让用户有时间点击下拉选项
  setTimeout(() => {
    showAllowanceAddressDropdown.value = false
  }, 200)
}

// 处理接收地址输入框失去焦点
const handleToAddressBlur = () => {
  // 延迟隐藏，让用户有时间点击下拉选项
  setTimeout(() => {
    showToAddressDropdown.value = false
  }, 200)
}

// 处理交易类型变化
const handleTransactionTypeChange = () => {
  if (transactionType.value === 'eth') {
    form.value.symbol = 'ETH'
    selectedToken.value = null
    contractOperationType.value = 'transfer'
  } else {
    // 保持当前选择的代币符号，不重置
    if (selectedToken.value) {
      form.value.symbol = selectedToken.value.symbol
    }
    contractOperationType.value = 'transfer'
  }
}

// 处理合约操作类型变化
const handleContractOperationTypeChange = () => {
  // 根据操作类型调整表单字段
  switch (contractOperationType.value) {
    case 'transfer':
      // 转账：需要发送地址、接收地址、金额
      // 确保接收地址字段可见
      break
    case 'approve':
      // 授权：需要授权地址、金额
      // 确保接收地址字段可见
      break
    case 'transferFrom':
      // 授权转账：需要发送地址、接收地址、金额
      // 确保接收地址字段可见
      // 清空授权地址，需要重新选择
      form.value.allowance_address = ''
      authorizedAddresses.value = []
      // 如果已有发送地址，查询授权关系
      if (form.value.from_address) {
        loadAuthorizedAddresses(form.value.from_address)
      }
      break
    
  }
}

// 选择代币
const selectToken = (token: ERC20Token) => {
  selectedToken.value = token
  form.value.symbol = token.symbol
  selectedTokenSearch.value = token.symbol
  showTokenDropdown.value = false
}

// 处理代币下拉框失去焦点
const handleTokenDropdownBlur = () => {
  setTimeout(() => {
    showTokenDropdown.value = false
  }, 200)
}

// 格式化代币余额
const formatTokenBalance = (token: ERC20Token) => {
  if (!token.balance) return '0'
  const num = parseFloat(token.balance)
  if (isNaN(num)) return '0'
  return num.toFixed(6)
}

// 格式化余额显示
const formatBalance = (balance: string | undefined) => {
  if (!balance) return '0'
  const num = parseFloat(balance)
  if (isNaN(num)) return '0'
  return num.toFixed(6)
}

// 转换为wei单位 - 处理整数金额
const formatToWei = (amount: string) => {
  if (!amount) return '0'
  const num = parseFloat(amount)
  if (isNaN(num)) return '0'
  // 1 ETH = 10^18 wei，返回整数格式
  const wei = Math.floor(num * Math.pow(10, 18))
  return wei.toString()
}

// 转换为代币最小单位 - 处理整数金额
const formatToTokenUnits = (amount: string, decimals: number) => {
  if (!amount) return '0'
  const num = parseFloat(amount)
  if (isNaN(num)) return '0'
  // 转换为代币的最小单位，返回整数格式
  const units = Math.floor(num * Math.pow(10, decimals))
  return units.toString()
}

// 提交表单
const handleSubmit = async () => {
  try {
    isSubmitting.value = true
    
    // 验证必填字段
    if (!form.value.symbol) {
      alert('请选择代币')
      return
    }
    
    if (!form.value.from_address) {
      alert('请选择发送地址')
      return
    }
    
    if (transactionType.value === 'erc20' && !selectedToken.value) {
      alert('请选择代币')
      return
    }
    
    // 验证接收地址与金额
    if (!form.value.to_address) {
      alert('请选择接收地址')
      return
    }

    if (!form.value.amount) {
      alert('请输入交易金额')
      return
    }
    
    // transferFrom操作需要授权地址
    if (contractOperationType.value === 'transferFrom' && !form.value.allowance_address) {
      alert('请选择代币持有者地址')
      return
    }
    
    // 构建完整的提交数据
    const submitData = {
      ...form.value,
      // 添加交易类型和合约操作类型
      transaction_type: transactionType.value === 'eth' ? 'coin' : 'token',
      contract_operation_type: contractOperationType.value,
      // 如果是代币交易，添加代币合约地址
      token_contract_address: selectedToken.value?.contract_address || '',
      // 确保金额为整数格式
      amount: form.value.amount ? (transactionType.value === 'eth' ? formatToWei(form.value.amount) : formatToTokenUnits(form.value.amount, selectedToken.value?.decimals || 18)) : '0',
      // 确保手续费为整数格式
      fee: form.value.fee || '0'
    }
    
    console.log('提交的交易数据:', submitData)
    
    let response
    if (props.isEditMode && props.transaction) {
      // 编辑模式：更新交易
      response = await updateUserTransaction(props.transaction.id, submitData)
    } else {
      // 创建模式：创建新交易
      response = await createUserTransaction(submitData)
    }
    
    if (response.success) {
      // 操作成功，触发事件
      if (props.isEditMode) {
        emit('updated', response.data)
      } else {
        emit('created', response.data)
      }
      emit('close')
      
      // 重置表单
      resetForm()
    } else {
      alert((props.isEditMode ? '更新' : '创建') + '交易失败: ' + response.message)
    }
  } catch (error) {
    console.error((props.isEditMode ? '更新' : '创建') + '交易失败:', error)
    alert((props.isEditMode ? '更新' : '创建') + '交易失败，请重试')
  } finally {
    isSubmitting.value = false
  }
}

// 重置表单
const resetForm = () => {
  form.value = {
    chain: 'eth',
    symbol: 'ETH',
    from_address: '',
    to_address: '',
    amount: '',
    fee: '0',
    gas_limit: 21000,
    gas_price: '20',
    nonce: undefined,
    remark: '',
    allowance_address: ''
  }
  
  // 重置所有状态
  transactionType.value = 'eth'
  contractOperationType.value = 'transfer'
  selectedToken.value = null
  selectedTokenSearch.value = ''
  authorizedAddresses.value = []
  
  // 重置下拉框状态
  showFromAddressDropdown.value = false
  showAllowanceAddressDropdown.value = false
  showToAddressDropdown.value = false
  showTokenDropdown.value = false
  
  // 重新初始化ETH字段
  initEthFields()
}

// 监听模态框显示状态，每次打开时重置表单
watch(() => props.show, (newShow) => {
  if (newShow) {
    // 模态框打开时重置表单
    resetForm()
    // 如果是编辑模式，重新初始化编辑数据
    if (props.isEditMode && props.transaction) {
      initEditForm()
    }
  }
})

// 组件挂载时加载地址列表和初始化ETH字段
onMounted(() => {
  loadAddresses()
  loadTokens()
  initEthFields()
  // 如果初始就是编辑模式，初始化编辑数据
  if (props.isEditMode && props.transaction) {
    initEditForm()
  }
})
</script>
