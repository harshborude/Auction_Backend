import { Link } from 'react-router-dom'
import { Auction } from '@/types'
import { useCountdown } from '@/hooks/useCountdown'
import StatusBadge from './StatusBadge'
import styles from './AuctionCard.module.css'

export default function AuctionCard({ auction }: { auction: Auction }) {
  const { label, state } = useCountdown(auction.Status === 'ACTIVE' ? auction.EndTime : null)

  const currentBid = auction.CurrentHighestBid || auction.StartingPrice

  return (
    <Link to={`/auctions/${auction.ID}`} className={styles.card}>
      <div className={styles.imageWrap}>
        {auction.ImageURL ? (
          <img src={auction.ImageURL} alt={auction.Title} className={styles.image} />
        ) : (
          <div className={styles.placeholder}>No image</div>
        )}
        <div className={styles.statusWrap}>
          <StatusBadge status={auction.Status} />
        </div>
      </div>
      <div className={styles.body}>
        <h3 className={styles.title}>{auction.Title}</h3>
        <div className={styles.meta}>
          <div>
            <div className={styles.metaLabel}>Current bid</div>
            <div className={styles.metaValue}>${currentBid.toLocaleString()}</div>
          </div>
          {auction.Status === 'ACTIVE' && (
            <div className={styles.countdown} data-state={state}>
              <div className={styles.metaLabel}>Ends in</div>
              <div className={styles.metaValue}>{label}</div>
            </div>
          )}
        </div>
        <div className={styles.bids}>
          {auction.BidCount} bid{auction.BidCount !== 1 ? 's' : ''}
        </div>
      </div>
    </Link>
  )
}
