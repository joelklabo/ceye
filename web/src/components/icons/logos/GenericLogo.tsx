import { Server } from 'lucide-react'

export function GenericLogo({ className = "h-6 w-6", provider }: { className?: string; provider?: string }) {
  // If provider name given, show first letter as monogram
  if (provider && provider.length > 0) {
    const letter = provider[0].toUpperCase()
    return (
      <div className={`${className} flex items-center justify-center rounded bg-muted text-muted-foreground font-bold text-sm`}>
        {letter}
      </div>
    )
  }
  
  // Otherwise show generic server icon
  return <Server className={className} aria-label="Provider" />
}
