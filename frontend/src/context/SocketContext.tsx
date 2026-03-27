import React, { createContext, useCallback, useContext, useEffect, useRef } from 'react'
import { WsMessage } from '@/types'
import { useAuth } from './AuthContext'

type MessageHandler = (msg: WsMessage) => void

interface SocketContextValue {
  joinAuction: (id: number) => void
  leaveAuction: (id: number) => void
  subscribe: (handler: MessageHandler) => () => void
}

const SocketContext = createContext<SocketContextValue | null>(null)

export function SocketProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth()
  const wsRef = useRef<WebSocket | null>(null)
  const handlersRef = useRef<Set<MessageHandler>>(new Set())
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const connect = useCallback(() => {
    const token = localStorage.getItem('access_token')
    if (!token) return

    const apiUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
    const wsUrl = apiUrl.replace(/^http/, 'ws')
    const ws = new WebSocket(`${wsUrl}/ws?token=${token}`)
    wsRef.current = ws

    ws.onopen = () => {
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current)
        reconnectTimer.current = null
      }
    }

    ws.onmessage = (event) => {
      try {
        const msg: WsMessage = JSON.parse(event.data)
        handlersRef.current.forEach((h) => h(msg))
      } catch {
        // ignore malformed messages
      }
    }

    ws.onclose = () => {
      wsRef.current = null
      // Reconnect with backoff if still authenticated
      if (localStorage.getItem('access_token')) {
        reconnectTimer.current = setTimeout(connect, 3000)
      }
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [])

  useEffect(() => {
    if (isAuthenticated) {
      connect()
    }
    return () => {
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
      wsRef.current = null
    }
  }, [isAuthenticated, connect])

  const send = useCallback((data: object) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data))
    }
  }, [])

  const joinAuction = useCallback(
    (id: number) => send({ type: 'JOIN_AUCTION', auction_id: id }),
    [send]
  )

  const leaveAuction = useCallback(
    (id: number) => send({ type: 'LEAVE_AUCTION', auction_id: id }),
    [send]
  )

  const subscribe = useCallback((handler: MessageHandler) => {
    handlersRef.current.add(handler)
    return () => handlersRef.current.delete(handler)
  }, [])

  return (
    <SocketContext.Provider value={{ joinAuction, leaveAuction, subscribe }}>
      {children}
    </SocketContext.Provider>
  )
}

export function useSocket() {
  const ctx = useContext(SocketContext)
  if (!ctx) throw new Error('useSocket must be used within SocketProvider')
  return ctx
}
