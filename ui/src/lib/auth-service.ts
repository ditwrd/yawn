import type {
  LoginRequest,
  RegisterRequest,
  RefreshRequest,
  LoginResponse,
  RefreshResponse,
  UserInfo,
  LogoutRequest,
  ErrorResponse
} from '@shared/dto'
import type { User } from '@shared/models'
import type {
  LoginCredentials,
  RegisterData,
  AuthResponse,
  AuthTokens,
  UserWithName
} from '@/types/api'

const API_BASE_URL = 'http://localhost:8080/api/v1'

class AuthService {
  private accessToken: string | null = null

  constructor() {
    // Load access token from memory/storage on initialization
    if (typeof window !== 'undefined') {
      this.accessToken = sessionStorage.getItem('accessToken')
    }
  }

  /**
   * Set the access token in memory and sessionStorage
   */
  private setAccessToken(token: string | null) {
    this.accessToken = token
    if (typeof window !== 'undefined') {
      if (token) {
        sessionStorage.setItem('accessToken', token)
      } else {
        sessionStorage.removeItem('accessToken')
      }
    }
  }

  /**
   * Get the current access token
   */
  public getAccessToken(): string | null {
    return this.accessToken
  }

  /**
   * Make authenticated API requests with automatic token injection
   */
  private async authenticatedRequest<T>(
    url: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers = new Headers(options.headers)

    // Set content type if not already set
    if (!headers.has('Content-Type') && (options.body || options.method !== 'GET')) {
      headers.set('Content-Type', 'application/json')
    }

    // Add authorization header if we have a token
    if (this.accessToken) {
      headers.set('Authorization', `Bearer ${this.accessToken}`)
    }

    const response = await fetch(`${API_BASE_URL}${url}`, {
      ...options,
      headers,
      credentials: 'include', // Include httpOnly cookies
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`)
    }

    return response.json()
  }

  /**
   * Login with email and password
   */
  async login(credentials: LoginCredentials): Promise<AuthResponse> {
    try {
      // Transform credentials to match API format (LoginRequest)
      const loginRequest: LoginRequest = {
        email: credentials.email,
        password: credentials.password
      }

      // For login, don't use authenticatedRequest since we don't have a token yet
      const response = await fetch(`${API_BASE_URL}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(loginRequest),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        const error = errorData as ErrorResponse
        throw new Error(error.error || `HTTP ${response.status}: ${response.statusText}`)
      }

      const data: LoginResponse = await response.json()

      // Transform API response to match our AuthResponse format
      const userWithName: UserWithName = {
        ...data.user,
        name: data.user.email?.split('@')[0] || 'User' // Use email prefix as name for now
      }

      const authResponse: AuthResponse = {
        success: true,
        data: {
          user: userWithName,
          tokens: {
            accessToken: data.access_token,
            refreshToken: data.refresh_token,
            expiresAt: Date.now() + (data.expires_in * 1000),
            tokenType: data.token_type,
          }
        }
      }

      // Store access token in memory
      this.setAccessToken(authResponse.data.tokens.accessToken)

      return authResponse
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Login failed'
      throw new Error(errorMessage)
    }
  }

  /**
   * Register a new user
   */
  async register(data: RegisterData): Promise<AuthResponse> {
    try {
      // Transform registration data to match API format (RegisterRequest)
      const registerRequest: RegisterRequest = {
        email: data.email,
        password: data.password
        // Note: API doesn't seem to expect name field based on DTO
      }

      // For register, don't use authenticatedRequest since we don't have a token yet
      const response = await fetch(`${API_BASE_URL}/auth/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(registerRequest),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        const error = errorData as ErrorResponse
        throw new Error(error.error || `HTTP ${response.status}: ${response.statusText}`)
      }

      const userData: UserInfo = await response.json()

      // Transform API response to match our AuthResponse format
      const userWithName: UserWithName = {
        ...userData,
        name: userData.email?.split('@')[0] || 'User' // Use email prefix as name for now
      }

      const authResponse: AuthResponse = {
        success: true,
        data: {
          user: userWithName,
          tokens: {
            accessToken: '', // Register doesn't return tokens, user needs to login
            refreshToken: '',
            expiresAt: 0,
            tokenType: 'Bearer',
          }
        }
      }

      // Registration doesn't authenticate user - they need to login separately

      return authResponse
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Registration failed'
      throw new Error(errorMessage)
    }
  }

  /**
   * Logout the user
   */
  async logout(): Promise<void> {
    try {
      // Call backend logout endpoint with access token
      if (this.accessToken) {
        const logoutRequest: LogoutRequest = {
          access_token: this.accessToken
        }

        await fetch(`${API_BASE_URL}/auth/logout`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.accessToken}`,
          },
          credentials: 'include',
          body: JSON.stringify(logoutRequest),
        })
      }
    } catch (error) {
      // Continue with local logout even if backend call fails
      console.warn('Logout API call failed:', error)
    } finally {
      // Clear local state
      this.setAccessToken(null)
    }
  }

  /**
   * Refresh the access token using the refresh token (stored in httpOnly cookie)
   */
  async refreshToken(refreshToken: string): Promise<AuthTokens> {
    try {
      const refreshRequest: RefreshRequest = {
        refresh_token: refreshToken
      }

      const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(refreshRequest),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        const error = errorData as ErrorResponse
        throw new Error(error.error || `HTTP ${response.status}: ${response.statusText}`)
      }

      const data: RefreshResponse = await response.json()

      // Transform API response
      const newTokens: AuthTokens = {
        accessToken: data.access_token,
        refreshToken: refreshToken, // Keep existing refresh token
        expiresAt: Date.now() + (data.expires_in * 1000),
        tokenType: data.token_type,
      }

      // Update access token in memory
      this.setAccessToken(newTokens.accessToken)

      return newTokens
    } catch (error) {
      // If refresh fails, clear access token
      this.setAccessToken(null)
      throw new Error('Session expired. Please login again.')
    }
  }

  /**
   * Request password reset
   */
  async requestPasswordReset(email: string): Promise<void> {
    await this.authenticatedRequest<void>('/auth/password-reset/request', {
      method: 'POST',
      body: JSON.stringify({ email }),
    })
  }

  /**
   * Reset password with token
   */
  async resetPassword(token: string, newPassword: string): Promise<void> {
    await this.authenticatedRequest<void>('/auth/password-reset/confirm', {
      method: 'POST',
      body: JSON.stringify({ token, password: newPassword }),
    })
  }

  /**
   * Enable MFA for the user
   */
  async enableMFA(): Promise<{ secret: string; qrCode: string }> {
    const response = await this.authenticatedRequest<{ data: { secret: string; qrCode: string } }>('/auth/mfa/enable', {
      method: 'POST',
    })
    return response.data
  }

  /**
   * Verify and enable MFA with TOTP code
   */
  async verifyAndEnableMFA(code: string): Promise<void> {
    await this.authenticatedRequest<void>('/auth/mfa/verify', {
      method: 'POST',
      body: JSON.stringify({ code }),
    })
  }

  /**
   * Disable MFA for the user
   */
  async disableMFA(password: string): Promise<void> {
    await this.authenticatedRequest<void>('/auth/mfa/disable', {
      method: 'POST',
      body: JSON.stringify({ password }),
    })
  }

  /**
   * Verify MFA code during login
   */
  async verifyMFACode(code: string): Promise<AuthResponse> {
    const response = await this.authenticatedRequest<AuthResponse>('/auth/mfa/verify-login', {
      method: 'POST',
      body: JSON.stringify({ code }),
    })

    if (response.success && response.data) {
      // Store access token in memory
      this.setAccessToken(response.data.tokens.accessToken)
    }

    return response
  }

  /**
   * Get current user profile
   */
  async getCurrentUser(): Promise<User> {
    const response = await this.authenticatedRequest<{ data: User }>('/auth/me')
    return response.data
  }

  /**
   * Update user profile
   */
  async updateProfile(data: Partial<User>): Promise<User> {
    const response = await this.authenticatedRequest<{ data: User }>('/auth/profile', {
      method: 'PATCH',
      body: JSON.stringify(data),
    })

    // Return updated user data - let the store handle the update

    return response.data
  }

  /**
   * Change password
   */
  async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    await this.authenticatedRequest<void>('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify({
        currentPassword,
        newPassword,
      }),
    })
  }

  /**
   * Check if access token is expired or will expire soon
   */
  isTokenExpired(expiresAt: number, bufferMs: number = 5 * 60 * 1000): boolean {
    return Date.now() >= (expiresAt - bufferMs) // 5-minute buffer
  }
}

// Export singleton instance
export const authService = new AuthService()
export default authService