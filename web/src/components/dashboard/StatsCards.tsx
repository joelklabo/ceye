import { motion } from 'framer-motion'
import { Activity, Clock, CheckCircle2, XCircle } from 'lucide-react'
import type { StatsData } from '@/types'

interface StatsCardsProps {
  stats: StatsData
}

interface StatCardProps {
  title: string
  value: number
  icon: React.ReactNode
  gradient: string
  delay: number
}

function StatCard({ title, value, icon, gradient, delay }: StatCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, delay }}
      className={`relative overflow-hidden rounded-lg border border-border bg-gradient-to-br ${gradient} p-6`}
    >
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm font-medium text-muted-foreground">{title}</p>
          <motion.p
            key={value}
            initial={{ scale: 1.2, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ duration: 0.2 }}
            className="mt-2 text-4xl font-bold"
          >
            {value}
          </motion.p>
        </div>
        <div className="rounded-full bg-background/20 p-3">
          {icon}
        </div>
      </div>
      
      {/* Pulse effect on value changes */}
      {value > 0 && (
        <motion.div
          key={`pulse-${value}`}
          initial={{ scale: 0.8, opacity: 0.5 }}
          animate={{ scale: 1.5, opacity: 0 }}
          transition={{ duration: 0.6 }}
          className="absolute inset-0 rounded-lg border-2 border-current"
        />
      )}
    </motion.div>
  )
}

export function StatsCards({ stats }: StatsCardsProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <StatCard
        title="Running"
        value={stats.running}
        icon={<Activity className="h-6 w-6 text-purple-400" />}
        gradient="from-purple-500/10 to-purple-500/5"
        delay={0}
      />
      <StatCard
        title="Queued"
        value={stats.queued}
        icon={<Clock className="h-6 w-6 text-amber-400" />}
        gradient="from-amber-500/10 to-amber-500/5"
        delay={0.1}
      />
      <StatCard
        title="Success"
        value={stats.success}
        icon={<CheckCircle2 className="h-6 w-6 text-green-400" />}
        gradient="from-green-500/10 to-green-500/5"
        delay={0.2}
      />
      <StatCard
        title="Failed"
        value={stats.failed}
        icon={<XCircle className="h-6 w-6 text-red-400" />}
        gradient="from-red-500/10 to-red-500/5"
        delay={0.3}
      />
    </div>
  )
}
