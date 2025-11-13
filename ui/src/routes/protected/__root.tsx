import { Outlet, createFileRoute, useNavigate } from '@tanstack/react-router'
import { AuthGuard } from '@/components/auth/AuthGuard'
import { AppShell } from '@/components/layout/app-shell'
import { useState } from 'react'

export const Route = createFileRoute('/protected/__root')({
  component: ProtectedLayout,
})

function ProtectedLayout() {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const navigate = useNavigate()

  const handleSidebarToggle = () => {
    setSidebarCollapsed(!sidebarCollapsed)
  }

  return (
    <AuthGuard requireAuth={true}>
      <AppShell
        sidebarCollapsed={sidebarCollapsed}
        onSidebarToggle={handleSidebarToggle}
      >
        <Outlet />
      </AppShell>
    </AuthGuard>
  )
}