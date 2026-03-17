class SocketManager {

    constructor(token) {
        this.token = token;
        this.listeners = [];
        this.shouldReconnect = true;
        this.connect();
    }

    connect() {

        this.socket = new WebSocket(
            `ws://localhost:8080/ws?token=${this.token}`
        );

        this.socket.onopen = () => {
            console.log("WebSocket Connected");
        };

        this.socket.onmessage = (event) => {
            const msg = JSON.parse(event.data);

            this.listeners.forEach(cb => cb(msg));
        };

        this.socket.onclose = () => {
            console.log("WebSocket Disconnected");

            if (this.shouldReconnect) {
                setTimeout(() => this.connect(), 3000);
            }
        };
    }

    send(data) {
        if (this.socket?.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(data));
        } else {
            console.warn("WebSocket not ready");
        }
    }

    onMessage(callback) {
        this.listeners.push(callback);

        // return unsubscribe (IMPORTANT)
        return () => {
            this.listeners = this.listeners.filter(cb => cb !== callback);
        };
    }

    disconnect() {
        this.shouldReconnect = false;

        if (this.socket) {
            this.socket.close();
        }

        this.listeners = [];
    }
}

export default SocketManager;