import { createFileRoute } from '@tanstack/react-router'
import { useSearch } from '@tanstack/react-router'
import { LoginForm } from '@/components/auth'

export const Route = createFileRoute('/auth/login')({
  component: Login,
})

function Login() {
  const { redirect } = useSearch({ from: '/auth/login' })

  return (
    <div className="w-full">
      <LoginForm redirectTo={redirect} />
    </div>
  )
}