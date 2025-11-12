import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuth } from '@/stores/auth-store'

export const Route = createFileRoute('/auth/__root')({
  beforeLoad: ({ context }) => {
    // If user is already authenticated, redirect to dashboard
    if (useAuth()) {
      throw redirect({
        to: '/protected/dashboard',
        replace: true,
      })
    }
  },
  component: AuthLayout,
})

function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        {children}
      </div>
    </div>
  )
}