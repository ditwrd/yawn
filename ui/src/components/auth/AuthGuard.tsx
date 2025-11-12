import type { ReactNode } from 'react'
import { Navigate, useLocation } from '@tanstack/react-router'
import { useAuth, useUser } from '@/stores/auth-store'

interface AuthGuardProps {
  children: ReactNode
  requireAuth?: boolean
  roles?: string[]
  redirectTo?: string
}

export function AuthGuard({
  children,
  requireAuth = true,
  roles = [],
  redirectTo = '/login'
}: AuthGuardProps) {
  const location = useLocation()
  const isAuthenticated = useAuth()
  const user = useUser()

  if (requireAuth && !isAuthenticated) {
    return <Navigate to={redirectTo} search={{ redirect: location.href }} />
  }

  // Check role-based access if roles are specified
  if (roles.length > 0 && user) {
    const hasRequiredRole = roles.includes(user.role)
    if (!hasRequiredRole) {
      return <Navigate to="/" /> // TODO: Implement unauthorized page
    }
  }

  return <>{children}</>
}