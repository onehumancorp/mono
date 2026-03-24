import React, { useState, useEffect } from 'react'
import { invoke, AIProvider } from '../../lib/tauri'
import { Plus, Trash2, Edit2, Check, X, Eye, EyeOff, ChevronDown, ChevronUp, RefreshCw, Star } from 'lucide-react'
import { logger } from '../../lib/logger'

interface OfficialProvider {
  id: string
  name: string
  default_base_url: string
  default_models: string[]
  requires_api_key: boolean
  website: string
}

export default function AIConfig() {
  const [providers, setProviders] = useState<AIProvider[]>([])
  const [officialProviders, setOfficialProviders] = useState<OfficialProvider[]>([])
  const [primaryModel, setPrimaryModel] = useState('')
  const [loading, setLoading] = useState(true)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [showKeyFor, setShowKeyFor] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [editForm, setEditForm] = useState<Partial<AIProvider>>({})
  const [newModelInput, setNewModelInput] = useState('')

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [aiCfg, official] = await Promise.all([
        invoke<{ providers: AIProvider[]; primary_model: string }>('get_ai_config'),
        invoke<OfficialProvider[]>('get_official_providers'),
      ])
      setProviders(aiCfg.providers || [])
      setPrimaryModel(aiCfg.primary_model || '')
      setOfficialProviders(official)
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }

  const startEdit = (provider: AIProvider) => {
    setEditingId(provider.id)
    setEditForm({ ...provider })
  }

  const cancelEdit = () => {
    setEditingId(null)
    setEditForm({})
  }

  const saveProvider = async () => {
    if (!editForm.id) return
    setSaving(true)
    try {
      await invoke('save_provider', {
        id: editForm.id,
        baseUrl: editForm.base_url || '',
        apiKey: editForm.api_key || '',
        models: editForm.models || [],
      })
      await loadData()
      setEditingId(null)
      setError(null)
    } catch (e) {
      setError(String(e))
    } finally {
      setSaving(false)
    }
  }

  const deleteProvider = async (id: string) => {
    if (!confirm(`Delete provider "${id}"?`)) return
    try {
      await invoke('delete_provider', { id })
      await loadData()
    } catch (e) {
      setError(String(e))
    }
  }

  const setAsPrimary = async (model: string) => {
    try {
      await invoke('set_primary_model', { model })
      setPrimaryModel(model)
    } catch (e) {
      setError(String(e))
    }
  }

  const addOfficialProvider = (op: OfficialProvider) => {
    const exists = providers.find(p => p.id === op.id)
    if (exists) {
      startEdit(exists)
      return
    }
    const newProvider: AIProvider = {
      id: op.id,
      name: op.name,
      base_url: op.default_base_url,
      api_key: '',
      models: [...op.default_models],
      is_official: true,
    }
    setEditingId(op.id)
    setEditForm(newProvider)
    setProviders(prev => [...prev, newProvider])
  }

  const addModelToEdit = () => {
    if (!newModelInput.trim()) return
    setEditForm(f => ({ ...f, models: [...(f.models || []), newModelInput.trim()] }))
    setNewModelInput('')
  }

  const removeModelFromEdit = (model: string) => {
    setEditForm(f => ({ ...f, models: (f.models || []).filter(m => m !== model) }))
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-48">
        <RefreshCw size={24} className="animate-spin text-blue-500" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {error && (
        <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-700 dark:text-red-400 text-sm">
          {error}
        </div>
      )}

      {/* Primary model */}
      <div className="card p-5">
        <h2 className="section-title mb-1">Primary Model</h2>
        <p className="section-subtitle mb-4">The default model used for all agents</p>
        <div className="flex gap-3">
          <input
            className="input flex-1"
            value={primaryModel}
            onChange={e => setPrimaryModel(e.target.value)}
            placeholder="e.g. anthropic/claude-sonnet-4-5"
          />
          <button className="btn-primary" onClick={() => setAsPrimary(primaryModel)}>
            <Check size={16} />
            Set
          </button>
        </div>
      </div>

      {/* Official providers catalog */}
      <div className="card p-5">
        <h2 className="section-title mb-1">Provider Catalog</h2>
        <p className="section-subtitle mb-4">Click to configure a provider</p>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
          {officialProviders.map(op => {
            const configured = providers.find(p => p.id === op.id)
            return (
              <button
                key={op.id}
                onClick={() => addOfficialProvider(op)}
                className={`flex flex-col items-start p-3 rounded-lg border text-left transition-colors ${
                  configured
                    ? 'border-blue-300 dark:border-blue-700 bg-blue-50 dark:bg-blue-900/20'
                    : 'border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-700 hover:bg-gray-50 dark:hover:bg-gray-800'
                }`}
              >
                <div className="flex items-center gap-2 w-full">
                  <span className="font-medium text-sm text-gray-900 dark:text-gray-100">{op.name}</span>
                  {configured && <Check size={12} className="text-blue-500 ml-auto" />}
                </div>
                <span className="text-xs text-gray-400 mt-0.5">
                  {op.default_models.length > 0 ? `${op.default_models.length} models` : 'Custom'}
                </span>
              </button>
            )
          })}
        </div>
      </div>

      {/* Configured providers */}
      <div className="card p-5">
        <h2 className="section-title mb-4">Configured Providers</h2>
        {providers.length === 0 ? (
          <p className="text-gray-500 dark:text-gray-400 text-sm italic">
            No providers configured. Click a provider above to get started.
          </p>
        ) : (
          <div className="space-y-3">
            {providers.map(provider => (
              <div
                key={provider.id}
                className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden"
              >
                <div
                  className="flex items-center gap-3 px-4 py-3 bg-gray-50 dark:bg-gray-800/50 cursor-pointer"
                  onClick={() =>
                    setExpandedId(expandedId === provider.id ? null : provider.id)
                  }
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm text-gray-900 dark:text-gray-100">
                        {provider.name}
                      </span>
                      <span className="text-xs text-gray-400">({provider.id})</span>
                    </div>
                    <div className="text-xs text-gray-500 mt-0.5 truncate max-w-xs">
                      {provider.base_url}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {provider.models.map(m => (
                      <button
                        key={m}
                        onClick={e => {
                          e.stopPropagation()
                          setAsPrimary(`${provider.id}/${m}`)
                        }}
                        className={`text-xs px-2 py-0.5 rounded-full border transition-colors ${
                          primaryModel === `${provider.id}/${m}`
                            ? 'bg-yellow-100 dark:bg-yellow-900/30 border-yellow-300 dark:border-yellow-700 text-yellow-800 dark:text-yellow-400'
                            : 'border-gray-200 dark:border-gray-600 hover:border-blue-400 text-gray-600 dark:text-gray-400'
                        }`}
                        title="Set as primary model"
                      >
                        {primaryModel === `${provider.id}/${m}` && (
                          <Star size={10} className="inline mr-0.5" />
                        )}
                        {m}
                      </button>
                    ))}
                  </div>
                  <div className="flex items-center gap-1 ml-2">
                    <button
                      onClick={e => {
                        e.stopPropagation()
                        startEdit(provider)
                      }}
                      className="p-1.5 text-gray-400 hover:text-blue-500 transition-colors"
                    >
                      <Edit2 size={14} />
                    </button>
                    <button
                      onClick={e => {
                        e.stopPropagation()
                        deleteProvider(provider.id)
                      }}
                      className="p-1.5 text-gray-400 hover:text-red-500 transition-colors"
                    >
                      <Trash2 size={14} />
                    </button>
                    {expandedId === provider.id ? (
                      <ChevronUp size={14} className="text-gray-400" />
                    ) : (
                      <ChevronDown size={14} className="text-gray-400" />
                    )}
                  </div>
                </div>

                {/* Expanded edit form */}
                {editingId === provider.id && (
                  <div className="p-4 border-t border-gray-200 dark:border-gray-700 space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
                          Base URL
                        </label>
                        <input
                          className="input"
                          value={editForm.base_url || ''}
                          onChange={e => setEditForm(f => ({ ...f, base_url: e.target.value }))}
                          placeholder="https://api.example.com"
                        />
                      </div>
                      <div>
                        <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
                          API Key
                        </label>
                        <div className="relative">
                          <input
                            className="input pr-10"
                            type={showKeyFor === provider.id ? 'text' : 'password'}
                            value={editForm.api_key || ''}
                            onChange={e => setEditForm(f => ({ ...f, api_key: e.target.value }))}
                            placeholder="sk-..."
                          />
                          <button
                            className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                            onClick={() =>
                              setShowKeyFor(showKeyFor === provider.id ? null : provider.id)
                            }
                          >
                            {showKeyFor === provider.id ? <EyeOff size={14} /> : <Eye size={14} />}
                          </button>
                        </div>
                      </div>
                    </div>

                    <div>
                      <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">
                        Models
                      </label>
                      <div className="flex flex-wrap gap-2 mb-2">
                        {(editForm.models || []).map(m => (
                          <span
                            key={m}
                            className="inline-flex items-center gap-1 px-2 py-1 bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded text-xs"
                          >
                            {m}
                            <button
                              onClick={() => removeModelFromEdit(m)}
                              className="hover:text-red-500"
                            >
                              <X size={10} />
                            </button>
                          </span>
                        ))}
                      </div>
                      <div className="flex gap-2">
                        <input
                          className="input flex-1"
                          value={newModelInput}
                          onChange={e => setNewModelInput(e.target.value)}
                          placeholder="Add model name..."
                          onKeyDown={e => e.key === 'Enter' && addModelToEdit()}
                        />
                        <button className="btn-secondary" onClick={addModelToEdit}>
                          <Plus size={14} />
                          Add
                        </button>
                      </div>
                    </div>

                    <div className="flex gap-2 pt-2">
                      <button className="btn-primary" onClick={saveProvider} disabled={saving}>
                        {saving ? <RefreshCw size={14} className="animate-spin" /> : <Check size={14} />}
                        Save
                      </button>
                      <button className="btn-secondary" onClick={cancelEdit}>
                        <X size={14} />
                        Cancel
                      </button>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
