import styles from "./UserTable.module.css";

function UserTable({ users, onAssign }) {

    return (
        <div className={styles.table}>

            <div className={styles.header}>
                <span>Username</span>
                <span>Email</span>
                <span>Role</span>
                <span>Credits</span>
                <span>Action</span>
            </div>

            {users.map((u) => {

                const available =
                    (u.Wallet?.Balance || 0) -
                    (u.Wallet?.ReservedBalance || 0);

                return (
                    <div key={u.ID} className={styles.row}>
                        <span>{u.Username}</span>
                        <span>{u.Email}</span>
                        <span>{u.Role}</span>
                        <span>{available}</span>

                        <button onClick={() => onAssign(u)}>
                            Assign
                        </button>
                    </div>
                );
            })}

        </div>
    );
}

export default UserTable;