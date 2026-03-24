import React, { useState, useEffect, useRef, useCallback } from 'react'
import { RefreshCw, Trash2, Download, Search } from 'lucide-react'
import { invoke } from '../../lib/tauri'

export default function Logs() {
  const [lines, setLines] = useState<string[]>([])
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [lineCount, setLineCount] = useState(200)
  const [filter, setFilter] = useState('')
  const [loading, setLoading] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)
  const [autoScroll, setAutoScroll] = useState(true)

  const fetchLogs = useCallback(async () => {
    try {
      const result = await invoke<string[]>('get_logs', { lines: lineCount })
      setLines(result || [])
    } catch {
      // service may not be running
    }
  }, [lineCount])

  useEffect(() => {
    setLoading(true)
    fetchLogs().finally(() => setLoading(false))
  }, [fetchLogs])

  useEffect(() => {
    if (!autoRefresh) return
    const interval = setInterval(fetchLogs, 2000)
    return () => clearInterval(interval)
  }, [autoRefresh, fetchLogs])

  useEffect(() => {
    if (autoScroll && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [lines, autoScroll])

  const filteredLines = filter
    ? lines.filter(l => l.toLowerCase().includes(filter.toLowerCase()))
    : lines

  const downloadLogs = () => {
    const content = lines.join('\n')
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `openclaw-${new Date().toISOString().slice(0, 19)}.log`
    a.click()
    URL.revokeObjectURL(url)
  }

  const getLineClass = (line: string) => {
    const lower = line.toLowerCase()
    if (lower.includes('error') || lower.includes('err ') || lower.includes('[error]')) return 'text-red-400'
    if (lower.includes('warn') || lower.includes('[warn]')) return 'text-yellow-400'
    if (lower.includes('info') || lower.includes('[info]')) return 'text-blue-400'
    if (lower.includes('debug') || lower.includes('[debug]')) return 'text-gray-500'
    return 'text-gray-300'
  }

  return (
    <div className="flex flex-col h-full space-y-3" style={{ minHeight: 0 }}>
      {/* Toolbar */}
      <div className="flex items-center gap-3 flex-shrink-0">
        <div className="relative flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            className="input pl-9 text-sm"
            value={filter}
            onChange={e => setFilter(e.target.value)}
            placeholder="Filter logs..."
          />
        </div>
        <select
          className="input w-32"
          value={lineCount}
          onChange={e => setLineCount(Number(e.target.value))}
        >
          <option value={50}>50 lines</option>
          <option value={100}>100 lines</option>
          <option value={200}>200 lines</option>
          <option value={500}>500 lines</option>
          <option value={1000}>1000 lines</option>
        </select>
        <label className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 cursor-pointer">
          <input
            type="checkbox"
            checked={autoRefresh}
            onChange={e => setAutoRefresh(e.target.checked)}
            className="w-4 h-4 rounded text-blue-500"
          />
          Auto-refresh
        </label>
        <label className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 cursor-pointer">
          <input
            type="checkbox"
            checked={autoScroll}
            onChange={e => setAutoScroll(e.target.checked)}
            className="w-4 h-4 rounded text-blue-500"
          />
          Auto-scroll
        </label>
        <button className="btn-secondary" onClick={fetchLogs} disabled={loading}>
          <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
          Refresh
        </button>
        <button className="btn-secondary" onClick={downloadLogs}>
          <Download size={14} />
          Export
        </button>
        <button className="btn-secondary" onClick={() => setLines([])}>
          <Trash2 size={14} />
          Clear
        </button>
      </div>

      {/* Log output */}
      <div className="flex-1 bg-gray-950 dark:bg-black rounded-xl overflow-auto log-output border border-gray-200 dark:border-gray-800">
        {filteredLines.length === 0 ? (
          <div className="flex items-center justify-center h-full text-gray-600 text-sm italic">
            {loading ? 'Loading logs...' : 'No log output'}
          </div>
        ) : (
          <div className="p-4 space-y-0">
            {filteredLines.map((line, i) => (
              <div key={i} className={`text-xs leading-relaxed ${getLineClass(line)}`}>
                <span className="text-gray-700 mr-3 select-none">{String(i + 1).padStart(4, ' ')}</span>
                {line || '\u00A0'}
              </div>
            ))}
            <div ref={bottomRef} />
          </div>
        )}
      </div>

      <div className="flex items-center justify-between text-xs text-gray-400 flex-shrink-0">
        <span>
          {filter ? `${filteredLines.length} of ${lines.length}` : lines.length} lines
        </span>
        {autoRefresh && (
          <span className="flex items-center gap-1">
            <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
            Live
          </span>
        )}
      </div>
    </div>
  )
}
