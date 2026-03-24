import React, { useState, useEffect } from 'react'
import {
  Package,
  Search,
  Download,
  Trash2,
  ToggleRight,
  ToggleLeft,
  Settings2,
  RefreshCw,
  Eye,
  EyeOff,
  ChevronDown,
  ChevronUp,
  Plus,
  Check,
} from 'lucide-react'
import { invoke, Skill } from '../../lib/tauri'

type FilterCategory = 'all' | 'builtin' | 'official' | 'community'

export default function Skills() {
  const [skills, setSkills] = useState<Skill[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<FilterCategory>('all')
  const [search, setSearch] = useState('')
  const [expandedConfig, setExpandedConfig] = useState<string | null>(null)
  const [configValues, setConfigValues] = useState<Record<string, Record<string, unknown>>>({})
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const [customSource, setCustomSource] = useState('')
  const [showCustomInstall, setShowCustomInstall] = useState(false)
  const [showPasswordFor, setShowPasswordFor] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadSkills()
  }, [])

  const loadSkills = async () => {
    setLoading(true)
    try {
      const list = await invoke<Skill[]>('get_skills_list')
      setSkills(list || [])
      const vals: Record<string, Record<string, unknown>> = {}
      for (const s of list || []) {
        vals[s.name] = s.config_values || {}
      }
      setConfigValues(vals)
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }

  const installSkill = async (name: string) => {
    setActionLoading(`install_${name}`)
    try {
      await invoke('install_skill', { name })
      await loadSkills()
    } catch (e) {
      setError(String(e))
    } finally {
      setActionLoading(null)
    }
  }

  const uninstallSkill = async (name: string) => {
    if (!confirm(`Uninstall skill "${name}"?`)) return
    setActionLoading(`uninstall_${name}`)
    try {
      await invoke('uninstall_skill', { name })
      await loadSkills()
    } catch (e) {
      setError(String(e))
    } finally {
      setActionLoading(null)
    }
  }

  const installCustom = async () => {
    if (!customSource.trim()) return
    setActionLoading('custom')
    try {
      await invoke('install_custom_skill', { source: customSource.trim() })
      await loadSkills()
      setCustomSource('')
      setShowCustomInstall(false)
    } catch (e) {
      setError(String(e))
    } finally {
      setActionLoading(null)
    }
  }

  const saveSkillConfig = async (name: string) => {
    setActionLoading(`config_${name}`)
    try {
      await invoke('save_skill_config', { name, config: configValues[name] || {} })
    } catch (e) {
      setError(String(e))
    } finally {
      setActionLoading(null)
    }
  }

  const updateConfigValue = (skillName: string, key: string, value: unknown) => {
    setConfigValues(prev => ({
      ...prev,
      [skillName]: { ...(prev[skillName] || {}), [key]: value },
    }))
  }

  const filteredSkills = skills.filter(s => {
    const matchCat = filter === 'all' || s.category === filter
    const matchSearch = !search || s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.description.toLowerCase().includes(search.toLowerCase())
    return matchCat && matchSearch
  })

  const categoryBadge = (cat: string) => {
    const styles: Record<string, string> = {
      builtin: 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300',
      official: 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300',
      community: 'bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300',
    }
    return (
      <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${styles[cat] || ''}`}>
        {cat}
      </span>
    )
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-48">
        <RefreshCw size={24} className="animate-spin text-blue-500" />
      </div>
    )
  }

  return (
    <div className="space-y-5">
      {error && (
        <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-700 dark:text-red-400 text-sm">
          {error}
          <button onClick={() => setError(null)} className="ml-2 underline">Dismiss</button>
        </div>
      )}

      {/* Toolbar */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            className="input pl-9"
            placeholder="Search skills..."
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
        </div>
        <div className="flex gap-1 bg-gray-100 dark:bg-gray-800 rounded-lg p-1">
          {(['all', 'builtin', 'official', 'community'] as FilterCategory[]).map(cat => (
            <button
              key={cat}
              onClick={() => setFilter(cat)}
              className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors capitalize ${
                filter === cat
                  ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm'
                  : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>
        <button
          className="btn-secondary"
          onClick={() => setShowCustomInstall(s => !s)}
        >
          <Plus size={16} />
          Custom Install
        </button>
        <button className="btn-secondary" onClick={loadSkills}>
          <RefreshCw size={16} />
          Refresh
        </button>
      </div>

      {/* Custom install */}
      {showCustomInstall && (
        <div className="card p-4">
          <h3 className="font-medium text-gray-900 dark:text-gray-100 mb-3">Install Custom Skill</h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-3">
            Enter an npm package name (e.g. <code className="bg-gray-100 dark:bg-gray-700 px-1 rounded">@myorg/my-skill</code>) or a local path (e.g. <code className="bg-gray-100 dark:bg-gray-700 px-1 rounded">/path/to/skill</code>).
          </p>
          <div className="flex gap-3">
            <input
              className="input flex-1"
              value={customSource}
              onChange={e => setCustomSource(e.target.value)}
              placeholder="npm package name or /local/path"
              onKeyDown={e => e.key === 'Enter' && installCustom()}
            />
            <button
              className="btn-primary"
              onClick={installCustom}
              disabled={actionLoading === 'custom'}
            >
              {actionLoading === 'custom' ? <RefreshCw size={14} className="animate-spin" /> : <Download size={14} />}
              Install
            </button>
          </div>
        </div>
      )}

      {/* Skills list */}
      <div className="space-y-2">
        {filteredSkills.length === 0 ? (
          <div className="card p-10 flex flex-col items-center text-center">
            <Package size={40} className="text-gray-300 dark:text-gray-600 mb-3" />
            <p className="text-gray-500 dark:text-gray-400">No skills found.</p>
          </div>
        ) : (
          filteredSkills.map(skill => {
            const isExpanded = expandedConfig === skill.name
            const cfgKey = `config_${skill.name}`
            return (
              <div key={skill.name} className="card overflow-hidden">
                <div className="flex items-center gap-4 px-5 py-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="font-medium text-gray-900 dark:text-gray-100">{skill.name}</span>
                      <span className="text-xs text-gray-400">v{skill.version}</span>
                      {categoryBadge(skill.category)}
                      {skill.installed && skill.enabled && (
                        <span className="badge-running">
                          <span className="w-1.5 h-1.5 rounded-full bg-green-500" />
                          Active
                        </span>
                      )}
                      {skill.installed && !skill.enabled && (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400">
                          Disabled
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5 truncate">
                      {skill.description}
                    </p>
                  </div>

                  <div className="flex items-center gap-2 flex-shrink-0">
                    {skill.installed && skill.config_schema?.length > 0 && (
                      <button
                        onClick={() => setExpandedConfig(isExpanded ? null : skill.name)}
                        className="p-1.5 text-gray-400 hover:text-blue-500 transition-colors"
                        title="Configure"
                      >
                        {isExpanded ? <ChevronUp size={16} /> : <Settings2 size={16} />}
                      </button>
                    )}

                    {skill.installed ? (
                      <>
                        <button
                          onClick={() => uninstallSkill(skill.name)}
                          disabled={actionLoading === `uninstall_${skill.name}`}
                          className="p-1.5 text-gray-400 hover:text-red-500 transition-colors"
                          title="Uninstall"
                        >
                          {actionLoading === `uninstall_${skill.name}` ? (
                            <RefreshCw size={16} className="animate-spin" />
                          ) : (
                            <Trash2 size={16} />
                          )}
                        </button>
                      </>
                    ) : (
                      <button
                        onClick={() => installSkill(skill.name)}
                        disabled={actionLoading === `install_${skill.name}`}
                        className="btn-primary text-xs py-1.5"
                      >
                        {actionLoading === `install_${skill.name}` ? (
                          <RefreshCw size={12} className="animate-spin" />
                        ) : (
                          <Download size={12} />
                        )}
                        Install
                      </button>
                    )}
                  </div>
                </div>

                {isExpanded && skill.config_schema?.length > 0 && (
                  <div className="px-5 pb-5 border-t border-gray-100 dark:border-gray-700 pt-4 space-y-4">
                    <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300">Configuration</h4>
                    <div className="grid grid-cols-2 gap-4">
                      {skill.config_schema.map(field => {
                        const val = String(configValues[skill.name]?.[field.key] ?? field.default_value ?? '')
                        return (
                          <div key={field.key}>
                            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
                              {field.label}
                              {field.required && <span className="text-red-500 ml-1">*</span>}
                            </label>
                            {field.field_type === 'toggle' ? (
                              <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                  type="checkbox"
                                  checked={val === 'true'}
                                  onChange={e =>
                                    updateConfigValue(skill.name, field.key, String(e.target.checked))
                                  }
                                  className="w-4 h-4 rounded text-blue-500"
                                />
                                <span className="text-sm text-gray-600 dark:text-gray-400">Enabled</span>
                              </label>
                            ) : field.field_type === 'select' ? (
                              <select
                                className="input"
                                value={val}
                                onChange={e => updateConfigValue(skill.name, field.key, e.target.value)}
                              >
                                {field.options.map(opt => (
                                  <option key={opt} value={opt}>{opt}</option>
                                ))}
                              </select>
                            ) : (
                              <div className="relative">
                                <input
                                  className="input pr-8"
                                  type={
                                    field.field_type === 'password' &&
                                    showPasswordFor !== `${skill.name}_${field.key}`
                                      ? 'password'
                                      : 'text'
                                  }
                                  value={val}
                                  onChange={e => updateConfigValue(skill.name, field.key, e.target.value)}
                                />
                                {field.field_type === 'password' && (
                                  <button
                                    className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                                    onClick={() =>
                                      setShowPasswordFor(
                                        showPasswordFor === `${skill.name}_${field.key}`
                                          ? null
                                          : `${skill.name}_${field.key}`
                                      )
                                    }
                                  >
                                    {showPasswordFor === `${skill.name}_${field.key}` ? (
                                      <EyeOff size={14} />
                                    ) : (
                                      <Eye size={14} />
                                    )}
                                  </button>
                                )}
                              </div>
                            )}
                          </div>
                        )
                      })}
                    </div>
                    <button
                      className="btn-primary"
                      onClick={() => saveSkillConfig(skill.name)}
                      disabled={actionLoading === cfgKey}
                    >
                      {actionLoading === cfgKey ? (
                        <RefreshCw size={14} className="animate-spin" />
                      ) : (
                        <Check size={14} />
                      )}
                      Save Configuration
                    </button>
                  </div>
                )}
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
