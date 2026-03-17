import { useEffect, useState } from "react";
import { fetchUsers, assignCredits, createAuction, endAuction, cancelAuction } from "../api/admin";
import { fetchAuctions } from "../api/auctions";

import UserTable from "../components/admin/UserTable";
import AssignCreditsModal from "../components/admin/AssignCreditsModal";
import CreateAuctionForm from "../components/admin/CreateAuctionForm";
import AdminAuctionTable from "../components/admin/AdminAuctionTable";

import styles from "./AdminDashboard.module.css";

function AdminDashboard() {

    const [users, setUsers] = useState([]);
    const [filtered, setFiltered] = useState([]);
    const [search, setSearch] = useState("");
    const [selectedUser, setSelectedUser] = useState(null);

    const [auctions, setAuctions] = useState([]);

    // 🔹 Load users
    useEffect(() => {
        const load = async () => {
            const data = await fetchUsers();
            setUsers(data);
            setFiltered(data);
        };
        load();
    }, []);

    // 🔹 Load auctions
    const loadAuctions = async () => {
        const data = await fetchAuctions();
        setAuctions(data || []);
    };

    useEffect(() => {
        loadAuctions();
    }, []);

    // 🔹 Search
    useEffect(() => {
        const result = users.filter((u) =>
            u.Username.toLowerCase().includes(search.toLowerCase())
        );
        setFiltered(result);
    }, [search, users]);

    // 🔹 Assign credits
    const handleAssign = async (amount) => {
        await assignCredits(selectedUser.ID, amount);
        const data = await fetchUsers();
        setUsers(data);
        setFiltered(data);
        setSelectedUser(null);
    };

    // 🔹 Create auction
    const handleCreateAuction = async (data) => {
        await createAuction(data);
        loadAuctions();
    };

    // 🔹 End auction
    const handleEnd = async (id) => {
        await endAuction(id);
        loadAuctions();
    };

    // 🔹 Cancel auction
    const handleCancel = async (id) => {
        await cancelAuction(id);
        loadAuctions();
    };

    return (
        <div className={styles.container}>

            <h2>Admin Dashboard</h2>

            {/* USERS */}
            <h3>Users</h3>

            <input
                type="text"
                placeholder="Search users..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
            />

            <UserTable users={filtered} onAssign={setSelectedUser} />

            {selectedUser && (
                <AssignCreditsModal
                    user={selectedUser}
                    onClose={() => setSelectedUser(null)}
                    onSubmit={handleAssign}
                />
            )}

            {/* AUCTIONS */}
            <h3>Auctions</h3>

            <CreateAuctionForm onCreate={handleCreateAuction} />

            <AdminAuctionTable
                auctions={auctions}
                onEnd={handleEnd}
                onCancel={handleCancel}
            />

        </div>
    );
}

export default AdminDashboard;