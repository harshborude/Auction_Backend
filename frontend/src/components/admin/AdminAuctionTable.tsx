import { useState } from 'react'
import { Auction } from '@/types'
import { endAuction, cancelAuction } from '@/api/admin'
import StatusBadge from '@/components/StatusBadge'
import styles from './AdminAuctionTable.module.css'

function formatDate(str: string) {
  return new Date(str).toLocaleString('en-US', {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
  })
}

export default function AdminAuctionTable({
  auctions,
  onUpdated,
}: {
  auctions: Auction[]
  onUpdated: () => void
}) {
  const [loading, setLoading] = useState<{ id: number; action: string } | null>(null)

  async function handleAction(id: number, action: 'end' | 'cancel') {
    const label = action === 'end' ? 'Force end' : 'Cancel'
    if (!window.confirm(`${label} auction #${id}?`)) return
    setLoading({ id, action })
    try {
      if (action === 'end') await endAuction(id)
      else await cancelAuction(id)
      onUpdated()
    } catch {
      // silently ignore
    } finally {
      setLoading(null)
    }
  }

  return (
    <table className={styles.table}>
      <thead className={styles.thead}>
        <tr>
          <th>Title</th>
          <th>Current bid</th>
          <th>Status</th>
          <th>Ends</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody className={styles.tbody}>
        {auctions.map((a) => (
          <tr key={a.ID}>
            <td>{a.Title}</td>
            <td>${(a.CurrentHighestBid || a.StartingPrice).toLocaleString()}</td>
            <td><StatusBadge status={a.Status} /></td>
            <td style={{ color: 'var(--text-muted)', fontSize: 13 }}>{formatDate(a.EndTime)}</td>
            <td>
              <div className={styles.actions}>
                {a.Status === 'ACTIVE' && (
                  <>
                    <button
                      className="btn btn-ghost btn-sm"
                      onClick={() => handleAction(a.ID, 'end')}
                      disabled={loading?.id === a.ID}
                    >
                      {loading?.id === a.ID && loading.action === 'end' ? '…' : 'Force end'}
                    </button>
                    <button
                      className="btn btn-danger btn-sm"
                      onClick={() => handleAction(a.ID, 'cancel')}
                      disabled={loading?.id === a.ID}
                    >
                      {loading?.id === a.ID && loading.action === 'cancel' ? '…' : 'Cancel'}
                    </button>
                  </>
                )}
                {a.Status !== 'ACTIVE' && (
                  <span style={{ fontSize: 13, color: 'var(--text-dim)' }}>—</span>
                )}
              </div>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
