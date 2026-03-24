import React, { useState, useEffect } from 'react'
import {
  Shield,
  AlertTriangle,
  CheckCircle,
  RefreshCw,
  Wrench,
  ChevronDown,
  ChevronUp,
} from 'lucide-react'
import { invoke, SecurityIssue } from '../../lib/tauri'

export default function Security() {
  const [issues, setIssues] = useState<SecurityIssue[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(false)
  const [fixing, setFixing] = useState(false)
  const [expanded, setExpanded] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [lastScanned, setLastScanned] = useState<Date | null>(null)

  const runScan = async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await invoke<SecurityIssue[]>('run_security_scan')
      setIssues(result || [])
      setSelected(new Set())
      setLastScanned(new Date())
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }

  const fixSelected = async () => {
    if (selected.size === 0) return
    setFixing(true)
    try {
      await invoke('fix_security_issues', { issueIds: Array.from(selected) })
      await runScan()
    } catch (e) {
      setError(String(e))
    } finally {
      setFixing(false)
    }
  }

  const toggleSelect = (id: string) => {
    setSelected(prev => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const selectAllFixable = () => {
    const fixableIds = issues.filter(i => i.fixable && !i.fixed).map(i => i.id)
    setSelected(new Set(fixableIds))
  }

  const severityIcon = (severity: string) => {
    switch (severity) {
      case 'high':
        return <AlertTriangle size={16} className="text-red-500" />
      case 'medium':
        return <AlertTriangle size={16} className="text-yellow-500" />
      case 'low':
        return <AlertTriangle size={16} className="text-blue-500" />
      default:
        return <AlertTriangle size={16} className="text-gray-500" />
    }
  }

  const highCount = issues.filter(i => i.severity === 'high' && !i.fixed).length
  const medCount = issues.filter(i => i.severity === 'medium' && !i.fixed).length
  const lowCount = issues.filter(i => i.severity === 'low' && !i.fixed).length
  const fixedCount = issues.filter(i => i.fixed).length

  return (
    <div className="space-y-5">
      {error && (
        <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-700 dark:text-red-400 text-sm">
          {error}
        </div>
      )}

      {/* Summary cards */}
      {issues.length > 0 && (
        <div className="grid grid-cols-4 gap-4">
          <div className="card p-4 text-center">
            <div className="text-3xl font-bold text-red-600 dark:text-red-400">{highCount}</div>
            <div className="text-sm text-gray-500 mt-1">High Risk</div>
          </div>
          <div className="card p-4 text-center">
            <div className="text-3xl font-bold text-yellow-600 dark:text-yellow-400">{medCount}</div>
            <div className="text-sm text-gray-500 mt-1">Medium Risk</div>
          </div>
          <div className="card p-4 text-center">
            <div className="text-3xl font-bold text-blue-600 dark:text-blue-400">{lowCount}</div>
            <div className="text-sm text-gray-500 mt-1">Low Risk</div>
          </div>
          <div className="card p-4 text-center">
            <div className="text-3xl font-bold text-green-600 dark:text-green-400">{fixedCount}</div>
            <div className="text-sm text-gray-500 mt-1">Fixed</div>
          </div>
        </div>
      )}

      {/* Controls */}
      <div className="card p-5">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="section-title">Security Scan</h2>
            {lastScanned && (
              <p className="section-subtitle">Last scan: {lastScanned.toLocaleTimeString()}</p>
            )}
          </div>
          <div className="flex gap-3">
            {issues.some(i => i.fixable && !i.fixed) && (
              <button className="btn-secondary" onClick={selectAllFixable}>
                Select All Fixable
              </button>
            )}
            {selected.size > 0 && (
              <button className="btn-success" onClick={fixSelected} disabled={fixing}>
                {fixing ? (
                  <RefreshCw size={16} className="animate-spin" />
                ) : (
                  <Wrench size={16} />
                )}
                Fix Selected ({selected.size})
              </button>
            )}
            <button className="btn-primary" onClick={runScan} disabled={loading}>
              {loading ? <RefreshCw size={16} className="animate-spin" /> : <Shield size={16} />}
              Run Scan
            </button>
          </div>
        </div>

        {issues.length === 0 && !loading && (
          <div className="flex flex-col items-center py-10 text-center">
            {lastScanned ? (
              <>
                <CheckCircle size={40} className="text-green-500 mb-3" />
                <p className="text-green-700 dark:text-green-400 font-medium">No security issues found!</p>
                <p className="text-sm text-gray-500 mt-1">Your OpenClaw installation looks secure.</p>
              </>
            ) : (
              <>
                <Shield size={40} className="text-gray-300 dark:text-gray-600 mb-3" />
                <p className="text-gray-500 dark:text-gray-400">Click "Run Scan" to check for security issues.</p>
              </>
            )}
          </div>
        )}

        {loading && (
          <div className="flex items-center justify-center py-10">
            <RefreshCw size={24} className="animate-spin text-blue-500 mr-3" />
            <span className="text-gray-500">Scanning for security issues...</span>
          </div>
        )}
      </div>

      {/* Issues list */}
      {issues.length > 0 && (
        <div className="space-y-2">
          {issues.map(issue => (
            <div
              key={issue.id}
              className={`card overflow-hidden ${issue.fixed ? 'opacity-60' : ''}`}
            >
              <div className="flex items-center gap-4 px-5 py-4">
                {issue.fixable && !issue.fixed && (
                  <input
                    type="checkbox"
                    checked={selected.has(issue.id)}
                    onChange={() => toggleSelect(issue.id)}
                    className="w-4 h-4 rounded text-blue-500 flex-shrink-0"
                  />
                )}
                {issue.fixed && (
                  <CheckCircle size={18} className="text-green-500 flex-shrink-0" />
                )}
                {!issue.fixable && !issue.fixed && (
                  <div className="w-4 h-4 flex-shrink-0" />
                )}

                <div className="flex-shrink-0">{severityIcon(issue.severity)}</div>

                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="font-medium text-gray-900 dark:text-gray-100">{issue.title}</span>
                    <span className={`badge-${issue.severity}`}>{issue.severity}</span>
                    <span className="text-xs text-gray-400">{issue.category}</span>
                    {issue.fixed && (
                      <span className="text-xs px-2 py-0.5 rounded-full bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400">
                        Fixed
                      </span>
                    )}
                  </div>
                  <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5 truncate">
                    {issue.description}
                  </p>
                </div>

                <button
                  onClick={() => setExpanded(expanded === issue.id ? null : issue.id)}
                  className="p-1.5 text-gray-400 hover:text-gray-600 transition-colors"
                >
                  {expanded === issue.id ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
                </button>
              </div>

              {expanded === issue.id && (
                <div className="px-5 pb-4 border-t border-gray-100 dark:border-gray-700 pt-3">
                  {issue.detail && (
                    <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-3 text-sm text-gray-700 dark:text-gray-300 font-mono mb-3">
                      {issue.detail}
                    </div>
                  )}
                  {issue.fixable && !issue.fixed && (
                    <button
                      className="btn-success text-sm"
                      onClick={async () => {
                        setFixing(true)
                        try {
                          await invoke('fix_security_issues', { issueIds: [issue.id] })
                          await runScan()
                        } catch (e) {
                          setError(String(e))
                        } finally {
                          setFixing(false)
                        }
                      }}
                    >
                      <Wrench size={14} />
                      Fix This Issue
                    </button>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
