import { motion } from 'framer-motion'
import { useState, useEffect } from 'react'
import { Circle, Wifi, WifiOff } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ConnectionIndicatorProps {
  isConnected: boolean
  lastUpdate: Date | null
  onUpdate?: () => void
}

export function ConnectionIndicator({ isConnected, lastUpdate, onUpdate }: ConnectionIndicatorProps) {
  const [isPulsing, setIsPulsing] = useState(false)

  // Trigger pulse animation when lastUpdate changes
  useEffect(() => {
    if (lastUpdate && isConnected) {
      setIsPulsing(true)
      onUpdate?.()
      
      // Reset pulse after animation completes
      const timer = setTimeout(() => setIsPulsing(false), 1000)
      return () => clearTimeout(timer)
    }
  }, [lastUpdate, isConnected, onUpdate])

  return (
    <div className="flex items-center gap-3" data-testid="connection-indicator">
      {/* Connection Status with Pulse */}
      <div className="relative flex items-center gap-2">
        {/* Pulse Ring - Only when receiving messages */}
        {isPulsing && isConnected && (
          <motion.div
            initial={{ scale: 1, opacity: 0.8 }}
            animate={{ scale: 2.5, opacity: 0 }}
            transition={{ duration: 0.8, ease: 'easeOut' }}
            className="absolute left-0 h-2 w-2 rounded-full bg-green-400"
          />
        )}
        
        {/* Status Indicator */}
        <motion.div
          animate={isPulsing ? { scale: [1, 1.3, 1] } : {}}
          transition={{ duration: 0.3 }}
          className="relative"
        >
          <Circle
            className={cn(
              'h-2 w-2 transition-colors duration-200',
              isConnected ? 'fill-green-400 text-green-400' : 'fill-red-400 text-red-400'
            )}
          />
        </motion.div>
        
        {/* Status Text */}
        {lastUpdate && isConnected && (
          <motion.span
            key={lastUpdate.toISOString()}
            initial={{ opacity: 0.5 }}
            animate={{ opacity: 1 }}
            className="text-xs text-muted-foreground"
          >
            {formatTimeAgo(lastUpdate)}
          </motion.span>
        )}
      </div>

      {/* Live Badge with Glow */}
      {isConnected && (
        <motion.div
          animate={isPulsing ? {
            boxShadow: [
              '0 0 0px rgba(34, 197, 94, 0)',
              '0 0 20px rgba(34, 197, 94, 0.6)',
              '0 0 0px rgba(34, 197, 94, 0)'
            ]
          } : {}}
          transition={{ duration: 0.8 }}
          className="flex items-center gap-1.5 rounded-full bg-green-500/10 border border-green-500/20 px-2.5 py-1"
        >
          <Wifi className="h-3 w-3 text-green-400" />
          <span className="text-xs font-medium text-green-400">LIVE</span>
        </motion.div>
      )}
      
      {!isConnected && (
        <div className="flex items-center gap-1.5 rounded-full bg-red-500/10 border border-red-500/20 px-2.5 py-1">
          <WifiOff className="h-3 w-3 text-red-400" />
          <span className="text-xs font-medium text-red-400">OFFLINE</span>
        </div>
      )}
    </div>
  )
}

function formatTimeAgo(date: Date): string {
  const now = new Date()
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000)
  
  if (seconds < 5) return 'just now'
  if (seconds < 60) return `${seconds}s ago`
  
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  
  return date.toLocaleDateString()
}
