import { Outlet, createRootRouteWithContext, useNavigate, useRouterState } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'
import { Suspense, useState } from 'react'

import { AppShell } from '../components/layout/app-shell'
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
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  const handleSidebarToggle = () => {
    setSidebarCollapsed(!sidebarCollapsed)
  }

  return (
    <ErrorBoundary>
      <Suspense
        fallback={
          <div className="flex items-center justify-center min-h-screen bg-background">
            <LoadingSpinner size="lg" />
          </div>
        }
      >
        <AppShell
          sidebarCollapsed={sidebarCollapsed}
          onSidebarToggle={handleSidebarToggle}
        >
          <Outlet />
        </AppShell>

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
  )
}

function RootErrorComponent({ error }: { error: Error }) {
  const navigate = useNavigate()

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-background text-foreground p-8">
      <div className="max-w-md w-full space-y-6 text-center">
        <div className="space-y-2">
          <h1 className="text-4xl font-bold text-destructive">Oops!</h1>
          <h2 className="text-xl font-semibold">Something went wrong</h2>
        </div>

        <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4">
          <p className="text-destructive font-mono text-sm">
            {error.message || 'An unexpected error occurred'}
          </p>
        </div>

        <div className="space-y-3">
          <button
            onClick={() => navigate({ to: '/' })}
            className="w-full px-4 py-2 bg-primary hover:bg-primary/90 text-primary-foreground rounded-lg font-medium transition-colors btn-hover-lift"
          >
            Go Home
          </button>
          <button
            onClick={() => window.location.reload()}
            className="w-full px-4 py-2 bg-secondary hover:bg-secondary/80 text-secondary-foreground rounded-lg font-medium transition-colors btn-hover-lift"
          >
            Refresh Page
          </button>
        </div>

        {import.meta.env.DEV && (
          <details className="mt-6">
            <summary className="cursor-pointer text-muted-foreground hover:text-foreground">
              Error Details (Development)
            </summary>
            <pre className="mt-2 text-xs text-muted-foreground overflow-auto bg-card p-3 rounded border border-border">
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
    <div className="flex flex-col items-center justify-center min-h-screen bg-background text-foreground p-8">
      <div className="max-w-md w-full space-y-6 text-center">
        <div className="space-y-2">
          <h1 className="text-4xl font-bold text-muted-foreground">404</h1>
          <h2 className="text-xl font-semibold">Page Not Found</h2>
        </div>

        <p className="text-muted-foreground">
          The page you're looking for doesn't exist or has been moved.
        </p>

        <button
          onClick={() => navigate({ to: '/' })}
          className="w-full px-4 py-2 bg-primary hover:bg-primary/90 text-primary-foreground rounded-lg font-medium transition-colors btn-hover-lift"
        >
          Go Home
        </button>
      </div>
    </div>
  )
}