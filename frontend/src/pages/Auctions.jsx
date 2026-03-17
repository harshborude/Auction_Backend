import { useEffect, useState } from "react"
import { fetchAuctions } from "../api/auctions"
import AuctionCard from "../components/AuctionCard"
import styles from "./Auctions.module.css"

function Auctions() {

    const [auctions, setAuctions] = useState([])
    const [loading, setLoading] = useState(true)

    useEffect(() => {

        const loadAuctions = async () => {

            try {

                const data = await fetchAuctions()

                setAuctions(data.auctions)

            } catch (err) {

                console.error("Failed to load auctions", err)

            }

            setLoading(false)

        }

        loadAuctions()

    }, [])

    if (loading) return <p>Loading auctions...</p>

    return (

        <div className={styles.container}>

            <h2>Live Auctions</h2>

            <div className={styles.grid}>

                {auctions.map((auction) => (

                    <AuctionCard key={auction.id} auction={auction} />

                ))}

            </div>

        </div>

    )

}

export default Auctions