import request from '../request'
import type { ApiResponse } from '../types'

// 解析配置类型
export interface ParserConfig {
  id?: number
  contract_address: string
  parser_type: 'input_data' | 'event_log'
  function_signature: string
  function_name: string
  function_description: string
  param_config: any[]
  parser_rules: any
  display_format: string
  is_active: boolean
  priority: number
  // 日志解析配置字段
  logs_parser_type?: string
  event_signature?: string
  event_name?: string
  event_description?: string
  logs_param_config?: any[]
  logs_parser_rules?: any
  logs_display_format?: string
}

// 创建解析配置请求
export interface CreateParserConfigRequest extends ParserConfig {}

// 更新解析配置请求
export interface UpdateParserConfigRequest extends Partial<ParserConfig> {
  id: number
}

/**
 * 创建解析配置
 */
export function createParserConfig(data: CreateParserConfigRequest): Promise<ApiResponse<ParserConfig>> {
  console.log('🌐 创建解析配置:', data)
  return request({
    url: '/api/v1/parser-configs',
    method: 'POST',
    data
  })
}

/**
 * 更新解析配置
 */
export function updateParserConfig(data: UpdateParserConfigRequest): Promise<ApiResponse<ParserConfig>> {
  console.log('🌐 更新解析配置:', data)
  return request({
    url: `/api/v1/parser-configs/${data.id}`,
    method: 'PUT',
    data
  })
}

/**
 * 删除解析配置
 */
export function deleteParserConfig(id: number): Promise<ApiResponse<null>> {
  console.log('🌐 删除解析配置:', id)
  return request({
    url: `/api/v1/parser-configs/${id}`,
    method: 'DELETE'
  })
}

/**
 * 批量保存解析配置（创建或更新）
 */
export function batchSaveParserConfigs(contractAddress: string, configs: ParserConfig[]): Promise<ApiResponse<any>> {
  console.log('🌐 批量保存解析配置:', { contractAddress, configs })
  
  const promises = configs.map(config => {
    if (config.id) {
      // 如果有ID，说明是更新 - 只传递需要更新的字段，不包含ID和contract_address
      const updateData: UpdateParserConfigRequest = {
        id: config.id, // ID是必需的，用于路由
        parser_type: config.parser_type,
        function_signature: config.function_signature,
        function_name: config.function_name,
        function_description: config.function_description,
        param_config: config.param_config,
        parser_rules: config.parser_rules,
        display_format: config.display_format,
        is_active: config.is_active,
        priority: config.priority,
        // 日志解析配置字段
        logs_parser_type: config.logs_parser_type,
        event_signature: config.event_signature,
        event_name: config.event_name,
        event_description: config.event_description,
        logs_param_config: config.logs_param_config,
        logs_parser_rules: config.logs_parser_rules,
        logs_display_format: config.logs_display_format
      }
      console.log('🔄 更新解析配置，数据:', updateData)
      return updateParserConfig(updateData)
    } else {
      // 如果没有ID，说明是创建
      console.log('🆕 创建新解析配置，数据:', config)
      return createParserConfig(config)
    }
  })
  
  return Promise.all(promises).then(results => {
    const success = results.every(result => result.success)
    return {
      success,
      data: results.map(result => result.data),
      message: success ? '所有解析配置保存成功' : '部分解析配置保存失败'
    }
  })
}
