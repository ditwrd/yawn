import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Loader2 } from 'lucide-react'

interface GoogleOAuthButtonProps {
  mode: 'login' | 'register'
  className?: string
  disabled?: boolean
}

export function GoogleOAuthButton({ mode, className, disabled }: GoogleOAuthButtonProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleGoogleAuth = async () => {
    setIsLoading(true)
    setError(null)

    try {
      // TODO: Implement real Google OAuth integration
      // For now, this is a placeholder that simulates the OAuth flow

      // Step 1: Redirect to Google OAuth consent screen
      const googleAuthUrl = new URL('https://accounts.google.com/o/oauth2/v2/auth')
      googleAuthUrl.searchParams.set('client_id', 'YOUR_GOOGLE_CLIENT_ID')
      googleAuthUrl.searchParams.set('redirect_uri', `${window.location.origin}/auth/google/callback`)
      googleAuthUrl.searchParams.set('response_type', 'code')
      googleAuthUrl.searchParams.set('scope', 'openid email profile')
      googleAuthUrl.searchParams.set('state', generateRandomState())

      console.log(`Google OAuth ${mode} - Redirecting to:`, googleAuthUrl.toString())

      // Simulate OAuth flow for demonstration
      setTimeout(() => {
        // Simulate successful OAuth callback
        console.log(`Google OAuth ${mode} simulated success`)
        setIsLoading(false)

        // TODO: Handle OAuth callback, exchange code for tokens, and create/update user
        alert(`Google OAuth ${mode} is not yet implemented. This is a placeholder.`)
      }, 2000)

    } catch (error) {
      setError(error instanceof Error ? error.message : `Google ${mode} failed`)
      setIsLoading(false)
    }
  }

  const generateRandomState = () => {
    return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15)
  }

  return (
    <div className={className}>
      {error && (
        <Alert variant="destructive" className="mb-4">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <Button
        type="button"
        variant="outline"
        className="w-full"
        disabled={disabled || isLoading}
        onClick={handleGoogleAuth}
      >
        {isLoading ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            Connecting to Google...
          </>
        ) : (
          <>
            <svg className="mr-2 h-4 w-4" viewBox="0 0 24 24">
              <path
                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                fill="#4285F4"
              />
              <path
                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                fill="#34A853"
              />
              <path
                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                fill="#FBBC05"
              />
              <path
                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                fill="#EA4335"
              />
            </svg>
            Continue with Google
          </>
        )}
      </Button>

      <div className="mt-2 text-xs text-gray-500 text-center">
        <p>Google OAuth integration is coming soon</p>
        <p>This button currently demonstrates the UI flow</p>
      </div>
    </div>
  )
}

// TODO: Implement OAuth callback route handler
// This would handle the redirect from Google after user consent
export function handleGoogleOAuthCallback(code: string, state: string) {
  // Exchange authorization code for access tokens
  // Get user profile from Google
  // Create or update user in your system
  // Issue JWT tokens and authenticate user
  console.log('Google OAuth callback:', { code, state })

  // Placeholder implementation
  return {
    success: false,
    message: 'Google OAuth integration not yet implemented'
  }
}