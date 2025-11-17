import { StatsCards } from '@/components/dashboard/StatsCards'
import { RunsTable } from '@/components/dashboard/RunsTable'
import { ProviderCards } from '@/components/dashboard/ProviderCards'
import { ActivityFeed } from '@/components/dashboard/ActivityFeed'
import { useDashboard } from '@/contexts/DashboardContext'
import { Circle } from 'lucide-react'
import { cn } from '@/lib/utils'

function App() {
  const { runs, stats, providers, isConnected, lastUpdate } = useDashboard()

  // For activity feed, we'll generate from runs for now
  // In the future, this could be its own WebSocket stream
  const activityItems = runs.slice(0, 10).map((run) => ({
    id: run.ID,
    type: run.Status === 'completed' && run.Conclusion === 'success' 
      ? 'success' as const
      : run.Status === 'failed' || run.Conclusion === 'failure'
      ? 'failure' as const
      : run.Status === 'in_progress'
      ? 'started' as const
      : 'queued' as const,
    message: `${run.WorkflowName} • ${run.Repo}/${run.Branch}`,
    timestamp: new Date(run.UpdatedAt),
  }))

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="border-b border-border px-8 py-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">🔍 ceye</h1>
            <p className="text-sm text-muted-foreground mt-1">
              CI/CD Monitoring Dashboard
            </p>
          </div>
          <div className="flex items-center gap-2 text-sm">
            <Circle
              className={cn(
                'h-2 w-2',
                isConnected ? 'fill-green-400 text-green-400' : 'fill-red-400 text-red-400'
              )}
            />
            <span className="text-muted-foreground">
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
            {lastUpdate && (
              <span className="text-muted-foreground">
                • {lastUpdate.toLocaleTimeString()}
              </span>
            )}
          </div>
        </div>
      </header>
      <main className="container mx-auto p-8 space-y-8">
        <StatsCards stats={stats} />
        
        <div className="grid gap-8 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <RunsTable runs={runs} />
          </div>
          <div className="space-y-8">
            <ProviderCards providers={providers} />
            <ActivityFeed items={activityItems} />
          </div>
        </div>
      </main>
    </div>
  )
}

export default App
