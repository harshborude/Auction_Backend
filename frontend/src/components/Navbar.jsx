import { Link } from "react-router-dom"
import styles from "./Navbar.module.css"

function Navbar() {
    return (
        <nav className={styles.navbar}>
            <Link to="/">Home</Link>

            <Link to="/auctions">Auctions</Link>

            <Link to="/wallet">Wallet</Link>

            <Link to="/login">Login</Link>
        </nav>
    )
}

export default Navbar