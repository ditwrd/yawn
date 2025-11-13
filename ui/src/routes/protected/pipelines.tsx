import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { FileText, Play, Pause, CheckCircle, XCircle, Clock } from 'lucide-react'

export const Route = createFileRoute('/protected/pipelines')({
  component: Pipelines,
})

function Pipelines() {
  const pipelines = [
    {
      id: '1',
      name: 'Customer Data ETL',
      description: 'Extract, transform, and load customer data from multiple sources',
      status: 'running',
      progress: 75,
      lastRun: '10 minutes ago',
      duration: '5m 23s',
      successRate: '98.5%',
    },
    {
      id: '2',
      name: 'ML Model Training',
      description: 'Daily training of recommendation models',
      status: 'completed',
      progress: 100,
      lastRun: '2 hours ago',
      duration: '1h 45m',
      successRate: '95.2%',
    },
    {
      id: '3',
      name: 'Analytics Report Generation',
      description: 'Generate weekly analytics reports',
      status: 'failed',
      progress: 45,
      lastRun: '1 day ago',
      duration: 'Failed',
      successRate: '89.1%',
    },
    {
      id: '4',
      name: 'Data Quality Checks',
      description: 'Validate data quality and integrity',
      status: 'idle',
      progress: 0,
      lastRun: '3 days ago',
      duration: '12m 45s',
      successRate: '99.8%',
    },
  ]

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running': return <Play className="h-4 w-4 text-green-500" />
      case 'completed': return <CheckCircle className="h-4 w-4 text-blue-500" />
      case 'failed': return <XCircle className="h-4 w-4 text-red-500" />
      case 'idle': return <Pause className="h-4 w-4 text-gray-500" />
      default: return <Clock className="h-4 w-4 text-gray-500" />
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'running': return <Badge variant="default" className="text-xs">Running</Badge>
      case 'completed': return <Badge variant="secondary" className="text-xs">Completed</Badge>
      case 'failed': return <Badge variant="destructive" className="text-xs">Failed</Badge>
      case 'idle': return <Badge variant="outline" className="text-xs">Idle</Badge>
      default: return <Badge variant="outline" className="text-xs">Unknown</Badge>
    }
  }

  return (
    <div className="space-y-6 fade-in">
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <h1 className="text-responsive-2xl font-semibold text-foreground font-mono">
            Pipelines
          </h1>
          <p className="text-muted-foreground font-mono">
            Monitor and manage your data pipelines
          </p>
        </div>

        <Button className="btn-hover-lift font-mono">
          <FileText className="h-4 w-4 mr-2" />
          New Pipeline
        </Button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 lg:gap-6">
        {pipelines.map((pipeline) => (
          <Card key={pipeline.id} className="card-hover">
            <CardHeader className="pb-3">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-2 min-w-0">
                  {getStatusIcon(pipeline.status)}
                  <CardTitle className="text-primary font-mono truncate">
                    {pipeline.name}
                  </CardTitle>
                </div>
                {getStatusBadge(pipeline.status)}
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm text-muted-foreground font-mono line-clamp-2">
                {pipeline.description}
              </p>

              {pipeline.status === 'running' && (
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-mono">Progress</span>
                    <span className="font-mono">{pipeline.progress}%</span>
                  </div>
                  <Progress value={pipeline.progress} className="h-2" />
                </div>
              )}

              <div className="grid grid-cols-2 gap-4 text-xs">
                <div>
                  <span className="text-muted-foreground font-mono block">Last Run</span>
                  <span className="font-mono">{pipeline.lastRun}</span>
                </div>
                <div>
                  <span className="text-muted-foreground font-mono block">Duration</span>
                  <span className="font-mono">{pipeline.duration}</span>
                </div>
                <div>
                  <span className="text-muted-foreground font-mono block">Success Rate</span>
                  <span className="font-mono text-green-500">{pipeline.successRate}</span>
                </div>
                <div>
                  <span className="text-muted-foreground font-mono block">Status</span>
                  <div className="flex items-center gap-1">
                    {getStatusIcon(pipeline.status)}
                    <span className="font-mono capitalize">{pipeline.status}</span>
                  </div>
                </div>
              </div>

              <div className="flex gap-2 pt-2">
                <Button
                  size="sm"
                  variant="outline"
                  className="flex-1 font-mono text-xs"
                >
                  View Logs
                </Button>
                <Button
                  size="sm"
                  className="flex-1 font-mono text-xs"
                  disabled={pipeline.status === 'running'}
                >
                  {pipeline.status === 'running' ? 'Running...' : 'Run Now'}
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}