import * as React from "react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  ChevronDown,
  FolderOpen,
  Search,
  Plus,
  Star,
  Clock,
  Users,
  Check,
  Sparkles,
} from "lucide-react"

interface Project {
  id: string
  name: string
  description?: string
  role: 'owner' | 'maintainer' | 'viewer'
  status: 'active' | 'archived' | 'suspended'
  lastAccessed?: Date
  isFavorite?: boolean
  memberCount?: number
}

interface ProjectSwitcherProps {
  currentProject?: Project
  projects: Project[]
  onProjectSelect: (project: Project) => void
  onCreateProject?: () => void
  className?: string
  variant?: 'default' | 'compact'
}

// Mock data for development
const mockProjects: Project[] = [
  {
    id: '1',
    name: 'Data Pipeline',
    description: 'ETL pipeline for customer analytics',
    role: 'owner',
    status: 'active',
    lastAccessed: new Date('2024-01-15'),
    isFavorite: true,
    memberCount: 5,
  },
  {
    id: '2',
    name: 'ML Models',
    description: 'Machine learning model training',
    role: 'maintainer',
    status: 'active',
    lastAccessed: new Date('2024-01-10'),
    memberCount: 8,
  },
  {
    id: '3',
    name: 'Analytics Dashboard',
    description: 'Real-time analytics dashboard',
    role: 'viewer',
    status: 'active',
    lastAccessed: new Date('2024-01-08'),
    isFavorite: true,
  },
  {
    id: '4',
    name: 'Legacy Reports',
    description: 'Legacy reporting system',
    role: 'owner',
    status: 'archived',
    lastAccessed: new Date('2023-12-20'),
  },
]

export function ProjectSwitcher({
  currentProject,
  projects = mockProjects,
  onProjectSelect,
  onCreateProject,
  className,
  variant = 'default',
}: ProjectSwitcherProps) {
  const [searchQuery, setSearchQuery] = React.useState('')
  const [open, setOpen] = React.useState(false)

  // Filter projects based on search query
  const filteredProjects = React.useMemo(() => {
    if (!searchQuery) return projects

    return projects.filter(project =>
      project.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      project.description?.toLowerCase().includes(searchQuery.toLowerCase())
    )
  }, [projects, searchQuery])

  // Group projects by status and favorites
  const favoriteProjects = React.useMemo(() =>
    filteredProjects.filter(p => p.isFavorite && p.status === 'active'),
    [filteredProjects]
  )

  const recentProjects = React.useMemo(() =>
    filteredProjects
      .filter(p => p.status === 'active' && !p.isFavorite)
      .sort((a, b) => (b.lastAccessed?.getTime() || 0) - (a.lastAccessed?.getTime() || 0))
      .slice(0, 3),
    [filteredProjects]
  )

  const activeProjects = React.useMemo(() =>
    filteredProjects.filter(p => p.status === 'active'),
    [filteredProjects]
  )

  const archivedProjects = React.useMemo(() =>
    filteredProjects.filter(p => p.status === 'archived'),
    [filteredProjects]
  )

  const getRoleBadgeVariant = (role: Project['role']) => {
    switch (role) {
      case 'owner': return 'default'
      case 'maintainer': return 'secondary'
      case 'viewer': return 'outline'
      default: return 'outline'
    }
  }

  const ProjectItem = ({ project, isSelected = false }: { project: Project; isSelected?: boolean }) => (
    <div
      className={cn(
        "group interactive-item tap-target",
        "flex items-center gap-3 p-2 rounded-md",
        "hover:bg-accent cursor-pointer",
        "transition-all duration-150",
        isSelected && "bg-accent/50"
      )}
      onClick={() => {
        onProjectSelect(project)
        setOpen(false)
      }}
      role="option"
      aria-selected={isSelected}
    >
      <div className="flex items-center gap-2 flex-1 min-w-0">
        <FolderOpen className="h-4 w-4 text-muted-foreground flex-shrink-0" />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-mono text-sm truncate">{project.name}</span>
            {isSelected && (
              <Check className="h-3 w-3 text-primary flex-shrink-0" />
            )}
          </div>
          {project.description && variant === 'default' && (
            <p className="text-xs text-muted-foreground truncate">
              {project.description}
            </p>
          )}
        </div>
      </div>

      <div className="flex items-center gap-1 flex-shrink-0">
        {project.isFavorite && (
          <Star className="h-3 w-3 text-yellow-500 fill-current" aria-hidden="true" />
        )}
        <Badge variant={getRoleBadgeVariant(project.role)} className="text-xs">
          {project.role}
        </Badge>
      </div>
    </div>
  )

  if (variant === 'compact') {
    return (
      <DropdownMenu open={open} onOpenChange={setOpen}>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            className={cn(
              "h-8 px-2 justify-start text-foreground",
              "font-mono text-sm",
              "btn-hover-lift focus-enhanced tap-target",
              className
            )}
            aria-label="Switch project"
          >
            <FolderOpen className="h-4 w-4 mr-2" />
            <span className="truncate">
              {currentProject?.name || 'Select Project'}
            </span>
            <ChevronDown className="h-4 w-4 ml-auto" />
          </Button>
        </DropdownMenuTrigger>

        <DropdownMenuContent
          className="w-80 font-mono text-sm"
          sideOffset={8}
          align="start"
        >
          <DropdownMenuLabel className="text-xs font-medium text-muted-foreground">
            Switch Project
          </DropdownMenuLabel>

          <DropdownMenuSeparator />

          {favoriteProjects.length > 0 && (
            <>
              <div className="px-2 py-1 text-xs font-medium text-muted-foreground flex items-center gap-1">
                <Star className="h-3 w-3" />
                Favorites
              </div>
              {favoriteProjects.map(project => (
                <DropdownMenuItem
                  key={project.id}
                  asChild
                  className="p-0"
                >
                  <ProjectItem project={project} isSelected={currentProject?.id === project.id} />
                </DropdownMenuItem>
              ))}
              <DropdownMenuSeparator />
            </>
          )}

          {recentProjects.length > 0 && (
            <>
              <div className="px-2 py-1 text-xs font-medium text-muted-foreground flex items-center gap-1">
                <Clock className="h-3 w-3" />
                Recent
              </div>
              {recentProjects.map(project => (
                <DropdownMenuItem
                  key={project.id}
                  asChild
                  className="p-0"
                >
                  <ProjectItem project={project} isSelected={currentProject?.id === project.id} />
                </DropdownMenuItem>
              ))}
              <DropdownMenuSeparator />
            </>
          )}

          <DropdownMenuItem
            className="text-primary hover:bg-primary/10 cursor-pointer"
            onClick={() => {
              onCreateProject?.()
              setOpen(false)
            }}
          >
            <Plus className="mr-2 h-4 w-4" />
            Create New Project
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    )
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className={cn(
            "w-full justify-between text-left font-mono",
            "h-10 px-3",
            "btn-hover-lift focus-enhanced tap-target",
            className
          )}
          aria-label="Switch project"
        >
          <div className="flex items-center gap-2 min-w-0 flex-1">
            <FolderOpen className="h-4 w-4 text-muted-foreground flex-shrink-0" />
            <span className="truncate">
              {currentProject?.name || 'Select Project'}
            </span>
            {currentProject?.memberCount && (
              <div className="flex items-center gap-1 text-xs text-muted-foreground">
                <Users className="h-3 w-3" />
                {currentProject.memberCount}
              </div>
            )}
          </div>
          <ChevronDown className="h-4 w-4 text-muted-foreground flex-shrink-0" />
        </Button>
      </PopoverTrigger>

      <PopoverContent className="w-96 p-0" align="start" sideOffset={8}>
        <Command className="font-mono">
          <div className="flex items-center border-b px-3">
            <Search className="mr-2 h-4 w-4 text-muted-foreground shrink-0" />
            <CommandInput
              placeholder="Search projects..."
              value={searchQuery}
              onValueChange={setSearchQuery}
              className="flex-1 h-10 border-0 focus:ring-0"
              aria-label="Search projects"
            />
          </div>

          <CommandList>
            {filteredProjects.length === 0 ? (
              <CommandEmpty>No projects found.</CommandEmpty>
            ) : (
              <>
                {favoriteProjects.length > 0 && (
                  <CommandGroup heading="Favorites">
                    {favoriteProjects.map(project => (
                      <CommandItem key={project.id} asChild>
                        <ProjectItem project={project} isSelected={currentProject?.id === project.id} />
                      </CommandItem>
                    ))}
                  </CommandGroup>
                )}

                {recentProjects.length > 0 && (
                  <CommandGroup heading="Recent">
                    {recentProjects.map(project => (
                      <CommandItem key={project.id} asChild>
                        <ProjectItem project={project} isSelected={currentProject?.id === project.id} />
                      </CommandItem>
                    ))}
                  </CommandGroup>
                )}

                {activeProjects.length > favoriteProjects.length + recentProjects.length && (
                  <CommandGroup heading="All Projects">
                    {activeProjects
                      .filter(p => !p.isFavorite && !recentProjects.includes(p))
                      .map(project => (
                        <CommandItem key={project.id} asChild>
                          <ProjectItem project={project} isSelected={currentProject?.id === project.id} />
                        </CommandItem>
                      ))}
                  </CommandGroup>
                )}

                {archivedProjects.length > 0 && (
                  <>
                    <CommandSeparator />
                    <CommandGroup heading="Archived">
                      {archivedProjects.map(project => (
                        <CommandItem key={project.id} asChild>
                          <ProjectItem project={project} isSelected={currentProject?.id === project.id} />
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  </>
                )}

                <CommandSeparator />
                <CommandGroup>
                  <CommandItem
                    className="text-primary hover:bg-primary/10 cursor-pointer"
                    onSelect={() => {
                      onCreateProject?.()
                      setOpen(false)
                    }}
                  >
                    <Plus className="mr-2 h-4 w-4" />
                    <span>Create New Project</span>
                  </CommandItem>
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}