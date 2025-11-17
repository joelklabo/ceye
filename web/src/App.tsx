import { StatsCards } from '@/components/dashboard/StatsCards'
import { RunsTable } from '@/components/dashboard/RunsTable'
import { ProviderCards } from '@/components/dashboard/ProviderCards'
import { ActivityFeed } from '@/components/dashboard/ActivityFeed'
import type { StatsData, Run, ProviderHealth } from '@/types'

function App() {
  // Mock data for now - will be replaced with WebSocket data in Phase 0.3
  const stats: StatsData = {
    running: 3,
    queued: 2,
    success: 142,
    failed: 8,
  }

  const mockProviders: Record<string, ProviderHealth> = {
    github: {
      LastError: '',
      ErrorCount: 0,
      LastSuccess: new Date(Date.now() - 120000).toISOString(),
    },
  }

  const mockActivity = [
    {
      id: '1',
      type: 'success' as const,
      message: 'Build completed successfully for ceye/main',
      timestamp: new Date(Date.now() - 120000),
    },
    {
      id: '2',
      type: 'started' as const,
      message: 'CI Tests started for ceye/main',
      timestamp: new Date(Date.now() - 60000),
    },
    {
      id: '3',
      type: 'failure' as const,
      message: 'Deploy failed for api-server/develop',
      timestamp: new Date(Date.now() - 7200000),
    },
  ]

  const mockRuns: Run[] = [
    {
      ID: '1',
      Provider: 'github',
      Repo: 'ceye',
      WorkflowName: 'CI Tests',
      Status: 'in_progress',
      Conclusion: '',
      Branch: 'main',
      CommitSHA: 'abc123',
      StartedAt: new Date(Date.now() - 300000).toISOString(),
      UpdatedAt: new Date(Date.now() - 60000).toISOString(),
      Duration: 240000000000,
      URL: 'https://github.com/example/ceye/actions/runs/1',
    },
    {
      ID: '2',
      Provider: 'github',
      Repo: 'ceye',
      WorkflowName: 'Build',
      Status: 'completed',
      Conclusion: 'success',
      Branch: 'main',
      CommitSHA: 'def456',
      StartedAt: new Date(Date.now() - 3600000).toISOString(),
      UpdatedAt: new Date(Date.now() - 3300000).toISOString(),
      Duration: 180000000000,
      URL: 'https://github.com/example/ceye/actions/runs/2',
    },
    {
      ID: '3',
      Provider: 'github',
      Repo: 'api-server',
      WorkflowName: 'Deploy',
      Status: 'failed',
      Conclusion: 'failure',
      Branch: 'develop',
      CommitSHA: 'ghi789',
      StartedAt: new Date(Date.now() - 7200000).toISOString(),
      UpdatedAt: new Date(Date.now() - 7000000).toISOString(),
      Duration: 420000000000,
      URL: 'https://github.com/example/api-server/actions/runs/3',
    },
  ]

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="border-b border-border px-8 py-4">
        <h1 className="text-2xl font-bold">🔍 ceye</h1>
        <p className="text-sm text-muted-foreground mt-1">
          CI/CD Monitoring Dashboard
        </p>
      </header>
      <main className="container mx-auto p-8 space-y-8">
        <StatsCards stats={stats} />
        
        <div className="grid gap-8 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <RunsTable runs={mockRuns} />
          </div>
          <div className="space-y-8">
            <ProviderCards providers={mockProviders} />
            <ActivityFeed items={mockActivity} />
          </div>
        </div>
      </main>
    </div>
  )
}

export default App
