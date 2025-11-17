import { motion, AnimatePresence } from 'framer-motion'
import { useState, memo, useMemo } from 'react'
import { Search, ArrowUpDown, ExternalLink } from 'lucide-react'
import type { Run } from '@/types'
import { cn } from '@/lib/utils'

interface RunsTableProps {
  runs: Run[]
}

type SortField = 'Provider' | 'Repo' | 'Status' | 'UpdatedAt'
type SortDirection = 'asc' | 'desc'

interface RunRowProps {
  run: Run
}

// Memoized row component to prevent unnecessary re-renders
const RunRow = memo(({ run }: RunRowProps) => {
  return (
    <motion.tr
      layout
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.15, layout: { duration: 0.2 } }}
      className="border-b border-border last:border-0 hover:bg-muted/30 transition-colors"
    >
      <td className="px-4 py-3 text-sm font-medium">
        {run.Provider}
      </td>
      <td className="px-4 py-3 text-sm">
        <a
          href={run.URL}
          target="_blank"
          rel="noopener noreferrer"
          className="hover:text-primary hover:underline"
        >
          {run.Repo}
        </a>
      </td>
      <td className="px-4 py-3 text-sm text-muted-foreground max-w-xs truncate">
        {run.WorkflowName}
      </td>
      <td className="px-4 py-3">
        <span
          className={cn(
            'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
            getStatusColor(run.Status, run.Conclusion)
          )}
        >
          {run.Conclusion || run.Status}
        </span>
      </td>
      <td className="px-4 py-3 text-sm text-muted-foreground">
        {formatDuration(run.Duration)}
      </td>
      <td className="px-4 py-3 text-sm text-muted-foreground">
        {formatTime(run.UpdatedAt)}
      </td>
      <td className="px-4 py-3 text-sm">
        <a
          href={run.URL}
          target="_blank"
          rel="noopener noreferrer"
          data-testid="external-link-icon"
          className="inline-flex items-center gap-1 text-muted-foreground hover:text-primary transition-colors"
          title="View on GitHub"
        >
          <ExternalLink className="h-4 w-4" />
        </a>
      </td>
    </motion.tr>
  )
})

function getStatusColor(status: string, conclusion: string): string {
  if (status === 'in_progress') return 'text-purple-400 bg-purple-400/10'
  if (status === 'queued') return 'text-amber-400 bg-amber-400/10'
  if (status === 'completed' && conclusion === 'success') return 'text-green-400 bg-green-400/10'
  if (status === 'failed' || conclusion === 'failure') return 'text-red-400 bg-red-400/10'
  if (status === 'cancelled') return 'text-gray-400 bg-gray-400/10'
  return 'text-muted-foreground bg-muted'
}

function formatDuration(ms: number): string {
  if (ms === 0) return '-'
  const seconds = Math.floor(ms / 1000000000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  
  if (hours > 0) return `${hours}h ${minutes % 60}m`
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`
  return `${seconds}s`
}

function formatTime(timestamp: string): string {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  
  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  const diffHours = Math.floor(diffMins / 60)
  if (diffHours < 24) return `${diffHours}h ago`
  const diffDays = Math.floor(diffHours / 24)
  return `${diffDays}d ago`
}

export function RunsTable({ runs }: RunsTableProps) {
  const [search, setSearch] = useState('')
  const [sortField, setSortField] = useState<SortField>('UpdatedAt')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortDirection('asc')
    }
  }

  // Memoize expensive filtering and sorting operations
  const filteredRuns = useMemo(() => {
    return runs.filter(run => {
      const searchLower = search.toLowerCase()
      return (
        run.Repo.toLowerCase().includes(searchLower) ||
        run.Provider.toLowerCase().includes(searchLower) ||
        run.WorkflowName.toLowerCase().includes(searchLower) ||
        run.Branch.toLowerCase().includes(searchLower)
      )
    })
  }, [runs, search])

  const sortedRuns = useMemo(() => {
    return [...filteredRuns].sort((a, b) => {
      let aVal: any = a[sortField]
      let bVal: any = b[sortField]
      
      if (sortDirection === 'asc') {
        return aVal > bVal ? 1 : -1
      }
      return aVal < bVal ? 1 : -1
    })
  }, [filteredRuns, sortField, sortDirection])

  if (runs.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-card p-12 text-center">
        <p className="text-muted-foreground">No runs yet. Waiting for CI data...</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          type="text"
          placeholder="Search runs..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full rounded-lg border border-border bg-background pl-10 pr-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      {/* Table */}
      <div className="rounded-lg border border-border bg-card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="border-b border-border bg-muted/50">
              <tr>
                <th className="px-4 py-3 text-left">
                  <button
                    onClick={() => handleSort('Provider')}
                    className="flex items-center gap-1 text-sm font-medium hover:text-foreground"
                  >
                    Provider
                    <ArrowUpDown className="h-3 w-3" />
                  </button>
                </th>
                <th className="px-4 py-3 text-left">
                  <button
                    onClick={() => handleSort('Repo')}
                    className="flex items-center gap-1 text-sm font-medium hover:text-foreground"
                  >
                    Repo
                    <ArrowUpDown className="h-3 w-3" />
                  </button>
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium">
                  Workflow
                </th>
                <th className="px-4 py-3 text-left">
                  <button
                    onClick={() => handleSort('Status')}
                    className="flex items-center gap-1 text-sm font-medium hover:text-foreground"
                  >
                    Status
                    <ArrowUpDown className="h-3 w-3" />
                  </button>
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium">
                  Duration
                </th>
                <th className="px-4 py-3 text-left">
                  <button
                    onClick={() => handleSort('UpdatedAt')}
                    className="flex items-center gap-1 text-sm font-medium hover:text-foreground"
                  >
                    Time
                    <ArrowUpDown className="h-3 w-3" />
                  </button>
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium">
                  Link
                </th>
              </tr>
            </thead>
            <tbody>
              <AnimatePresence mode="popLayout">
                {sortedRuns.map((run) => (
                  <RunRow key={run.ID} run={run} />
                ))}
              </AnimatePresence>
            </tbody>
          </table>
        </div>
      </div>

      {filteredRuns.length === 0 && search && (
        <p className="text-center text-sm text-muted-foreground py-4">
          No runs match "{search}"
        </p>
      )}
    </div>
  )
}
