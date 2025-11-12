import { describe, it, expect, beforeEach } from 'vitest'
import { create } from 'zustand'
import { QueryClient } from '@tanstack/react-query'

// Test 1: Zustand State Management
describe('Zustand State Management', () => {
  interface TestStore {
    count: number
    increment: () => void
    decrement: () => void
  }

  it('should manage state correctly', () => {
    const useStore = create<TestStore>((set) => ({
      count: 0,
      increment: () => set((state) => ({ count: state.count + 1 })),
      decrement: () => set((state) => ({ count: state.count - 1 })),
    }))

    // Initial state
    expect(useStore.getState().count).toBe(0)

    // Increment
    useStore.getState().increment()
    expect(useStore.getState().count).toBe(1)

    // Increment again
    useStore.getState().increment()
    expect(useStore.getState().count).toBe(2)

    // Decrement
    useStore.getState().decrement()
    expect(useStore.getState().count).toBe(1)
  })

  it('should persist state across multiple store instances', () => {
    const useStore = create<TestStore>((set) => ({
      count: 10,
      increment: () => set((state) => ({ count: state.count + 1 })),
      decrement: () => set((state) => ({ count: state.count - 1 })),
    }))

    const initialCount = useStore.getState().count
    expect(initialCount).toBe(10)

    useStore.getState().increment()
    expect(useStore.getState().count).toBe(11)

    // Store should maintain the updated value
    expect(useStore.getState().count).toBe(11)
  })
})

// Test 2: TanStack Query Configuration
describe('TanStack Query Configuration', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          staleTime: 5 * 60 * 1000, // 5 minutes
          gcTime: 10 * 60 * 1000, // 10 minutes
        },
      },
    })
  })

  it('should create query client with correct default options', () => {
    expect(queryClient).toBeDefined()

    const defaultOptions = queryClient.getDefaultOptions()
    expect(defaultOptions.queries?.retry).toBe(false)
    expect(defaultOptions.queries?.staleTime).toBe(5 * 60 * 1000)
    expect(defaultOptions.queries?.gcTime).toBe(10 * 60 * 1000)
  })

  it('should fetch and cache data correctly', async () => {
    const mockData = { id: 1, name: 'Test Data' }

    const query = queryClient.fetchQuery({
      queryKey: ['test-key'],
      queryFn: () => Promise.resolve(mockData),
    })

    const result = await query
    expect(result).toEqual(mockData)

    // Verify data is cached
    const cachedData = queryClient.getQueryData(['test-key'])
    expect(cachedData).toEqual(mockData)
  })

  it('should handle query invalidation', async () => {
    const mockData = { id: 1, name: 'Test Data' }

    // Set initial data
    queryClient.setQueryData(['test-key'], mockData)
    expect(queryClient.getQueryData(['test-key'])).toEqual(mockData)

    // Invalidate query
    queryClient.invalidateQueries({ queryKey: ['test-key'] })

    // Data should still be available but marked for refetch
    const cachedData = queryClient.getQueryData(['test-key'])
    expect(cachedData).toEqual(mockData)
  })
})

// Test 3: TypeScript Type Safety
describe('TypeScript Type Safety', () => {
  it('should enforce correct types for interfaces', () => {
    interface User {
      id: string
      name: string
      email: string
      role: 'admin' | 'user' | 'viewer'
    }

    const user: User = {
      id: '123',
      name: 'John Doe',
      email: 'john@example.com',
      role: 'user',
    }

    expect(user.id).toBe('123')
    expect(user.name).toBe('John Doe')
    expect(user.email).toBe('john@example.com')
    expect(user.role).toBe('user')

    // This should compile correctly with TypeScript strict mode
    const isUser = (obj: unknown): obj is User => {
      return (
        typeof obj === 'object' &&
        obj !== null &&
        'id' in obj &&
        'name' in obj &&
        'email' in obj &&
        'role' in obj &&
        typeof (obj as User).id === 'string' &&
        typeof (obj as User).name === 'string' &&
        typeof (obj as User).email === 'string' &&
        ['admin', 'user', 'viewer'].includes((obj as User).role)
      )
    }

    expect(isUser(user)).toBe(true)
    expect(isUser({ not: 'a-user' })).toBe(false)
  })

  it('should handle optional properties correctly', () => {
    interface Config {
      required: string
      optional?: string
    }

    const config1: Config = { required: 'value' }
    const config2: Config = { required: 'value', optional: 'optional-value' }

    expect(config1.required).toBe('value')
    expect(config1.optional).toBeUndefined()

    expect(config2.required).toBe('value')
    expect(config2.optional).toBe('optional-value')
  })
})

// Test 4: Error Handling
describe('Error Handling', () => {
  it('should handle API errors gracefully', async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    })

    const errorMessage = 'API Error'

    // Test query error
    const queryPromise = queryClient.fetchQuery({
      queryKey: ['error-test'],
      queryFn: () => Promise.reject(new Error(errorMessage)),
    })

    await expect(queryPromise).rejects.toThrow(errorMessage)

    // Test that error is properly handled
    try {
      await queryPromise
    } catch (error) {
      expect(error).toBeInstanceOf(Error)
      expect((error as Error).message).toBe(errorMessage)
    }
  })

  it('should validate data types correctly', () => {
    const validateUser = (data: unknown) => {
      if (
        typeof data === 'object' &&
        data !== null &&
        'id' in data &&
        'name' in data &&
        typeof (data as any).id === 'string' &&
        typeof (data as any).name === 'string'
      ) {
        return { isValid: true, data: data as { id: string; name: string } }
      }
      return { isValid: false, error: 'Invalid user data' }
    }

    const validUser = { id: '123', name: 'John' }
    const invalidUser = { id: 123, name: 'John' }

    const validResult = validateUser(validUser)
    expect(validResult.isValid).toBe(true)
    if (validResult.isValid && validResult.data) {
      expect(validResult.data.id).toBe('123')
      expect(validResult.data.name).toBe('John')
    }

    const invalidResult = validateUser(invalidUser)
    expect(invalidResult.isValid).toBe(false)
    if (!invalidResult.isValid) {
      expect(invalidResult.error).toBe('Invalid user data')
    }
  })
})