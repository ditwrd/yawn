import {
  QueryClient,
  useQuery,
  useMutation
} from '@tanstack/react-query'
import type {
  QueryKey,
  QueryFunction,
  QueryObserverOptions,
  UseQueryOptions,
  UseMutationOptions,
  InvalidateQueryFilters
} from '@tanstack/react-query'
import { queryKeys } from '@/types/api'
import type { ApiResponse, PaginatedResponse } from '@/types/api'

// Generic API fetcher with error handling
export async function apiFetch<T>(
  url: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> {
  try {
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({}))
      throw new Error(error.message || `HTTP error! status: ${response.status}`)
    }

    return await response.json()
  } catch (error) {
    console.error('API fetch error:', error)
    throw error
  }
}

// Enhanced query wrapper with type safety and default options
export function useApiQuery<T>(
  queryKey: QueryKey,
  queryFn: QueryFunction<T>,
  options?: Omit<UseQueryOptions<T>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey,
    queryFn,
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
    retry: (failureCount, error: any) => {
      if (error?.status >= 400 && error?.status < 500) return false
      return failureCount < 3
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    ...options,
  })
}

// Enhanced mutation wrapper with optimistic updates
export function useApiMutation<TData, TError, TVariables>(
  mutationFn: (variables: TVariables) => Promise<TData>,
  options?: UseMutationOptions<TData, TError, TVariables>
) {
  return useMutation({
    mutationFn,
    retry: (failureCount, error: any) => {
      if (error?.status >= 400 && error?.status < 500) return false
      return failureCount < 2
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),
    ...options,
  })
}

// Prefetching utility for better UX
export function prefetchQuery<T>(
  queryClient: QueryClient,
  queryKey: QueryKey,
  queryFn: QueryFunction<T>,
  options?: Omit<QueryObserverOptions<T>, 'queryKey' | 'queryFn'>
) {
  return queryClient.prefetchQuery({
    queryKey,
    queryFn,
    staleTime: 5 * 60 * 1000, // 5 minutes
    ...options,
  })
}

// Cache invalidation utilities
export const invalidateQueries = (
  queryClient: QueryClient,
  filters?: InvalidateQueryFilters
) => {
  return queryClient.invalidateQueries(filters)
}

// Specific invalidation helpers
export const invalidateUserQueries = (queryClient: QueryClient) => {
  return queryClient.invalidateQueries({ queryKey: queryKeys.users.all })
}

export const invalidateProjectQueries = (queryClient: QueryClient) => {
  return queryClient.invalidateQueries({ queryKey: queryKeys.projects.all })
}

export const invalidateAuthQueries = (queryClient: QueryClient) => {
  return queryClient.invalidateQueries({ queryKey: queryKeys.auth.all })
}

// Optimistic update helpers
export const createOptimisticUpdate = <T>(
  queryClient: QueryClient,
  queryKey: QueryKey,
  updateFn: (old: T | undefined) => T
) => {
  const previousValue = queryClient.getQueryData<T>(queryKey)

  queryClient.setQueryData<T>(queryKey, updateFn(previousValue))

  return () => {
    queryClient.setQueryData<T>(queryKey, previousValue)
  }
}

// Pagination query helper
export function usePaginatedQuery<T>(
  baseQueryKey: QueryKey,
  fetchPage: (page: number, limit: number) => Promise<PaginatedResponse<T>>,
  page: number = 1,
  limit: number = 10,
  options?: UseQueryOptions<PaginatedResponse<T>>
) {
  return useApiQuery(
    [...baseQueryKey, page, limit],
    () => fetchPage(page, limit),
    {
      placeholderData: (previousData) => previousData, // Keep previous data while loading new page
      ...options,
    }
  )
}

// Infinite query helper for endless scrolling (to be implemented with useInfiniteQuery)
// export function useInfiniteQuery<T>(
//   baseQueryKey: QueryKey,
//   fetchPage: (pageParam: number) => Promise<{ data: T[]; nextPage?: number }>,
//   options?: UseInfiniteQueryOptions<{ data: T[]; nextPage?: number }>
// ) {
//   // TODO: Implement with useInfiniteQuery from TanStack Query
// }

// Background refetch utilities
export const refetchQueriesInBackground = (
  queryClient: QueryClient,
  filters?: InvalidateQueryFilters
) => {
  return queryClient.refetchQueries({
    ...filters,
    type: 'active',
  })
}

// Query health check
export const getQueryHealth = (queryClient: QueryClient) => {
  const cache = queryClient.getQueryCache()
  const queries = cache.getAll()

  return {
    totalQueries: queries.length,
    activeQueries: queries.filter(q => q.getObserversCount() > 0).length,
    staleQueries: queries.filter(q => q.isStale()).length,
    errorQueries: queries.filter(q => q.state.status === 'error').length,
  }
}

// Query cache cleanup utility
export const cleanupQueries = (
  queryClient: QueryClient,
  maxAge: number = 30 * 60 * 1000 // 30 minutes
) => {
  const cache = queryClient.getQueryCache()
  const queries = cache.getAll()

  queries.forEach(query => {
    if (query.state.dataUpdatedAt < Date.now() - maxAge) {
      cache.remove(query)
    }
  })
}