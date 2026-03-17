import styles from "./AdminAuctionTable.module.css";

function AdminAuctionTable({ auctions, onEnd, onCancel }) {

    return (
        <div className={styles.table}>

            <div className={styles.header}>
                <span>Title</span>
                <span>Current Bid</span>
                <span>Status</span>
                <span>Ends</span>
                <span>Actions</span>
            </div>

            {auctions.map((a) => (
                <div key={a.ID} className={styles.row}>

                    <span>{a.Title}</span>
                    <span>{a.CurrentHighestBid}</span>
                    <span>{a.Status}</span>
                    <span>{new Date(a.EndTime).toLocaleString()}</span>

                    <div>
                        <button onClick={() => onEnd(a.ID)}>End</button>
                        <button onClick={() => onCancel(a.ID)}>Cancel</button>
                    </div>

                </div>
            ))}

        </div>
    );
}

export default AdminAuctionTable;