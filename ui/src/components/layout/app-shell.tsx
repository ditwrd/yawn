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
    <div className={cn("flex h-screen bg-background", className)}>
      <Sidebar collapsed={sidebarCollapsed} onToggle={onSidebarToggle} />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header onSidebarToggle={onSidebarToggle} sidebarCollapsed={sidebarCollapsed} />
        <main className="flex-1 overflow-auto p-6">
          {children}
        </main>
      </div>
    </div>
  )
}