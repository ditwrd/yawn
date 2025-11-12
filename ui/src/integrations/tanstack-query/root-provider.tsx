import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactNode } from 'react'
import { isRecoverableError } from '@/lib/error-handling'

// Enhanced QueryClient configuration with comprehensive error handling, caching strategies, and performance optimizations
export function getContext() {
  const queryClient: QueryClient = new QueryClient({
    defaultOptions: {
      queries: {
        // Enhanced retry configuration with exponential backoff
        retry: (failureCount, error: any) => {
          // Don't retry on 4xx errors (client errors)
          if (error?.status >= 400 && error?.status < 500) {
            return false
          }
          // Don't retry on non-recoverable errors
          if (!isRecoverableError(error)) {
            return false
          }
          // Retry up to 3 times for other errors
          return failureCount < 3
        },
        retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),

        // Enhanced stale time configurations by query type
        staleTime: 5 * 60 * 1000, // 5 minutes default
        gcTime: 10 * 60 * 1000, // 10 minutes garbage collection

        // Smart refetch configurations
        refetchOnWindowFocus: false, // Disabled for better UX
        refetchOnReconnect: true, // Enabled for network recovery
        refetchOnMount: true, // Enabled for stale data

        // Performance optimizations
        refetchIntervalInBackground: false, // Don't refetch when tab is inactive
        structuralSharing: true, // Enable structural sharing for performance

        // Error handling
        throwOnError: false, // Handle errors gracefully

        // Network status handling
        networkMode: 'online', // Only fetch when online

        // Query status and caching
        initialDataUpdatedAt: 0,
        select: undefined, // Can be overridden per query
        placeholderData: undefined, // Can be set per query for optimistic UI
      },
      mutations: {
        // Enhanced retry configuration for mutations
        retry: (failureCount, error: any) => {
          // Don't retry mutations on 4xx errors
          if (error?.status >= 400 && error?.status < 500) {
            return false
          }
          // Don't retry non-recoverable errors
          if (!isRecoverableError(error)) {
            return false
          }
          return failureCount < 2 // Retry mutations up to 2 times
        },
        retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),

        // Error handling
        throwOnError: false, // Handle errors gracefully

        // Network status handling
        networkMode: 'online', // Only mutate when online

        // Enhanced mutation handling with optimistic updates
        onMutate: async (variables) => {
          // Cancel any outgoing refetches for this mutation
          await queryClient.cancelQueries({ queryKey: ['mutation'] })

          // Snapshot previous value
          const previousValue = queryClient.getQueryData(['mutation'])

          // Optimistically update to the new value
          queryClient.setQueryData(['mutation'], variables)

          // Return context with previous value for rollback
          return { previousValue, variables }
        },
        onError: (error, variables, context: any) => {
          // Roll back on error with context
          if (context?.previousValue) {
            queryClient.setQueryData(['mutation'], context.previousValue)
          }

          // Log error for debugging
          console.error('Mutation error:', error)
        },
        onSettled: async (_data, _error, _variables, context) => {
          // Always refetch after mutation settles
          await queryClient.invalidateQueries({ queryKey: ['mutation'] })

          // Invalidate related queries if needed
          if (context?.variables?.invalidateQueries) {
            context.variables.invalidateQueries.forEach((queryKey: any) => {
              queryClient.invalidateQueries({ queryKey })
            })
          }
        },
        onSuccess: (data, variables, _context) => {
          // Auto-update related cache entries on success
          if (variables?.updateCache) {
            variables.updateCache.forEach((update: any) => {
              queryClient.setQueryData(update.queryKey, update.data)
            })
          }
        },
      },
    },
  })

  // Global query cache manipulation utilities
  queryClient.setQueryDefaults(['assets'], {
    staleTime: 2 * 60 * 1000, // 2 minutes for assets
    gcTime: 5 * 60 * 1000, // 5 minutes garbage collection
  })

  queryClient.setQueryDefaults(['projects'], {
    staleTime: 10 * 60 * 1000, // 10 minutes for projects
    gcTime: 15 * 60 * 1000, // 15 minutes garbage collection
  })

  queryClient.setQueryDefaults(['pipelines'], {
    staleTime: 1 * 60 * 1000, // 1 minute for pipelines
    gcTime: 3 * 60 * 1000, // 3 minutes garbage collection
  })

  queryClient.setQueryDefaults(['users'], {
    staleTime: 15 * 60 * 1000, // 15 minutes for users
    gcTime: 30 * 60 * 1000, // 30 minutes garbage collection
  })

  return {
    queryClient,
  }
}

// Enhanced Provider component with error boundaries
export function Provider({
  children,
  queryClient,
}: {
  children: React.ReactNode
  queryClient: QueryClient
}) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}

// Hook for accessing query client utilities
export function useQueryClientUtils() {
  const queryClient = new QueryClient()

  return {
    // Prefetch utilities
    prefetchQuery: queryClient.prefetchQuery.bind(queryClient),
    prefetchInfiniteQuery: queryClient.prefetchInfiniteQuery.bind(queryClient),

    // Cache manipulation
    setQueryData: queryClient.setQueryData.bind(queryClient),
    getQueryData: queryClient.getQueryData.bind(queryClient),
    removeQueries: queryClient.removeQueries.bind(queryClient),
    invalidateQueries: queryClient.invalidateQueries.bind(queryClient),
    refetchQueries: queryClient.refetchQueries.bind(queryClient),

    // Cache inspection
    getQueryCache: () => queryClient.getQueryCache(),
    getMutationCache: () => queryClient.getMutationCache(),
  }
}

// Custom hook for query state management
export function useQueryState() {
  const queryClient = new QueryClient()

  return {
    // Get all queries
    getQueries: () => queryClient.getQueryCache().getAll(),

    // Get active queries
    getActiveQueries: () => queryClient.getQueryCache().findAll({ active: true }),

    // Get inactive queries
    getInactiveQueries: () => queryClient.getQueryCache().findAll({ active: false }),

    // Get stale queries
    getStaleQueries: () => queryClient.getQueryCache().findAll({ stale: true }),

    // Clear cache
    clearCache: () => queryClient.clear(),

    // Get cache size
    getCacheSize: () => queryClient.getQueryCache().getAll().length,
  }
}
