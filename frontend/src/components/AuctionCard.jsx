import { Link } from "react-router-dom"
import styles from "./AuctionCard.module.css"

function AuctionCard({ auction }) {

    return (
        <div className={styles.card}>

            <img
                src={auction.image_url}
                alt={auction.title}
                className={styles.image}
            />

            <div className={styles.info}>

                <h3>{auction.title}</h3>

                <p>
                    Current Bid: <strong>${auction.current_highest_bid}</strong>
                </p>

                <p>Bids: {auction.bid_count}</p>

                <Link to={`/auctions/${auction.id}`} className={styles.button}>
                    View Auction
                </Link>

            </div>

        </div>
    )
}

export default AuctionCard