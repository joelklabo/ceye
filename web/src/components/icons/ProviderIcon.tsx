import { GitHubLogo } from './logos/GitHubLogo'
import { AzureLogo } from './logos/AzureLogo'
import { GitLabLogo } from './logos/GitLabLogo'
import { AWSLogo } from './logos/AWSLogo'
import { GenericLogo } from './logos/GenericLogo'
import { useState } from 'react'

interface ProviderIconProps {
  provider: string
  logoPath?: string
  size?: 'xs' | 'sm' | 'md' | 'lg'
  className?: string
  fallback?: 'monogram' | 'icon'
}

const sizeClasses = {
  xs: 'h-4 w-4',
  sm: 'h-5 w-5',
  md: 'h-6 w-6',
  lg: 'h-8 w-8',
}

export function ProviderIcon({
  provider,
  logoPath,
  size = 'md',
  className,
  fallback = 'monogram'
}: ProviderIconProps) {
  const [imageError, setImageError] = useState(false)
  const sizeClass = className || sizeClasses[size]
  
  // If custom logo path provided and not errored, try to load it
  if (logoPath && !imageError) {
    return (
      <img
        src={logoPath}
        alt={`${provider} logo`}
        className={sizeClass}
        data-provider-logo
        onError={() => setImageError(true)}
      />
    )
  }
  
  // Check for built-in logos (case-insensitive)
  const providerLower = provider.toLowerCase()
  
  if (providerLower.includes('github')) {
    return <GitHubLogo className={sizeClass} />
  }
  
  if (providerLower.includes('azure')) {
    return <AzureLogo className={sizeClass} />
  }
  
  if (providerLower.includes('gitlab')) {
    return <GitLabLogo className={sizeClass} />
  }
  
  if (providerLower.includes('aws') || providerLower.includes('codebuild')) {
    return <AWSLogo className={sizeClass} />
  }
  
  // Fallback to generic logo
  if (fallback === 'monogram') {
    return <GenericLogo className={sizeClass} provider={provider} />
  }
  
  return <GenericLogo className={sizeClass} />
}
