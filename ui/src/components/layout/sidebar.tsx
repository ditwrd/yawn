import * as React from "react"
import { Link } from "@tanstack/react-router"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import {
  LayoutDashboard,
  FolderOpen,
  Settings,
  ChevronLeft,
  ChevronRight,
  Moon,
  Plus,
  Search,
  Users,
  FileText,
  BarChart3,
  Menu,
  X,
} from "lucide-react"

interface SidebarProps {
  className?: string
  collapsed?: boolean
  onToggle?: () => void
}

const navigationItems = [
  {
    title: "Dashboard",
    href: "/dashboard",
    icon: LayoutDashboard,
    description: "View your workspace overview",
  },
  {
    title: "Projects",
    href: "/projects",
    icon: FolderOpen,
    badge: "3",
    description: "Manage your data projects",
  },
  {
    title: "Pipelines",
    href: "/pipelines",
    icon: FileText,
    description: "Create and manage pipelines",
  },
  {
    title: "Analytics",
    href: "/analytics",
    icon: BarChart3,
    description: "View performance metrics",
  },
  {
    title: "Team",
    href: "/team",
    icon: Users,
    description: "Collaborate with your team",
  },
  {
    title: "Settings",
    href: "/settings",
    icon: Settings,
    description: "Configure your preferences",
  },
]

export function Sidebar({ className, collapsed = false, onToggle }: SidebarProps) {
  return (
    <aside
      className={cn(
        "flex flex-col bg-sidebar border-r border-sidebar-border",
        "sidebar-transition",
        collapsed
          ? "w-16 lg:w-16"
          : "w-72 sm:w-64 lg:w-64",
        "h-screen overflow-hidden",
        className
      )}
      role="navigation"
      aria-label="Main navigation"
    >
      {/* Header */}
      <header className="flex h-16 items-center justify-between shrink-0 px-3 lg:px-4">
        {!collapsed && (
          <div className="flex items-center gap-2 lg:gap-3 fade-in">
            <Moon
              className="h-5 w-5 lg:h-6 lg:w-6 text-sidebar-primary flex-shrink-0"
              aria-hidden="true"
            />
            <span
              className="text-responsive-base font-semibold text-sidebar-foreground truncate"
              role="heading"
              aria-level={1}
            >
              YAWN
            </span>
          </div>
        )}

        <Button
          variant="ghost"
          size="icon"
          onClick={onToggle}
          className={cn(
            "h-8 w-8 lg:h-8 lg:w-8",
            "text-sidebar-foreground hover:bg-sidebar-accent",
            "focus-enhanced tap-target",
            "btn-hover-lift"
          )}
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          aria-expanded={!collapsed}
        >
          {collapsed ? (
            <ChevronRight className="h-4 w-4" />
          ) : (
            <>
              <ChevronLeft className="h-4 w-4 hidden lg:block" />
              <X className="h-4 w-4 lg:hidden" />
            </>
          )}
        </Button>
      </header>

      <Separator className="bg-sidebar-border shrink-0" />

      {/* Search - Hidden on mobile when collapsed */}
      {!collapsed && (
        <div className="p-3 lg:p-4 fade-in">
          <div className="relative">
            <Search
              className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
              aria-hidden="true"
            />
            <input
              type="text"
              placeholder="Search..."
              className={cn(
                "w-full rounded-md bg-input border-sidebar-border",
                "px-10 py-2 text-sm text-foreground",
                "placeholder:text-muted-foreground",
                "focus:outline-none focus:ring-2 focus:ring-sidebar-ring",
                "focus:ring-offset-2 focus:ring-offset-sidebar",
                "transition-all duration-fast"
              )}
              aria-label="Search navigation"
            />
          </div>
        </div>
      )}

      {!collapsed && (
        <Separator className="bg-sidebar-border shrink-0" />
      )}

      {/* Navigation */}
      <nav
        className="flex-1 overflow-y-auto p-3 lg:p-4"
        aria-label="Primary navigation"
      >
        <ul className="space-y-1 lg:space-y-2">
          {navigationItems.map((item) => (
            <li key={item.href}>
              <Link
                to={item.href}
                className={cn(
                  "group interactive-item tap-target",
                  "flex items-center gap-2 lg:gap-3",
                  "rounded-md px-2 lg:px-3 py-2",
                  "text-sm font-medium text-sidebar-foreground",
                  "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
                  "focus:bg-sidebar-accent focus:text-sidebar-accent-foreground",
                  "transition-all duration-fast",
                  collapsed && "justify-center px-2"
                )}
                aria-label={item.title}
                title={collapsed ? item.title : item.description}
              >
                <item.icon
                  className={cn(
                    "h-5 w-5 flex-shrink-0",
                    "group-hover:text-sidebar-primary",
                    "transition-colors duration-fast"
                  )}
                  aria-hidden="true"
                />
                {!collapsed && (
                  <span className="truncate fade-in">{item.title}</span>
                )}
                {!collapsed && item.badge && (
                  <Badge
                    variant="secondary"
                    className="text-xs fade-in"
                    aria-label={`${item.badge} items`}
                  >
                    {item.badge}
                  </Badge>
                )}
              </Link>
            </li>
          ))}
        </ul>
      </nav>

      <Separator className="bg-sidebar-border shrink-0" />

      {/* Footer */}
      <footer className="p-3 lg:p-4 shrink-0">
        <Button
          variant="default"
          size={collapsed ? "icon" : "default"}
          className={cn(
            "w-full",
            "bg-sidebar-primary text-sidebar-primary-foreground",
            "hover:bg-sidebar-primary/90 focus:ring-sidebar-ring",
            "btn-hover-lift focus-enhanced",
            collapsed && "aspect-square"
          )}
          aria-label={collapsed ? "Create new project" : "New Project"}
        >
          <Plus className="h-4 w-4" />
          {!collapsed && <span className="fade-in">New Project</span>}
        </Button>
      </footer>
    </aside>
  )
}