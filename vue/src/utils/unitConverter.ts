// 单位转换工具函数

/**
 * 将Wei转换为Gwei
 * @param weiStr Wei字符串
 * @returns Gwei字符串
 */
export const convertWeiToGwei = (weiStr: string): string => {
  if (!weiStr || weiStr === '0') return '0'
  
  try {
    const weiBig = BigInt(weiStr)
    // 使用BigInt进行精确计算，避免科学计数法
    const gweiBig = weiBig / BigInt(1e9)
    const remainder = weiBig % BigInt(1e9)
    
    // 如果有余数，计算小数部分
    if (remainder > 0) {
      // 将余数转换为字符串，然后手动计算小数部分
      const remainderStr = remainder.toString()
      const paddedRemainder = remainderStr.padStart(9, '0')
      const decimalPart = '0.' + paddedRemainder
      
      // 完全避免使用Number()进行小数运算
      const gweiStr = gweiBig.toString()
      if (gweiStr === '0') {
        // 如果整数部分是0，直接返回小数部分
        return decimalPart.replace(/\.?0+$/, '') || '0'
      } else {
        // 如果整数部分不为0，拼接整数和小数部分
        return gweiStr + '.' + paddedRemainder.replace(/0+$/, '')
      }
    } else {
      return gweiBig.toString()
    }
  } catch (error) {
    console.error('Wei转Gwei失败:', error)
    return '0'
  }
}

/**
 * 将Gwei转换为Wei
 * @param gweiStr Gwei字符串
 * @returns Wei字符串
 */
export const convertGweiToWei = (gweiStr: string): string => {
  if (!gweiStr || gweiStr === '0') return '0'
  
  try {
    const gweiNum = parseFloat(gweiStr)
    const wei = Math.floor(gweiNum * 1e9)
    return wei.toString()
  } catch (error) {
    console.error('Gwei转Wei失败:', error)
    return '0'
  }
}

/**
 * 格式化费率显示（Wei转Gwei，保留2位小数）
 * @param feeWei Wei字符串
 * @returns 格式化的Gwei字符串
 */
export const formatFeeForDisplay = (feeWei: string): string => {
  if (!feeWei) return '0'
  
  try {
    const weiBig = BigInt(feeWei)
    // 使用BigInt进行精确计算，避免科学计数法
    const gweiBig = weiBig / BigInt(1e9)
    const remainder = weiBig % BigInt(1e9)
    
    // 如果有余数，计算小数部分
    if (remainder > 0) {
      // 将余数转换为字符串，然后手动计算小数部分
      const remainderStr = remainder.toString()
      const paddedRemainder = remainderStr.padStart(9, '0')
      
      // 完全避免使用Number()进行小数运算
      const gweiStr = gweiBig.toString()
      if (gweiStr === '0') {
        // 如果整数部分是0，直接返回小数部分
        const decimalPart = '0.' + paddedRemainder
        return parseFloat(decimalPart).toFixed(9)
      } else {
        // 如果整数部分不为0，拼接整数和小数部分
        const fullDecimal = gweiStr + '.' + paddedRemainder
        return parseFloat(fullDecimal).toFixed(9)
      }
    } else {
      return Number(gweiBig).toFixed(9)
    }
  } catch (error) {
    console.error('费率格式化失败:', error)
    return '0'
  }
}

// 测试用例
export const testUnitConversion = () => {
  console.log('=== 单位转换测试 ===')
  
  // 测试Wei到Gwei转换
  const testCases = [
    { wei: '2000000000', expected: '2' },      // 2 Gwei
    { wei: '30000000000', expected: '30' },    // 30 Gwei
    { wei: '3140820000000000', expected: '3140.82' }, // 3140.82 Gwei
    { wei: '5456811000000000', expected: '5456.811' }, // 5456.811 Gwei
    { wei: '1', expected: '0.000000001' },     // 1 Wei = 0.000000001 Gwei
    { wei: '1000000000', expected: '1' },      // 1 Gwei
    { wei: '1500000000', expected: '1.5' },    // 1.5 Gwei
  ]
  
  testCases.forEach(({ wei, expected }) => {
    const result = convertWeiToGwei(wei)
    const isCorrect = Math.abs(parseFloat(result) - parseFloat(expected)) < 0.000000001
    console.log(`Wei: ${wei} -> Gwei: ${result} (期望: ${expected}) ${isCorrect ? '✅' : '❌'}`)
  })
  
  // 测试Gwei到Wei转换
  const gweiTestCases = [
    { gwei: '2', expected: '2000000000' },
    { gwei: '30', expected: '30000000000' },
    { gwei: '3140.82', expected: '3140820000000000' },
    { gwei: '5456.811', expected: '5456811000000000' },
    { gwei: '0.000000001', expected: '1' },
    { gwei: '1', expected: '1000000000' },
    { gwei: '1.5', expected: '1500000000' },
  ]
  
  gweiTestCases.forEach(({ gwei, expected }) => {
    const result = convertGweiToWei(gwei)
    const isCorrect = result === expected
    console.log(`Gwei: ${gwei} -> Wei: ${result} (期望: ${expected}) ${isCorrect ? '✅' : '❌'}`)
  })
  
  // 测试科学计数法问题
  console.log('\n=== 科学计数法测试 ===')
  const scientificTestCases = [
    '1',           // 1 Wei
    '1000000000',  // 1 Gwei
    '2000000000',  // 2 Gwei
    '30000000000', // 30 Gwei
  ]
  
  scientificTestCases.forEach(wei => {
    const result = convertWeiToGwei(wei)
    const hasScientificNotation = result.includes('e') || result.includes('E')
    console.log(`Wei: ${wei} -> Gwei: ${result} ${hasScientificNotation ? '❌ 科学计数法' : '✅ 正常格式'}`)
  })
}
