// src/App.tsx
import React, { useState } from 'react';
import { useConnection, useWallet } from '@solana/wallet-adapter-react';
import { WalletMultiButton } from '@solana/wallet-adapter-react-ui';
import { Keypair, SystemProgram, Transaction, PublicKey } from '@solana/web3.js';
import {
    MINT_SIZE,
    TOKEN_PROGRAM_ID,
    getMinimumBalanceForRentExemptMint,
    createInitializeMintInstruction,
    getAssociatedTokenAddressSync,
    createAssociatedTokenAccountInstruction,
    createMintToInstruction,
} from '@solana/spl-token';

interface TokenBalance {
    mintAddress: string;
    balance: string;
}

function App() {
    const { connection } = useConnection();
    const { publicKey, sendTransaction } = useWallet();
    
    // State for the new minting functionality
    const [mintAddress, setMintAddress] = useState('');
    const [mintAmount, setMintAmount] = useState('');

    const [tokenBalances, setTokenBalances] = useState<TokenBalance[] | null>(null);

    const createMint = async () => {
        if (!publicKey) { return; }
        const mintKeypair = Keypair.generate();
        try {
            const lamports = await getMinimumBalanceForRentExemptMint(connection);
            const transaction = new Transaction().add(
                SystemProgram.createAccount({ fromPubkey: publicKey, newAccountPubkey: mintKeypair.publicKey, space: MINT_SIZE, lamports, programId: TOKEN_PROGRAM_ID }),
                createInitializeMintInstruction(mintKeypair.publicKey, 9, publicKey, publicKey)
            );
            const signature = await sendTransaction(transaction, connection, { signers: [mintKeypair] });
            await connection.confirmTransaction(signature, 'confirmed');
            const newMintAddress = mintKeypair.publicKey.toBase58();
            setMintAddress(newMintAddress); // Automatically fill the input with the new mint address
            alert(`创建成功! 铸币地址: ${newMintAddress}`);
        } catch (error) {
            console.error('❌ 创建失败:', error);
            alert(`❌ 创建失败: ${(error as Error).message}`);
        }
    };
    
    // --- NEW FUNCTION: Mint Tokens ---
    const mintTokens = async () => {
        if (!publicKey || !mintAddress || !mintAmount) {
            alert('请连接钱包, 并输入铸币地址和数量!');
            return;
        }

        console.log("🚀 开始发行代币...");
        
        try {
            const mintPublicKey = new PublicKey(mintAddress);
            const decimals = 9; // 我们在创建时设置了9位小数

            // 1. 找到或创建关联代币账户 (ATA)
            // 这是接收代币的标准地址
            const destinationAta = getAssociatedTokenAddressSync(mintPublicKey, publicKey);
            console.log(`✅ 目标代币账户 (ATA): ${destinationAta.toBase58()}`);

            const transaction = new Transaction();
            
            // 检查ATA是否存在，如果不存在则添加创建指令
            const destinationAtaInfo = await connection.getAccountInfo(destinationAta);
            if (!destinationAtaInfo) {
                console.log("ℹ️ 目标代币账户不存在，将创建它...");
                transaction.add(
                    createAssociatedTokenAccountInstruction(
                        publicKey, // Payer
                        destinationAta, // ATA Address
                        publicKey, // Owner
                        mintPublicKey // Mint
                    )
                );
            }

            // 2. 添加发行代币的指令 (MintTo)
            const amountInSmallestUnit = BigInt(parseFloat(mintAmount) * Math.pow(10, decimals));
            transaction.add(
                createMintToInstruction(
                    mintPublicKey, // Mint Address
                    destinationAta, // Destination Token Account
                    publicKey, // Mint Authority (YOU!)
                    amountInSmallestUnit // Amount, in the smallest unit
                )
            );

            // 3. 发送交易
            const signature = await sendTransaction(transaction, connection);
            console.log("✍️ 交易已发送，等待确认...");
            await connection.confirmTransaction(signature, 'confirmed');

            console.log("✅ 发行成功!");
            console.log(`🔗 交易链接: https://explorer.solana.com/tx/${signature}?cluster=devnet`);
            alert(`成功发行 ${mintAmount} 个代币!`);

            // 发行成功后自动刷新余额
            checkBalances();

        } catch (error) {
            console.error('❌ 发行失败:', error);
            alert(`❌ 发行失败: ${(error as Error).message}`);
        }
    };


    const checkBalances = async () => {
        if (!publicKey) { return; }
        setTokenBalances(null);
        try {
            const tokenAccounts = await connection.getParsedTokenAccountsByOwner(publicKey, { programId: TOKEN_PROGRAM_ID });
            if (tokenAccounts.value.length === 0) {
                setTokenBalances([]);
                return;
            }
            const balances: TokenBalance[] = [];
            tokenAccounts.value.forEach((account) => {
                const info = account.account.data.parsed.info;
                if (info.tokenAmount.uiAmount > 0) {
                    balances.push({ mintAddress: info.mint, balance: info.tokenAmount.uiAmountString });
                }
            });
            setTokenBalances(balances);
        } catch (error) {
            console.error('❌ 查询余额失败:', error);
        }
    };

    return (
        <div style={{ padding: '20px', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '20px', textAlign: 'center' }}>
            <h1>Solana DApp 工具箱</h1>
            <WalletMultiButton />
            
            {publicKey && (
                <div style={{ border: '1px solid #ccc', padding: '20px', borderRadius: '8px', width: '800px' }}>
                    <h2>第一步：创建代币种类</h2>
                    <button onClick={createMint}>点我创建一个新的铸币地址</button>
                </div>
            )}
            
            {publicKey && (
                <div style={{ border: '1px solid #ccc', padding: '20px', borderRadius: '8px', width: '800px' }}>
                    <h2>第二步：为自己发行代币</h2>
                    <div>
                        <label>铸币地址 (Mint Address): </label>
                        <input
                            type="text"
                            value={mintAddress}
                            onChange={(e) => setMintAddress(e.target.value)}
                            placeholder="请粘贴你刚创建的铸币地址"
                            style={{ width: '400px', margin: '10px' }}
                        />
                    </div>
                    <div>
                        <label>发行数量 (Amount): </label>
                        <input
                            type="number"
                            value={mintAmount}
                            onChange={(e) => setMintAmount(e.target.value)}
                            placeholder="例如: 100"
                            style={{ width: '400px', margin: '10px' }}
                        />
                    </div>
                    <button onClick={mintTokens}>发行代币到我的钱包</button>
                </div>
            )}

            {publicKey && (
                <div style={{ border: '1px solid #ccc', padding: '20px', borderRadius: '8px', width: '800px' }}>
                    <h2>第三步：查询我的代币余额</h2>
                    <button onClick={checkBalances}>刷新我的代币余额</button>
                    {tokenBalances && (
                        <div>
                          {tokenBalances.length === 0 ? <p>你没有任何持有余额的代币。</p> : (
                            <table border={1} style={{ margin: '20px auto', borderCollapse: 'collapse', width: '100%' }}>
                              <thead><tr><th style={{padding:'8px'}}>代币种类 (Mint Address)</th><th style={{padding:'8px'}}>余额</th></tr></thead>
                              <tbody>{tokenBalances.map((t) => (<tr key={t.mintAddress}>
                                  <td style={{padding:'8px', fontFamily:'monospace'}}>{t.mintAddress}</td>
                                  <td style={{padding:'8px', textAlign:'right'}}>{t.balance}</td>
                                </tr>))}</tbody>
                            </table>
                          )}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

export default App;