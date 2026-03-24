import React from 'react'
import {
  LayoutDashboard,
  Brain,
  MessageSquare,
  Bot,
  Package,
  Shield,
  FlaskConical,
  FileText,
  Settings,
  Zap,
} from 'lucide-react'
import { Page } from '../../App'

interface SidebarProps {
  currentPage: Page
  onNavigate: (page: Page) => void
}

const navItems: { id: Page; label: string; icon: React.ReactNode }[] = [
  { id: 'dashboard', label: 'Dashboard', icon: <LayoutDashboard size={18} /> },
  { id: 'aiconfig', label: 'AI Models', icon: <Brain size={18} /> },
  { id: 'channels', label: 'Channels', icon: <MessageSquare size={18} /> },
  { id: 'agents', label: 'Agents', icon: <Bot size={18} /> },
  { id: 'skills', label: 'Skills', icon: <Package size={18} /> },
  { id: 'security', label: 'Security', icon: <Shield size={18} /> },
  { id: 'testing', label: 'Diagnostics', icon: <FlaskConical size={18} /> },
  { id: 'logs', label: 'Logs', icon: <FileText size={18} /> },
  { id: 'settings', label: 'Settings', icon: <Settings size={18} /> },
]

export default function Sidebar({ currentPage, onNavigate }: SidebarProps) {
  return (
    <div className="flex flex-col w-60 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-700 h-full">
      {/* Logo */}
      <div className="flex items-center gap-3 px-5 py-4 border-b border-gray-200 dark:border-gray-700">
        <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
          <Zap size={18} className="text-white" />
        </div>
        <div>
          <div className="font-bold text-gray-900 dark:text-gray-100 text-sm">OpenClaw</div>
          <div className="text-xs text-gray-500 dark:text-gray-400">Manager</div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-3 py-4 space-y-0.5 overflow-y-auto">
        {navItems.map(item => (
          <button
            key={item.id}
            onClick={() => onNavigate(item.id)}
            className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
              currentPage === item.id
                ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300'
                : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-100'
            }`}
          >
            <span
              className={
                currentPage === item.id
                  ? 'text-blue-600 dark:text-blue-400'
                  : 'text-gray-400 dark:text-gray-500'
              }
            >
              {item.icon}
            </span>
            {item.label}
          </button>
        ))}
      </nav>

      {/* Footer */}
      <div className="px-5 py-3 border-t border-gray-200 dark:border-gray-700">
        <p className="text-xs text-gray-400 dark:text-gray-500">v1.0.0 · OpenClaw Manager</p>
      </div>
    </div>
  )
}
