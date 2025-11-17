// Core types matching Go backend

export interface Run {
  ID: string
  Provider: string
  Repo: string
  WorkflowName: string
  Status: RunStatus
  Conclusion: string
  Branch: string
  CommitSHA: string
  StartedAt: string
  UpdatedAt: string
  Duration: number
  URL: string
}

export type RunStatus = 
  | "unknown"
  | "queued"
  | "in_progress"
  | "completed"
  | "failed"
  | "cancelled"

export interface ProviderHealth {
  LastError: string
  ErrorCount: number
  LastSuccess: string
}

export interface ProviderMeta {
  logo?: string
}

export interface DashboardMessage {
  type: string
  timestamp: string
  runs?: Run[]
  providers?: string[]
  status?: Record<string, string>
  health?: Record<string, ProviderHealth>
  meta?: Record<string, ProviderMeta>
  totals?: Record<string, number>
  alert_count?: number
  version?: string
  git_commit?: string
  build_time?: string
}

export interface StatsData {
  running: number
  queued: number
  success: number
  failed: number
}
