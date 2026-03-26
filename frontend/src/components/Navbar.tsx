import { Link, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '@/context/AuthContext'
import styles from './Navbar.module.css'

export default function Navbar() {
  const { user, wallet, isAuthenticated, isAdmin, logout } = useAuth()
  const navigate = useNavigate()

  async function handleLogout() {
    await logout()
    navigate('/login')
  }

  const available = wallet ? wallet.Balance - wallet.ReservedBalance : 0

  return (
    <nav className={styles.nav}>
      <div className={styles.inner}>
        <Link to="/" className={styles.brand}>AUCTION</Link>
        <div className={styles.links}>
          <NavLink to="/auctions" className={({ isActive }) => (isActive ? styles.active : '')}>
            Auctions
          </NavLink>
          {isAuthenticated && (
            <NavLink to="/wallet" className={({ isActive }) => (isActive ? styles.active : '')}>
              Wallet
            </NavLink>
          )}
          {isAdmin && (
            <NavLink to="/admin" className={({ isActive }) => (isActive ? styles.active : '')}>
              Admin
            </NavLink>
          )}
        </div>
        <div className={styles.right}>
          {isAuthenticated ? (
            <>
              <span className={styles.balance}>${available.toLocaleString()}</span>
              <span className={styles.username}>{user?.Username}</span>
              <button className={styles.logout} onClick={handleLogout}>Logout</button>
            </>
          ) : (
            <>
              <Link to="/login" className={styles.loginLink}>Login</Link>
              <Link to="/register" className={styles.registerBtn}>Register</Link>
            </>
          )}
        </div>
      </div>
    </nav>
  )
}
