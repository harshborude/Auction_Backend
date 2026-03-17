import { createContext, useState, useEffect } from "react";
import { loginUser, getCurrentUser } from "../api/auth";

export const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const initAuth = async () => {
            const token = localStorage.getItem("access_token");
            if (token) {
                try {
                    const res = await getCurrentUser();
                    setUser(res.data);
                    setIsAuthenticated(true);
                } catch (err) {
                    console.error("Session invalid or expired");
                    localStorage.removeItem("access_token");
                    localStorage.removeItem("refresh_token");
                }
            }
            setLoading(false);
        };
        initAuth();
    }, []);

    const login = async (credentials) => {
        const res = await loginUser(credentials);
        localStorage.setItem("access_token", res.data.access_token);
        localStorage.setItem("refresh_token", res.data.refresh_token);
        setUser(res.data.user);
        setIsAuthenticated(true);
    };

    const logout = () => {
        localStorage.removeItem("access_token");
        localStorage.removeItem("refresh_token");
        setUser(null);
        setIsAuthenticated(false);
        window.location.href = "/login";
    };

    return (
        <AuthContext.Provider value={{ user, isAuthenticated, login, logout, loading }}>
            {/* Don't render the app until we know the user's auth status */}
            {!loading && children}
        </AuthContext.Provider>
    );
};