import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import {
  User,
  Bell,
  Shield,
  Database,
  Moon,
  Key,
  HelpCircle,
  ChevronRight
} from 'lucide-react'

export const Route = createFileRoute('/protected/settings')({
  component: Settings,
})

function Settings() {
  const settingsSections = [
    {
      title: 'Profile',
      description: 'Manage your personal information',
      icon: User,
      href: '/settings/profile',
      badge: null,
    },
    {
      title: 'Notifications',
      description: 'Configure notification preferences',
      icon: Bell,
      href: '/settings/notifications',
      badge: '3',
    },
    {
      title: 'Security',
      description: 'Password and authentication settings',
      icon: Shield,
      href: '/settings/security',
      badge: null,
    },
    {
      title: 'API Keys',
      description: 'Manage your API keys and tokens',
      icon: Key,
      href: '/settings/api-keys',
      badge: null,
    },
    {
      title: 'Data & Privacy',
      description: 'Export or delete your data',
      icon: Database,
      href: '/settings/data',
      badge: null,
    },
    {
      title: 'Appearance',
      description: 'Theme and display preferences',
      icon: Moon,
      href: '/settings/appearance',
      badge: null,
    },
    {
      title: 'Help & Support',
      description: 'Get help and contact support',
      icon: HelpCircle,
      href: '/settings/help',
      badge: null,
    },
  ]

  return (
    <div className="space-y-6 fade-in">
      <div className="space-y-2">
        <h1 className="text-responsive-2xl font-semibold text-foreground font-mono">
          Settings
        </h1>
        <p className="text-muted-foreground font-mono">
          Manage your account settings and preferences
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Profile Summary */}
        <div className="lg:col-span-1">
          <Card>
            <CardHeader>
              <CardTitle className="text-primary font-mono">Profile Summary</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center gap-3">
                <div className="h-12 w-12 rounded-full bg-primary/10 flex items-center justify-center">
                  <User className="h-6 w-6 text-primary" />
                </div>
                <div className="min-w-0">
                  <h3 className="font-mono font-medium truncate">John Doe</h3>
                  <p className="text-sm text-muted-foreground font-mono truncate">john@example.com</p>
                </div>
              </div>

              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground font-mono">Member Since</span>
                  <span className="font-mono">Jan 2024</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground font-mono">Plan</span>
                  <Badge variant="default" className="text-xs">Pro</Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground font-mono">Projects</span>
                  <span className="font-mono">12</span>
                </div>
              </div>

              <Separator />

              <Button variant="outline" className="w-full font-mono">
                Edit Profile
              </Button>
            </CardContent>
          </Card>
        </div>

        {/* Settings Sections */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle className="text-primary font-mono">All Settings</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-1">
                {settingsSections.map((section) => (
                  <div
                    key={section.title}
                    className="group interactive-item tap-target p-3 rounded-lg hover:bg-accent cursor-pointer transition-all duration-150"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3 min-w-0">
                        <section.icon className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors flex-shrink-0" />
                        <div className="min-w-0">
                          <h4 className="font-mono font-medium text-foreground group-hover:text-primary transition-colors">
                            {section.title}
                          </h4>
                          <p className="text-sm text-muted-foreground font-mono">
                            {section.description}
                          </p>
                        </div>
                      </div>

                      <div className="flex items-center gap-2 flex-shrink-0">
                        {section.badge && (
                          <Badge variant="secondary" className="text-xs">
                            {section.badge}
                          </Badge>
                        )}
                        <ChevronRight className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle className="text-primary font-mono">Quick Actions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            <Button variant="outline" className="font-mono justify-start">
              <Key className="h-4 w-4 mr-2" />
              Generate API Key
            </Button>
            <Button variant="outline" className="font-mono justify-start">
              <Database className="h-4 w-4 mr-2" />
              Export Data
            </Button>
            <Button variant="outline" className="font-mono justify-start">
              <HelpCircle className="h-4 w-4 mr-2" />
              View Documentation
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}