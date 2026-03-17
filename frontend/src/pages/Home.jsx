import styles from "./Home.module.css"
import { Link } from "react-router-dom"

function Home() {
    return (
        <div className={styles.container}>

            {/* HERO SECTION */}
            <section className={styles.hero}>
                <h1>Real-Time Auction Platform</h1>

                <p>
                    Participate in live auctions, place bids in real time, and win exclusive items.
                </p>

                <Link to="/auctions" className={styles.cta}>
                    View Live Auctions
                </Link>
            </section>


            {/* HOW IT WORKS */}
            <section className={styles.how}>
                <h2>How It Works</h2>

                <div className={styles.steps}>
                    <div>
                        <h3>Create an Account</h3>
                        <p>Sign up and receive bidding credits.</p>
                    </div>

                    <div>
                        <h3>Join Auctions</h3>
                        <p>Browse active auctions and place bids in real time.</p>
                    </div>

                    <div>
                        <h3>Win Items</h3>
                        <p>If your bid is highest when the auction ends, you win.</p>
                    </div>
                </div>

            </section>

        </div>
    )
}

export default Home