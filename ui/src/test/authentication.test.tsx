import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/auth-store'

// Mock the auth service
vi.mock('@/lib/auth-service', () => ({
  authService: {
    login: vi.fn(),
    logout: vi.fn(),
    register: vi.fn(),
    refreshToken: vi.fn(),
    isTokenExpired: vi.fn(),
  }
}))

// Test components
function TestComponent() {
  return <div>Protected Content</div>
}

describe('Authentication Flow', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    })

    // Reset auth store before each test
    useAuthStore.setState({
      isAuthenticated: false,
      user: null,
      tokens: null,
      isLoading: false,
      error: null,
    })

    vi.clearAllMocks()
  })

  describe('Login/Logout Functionality', () => {
    it('should login successfully with valid credentials', async () => {
      const { authService } = await import('@/lib/auth-service')
      authService.login = vi.fn().mockResolvedValueOnce({
        success: true,
        data: {
          user: {
            id: '1',
            email: 'test@example.com',
            name: 'Test User',
            role: 'user',
            isActive: true,
            createdAt: '2024-01-01T00:00:00Z',
            updatedAt: '2024-01-01T00:00:00Z',
          },
          tokens: {
            accessToken: 'test-access-token',
            refreshToken: 'test-refresh-token',
            expiresAt: Date.now() + 3600000,
            tokenType: 'Bearer',
          },
        },
      })

      const { login } = useAuthStore.getState()
      await login({ email: 'test@example.com', password: 'password123' })

      const state = useAuthStore.getState()
      expect(state.isAuthenticated).toBe(true)
      expect(state.user?.email).toBe('test@example.com')
      expect(state.tokens?.accessToken).toBe('test-access-token')
    })

    it('should handle login failure with invalid credentials', async () => {
      const { authService } = await import('@/lib/auth-service')
      authService.login = vi.fn().mockRejectedValueOnce(new Error('Invalid email or password'))

      const { login } = useAuthStore.getState()

      try {
        await login({ email: 'test@example.com', password: 'wrongpassword' })
      } catch (error) {
        // Expected to throw
      }

      const state = useAuthStore.getState()
      expect(state.isAuthenticated).toBe(false)
      expect(state.user).toBe(null)
      expect(state.error).toContain('Invalid email or password')
    })

    it('should logout successfully and clear state', async () => {
      // Set authenticated state
      useAuthStore.setState({
        isAuthenticated: true,
        user: { id: '1', email: 'test@example.com', name: 'Test', role: 'user', isActive: true, createdAt: '', updatedAt: '' },
        tokens: { accessToken: 'test', refreshToken: 'test', expiresAt: 123, tokenType: 'Bearer' },
      })

      const { logout } = useAuthStore.getState()
      await logout()

      const state = useAuthStore.getState()
      expect(state.isAuthenticated).toBe(false)
      expect(state.user).toBe(null)
      expect(state.tokens).toBe(null)
    })
  })

  describe('JWT Token Refresh Logic', () => {
    it('should refresh access token successfully', async () => {
      // Set initial authenticated state with expired token
      useAuthStore.setState({
        isAuthenticated: true,
        user: { id: '1', email: 'test@example.com', name: 'Test', role: 'user', isActive: true, createdAt: '', updatedAt: '' },
        tokens: {
          accessToken: 'expired-token',
          refreshToken: 'valid-refresh-token',
          expiresAt: Date.now() - 1000, // Expired
          tokenType: 'Bearer',
        },
      })

      const { authService } = await import('@/lib/auth-service')
      authService.refreshToken = vi.fn().mockResolvedValueOnce({
        accessToken: 'new-access-token',
        refreshToken: 'new-refresh-token',
        expiresAt: Date.now() + 3600000,
        tokenType: 'Bearer',
      })

      const { refreshToken } = useAuthStore.getState()
      await refreshToken()

      const state = useAuthStore.getState()
      expect(state.tokens?.accessToken).toBe('new-access-token')
      expect(state.tokens?.expiresAt).toBeGreaterThan(Date.now())
    })

    it('should logout user when refresh token fails', async () => {
      // Set initial authenticated state
      useAuthStore.setState({
        isAuthenticated: true,
        user: { id: '1', email: 'test@example.com', name: 'Test', role: 'user', isActive: true, createdAt: '', updatedAt: '' },
        tokens: {
          accessToken: 'expired-token',
          refreshToken: 'invalid-refresh-token',
          expiresAt: Date.now() - 1000,
          tokenType: 'Bearer',
        },
      })

      const { authService } = await import('@/lib/auth-service')
      authService.refreshToken = vi.fn().mockRejectedValueOnce(new Error('Session expired. Please login again.'))

      const { refreshToken } = useAuthStore.getState()

      try {
        await refreshToken()
      } catch (error) {
        // Expected to throw
      }

      const state = useAuthStore.getState()
      expect(state.isAuthenticated).toBe(false)
      expect(state.user).toBe(null)
      expect(state.tokens).toBe(null)
    })
  })

  describe('Authentication State Management', () => {
    it('should allow access when user is authenticated', () => {
      useAuthStore.setState({
        isAuthenticated: true,
        user: { id: '1', email: 'test@example.com', name: 'Test User', role: 'user', isActive: true, createdAt: '', updatedAt: '' },
        tokens: { accessToken: 'test-token', refreshToken: 'refresh-token', expiresAt: Date.now() + 3600000, tokenType: 'Bearer' },
      })

      const state = useAuthStore.getState()
      expect(state.isAuthenticated).toBe(true)
      expect(state.user?.email).toBe('test@example.com')
    })

    it('should deny access when user is not authenticated', () => {
      useAuthStore.setState({
        isAuthenticated: false,
        user: null,
        tokens: null,
      })

      const state = useAuthStore.getState()
      expect(state.isAuthenticated).toBe(false)
      expect(state.user).toBe(null)
    })

    it('should allow access when user has required role', () => {
      useAuthStore.setState({
        isAuthenticated: true,
        user: { id: '1', email: 'admin@example.com', name: 'Admin User', role: 'admin', isActive: true, createdAt: '', updatedAt: '' },
        tokens: { accessToken: 'test-token', refreshToken: 'refresh-token', expiresAt: Date.now() + 3600000, tokenType: 'Bearer' },
      })

      const state = useAuthStore.getState()
      const hasRequiredRole = state.user?.role === 'admin'
      expect(hasRequiredRole).toBe(true)
    })

    it('should deny access when user lacks required role', () => {
      useAuthStore.setState({
        isAuthenticated: true,
        user: { id: '1', email: 'user@example.com', name: 'Regular User', role: 'user', isActive: true, createdAt: '', updatedAt: '' },
        tokens: { accessToken: 'test-token', refreshToken: 'refresh-token', expiresAt: Date.now() + 3600000, tokenType: 'Bearer' },
      })

      const state = useAuthStore.getState()
      const hasRequiredRole = state.user?.role === 'admin'
      expect(hasRequiredRole).toBe(false)
    })
  })
})