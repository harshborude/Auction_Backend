import { createContext, useEffect, useState, useContext } from "react";
import SocketManager from "../utils/websocket";
import { AuthContext } from "./AuthContext";

export const SocketContext = createContext();

export const SocketProvider = ({ children }) => {
    const { isAuthenticated } = useContext(AuthContext);
    const [socket, setSocket] = useState(null);

    useEffect(() => {
        let manager = null;

        if (isAuthenticated) {
            const token = localStorage.getItem("access_token");
            manager = new SocketManager(token);
            setSocket(manager);
        }

        // Cleanup: Disconnect when user logs out or unmounts
        return () => {
            if (manager) {
                manager.disconnect();
            }
        };
    }, [isAuthenticated]);

    const joinAuction = (auctionId) => {
        if (socket) {
            socket.send({ type: "JOIN_AUCTION", auction_id: parseInt(auctionId) });
        }
    };

    const leaveAuction = (auctionId) => {
        if (socket) {
            socket.send({ type: "LEAVE_AUCTION", auction_id: parseInt(auctionId) });
        }
    };

    return (
        <SocketContext.Provider value={{ socket, joinAuction, leaveAuction }}>
            {children}
        </SocketContext.Provider>
    );
};