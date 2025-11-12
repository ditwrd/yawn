import { Outlet, createRootRouteWithContext, useNavigate, useRouterState } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'
import { Suspense } from 'react'

import Header from '../components/Header'
import { ErrorBoundary } from '../components/error/ErrorBoundary'
import { LoadingSpinner } from '../components/ui/LoadingSpinner'

import TanStackQueryDevtools from '../integrations/tanstack-query/devtools'

import type { QueryClient } from '@tanstack/react-query'

interface MyRouterContext {
  queryClient: QueryClient
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
  component: RootComponent,
  errorComponent: RootErrorComponent,
  notFoundComponent: RootNotFoundComponent,
})

function RootComponent() {
  const isLoading = useRouterState().isLoading

  return (
    <div className="min-h-screen bg-black text-white">
      <ErrorBoundary>
        <Suspense
          fallback={
            <div className="flex items-center justify-center min-h-screen">
              <LoadingSpinner size="lg" />
            </div>
          }
        >
          <Header />
          <main className="container mx-auto px-4 py-8">
            <Outlet />
          </main>

          {/* Global loading indicator */}
          {isLoading && (
            <div className="fixed top-4 right-4 z-50">
              <LoadingSpinner size="sm" />
            </div>
          )}

          {/* Dev tools - only in development */}
          {import.meta.env.DEV && (
            <TanStackDevtools
              config={{
                position: 'bottom-right',
              }}
              plugins={[
                {
                  name: 'Tanstack Router',
                  render: <TanStackRouterDevtoolsPanel />,
                },
                TanStackQueryDevtools,
              ]}
            />
          )}
        </Suspense>
      </ErrorBoundary>
    </div>
  )
}

function RootErrorComponent({ error }: { error: Error }) {
  const navigate = useNavigate()

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-black text-white p-8">
      <div className="max-w-md w-full space-y-6 text-center">
        <div className="space-y-2">
          <h1 className="text-4xl font-bold text-red-500">Oops!</h1>
          <h2 className="text-xl font-semibold">Something went wrong</h2>
        </div>

        <div className="bg-red-900/20 border border-red-800 rounded-lg p-4">
          <p className="text-red-300 font-mono text-sm">
            {error.message || 'An unexpected error occurred'}
          </p>
        </div>

        <div className="space-y-3">
          <button
            onClick={() => navigate({ to: '/' })}
            className="w-full px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded-lg font-medium transition-colors"
          >
            Go Home
          </button>
          <button
            onClick={() => window.location.reload()}
            className="w-full px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg font-medium transition-colors"
          >
            Refresh Page
          </button>
        </div>

        {import.meta.env.DEV && (
          <details className="mt-6">
            <summary className="cursor-pointer text-gray-400 hover:text-white">
              Error Details (Development)
            </summary>
            <pre className="mt-2 text-xs text-gray-500 overflow-auto bg-gray-900 p-3 rounded border border-gray-800">
              {error.stack}
            </pre>
          </details>
        )}
      </div>
    </div>
  )
}

function RootNotFoundComponent() {
  const navigate = useNavigate()

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-black text-white p-8">
      <div className="max-w-md w-full space-y-6 text-center">
        <div className="space-y-2">
          <h1 className="text-4xl font-bold text-gray-400">404</h1>
          <h2 className="text-xl font-semibold">Page Not Found</h2>
        </div>

        <p className="text-gray-400">
          The page you're looking for doesn't exist or has been moved.
        </p>

        <button
          onClick={() => navigate({ to: '/' })}
          className="w-full px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded-lg font-medium transition-colors"
        >
          Go Home
        </button>
      </div>
    </div>
  )
}