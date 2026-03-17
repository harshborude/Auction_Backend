import { createContext, useEffect, useRef, useContext } from "react";
import SocketManager from "../utils/websocket";
import { AuthContext } from "./AuthContext";

export const SocketContext = createContext();

export const SocketProvider = ({ children }) => {
    const { isAuthenticated } = useContext(AuthContext);

    const socketRef = useRef(null);

    useEffect(() => {

        if (isAuthenticated && !socketRef.current) {
            const token = localStorage.getItem("access_token");

            socketRef.current = new SocketManager(token);
        }

        // 🔥 Cleanup ONLY on logout
        if (!isAuthenticated && socketRef.current) {
            socketRef.current.disconnect();
            socketRef.current = null;
        }

    }, [isAuthenticated]);

    const joinAuction = (auctionId) => {
        if (socketRef.current) {
            socketRef.current.send({
                type: "JOIN_AUCTION",
                auction_id: parseInt(auctionId)
            });
        }
    };

    const leaveAuction = (auctionId) => {
        if (socketRef.current) {
            socketRef.current.send({
                type: "LEAVE_AUCTION",
                auction_id: parseInt(auctionId)
            });
        }
    };

    return (
        <SocketContext.Provider value={{
            socket: socketRef.current,
            joinAuction,
            leaveAuction
        }}>
            {children}
        </SocketContext.Provider>
    );
};