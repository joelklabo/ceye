import { useEventStream } from '../../hooks/useEventStream'
import { TimelineEvent } from '../../types'
import { motion, AnimatePresence } from 'framer-motion'
import { Trash2, ChevronDown, ChevronRight, Circle } from 'lucide-react'
import { useState, useCallback } from 'react'

export function EventTimeline() {
  const { events, isConnected, clearEvents } = useEventStream({ enabled: true })
  const [expandedEvent, setExpandedEvent] = useState<string | null>(null)

  const getCategoryColor = useCallback((category: string) => {
    switch (category) {
      case 'WebSocket':
        return 'text-blue-400'
      case 'Webhook':
        return 'text-purple-400'
      case 'Polling':
        return 'text-yellow-400'
      case 'UI':
        return 'text-green-400'
      case 'Error':
        return 'text-red-400'
      default:
        return 'text-gray-400'
    }
  }, [])

  return (
    <div className="p-4 space-y-2">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <div className="text-sm text-muted-foreground">
            {events.length} event {events.length === 1 ? 'entry' : 'entries'}
          </div>
          <div className="flex items-center gap-1 text-xs">
            <Circle
              className={`h-2 w-2 ${isConnected ? 'fill-green-500 text-green-500' : 'fill-red-500 text-red-500'}`}
            />
            <span className={isConnected ? 'text-green-500' : 'text-red-500'}>
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
          </div>
        </div>
        <button
          onClick={clearEvents}
          className="flex items-center gap-1 px-2 py-1 text-xs bg-muted hover:bg-muted/80 rounded"
        >
          <Trash2 className="h-3 w-3" />
          Clear
        </button>
      </div>

      {events.length === 0 ? (
        <div className="text-center py-8 text-sm text-muted-foreground">
          No events yet. Events will appear here in real-time...
        </div>
      ) : (
        <div className="space-y-1 font-mono text-xs">
          {events.map((event: TimelineEvent) => (
            <motion.div
              key={event.id}
              initial={{ opacity: 0, x: -10 }}
              animate={{ opacity: 1, x: 0 }}
              className="rounded border border-border bg-card/50 overflow-hidden"
            >
              <button
                onClick={() => setExpandedEvent(expandedEvent === event.id ? null : event.id)}
                className="w-full px-3 py-2 text-left hover:bg-muted/20 transition-colors"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    {expandedEvent === event.id ? (
                      <ChevronDown className="h-3 w-3" />
                    ) : (
                      <ChevronRight className="h-3 w-3' />
                    )}
                    <span className="text-xs text-muted-foreground shrink-0">
                      {new Date(event.timestamp).toLocaleTimeString()}
                    </span>
                    <span className={`font-semibold shrink-0 ${getCategoryColor(event.category)}`}>
                      [{event.category}]
                    </span>
                    <span className="text-foreground break-all">{event.message}</span>
                  </div>
                </div>
              </button>
              
              <AnimatePresence>
                {expandedEvent === event.id && event.details && (
                  <motion.div
                    initial={{ height: 0 }}
                    animate={{ height: 'auto' }}
                    exit={{ height: 0 }}
                    transition={{ duration: 0.2 }}
                    className="overflow-hidden"
                  >
                    <pre className="text-[10px] p-3 bg-muted/50 overflow-x-auto">
                      {JSON.stringify(event.details, null, 2)}
                    </pre>
                  </motion.div>
                )}
              </AnimatePresence>
            </motion.div>
          ))}
        </div>
      )}
    </div>
  )
}
