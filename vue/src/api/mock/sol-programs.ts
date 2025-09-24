export const handleMockCreateProgram = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        success: true,
        message: 'mock create program ok',
        data: { id: Date.now(), ...data },
        timestamp: Date.now()
      })
    }, 300)
  })
}

export const handleMockUpdateProgram = (id: number, data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        success: true,
        message: 'mock update program ok',
        data: { id, ...data },
        timestamp: Date.now()
      })
    }, 300)
  })
}

export const handleMockDeleteProgram = (id: number): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        success: true,
        message: 'mock delete program ok',
        data: null,
        timestamp: Date.now()
      })
    }, 300)
  })
}

export const handleMockGetProgram = (id: number): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        success: true,
        message: 'mock get program ok',
        data: {
          id,
          program_id: 'Mock1111111111111111111111111111111111111',
          name: 'Mock Program',
          is_system: false,
          status: 'active'
        },
        timestamp: Date.now()
      })
    }, 300)
  })
}

export const handleMockListPrograms = (params: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        success: true,
        message: 'mock list programs ok',
        data: [
          { id: 1, program_id: 'Mock1111111111111111111111111111111111111', name: 'Mock Program A', status: 'active' },
          { id: 2, program_id: 'Mock2222222222222222222222222222222222222', name: 'Mock Program B', status: 'inactive' }
        ],
        pagination: { page: params?.page ?? 1, page_size: params?.page_size ?? 20, total: 2 },
        timestamp: Date.now()
      })
    }, 300)
  })
}


