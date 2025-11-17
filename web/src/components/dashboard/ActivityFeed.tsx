import { motion, AnimatePresence } from 'framer-motion'
import { CheckCircle2, XCircle, Clock, Activity as ActivityIcon, ExternalLink, ChevronDown, ChevronUp } from 'lucide-react'
import React, { useState } from 'react'

interface ActivityItem {
  id: string
  type: 'success' | 'failure' | 'started' | 'queued'
  message: string
  timestamp: Date
  // Rich data
  duration?: number  // Duration in seconds
  commitSHA?: string
  commitMessage?: string
  branch?: string
  repo?: string
  conclusion?: string
  url?: string
}

interface ActivityItemRowProps {
  item: ActivityItem
  index: number
}

const ActivityItemRow = React.memo(function ActivityItemRow({ item, index }: ActivityItemRowProps) {
  const [isExpanded, setIsExpanded] = useState(false)

  return (
    <motion.div
      key={item.id}
      data-testid="activity-item"
      layout
      transition={{ duration: 0.2, delay: index * 0.03 }}
      className="border-b border-border last:border-0"
    >
      <div 
        className="flex items-start gap-3 p-4 hover:bg-muted/20 transition-colors cursor-pointer"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="mt-0.5">{getIcon(item.type)}</div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium">{item.message}</p>
            <div className="flex items-center gap-2">
              {item.url && (
                <a
                  href={item.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  data-testid="activity-external-link"
                  className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
                  title="View on GitHub"
                  onClick={(e) => e.stopPropagation()}
                >
                  <ExternalLink className="h-3 w-3" />
                </a>
              )}
              {isExpanded ? (
                <ChevronUp className="h-4 w-4 text-muted-foreground" />
              ) : (
                <ChevronDown className="h-4 w-4 text-muted-foreground" />
              )}
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2 mt-1 text-xs text-muted-foreground">
            {item.duration !== undefined && item.duration > 0 && (
              <span>Duration: {formatDuration(item.duration)}</span>
            )}
            {item.duration !== undefined && item.duration > 0 && item.commitSHA && (
              <span>•</span>
            )}
            {item.commitSHA && (
              <span className="font-mono">{item.commitSHA.substring(0, 7)}</span>
            )}
            {item.commitMessage && (
              <span>•</span>
            )}
            {item.commitMessage && (
              <span className="italic truncate max-w-xs">
                "{item.commitMessage.split('\n')[0]}"
              </span>
            )}
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            {formatTime(item.timestamp)}
          </p>
        </div>
      </div>
      
      <AnimatePresence>
        {isExpanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="overflow-hidden bg-muted/30 px-4 py-3 border-t border-border"
          >
            <div className="text-xs space-y-2">
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <span className="font-semibold text-muted-foreground">Status:</span>
                  <span className="ml-2">{item.type}</span>
                </div>
                {item.conclusion && (
                  <div>
                    <span className="font-semibold text-muted-foreground">Conclusion:</span>
                    <span className="ml-2">{item.conclusion}</span>
                  </div>
                )}
                {item.repo && (
                  <div>
                    <span className="font-semibold text-muted-foreground">Repository:</span>
                    <span className="ml-2">{item.repo}</span>
                  </div>
                )}
                {item.branch && (
                  <div>
                    <span className="font-semibold text-muted-foreground">Branch:</span>
                    <span className="ml-2">{item.branch}</span>
                  </div>
                )}
                {item.commitSHA && (
                  <div className="col-span-2">
                    <span className="font-semibold text-muted-foreground">Commit:</span>
                    <span className="ml-2 font-mono">{item.commitSHA}</span>
                  </div>
                )}
                {item.url && (
                  <div className="col-span-2">
                    <span className="font-semibold text-muted-foreground">URL:</span>
                    <a 
                      href={item.url} 
                      target="_blank" 
                      rel="noopener noreferrer"
                      className="ml-2 text-primary hover:underline break-all"
                      onClick={(e) => e.stopPropagation()}
                    >
                      {item.url}
                    </a>
                  </div>
                )}
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  )
}

interface ActivityFeedProps {
  items: ActivityItem[]
  maxItems?: number
}

function getIcon(type: string) {
  switch (type) {
    case 'success':
      return <CheckCircle2 className="h-4 w-4 text-green-400" />
    case 'failure':
      return <XCircle className="h-4 w-4 text-red-400" />
    case 'started':
      return <ActivityIcon className="h-4 w-4 text-purple-400" />
    case 'queued':
      return <Clock className="h-4 w-4 text-amber-400" />
    default:
      return <ActivityIcon className="h-4 w-4 text-muted-foreground" />
  }
}

function formatTime(date: Date): string {
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function formatDuration(seconds: number): string {
  if (seconds < 60) {
    return `${Math.round(seconds)}s`
  }
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = Math.round(seconds % 60)
  if (minutes < 60) {
    return `${minutes}m ${remainingSeconds}s`
  }
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return `${hours}h ${remainingMinutes}m`
}

export function ActivityFeed({ items, maxItems = 10 }: ActivityFeedProps) {
  const [expanded, setExpanded] = useState(true)
  const displayItems = items.slice(0, maxItems)

  return (
    <div className="rounded-lg border border-border bg-card" data-testid="activity-feed">
      <div className="flex items-center justify-between border-b border-border p-4">
        <h2 className="text-lg font-semibold">Activity</h2>
        <button
          onClick={() => setExpanded(!expanded)}
          className="text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          {expanded ? 'Collapse' : 'Expand'}
        </button>
      </div>

      <AnimatePresence>
        {expanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="overflow-hidden"
          >
            <div className="max-h-96 overflow-y-auto">
              {displayItems.length === 0 ? (
                <div className="p-8 text-center">
                  <p className="text-sm text-muted-foreground">No activity yet</p>
                </div>
              ) : (
                <div>
                  {displayItems.map((item, index) => (
                    <ActivityItemRow key={item.id} item={item} index={index} />
                  ))}
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
