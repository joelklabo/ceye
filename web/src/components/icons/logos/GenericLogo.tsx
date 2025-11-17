import { Server } from 'lucide-react'

// Generate consistent color from provider name
function hashStringToHSL(str: string): string {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash)
    hash = hash & hash // Convert to 32bit integer
  }
  
  // Use hash to generate hue (0-360)
  // Keep saturation and lightness consistent for good visibility
  const hue = Math.abs(hash % 360)
  const saturation = 65 // Moderate saturation for pleasant colors
  const lightness = 55  // Mid lightness for good contrast
  
  return `hsl(${hue}, ${saturation}%, ${lightness}%)`
}

export function GenericLogo({ className = "h-6 w-6", provider }: { className?: string; provider?: string }) {
  // If provider name given, show first letter as monogram with color
  if (provider && provider.length > 0) {
    const letter = provider[0].toUpperCase()
    const bgColor = hashStringToHSL(provider.toLowerCase())
    
    return (
      <div 
        className={`${className} flex items-center justify-center rounded font-bold text-sm text-white`}
        style={{ backgroundColor: bgColor }}
      >
        {letter}
      </div>
    )
  }
  
  // Otherwise show generic server icon
  return <Server className={className} aria-label="Provider" />
}
