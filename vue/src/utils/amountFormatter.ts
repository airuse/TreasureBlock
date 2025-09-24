/**
 * 金额格式化工具函数
 * 使用BigInt确保精确计算，避免JavaScript Number的精度问题
 */

/**
 * 格式化代币金额
 * @param amount 原始金额字符串（整数格式）
 * @param symbol 代币符号
 * @param decimals 代币精度
 * @returns 格式化后的金额字符串
 */
export function formatTokenAmount(amount: string, symbol: string, decimals: number | undefined): string {
  if (!amount || amount === '0') return '0'
  
  // 检查是否是小数，如果是小数，直接返回（可能是旧数据或显示格式）
  if (amount.includes('.')) {
    return amount
  }
  
  // 将字符串转换为BigInt（因为数据库中存储的是整数）
  let intAmount: bigint
  try {
    intAmount = BigInt(amount)
  } catch (error) {
    console.error(`无法转换金额为BigInt: ${amount}`, error)
    return amount // 如果转换失败，返回原始值
  }
  
  if (intAmount === 0n) return '0'
  
  // 如果明确提供了精度，使用提供的精度
  if (decimals !== undefined && decimals >= 0) {
    return formatAmountWithDecimals(intAmount, decimals)
  }
  
  // 如果没有提供精度，根据币种智能判断
  if (symbol === 'ETH' || symbol === 'BNB') {
    // ETH/BNB 使用18位精度
    return formatAmountWithDecimals(intAmount, 18)
  } else if (symbol === 'SOL') {
    // SOL 使用9位精度
    return formatAmountWithDecimals(intAmount, 9)
  } else if (symbol === 'USDC' || symbol === 'USDT') {
    // USDC/USDT使用6位精度
    return formatAmountWithDecimals(intAmount, 6)
  } else if (symbol === 'DAI') {
    // DAI使用18位精度
    return formatAmountWithDecimals(intAmount, 18)
  } else {
    // 其他代币，尝试智能判断精度
    // 如果数值很大，可能是原始精度，需要转换
    if (intAmount > BigInt('1000000000000')) { // 10^12
      // 尝试常见的精度：6, 8, 18
      const possibleDecimals = [6, 8, 18]
      for (const dec of possibleDecimals) {
        const formatted = formatAmountWithDecimals(intAmount, dec)
        const readableAmount = parseFloat(formatted)
        // 如果转换后的数值在合理范围内（0.000001 到 1000000），使用这个精度
        if (readableAmount >= 0.000001 && readableAmount <= 1000000) {
          return formatted
        }
      }
    }
    
    // 如果无法确定，直接返回原始值
    return amount
  }
}

/**
 * 使用指定精度格式化金额
 * @param intAmount BigInt格式的金额
 * @param decimals 精度位数
 * @returns 格式化后的金额字符串
 */
function formatAmountWithDecimals(intAmount: bigint, decimals: number): string {
  const factor = BigInt(Math.pow(10, decimals).toString())
  
  // 使用BigInt进行精确计算，避免精度丢失
  const quotient = intAmount / factor
  const remainder = intAmount % factor
  
  // 将余数转换为小数部分
  const remainderStr = remainder.toString().padStart(decimals, '0')
  const decimalPart = remainderStr.slice(0, Math.min(decimals, 8))
  
  // 组合整数部分和小数部分
  let result: string
  if (quotient === 0n) {
    result = '0.' + decimalPart
  } else {
    result = quotient.toString() + (decimalPart ? '.' + decimalPart : '')
  }
  
  // 去掉末尾的零和小数点
  result = result.replace(/0+$/, '') // 去掉末尾的零
  if (result.endsWith('.')) {
    result = result.slice(0, -1) // 去掉末尾的小数点
  }
  
  return result
}

/**
 * 将用户输入的金额转换为代币最小单位（整数格式）
 * @param amount 用户输入的金额字符串
 * @param decimals 代币精度
 * @returns 代币最小单位的字符串
 */
export function convertToTokenUnits(amount: string, decimals: number): string {
  if (!amount) return '0'
  
  const num = parseFloat(amount)
  if (isNaN(num)) return '0'
  
  // 使用BigInt确保精确计算
  const factor = BigInt(Math.pow(10, decimals).toString())
  const amountBig = BigInt(Math.floor(num * Math.pow(10, decimals)).toString())
  
  return amountBig.toString()
}

/**
 * 将代币最小单位转换为用户可读的金额
 * @param units 代币最小单位字符串
 * @param decimals 代币精度
 * @returns 用户可读的金额字符串
 */
export function convertFromTokenUnits(units: string, decimals: number): string {
  if (!units || units === '0') return '0'
  
  try {
    const intAmount = BigInt(units)
    return formatAmountWithDecimals(intAmount, decimals)
  } catch (error) {
    console.error(`无法转换代币单位: ${units}`, error)
    return units
  }
}
