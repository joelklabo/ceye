import { useState, useEffect, useCallback, useRef } from 'react'

export interface LogEntry {
  type: string
  timestamp: string
  level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR'
  component: string
  message: string
}

interface UseLogStreamOptions {
  enabled?: boolean
  maxEntries?: number
}

export function useLogStream(options: UseLogStreamOptions = {}) {
  const { enabled = true, maxEntries = 500 } = options
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)

  const clearLogs = useCallback(() => {
    setLogs([])
  }, [])

  useEffect(() => {
    if (!enabled) return

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${wsProtocol}//${window.location.host}/debug/logs`

    const connect = () => {
      try {
        const ws = new WebSocket(wsUrl)
        wsRef.current = ws

        ws.onopen = () => {
          setIsConnected(true)
        }

        ws.onmessage = (event) => {
          try {
            const logEntry: LogEntry = JSON.parse(event.data)
            setLogs((prev) => [...prev, logEntry].slice(-maxEntries))
          } catch (e) {
            console.error('Failed to parse log message:', e)
          }
        }

        ws.onerror = (error) => {
          console.error('Log WebSocket error:', error)
          setIsConnected(false)
        }

        ws.onclose = () => {
          setIsConnected(false)
          // Auto-reconnect after 2 seconds
          setTimeout(() => {
            if (enabled) connect()
          }, 2000)
        }
      } catch (error) {
        console.error('Failed to create WebSocket:', error)
      }
    }

    connect()

    return () => {
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [enabled, maxEntries])

  return {
    logs,
    isConnected,
    clearLogs,
  }
}
