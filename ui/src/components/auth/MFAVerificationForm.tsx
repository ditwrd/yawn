import { useState } from 'react'
import { useAuthActions, useAuthLoading, useAuthError } from '@/stores/auth-store'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Loader2, Shield, Smartphone } from 'lucide-react'

interface MFAVerificationFormProps {
  onVerified?: () => void
  className?: string
}

export function MFAVerificationForm({ onVerified, className }: MFAVerificationFormProps) {
  const [code, setCode] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const { verifyMFACode } = useAuthActions()
  const authLoading = useAuthLoading()
  const authError = useAuthError()

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value.replace(/\D/g, '') // Only allow digits
    if (value.length <= 6) {
      setCode(value)
      setError(null)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (code.length !== 6) {
      setError('Please enter a 6-digit verification code')
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      await verifyMFACode(code)
      onVerified?.()
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Invalid verification code')
    } finally {
      setIsLoading(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleSubmit(e as any)
    }
  }

  return (
    <Card className={`w-full max-w-md mx-auto ${className}`}>
      <CardHeader className="space-y-1 text-center">
        <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-blue-100">
          <Smartphone className="h-6 w-6 text-blue-600" />
        </div>
        <CardTitle className="text-2xl font-bold">Verify your identity</CardTitle>
        <CardDescription>
          Enter the 6-digit code from your authenticator app
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit}>
        <CardContent className="space-y-4">
          {(error || authError) && (
            <Alert variant="destructive">
              <AlertDescription>{error || authError}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <Label htmlFor="mfa-code">Verification Code</Label>
            <Input
              id="mfa-code"
              type="text"
              inputMode="numeric"
              pattern="[0-9]*"
              maxLength={6}
              placeholder="000000"
              value={code}
              onChange={handleInputChange}
              onKeyDown={handleKeyDown}
              className="text-center text-lg tracking-widest"
              disabled={isLoading || authLoading}
              required
              autoFocus
            />
            <p className="text-sm text-gray-500 text-center">
              Open your authenticator app and enter the 6-digit code
            </p>
          </div>

          <div className="bg-blue-50 p-4 rounded-lg">
            <div className="flex items-start space-x-3">
              <Shield className="h-5 w-5 text-blue-600 mt-0.5" />
              <div className="text-sm text-blue-800">
                <p className="font-medium mb-1">Having trouble?</p>
                <ul className="space-y-1 text-blue-700">
                  <li>• Make sure your device's time is synced</li>
                  <li>• Try generating a new code</li>
                  <li>• Check if you're using the correct account</li>
                </ul>
              </div>
            </div>
          </div>
        </CardContent>

        <CardFooter className="flex flex-col space-y-4">
          <Button
            type="submit"
            className="w-full"
            disabled={code.length !== 6 || isLoading || authLoading}
          >
            {(isLoading || authLoading) ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Verifying...
              </>
            ) : (
              'Verify'
            )}
          </Button>

          <div className="text-center text-sm space-y-2">
            <button
              type="button"
              className="text-blue-600 hover:text-blue-500 hover:underline"
              onClick={() => {
                // TODO: Implement backup code recovery
                console.log('Use backup code (placeholder)')
              }}
              disabled={isLoading || authLoading}
            >
              Use a backup code instead
            </button>
            <br />
            <button
              type="button"
              className="text-gray-600 hover:text-gray-500 hover:underline"
              onClick={() => {
                // TODO: Implement contact support
                console.log("Can't access your device ? (placeholder)")
              }}
              disabled={isLoading || authLoading}
            >
              Can't access your device?
            </button>
          </div>
        </CardFooter>
      </form>
    </Card>
  )
}
