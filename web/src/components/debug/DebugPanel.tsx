import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Bug, X, Trash2, ChevronDown, ChevronRight } from 'lucide-react'

interface WebSocketMessage {
  timestamp: Date
  type: string
  direction: 'sent' | 'received'
  data: any
}

export function DebugPanel() {
  const [isOpen, setIsOpen] = useState(() => {
    // Persist state in localStorage
    return localStorage.getItem('debugPanelOpen') === 'true'
  })
  const [activeTab, setActiveTab] = useState<'websocket' | 'logs' | 'events'>('websocket')
  const [messages, setMessages] = useState<WebSocketMessage[]>([])
  const [expandedMessage, setExpandedMessage] = useState<number | null>(null)

  // Persist panel state
  useEffect(() => {
    localStorage.setItem('debugPanelOpen', String(isOpen))
  }, [isOpen])

  // Intercept WebSocket messages
  useEffect(() => {
    if (!isOpen) return

    // Store original WebSocket
    const OriginalWebSocket = window.WebSocket
    
    // Override WebSocket
    window.WebSocket = function(url: string | URL, protocols?: string | string[]) {
      const ws = new OriginalWebSocket(url, protocols)
      
      // Intercept incoming messages
      ws.addEventListener('message', (event) => {
        try {
          const data = JSON.parse(event.data)
          setMessages(prev => [...prev, {
            timestamp: new Date(),
            type: data.Type || 'unknown',
            direction: 'received' as const,
            data: data
          }].slice(-50)) // Keep last 50 messages
        } catch (e) {
          // Not JSON, ignore
        }
      })

      return ws
    } as any

    return () => {
      // Restore original WebSocket
      window.WebSocket = OriginalWebSocket
    }
  }, [isOpen])

  const clearMessages = () => {
    setMessages([])
    setExpandedMessage(null)
  }

  return (
    <>
      {/* Toggle Button */}
      <motion.button
        onClick={() => setIsOpen(!isOpen)}
        className="fixed bottom-4 right-4 z-50 p-3 rounded-full bg-primary text-primary-foreground shadow-lg hover:shadow-xl transition-shadow"
        whileHover={{ scale: 1.1 }}
        whileTap={{ scale: 0.9 }}
        aria-label="Debug Panel"
      >
        <Bug className="h-5 w-5" />
      </motion.button>

      {/* Debug Panel */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ x: 400 }}
            animate={{ x: 0 }}
            exit={{ x: 400 }}
            transition={{ type: 'spring', damping: 25 }}
            className="fixed right-0 top-0 bottom-0 w-96 bg-card border-l border-border shadow-2xl z-40 flex flex-col"
          >
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b border-border">
              <h2 className="text-lg font-semibold flex items-center gap-2">
                <Bug className="h-5 w-5" />
                Debug Panel
              </h2>
              <button
                onClick={() => setIsOpen(false)}
                className="p-1 hover:bg-muted rounded"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            {/* Tabs */}
            <div className="flex border-b border-border">
              <button
                onClick={() => setActiveTab('websocket')}
                className={`flex-1 px-4 py-2 text-sm font-medium transition-colors ${
                  activeTab === 'websocket'
                    ? 'text-primary border-b-2 border-primary'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                WebSocket
              </button>
              <button
                onClick={() => setActiveTab('logs')}
                className={`flex-1 px-4 py-2 text-sm font-medium transition-colors ${
                  activeTab === 'logs'
                    ? 'text-primary border-b-2 border-primary'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                Logs
              </button>
              <button
                onClick={() => setActiveTab('events')}
                className={`flex-1 px-4 py-2 text-sm font-medium transition-colors ${
                  activeTab === 'events'
                    ? 'text-primary border-b-2 border-primary'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                Events
              </button>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-y-auto">
              {activeTab === 'websocket' && (
                <div className="p-4 space-y-2">
                  <div className="flex items-center justify-between mb-4">
                    <div className="text-sm text-muted-foreground">
                      {messages.length} messages
                    </div>
                    <button
                      onClick={clearMessages}
                      className="flex items-center gap-1 px-2 py-1 text-xs bg-muted hover:bg-muted/80 rounded"
                    >
                      <Trash2 className="h-3 w-3" />
                      Clear
                    </button>
                  </div>

                  {messages.length === 0 ? (
                    <div className="text-center py-8 text-sm text-muted-foreground">
                      No messages yet. Waiting for WebSocket activity...
                    </div>
                  ) : (
                    <div className="space-y-2">
                      {messages.map((msg, index) => (
                        <motion.div
                          key={index}
                          initial={{ opacity: 0, y: -10 }}
                          animate={{ opacity: 1, y: 0 }}
                          className="rounded border border-border bg-card/50 overflow-hidden"
                        >
                          <button
                            onClick={() => setExpandedMessage(expandedMessage === index ? null : index)}
                            className="w-full px-3 py-2 text-left hover:bg-muted/20 transition-colors"
                          >
                            <div className="flex items-center justify-between">
                              <div className="flex items-center gap-2">
                                {expandedMessage === index ? (
                                  <ChevronDown className="h-3 w-3" />
                                ) : (
                                  <ChevronRight className="h-3 w-3" />
                                )}
                                <span className="text-xs font-mono text-primary">
                                  {msg.type}
                                </span>
                                <span className="text-xs text-muted-foreground">
                                  {msg.timestamp.toLocaleTimeString()}
                                </span>
                              </div>
                              <span className="text-xs px-2 py-0.5 rounded bg-green-500/20 text-green-400">
                                ↓ received
                              </span>
                            </div>
                          </button>
                          
                          <AnimatePresence>
                            {expandedMessage === index && (
                              <motion.div
                                initial={{ height: 0 }}
                                animate={{ height: 'auto' }}
                                exit={{ height: 0 }}
                                transition={{ duration: 0.2 }}
                                className="overflow-hidden"
                              >
                                <pre className="text-[10px] p-3 bg-muted/50 overflow-x-auto">
                                  {JSON.stringify(msg.data, null, 2)}
                                </pre>
                              </motion.div>
                            )}
                          </AnimatePresence>
                        </motion.div>
                      ))}
                    </div>
                  )}
                </div>
              )}

              {activeTab === 'logs' && (
                <div className="p-4">
                  <div className="text-sm text-muted-foreground">
                    Frontend console logs will appear here...
                  </div>
                  <div className="mt-4 text-xs text-muted-foreground">
                    Coming soon: Real-time console.log() capture
                  </div>
                </div>
              )}

              {activeTab === 'events' && (
                <div className="p-4">
                  <div className="text-sm text-muted-foreground">
                    Event timeline will appear here...
                  </div>
                  <div className="mt-4 text-xs text-muted-foreground">
                    Coming soon: Visual event timeline
                  </div>
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  )
}
