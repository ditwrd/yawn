// Import shared types from generated DTOs and models
import type {
  // Auth DTOs
  LoginRequest,
  RegisterRequest,
  LoginResponse,
  RefreshResponse,
  UserInfo,
  // Project DTOs
  ProjectResponse,
  ProjectMemberInfo,
  CreateProjectRequest,
  UpdateProjectRequest,
  ProjectListResponse,
  // User DTOs
  UserResponse,
  UpdateUserRequest,
  // Asset DTOs
  AssetResponse,
  AssetListResponse,
  CreateAssetRequest,
  UpdateAssetRequest,
  // Pipeline DTOs
  PipelineResponse,
  PipelineListResponse,
  CreatePipelineRequest,
  UpdatePipelineRequest,
  // Error DTOs
  ErrorResponse,
  ValidationErrorResponse
} from '@shared/dto'

import type {
  // Domain Models
  User,
  Project,
  Asset,
  // Enums and constants
  UserRole
} from '@shared/models'

// Base API response types - extending shared error types
export interface ApiResponse<T = unknown> {
  data: T
  success: boolean
  message?: string
  errors?: string[]
}

export interface PaginatedResponse<T> {
  data: T[]
  pagination: {
    page: number
    limit: number
    total: number
    totalPages: number
    hasNext: boolean
    hasPrev: boolean
  }
}

// Re-export shared types for convenience
export type {
  LoginRequest,
  RegisterRequest,
  RefreshResponse,
  UserInfo,
  // User types
  User,
  UserResponse,
  UpdateUserRequest,
  // Project types
  Project,
  ProjectResponse,
  ProjectMemberInfo as ProjectMember,
  CreateProjectRequest as CreateProjectData,
  UpdateProjectRequest as UpdateProjectData,
  ProjectListResponse,
  // Asset types
  Asset,
  AssetResponse,
  AssetListResponse,
  CreateAssetRequest,
  UpdateAssetRequest,
  // Pipeline types
  PipelineResponse,
  PipelineListResponse,
  CreatePipelineRequest,
  UpdatePipelineRequest,
  // Error types
  ErrorResponse as ApiError,
  ValidationErrorResponse as ValidationError
}

// Legacy type aliases for backward compatibility
export interface LoginCredentials extends LoginRequest {
  rememberMe?: boolean
}

export interface RegisterData {
  email: string
  name: string
  password: string
  confirmPassword: string
}

// User interface with name field for UI compatibility
export interface UserWithName extends User {
  name: string
}

// Auth response wrapper for UI compatibility
export interface AuthResponse extends ApiResponse<{
  user: UserWithName
  tokens: AuthTokens
}> {}

export interface CreateUserData {
  email: string
  name: string
  password: string
  role?: UserRole
}

export interface UpdateUserData {
  name?: string
  email?: string
  role?: UserRole
  isActive?: boolean
}

// Auth token interface extracted from LoginResponse - using camelCase for UI
export interface AuthTokens {
  accessToken: string
  refreshToken: string
  tokenType: string
  expiresAt: number
}

// Additional error types for UI-specific needs
export interface UiApiError extends ErrorResponse {
  timestamp: string
}

export interface UiValidationError extends ValidationErrorResponse {
  value?: unknown
}

// Query key types for TanStack Query - expanded for comprehensive API coverage
export const queryKeys = {
  users: {
    all: ['users'] as const,
    lists: () => [...queryKeys.users.all, 'list'] as const,
    list: (filters: Record<string, unknown>) => [...queryKeys.users.lists(), filters] as const,
    details: () => [...queryKeys.users.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.users.details(), id] as const,
  },
  projects: {
    all: ['projects'] as const,
    lists: () => [...queryKeys.projects.all, 'list'] as const,
    list: (filters: Record<string, unknown>) => [...queryKeys.projects.lists(), filters] as const,
    details: () => [...queryKeys.projects.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.projects.details(), id] as const,
    members: (id: string) => [...queryKeys.projects.detail(id), 'members'] as const,
  },
  assets: {
    all: ['assets'] as const,
    lists: () => [...queryKeys.assets.all, 'list'] as const,
    list: (filters: Record<string, unknown>) => [...queryKeys.assets.lists(), filters] as const,
    details: () => [...queryKeys.assets.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.assets.details(), id] as const,
    search: (query: string) => [...queryKeys.assets.all, 'search', query] as const,
    versions: (id: string) => [...queryKeys.assets.detail(id), 'versions'] as const,
  },
  pipelines: {
    all: ['pipelines'] as const,
    lists: () => [...queryKeys.pipelines.all, 'list'] as const,
    list: (filters: Record<string, unknown>) => [...queryKeys.pipelines.lists(), filters] as const,
    details: () => [...queryKeys.pipelines.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.pipelines.details(), id] as const,
    executions: (id: string) => [...queryKeys.pipelines.detail(id), 'executions'] as const,
    dependencies: (id: string) => [...queryKeys.pipelines.detail(id), 'dependencies'] as const,
    trigger: (id: string) => [...queryKeys.pipelines.detail(id), 'trigger'] as const,
  },
  repositories: {
    all: ['repositories'] as const,
    lists: () => [...queryKeys.repositories.all, 'list'] as const,
    list: (filters: Record<string, unknown>) => [...queryKeys.repositories.lists(), filters] as const,
    details: () => [...queryKeys.repositories.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.repositories.details(), id] as const,
    sync: (id: string) => [...queryKeys.repositories.detail(id), 'sync'] as const,
    status: (id: string) => [...queryKeys.repositories.detail(id), 'status'] as const,
    validate: () => [...queryKeys.repositories.all, 'validate'] as const,
  },
  auth: {
    all: ['auth'] as const,
    user: () => [...queryKeys.auth.all, 'user'] as const,
    tokens: () => [...queryKeys.auth.all, 'tokens'] as const,
    login: () => [...queryKeys.auth.all, 'login'] as const,
    register: () => [...queryKeys.auth.all, 'register'] as const,
    refresh: () => [...queryKeys.auth.all, 'refresh'] as const,
    logout: () => [...queryKeys.auth.all, 'logout'] as const,
  },
  gitops: {
    all: ['gitops'] as const,
    sync: () => [...queryKeys.gitops.all, 'sync'] as const,
    status: () => [...queryKeys.gitops.all, 'status'] as const,
    pending: () => [...queryKeys.gitops.all, 'pending'] as const,
    metrics: () => [...queryKeys.gitops.all, 'metrics'] as const,
    webhooks: () => [...queryKeys.gitops.all, 'webhooks'] as const,
  },
} as const

// Type guards - updated for shared types
export function isApiError(error: unknown): error is ErrorResponse {
  return (
    typeof error === 'object' &&
    error !== null &&
    'code' in error &&
    'message' in error
  )
}

export function isValidationError(error: unknown): error is ValidationErrorResponse {
  return (
    isApiError(error) &&
    'fields' in error
  )
}

export function isUiApiError(error: unknown): error is UiApiError {
  return (
    isApiError(error) &&
    typeof error === 'object' &&
    error !== null &&
    'timestamp' in error
  )
}

export function isSuccessResponse<T>(response: ApiResponse<T>): response is ApiResponse<T> & { success: true } {
  return response.success === true
}