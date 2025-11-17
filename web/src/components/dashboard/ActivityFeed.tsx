import { motion, AnimatePresence } from 'framer-motion'
import { CheckCircle2, XCircle, Clock, Activity as ActivityIcon } from 'lucide-react'
import { useState } from 'react'

interface ActivityItem {
  id: string
  type: 'success' | 'failure' | 'started' | 'queued'
  message: string
  timestamp: Date
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

export function ActivityFeed({ items, maxItems = 10 }: ActivityFeedProps) {
  const [expanded, setExpanded] = useState(true)
  const displayItems = items.slice(0, maxItems)

  return (
    <div className="rounded-lg border border-border bg-card">
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
                <div className="divide-y divide-border">
                  {displayItems.map((item, index) => (
                    <motion.div
                      key={item.id}
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ duration: 0.2, delay: index * 0.03 }}
                      className="flex items-start gap-3 p-4 hover:bg-muted/20 transition-colors"
                    >
                      <div className="mt-0.5">{getIcon(item.type)}</div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm">{item.message}</p>
                        <p className="text-xs text-muted-foreground mt-1">
                          {formatTime(item.timestamp)}
                        </p>
                      </div>
                    </motion.div>
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
