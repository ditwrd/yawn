import { createFileRoute } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Moon } from 'lucide-react'

export const Route = createFileRoute('/')({
  component: App,
})

function App() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <main className="container-responsive py-12 fade-in">
        <div className="max-w-4xl mx-auto space-y-12">
          {/* Hero Section */}
          <div className="text-center space-y-6">
            <div className="flex justify-center">
              <div className="p-4 bg-primary/10 rounded-full">
                <Moon className="h-12 w-12 text-primary" />
              </div>
            </div>
            <div className="space-y-4">
              <h1 className="text-responsive-2xl font-bold font-mono">
                Welcome to YAWN
              </h1>
              <p className="text-responsive-lg text-muted-foreground font-mono max-w-2xl mx-auto">
                A data workflow platform with the sleepy YAWN aesthetic - built with
                <span className="text-primary"> purple/indigo accents</span> and
                <span className="text-primary"> JetBrains Mono</span> typography.
              </p>
            </div>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Button className="btn-hover-lift font-mono">
                Get Started
              </Button>
              <Button variant="outline" className="font-mono">
                View Documentation
              </Button>
            </div>
          </div>

          {/* Design System Showcase */}
          <div className="space-y-8">
            <h2 className="text-responsive-xl font-semibold text-center font-mono">
              Design System Features
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <Card className="card-hover">
                <CardHeader>
                  <CardTitle className="text-primary font-mono">Dark Theme</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-muted-foreground font-mono text-sm">
                    Sleepy YAWN aesthetic with purple/indigo accents and high contrast ratios
                  </p>
                </CardContent>
              </Card>

              <Card className="card-hover">
                <CardHeader>
                  <CardTitle className="text-primary font-mono">Typography</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-muted-foreground font-mono text-sm">
                    JetBrains Mono throughout with responsive text scaling and tabular numbers
                  </p>
                </CardContent>
              </Card>

              <Card className="card-hover">
                <CardHeader>
                  <CardTitle className="text-primary font-mono">Mobile First</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-muted-foreground font-mono text-sm">
                    Responsive design system with mobile-first breakpoints and touch-friendly targets
                  </p>
                </CardContent>
              </Card>

              <Card className="card-hover">
                <CardHeader>
                  <CardTitle className="text-primary font-mono">Accessibility</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-muted-foreground font-mono text-sm">
                    WCAG 2.1 AA compliant with focus management and keyboard navigation
                  </p>
                </CardContent>
              </Card>

              <Card className="card-hover">
                <CardHeader>
                  <CardTitle className="text-primary font-mono">Micro-interactions</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-muted-foreground font-mono text-sm">
                    Subtle animations and transitions with reduced motion support
                  </p>
                </CardContent>
              </Card>

              <Card className="card-hover">
                <CardHeader>
                  <CardTitle className="text-primary font-mono">Performance</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-muted-foreground font-mono text-sm">
                    Optimized CSS with efficient transitions and minimal bundle impact
                  </p>
                </CardContent>
              </Card>
            </div>
          </div>

          {/* Quick Test Section */}
          <div className="space-y-6">
            <h3 className="text-responsive-lg font-semibold text-center font-mono">
              Interactive Elements
            </h3>
            <div className="flex flex-wrap gap-3 justify-center">
              <Button size="sm" className="btn-hover-lift font-mono">Small</Button>
              <Button size="default" className="btn-hover-lift font-mono">Default</Button>
              <Button size="lg" className="btn-hover-lift font-mono">Large</Button>
              <Button variant="outline" className="font-mono">Outline</Button>
              <Button variant="secondary" className="font-mono">Secondary</Button>
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
