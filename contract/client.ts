// 导入 Solana 和 SPL Token 的相关工具
import { Connection, Keypair, clusterApiUrl } from "@solana/web3.js";
import { createMint, TOKEN_2022_PROGRAM_ID } from "@solana/spl-token";

// ----------------------------------------------------
// 主要操作逻辑
// ----------------------------------------------------

// (1) 连接到 Solana 的 Devnet (测试网)
const connection = new Connection(clusterApiUrl("devnet"), "confirmed");

// (2) 获取 Playground 提供的钱包作为付款方和权限方
// Playground 会自动处理这个钱包，我们直接用就行
const payerAndAuthority = pg.wallet.keypair;

// (3) 生成一个新的密钥对，它的公钥将作为我们铸币账户的地址
const mintKeypair = Keypair.generate();

console.log("准备创建铸币账户...");
console.log(`付款方地址: ${payerAndAuthority.publicKey.toBase58()}`);
console.log(`新铸币账户地址: ${mintKeypair.publicKey.toBase58()}`);

// (4) 调用`createMint`函数来发送交易
// 这个函数会帮我们创建并初始化好一个铸币账户
const mintAddress = await createMint(
  connection,
  payerAndAuthority, // 付款方
  payerAndAuthority.publicKey, // 铸币权限 (设置为我们自己)
  payerAndAuthority.publicKey, // 冻结权限 (也设置为我们自己)
  9, // 我们代币的小数位数 (这里设为9)
  mintKeypair, // 铸币账户的密钥对
  undefined, // 确认选项，默认即可
  TOKEN_2022_PROGRAM_ID // 使用最新的 Token-2022 标准
);

console.log("✅ 铸币账户创建成功!");
console.log(`✅ Mint Account Address: ${mintAddress.toBase58()}`);
