import { createFileRoute } from '@tanstack/react-router'
import { useNavigate } from '@tanstack/react-router'
import { MFAVerificationForm } from '@/components/auth'

export const Route = createFileRoute('/auth/verify-mfa')({
  component: VerifyMFA,
})

function VerifyMFA() {
  const navigate = useNavigate()

  const handleVerified = () => {
    // Redirect to dashboard after successful MFA verification
    navigate({ to: '/protected/dashboard', replace: true })
  }

  return (
    <div className="w-full">
      <MFAVerificationForm onVerified={handleVerified} />
    </div>
  )
}