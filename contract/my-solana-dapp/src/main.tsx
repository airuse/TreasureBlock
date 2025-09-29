import React, { useMemo } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import './index.css'

import { ConnectionProvider, WalletProvider } from '@solana/wallet-adapter-react'
import { WalletAdapterNetwork } from '@solana/wallet-adapter-base'
import { PhantomWalletAdapter, SolflareWalletAdapter } from '@solana/wallet-adapter-wallets'
import { WalletModalProvider } from '@solana/wallet-adapter-react-ui'

import '@solana/wallet-adapter-react-ui/styles.css'

const Main = () => {
    const network = WalletAdapterNetwork.Mainnet
    // 使用更可靠的 RPC 端点，避免 403 错误
    const endpoint = useMemo(() => {
        // 推荐使用免费的公共 RPC 端点
        return 'https://empty-hardworking-spree.solana-mainnet.quiknode.pro/d1cb31945987ed11caa6f3e63c2b8b86b00d41ff'
    }, [network])
    const wallets = useMemo(() => [new PhantomWalletAdapter(), new SolflareWalletAdapter()], [network])

    return (
        <ConnectionProvider endpoint={endpoint}>
            <WalletProvider wallets={wallets} autoConnect>
                <WalletModalProvider>
                    <App />
                </WalletModalProvider>
            </WalletProvider>
        </ConnectionProvider>
    )
}

createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Main />
  </React.StrictMode>,
)
