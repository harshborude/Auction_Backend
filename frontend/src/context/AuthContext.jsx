import { createContext, useState, useEffect } from "react";
import { loginUser, getCurrentUser, logoutUser } from "../api/auth";
import { getWallet } from "../api/wallet";
import { useNavigate } from "react-router-dom";

export const AuthContext = createContext();

export const AuthProvider = ({ children }) => {

    const navigate = useNavigate();

    const [user, setUser] = useState(null);
    const [wallet, setWallet] = useState(null);
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [loading, setLoading] = useState(true);

    // 🔹 Initialize session
    useEffect(() => {

        const initAuth = async () => {

            const token = localStorage.getItem("access_token");

            if (token) {
                try {
                    // ✅ fetch user
                    const userRes = await getCurrentUser();
                    setUser(userRes.data);
                    setIsAuthenticated(true);

                    // ✅ fetch wallet
                    const walletRes = await getWallet();

                    setWallet({
                        available: walletRes.data.Balance - walletRes.data.ReservedBalance,
                        reserved: walletRes.data.ReservedBalance,
                        total: walletRes.data.Balance
                    });

                } catch (err) {
                    console.log("Session expired");

                    localStorage.removeItem("access_token");
                    localStorage.removeItem("refresh_token");
                }
            }

            setLoading(false);
        };

        initAuth();

    }, []);

    // 🔹 Login
    const login = async (credentials) => {

        const res = await loginUser(credentials);

        localStorage.setItem("access_token", res.data.access_token);
        localStorage.setItem("refresh_token", res.data.refresh_token);

        setUser(res.data.user);
        setIsAuthenticated(true);

        // ✅ fetch wallet immediately
        try {
            const walletRes = await getWallet();

            setWallet({
                available: walletRes.data.Balance - walletRes.data.ReservedBalance,
                reserved: walletRes.data.ReservedBalance,
                total: walletRes.data.Balance
            });
        } catch (err) {
            console.log("Wallet fetch failed");
        }

        navigate("/auctions");
    };

    // 🔹 Logout
    const logout = async () => {

        try {
            const refreshToken = localStorage.getItem("refresh_token");

            if (refreshToken) {
                await logoutUser(refreshToken);
            }

        } catch (err) {
            console.log("Logout error (safe to ignore)");
        }

        localStorage.removeItem("access_token");
        localStorage.removeItem("refresh_token");

        setUser(null);
        setWallet(null);
        setIsAuthenticated(false);

        navigate("/login");
    };

    return (
        <AuthContext.Provider value={{
            user,
            wallet,
            isAuthenticated,
            login,
            logout,
            loading
        }}>
            {!loading && children}
        </AuthContext.Provider>
    );
};