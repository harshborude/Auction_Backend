import { useState, useEffect } from "react"
import { getWallet, getTransactions } from "../api/wallet"
import styles from "./Wallet.module.css"

import { mapWallet, mapTransaction } from "../utils/mapper"



function Wallet() {

    const [wallet, setWallet] = useState(null)
    const [transactions, setTransactions] = useState([])
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        fetchWalletData()
    }, [])

    const fetchWalletData = async () => {
        try {

            const walletRes = await getWallet()
            const txnRes = await getTransactions()

            setWallet(mapWallet(walletRes.data || {}))
            setTransactions((txnRes.data.transactions || []).map(mapTransaction))
            console.log(walletRes.data)
            console.log(txnRes.data.transactions)
        } catch (err) {
            console.error("Failed to fetch wallet data", err)
        } finally {
            setLoading(false)
        }
    }

    if (loading) return <p>Loading wallet...</p>

    if (!wallet) return <p>Wallet not found</p>

    return (

        <div className={styles.container}>

            <h2>Wallet</h2>

            {/* 🔹 Wallet Info */}
            <div className={styles.balanceBox}>

                <p><strong>Available Credits:</strong> ₹{wallet?.available_credits || 0}</p>

                <p><strong>Reserved Credits:</strong> ₹{wallet?.reserved_credits || 0}</p>

                <p><strong>Total Credits:</strong>₹{wallet?.total_credits || 0}</p>

            </div>

            {/* 🔹 Transactions */}
            <div className={styles.transactions}>

                <h3>Transactions</h3>

                {transactions.length === 0 ? (
                    <p>No transactions yet</p>
                ) : (
                    transactions.map((t) => (
                        <div key={t.id} className={styles.transactionItem}>
                            <p>{t.type}</p>
                            <p>₹{t.amount}</p>
                            <p>{t.description}</p>
                        </div>
                    ))
                )}

            </div>

        </div>

    )
}

export default Wallet