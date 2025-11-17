import { motion } from 'framer-motion'

interface SkeletonProps {
  className?: string
  variant?: 'text' | 'circular' | 'rectangular'
  width?: string | number
  height?: string | number
  animation?: 'pulse' | 'wave' | 'none'
}

export function Skeleton({
  className = '',
  variant = 'rectangular',
  width,
  height,
  animation = 'pulse',
}: SkeletonProps) {
  const variantClasses = {
    text: 'rounded',
    circular: 'rounded-full',
    rectangular: 'rounded-lg',
  }

  const getAnimation = () => {
    if (animation === 'pulse') {
      return {
        opacity: [0.4, 0.8, 0.4],
        transition: {
          duration: 1.5,
          repeat: Infinity as number,
          ease: 'easeInOut' as const,
        },
      }
    }
    if (animation === 'wave') {
      return {
        x: ['-100%', '100%'],
        transition: {
          duration: 1.5,
          repeat: Infinity as number,
          ease: 'linear' as const,
        },
      }
    }
    return {}
  }

  const style = {
    width: width ? (typeof width === 'number' ? `${width}px` : width) : '100%',
    height: height ? (typeof height === 'number' ? `${height}px` : height) : '20px',
  }

  return (
    <motion.div
      className={`bg-muted ${variantClasses[variant]} ${className}`}
      style={style}
      animate={getAnimation()}
    />
  )
}

// Common skeleton patterns
export function SkeletonCard() {
  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <Skeleton width="40%" height={16} className="mb-3" />
          <Skeleton width="60%" height={36} />
        </div>
        <Skeleton variant="circular" width={48} height={48} />
      </div>
    </div>
  )
}

export function SkeletonTable({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-3">
      {/* Header */}
      <div className="flex gap-4">
        {[100, 150, 120, 100, 80].map((width, i) => (
          <Skeleton key={i} width={width} height={16} />
        ))}
      </div>
      {/* Rows */}
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex gap-4">
          {[100, 150, 120, 100, 80].map((width, j) => (
            <Skeleton key={j} width={width} height={20} />
          ))}
        </div>
      ))}
    </div>
  )
}

export function SkeletonText({ lines = 3 }: { lines?: number }) {
  return (
    <div className="space-y-2">
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton
          key={i}
          width={i === lines - 1 ? '80%' : '100%'}
          height={16}
        />
      ))}
    </div>
  )
}
