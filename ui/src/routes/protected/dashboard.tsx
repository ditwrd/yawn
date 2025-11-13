import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export const Route = createFileRoute('/protected/dashboard')({
  component: Dashboard,
})

function Dashboard() {
  return (
    <div id="main-content" className="space-y-6 fade-in">
      <div className="space-y-2">
        <h1 className="text-responsive-2xl font-semibold text-foreground font-mono">
          Dashboard
        </h1>
        <p className="text-muted-foreground font-mono">
          Welcome to your protected workspace
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 lg:gap-6">
        <Card className="card-hover">
          <CardHeader>
            <CardTitle className="text-primary font-mono">Projects</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-responsive-xl font-bold text-foreground font-mono">
              12
            </div>
            <p className="text-muted-foreground font-mono text-sm">Active projects</p>
          </CardContent>
        </Card>

        <Card className="card-hover">
          <CardHeader>
            <CardTitle className="text-primary font-mono">Tasks</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-responsive-xl font-bold text-foreground font-mono">
              48
            </div>
            <p className="text-muted-foreground font-mono text-sm">Pending tasks</p>
          </CardContent>
        </Card>

        <Card className="card-hover">
          <CardHeader>
            <CardTitle className="text-primary font-mono">Team</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-responsive-xl font-bold text-foreground font-mono">
              6
            </div>
            <p className="text-muted-foreground font-mono text-sm">Team members</p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}