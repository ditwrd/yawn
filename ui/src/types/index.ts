// Export all API types
export * from './api'

// Common component props
export interface BaseComponentProps {
  className?: string
  children?: React.ReactNode
}

export interface ButtonProps extends BaseComponentProps {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  disabled?: boolean
  loading?: boolean
  onClick?: () => void
  type?: 'button' | 'submit' | 'reset'
}

export interface InputProps extends BaseComponentProps {
  type?: string
  placeholder?: string
  value?: string
  onChange?: (value: string) => void
  error?: string
  disabled?: boolean
  required?: boolean
}

// Form types
export interface FormField {
  name: string
  label: string
  type: string
  required?: boolean
  placeholder?: string
  validation?: Record<string, unknown>
}

export interface FormState {
  values: Record<string, unknown>
  errors: Record<string, string>
  touched: Record<string, boolean>
  isSubmitting: boolean
  isValid: boolean
}

// Navigation types
export interface NavItem {
  id: string
  label: string
  href: string
  icon?: string
  badge?: number
  isActive?: boolean
  children?: NavItem[]
}

export interface BreadcrumbItem {
  label: string
  href?: string
  isActive?: boolean
}

// Utility types
export type Optional<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>
export type RequiredBy<T, K extends keyof T> = T & Required<Pick<T, K>>
export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P]
}

// Event types
export interface CustomEvent<T = unknown> {
  type: string
  payload: T
  timestamp: number
}

export type EventHandler<T = unknown> = (event: CustomEvent<T>) => void

// Theme types
export type ThemeMode = 'dark' | 'light'
export type ColorScheme = 'purple' | 'indigo' | 'blue' | 'green'

export interface Theme {
  mode: ThemeMode
  colorScheme: ColorScheme
  fontSize: 'sm' | 'base' | 'lg'
}

// Status types
export type Status = 'idle' | 'loading' | 'success' | 'error'
export type Priority = 'low' | 'medium' | 'high' | 'critical'

// Search and filter types
export interface SearchFilters {
  query?: string
  page?: number
  limit?: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

export interface SearchState {
  filters: SearchFilters
  results: unknown[]
  totalCount: number
  isLoading: boolean
  error?: string
}