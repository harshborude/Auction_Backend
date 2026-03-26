import { Navigate } from 'react-router-dom'
import { useAuth } from '@/context/AuthContext'

interface Props {
  children: React.ReactNode
  requiredRole?: 'ADMIN'
}

export default function ProtectedRoute({ children, requiredRole }: Props) {
  const { isAuthenticated, isAdmin, loading } = useAuth()

  if (loading) return null

  if (!isAuthenticated) return <Navigate to="/login" replace />

  if (requiredRole === 'ADMIN' && !isAdmin) return <Navigate to="/" replace />

  return <>{children}</>
}
