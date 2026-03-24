import React, { useState } from 'react'
import { ThemeProvider } from './lib/ThemeContext'
import Sidebar from './components/Layout/Sidebar'
import Header from './components/Layout/Header'
import Dashboard from './components/Dashboard'
import AIConfig from './components/AIConfig'
import Channels from './components/Channels'
import Agents from './components/Agents'
import Skills from './components/Skills'
import Security from './components/Security'
import Testing from './components/Testing'
import Logs from './components/Logs'
import Settings from './components/Settings'

export type Page =
  | 'dashboard'
  | 'aiconfig'
  | 'channels'
  | 'agents'
  | 'skills'
  | 'security'
  | 'testing'
  | 'logs'
  | 'settings'

export default function App() {
  const [currentPage, setCurrentPage] = useState<Page>('dashboard')

  const renderPage = () => {
    switch (currentPage) {
      case 'dashboard':
        return <Dashboard />
      case 'aiconfig':
        return <AIConfig />
      case 'channels':
        return <Channels />
      case 'agents':
        return <Agents />
      case 'skills':
        return <Skills />
      case 'security':
        return <Security />
      case 'testing':
        return <Testing />
      case 'logs':
        return <Logs />
      case 'settings':
        return <Settings />
      default:
        return <Dashboard />
    }
  }

  return (
    <ThemeProvider>
      <div className="flex h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 overflow-hidden">
        <Sidebar currentPage={currentPage} onNavigate={setCurrentPage} />
        <div className="flex flex-col flex-1 overflow-hidden">
          <Header currentPage={currentPage} />
          <main className="flex-1 overflow-auto p-6">{renderPage()}</main>
        </div>
      </div>
    </ThemeProvider>
  )
}
