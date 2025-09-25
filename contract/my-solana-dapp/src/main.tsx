import React, { useMemo } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import './index.css'

import { ConnectionProvider, WalletProvider } from '@solana/wallet-adapter-react'
import { WalletAdapterNetwork } from '@solana/wallet-adapter-base'
import { PhantomWalletAdapter, SolflareWalletAdapter } from '@solana/wallet-adapter-wallets'
import { WalletModalProvider } from '@solana/wallet-adapter-react-ui'
import { clusterApiUrl } from '@solana/web3.js'

import '@solana/wallet-adapter-react-ui/styles.css'

const Main = () => {
    const network = WalletAdapterNetwork.Devnet
    const endpoint = useMemo(() => clusterApiUrl(network), [network])
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
