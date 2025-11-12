import { createFileRoute } from '@tanstack/react-router'
import { useSearch } from '@tanstack/react-router'
import { RegisterForm } from '@/components/auth'

export const Route = createFileRoute('/auth/register')({
  component: Register,
})

function Register() {
  const { redirect } = useSearch({ from: '/auth/register' })

  return (
    <div className="w-full">
      <RegisterForm redirectTo={redirect} />
    </div>
  )
}