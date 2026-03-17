import { useEffect, useState, useContext, useMemo } from "react"
import { useParams } from "react-router-dom"
import { SocketContext } from "../context/SocketContext"
import { AuthContext } from "../context/AuthContext"
import { mapAuction, mapBid } from "../utils/mapper"
import { fetchAuction, fetchBids, placeBid } from "../api/auctions"
import styles from "./AuctionDetail.module.css"

function AuctionDetail() {

    const { id } = useParams()

    const { socket, joinAuction, leaveAuction } = useContext(SocketContext)
    const { user, wallet } = useContext(AuthContext)

    const [auction, setAuction] = useState(null)
    const [bids, setBids] = useState([])
    const [bidAmount, setBidAmount] = useState("")
    const [loading, setLoading] = useState(true)
    const [timeLeft, setTimeLeft] = useState("")
    const [error, setError] = useState("")
    const [placingBid, setPlacingBid] = useState(false)

    // 🔹 Fetch auction + bids
    useEffect(() => {

        const loadData = async () => {
            try {

                const auctionData = await fetchAuction(id)
                const bidsData = await fetchBids(id)

                setAuction(mapAuction(auctionData))
                setBids((bidsData || []).map(mapBid))

            } catch (err) {
                console.error(err)
                setError("Failed to load auction")
            }

            setLoading(false)
        }

        loadData()

    }, [id])

    // 🔹 Countdown Timer
    useEffect(() => {

        if (!auction) return

        const interval = setInterval(() => {

            const now = Date.now()
            const end = new Date(auction.end_time).getTime()

            const diff = end - now

            if (diff <= 0) {
                setTimeLeft("Auction Ended")
                clearInterval(interval)
                return
            }

            const h = Math.floor(diff / (1000 * 60 * 60))
            const m = Math.floor((diff / (1000 * 60)) % 60)
            const s = Math.floor((diff / 1000) % 60)

            setTimeLeft(`${h}h ${m}m ${s}s`)

        }, 1000)

        return () => clearInterval(interval)

    }, [auction])

    // 🔹 Derived states (IMPORTANT)
    const isHighestBidder = useMemo(() => {
        return auction?.current_highest_bidder_id === user?.id
    }, [auction, user])

    const canBid = useMemo(() => {
        if (!auction || auction.status !== "ACTIVE") return false
        if (!user) return false
        if (!wallet) return false

        const amount = Number(bidAmount)
        if (!amount) return false
        if (amount <= auction.current_highest_bid) return false
        if (amount > wallet.available) return false

        return true
    }, [auction, user, wallet, bidAmount])

    // 🔹 WebSocket setup
    useEffect(() => {

        if (!socket) return

        joinAuction(id)

        const handler = (msg) => {

            switch (msg.type) {

                case "BID_UPDATE":
                    setAuction(prev => ({
                        ...prev,
                        current_highest_bid: msg.amount,
                        current_highest_bidder_id: msg.bidder_id,
                        bid_count: prev.bid_count + 1
                    }))

                    setBids(prev => [
                        {
                            amount: msg.amount,
                            user_id: msg.bidder_id
                        },
                        ...prev
                    ])
                    break

                case "OUTBID":
                    if (msg.bidder_id === user?.id) {
                        setError("You have been outbid!")
                    }
                    break

                case "AUCTION_EXTENDED":
                    setAuction(prev => ({
                        ...prev,
                        end_time: msg.end_time
                    }))
                    break

                case "AUCTION_END":
                    setError("Auction ended")
                    break

                default:
                    break
            }

        }

        const unsubscribe = socket.onMessage(handler);

        return () => {
            unsubscribe();
            leaveAuction(id);
        };

    }, [socket, id, user])

    // 🔹 Place Bid
    const handleBid = async () => {

        setError("")

        const amount = Number(bidAmount)

        if (amount <= auction.current_highest_bid) {
            setError("Bid must be higher than current price")
            return
        }

        if (amount > wallet.available) {
            setError("Insufficient credits")
            return
        }

        try {

            setPlacingBid(true)

            await placeBid(id, amount)

            setBidAmount("")

        } catch (err) {
            setError(err.response?.data?.error || "Bid failed")
        } finally {
            setPlacingBid(false)
        }

    }

    if (loading) return <p>Loading...</p>
    if (!auction) return <p>Auction not found</p>

    return (

        <div className={styles.container}>

            <div className={styles.left}>
                <img src={auction.image_url} alt={auction.title} />
            </div>

            <div className={styles.right}>

                <h2>{auction.title}</h2>

                <p>{auction.description}</p>

                <h3>Current Bid: ${auction.current_highest_bid}</h3>

                <p>Bids: {auction.bid_count}</p>

                <p>Status: {auction.status}</p>

                <p><strong>Ends in:</strong> {timeLeft}</p>

                {/* 🔥 Status indicators */}
                {isHighestBidder && (
                    <p className={styles.winning}>You are winning</p>
                )}

                {error && (
                    <p className={styles.error}>{error}</p>
                )}

                {/* 🔥 Bid Box */}
                {auction.status === "ACTIVE" && (
                    <div className={styles.bidBox}>

                        <input
                            type="number"
                            value={bidAmount}
                            onChange={(e) => setBidAmount(e.target.value)}
                            placeholder="Enter bid amount"
                        />

                        <button
                            onClick={handleBid}
                            disabled={!canBid || placingBid}
                        >
                            {placingBid ? "Placing..." : "Place Bid"}
                        </button>

                    </div>
                )}

            </div>

            {/* 🔥 Bid History */}
            <div className={styles.bids}>

                <h3>Bid History</h3>

                {bids.map((b, i) => (
                    <div key={i}>
                        User {b.user_id} bid ${b.amount}
                    </div>
                ))}

            </div>

        </div>

    )

}

export default AuctionDetail