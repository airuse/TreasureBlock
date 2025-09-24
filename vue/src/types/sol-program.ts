// ==================== Sol Program 业务实体类型 ====================

export interface SolProgram {
  id?: number
  program_id: string
  name: string
  alias?: string
  category?: string
  type?: string
  is_system?: boolean
  version?: string
  status?: string
  description?: string
  instruction_rules?: any
  event_rules?: any
  sample_data?: any
  ctime?: string
  mtime?: string
}

export interface ListProgramsRequest {
  page: number
  page_size: number
  keyword?: string
}


