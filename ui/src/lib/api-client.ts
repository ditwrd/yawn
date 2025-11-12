import type {
  // Asset DTOs
  AssetResponse,
  AssetListResponse,
  CreateAssetRequest,
  UpdateAssetRequest,
  AssetDeleteResponse,
  AssetSearchRequest,
  AssetSearchResponse,

  // Project DTOs
  ProjectResponse,
  ProjectListResponse,
  CreateProjectRequest,
  UpdateProjectRequest,
  ProjectDeleteResponse,

  // Pipeline DTOs
  PipelineResponse,
  PipelineListResponse,
  CreatePipelineRequest,
  UpdatePipelineRequest,
  PipelineDeleteResponse,
  PipelineExecutionResponse,
  TriggerExecutionRequest,
  TriggerExecutionResponse,

  // Repository DTOs
  RepositoryResponse,
  GitRepositoryResponse,
  GitRepositoryListResponse,

  // Auth DTOs
  ErrorResponse,
  ValidationErrorResponse,

  // User DTOs
  UserResponse,
  UserListResponse,
  UpdateUserRequest,
} from '@shared/dto'

import { authService } from './auth-service'

const API_BASE_URL = 'http://localhost:8080/api/v1'

// API Error types
export interface APIError extends Error {
  status?: number
  code?: string
  details?: string
  fields?: Record<string, string[] | string>
}

// Request/Response types
export interface ApiResponse<T = any> {
  data: T
  message?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
}

// HTTP Methods
type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

// Request options
interface RequestOptions extends RequestInit {
  params?: Record<string, string | number | boolean>
}

// Request cache for deduplication
const requestCache = new Map<string, Promise<any>>()

/**
 * Enhanced API Client with type safety, error handling, and request deduplication
 */
class ApiClient {
  private baseURL: string
  private defaultHeaders: Record<string, string>

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL
    this.defaultHeaders = {
      'Content-Type': 'application/json',
    }
  }

  /**
   * Build URL with query parameters
   */
  private buildURL(endpoint: string, params?: Record<string, string | number | boolean>): string {
    const url = new URL(endpoint, this.baseURL)

    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          url.searchParams.append(key, String(value))
        }
      })
    }

    return url.toString()
  }

  /**
   * Create cache key for request deduplication
   */
  private getCacheKey(method: HTTPMethod, url: string, body?: any): string {
    const bodyKey = body ? JSON.stringify(body) : ''
    return `${method}:${url}:${bodyKey}`
  }

  /**
   * Handle API errors and transform them to consistent format
   */
  private handleApiError(error: any, response?: Response): APIError {
    const apiError: APIError = new Error(error.message || 'Request failed')

    if (response) {
      apiError.status = response.status
    }

    if (error.code) {
      apiError.code = error.code
    }

    if (error.details) {
      apiError.details = error.details
    }

    if (error.fields) {
      apiError.fields = error.fields
    }

    return apiError
  }

  /**
   * Check if token needs refresh and handle it
   */
  private async ensureValidToken(): Promise<void> {
    const tokens = authService.getAccessToken()

    if (!tokens) {
      throw new Error('No access token available')
    }

    // Token refresh logic would be handled by auth service
    // This is a placeholder for automatic token refresh
  }

  /**
   * Make HTTP request with interceptors and error handling
   */
  private async request<T>(
    method: HTTPMethod,
    endpoint: string,
    options: RequestOptions = {}
  ): Promise<ApiResponse<T>> {
    try {
      // Ensure we have a valid token
      await this.ensureValidToken()

      const { params, ...fetchOptions } = options
      const url = this.buildURL(endpoint, params)

      // Check for request deduplication (only for GET requests)
      const cacheKey = this.getCacheKey(method, url, fetchOptions.body)
      if (method === 'GET' && requestCache.has(cacheKey)) {
        return requestCache.get(cacheKey)!
      }

      // Prepare headers
      const headers = new Headers({
        ...this.defaultHeaders,
        ...fetchOptions.headers,
      })

      // Add authorization header
      const token = authService.getAccessToken()
      if (token) {
        headers.set('Authorization', `Bearer ${token}`)
      }

      // Create the request promise
      const requestPromise = (async () => {
        const response = await fetch(url, {
          method,
          headers,
          credentials: 'include',
          ...fetchOptions,
        })

        // Handle non-successful responses
        if (!response.ok) {
          let errorData
          try {
            errorData = await response.json()
          } catch {
            errorData = { error: `HTTP ${response.status}: ${response.statusText}` }
          }

          throw this.handleApiError(errorData, response)
        }

        // Parse successful response
        const data = await response.json()

        return { data }
      })()

      // Cache GET requests
      if (method === 'GET') {
        requestCache.set(cacheKey, requestPromise)

        // Clear cache after request completes
        requestPromise.finally(() => {
          requestCache.delete(cacheKey)
        })
      }

      return await requestPromise
    } catch (error) {
      throw this.handleApiError(error)
    }
  }

  /**
   * GET request
   */
  async get<T>(endpoint: string, options: Omit<RequestOptions, 'body' | 'method'> = {}): Promise<ApiResponse<T>> {
    return this.request<T>('GET', endpoint, options)
  }

  /**
   * POST request
   */
  async post<T>(endpoint: string, data?: any, options: RequestOptions = {}): Promise<ApiResponse<T>> {
    return this.request<T>('POST', endpoint, {
      ...options,
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  /**
   * PUT request
   */
  async put<T>(endpoint: string, data?: any, options: RequestOptions = {}): Promise<ApiResponse<T>> {
    return this.request<T>('PUT', endpoint, {
      ...options,
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  /**
   * PATCH request
   */
  async patch<T>(endpoint: string, data?: any, options: RequestOptions = {}): Promise<ApiResponse<T>> {
    return this.request<T>('PATCH', endpoint, {
      ...options,
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  /**
   * DELETE request
   */
  async delete<T>(endpoint: string, options: RequestOptions = {}): Promise<ApiResponse<T>> {
    return this.request<T>('DELETE', endpoint, options)
  }

  // Asset Management endpoints
  async getAssets(params?: { page?: number; limit?: number; project_id?: string }): Promise<ApiResponse<AssetListResponse>> {
    return this.get<AssetListResponse>('/assets', { params })
  }

  async getAsset(id: string): Promise<ApiResponse<AssetResponse>> {
    return this.get<AssetResponse>(`/assets/${id}`)
  }

  async createAsset(data: CreateAssetRequest): Promise<ApiResponse<AssetResponse>> {
    return this.post<AssetResponse>('/assets', data)
  }

  async updateAsset(id: string, data: UpdateAssetRequest): Promise<ApiResponse<AssetResponse>> {
    return this.put<AssetResponse>(`/assets/${id}`, data)
  }

  async deleteAsset(id: string): Promise<ApiResponse<AssetDeleteResponse>> {
    return this.delete<AssetDeleteResponse>(`/assets/${id}`)
  }

  async searchAssets(data: AssetSearchRequest): Promise<ApiResponse<AssetSearchResponse>> {
    return this.post<AssetSearchResponse>('/assets/search', data)
  }

  // Project Management endpoints
  async getProjects(params?: { page?: number; limit?: number }): Promise<ApiResponse<ProjectListResponse>> {
    return this.get<ProjectListResponse>('/projects', { params })
  }

  async getProject(id: string): Promise<ApiResponse<ProjectResponse>> {
    return this.get<ProjectResponse>(`/projects/${id}`)
  }

  async createProject(data: CreateProjectRequest): Promise<ApiResponse<ProjectResponse>> {
    return this.post<ProjectResponse>('/projects', data)
  }

  async updateProject(id: string, data: UpdateProjectRequest): Promise<ApiResponse<ProjectResponse>> {
    return this.put<ProjectResponse>(`/projects/${id}`, data)
  }

  async deleteProject(id: string): Promise<ApiResponse<ProjectDeleteResponse>> {
    return this.delete<AssetDeleteResponse>(`/projects/${id}`)
  }

  // Pipeline Management endpoints
  async getPipelines(params?: { page?: number; limit?: number; project_id?: string }): Promise<ApiResponse<PipelineListResponse>> {
    return this.get<PipelineListResponse>('/pipelines', { params })
  }

  async getPipeline(id: string): Promise<ApiResponse<PipelineResponse>> {
    return this.get<PipelineResponse>(`/pipelines/${id}`)
  }

  async createPipeline(data: CreatePipelineRequest): Promise<ApiResponse<PipelineResponse>> {
    return this.post<PipelineResponse>('/pipelines', data)
  }

  async updatePipeline(id: string, data: UpdatePipelineRequest): Promise<ApiResponse<PipelineResponse>> {
    return this.put<PipelineResponse>(`/pipelines/${id}`, data)
  }

  async deletePipeline(id: string): Promise<ApiResponse<PipelineDeleteResponse>> {
    return this.delete<PipelineDeleteResponse>(`/pipelines/${id}`)
  }

  async triggerPipeline(id: string, data?: TriggerExecutionRequest): Promise<ApiResponse<TriggerExecutionResponse>> {
    return this.post<TriggerExecutionResponse>(`/pipelines/${id}/execute`, data)
  }

  async getPipelineExecutions(pipelineId: string, params?: { page?: number; limit?: number }): Promise<ApiResponse<PipelineExecutionResponse[]>> {
    return this.get<PipelineExecutionResponse[]>(`/pipelines/${pipelineId}/executions`, { params })
  }

  // Repository Management endpoints
  async getRepositories(params?: { page?: number; limit?: number; project_id?: string }): Promise<ApiResponse<GitRepositoryListResponse>> {
    return this.get<GitRepositoryListResponse>('/repositories', { params })
  }

  async getRepository(id: string): Promise<ApiResponse<GitRepositoryResponse>> {
    return this.get<GitRepositoryResponse>(`/repositories/${id}`)
  }

  async syncRepository(id: string, data?: { branch?: string; force?: boolean }): Promise<ApiResponse<any>> {
    return this.post<any>(`/repositories/${id}/sync`, data)
  }

  // User Management endpoints
  async getUsers(params?: { page?: number; limit?: number }): Promise<ApiResponse<UserListResponse>> {
    return this.get<UserListResponse>('/users', { params })
  }

  async getUser(id: string): Promise<ApiResponse<UserResponse>> {
    return this.get<UserResponse>(`/users/${id}`)
  }

  async updateUser(id: string, data: UpdateUserRequest): Promise<ApiResponse<UserResponse>> {
    return this.put<UserResponse>(`/users/${id}`, data)
  }

  async deleteUser(id: string): Promise<ApiResponse<any>> {
    return this.delete<any>(`/users/${id}`)
  }
}

// Create and export singleton instance
export const apiClient = new ApiClient()

// Export factory function for testing
export function createApiClient(baseURL?: string): ApiClient {
  return new ApiClient(baseURL)
}

export default apiClient