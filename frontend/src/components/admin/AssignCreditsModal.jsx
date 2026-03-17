import { useState } from "react";
import styles from "./AssignCreditsModal.module.css";

function AssignCreditsModal({ user, onClose, onSubmit }) {

    const [amount, setAmount] = useState("");

    const handleSubmit = () => {
        if (!amount) return;
        onSubmit(Number(amount));
    };

    return (
        <div className={styles.overlay}>

            <div className={styles.modal}>
                <h3>Assign Credits</h3>

                <p>User: {user.Username}</p>

                <input
                    type="number"
                    placeholder="Enter amount"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                />

                <div className={styles.actions}>
                    <button onClick={handleSubmit}>OK</button>
                    <button onClick={onClose}>Cancel</button>
                </div>
            </div>

        </div>
    );
}

export default AssignCreditsModal;