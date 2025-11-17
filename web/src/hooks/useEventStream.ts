import { useState, useEffect, useRef, useCallback } from 'react'

interface TimelineEvent {
  timestamp: string // ISO string
  type: string      // e.g., "websocket_message", "webhook_received", "provider_poll"
  category: string  // e.g., "WebSocket", "Webhook", "Polling", "UI", "Error"
  message: string   // Short description
  details?: any     // Full event payload
  id: string        // Unique ID for React keys
}

interface UseEventStreamOptions {
  enabled: boolean
  maxEntries?: number
}

export function useEventStream(options: UseEventStreamOptions) {
  const { enabled, maxEntries = 500 } = options
  const [events, setEvents] = useState<TimelineEvent[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  const connect = useCallback(() => {
    if (wsRef.current && (wsRef.current.readyState === WebSocket.OPEN || wsRef.current.readyState === WebSocket.CONNECTING)) {
      return
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/debug/events` // New WebSocket endpoint for events

    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => {
      console.log('EventStream: Connected to /debug/events')
      setIsConnected(true)
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
        reconnectTimeoutRef.current = null
      }
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        // Assuming the backend sends events in a structured format
        // For now, we'll just add it as a generic event
        const newEvent: TimelineEvent = {
          id: `${data.timestamp}-${Math.random()}`, // Simple unique ID
          timestamp: data.timestamp,
          type: data.type || 'unknown',
          category: data.category || 'General',
          message: data.message || JSON.stringify(data),
          details: data.details,
        }
        setEvents(prev => {
          const newEvents = [...prev, newEvent]
          if (newEvents.length > maxEntries) {
            return newEvents.slice(newEvents.length - maxEntries)
          }
          return newEvents
        })
      } catch (e) {
        console.error('EventStream: Failed to parse event message:', e)
      }
    }

    ws.onclose = () => {
      console.log('EventStream: Disconnected from /debug/events')
      setIsConnected(false)
      // Attempt to reconnect after a delay
      if (enabled) { // Only reconnect if the stream is still enabled
        reconnectTimeoutRef.current = setTimeout(() => {
          console.log('EventStream: Attempting to reconnect...')
          connect()
        }, 3000) // Reconnect after 3 seconds
      }
    }

    ws.onerror = (error) => {
      console.error('EventStream: WebSocket error:', error)
      ws.close()
    }
  }, [enabled, maxEntries])

  useEffect(() => {
    if (enabled) {
      connect()
    } else {
      wsRef.current?.close()
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
        reconnectTimeoutRef.current = null
      }
      setIsConnected(false)
    }

    return () => {
      wsRef.current?.close()
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
        reconnectTimeoutRef.current = null
      }
    }
  }, [enabled, connect])

  const clearEvents = useCallback(() => {
    setEvents([])
  }, [])

  return { events, isConnected, clearEvents }
}
