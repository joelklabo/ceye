import { motion } from 'framer-motion'
import { Circle } from 'lucide-react'
import type { ProviderHealth } from '@/types'
import { cn } from '@/lib/utils'
import { ProviderIcon } from '@/components/icons/ProviderIcon'

interface ProviderCardsProps {
  providers: Record<string, ProviderHealth>
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

export function ProviderCards({ providers }: ProviderCardsProps) {
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
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">Provider Health</h2>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {providerEntries.map(([name, health], index) => {
          const isHealthy = health.ErrorCount === 0
          const lastUpdate = health.LastSuccess || health.LastError

          return (
            <motion.div
              key={name}
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.2, delay: index * 0.05 }}
              whileHover={{ scale: 1.02 }}
              className="rounded-lg border border-border bg-card p-4 hover:shadow-lg transition-shadow"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3">
                  <ProviderIcon provider={name} size="md" />
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
            </motion.div>
          )
        })}
      </div>
    </div>
  )
}
