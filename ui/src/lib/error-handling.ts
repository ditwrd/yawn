import type { APIError } from './api-client'

// Error severity levels
export type ErrorSeverity = 'low' | 'medium' | 'high' | 'critical'

// Error categories
export type ErrorCategory =
  | 'network'
  | 'authentication'
  | 'authorization'
  | 'validation'
  | 'server'
  | 'client'
  | 'timeout'
  | 'unknown'

// User-friendly error messages
const ERROR_MESSAGES: Record<string, string> = {
  // Network errors
  NETWORK_ERROR: 'Network connection failed. Please check your internet connection.',
  TIMEOUT_ERROR: 'Request timed out. Please try again.',
  OFFLINE_ERROR: 'You appear to be offline. Please check your connection.',

  // Authentication errors
  UNAUTHORIZED: 'Please log in to access this feature.',
  TOKEN_EXPIRED: 'Your session has expired. Please log in again.',
  INVALID_CREDENTIALS: 'Invalid email or password.',
  ACCOUNT_LOCKED: 'Your account has been locked. Please contact support.',

  // Authorization errors
  FORBIDDEN: 'You don\'t have permission to perform this action.',
  INSUFFICIENT_PERMISSIONS: 'Your role doesn\'t have sufficient permissions.',

  // Validation errors
  VALIDATION_ERROR: 'Please check your input and try again.',
  INVALID_INPUT: 'Invalid input provided.',
  REQUIRED_FIELD: 'This field is required.',
  INVALID_FORMAT: 'Invalid format.',

  // Server errors
  INTERNAL_ERROR: 'Something went wrong. Please try again later.',
  SERVICE_UNAVAILABLE: 'Service temporarily unavailable. Please try again later.',
  RATE_LIMITED: 'Too many requests. Please wait and try again.',

  // Resource errors
  NOT_FOUND: 'The requested resource was not found.',
  ALREADY_EXISTS: 'This resource already exists.',
  CONFLICT: 'There was a conflict with your request.',

  // Default fallback
  UNKNOWN_ERROR: 'An unexpected error occurred. Please try again.',
}

/**
 * Categorize API errors based on status code and error details
 */
export function categorizeError(error: APIError): ErrorCategory {
  // Network-related errors
  if (!error.status) {
    if (error.message.includes('fetch') || error.message.includes('network')) {
      return 'network'
    }
    if (error.message.includes('timeout')) {
      return 'timeout'
    }
    return 'unknown'
  }

  const status = error.status

  // Authentication errors (401)
  if (status === 401) {
    if (error.code === 'TOKEN_EXPIRED' || error.message.includes('expired')) {
      return 'authentication'
    }
    return 'authentication'
  }

  // Authorization errors (403)
  if (status === 403) {
    return 'authorization'
  }

  // Validation errors (422)
  if (status === 422) {
    return 'validation'
  }

  // Client errors (400-499)
  if (status >= 400 && status < 500) {
    if (status === 404) {
      return 'client'
    }
    if (status === 409) {
      return 'client'
    }
    return 'client'
  }

  // Server errors (500-599)
  if (status >= 500) {
    return 'server'
  }

  return 'unknown'
}

/**
 * Determine error severity based on category and status
 */
export function getErrorSeverity(error: APIError): ErrorSeverity {
  const category = categorizeError(error)

  switch (category) {
    case 'network':
    case 'timeout':
      return 'medium'

    case 'authentication':
      return 'medium'

    case 'authorization':
      return 'medium'

    case 'validation':
      return 'low'

    case 'server':
      return error.status === 503 ? 'high' : 'critical'

    case 'client':
      return error.status === 404 ? 'low' : 'medium'

    default:
      return 'medium'
  }
}

/**
 * Get user-friendly error message
 */
export function getUserFriendlyMessage(error: APIError): string {
  const category = categorizeError(error)

  // Check for specific error codes first
  if (error.code && ERROR_MESSAGES[error.code]) {
    return ERROR_MESSAGES[error.code]
  }

  // Check status-specific messages
  if (error.status) {
    switch (error.status) {
      case 400:
        return ERROR_MESSAGES.INVALID_INPUT
      case 401:
        return ERROR_MESSAGES.UNAUTHORIZED
      case 403:
        return ERROR_MESSAGES.FORBIDDEN
      case 404:
        return ERROR_MESSAGES.NOT_FOUND
      case 409:
        return ERROR_MESSAGES.ALREADY_EXISTS
      case 422:
        return ERROR_MESSAGES.VALIDATION_ERROR
      case 429:
        return ERROR_MESSAGES.RATE_LIMITED
      case 500:
        return ERROR_MESSAGES.INTERNAL_ERROR
      case 502:
      case 503:
      case 504:
        return ERROR_MESSAGES.SERVICE_UNAVAILABLE
    }
  }

  // Check category-specific messages
  switch (category) {
    case 'network':
      return ERROR_MESSAGES.NETWORK_ERROR
    case 'timeout':
      return ERROR_MESSAGES.TIMEOUT_ERROR
    case 'authentication':
      return ERROR_MESSAGES.UNAUTHORIZED
    case 'authorization':
      return ERROR_MESSAGES.FORBIDDEN
    case 'validation':
      return ERROR_MESSAGES.VALIDATION_ERROR
    case 'server':
      return ERROR_MESSAGES.INTERNAL_ERROR
  }

  // Use the error message if it's user-friendly
  if (error.message && !error.message.includes('HTTP') && !error.message.toLowerCase().includes('error')) {
    return error.message
  }

  return ERROR_MESSAGES.UNKNOWN_ERROR
}

/**
 * Extract field validation errors
 */
export function getFieldErrors(error: APIError): Record<string, string[]> {
  if (!error.fields || typeof error.fields !== 'object') {
    return {}
  }

  const fieldErrors: Record<string, string[]> = {}

  // Handle different field error formats
  if (Array.isArray(error.fields)) {
    // If fields is an array of validation error objects
    error.fields.forEach((field: any) => {
      if (field.field && field.message) {
        fieldErrors[field.field] = [field.message]
      }
    })
  } else {
    // If fields is a record
    Object.entries(error.fields).forEach(([field, messages]) => {
      if (typeof messages === 'string') {
        fieldErrors[field] = [messages]
      } else if (Array.isArray(messages)) {
        fieldErrors[field] = messages.map(String)
      }
    })
  }

  return fieldErrors
}

/**
 * Determine if error is recoverable (can be retried)
 */
export function isRecoverableError(error: APIError): boolean {
  const category = categorizeError(error)
  const severity = getErrorSeverity(error)

  // Don't retry authentication and authorization errors
  if (category === 'authentication' || category === 'authorization') {
    return false
  }

  // Don't retry validation errors
  if (category === 'validation') {
    return false
  }

  // Don't retry client errors (except timeout)
  if (category === 'client' && severity !== 'medium') {
    return false
  }

  // Retry network, timeout, and server errors
  return ['network', 'timeout', 'server'].includes(category)
}

/**
 * Create error notification object
 */
export function createErrorNotification(error: APIError) {
  return {
    id: Math.random().toString(36).substr(2, 9),
    type: 'error' as const,
    title: 'Error',
    message: getUserFriendlyMessage(error),
    severity: getErrorSeverity(error),
    category: categorizeError(error),
    fieldErrors: getFieldErrors(error),
    timestamp: new Date().toISOString(),
    originalError: error,
  }
}

/**
 * Enhanced error handler for logging and monitoring
 */
export class ErrorHandler {
  private static instance: ErrorHandler
  private errorLog: Array<{ error: APIError; timestamp: string; context?: any }> = []

  static getInstance(): ErrorHandler {
    if (!ErrorHandler.instance) {
      ErrorHandler.instance = new ErrorHandler()
    }
    return ErrorHandler.instance
  }

  /**
   * Log error for debugging and monitoring
   */
  log(error: APIError, context?: any): void {
    const logEntry = {
      error: {
        ...error,
        message: error.message,
        status: error.status,
        code: error.code,
      },
      timestamp: new Date().toISOString(),
      context,
    }

    this.errorLog.push(logEntry)

    // Keep only last 100 errors in memory
    if (this.errorLog.length > 100) {
      this.errorLog = this.errorLog.slice(-100)
    }

    // Log to console in development
    if (import.meta.env.DEV) {
      console.group(`🚨 API Error [${error.status || 'NETWORK'}]`)
      console.error('Error:', error)
      if (context) console.log('Context:', context)
      console.groupEnd()
    }

    // In production, you would send this to your error monitoring service
    // Examples: Sentry, LogRocket, etc.
  }

  /**
   * Get recent errors for debugging
   */
  getRecentErrors(limit: number = 10): Array<{ error: APIError; timestamp: string; context?: any }> {
    return this.errorLog.slice(-limit)
  }

  /**
   * Clear error log
   */
  clearLog(): void {
    this.errorLog = []
  }

  /**
   * Get error statistics
   */
  getErrorStats(): Record<ErrorCategory, number> {
    const stats: Record<ErrorCategory, number> = {
      network: 0,
      authentication: 0,
      authorization: 0,
      validation: 0,
      server: 0,
      client: 0,
      timeout: 0,
      unknown: 0,
    }

    this.errorLog.forEach(({ error }) => {
      const category = categorizeError(error)
      stats[category]++
    })

    return stats
  }
}

// Export singleton instance
export const errorHandler = ErrorHandler.getInstance()

/**
 * Hook for handling errors in React components
 */
export function useErrorHandler() {
  const handleError = (error: APIError | Error, context?: any) => {
    const apiError = error as APIError

    // Log the error
    errorHandler.log(apiError, context)

    // Create and return notification
    return createErrorNotification(apiError)
  }

  const getErrorMessage = (error: APIError | Error): string => {
    const apiError = error as APIError
    return getUserFriendlyMessage(apiError)
  }

  const getFieldErrors = (error: APIError | Error): Record<string, string[]> => {
    const apiError = error as APIError
    return getFieldErrors(apiError)
  }

  const shouldRetry = (error: APIError | Error): boolean => {
    const apiError = error as APIError
    return isRecoverableError(apiError)
  }

  return {
    handleError,
    getErrorMessage,
    getFieldErrors,
    shouldRetry,
    errorLog: errorHandler.getRecentErrors(),
    errorStats: errorHandler.getErrorStats(),
  }
}