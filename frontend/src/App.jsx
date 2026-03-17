import { Routes, Route, Navigate } from "react-router-dom";
import { useContext } from "react";

import Navbar from "./components/Navbar";
import Home from "./pages/Home";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Auctions from "./pages/Auctions";
import AuctionDetail from "./pages/AuctionDetail";
import Wallet from "./pages/Wallet";
import AdminDashboard from "./pages/AdminDashboard";
import ProtectedRoute from "./components/ProtectedRoute";
import { AuthContext } from "./context/AuthContext";

function App() {

    const { user } = useContext(AuthContext);

    return (
        <>
            <Navbar />

            <Routes>
                <Route path="/" element={<Home />} />

                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Register />} />

                <Route path="/auctions" element={<Auctions />} />
                <Route path="/auctions/:id" element={<AuctionDetail />} />

                <Route
                    path="/wallet"
                    element={
                        <ProtectedRoute>
                            <Wallet />
                        </ProtectedRoute>
                    }
                />

                {/* 🔥 ADMIN ROUTE */}
                <Route
                    path="/admin"
                    element={
                        user?.Role === "ADMIN"
                            ? <AdminDashboard />
                            : <Navigate to="/" />
                    }
                />
            </Routes>
        </>
    );
}

export default App;