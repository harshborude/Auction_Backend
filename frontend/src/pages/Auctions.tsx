import { useEffect, useState } from 'react'
import { fetchAuctions } from '@/api/auctions'
import { Auction } from '@/types'
import AuctionCard from '@/components/AuctionCard'
import Pagination from '@/components/Pagination'
import Skeleton from '@/components/Skeleton'
import styles from './Auctions.module.css'

const LIMIT = 12

function CardSkeleton() {
  return (
    <div className={styles.skeletonCard}>
      <Skeleton height={158} borderRadius="0" />
      <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
        <Skeleton height={18} width="75%" />
        <Skeleton height={14} width="50%" />
        <Skeleton height={22} width="40%" />
      </div>
    </div>
  )
}

export default function Auctions() {
  const [auctions, setAuctions] = useState<Auction[]>([])
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    setLoading(true)
    setError('')
    fetchAuctions(page, LIMIT)
      .then(({ data }) => {
        setAuctions(data.auctions ?? [])
        setHasMore((data.auctions ?? []).length === LIMIT)
      })
      .catch(() => setError('Failed to load auctions'))
      .finally(() => setLoading(false))
  }, [page])

  return (
    <div className="page">
      <div className="container">
        <div className={styles.header}>
          <h1 className={styles.title}>Live Auctions</h1>
          {!loading && <span className={styles.count}>{auctions.length} active</span>}
        </div>

        {error && <p className="error-text">{error}</p>}

        {loading ? (
          <div className={styles.skeletonGrid}>
            {Array.from({ length: 6 }).map((_, i) => <CardSkeleton key={i} />)}
          </div>
        ) : auctions.length === 0 ? (
          <div className={styles.empty}>
            <div className={styles.emptyIcon}>🔨</div>
            <p>No active auctions right now. Check back soon.</p>
          </div>
        ) : (
          <div className={styles.grid}>
            {auctions.map((a) => <AuctionCard key={a.ID} auction={a} />)}
          </div>
        )}

        {!loading && auctions.length > 0 && (
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
