import { Link } from "react-router-dom"
import { useContext } from "react"
import { AuthContext } from "../context/AuthContext"
import styles from "./Navbar.module.css"

function Navbar() {

    const { isAuthenticated, logout, user, wallet } = useContext(AuthContext)

    return (
        <nav className={styles.navbar}>

            <div className={styles.left}>
                <Link to="/">Home</Link>
                <Link to="/auctions">Auctions</Link>
                <Link to="/wallet">Wallet</Link>
            </div>

            <div className={styles.right}>
                {isAuthenticated && user && (
                    <>
                        <span className={styles.user}>
                            {user.Username || "User"}
                        </span>

                        {wallet && (
                            <span className={styles.credits}>
                                💰 {wallet.available}
                            </span>
                        )}
                    </>
                )}

                {isAuthenticated ? (
                    <button onClick={logout}>Logout</button>
                ) : (
                    <Link to="/login">Login</Link>
                )}
            </div>

        </nav>
    )
}

export default Navbar