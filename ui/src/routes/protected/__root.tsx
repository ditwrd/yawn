import { Outlet, createFileRoute } from '@tanstack/react-router'
import { AuthGuard } from '@/components/auth/AuthGuard'

export const Route = createFileRoute('/protected/__root')({
  component: ProtectedLayout,
})

function ProtectedLayout() {
  return (
    <AuthGuard requireAuth={true}>
      <div className="space-y-6">
        <header className="border-b border-gray-800 pb-4">
          <h1 className="text-2xl font-bold text-purple-400">Protected Area</h1>
          <p className="text-gray-400">This area requires authentication to access</p>
        </header>
        <Outlet />
      </div>
    </AuthGuard>
  )
}