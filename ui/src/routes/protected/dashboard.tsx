import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export const Route = createFileRoute('/protected/dashboard')({
  component: Dashboard,
})

function Dashboard() {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-xl font-semibold text-white">Dashboard</h2>
        <p className="text-gray-400">Welcome to your protected dashboard</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <Card className="bg-black">
          <CardHeader>
            <CardTitle className="text-purple-400">Projects</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-white">12</div>
            <p className="text-gray-400">Active projects</p>
          </CardContent>
        </Card>

        <Card className="bg-black">
          <CardHeader>
            <CardTitle className="text-purple-400">Tasks</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-white">48</div>
            <p className="text-gray-400">Pending tasks</p>
          </CardContent>
        </Card>

        <Card className="bg-black">
          <CardHeader>
            <CardTitle className="text-purple-400">Team</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-white">6</div>
            <p className="text-gray-400">Team members</p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}