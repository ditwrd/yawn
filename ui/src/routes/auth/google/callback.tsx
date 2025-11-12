import { createFileRoute, useNavigate, useSearch } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Loader2, AlertCircle } from 'lucide-react'
import { useState, useEffect } from 'react'
import { handleGoogleOAuthCallback } from '@/components/auth/GoogleOAuthButton'

export const Route = createFileRoute('/auth/google/callback')({
  component: GoogleOAuthCallback,
})

function GoogleOAuthCallback() {
  const navigate = useNavigate()
  const { code, state, error } = useSearch({ from: '/auth/google/callback' })
  const [isLoading, setIsLoading] = useState(true)
  const [callbackError, setCallbackError] = useState<string | null>(null)

  useEffect(() => {
    const processCallback = async () => {
      try {
        if (error) {
          throw new Error(error)
        }

        if (!code || !state) {
          throw new Error('Missing required OAuth parameters')
        }

        // Process the OAuth callback
        const result = handleGoogleOAuthCallback(code, state)

        if (result.success) {
          // Redirect to dashboard on successful authentication
          navigate({ to: '/protected/dashboard', replace: true })
        } else {
          throw new Error(result.message || 'OAuth authentication failed')
        }
      } catch (error) {
        setCallbackError(error instanceof Error ? error.message : 'Authentication failed')
      } finally {
        setIsLoading(false)
      }
    }

    processCallback()
  }, [code, state, error, navigate])

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <Card className="w-full max-w-md">
          <CardContent className="flex flex-col items-center space-y-4 p-6">
            <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
            <CardTitle className="text-xl">Completing authentication...</CardTitle>
            <p className="text-gray-600 text-center">
              Please wait while we verify your Google account
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (callbackError) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <Card className="w-full max-w-md">
          <CardHeader>
            <div className="flex items-center space-x-2">
              <AlertCircle className="h-5 w-5 text-red-500" />
              <CardTitle className="text-xl text-red-600">Authentication Failed</CardTitle>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert variant="destructive">
              <AlertDescription>{callbackError}</AlertDescription>
            </Alert>

            <div className="space-y-2">
              <p className="text-sm text-gray-600">
                This could be due to:
              </p>
              <ul className="text-sm text-gray-600 list-disc list-inside space-y-1">
                <li>OAuth integration not yet implemented</li>
                <li>Invalid or expired authorization code</li>
                <li>Configuration issues with Google OAuth</li>
                <li>Network connectivity problems</li>
              </ul>
            </div>

            <div className="flex flex-col space-y-2 pt-4">
              <Button
                onClick={() => navigate({ to: '/auth/login', replace: true })}
                className="w-full"
              >
                Back to Login
              </Button>
              <Button
                variant="outline"
                onClick={() => navigate({ to: '/', replace: true })}
                className="w-full"
              >
                Go to Homepage
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return null // Should redirect before reaching here
}