import { useState } from 'react'
import { User } from '@/types'
import { assignCredits, banUser } from '@/api/admin'
import styles from './UserTable.module.css'

export default function UserTable({ users, onUpdated }: { users: User[]; onUpdated: () => void }) {
  const [assigning, setAssigning] = useState<number | null>(null)
  const [banning, setBanning] = useState<number | null>(null)
  const [amounts, setAmounts] = useState<Record<number, string>>({})
  const [error, setError] = useState('')

  async function handleAssign(userId: number) {
    const amount = Number(amounts[userId])
    if (!amount || amount <= 0) return
    setAssigning(userId)
    setError('')
    try {
      await assignCredits(userId, amount)
      setAmounts((prev) => ({ ...prev, [userId]: '' }))
      onUpdated()
    } catch {
      setError('Failed to assign credits')
    } finally {
      setAssigning(null)
    }
  }

  async function handleBan(userId: number, username: string) {
    if (!window.confirm(`Ban user "${username}"? They will no longer be able to log in.`)) return
    setBanning(userId)
    setError('')
    try {
      await banUser(userId)
      onUpdated()
    } catch {
      setError('Failed to ban user')
    } finally {
      setBanning(null)
    }
  }

  return (
    <>
      {error && <p className="error-text" style={{ marginBottom: 12 }}>{error}</p>}
      <table className={styles.table}>
        <thead className={styles.thead}>
          <tr>
            <th>Username</th>
            <th>Email</th>
            <th>Role</th>
            <th>Status</th>
            <th>Balance</th>
            <th>Assign credits</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody className={styles.tbody}>
          {users.map((u) => {
            const available = u.Wallet ? u.Wallet.Balance - u.Wallet.ReservedBalance : 0
            return (
              <tr key={u.ID}>
                <td>{u.Username}</td>
                <td style={{ color: 'var(--text-muted)' }}>{u.Email}</td>
                <td>
                  <span className={`${styles.role} ${u.Role === 'ADMIN' ? styles.roleAdmin : styles.roleUser}`}>
                    {u.Role}
                  </span>
                </td>
                <td>
                  {u.IsActive ? (
                    <span style={{ color: 'var(--green)', fontSize: 12 }}>Active</span>
                  ) : (
                    <span style={{ color: 'var(--red)', fontSize: 12 }}>Banned</span>
                  )}
                </td>
                <td>${available.toLocaleString()}</td>
                <td>
                  <div className={styles.assignRow}>
                    <input
                      className="form-input"
                      type="number"
                      min={1}
                      placeholder="Amount"
                      value={amounts[u.ID] ?? ''}
                      onChange={(e) => setAmounts((prev) => ({ ...prev, [u.ID]: e.target.value }))}
                    />
                    <button
                      className="btn btn-primary btn-sm"
                      onClick={() => handleAssign(u.ID)}
                      disabled={assigning === u.ID || !amounts[u.ID]}
                    >
                      {assigning === u.ID ? '…' : 'Add'}
                    </button>
                  </div>
                </td>
                <td>
                  {u.IsActive && u.Role !== 'ADMIN' && (
                    <button
                      className="btn btn-danger btn-sm"
                      onClick={() => handleBan(u.ID, u.Username)}
                      disabled={banning === u.ID}
                    >
                      {banning === u.ID ? '…' : 'Ban'}
                    </button>
                  )}
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </>
  )
}
