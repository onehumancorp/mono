import React, { useState, useEffect } from 'react'
import { Settings, Save, RefreshCw, Check, Download, Trash2, AlertTriangle } from 'lucide-react'
import { invoke } from '../../lib/tauri'

interface AppSettings {
  language: string
  auto_start: boolean
  minimize_to_tray: boolean
  start_service_on_launch: boolean
  check_updates_on_start: boolean
  log_level: string
}

const defaultSettings: AppSettings = {
  language: 'en',
  auto_start: false,
  minimize_to_tray: true,
  start_service_on_launch: true,
  check_updates_on_start: true,
  log_level: 'info',
}

export default function SettingsPage() {
  const [settings, setSettings] = useState<AppSettings>(defaultSettings)
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [envInfo, setEnvInfo] = useState<{ openclaw_version: string | null } | null>(null)
  const [updateStatus, setUpdateStatus] = useState<string | null>(null)
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [updating, setUpdating] = useState(false)
  const [uninstalling, setUninstalling] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadSettings()
    loadEnvInfo()
  }, [])

  const loadSettings = async () => {
    try {
      const cfg = await invoke<Record<string, unknown>>('get_config')
      if (cfg && cfg.__app_settings) {
        setSettings({ ...defaultSettings, ...(cfg.__app_settings as Partial<AppSettings>) })
      }
    } catch {
      // use defaults
    }
  }

  const loadEnvInfo = async () => {
    try {
      const info = await invoke<{ openclaw_version: string | null }>('get_system_info')
      setEnvInfo(info)
    } catch {
      //
    }
  }

  const saveSettings = async () => {
    setSaving(true)
    try {
      const cfg = await invoke<Record<string, unknown>>('get_config').catch(() => ({}))
      await invoke('save_config', { config: { ...cfg, __app_settings: settings } })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch (e) {
      setError(String(e))
    } finally {
      setSaving(false)
    }
  }

  const checkUpdate = async () => {
    setCheckingUpdate(true)
    setUpdateStatus(null)
    try {
      const result = await invoke<{ has_update: boolean; latest_version: string; current_version: string }>('check_openclaw_update')
      if (result.has_update) {
        setUpdateStatus(`Update available: ${result.current_version} → ${result.latest_version}`)
      } else {
        setUpdateStatus(`Up to date (${result.current_version})`)
      }
    } catch (e) {
      setUpdateStatus(`Check failed: ${String(e)}`)
    } finally {
      setCheckingUpdate(false)
    }
  }

  const updateOpenclaw = async () => {
    setUpdating(true)
    try {
      await invoke('update_openclaw')
      await loadEnvInfo()
      setUpdateStatus('Update successful!')
    } catch (e) {
      setError(String(e))
    } finally {
      setUpdating(false)
    }
  }

  const uninstallOpenclaw = async () => {
    if (!confirm('Are you sure you want to uninstall OpenClaw? This will remove the npm package.')) return
    setUninstalling(true)
    try {
      await invoke('uninstall_openclaw')
      await loadEnvInfo()
    } catch (e) {
      setError(String(e))
    } finally {
      setUninstalling(false)
    }
  }

  const update = (key: keyof AppSettings, value: unknown) => {
    setSettings(s => ({ ...s, [key]: value }))
  }

  return (
    <div className="space-y-6 max-w-2xl">
      {error && (
        <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-700 dark:text-red-400 text-sm">
          {error}
          <button onClick={() => setError(null)} className="ml-2 underline">Dismiss</button>
        </div>
      )}

      {/* App preferences */}
      <div className="card p-5">
        <h2 className="section-title flex items-center gap-2 mb-4">
          <Settings size={18} />
          Application Preferences
        </h2>

        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Language</label>
            <select
              className="input w-48"
              value={settings.language}
              onChange={e => update('language', e.target.value)}
            >
              <option value="en">English</option>
              <option value="zh">中文</option>
              <option value="ja">日本語</option>
              <option value="ko">한국어</option>
            </select>
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Log Level</label>
            <select
              className="input w-48"
              value={settings.log_level}
              onChange={e => update('log_level', e.target.value)}
            >
              <option value="debug">Debug</option>
              <option value="info">Info</option>
              <option value="warn">Warn</option>
              <option value="error">Error</option>
            </select>
          </div>

          <div className="space-y-3 pt-2">
            {(
              [
                { key: 'auto_start', label: 'Launch at startup', description: 'Start OpenClaw Manager when you log in' },
                { key: 'minimize_to_tray', label: 'Minimize to tray', description: 'Keep running in the system tray when closed' },
                { key: 'start_service_on_launch', label: 'Start service on launch', description: 'Automatically start the OpenClaw service when the app opens' },
                { key: 'check_updates_on_start', label: 'Check for updates on startup', description: 'Automatically check for OpenClaw updates' },
              ] as { key: keyof AppSettings; label: string; description: string }[]
            ).map(({ key, label, description }) => (
              <label key={key} className="flex items-start gap-3 cursor-pointer group">
                <div className="pt-0.5">
                  <input
                    type="checkbox"
                    checked={settings[key] as boolean}
                    onChange={e => update(key, e.target.checked)}
                    className="w-4 h-4 rounded text-blue-500"
                  />
                </div>
                <div>
                  <div className="text-sm font-medium text-gray-800 dark:text-gray-200">{label}</div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">{description}</div>
                </div>
              </label>
            ))}
          </div>
        </div>

        <div className="mt-5 pt-5 border-t border-gray-100 dark:border-gray-700">
          <button className="btn-primary" onClick={saveSettings} disabled={saving}>
            {saving ? (
              <RefreshCw size={14} className="animate-spin" />
            ) : saved ? (
              <Check size={14} />
            ) : (
              <Save size={14} />
            )}
            {saved ? 'Saved!' : 'Save Settings'}
          </button>
        </div>
      </div>

      {/* OpenClaw management */}
      <div className="card p-5">
        <h2 className="section-title mb-1">OpenClaw Installation</h2>
        <p className="section-subtitle mb-4">
          Installed version: <code className="bg-gray-100 dark:bg-gray-700 px-1 rounded text-xs">{envInfo?.openclaw_version || 'Not installed'}</code>
        </p>

        <div className="flex flex-wrap gap-3">
          <button className="btn-secondary" onClick={checkUpdate} disabled={checkingUpdate}>
            {checkingUpdate ? <RefreshCw size={14} className="animate-spin" /> : <RefreshCw size={14} />}
            Check for Updates
          </button>
          {updateStatus?.includes('Update available') && (
            <button className="btn-success" onClick={updateOpenclaw} disabled={updating}>
              {updating ? <RefreshCw size={14} className="animate-spin" /> : <Download size={14} />}
              Update Now
            </button>
          )}
          <button className="btn-danger" onClick={uninstallOpenclaw} disabled={uninstalling}>
            {uninstalling ? <RefreshCw size={14} className="animate-spin" /> : <Trash2 size={14} />}
            Uninstall OpenClaw
          </button>
        </div>

        {updateStatus && (
          <div className={`mt-3 text-sm flex items-center gap-2 ${updateStatus.includes('available') ? 'text-yellow-600 dark:text-yellow-400' : 'text-green-600 dark:text-green-400'}`}>
            {updateStatus.includes('available') ? <AlertTriangle size={14} /> : <Check size={14} />}
            {updateStatus}
          </div>
        )}
      </div>

      {/* About */}
      <div className="card p-5">
        <h2 className="section-title mb-3">About</h2>
        <div className="space-y-1 text-sm text-gray-600 dark:text-gray-400">
          <p>OpenClaw Manager v1.0.0</p>
          <p>Built with Tauri 2.0 + React 18 + TypeScript</p>
          <p>
            <a
              href="https://github.com/miaoxworld/openclaw-manager"
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-500 hover:underline"
            >
              View on GitHub ↗
            </a>
          </p>
        </div>
      </div>
    </div>
  )
}
