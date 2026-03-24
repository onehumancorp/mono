import React, { useState, useEffect } from 'react'
import {
  MessageSquare,
  ChevronDown,
  ChevronUp,
  Save,
  ExternalLink,
  Check,
  RefreshCw,
  ToggleLeft,
  ToggleRight,
} from 'lucide-react'
import { invoke } from '../../lib/tauri'

interface ChannelField {
  key: string
  label: string
  type: 'text' | 'password' | 'select' | 'textarea'
  placeholder?: string
  options?: string[]
  required?: boolean
}

interface ChannelDef {
  id: string
  name: string
  icon: string
  description: string
  guide_url: string
  fields: ChannelField[]
}

const CHANNEL_DEFS: ChannelDef[] = [
  {
    id: 'telegram',
    name: 'Telegram',
    icon: '✈️',
    description: 'Connect via Telegram Bot API',
    guide_url: 'https://core.telegram.org/bots/tutorial',
    fields: [
      { key: 'bot_token', label: 'Bot Token', type: 'password', placeholder: '123456:ABC-DEF1234...', required: true },
      { key: 'allowed_chats', label: 'Allowed Chat IDs (comma separated)', type: 'text', placeholder: '-100123456,987654' },
      { key: 'group_mode', label: 'Group Mode', type: 'select', options: ['off', 'mention', 'always'] },
    ],
  },
  {
    id: 'discord',
    name: 'Discord',
    icon: '🎮',
    description: 'Connect via Discord Bot',
    guide_url: 'https://discord.com/developers/docs/intro',
    fields: [
      { key: 'bot_token', label: 'Bot Token', type: 'password', placeholder: 'MTxxxxxxx.xxx.xxx', required: true },
      { key: 'server_id', label: 'Server ID', type: 'text', placeholder: '123456789012345678' },
      { key: 'channel_id', label: 'Default Channel ID', type: 'text', placeholder: '123456789012345678' },
    ],
  },
  {
    id: 'slack',
    name: 'Slack',
    icon: '💼',
    description: 'Connect via Slack App',
    guide_url: 'https://api.slack.com/start/building',
    fields: [
      { key: 'bot_token', label: 'Bot Token', type: 'password', placeholder: 'xoxb-...', required: true },
      { key: 'signing_secret', label: 'Signing Secret', type: 'password', placeholder: 'abc123...', required: true },
      { key: 'channel_id', label: 'Default Channel ID', type: 'text', placeholder: 'C01234567' },
    ],
  },
  {
    id: 'feishu',
    name: 'Feishu / Lark',
    icon: '🐦',
    description: 'Connect via Feishu (Lark) Bot',
    guide_url: 'https://open.feishu.cn/document/home/index',
    fields: [
      { key: 'app_id', label: 'App ID', type: 'text', placeholder: 'cli_xxx', required: true },
      { key: 'app_secret', label: 'App Secret', type: 'password', placeholder: 'xxx', required: true },
      { key: 'verification_token', label: 'Verification Token', type: 'password', placeholder: 'xxx' },
      { key: 'encrypt_key', label: 'Encrypt Key', type: 'password', placeholder: 'xxx' },
      { key: 'region', label: 'Region', type: 'select', options: ['cn', 'us', 'sg', 'jp'] },
    ],
  },
  {
    id: 'wechat',
    name: 'WeChat',
    icon: '💬',
    description: 'Connect via WeChat Official Account',
    guide_url: 'https://developers.weixin.qq.com/doc/offiaccount/en_US/Getting_Started/Overview.html',
    fields: [
      { key: 'app_id', label: 'AppID', type: 'text', placeholder: 'wx...', required: true },
      { key: 'app_secret', label: 'AppSecret', type: 'password', placeholder: 'xxx', required: true },
      { key: 'token', label: 'Token', type: 'text', placeholder: 'xxx', required: true },
      { key: 'encoding_aes_key', label: 'EncodingAESKey', type: 'password', placeholder: '43 chars' },
    ],
  },
  {
    id: 'imessage',
    name: 'iMessage',
    icon: '📱',
    description: 'Connect via AppleScript on macOS',
    guide_url: 'https://support.apple.com/guide/messages/welcome/mac',
    fields: [
      { key: 'handle', label: 'Your iMessage Handle', type: 'text', placeholder: 'your@email.com or +1234567890', required: true },
      { key: 'allowed_contacts', label: 'Allowed Contacts (comma separated)', type: 'text', placeholder: '+1234567890' },
    ],
  },
  {
    id: 'dingtalk',
    name: 'DingTalk',
    icon: '📟',
    description: 'Connect via DingTalk Robot',
    guide_url: 'https://open.dingtalk.com/document/orgapp/overview-of-dingtalk-bot',
    fields: [
      { key: 'app_key', label: 'App Key', type: 'text', placeholder: 'dingtalk_xxx', required: true },
      { key: 'app_secret', label: 'App Secret', type: 'password', placeholder: 'xxx', required: true },
      { key: 'webhook_url', label: 'Webhook URL', type: 'text', placeholder: 'https://oapi.dingtalk.com/...' },
    ],
  },
  {
    id: 'qq',
    name: 'QQ',
    icon: '🐧',
    description: 'Connect via NapCat or Lagrange',
    guide_url: 'https://napneko.github.io/en-US/guide/start-install',
    fields: [
      { key: 'bot_id', label: 'Bot QQ ID', type: 'text', placeholder: '123456789', required: true },
      { key: 'bot_token', label: 'Bot Token', type: 'password', placeholder: 'xxx' },
      { key: 'ws_url', label: 'WebSocket URL', type: 'text', placeholder: 'ws://127.0.0.1:3001' },
    ],
  },
  {
    id: 'whatsapp',
    name: 'WhatsApp',
    icon: '📞',
    description: 'Connect via WhatsApp Business Cloud API',
    guide_url: 'https://developers.facebook.com/docs/whatsapp/cloud-api',
    fields: [
      { key: 'phone_number_id', label: 'Phone Number ID', type: 'text', placeholder: '123456789', required: true },
      { key: 'access_token', label: 'Access Token', type: 'password', placeholder: 'EAAxx...', required: true },
      { key: 'verify_token', label: 'Webhook Verify Token', type: 'text', placeholder: 'my_secret_token' },
    ],
  },
  {
    id: 'line',
    name: 'LINE',
    icon: '🟢',
    description: 'Connect via LINE Messaging API',
    guide_url: 'https://developers.line.biz/en/docs/messaging-api/',
    fields: [
      { key: 'channel_access_token', label: 'Channel Access Token', type: 'password', placeholder: 'xxx', required: true },
      { key: 'channel_secret', label: 'Channel Secret', type: 'password', placeholder: 'xxx', required: true },
    ],
  },
]

type ChannelConfigMap = Record<string, Record<string, unknown>>

export default function Channels() {
  const [configs, setConfigs] = useState<ChannelConfigMap>({})
  const [expanded, setExpanded] = useState<string | null>(null)
  const [saving, setSaving] = useState<string | null>(null)
  const [saved, setSaved] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadConfigs()
  }, [])

  const loadConfigs = async () => {
    setLoading(true)
    try {
      const data = await invoke<Record<string, Record<string, string>>>('get_channels_config')
      setConfigs(data || {})
    } catch {
      setConfigs({})
    } finally {
      setLoading(false)
    }
  }

  const updateField = (channelId: string, key: string, value: string) => {
    setConfigs(prev => ({
      ...prev,
      [channelId]: { ...(prev[channelId] || {}), [key]: value },
    }))
  }

  const toggleEnabled = (channelId: string) => {
    setConfigs(prev => ({
      ...prev,
      [channelId]: { ...(prev[channelId] || {}), enabled: !prev[channelId]?.enabled },
    }))
  }

  const saveChannel = async (channelId: string) => {
    setSaving(channelId)
    try {
      await invoke('save_channel_config', { channel: channelId, config: configs[channelId] || {} })
      setSaved(channelId)
      setTimeout(() => setSaved(null), 2000)
    } catch (e) {
      alert(String(e))
    } finally {
      setSaving(null)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-48">
        <RefreshCw size={24} className="animate-spin text-blue-500" />
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {CHANNEL_DEFS.map(ch => {
        const cfg = configs[ch.id] || {}
        const isExpanded = expanded === ch.id
        const isEnabled = !!cfg.enabled

        return (
          <div key={ch.id} className="card overflow-hidden">
            <div
              className="flex items-center gap-3 px-5 py-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
              onClick={() => setExpanded(isExpanded ? null : ch.id)}
            >
              <span className="text-2xl">{ch.icon}</span>
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-gray-900 dark:text-gray-100">{ch.name}</span>
                  {isEnabled && (
                    <span className="badge-running">
                      <span className="w-1.5 h-1.5 rounded-full bg-green-500" />
                      Enabled
                    </span>
                  )}
                </div>
                <p className="text-sm text-gray-500 dark:text-gray-400">{ch.description}</p>
              </div>
              <div className="flex items-center gap-3">
                <button
                  onClick={e => {
                    e.stopPropagation()
                    toggleEnabled(ch.id)
                  }}
                  className="text-gray-400 hover:text-blue-500 transition-colors"
                >
                  {isEnabled ? (
                    <ToggleRight size={24} className="text-blue-500" />
                  ) : (
                    <ToggleLeft size={24} />
                  )}
                </button>
                {isExpanded ? <ChevronUp size={16} className="text-gray-400" /> : <ChevronDown size={16} className="text-gray-400" />}
              </div>
            </div>

            {isExpanded && (
              <div className="px-5 pb-5 border-t border-gray-100 dark:border-gray-700 pt-4 space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  {ch.fields.map(field => (
                    <div key={field.key} className={field.type === 'textarea' ? 'col-span-2' : ''}>
                      <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
                        {field.label}
                        {field.required && <span className="text-red-500 ml-1">*</span>}
                      </label>
                      {field.type === 'select' ? (
                        <select
                          className="input"
                          value={String(cfg[field.key] ?? '')}
                          onChange={e => updateField(ch.id, field.key, e.target.value)}
                        >
                          <option value="">— Select —</option>
                          {field.options?.map(opt => (
                            <option key={opt} value={opt}>
                              {opt}
                            </option>
                          ))}
                        </select>
                      ) : field.type === 'textarea' ? (
                        <textarea
                          className="input h-20 resize-none"
                          value={String(cfg[field.key] ?? '')}
                          onChange={e => updateField(ch.id, field.key, e.target.value)}
                          placeholder={field.placeholder}
                        />
                      ) : (
                        <input
                          className="input"
                          type={field.type === 'password' ? 'password' : 'text'}
                          value={String(cfg[field.key] ?? '')}
                          onChange={e => updateField(ch.id, field.key, e.target.value)}
                          placeholder={field.placeholder}
                        />
                      )}
                    </div>
                  ))}
                </div>

                <div className="flex items-center gap-3 pt-2">
                  <button
                    className="btn-primary"
                    onClick={() => saveChannel(ch.id)}
                    disabled={saving === ch.id}
                  >
                    {saving === ch.id ? (
                      <RefreshCw size={14} className="animate-spin" />
                    ) : saved === ch.id ? (
                      <Check size={14} />
                    ) : (
                      <Save size={14} />
                    )}
                    {saved === ch.id ? 'Saved!' : 'Save'}
                  </button>
                  <a
                    href={ch.guide_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="btn-secondary text-sm"
                    onClick={e => e.stopPropagation()}
                  >
                    <ExternalLink size={14} />
                    Setup Guide
                  </a>
                </div>
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
