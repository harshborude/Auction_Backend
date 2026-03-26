import { FormEvent, useCallback, useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { fetchAuction, fetchBids, placeBid } from '@/api/auctions'
import { Auction, Bid, WsMessage } from '@/types'
import { useAuth } from '@/context/AuthContext'
import { useSocket } from '@/context/SocketContext'
import { useCountdown } from '@/hooks/useCountdown'
import StatusBadge from '@/components/StatusBadge'
import styles from './AuctionDetail.module.css'

function timeAgo(dateStr: string) {
  const diff = Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000)
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}

export default function AuctionDetail() {
  const { id } = useParams<{ id: string }>()
  const auctionId = Number(id)
  const navigate = useNavigate()
  const { user, wallet, isAuthenticated, refreshWallet } = useAuth()
  const { joinAuction, leaveAuction, subscribe } = useSocket()

  const [auction, setAuction] = useState<Auction | null>(null)
  const [bids, setBids] = useState<Bid[]>([])
  const [loading, setLoading] = useState(true)
  const [bidAmount, setBidAmount] = useState('')
  const [bidError, setBidError] = useState('')
  const [bidding, setBidding] = useState(false)
  const [extended, setExtended] = useState(false)
  const [endTime, setEndTime] = useState<string | null>(null)

  const { label: countdown, state: countdownState } = useCountdown(
    auction?.Status === 'ACTIVE' ? endTime : null
  )

  useEffect(() => {
    setLoading(true)
    Promise.all([fetchAuction(auctionId), fetchBids(auctionId)])
      .then(([{ data: a }, { data: b }]) => {
        setAuction(a)
        setEndTime(a.EndTime)
        setBids(b)
      })
      .catch(() => navigate('/auctions'))
      .finally(() => setLoading(false))
  }, [auctionId, navigate])

  const handleMessage = useCallback(
    (msg: WsMessage) => {
      if (msg.auction_id !== auctionId) return

      if (msg.type === 'BID_UPDATE') {
        setAuction((prev) =>
          prev
            ? {
                ...prev,
                CurrentHighestBid: msg.amount,
                CurrentHighestBidderID: msg.bidder_id,
                BidCount: prev.BidCount + 1,
              }
            : prev
        )
        setBids((prev) => [
          {
            ID: Date.now(),
            UserID: msg.bidder_id,
            Amount: msg.amount,
            AuctionID: auctionId,
            user: {
              ID: msg.bidder_id,
              Username: msg.bidder_id === user?.ID ? user.Username : `User ${msg.bidder_id}`,
            },
            CreatedAt: new Date().toISOString(),
          },
          ...prev,
        ])
      } else if (msg.type === 'AUCTION_EXTENDED') {
        if (msg.end_time) setEndTime(msg.end_time)
        setExtended(true)
        setTimeout(() => setExtended(false), 5000)
      } else if (msg.type === 'AUCTION_END') {
        setAuction((prev) => (prev ? { ...prev, Status: 'ENDED' } : prev))
      }
    },
    [auctionId, user]
  )

  useEffect(() => {
    if (!isAuthenticated || !auction) return
    joinAuction(auctionId)
    const unsub = subscribe(handleMessage)
    return () => {
      leaveAuction(auctionId)
      unsub()
    }
  }, [auctionId, isAuthenticated, auction, joinAuction, leaveAuction, subscribe, handleMessage])

  async function handleBid(e: FormEvent) {
    e.preventDefault()
    setBidError('')
    const amount = Number(bidAmount)
    if (!amount || amount <= 0) {
      setBidError('Enter a valid amount')
      return
    }
    setBidding(true)
    try {
      await placeBid(auctionId, amount)
      setBidAmount('')
      await refreshWallet()
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error
      setBidError(msg || 'Failed to place bid')
    } finally {
      setBidding(false)
    }
  }

  if (loading) {
    return (
      <div className="page">
        <div className="container">
          <p style={{ color: 'var(--text-muted)' }}>Loading…</p>
        </div>
      </div>
    )
  }

  if (!auction) return null

  const available = wallet ? wallet.Balance - wallet.ReservedBalance : 0
  const isWinning = auction.CurrentHighestBidderID === user?.ID
  const isSeller = auction.CreatedBy === user?.ID
  const minBid =
    auction.BidCount === 0
      ? auction.StartingPrice
      : isWinning
      ? auction.CurrentHighestBid + 1
      : auction.CurrentHighestBid + auction.BidIncrement
  const canBid = isAuthenticated && auction.Status === 'ACTIVE' && !isSeller

  return (
    <div className="page">
      <div className="container">
        <Link to="/auctions" className={styles.back}>← Back to auctions</Link>
        <div className={styles.layout}>
          {/* Left: image + description + bid history */}
          <div>
            {auction.ImageURL ? (
              <img src={auction.ImageURL} alt={auction.Title} className={styles.image} />
            ) : (
              <div className={styles.imagePlaceholder}>No image</div>
            )}
            <div className={styles.titleRow}>
              <h1 className={styles.title}>{auction.Title}</h1>
              <StatusBadge status={auction.Status} />
            </div>
            {auction.Description && (
              <p className={styles.description}>{auction.Description}</p>
            )}

            <hr className={styles.divider} />

            <p className={styles.bidsTitle}>Bid history ({bids.length})</p>
            {bids.length === 0 ? (
              <p className={styles.noBids}>No bids yet. Be the first!</p>
            ) : (
              <div className={styles.bidList}>
                {bids.map((bid, i) => (
                  <div key={bid.ID ?? i} className={styles.bidRow}>
                    <span className={`${styles.bidUser} ${bid.UserID === user?.ID ? styles.isYou : ''}`}>
                      {bid.UserID === user?.ID
                        ? 'You'
                        : bid.user?.Username ?? `User ${bid.UserID}`}
                    </span>
                    <span className={styles.bidAmount}>${bid.Amount.toLocaleString()}</span>
                    <span className={styles.bidTime}>{timeAgo(bid.CreatedAt)}</span>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Right: bid panel */}
          <div className={styles.panel}>
            <div>
              <p className={styles.currentBidLabel}>
                {auction.BidCount === 0 ? 'Starting price' : 'Current bid'}
              </p>
              <p className={styles.currentBidAmount}>
                ${(auction.CurrentHighestBid || auction.StartingPrice).toLocaleString()}
              </p>
            </div>

            {auction.Status === 'ACTIVE' && (
              <div className={styles.panelRow}>
                <div>
                  <p className={styles.panelLabel}>Ends in</p>
                  <p className={styles.countdown} data-state={countdownState}>{countdown}</p>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <p className={styles.panelLabel}>Bids</p>
                  <p className={styles.panelValue}>{auction.BidCount}</p>
                </div>
              </div>
            )}

            {extended && (
              <div className={styles.bannerExtended}>⚡ Extended by 30s (anti-sniping)</div>
            )}

            <hr className={styles.panelDivider} />

            {auction.Status === 'ACTIVE' && canBid && (
              <>
                <p className={styles.walletInfo}>
                  Available: <span className={styles.walletBalance}>${available.toLocaleString()}</span>
                </p>
                <form onSubmit={handleBid}>
                  <div className={styles.bidInput}>
                    <input
                      className="form-input"
                      type="number"
                      value={bidAmount}
                      onChange={(e) => setBidAmount(e.target.value)}
                      placeholder={`≥ $${minBid.toLocaleString()}`}
                      min={minBid}
                    />
                    <button type="submit" className="btn btn-primary" disabled={bidding}>
                      {bidding ? '…' : 'Bid'}
                    </button>
                  </div>
                  <p className={styles.increment} style={{ marginTop: 6 }}>
                    Min increment: ${auction.BidIncrement.toLocaleString()}
                  </p>
                  {bidError && <p className="error-text" style={{ marginTop: 8 }}>{bidError}</p>}
                </form>
              </>
            )}

            {/* Status banners */}
            {auction.Status === 'ACTIVE' && isAuthenticated && !isSeller && (
              <div className={`${styles.banner} ${isWinning ? styles.bannerWinning : styles.bannerOutbid}`}>
                {isWinning
                  ? '🏆 You are currently winning'
                  : auction.BidCount > 0
                  ? '⬆ You have been outbid'
                  : 'Place the first bid!'}
              </div>
            )}

            {auction.Status === 'ENDED' && (
              <div className={`${styles.banner} ${styles.bannerEnded}`}>
                {auction.CurrentHighestBidderID === user?.ID
                  ? '🎉 You won this auction!'
                  : `Auction ended — won by User ${auction.CurrentHighestBidderID ?? '—'}`}
              </div>
            )}

            {auction.Status === 'CANCELLED' && (
              <div className={`${styles.banner} ${styles.bannerEnded}`}>
                This auction was cancelled.
              </div>
            )}

            {!isAuthenticated && auction.Status === 'ACTIVE' && (
              <div className={`${styles.banner} ${styles.bannerEnded}`}>
                <Link to="/login" style={{ color: 'var(--accent)' }}>Sign in</Link> to place a bid
              </div>
            )}

            {isSeller && (
              <div className={`${styles.banner} ${styles.bannerEnded}`}>
                You created this auction
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
