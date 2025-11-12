import { createFileRoute } from '@tanstack/react-router'
import { PasswordResetForm } from '@/components/auth'

export const Route = createFileRoute('/auth/forgot-password')({
  component: ForgotPassword,
})

function ForgotPassword() {
  return (
    <div className="w-full">
      <PasswordResetForm />
    </div>
  )
}