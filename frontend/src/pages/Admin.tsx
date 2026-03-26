import { useCallback, useEffect, useState } from 'react'
import { fetchUsers, getAdminAuctions } from '@/api/admin'
import { User, Auction } from '@/types'
import UserTable from '@/components/admin/UserTable'
import AdminAuctionTable from '@/components/admin/AdminAuctionTable'
import CreateAuctionForm from '@/components/admin/CreateAuctionForm'
import Skeleton from '@/components/Skeleton'
import styles from './Admin.module.css'

export default function Admin() {
  const [tab, setTab] = useState<'users' | 'auctions'>('users')
  const [users, setUsers] = useState<User[]>([])
  const [auctions, setAuctions] = useState<Auction[]>([])
  const [usersLoading, setUsersLoading] = useState(true)
  const [auctionsLoading, setAuctionsLoading] = useState(true)

  const loadUsers = useCallback(() => {
    setUsersLoading(true)
    fetchUsers()
      .then(({ data }) => setUsers(data))
      .finally(() => setUsersLoading(false))
  }, [])

  const loadAuctions = useCallback(() => {
    setAuctionsLoading(true)
    getAdminAuctions()
      .then(({ data }) => setAuctions(data.auctions ?? []))
      .finally(() => setAuctionsLoading(false))
  }, [])

  useEffect(() => { loadUsers() }, [loadUsers])
  useEffect(() => { loadAuctions() }, [loadAuctions])

  const activeCount = auctions.filter((a) => a.Status === 'ACTIVE').length

  return (
    <div className="page">
      <div className="container">
        <h1 className={styles.title}>Admin Dashboard</h1>

        <div className={styles.tabs}>
          <button
            className={`${styles.tab} ${tab === 'users' ? styles.active : ''}`}
            onClick={() => setTab('users')}
          >
            Users ({users.length})
          </button>
          <button
            className={`${styles.tab} ${tab === 'auctions' ? styles.active : ''}`}
            onClick={() => setTab('auctions')}
          >
            Auctions ({activeCount} active)
          </button>
        </div>

        {tab === 'users' && (
          <div className={styles.panel}>
            {usersLoading ? (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                {Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} height={48} />)}
              </div>
            ) : (
              <UserTable users={users} onUpdated={loadUsers} />
            )}
          </div>
        )}

        {tab === 'auctions' && (
          <div className={styles.panel}>
            <p className={styles.sectionLabel}>Create auction</p>
            <CreateAuctionForm onCreated={loadAuctions} />

            <div style={{ height: 40 }} />

            <p className={styles.sectionLabel}>All auctions</p>
            {auctionsLoading ? (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                {Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} height={48} />)}
              </div>
            ) : (
              <AdminAuctionTable auctions={auctions} onUpdated={loadAuctions} />
            )}
          </div>
        )}
      </div>
    </div>
  )
}
