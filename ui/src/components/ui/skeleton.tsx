import { cn } from '@/lib/utils'

interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'text' | 'circular' | 'rounded'
  animation?: 'pulse' | 'wave' | 'none'
}

/**
 * Skeleton component for loading states
 */
function Skeleton({
  className,
  variant = 'default',
  animation = 'pulse',
  ...props
}: SkeletonProps) {
  const variantClasses = {
    default: 'rounded-md',
    text: 'rounded-sm h-4',
    circular: 'rounded-full',
    rounded: 'rounded-lg',
  }

  const animationClasses = {
    pulse: 'animate-pulse',
    wave: 'animate-shimmer',
    none: '',
  }

  return (
    <div
      className={cn(
        'bg-gray-800',
        variantClasses[variant],
        animationClasses[animation],
        className
      )}
      {...props}
    />
  )
}

/**
 * Card skeleton for loading card layouts
 */
function CardSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn('rounded-lg border border-gray-800 bg-black p-6', className)}>
      <div className="space-y-4">
        <Skeleton variant="text" className="h-6 w-3/4" />
        <Skeleton variant="text" className="h-4 w-1/2" />
        <div className="space-y-2">
          <Skeleton variant="text" className="h-4 w-full" />
          <Skeleton variant="text" className="h-4 w-5/6" />
          <Skeleton variant="text" className="h-4 w-4/6" />
        </div>
      </div>
    </div>
  )
}

/**
 * Table skeleton for loading table data
 */
function TableSkeleton({
  rows = 5,
  columns = 4,
  className
}: {
  rows?: number
  columns?: number
  className?: string
}) {
  return (
    <div className={cn('w-full', className)}>
      <div className="border-b border-gray-800">
        <div className="grid gap-4 p-4" style={{ gridTemplateColumns: `repeat(${columns}, 1fr)` }}>
          {Array.from({ length: columns }).map((_, i) => (
            <Skeleton key={`header-${i}`} variant="text" className="h-5" />
          ))}
        </div>
      </div>
      {Array.from({ length: rows }).map((_, rowIndex) => (
        <div key={`row-${rowIndex}`} className="border-b border-gray-800">
          <div className="grid gap-4 p-4" style={{ gridTemplateColumns: `repeat(${columns}, 1fr)` }}>
            {Array.from({ length: columns }).map((_, colIndex) => (
              <Skeleton key={`cell-${rowIndex}-${colIndex}`} variant="text" className="h-4" />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

/**
 * List skeleton for loading list items
 */
function ListSkeleton({
  items = 5,
  showAvatar = false,
  className
}: {
  items?: number
  showAvatar?: boolean
  className?: string
}) {
  return (
    <div className={cn('space-y-4', className)}>
      {Array.from({ length: items }).map((_, i) => (
        <div key={i} className="flex items-center space-x-4 p-4 border border-gray-800 rounded-lg">
          {showAvatar && (
            <Skeleton variant="circular" className="h-10 w-10 flex-shrink-0" />
          )}
          <div className="flex-1 space-y-2">
            <Skeleton variant="text" className="h-5 w-3/4" />
            <Skeleton variant="text" className="h-4 w-1/2" />
          </div>
          <Skeleton variant="text" className="h-8 w-20" />
        </div>
      ))}
    </div>
  )
}

/**
 * Form skeleton for loading forms
 */
function FormSkeleton({
  fields = 4,
  showActions = true,
  className
}: {
  fields?: number
  showActions?: boolean
  className?: string
}) {
  return (
    <div className={cn('space-y-6', className)}>
      {Array.from({ length: fields }).map((_, i) => (
        <div key={i} className="space-y-2">
          <Skeleton variant="text" className="h-4 w-1/4" />
          <Skeleton variant="rounded" className="h-10 w-full" />
        </div>
      ))}
      {showActions && (
        <div className="flex justify-end space-x-3 pt-4">
          <Skeleton variant="rounded" className="h-10 w-24" />
          <Skeleton variant="rounded" className="h-10 w-32" />
        </div>
      )}
    </div>
  )
}

/**
 * Dashboard skeleton for loading dashboard layouts
 */
function DashboardSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn('space-y-6', className)}>
      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {Array.from({ length: 4 }).map((_, i) => (
          <CardSkeleton key={`stat-${i}`} />
        ))}
      </div>

      {/* Main content area */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <CardSkeleton />
        </div>
        <div className="space-y-6">
          <CardSkeleton />
          <CardSkeleton />
        </div>
      </div>
    </div>
  )
}

/**
 * Project skeleton for loading project items
 */
function ProjectSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn('border border-gray-800 rounded-lg p-6 space-y-4', className)}>
      <div className="flex items-start justify-between">
        <div className="space-y-2 flex-1">
          <Skeleton variant="text" className="h-6 w-3/4" />
          <Skeleton variant="text" className="h-4 w-1/2" />
        </div>
        <Skeleton variant="rounded" className="h-8 w-20" />
      </div>
      <Skeleton variant="text" className="h-4 w-full" />
      <Skeleton variant="text" className="h-4 w-5/6" />
      <div className="flex items-center justify-between pt-4 border-t border-gray-800">
        <div className="flex items-center space-x-4">
          <Skeleton variant="circular" className="h-8 w-8" />
          <Skeleton variant="text" className="h-4 w-24" />
        </div>
        <Skeleton variant="text" className="h-4 w-20" />
      </div>
    </div>
  )
}

/**
 * Pipeline skeleton for loading pipeline items
 */
function PipelineSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn('border border-gray-800 rounded-lg p-6 space-y-4', className)}>
      <div className="flex items-center justify-between">
        <div className="space-y-2 flex-1">
          <Skeleton variant="text" className="h-6 w-2/3" />
          <Skeleton variant="text" className="h-4 w-1/3" />
        </div>
        <Skeleton variant="circular" className="h-3 w-3" />
      </div>
      <Skeleton variant="text" className="h-4 w-full" />
      <div className="flex items-center space-x-2">
        <Skeleton variant="rounded" className="h-6 w-16" />
        <Skeleton variant="rounded" className="h-6 w-20" />
        <Skeleton variant="rounded" className="h-6 w-24" />
      </div>
    </div>
  )
}

/**
 * Asset skeleton for loading asset items
 */
function AssetSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn('border border-gray-800 rounded-lg p-4 space-y-3', className)}>
      <div className="flex items-center space-x-3">
        <Skeleton variant="circular" className="h-10 w-10" />
        <div className="space-y-1 flex-1">
          <Skeleton variant="text" className="h-5 w-1/2" />
          <Skeleton variant="text" className="h-4 w-1/3" />
        </div>
        <Skeleton variant="rounded" className="h-8 w-16" />
      </div>
      <Skeleton variant="text" className="h-4 w-full" />
      <div className="flex items-center justify-between pt-2">
        <Skeleton variant="text" className="h-4 w-24" />
        <Skeleton variant="text" className="h-4 w-20" />
      </div>
    </div>
  )
}

export {
  Skeleton,
  CardSkeleton,
  TableSkeleton,
  ListSkeleton,
  FormSkeleton,
  DashboardSkeleton,
  ProjectSkeleton,
  PipelineSkeleton,
  AssetSkeleton,
}