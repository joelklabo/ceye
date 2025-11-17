import { motion } from 'framer-motion'
import { Circle, ChevronDown, ChevronUp } from 'lucide-react'
import { useState } from 'react'
import type { ProviderHealth, ProviderMeta } from '@/types'
import { cn } from '@/lib/utils'
import { ProviderIcon } from '@/components/icons/ProviderIcon'

interface ProviderCardsProps {
  providers: Record<string, ProviderHealth>
  meta?: Record<string, ProviderMeta>
  onRefresh?: () => void // Added onRefresh prop
}

function formatTime(timestamp: string): string {
  if (!timestamp) return 'Never'
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

function ProviderCard({ name, health, logoPath }: {
  name: string
  health: ProviderHealth
  logoPath?: string
}) {
  const [showPayload, setShowPayload] = useState(false)
  const isHealthy = health.ErrorCount === 0
  const lastUpdate = health.LastSuccess || health.LastError

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.2 }}
      whileHover={{ scale: 1.01 }}
      className="rounded-lg border border-border bg-card p-4 hover:shadow-md transition-all"
    >
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3">
          <ProviderIcon provider={name} size="md" logoPath={logoPath} />
          <div>
            <h3 className="font-medium capitalize">{name}</h3>
            <p className="text-xs text-muted-foreground mt-1">
              Last update: {formatTime(lastUpdate)}
            </p>
          </div>
        </div>
        <motion.div
          animate={{
            scale: [1, 1.2, 1],
            opacity: [1, 0.8, 1],
          }}
          transition={{
            duration: 2,
            repeat: Infinity,
            ease: "easeInOut",
          }}
        >
          <Circle
            className={cn(
              'h-3 w-3',
              isHealthy ? 'fill-green-400 text-green-400' : 'fill-red-400 text-red-400'
            )}
          />
        </motion.div>
      </div>

      {/* Webhook information */}
      {health.LastWebhook && (
        <div className="mt-2 space-y-1">
          <div className="text-xs text-muted-foreground">
            Last webhook: <span className="font-mono">{health.LastWebhook.event_type}</span>
          </div>
          {health.LastWebhook.payload && (
            <button
              onClick={() => setShowPayload(!showPayload)}
              className="text-xs text-primary hover:underline flex items-center gap-1"
            >
              {showPayload ? (
                <>
                  <ChevronUp className="h-3 w-3" />
                  Hide Payload
                </>
              ) : (
                <>
                  <ChevronDown className="h-3 w-3" />
                  View Payload
                </>
              )}
            </button>
          )}
        </div>
      )}

      {/* Payload viewer */}
      {showPayload && health.LastWebhook?.payload && (
        <motion.pre
          initial={{ height: 0, opacity: 0 }}
          animate={{ height: 'auto', opacity: 1 }}
          exit={{ height: 0, opacity: 0 }}
          transition={{ duration: 0.2 }}
          className="mt-2 text-[10px] bg-muted p-2 rounded overflow-x-auto max-h-96 overflow-y-auto"
        >
          {JSON.stringify(JSON.parse(health.LastWebhook.payload), null, 2)}
        </motion.pre>
      )}

      {/* Message count */}
      {health.MessageCount !== undefined && health.MessageCount > 0 && (
        <div className="mt-1 text-xs text-muted-foreground">
          📨 {health.MessageCount} message{health.MessageCount !== 1 ? 's' : ''} received
        </div>
      )}

      {/* Error state */}
      {!isHealthy && health.ErrorCount > 0 && (
        <div className="mt-2 rounded bg-red-500/10 p-2 text-xs text-red-400">
          {health.ErrorCount} error{health.ErrorCount !== 1 ? 's' : ''} detected
        </div>
      )}

      {/* Healthy state */}
      {isHealthy && !health.LastWebhook && (
        <div className="mt-2 text-xs text-green-400">
          ✓ Healthy
        </div>
      )}
    </motion.div>
  )
}

export function ProviderCards({ providers, meta }: ProviderCardsProps) {
  const providerEntries = Object.entries(providers)

  if (providerEntries.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="text-lg font-semibold mb-4">Provider Health</h2>
        <p className="text-muted-foreground text-sm">No providers configured</p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border border-border bg-card">
      <div className="flex items-center justify-between"> {/* Added this div to contain title and debug button */}
        <h2 className="text-lg font-semibold">Provider Health</h2>
        <div data-testid="refresh-button-debug">Refresh Debug</div> {/* Always render for debugging */}
      </div>
      <div className="p-4 space-y-3">
        {providerEntries.map(([name, health]) => (
          <ProviderCard
            key={name}
            name={name}
            health={health}
            logoPath={meta?.[name]?.logo}
          />
        ))}
      </div>
    </div>
  )
}
