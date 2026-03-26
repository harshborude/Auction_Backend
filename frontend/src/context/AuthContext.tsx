import React, { createContext, useCallback, useContext, useEffect, useState } from 'react'
import { User, Wallet } from '@/types'
import { loginUser, logoutUser, getCurrentUser } from '@/api/auth'
import { getWallet } from '@/api/wallet'

interface AuthContextValue {
  user: User | null
  wallet: Wallet | null
  isAuthenticated: boolean
  isAdmin: boolean
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  refreshWallet: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [wallet, setWallet] = useState<Wallet | null>(null)
  const [loading, setLoading] = useState(true)

  const fetchWallet = useCallback(async () => {
    try {
      const { data } = await getWallet()
      setWallet(data)
    } catch {
      // non-critical
    }
  }, [])

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    if (!token) {
      setLoading(false)
      return
    }
    ;(async () => {
      try {
        const { data } = await getCurrentUser()
        setUser(data)
        await fetchWallet()
      } catch {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
      } finally {
        setLoading(false)
      }
    })()
  }, [fetchWallet])

  const login = useCallback(async (email: string, password: string) => {
    const { data } = await loginUser({ email, password })
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    const { data: me } = await getCurrentUser()
    setUser(me)
    const { data: w } = await getWallet()
    setWallet(w)
  }, [])

  const logout = useCallback(async () => {
    const refreshToken = localStorage.getItem('refresh_token') ?? ''
    try {
      await logoutUser(refreshToken)
    } catch {
      // proceed regardless
    }
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    setUser(null)
    setWallet(null)
  }, [])

  const refreshWallet = useCallback(async () => {
    await fetchWallet()
  }, [fetchWallet])

  return (
    <AuthContext.Provider
      value={{
        user,
        wallet,
        isAuthenticated: !!user,
        isAdmin: user?.Role === 'ADMIN',   // Role is PascalCase
        loading,
        login,
        logout,
        refreshWallet,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
