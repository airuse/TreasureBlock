export const handleMockListTxDetails = (params: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        success: true,
        data: [
          {
            id: 1,
            tx_id: 'mockTx111',
            slot: 100,
            blockhash: 'mockBlockhash',
            recent_blockhash: 'mockRecent',
            version: 'legacy',
            fee: 5000,
            compute_units: 100000,
            status: 'success',
            account_keys: '[]',
            pre_balances: '[]',
            post_balances: '[]',
            pre_token_balances: '[]',
            post_token_balances: '[]',
            logs: '[]',
            instructions: '[]',
            inner_instructions: '[]',
            loaded_addresses: '[]',
            rewards: '[]',
            events: '[]',
            raw_transaction: '{}',
            raw_meta: '{}',
            ctime: new Date().toISOString(),
            mtime: new Date().toISOString()
          }
        ],
        message: 'mock list tx details ok',
        timestamp: Date.now()
      })
    }, 300)
  })
}

export const handleMockGetArtifactsByTxId = (txId: string): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        success: true,
        data: {
          events: [
            { id: 1, tx_id: txId, slot: 100, event_index: 0, event_type: 'transfer', program_id: 'Mock111', from_address: 'A', to_address: 'B', amount: '10', mint: 'Mint111', decimals: 9, is_inner: false, asset_type: 'token', extra_data: '{}', ctime: new Date().toISOString() }
          ],
          instructions: [
            { id: 1, tx_id: txId, slot: 100, instruction_index: 0, program_id: 'Mock111', accounts: '[]', data: '', parsed_data: '{}', instruction_type: 'transfer', is_inner: false, stack_height: 1, ctime: new Date().toISOString() }
          ]
        },
        message: 'mock get artifacts ok',
        timestamp: Date.now()
      })
    }, 300)
  })
}


