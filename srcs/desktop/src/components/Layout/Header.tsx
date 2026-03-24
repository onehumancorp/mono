import React from 'react'
import { Sun, Moon } from 'lucide-react'
import { useTheme } from '../../lib/ThemeContext'
import { Page } from '../../App'

const pageTitles: Record<Page, { title: string; subtitle: string }> = {
  dashboard: { title: 'Dashboard', subtitle: 'Service status and controls' },
  aiconfig: { title: 'AI Model Config', subtitle: 'Configure AI providers and models' },
  channels: { title: 'Message Channels', subtitle: 'Manage messaging integrations' },
  agents: { title: 'Agent Management', subtitle: 'Create and configure AI agents' },
  skills: { title: 'Skills Library', subtitle: 'Browse and manage skill plugins' },
  security: { title: 'Security', subtitle: 'Security scan and risk management' },
  testing: { title: 'Diagnostics', subtitle: 'System checks and connection tests' },
  logs: { title: 'Service Logs', subtitle: 'Real-time log viewer' },
  settings: { title: 'Settings', subtitle: 'Application preferences' },
}

interface HeaderProps {
  currentPage: Page
}

export default function Header({ currentPage }: HeaderProps) {
  const { theme, toggleTheme } = useTheme()
  const info = pageTitles[currentPage]

  return (
    <header className="flex items-center justify-between px-6 py-4 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">
      <div>
        <h1 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{info.title}</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400">{info.subtitle}</p>
      </div>
      <button
        onClick={toggleTheme}
        className="p-2 rounded-lg text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
        title={theme === 'light' ? 'Switch to dark mode' : 'Switch to light mode'}
      >
        {theme === 'light' ? <Moon size={18} /> : <Sun size={18} />}
      </button>
    </header>
  )
}
