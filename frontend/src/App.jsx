import { BrowserRouter, Routes, Route } from "react-router-dom";

import Navbar from "./components/Navbar";
import Home from "./pages/Home";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Auctions from "./pages/Auctions";
import AuctionDetail from "./pages/AuctionDetail";
import Wallet from "./pages/Wallet";

function App() {
    return (
        <BrowserRouter>
            <Navbar />

            <Routes>
                <Route path="/" element={<Home />} />

                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Register />} />

                <Route path="/auctions" element={<Auctions />} />
                <Route path="/auctions/:id" element={<AuctionDetail />} />

                <Route path="/wallet" element={<Wallet />} />
            </Routes>
        </BrowserRouter>
    );
}

export default App;