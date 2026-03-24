import React, { useState, useEffect, useCallback } from 'react'
import { motion } from 'framer-motion'
import {
  Play,
  Square,
  RotateCcw,
  Stethoscope,
  Cpu,
  MemoryStick,
  Clock,
  Hash,
  Globe,
  RefreshCw,
  CheckCircle,
  XCircle,
  AlertCircle,
} from 'lucide-react'
import { invoke, ServiceStatus } from '../../lib/tauri'
import { logger } from '../../lib/logger'

export default function Dashboard() {
  const [status, setStatus] = useState<ServiceStatus | null>(null)
  const [logs, setLogs] = useState<string[]>([])
  const [loading, setLoading] = useState(false)
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const fetchStatus = useCallback(async () => {
    try {
      const s = await invoke<ServiceStatus>('get_service_status')
      setStatus(s)
      setError(null)
    } catch (e) {
      logger.error('Failed to get status', e)
      setError(String(e))
    }
  }, [])

  const fetchLogs = useCallback(async () => {
    try {
      const lines = await invoke<string[]>('get_logs', { lines: 100 })
      setLogs(lines)
    } catch (e) {
      logger.warn('Failed to fetch logs', e)
    }
  }, [])

  useEffect(() => {
    setLoading(true)
    Promise.all([fetchStatus(), fetchLogs()]).finally(() => setLoading(false))
    const interval = setInterval(() => {
      fetchStatus()
      fetchLogs()
    }, 3000)
    return () => clearInterval(interval)
  }, [fetchStatus, fetchLogs])

  const handleAction = async (action: string) => {
    setActionLoading(action)
    try {
      await invoke(`${action}_service`)
      await fetchStatus()
      setError(null)
    } catch (e) {
      setError(String(e))
    } finally {
      setActionLoading(null)
    }
  }

  const handleDiagnose = async () => {
    setActionLoading('diagnose')
    try {
      const results = await invoke<{ name: string; passed: boolean; message: string }[]>('run_doctor')
      const summary = results.map(r => `${r.passed ? '✓' : '✗'} ${r.name}: ${r.message}`).join('\n')
      setLogs(prev => [...summary.split('\n'), '---', ...prev])
    } catch (e) {
      setError(String(e))
    } finally {
      setActionLoading(null)
    }
  }

  const formatUptime = (seconds: number | null) => {
    if (seconds === null) return '—'
    const h = Math.floor(seconds / 3600)
    const m = Math.floor((seconds % 3600) / 60)
    const s = seconds % 60
    if (h > 0) return `${h}h ${m}m`
    if (m > 0) return `${m}m ${s}s`
    return `${s}s`
  }

  return (
    <div className="space-y-6">
      {/* Status cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatusCard
          label="Status"
          value={loading ? 'Loading…' : status?.running ? 'Running' : 'Stopped'}
          icon={
            status?.running ? (
              <CheckCircle size={20} className="text-green-500" />
            ) : (
              <XCircle size={20} className="text-red-500" />
            )
          }
          valueClass={status?.running ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}
        />
        <StatusCard
          label="Port"
          value={status ? String(status.port) : '—'}
          icon={<Globe size={20} className="text-blue-500" />}
        />
        <StatusCard
          label="PID"
          value={status?.pid ? String(status.pid) : '—'}
          icon={<Hash size={20} className="text-purple-500" />}
        />
        <StatusCard
          label="Memory"
          value={status?.memory_mb ? `${status.memory_mb.toFixed(1)} MB` : '—'}
          icon={<MemoryStick size={20} className="text-orange-500" />}
        />
        <StatusCard
          label="CPU"
          value={status?.cpu_percent !== null && status?.cpu_percent !== undefined ? `${status.cpu_percent.toFixed(1)}%` : '—'}
          icon={<Cpu size={20} className="text-cyan-500" />}
        />
        <StatusCard
          label="Uptime"
          value={formatUptime(status?.uptime_seconds ?? null)}
          icon={<Clock size={20} className="text-teal-500" />}
        />
      </div>

      {/* Controls */}
      <div className="card p-5">
        <h2 className="section-title mb-4">Service Controls</h2>
        {error && (
          <div className="mb-4 flex items-center gap-2 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-700 dark:text-red-400 text-sm">
            <AlertCircle size={16} />
            {error}
          </div>
        )}
        <div className="flex flex-wrap gap-3">
          <button
            className="btn-primary"
            disabled={!!actionLoading || status?.running}
            onClick={() => handleAction('start')}
          >
            {actionLoading === 'start' ? <RefreshCw size={16} className="animate-spin" /> : <Play size={16} />}
            Start
          </button>
          <button
            className="btn-danger"
            disabled={!!actionLoading || !status?.running}
            onClick={() => handleAction('stop')}
          >
            {actionLoading === 'stop' ? <RefreshCw size={16} className="animate-spin" /> : <Square size={16} />}
            Stop
          </button>
          <button
            className="btn-warning"
            disabled={!!actionLoading}
            onClick={() => handleAction('restart')}
          >
            {actionLoading === 'restart' ? <RefreshCw size={16} className="animate-spin" /> : <RotateCcw size={16} />}
            Restart
          </button>
          <button
            className="btn-secondary"
            disabled={!!actionLoading}
            onClick={handleDiagnose}
          >
            {actionLoading === 'diagnose' ? (
              <RefreshCw size={16} className="animate-spin" />
            ) : (
              <Stethoscope size={16} />
            )}
            Diagnose
          </button>
          <button
            className="btn-secondary ml-auto"
            onClick={() => { fetchStatus(); fetchLogs() }}
            disabled={!!actionLoading}
          >
            <RefreshCw size={16} />
            Refresh
          </button>
        </div>
      </div>

      {/* Logs */}
      <div className="card p-5">
        <div className="flex items-center justify-between mb-3">
          <h2 className="section-title">Real-time Logs</h2>
          <button
            className="text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
            onClick={() => setLogs([])}
          >
            Clear
          </button>
        </div>
        <div className="bg-gray-950 dark:bg-gray-900 rounded-lg p-4 h-64 overflow-y-auto log-output">
          {logs.length === 0 ? (
            <p className="text-gray-500 italic text-sm">No log output</p>
          ) : (
            logs.map((line, i) => (
              <div
                key={i}
                className={`text-xs leading-relaxed ${
                  line.includes('ERROR') || line.includes('✗')
                    ? 'text-red-400'
                    : line.includes('WARN')
                    ? 'text-yellow-400'
                    : line.includes('✓')
                    ? 'text-green-400'
                    : 'text-gray-300'
                }`}
              >
                {line || '\u00A0'}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}

function StatusCard({
  label,
  value,
  icon,
  valueClass = 'text-gray-900 dark:text-gray-100',
}: {
  label: string
  value: string
  icon: React.ReactNode
  valueClass?: string
}) {
  return (
    <motion.div
      className="card p-4 flex items-center gap-3"
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
    >
      <div className="p-2 rounded-lg bg-gray-50 dark:bg-gray-700/50">{icon}</div>
      <div>
        <p className="text-xs text-gray-500 dark:text-gray-400">{label}</p>
        <p className={`text-base font-semibold ${valueClass}`}>{value}</p>
      </div>
    </motion.div>
  )
}
