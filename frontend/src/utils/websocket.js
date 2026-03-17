class SocketManager {

    constructor(token) {
        this.token = token
        this.connect()
    }

    connect() {

        this.socket = new WebSocket(
            `ws://localhost:8080/ws?token=${this.token}`
        )

        this.socket.onopen = () => {
            console.log("WebSocket Connected")
        }

        this.socket.onclose = () => {
            console.log("WebSocket Disconnected")

            // auto reconnect
            setTimeout(() => this.connect(), 3000)
        }

    }

    send(data) {

        if (this.socket?.readyState === WebSocket.OPEN) {

            this.socket.send(JSON.stringify(data))

        } else {

            console.warn("WebSocket not ready")

        }

    }

    onMessage(callback) {

        if (!this.socket) return

        this.socket.onmessage = (event) => {

            const msg = JSON.parse(event.data)

            callback(msg)

        }

    }

    disconnect() {

        if (this.socket) {

            this.socket.close()

        }

    }

}

export default SocketManager