import * as React from "react"
import { Link } from "@tanstack/react-router"
import { cn } from "@/lib/utils"
import { ChevronRight, Home } from "lucide-react"

interface BreadcrumbItem {
  label: string
  href?: string
  current?: boolean
}

interface BreadcrumbProps {
  items: BreadcrumbItem[]
  className?: string
  showHome?: boolean
  maxItems?: number // For responsive truncation
}

export function Breadcrumb({
  items,
  className,
  showHome = true,
  maxItems = 5,
}: BreadcrumbProps) {
  // Truncate items on mobile if too many
  const [isMobile, setIsMobile] = React.useState(false)
  const [visibleItems, setVisibleItems] = React.useState(items)

  React.useEffect(() => {
    const checkMobile = () => {
      const mobile = window.innerWidth < 640
      setIsMobile(mobile)

      // On mobile, show fewer items
      if (mobile && items.length > maxItems) {
        const showCount = Math.floor(maxItems / 2)
        const truncated = [
          ...items.slice(0, showCount),
          { label: "...", href: undefined },
          ...items.slice(-showCount)
        ]
        setVisibleItems(truncated)
      } else {
        setVisibleItems(items)
      }
    }

    checkMobile()
    window.addEventListener('resize', checkMobile)
    return () => window.removeEventListener('resize', checkMobile)
  }, [items, maxItems])

  if (!visibleItems.length) return null

  return (
    <nav
      className={cn(
        "flex items-center space-x-1 text-sm text-muted-foreground",
        "font-mono",
        "fade-in",
        className
      )}
      aria-label="Breadcrumb navigation"
    >
      {showHome && (
        <Link
          to="/"
          className={cn(
            "interactive-item tap-target",
            "flex items-center hover:text-foreground",
            "focus:text-foreground focus-visible:ring-2 focus-visible:ring-ring",
            "focus-visible:ring-offset-2 focus-visible:ring-offset-background",
            "rounded-sm p-1 transition-colors duration-150",
            "text-muted-foreground hover:text-foreground"
          )}
          aria-label="Navigate to home"
        >
          <Home className="h-4 w-4" />
          <span className="sr-only">Home</span>
        </Link>
      )}

      {(showHome || visibleItems.length > 0) && (
        <ChevronRight
          className="h-4 w-4 flex-shrink-0"
          aria-hidden="true"
        />
      )}

      {visibleItems.map((item, index) => {
        const isLast = index === visibleItems.length - 1
        const isEllipsis = item.label === "..."

        return (
          <React.Fragment key={index}>
            {!isEllipsis && item.href && !isLast ? (
              <Link
                to={item.href}
                className={cn(
                  "interactive-item tap-target",
                  "hover:text-foreground focus:text-foreground",
                  "focus-visible:ring-2 focus-visible:ring-ring",
                  "focus-visible:ring-offset-2 focus-visible:ring-offset-background",
                  "rounded-sm p-1 transition-colors duration-150",
                  "truncate max-w-[120px] sm:max-w-[200px]",
                  item.current && "text-foreground font-medium"
                )}
                aria-current={item.current ? "page" : undefined}
                title={item.label}
              >
                {item.label}
              </Link>
            ) : (
              <span
                className={cn(
                  "truncate max-w-[120px] sm:max-w-[200px]",
                  isLast && "text-foreground font-medium",
                  isEllipsis && "text-muted-foreground"
                )}
                aria-current={isLast ? "page" : undefined}
              >
                {item.label}
              </span>
            )}

            {!isLast && (
              <ChevronRight
                className="h-4 w-4 flex-shrink-0"
                aria-hidden="true"
              />
            )}
          </React.Fragment>
        )
      })}
    </nav>
  )
}

// Helper function to generate breadcrumbs from routes
export function generateBreadcrumbs(pathname: string): BreadcrumbItem[] {
  const segments = pathname.split('/').filter(Boolean)
  const breadcrumbs: BreadcrumbItem[] = []

  segments.forEach((segment, index) => {
    const href = '/' + segments.slice(0, index + 1).join('/')
    const label = formatBreadcrumbLabel(segment)
    const isLast = index === segments.length - 1

    breadcrumbs.push({
      label,
      href: isLast ? undefined : href,
      current: isLast,
    })
  })

  return breadcrumbs
}

// Helper to format segment names
function formatBreadcrumbLabel(segment: string): string {
  return segment
    .split('-')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}