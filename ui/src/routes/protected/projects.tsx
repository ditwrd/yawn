import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { FolderOpen, Plus, Users, Star, Clock } from 'lucide-react'

export const Route = createFileRoute('/protected/projects')({
  component: Projects,
})

function Projects() {
  const projects = [
    {
      id: '1',
      name: 'Data Pipeline',
      description: 'ETL pipeline for customer analytics',
      status: 'active',
      role: 'owner',
      isFavorite: true,
      members: 5,
      lastUpdated: '2 hours ago',
    },
    {
      id: '2',
      name: 'ML Models',
      description: 'Machine learning model training',
      status: 'active',
      role: 'maintainer',
      isFavorite: false,
      members: 8,
      lastUpdated: '1 day ago',
    },
    {
      id: '3',
      name: 'Analytics Dashboard',
      description: 'Real-time analytics dashboard',
      status: 'active',
      role: 'viewer',
      isFavorite: true,
      members: 3,
      lastUpdated: '3 hours ago',
    },
    {
      id: '4',
      name: 'Legacy Reports',
      description: 'Legacy reporting system',
      status: 'archived',
      role: 'owner',
      isFavorite: false,
      members: 2,
      lastUpdated: '1 week ago',
    },
  ]

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active': return <Badge variant="default" className="text-xs">Active</Badge>
      case 'archived': return <Badge variant="secondary" className="text-xs">Archived</Badge>
      default: return <Badge variant="outline" className="text-xs">Unknown</Badge>
    }
  }

  const getRoleBadge = (role: string) => {
    switch (role) {
      case 'owner': return <Badge variant="default" className="text-xs">Owner</Badge>
      case 'maintainer': return <Badge variant="secondary" className="text-xs">Maintainer</Badge>
      case 'viewer': return <Badge variant="outline" className="text-xs">Viewer</Badge>
      default: return <Badge variant="outline" className="text-xs">Unknown</Badge>
    }
  }

  return (
    <div className="space-y-6 fade-in">
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <h1 className="text-responsive-2xl font-semibold text-foreground font-mono">
            Projects
          </h1>
          <p className="text-muted-foreground font-mono">
            Manage your data projects and collaborations
          </p>
        </div>

        <Button className="btn-hover-lift font-mono">
          <Plus className="h-4 w-4 mr-2" />
          New Project
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 lg:gap-6">
        {projects.map((project) => (
          <Card key={project.id} className="card-hover">
            <CardHeader className="pb-3">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-2 min-w-0">
                  <FolderOpen className="h-5 w-5 text-primary flex-shrink-0" />
                  <CardTitle className="text-primary font-mono truncate">
                    {project.name}
                  </CardTitle>
                  {project.isFavorite && (
                    <Star className="h-4 w-4 text-yellow-500 fill-current flex-shrink-0" />
                  )}
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-3">
              <p className="text-sm text-muted-foreground font-mono line-clamp-2">
                {project.description}
              </p>

              <div className="flex items-center gap-2">
                {getStatusBadge(project.status)}
                {getRoleBadge(project.role)}
              </div>

              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <div className="flex items-center gap-1">
                  <Users className="h-3 w-3" />
                  <span className="font-mono">{project.members}</span>
                </div>
                <div className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  <span className="font-mono">{project.lastUpdated}</span>
                </div>
              </div>

              <div className="flex gap-2 pt-2">
                <Button size="sm" variant="outline" className="flex-1 font-mono text-xs">
                  View
                </Button>
                <Button size="sm" className="flex-1 font-mono text-xs">
                  Open
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}