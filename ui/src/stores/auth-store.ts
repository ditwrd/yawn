import { create } from 'zustand'
import { persist, subscribeWithSelector } from 'zustand/middleware'
import { authService } from '@/lib/auth-service'
import type { User, AuthTokens, LoginCredentials, RegisterData } from '@/types/api'

interface AuthState {
  // State
  isAuthenticated: boolean
  user: User | null
  tokens: AuthTokens | null
  isLoading: boolean
  error: string | null

  // Actions
  login: (credentials: LoginCredentials) => Promise<void>
  logout: () => Promise<void>
  register: (data: RegisterData) => Promise<void>
  refreshToken: () => Promise<void>
  updateProfile: (data: Partial<User>) => Promise<void>
  verifyMFACode: (code: string) => Promise<void>
  setLoading: (loading: boolean) => void
  clearError: () => void
  setError: (error: string) => void

  // Token management
  checkTokenExpiry: () => Promise<void>
  autoRefreshToken: () => NodeJS.Timeout | null
}

export const useAuthStore = create<AuthState>()(
  // subscribeWithSelector(
    // persist(
      (set, get) => ({
        // Initial state
        isAuthenticated: false,
        user: null,
        tokens: null,
        isLoading: false,
        error: null,

        // Actions
        login: async (credentials: LoginCredentials) => {
          set({ isLoading: true, error: null })

          try {
            const response = await authService.login(credentials)

            if (response.success && response.data) {
              set({
                isAuthenticated: true,
                user: response.data.user,
                tokens: response.data.tokens,
                isLoading: false,
                error: null,
              })
            } else {
              throw new Error('Login failed')
            }
          } catch (error) {
            set({
              error: error instanceof Error ? error.message : 'Login failed',
              isLoading: false,
            })
            throw error
          }
        },

        logout: async () => {
          set({ isLoading: true })

          try {
            await authService.logout()
          } catch (error) {
            console.warn('Logout API call failed:', error)
          } finally {
            set({
              isAuthenticated: false,
              user: null,
              tokens: null,
              error: null,
              isLoading: false,
            })
          }
        },

        register: async (data: RegisterData) => {
          set({ isLoading: true, error: null })

          try {
            const response = await authService.register(data)

            if (response.success && response.data) {
              // Registration doesn't authenticate user, just show success
              set({
                isAuthenticated: false,
                user: null,
                tokens: null,
                isLoading: false,
                error: null,
              })
            } else {
              throw new Error('Registration failed')
            }
          } catch (error) {
            set({
              error: error instanceof Error ? error.message : 'Registration failed',
              isLoading: false,
            })
            throw error
          }
        },

        refreshToken: async () => {
          const { tokens } = get()
          if (!tokens?.refreshToken) return

          try {
            const newTokens = await authService.refreshToken(tokens.refreshToken)

            set({
              tokens: newTokens
            })
          } catch (error) {
            // If refresh fails, log out the user
            console.error('Token refresh failed:', error)
            get().logout()
            throw error
          }
        },

        updateProfile: async (data: Partial<User>) => {
          const { user } = get()
          if (!user) return

          set({ isLoading: true, error: null })

          try {
            const updatedUser = await authService.updateProfile(data)

            set({
              user: updatedUser,
              isLoading: false,
              error: null,
            })
          } catch (error) {
            set({
              error: error instanceof Error ? error.message : 'Profile update failed',
              isLoading: false,
            })
            throw error
          }
        },

        verifyMFACode: async (code: string) => {
          set({ isLoading: true, error: null })

          try {
            const response = await authService.verifyMFACode(code)

            if (response.success && response.data) {
              set({
                isAuthenticated: true,
                user: response.data.user,
                tokens: response.data.tokens,
                isLoading: false,
                error: null,
              })
            } else {
              throw new Error('MFA verification failed')
            }
          } catch (error) {
            set({
              error: error instanceof Error ? error.message : 'MFA verification failed',
              isLoading: false,
            })
            throw error
          }
        },

        setLoading: (loading: boolean) => {
          set({ isLoading: loading })
        },

        clearError: () => {
          set({ error: null })
        },

        setError: (error: string) => {
          set({ error })
        },

        checkTokenExpiry: async () => {
          const { tokens, refreshToken, isLoading } = get()

          // Prevent refresh if already loading to avoid infinite loops
          if (isLoading || !tokens) return

          if (authService.isTokenExpired(tokens.expiresAt)) {
            try {
              await refreshToken()
            } catch (error) {
              console.error('Auto token refresh failed:', error)
            }
          }
        },

        autoRefreshToken: () => {
          const { tokens } = get()

          if (!tokens) return null

          const timeUntilExpiry = tokens.expiresAt - Date.now()
          const refreshBuffer = 5 * 60 * 1000 // 5 minutes before expiry
          const refreshDelay = Math.max(0, timeUntilExpiry - refreshBuffer)

          return setTimeout(() => {
            get().checkTokenExpiry()
          }, refreshDelay)
        },
      })
      // {
      //   name: 'auth-store',
      //   partialize: (state) => ({
      //     isAuthenticated: state.isAuthenticated,
      //     user: state.user,
      //     tokens: state.tokens,
      //   }),
      // }
    // )
  // )
)

// Auto-refresh token disabled temporarily to fix infinite loop
// TODO: Re-enable once circular dependency is resolved
if (typeof window !== 'undefined') {
  // Check token expiry on store initialization only
  // useAuthStore.getState().checkTokenExpiry()
}

// Selectors for optimized re-renders - using shallow comparison to prevent infinite loops
export const useAuth = () => useAuthStore((state) => state.isAuthenticated)
export const useUser = () => useAuthStore((state) => state.user)
export const useAuthTokens = () => useAuthStore((state) => state.tokens)
export const useAuthLoading = () => useAuthStore((state) => state.isLoading)
export const useAuthError = () => useAuthStore((state) => state.error)

// Memoize actions to prevent infinite loops
export const useAuthActions = () => {
  const login = useAuthStore((state) => state.login)
  const logout = useAuthStore((state) => state.logout)
  const register = useAuthStore((state) => state.register)
  const refreshToken = useAuthStore((state) => state.refreshToken)
  const updateProfile = useAuthStore((state) => state.updateProfile)
  const verifyMFACode = useAuthStore((state) => state.verifyMFACode)
  const setLoading = useAuthStore((state) => state.setLoading)
  const clearError = useAuthStore((state) => state.clearError)
  const setError = useAuthStore((state) => state.setError)

  return {
    login,
    logout,
    register,
    refreshToken,
    updateProfile,
    verifyMFACode,
    setLoading,
    clearError,
    setError,
  }
}