// Tauri IPC wrapper - gracefully falls back when running in browser for development
import { invoke as tauriInvoke } from '@tauri-apps/api/core'

export async function invoke<T>(command: string, args?: Record<string, unknown>): Promise<T> {
  try {
    return await tauriInvoke<T>(command, args)
  } catch (error) {
    console.error(`[tauri] command '${command}' failed:`, error)
    throw error
  }
}

// Service status types
export interface ServiceStatus {
  running: boolean
  pid: number | null
  port: number
  uptime_seconds: number | null
  memory_mb: number | null
  cpu_percent: number | null
}

// Diagnostic result
export interface DiagnosticResult {
  name: string
  passed: boolean
  message: string
  suggestion: string | null
}

// Security issue
export interface SecurityIssue {
  id: string
  title: string
  description: string
  severity: 'high' | 'medium' | 'low'
  fixable: boolean
  fixed: boolean
  category: string
  detail: string | null
}

// AI Provider
export interface AIProvider {
  id: string
  name: string
  base_url: string
  api_key: string
  models: string[]
  is_official: boolean
}

// Agent
export interface Agent {
  id: string
  name: string
  emoji: string
  description: string
  model_override: string
  channels: string[]
  tools: string[]
  sandbox_mode: boolean
  workspace_isolation: boolean
  mention_mode: boolean
  sub_agent_permissions: string[]
  is_primary: boolean
}

// Skill
export interface Skill {
  name: string
  version: string
  description: string
  category: 'builtin' | 'official' | 'community'
  installed: boolean
  enabled: boolean
  config_schema: SkillConfigField[]
  config_values: Record<string, unknown>
}

export interface SkillConfigField {
  key: string
  label: string
  field_type: 'text' | 'password' | 'select' | 'toggle'
  default_value: string
  options: string[]
  required: boolean
}

// Channel config
export interface ChannelConfig {
  enabled: boolean
  [key: string]: unknown
}

// Environment info
export interface EnvInfo {
  node_version: string | null
  npm_version: string | null
  openclaw_version: string | null
  os: string
  arch: string
  platform: string
}
