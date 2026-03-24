import React, { useState, useEffect } from 'react'
import {
  Plus,
  Trash2,
  Edit2,
  Check,
  X,
  Star,
  RefreshCw,
  Bot,
  ChevronDown,
  ChevronUp,
} from 'lucide-react'
import { invoke, Agent } from '../../lib/tauri'

const AVAILABLE_CHANNELS = ['telegram', 'discord', 'slack', 'feishu', 'wechat', 'imessage', 'dingtalk', 'qq', 'whatsapp', 'line']
const AVAILABLE_TOOLS = ['web_search', 'code_exec', 'file_read', 'file_write', 'http_request', 'shell_exec', 'image_gen', 'pdf_read', 'email_send', 'calendar']

const emptyAgent = (): Omit<Agent, 'id'> => ({
  name: '',
  emoji: '🤖',
  description: '',
  model_override: '',
  channels: [],
  tools: [],
  sandbox_mode: false,
  workspace_isolation: false,
  mention_mode: false,
  sub_agent_permissions: [],
  is_primary: false,
})

export default function Agents() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [isCreating, setIsCreating] = useState(false)
  const [form, setForm] = useState<Omit<Agent, 'id'>>(emptyAgent())
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [expanded, setExpanded] = useState<string | null>(null)

  useEffect(() => {
    loadAgents()
  }, [])

  const loadAgents = async () => {
    setLoading(true)
    try {
      const list = await invoke<Agent[]>('get_agents_list')
      setAgents(list || [])
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }

  const startCreate = () => {
    setIsCreating(true)
    setEditingId(null)
    setForm(emptyAgent())
  }

  const startEdit = (agent: Agent) => {
    setIsCreating(false)
    setEditingId(agent.id)
    setForm({ ...agent })
    setExpanded(agent.id)
  }

  const cancel = () => {
    setIsCreating(false)
    setEditingId(null)
    setForm(emptyAgent())
  }

  const save = async () => {
    if (!form.name.trim()) return alert('Agent name is required')
    setSaving(true)
    try {
      const agentData: Agent = {
        id: editingId || `agent_${Date.now()}`,
        ...form,
      }
      await invoke('save_agent', { agent: agentData })
      await loadAgents()
      cancel()
      setError(null)
    } catch (e) {
      setError(String(e))
    } finally {
      setSaving(false)
    }
  }

  const deleteAgent = async (id: string) => {
    if (!confirm('Delete this agent?')) return
    try {
      await invoke('delete_agent', { id })
      await loadAgents()
    } catch (e) {
      setError(String(e))
    }
  }

  const setDefaultAgent = async (id: string) => {
    try {
      await invoke('set_default_agent', { id })
      await loadAgents()
    } catch (e) {
      setError(String(e))
    }
  }

  const toggleChannel = (ch: string) => {
    setForm(f => ({
      ...f,
      channels: f.channels.includes(ch) ? f.channels.filter(c => c !== ch) : [...f.channels, ch],
    }))
  }

  const toggleTool = (tool: string) => {
    setForm(f => ({
      ...f,
      tools: f.tools.includes(tool) ? f.tools.filter(t => t !== tool) : [...f.tools, tool],
    }))
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-48">
        <RefreshCw size={24} className="animate-spin text-blue-500" />
      </div>
    )
  }

  const renderForm = () => (
    <div className="space-y-5">
      <div className="grid grid-cols-12 gap-4">
        <div className="col-span-2">
          <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Emoji</label>
          <input
            className="input text-2xl text-center"
            value={form.emoji}
            onChange={e => setForm(f => ({ ...f, emoji: e.target.value }))}
            maxLength={2}
          />
        </div>
        <div className="col-span-10">
          <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
            Name <span className="text-red-500">*</span>
          </label>
          <input
            className="input"
            value={form.name}
            onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
            placeholder="My Agent"
          />
        </div>
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
        <textarea
          className="input h-20 resize-none"
          value={form.description}
          onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
          placeholder="What does this agent do?"
        />
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Model Override</label>
        <input
          className="input"
          value={form.model_override}
          onChange={e => setForm(f => ({ ...f, model_override: e.target.value }))}
          placeholder="Leave blank to use primary model"
        />
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">
          Bound Channels
        </label>
        <div className="flex flex-wrap gap-2">
          {AVAILABLE_CHANNELS.map(ch => (
            <button
              key={ch}
              onClick={() => toggleChannel(ch)}
              className={`px-3 py-1 rounded-full text-xs font-medium border transition-colors ${
                form.channels.includes(ch)
                  ? 'bg-blue-100 dark:bg-blue-900/30 border-blue-300 dark:border-blue-700 text-blue-700 dark:text-blue-300'
                  : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:border-blue-300'
              }`}
            >
              {ch}
            </button>
          ))}
        </div>
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">
          Tool Permissions
        </label>
        <div className="flex flex-wrap gap-2">
          {AVAILABLE_TOOLS.map(tool => (
            <button
              key={tool}
              onClick={() => toggleTool(tool)}
              className={`px-3 py-1 rounded-full text-xs font-medium border transition-colors ${
                form.tools.includes(tool)
                  ? 'bg-purple-100 dark:bg-purple-900/30 border-purple-300 dark:border-purple-700 text-purple-700 dark:text-purple-300'
                  : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:border-purple-300'
              }`}
            >
              {tool}
            </button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4">
        {(
          [
            { key: 'sandbox_mode', label: 'Sandbox Mode' },
            { key: 'workspace_isolation', label: 'Workspace Isolation' },
            { key: 'mention_mode', label: '@Mention Mode' },
          ] as { key: keyof typeof form; label: string }[]
        ).map(({ key, label }) => (
          <label key={key} className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={form[key] as boolean}
              onChange={e => setForm(f => ({ ...f, [key]: e.target.checked }))}
              className="w-4 h-4 rounded text-blue-500"
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">{label}</span>
          </label>
        ))}
      </div>

      <div className="flex gap-3 pt-2">
        <button className="btn-primary" onClick={save} disabled={saving}>
          {saving ? <RefreshCw size={14} className="animate-spin" /> : <Check size={14} />}
          {isCreating ? 'Create Agent' : 'Save Changes'}
        </button>
        <button className="btn-secondary" onClick={cancel}>
          <X size={14} />
          Cancel
        </button>
      </div>
    </div>
  )

  return (
    <div className="space-y-4">
      {error && (
        <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-700 dark:text-red-400 text-sm">
          {error}
        </div>
      )}

      <div className="flex items-center justify-between">
        <div>
          <h2 className="section-title">Agents ({agents.length})</h2>
          <p className="section-subtitle">Create and configure AI agents</p>
        </div>
        <button className="btn-primary" onClick={startCreate} disabled={isCreating}>
          <Plus size={16} />
          New Agent
        </button>
      </div>

      {isCreating && (
        <div className="card p-5">
          <h3 className="font-medium text-gray-900 dark:text-gray-100 mb-4">Create New Agent</h3>
          {renderForm()}
        </div>
      )}

      <div className="space-y-3">
        {agents.map(agent => (
          <div key={agent.id} className="card overflow-hidden">
            <div
              className="flex items-center gap-4 px-5 py-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50"
              onClick={() => setExpanded(expanded === agent.id ? null : agent.id)}
            >
              <div className="text-3xl">{agent.emoji}</div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-gray-900 dark:text-gray-100">{agent.name}</span>
                  {agent.is_primary && (
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-400">
                      <Star size={10} />
                      Primary
                    </span>
                  )}
                </div>
                <p className="text-sm text-gray-500 dark:text-gray-400 truncate">{agent.description}</p>
                {agent.channels.length > 0 && (
                  <div className="flex gap-1 mt-1 flex-wrap">
                    {agent.channels.map(c => (
                      <span key={c} className="text-xs px-2 py-0.5 rounded-full bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300">
                        {c}
                      </span>
                    ))}
                  </div>
                )}
              </div>
              <div className="flex items-center gap-1">
                <button
                  onClick={e => { e.stopPropagation(); setDefaultAgent(agent.id) }}
                  className="p-1.5 text-gray-400 hover:text-yellow-500 transition-colors"
                  title="Set as primary"
                >
                  <Star size={14} className={agent.is_primary ? 'fill-yellow-400 text-yellow-400' : ''} />
                </button>
                <button
                  onClick={e => { e.stopPropagation(); startEdit(agent) }}
                  className="p-1.5 text-gray-400 hover:text-blue-500 transition-colors"
                >
                  <Edit2 size={14} />
                </button>
                <button
                  onClick={e => { e.stopPropagation(); deleteAgent(agent.id) }}
                  className="p-1.5 text-gray-400 hover:text-red-500 transition-colors"
                >
                  <Trash2 size={14} />
                </button>
                {expanded === agent.id ? <ChevronUp size={14} className="text-gray-400" /> : <ChevronDown size={14} className="text-gray-400" />}
              </div>
            </div>

            {editingId === agent.id && expanded === agent.id && (
              <div className="px-5 pb-5 border-t border-gray-100 dark:border-gray-700 pt-4">
                {renderForm()}
              </div>
            )}
          </div>
        ))}

        {agents.length === 0 && !isCreating && (
          <div className="card p-10 flex flex-col items-center text-center">
            <Bot size={40} className="text-gray-300 dark:text-gray-600 mb-3" />
            <p className="text-gray-500 dark:text-gray-400">No agents configured yet.</p>
            <p className="text-sm text-gray-400 dark:text-gray-500 mt-1">Click "New Agent" to create your first agent.</p>
          </div>
        )}
      </div>
    </div>
  )
}
