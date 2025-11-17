export function AzureLogo({ className = "h-6 w-6" }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      className={className}
      fill="currentColor"
      aria-label="Azure DevOps"
    >
      <path d="M0 12.357L5.333 24l8.799-2.057V2.057L5.333 0 0 12.357zm18.667 1.035l-4.267-.802v5.753l-5.867 1.649L18.667 24h4.8l-4.8-10.608zm0-2.727L19.6 0H6.4l8.267 8.206 3.467-1.237v4.696h.533z" />
    </svg>
  )
}
