import * as React from "react"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { ThemeToggle } from "@/components/ui/theme-toggle"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Bell,
  ChevronDown,
  Search,
  Menu,
  Settings,
  User,
  LogOut,
  HelpCircle,
} from "lucide-react"

interface HeaderProps {
  className?: string
  sidebarCollapsed?: boolean
  onSidebarToggle?: () => void
}

export function Header({ className, sidebarCollapsed, onSidebarToggle }: HeaderProps) {
  return (
    <header
      className={cn(
        "flex h-14 lg:h-16 items-center justify-between",
        "border-b border-border bg-card shrink-0",
        "px-4 sm:px-6 transition-all duration-normal ease-in-out",
        className
      )}
      role="banner"
    >
      {/* Left side */}
      <div className="flex items-center gap-2 sm:gap-4">
        {/* Mobile menu toggle */}
        <Button
          variant="ghost"
          size="icon"
          onClick={onSidebarToggle}
          className={cn(
            "h-8 w-8 text-foreground hover:bg-accent",
            "btn-hover-lift focus-enhanced tap-target",
            "lg:hidden"
          )}
          aria-label="Toggle sidebar menu"
          aria-expanded={!sidebarCollapsed}
        >
          <Menu className="h-4 w-4" />
        </Button>

        {/* Current page indicator */}
        <div className="hidden sm:flex items-center gap-2">
          <Badge
            variant="outline"
            className="text-xs font-mono fade-in"
            aria-label="Current page: Dashboard"
          >
            Dashboard
          </Badge>
          <Separator
            orientation="vertical"
            className="h-4 lg:h-6 bg-border fade-in"
          />
        </div>
      </div>

      {/* Center - Search (hidden on mobile, visible on larger screens) */}
      <div className="hidden md:flex flex-1 max-w-md mx-4 lg:mx-8">
        <div className="relative w-full">
          <Search
            className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
            aria-hidden="true"
          />
          <input
            type="text"
            placeholder="Search projects, pipelines, or team..."
            className={cn(
              "w-full rounded-md bg-input border-border",
              "pl-10 pr-4 py-2 text-sm text-foreground",
              "placeholder:text-muted-foreground",
              "focus:outline-none focus:ring-2 focus:ring-ring",
              "focus:ring-offset-2 focus:ring-offset-card",
              "transition-all duration-fast",
              "font-mono"
            )}
            aria-label="Search"
          />
        </div>
      </div>

      {/* Right side */}
      <div className="flex items-center gap-1 sm:gap-2 lg:gap-4">
        {/* Help/Support */}
        <Button
          variant="ghost"
          size="icon"
          className={cn(
            "h-8 w-8 text-foreground hover:bg-accent",
            "btn-hover-lift focus-enhanced tap-target",
            "hidden sm:flex"
          )}
          aria-label="Help and support"
        >
          <HelpCircle className="h-4 w-4" />
        </Button>

        {/* Notifications */}
        <Button
          variant="ghost"
          size="icon"
          className={cn(
            "relative h-8 w-8 text-foreground hover:bg-accent",
            "btn-hover-lift focus-enhanced tap-target"
          )}
          aria-label="Notifications"
          aria-describedby="notification-count"
        >
          <Bell className="h-4 w-4" />
          <span
            id="notification-count"
            className="absolute -top-1 -right-1 flex h-3 w-3 lg:h-4 lg:w-4 items-center justify-center rounded-full bg-destructive text-xs text-destructive-foreground font-mono"
            aria-label="2 notifications"
          >
            2
          </span>
        </Button>

        {/* Theme Toggle */}
        <div className="hidden sm:block">
          <ThemeToggle />
        </div>

        {/* User Menu */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              className={cn(
                "flex items-center gap-2 h-8 px-2 text-foreground hover:bg-accent",
                "btn-hover-lift focus-enhanced tap-target"
              )}
              aria-label="User menu"
              aria-expanded={false}
            >
              <Avatar className="h-8 w-8">
                <AvatarImage
                  src="/avatars/user.jpg"
                  alt="User avatar"
                  className="rounded-full"
                />
                <AvatarFallback className="rounded-full bg-primary text-primary-foreground font-mono text-sm">
                  JD
                </AvatarFallback>
              </Avatar>
              <span className="hidden lg:block text-sm font-mono truncate max-w-24">
                John Doe
              </span>
              <ChevronDown className="h-4 w-4 hidden lg:block" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            align="end"
            className="w-56 font-mono text-sm"
            sideOffset={8}
          >
            <DropdownMenuLabel className="text-sm font-medium">
              <div className="flex flex-col space-y-1">
                <p className="font-mono">John Doe</p>
                <p className="text-xs text-muted-foreground font-mono truncate">
                  john@example.com
                </p>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              className="text-foreground hover:bg-accent cursor-pointer interactive-item tap-target"
              aria-label="Go to profile"
            >
              <User className="mr-2 h-4 w-4" aria-hidden="true" />
              <span>Profile</span>
            </DropdownMenuItem>
            <DropdownMenuItem
              className="text-foreground hover:bg-accent cursor-pointer interactive-item tap-target"
              aria-label="Go to settings"
            >
              <Settings className="mr-2 h-4 w-4" aria-hidden="true" />
              <span>Settings</span>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              className="text-destructive hover:bg-destructive/10 cursor-pointer interactive-item tap-target"
              aria-label="Log out"
            >
              <LogOut className="mr-2 h-4 w-4" aria-hidden="true" />
              <span>Log out</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}