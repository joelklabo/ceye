export function GitLabLogo({ className = "h-6 w-6" }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      className={className}
      fill="currentColor"
      aria-label="GitLab"
    >
      <path d="M23.546 10.93L13.067.452c-.604-.603-1.582-.603-2.188 0L.452 10.93c-.6.605-.6 1.584 0 2.188l10.427 10.426c.603.602 1.582.602 2.188 0l10.478-10.478c.6-.603.6-1.582 0-2.187M12 19.172l-7.172-7.172L12 4.828l7.172 7.172L12 19.172z" />
    </svg>
  )
}
