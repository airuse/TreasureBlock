package main

import (
	"blockChainBrowser/client/signer/pkg"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== BTC QR码解析测试 ===")

	// 读取测试数据
	data, err := os.ReadFile("test_btc_qr.json")
	if err != nil {
		fmt.Printf("❌ 读取测试文件失败: %v\n", err)
		return
	}

	qrData := string(data)
	fmt.Printf("QR码数据: %s\n", qrData)

	// 解析QR码数据
	transaction, err := pkg.ParseQRCodeData(qrData)
	if err != nil {
		fmt.Printf("❌ QR码数据解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ QR码数据解析成功\n")
	fmt.Printf("交易ID: %d\n", transaction.ID)
	fmt.Printf("链类型: %s\n", transaction.Type)
	fmt.Printf("发送地址: %s\n", transaction.Address)
	fmt.Printf("是否为BTC交易: %t\n", transaction.IsBTC())

	if transaction.MsgTx != nil {
		fmt.Printf("交易版本: %d\n", transaction.MsgTx.Version)
		fmt.Printf("锁定时间: %d\n", transaction.MsgTx.LockTime)
		fmt.Printf("交易输入数量: %d\n", len(transaction.MsgTx.TxIn))
		fmt.Printf("交易输出数量: %d\n", len(transaction.MsgTx.TxOut))

		fmt.Println("\n交易输入详情:")
		for i, txIn := range transaction.MsgTx.TxIn {
			fmt.Printf("  %d. 前交易: %s, 输出索引: %d, 序列号: %d\n",
				i+1, txIn.Txid, txIn.Vout, txIn.Sequence)
		}

		fmt.Println("\n交易输出详情:")
		totalOutput := int64(0)
		for i, txOut := range transaction.MsgTx.TxOut {
			fmt.Printf("  %d. 地址: %s, 金额: %d satoshi (%.8f BTC)\n",
				i+1, txOut.Address, txOut.ValueSatoshi, float64(txOut.ValueSatoshi)/1e8)
			totalOutput += txOut.ValueSatoshi
		}
		fmt.Printf("\n总输出金额: %d satoshi (%.8f BTC)\n", totalOutput, float64(totalOutput)/1e8)
	}

	// 测试JSON序列化
	jsonData, err := json.MarshalIndent(transaction, "", "  ")
	if err != nil {
		fmt.Printf("❌ JSON序列化失败: %v\n", err)
		return
	}

	fmt.Println("\n=== 序列化后的交易数据 ===")
	fmt.Println(string(jsonData))

	fmt.Println("\n✅ 测试完成")
}
