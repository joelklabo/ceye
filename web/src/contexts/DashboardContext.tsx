import { createContext, useContext, type ReactNode } from 'react'
import { useWebSocket } from '@/hooks/useWebSocket'
import type { Run, ProviderHealth, ProviderMeta, StatsData } from '@/types'

interface DashboardContextValue {
  runs: Run[]
  stats: StatsData
  providers: Record<string, ProviderHealth>
  meta: Record<string, ProviderMeta>
  isConnected: boolean
  isLoading: boolean
  error: string | null
  lastUpdate: Date | null
}

const DashboardContext = createContext<DashboardContextValue | null>(null)

interface DashboardProviderProps {
  children: ReactNode
}

function calculateStats(runs: Run[]): StatsData {
  const stats: StatsData = {
    running: 0,
    queued: 0,
    success: 0,
    failed: 0,
  }

  runs.forEach((run) => {
    if (run.Status === 'in_progress') {
      stats.running++
    } else if (run.Status === 'queued') {
      stats.queued++
    } else if (run.Status === 'completed' && run.Conclusion === 'success') {
      stats.success++
    } else if (run.Status === 'failed' || run.Conclusion === 'failure') {
      stats.failed++
    }
  })

  return stats
}

export function DashboardProvider({ children }: DashboardProviderProps) {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws`

  const { isConnected, lastMessage, error } = useWebSocket({
    url: wsUrl,
    onMessage: (message) => {
      console.log('Dashboard message received:', message.type)
    },
  })

  const runs = lastMessage?.runs || []
  const stats = lastMessage?.totals
    ? {
        running: lastMessage.totals.running || 0,
        queued: lastMessage.totals.queued || 0,
        success: lastMessage.totals.success || 0,
        failed: lastMessage.totals.failed || 0,
      }
    : calculateStats(runs)
  
  const providers = lastMessage?.health || {}
  const meta = lastMessage?.meta || {}
  const lastUpdate = lastMessage ? new Date(lastMessage.timestamp) : null

  // Loading state: we're loading if connected but no data yet
  const isLoading = isConnected && !lastMessage

  const value: DashboardContextValue = {
    runs,
    stats,
    providers,
    meta,
    isConnected,
    isLoading,
    error,
    lastUpdate,
  }

  return (
    <DashboardContext.Provider value={value}>
      {children}
    </DashboardContext.Provider>
  )
}

export function useDashboard() {
  const context = useContext(DashboardContext)
  if (!context) {
    throw new Error('useDashboard must be used within a DashboardProvider')
  }
  return context
}
