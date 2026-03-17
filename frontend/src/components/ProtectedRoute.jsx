import { useContext } from "react";
import { AuthContext } from "../context/AuthContext";
import { Navigate, useLocation } from "react-router-dom";

function ProtectedRoute({ children }) {

    const { isAuthenticated, loading } = useContext(AuthContext);
    const location = useLocation();

    // 🔹 Prevent flicker + unnecessary renders
    if (loading) {
        return <div style={{ padding: "20px" }}>Checking authentication...</div>;
    }

    // 🔹 Redirect with return path
    if (!isAuthenticated) {
        return <Navigate to="/login" state={{ from: location }} replace />;
    }

    return children;
}

export default ProtectedRoute;