import { Loader2, AlertCircle, RefreshCw } from 'lucide-react'
import { Button } from './button'
import { Alert, AlertDescription } from './alert'
import { cn } from '@/lib/utils'
import type { APIError } from '@/lib/api-client'
import { getUserFriendlyMessage, isRecoverableError } from '@/lib/error-handling'

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg'
  className?: string
  text?: string
}

/**
 * Loading spinner component
 */
function LoadingSpinner({ size = 'md', className, text }: LoadingSpinnerProps) {
  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-6 w-6',
    lg: 'h-8 w-8',
  }

  return (
    <div className={cn('flex items-center space-x-2', className)}>
      <Loader2 className={cn('animate-spin', sizeClasses[size])} />
      {text && <span className="text-sm text-gray-400">{text}</span>}
    </div>
  )
}

interface LoadingStateProps {
  isLoading: boolean
  error?: APIError | Error | null
  children: React.ReactNode
  fallback?: React.ReactNode
  errorFallback?: React.ReactNode
  onRetry?: () => void
  className?: string
}

/**
 * Generic loading state component that handles loading, error, and success states
 */
function LoadingState({
  isLoading,
  error,
  children,
  fallback,
  errorFallback,
  onRetry,
  className,
}: LoadingStateProps) {
  if (isLoading) {
    return <>{fallback || <LoadingSpinner />}</>
  }

  if (error) {
    if (errorFallback) {
      return <>{errorFallback}</>
    }

    const apiError = error as APIError
    const isRecoverable = onRetry && isRecoverableError(apiError)

    return (
      <div className={cn('space-y-4', className)}>
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            {getUserFriendlyMessage(apiError)}
          </AlertDescription>
        </Alert>

        {isRecoverable && (
          <div className="flex justify-center">
            <Button
              variant="outline"
              size="sm"
              onClick={onRetry}
              className="text-gray-400 border-gray-700 hover:bg-gray-800"
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              Retry
            </Button>
          </div>
        )}
      </div>
    )
  }

  return <>{children}</>
}

interface LoadingCardProps {
  title?: string
  description?: string
  isLoading: boolean
  error?: APIError | Error | null
  children: React.ReactNode
  className?: string
  onRetry?: () => void
}

/**
 * Card with integrated loading states
 */
function LoadingCard({
  title,
  description,
  isLoading,
  error,
  children,
  className,
  onRetry,
}: LoadingCardProps) {
  return (
    <div className={cn('rounded-lg border border-gray-800 bg-black', className)}>
      {(title || description) && (
        <div className="p-6 border-b border-gray-800">
          {title && (
            <h3 className="text-lg font-medium text-white">{title}</h3>
          )}
          {description && (
            <p className="text-sm text-gray-400 mt-1">{description}</p>
          )}
        </div>
      )}

      <div className="p-6">
        <LoadingState
          isLoading={isLoading}
          error={error}
          onRetry={onRetry}
          fallback={
            <div className="flex items-center justify-center py-8">
              <LoadingSpinner size="lg" text="Loading..." />
            </div>
          }
        >
          {children}
        </LoadingState>
      </div>
    </div>
  )
}

interface LoadingButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  isLoading?: boolean
  loadingText?: string
  children: React.ReactNode
}

/**
 * Button with integrated loading state
 */
function LoadingButton({
  isLoading = false,
  loadingText,
  children,
  disabled,
  className,
  ...props
}: LoadingButtonProps) {
  return (
    <Button
      {...props}
      disabled={disabled || isLoading}
      className={cn('relative', className)}
    >
      {isLoading && (
        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
      )}
      {isLoading ? (loadingText || 'Loading...') : children}
    </Button>
  )
}

interface ProgressiveLoadingProps {
  items: any[]
  isLoading: boolean
  error?: APIError | Error | null
  renderItem: (item: any, index: number) => React.ReactNode
  renderSkeleton?: (index: number) => React.ReactNode
  emptyMessage?: string
  emptyIcon?: React.ReactNode
  onRetry?: () => void
  className?: string
}

/**
 * Progressive loading component for lists
 */
function ProgressiveLoading({
  items,
  isLoading,
  error,
  renderItem,
  renderSkeleton,
  emptyMessage = 'No items found',
  emptyIcon,
  onRetry,
  className,
}: ProgressiveLoadingProps) {
  if (isLoading && items.length === 0) {
    return (
      <div className={cn('space-y-4', className)}>
        {Array.from({ length: 3 }).map((_, i) =>
          renderSkeleton ? renderSkeleton(i) : <LoadingSpinner key={i} />
        )}
      </div>
    )
  }

  if (error) {
    return (
      <div className={cn('flex flex-col items-center justify-center py-12 space-y-4', className)}>
        <AlertCircle className="h-8 w-8 text-red-500" />
        <p className="text-gray-400 text-center">{getUserFriendlyMessage(error as APIError)}</p>
        {onRetry && (
          <Button variant="outline" size="sm" onClick={onRetry}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        )}
      </div>
    )
  }

  if (!isLoading && items.length === 0) {
    return (
      <div className={cn('flex flex-col items-center justify-center py-12 space-y-4', className)}>
        {emptyIcon || <AlertCircle className="h-8 w-8 text-gray-600" />}
        <p className="text-gray-500 text-center">{emptyMessage}</p>
      </div>
    )
  }

  return (
    <div className={cn('space-y-4', className)}>
      {items.map((item, index) => renderItem(item, index))}
      {isLoading && items.length > 0 && (
        <div className="flex justify-center py-4">
          <LoadingSpinner size="sm" />
        </div>
      )}
    </div>
  )
}

interface LoadingOverlayProps {
  isLoading: boolean
  text?: string
  children: React.ReactNode
  className?: string
}

/**
 * Loading overlay component
 */
function LoadingOverlay({ isLoading, text, children, className }: LoadingOverlayProps) {
  return (
    <div className={cn('relative', className)}>
      {children}
      {isLoading && (
        <div className="absolute inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
          <LoadingSpinner size="lg" text={text || 'Loading...'} />
        </div>
      )}
    </div>
  )
}

interface OptimisticUpdateProps<T> {
  data: T
  isUpdating: boolean
  error?: APIError | Error | null
  optimisticData?: T
  children: (data: T, isUpdating: boolean) => React.ReactNode
  onRollback?: () => void
}

/**
 * Component for handling optimistic updates
 */
function OptimisticUpdate<T>({
  data,
  isUpdating,
  error,
  optimisticData,
  children,
  onRollback,
}: OptimisticUpdateProps<T>) {
  const displayData = isUpdating && optimisticData ? optimisticData : data

  if (error && optimisticData) {
    return (
      <div className="space-y-4">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Failed to update. Changes have been rolled back.
          </AlertDescription>
        </Alert>
        {onRollback && (
          <Button variant="outline" size="sm" onClick={onRollback}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Rollback
          </Button>
        )}
      </div>
    )
  }

  return <>{children(displayData, isUpdating)}</>
}

export {
  LoadingSpinner,
  LoadingState,
  LoadingCard,
  LoadingButton,
  ProgressiveLoading,
  LoadingOverlay,
  OptimisticUpdate,
}