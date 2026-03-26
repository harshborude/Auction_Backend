import { useEffect, useState } from 'react'
import { getWallet, getTransactions } from '@/api/wallet'
import { Transaction, Wallet as WalletType } from '@/types'
import Pagination from '@/components/Pagination'
import Skeleton from '@/components/Skeleton'
import styles from './Wallet.module.css'

const TX_CONFIG: Record<string, { label: string; dotClass: string; amountClass: string; sign: string }> = {
  ADMIN_ASSIGN: { label: 'Credits assigned', dotClass: styles.dotGreen, amountClass: styles.txPositive, sign: '+' },
  BID_RESERVE:  { label: 'Bid reserved',     dotClass: styles.dotAmber, amountClass: styles.txNeutral,  sign: '-' },
  BID_RELEASE:  { label: 'Bid released',     dotClass: styles.dotBlue,  amountClass: styles.txPositive, sign: '+' },
  AUCTION_WIN:  { label: 'Auction won',      dotClass: styles.dotRed,   amountClass: styles.txNegative, sign: '-' },
}

function formatDate(str: string) {
  return new Date(str).toLocaleDateString('en-US', {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
  })
}

export default function Wallet() {
  const [wallet, setWallet] = useState<WalletType | null>(null)
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(false)
  const [loading, setLoading] = useState(true)
  const [txLoading, setTxLoading] = useState(false)
  const LIMIT = 20

  useEffect(() => {
    getWallet().then(({ data }) => setWallet(data)).finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    setTxLoading(true)
    getTransactions(page, LIMIT)
      .then(({ data }) => {
        setTransactions(data.transactions ?? [])
        setHasMore((data.transactions ?? []).length === LIMIT)
      })
      .finally(() => setTxLoading(false))
  }, [page])

  const available = wallet ? wallet.Balance - wallet.ReservedBalance : 0

  return (
    <div className="page">
      <div className="container">
        <h1 className={styles.title}>Wallet</h1>

        {loading ? (
          <div style={{ display: 'flex', gap: 16, marginBottom: 40 }}>
            <Skeleton height={88} width={200} borderRadius="var(--radius)" />
            <Skeleton height={88} width={200} borderRadius="var(--radius)" />
          </div>
        ) : wallet && (
          <div className={styles.balanceGrid}>
            <div className={styles.balanceCard}>
              <p className={styles.balanceLabel}>Available</p>
              <p className={`${styles.balanceAmount} ${styles.available}`}>
                ${available.toLocaleString()}
              </p>
            </div>
            <div className={styles.balanceCard}>
              <p className={styles.balanceLabel}>Reserved in bids</p>
              <p className={`${styles.balanceAmount} ${styles.reserved}`}>
                ${wallet.ReservedBalance.toLocaleString()}
              </p>
            </div>
            <div className={styles.balanceCard}>
              <p className={styles.balanceLabel}>Total balance</p>
              <p className={styles.balanceAmount}>${wallet.Balance.toLocaleString()}</p>
            </div>
          </div>
        )}

        <p className={styles.sectionTitle}>Transaction history</p>

        {txLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} height={48} />)}
          </div>
        ) : transactions.length === 0 ? (
          <p className={styles.empty}>No transactions yet.</p>
        ) : (
          <div className={styles.txList}>
            {transactions.map((tx) => {
              const cfg = TX_CONFIG[tx.Type] ?? { label: tx.Type, dotClass: styles.dotBlue, amountClass: '', sign: '' }
              return (
                <div key={tx.ID} className={styles.txRow}>
                  <span className={`${styles.txDot} ${cfg.dotClass}`} />
                  <div className={styles.txInfo}>
                    <p className={styles.txType}>{cfg.label}</p>
                    {tx.Reference && <p className={styles.txRef}>{tx.Reference}</p>}
                  </div>
                  <span className={`${styles.txAmount} ${cfg.amountClass}`}>
                    {cfg.sign}${Math.abs(tx.Amount).toLocaleString()}
                  </span>
                  <span className={styles.txTime}>{formatDate(tx.CreatedAt)}</span>
                </div>
              )
            })}
          </div>
        )}

        {!txLoading && transactions.length > 0 && (
          <Pagination
            page={page}
            hasMore={hasMore}
            onPrev={() => setPage((p) => p - 1)}
            onNext={() => setPage((p) => p + 1)}
          />
        )}
      </div>
    </div>
  )
}
