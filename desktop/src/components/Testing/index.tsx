import React, { useState } from 'react'
import {
  FlaskConical,
  CheckCircle,
  XCircle,
  RefreshCw,
  Server,
  Brain,
  MessageSquare,
  Monitor,
} from 'lucide-react'
import { invoke, DiagnosticResult, EnvInfo } from '../../lib/tauri'

export default function Testing() {
  const [doctorResults, setDoctorResults] = useState<DiagnosticResult[]>([])
  const [doctorRunning, setDoctorRunning] = useState(false)
  const [envInfo, setEnvInfo] = useState<EnvInfo | null>(null)

  // AI test
  const [aiProvider, setAiProvider] = useState('')
  const [aiModel, setAiModel] = useState('')
  const [aiApiKey, setAiApiKey] = useState('')
  const [aiBaseUrl, setAiBaseUrl] = useState('')
  const [aiTestResult, setAiTestResult] = useState<{ success: boolean; message: string } | null>(null)
  const [aiTesting, setAiTesting] = useState(false)

  // Channel test
  const [chChannel, setChChannel] = useState('telegram')
  const [chTestResult, setChTestResult] = useState<{ success: boolean; message: string } | null>(null)
  const [chTesting, setChTesting] = useState(false)

  const runDoctor = async () => {
    setDoctorRunning(true)
    try {
      const [results, info] = await Promise.all([
        invoke<DiagnosticResult[]>('run_doctor'),
        invoke<EnvInfo>('get_system_info'),
      ])
      setDoctorResults(results || [])
      setEnvInfo(info)
    } catch (e) {
      setDoctorResults([{ name: 'Error', passed: false, message: String(e), suggestion: null }])
    } finally {
      setDoctorRunning(false)
    }
  }

  const testAI = async () => {
    if (!aiProvider || !aiModel) return alert('Provider and model are required')
    setAiTesting(true)
    setAiTestResult(null)
    try {
      const result = await invoke<{ success: boolean; message: string }>('test_ai_connection', {
        provider: aiProvider,
        model: aiModel,
        apiKey: aiApiKey,
        baseUrl: aiBaseUrl,
      })
      setAiTestResult(result)
    } catch (e) {
      setAiTestResult({ success: false, message: String(e) })
    } finally {
      setAiTesting(false)
    }
  }

  const testChannel = async () => {
    setChTesting(true)
    setChTestResult(null)
    try {
      const result = await invoke<{ success: boolean; message: string }>('test_channel', {
        channel: chChannel,
        config: {},
      })
      setChTestResult(result)
    } catch (e) {
      setChTestResult({ success: false, message: String(e) })
    } finally {
      setChTesting(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* System Doctor */}
      <div className="card p-5">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="section-title flex items-center gap-2">
              <Server size={18} />
              System Doctor
            </h2>
            <p className="section-subtitle">Check Node.js, OpenClaw, config, and environment</p>
          </div>
          <button className="btn-primary" onClick={runDoctor} disabled={doctorRunning}>
            {doctorRunning ? <RefreshCw size={16} className="animate-spin" /> : <FlaskConical size={16} />}
            Run Check
          </button>
        </div>

        {envInfo && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-4">
            <InfoBadge label="OS" value={envInfo.os} />
            <InfoBadge label="Arch" value={envInfo.arch} />
            <InfoBadge label="Node.js" value={envInfo.node_version || 'Not found'} ok={!!envInfo.node_version} />
            <InfoBadge label="OpenClaw" value={envInfo.openclaw_version || 'Not installed'} ok={!!envInfo.openclaw_version} />
          </div>
        )}

        {doctorRunning && (
          <div className="flex items-center gap-2 text-gray-500 py-4">
            <RefreshCw size={16} className="animate-spin" />
            Running diagnostics...
          </div>
        )}

        {doctorResults.length > 0 && (
          <div className="space-y-2">
            {doctorResults.map((r, i) => (
              <div
                key={i}
                className={`flex items-start gap-3 p-3 rounded-lg ${
                  r.passed
                    ? 'bg-green-50 dark:bg-green-900/10'
                    : 'bg-red-50 dark:bg-red-900/10'
                }`}
              >
                {r.passed ? (
                  <CheckCircle size={16} className="text-green-500 flex-shrink-0 mt-0.5" />
                ) : (
                  <XCircle size={16} className="text-red-500 flex-shrink-0 mt-0.5" />
                )}
                <div>
                  <div className="text-sm font-medium text-gray-900 dark:text-gray-100">{r.name}</div>
                  <div className={`text-xs mt-0.5 ${r.passed ? 'text-green-700 dark:text-green-400' : 'text-red-700 dark:text-red-400'}`}>
                    {r.message}
                  </div>
                  {r.suggestion && (
                    <div className="text-xs text-gray-500 dark:text-gray-400 mt-1 italic">
                      💡 {r.suggestion}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}

        {doctorResults.length === 0 && !doctorRunning && (
          <p className="text-gray-400 dark:text-gray-500 text-sm italic">
            Click "Run Check" to start diagnostics.
          </p>
        )}
      </div>

      {/* AI Connection Test */}
      <div className="card p-5">
        <h2 className="section-title flex items-center gap-2 mb-1">
          <Brain size={18} />
          AI Connection Test
        </h2>
        <p className="section-subtitle mb-4">Verify an AI provider is reachable and the API key works</p>

        <div className="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Provider</label>
            <input
              className="input"
              value={aiProvider}
              onChange={e => setAiProvider(e.target.value)}
              placeholder="e.g. anthropic"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Model</label>
            <input
              className="input"
              value={aiModel}
              onChange={e => setAiModel(e.target.value)}
              placeholder="e.g. claude-haiku-4-5"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">API Key</label>
            <input
              className="input"
              type="password"
              value={aiApiKey}
              onChange={e => setAiApiKey(e.target.value)}
              placeholder="sk-..."
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Base URL (optional)</label>
            <input
              className="input"
              value={aiBaseUrl}
              onChange={e => setAiBaseUrl(e.target.value)}
              placeholder="https://api.example.com"
            />
          </div>
        </div>

        <div className="flex items-center gap-4">
          <button className="btn-primary" onClick={testAI} disabled={aiTesting}>
            {aiTesting ? <RefreshCw size={16} className="animate-spin" /> : <Brain size={16} />}
            Test Connection
          </button>
          {aiTestResult && (
            <div
              className={`flex items-center gap-2 text-sm ${
                aiTestResult.success ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
              }`}
            >
              {aiTestResult.success ? <CheckCircle size={16} /> : <XCircle size={16} />}
              {aiTestResult.message}
            </div>
          )}
        </div>
      </div>

      {/* Channel Connectivity Test */}
      <div className="card p-5">
        <h2 className="section-title flex items-center gap-2 mb-1">
          <MessageSquare size={18} />
          Channel Connectivity Test
        </h2>
        <p className="section-subtitle mb-4">Test if a configured channel is reachable</p>

        <div className="flex items-end gap-4">
          <div className="flex-1">
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Channel</label>
            <select
              className="input"
              value={chChannel}
              onChange={e => setChChannel(e.target.value)}
            >
              {['telegram', 'discord', 'slack', 'feishu', 'wechat', 'dingtalk', 'qq', 'whatsapp', 'line'].map(ch => (
                <option key={ch} value={ch}>{ch}</option>
              ))}
            </select>
          </div>
          <button className="btn-primary" onClick={testChannel} disabled={chTesting}>
            {chTesting ? <RefreshCw size={16} className="animate-spin" /> : <MessageSquare size={16} />}
            Test Channel
          </button>
        </div>

        {chTestResult && (
          <div
            className={`mt-4 flex items-center gap-2 p-3 rounded-lg text-sm ${
              chTestResult.success
                ? 'bg-green-50 dark:bg-green-900/10 text-green-700 dark:text-green-400'
                : 'bg-red-50 dark:bg-red-900/10 text-red-700 dark:text-red-400'
            }`}
          >
            {chTestResult.success ? <CheckCircle size={16} /> : <XCircle size={16} />}
            {chTestResult.message}
          </div>
        )}
      </div>
    </div>
  )
}

function InfoBadge({ label, value, ok }: { label: string; value: string; ok?: boolean }) {
  return (
    <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
      <div className="text-xs text-gray-500 dark:text-gray-400">{label}</div>
      <div
        className={`text-sm font-medium mt-0.5 ${
          ok === undefined
            ? 'text-gray-900 dark:text-gray-100'
            : ok
            ? 'text-green-700 dark:text-green-400'
            : 'text-red-700 dark:text-red-400'
        }`}
      >
        {value}
      </div>
    </div>
  )
}
