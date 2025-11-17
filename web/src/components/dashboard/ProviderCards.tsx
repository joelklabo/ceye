import { useState, useEffect, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Circle, RefreshCw, ChevronDown, ChevronRight, Zap } from 'lucide-react'
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

export function ProviderCards({ providers, meta, onRefresh }: ProviderCardsProps) {
  const providerEntries = Object.entries(providers)
  const [expandedPayload, setExpandedPayload] = useState<string | null>(null)
  const [flashingProvider, setFlashingProvider] = useState<string | null>(null)
  const lastWebhookTime = useRef<Record<string, string>>({})

  // Detect new webhooks and trigger flash animation
  useEffect(() => {
    providerEntries.forEach(([name, health]) => {
      if (health.LastWebhook?.received_at) {
        const previousTime = lastWebhookTime.current[name]
        const currentTime = health.LastWebhook.received_at
        
        // If webhook time changed, trigger flash
        if (previousTime && previousTime !== currentTime) {
          setFlashingProvider(name)
          setTimeout(() => setFlashingProvider(null), 1000)
        }
        
        lastWebhookTime.current[name] = currentTime
      }
    })
  }, [providerEntries])

  if (providerEntries.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-card">
        <div className="flex items-center justify-between border-b border-border p-4">
          <h2 className="text-lg font-semibold">Provider Health</h2>
        </div>
        <div className="p-4">
          <p className="text-muted-foreground text-sm">No providers configured</p>
        </div>
      </div>
    )
  }

  return (
    <div className="rounded-lg border border-border bg-card" data-testid="provider-health">
      <div className="flex items-center justify-between border-b border-border p-4">
        <h2 className="text-lg font-semibold">Provider Health</h2>
        {onRefresh && (
          <button
            onClick={onRefresh}
            className="text-sm text-muted-foreground hover:text-foreground transition-colors"
            data-testid="provider-refresh-button"
          >
            <RefreshCw className="h-4 w-4" />
          </button>
        )}
      </div>
      <div className="p-4 pt-0 mt-4">
        <div className="space-y-3" data-testid="provider-cards-list"> {/* Removed grid layout */}
          {providerEntries.map(([name, health], index) => {
            const isHealthy = health.ErrorCount === 0
            const lastUpdate = health.LastSuccess || health.LastError
            const logoPath = meta?.[name]?.logo

            return (
              <motion.div
                key={name}
                initial={{ opacity: 0, scale: 0.95 }}
                animate={
                  flashingProvider === name
                    ? {
                        opacity: 1,
                        scale: 1,
                        borderColor: [
                          'hsl(var(--border))',
                          'hsl(var(--primary))',
                          'hsl(var(--border))'
                        ]
                      }
                    : { opacity: 1, scale: 1 }
                }
                transition={
                  flashingProvider === name
                    ? { duration: 0.8, borderColor: { duration: 0.8 } }
                    : { duration: 0.2, delay: index * 0.05 }
                }
                whileHover={{ scale: 1.02 }}
                className="rounded-lg border border-border bg-card p-4 hover:shadow-lg transition-shadow"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <ProviderIcon provider={name} size="md" logoPath={logoPath} />
                    <div>
                      <h3 className="font-medium capitalize">{name}</h3>
                      <p className="text-xs text-muted-foreground mt-1">
                        Last update: {formatTime(lastUpdate)}
                      </p>
                      {health.MessageCount !== undefined && health.MessageCount > 0 && (
                        <p className="text-xs text-muted-foreground mt-0.5">
                          {health.MessageCount} messages received
                        </p>
                      )}
                      {health.LastWebhook && (
                        <p className="text-xs text-primary mt-0.5">
                          Last webhook: {health.LastWebhook.event_type}
                        </p>
                      )}
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

                {!isHealthy && health.ErrorCount > 0 && (
                  <div className="mt-2 rounded bg-red-500/10 p-2 text-xs text-red-400">
                    {health.ErrorCount} error{health.ErrorCount !== 1 ? 's' : ''} detected
                  </div>
                )}

                {isHealthy && (
                  <div className="mt-2 text-xs text-green-400">
                    ✓ Healthy
                  </div>
                )}

                {/* Webhook information */}
                {health.LastWebhook && (
                  <div className="mt-3 pt-3 border-t border-border">
                    <div className="flex items-center justify-between text-xs">
                      <div className="flex items-center gap-2">
                        <Zap className="h-3 w-3 text-primary" />
                        <span className="text-muted-foreground">Last webhook:</span>
                        <span className="text-foreground font-medium">
                          {health.LastWebhook.event_type}
                        </span>
                      </div>
                      <span className="text-muted-foreground">
                        {formatTime(health.LastWebhook.received_at)}
                      </span>
                    </div>

                    {health.MessageCount !== undefined && health.MessageCount > 0 && (
                      <div className="mt-2 text-xs text-muted-foreground">
                        {health.MessageCount} messages received
                      </div>
                    )}

                    {/* Payload viewer */}
                    {health.LastWebhook.payload && (
                      <div className="mt-2">
                        <button
                          onClick={() => setExpandedPayload(expandedPayload === name ? null : name)}
                          className="flex items-center gap-1 text-xs text-primary hover:underline"
                        >
                          {expandedPayload === name ? (
                            <>
                              <ChevronDown className="h-3 w-3" />
                              Hide Payload
                            </>
                          ) : (
                            <>
                              <ChevronRight className="h-3 w-3" />
                              View Payload
                            </>
                          )}
                        </button>

                        <AnimatePresence>
                          {expandedPayload === name && (
                            <motion.pre
                              initial={{ height: 0, opacity: 0 }}
                              animate={{ height: 'auto', opacity: 1 }}
                              exit={{ height: 0, opacity: 0 }}
                              transition={{ duration: 0.2 }}
                              className="mt-2 rounded bg-muted p-2 text-[10px] overflow-x-auto max-h-64 overflow-y-auto"
                            >
                              {JSON.stringify(JSON.parse(health.LastWebhook.payload), null, 2)}
                            </motion.pre>
                          )}
                        </AnimatePresence>
                      </div>
                    )}
                  </div>
                )}
              </motion.div>
            )
          })}
        </div>
      </div>
    </div>
  )
}