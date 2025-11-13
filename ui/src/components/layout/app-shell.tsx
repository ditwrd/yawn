import * as React from "react"
import { cn } from "@/lib/utils"
import { Sidebar } from "./sidebar"
import { Header } from "./header"

interface AppShellProps {
  children: React.ReactNode
  className?: string
  sidebarCollapsed?: boolean
  onSidebarToggle?: () => void
}

export function AppShell({
  children,
  className,
  sidebarCollapsed = false,
  onSidebarToggle,
}: AppShellProps) {
  const [isMobile, setIsMobile] = React.useState(false)

  React.useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 1024)
    }

    checkMobile()
    window.addEventListener('resize', checkMobile)
    return () => window.removeEventListener('resize', checkMobile)
  }, [])

  // Close sidebar on mobile when clicking outside
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        isMobile &&
        !sidebarCollapsed &&
        !(event.target as Element).closest('aside') &&
        !(event.target as Element).closest('[aria-label="Toggle sidebar menu"]')
      ) {
        onSidebarToggle?.()
      }
    }

    if (isMobile && !sidebarCollapsed) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isMobile, sidebarCollapsed, onSidebarToggle])

  return (
    <div className={cn(
      "flex h-screen bg-background overflow-hidden",
      "layout-transition",
      className
    )}>
      <Sidebar
        collapsed={sidebarCollapsed}
        onToggle={onSidebarToggle}
        className="h-full"
      />

      <div className={cn(
        "flex flex-1 flex-col overflow-hidden min-w-0",
        "transition-all duration-300 ease-in-out"
      )}>
        <Header
          onSidebarToggle={onSidebarToggle}
          sidebarCollapsed={sidebarCollapsed}
          className="flex-shrink-0"
        />

        <main
          id="main-content"
          className={cn(
            "flex-1 overflow-auto",
            "container-responsive py-6 lg:py-8",
            "fade-in"
          )}
          role="main"
        >
          {children}
        </main>
      </div>

      {/* Skip to content link for accessibility */}
      <a
        href="#main-content"
        className="skip-to-content"
        aria-label="Skip to main content"
      >
        Skip to main content
      </a>
    </div>
  )
}