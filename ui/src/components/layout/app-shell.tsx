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
  return (
    <div className={cn(
      "flex h-screen bg-background overflow-hidden",
      "layout-transition",
      className
    )}>
      {/* Mobile Sidebar Overlay */}
      {!sidebarCollapsed && (
        <div
          className="mobile-sidebar-overlay lg:hidden"
          onClick={onSidebarToggle}
          aria-hidden="true"
        />
      )}

      <Sidebar
        collapsed={sidebarCollapsed}
        onToggle={onSidebarToggle}
        className={cn(
          "fixed lg:relative z-50 h-full",
          "sidebar-transition",
          sidebarCollapsed && "-translate-x-full lg:translate-x-0"
        )}
      />

      <div className={cn(
        "flex flex-1 flex-col overflow-hidden min-w-0",
        "transition-all duration-normal ease-in-out",
        sidebarCollapsed ? "lg:ml-0" : "lg:ml-0"
      )}>
        <Header
          onSidebarToggle={onSidebarToggle}
          sidebarCollapsed={sidebarCollapsed}
          className="flex-shrink-0"
        />

        <main className={cn(
          "flex-1 overflow-auto",
          "container-responsive py-6 lg:py-8",
          "fade-in"
        )}>
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