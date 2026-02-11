import { useState, useEffect } from 'react'
import TenantSelector from './components/TenantSelector'
import EventFeed from './components/EventFeed'
import EventStats from './components/EventStats'
import CreateTenantForm from './components/CreateTenantForm'
import EventIngestForm from './components/EventIngestForm'
import { useToast } from './components/Toast'
import { getErrorInfo, logError } from './lib/errors'

interface Tenant {
  id: string
  name: string
  api_key: string
  active: boolean
}

interface EventStats {
  total: number
  [key: string]: number
}

function App() {
  const [selectedTenant, setSelectedTenant] = useState<Tenant | null>(null)
  const [tenants, setTenants] = useState<Tenant[]>([])
  const [stats, setStats] = useState<EventStats | null>(null)
  const [showCreateTenant, setShowCreateTenant] = useState(false)
  const [showIngestForm, setShowIngestForm] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState<'connected' | 'disconnected' | 'connecting'>('disconnected')
  const [isLoading, setIsLoading] = useState(true)
  
  const toast = useToast()
  // Use relative paths when VITE_API_URL is empty (Docker/Nginx proxy)
  // In development, set VITE_API_URL=http://localhost:8080/api/v1
  const API_BASE = import.meta.env.VITE_API_URL || '/api/v1'

  useEffect(() => {
    fetchTenants()
  }, [])

  useEffect(() => {
    if (selectedTenant) {
      fetchStats()
      const interval = setInterval(fetchStats, 10000) // Refresh stats every 10s
      return () => clearInterval(interval)
    }
  }, [selectedTenant])

  const fetchTenants = async () => {
    setIsLoading(true)
    try {
      const response = await fetch(`${API_BASE}/tenants-with-keys`)
      if (!response.ok) {
        const errorInfo = await response.json()
        const info = getErrorInfo({ error: errorInfo.error })
        throw new Error(info.message)
      }
      const data = await response.json()
      setTenants(data.tenants || [])
    } catch (error) {
      logError(error, 'fetchTenants')
      const info = getErrorInfo(error)
      toast.showError(info.title, info.message)
    } finally {
      setIsLoading(false)
    }
  }

  const fetchStats = async () => {
    if (!selectedTenant) return
    try {
      const response = await fetch(`${API_BASE}/events/stats`, {
        headers: {
          'X-API-Key': selectedTenant.api_key,
        },
      })
      if (!response.ok) {
        const errorInfo = await response.json()
        const info = getErrorInfo({ error: errorInfo.error })
        throw new Error(info.message)
      }
      const data = await response.json()
      setStats(data.stats)
    } catch (error) {
      logError(error, 'fetchStats')
      // Don't show toast for stats errors - they can fail silently
    }
  }

  const handleTenantSelect = (tenant: Tenant) => {
    setSelectedTenant(tenant)
    setConnectionStatus('connecting')
    // Simulate connection status
    setTimeout(() => setConnectionStatus('connected'), 1000)
  }

  const handleTenantCreated = (newTenant: Tenant) => {
    setTenants([...tenants, newTenant])
    setShowCreateTenant(false)
    toast.showSuccess('Tenant Created', `${newTenant.name} has been created successfully.`)
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <h1 className="text-2xl font-bold text-gray-900">Event Dashboard</h1>
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                connectionStatus === 'connected' 
                  ? 'bg-green-100 text-green-800' 
                  : connectionStatus === 'connecting'
                  ? 'bg-yellow-100 text-yellow-800'
                  : 'bg-red-100 text-red-800'
              }`}>
                {connectionStatus === 'connected' ? '● Connected' : 
                 connectionStatus === 'connecting' ? '● Connecting' : '● Disconnected'}
              </span>
            </div>
            <div className="flex items-center space-x-4">
              <button
                onClick={() => setShowIngestForm(true)}
                disabled={!selectedTenant}
                className={`inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white ${
                  selectedTenant 
                    ? 'bg-indigo-600 hover:bg-indigo-700'
                    : 'bg-gray-400 cursor-not-allowed'
                }`}
              >
                Ingest Event
              </button>
              <button
                onClick={() => setShowCreateTenant(true)}
                className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
              >
                New Tenant
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Tenant Selector */}
        <div className="mb-8">
          <TenantSelector
            tenants={tenants}
            selectedTenant={selectedTenant}
            onSelect={handleTenantSelect}
            isLoading={isLoading}
          />
        </div>

        {selectedTenant ? (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Event Feed */}
            <div className="lg:col-span-2">
              <EventFeed
                tenantId={selectedTenant.id}
                tenantApiKey={selectedTenant.api_key}
                tenantName={selectedTenant.name}
                connectionStatus={connectionStatus}
              />
            </div>

            {/* Stats Sidebar */}
            <div className="lg:col-span-1">
              <EventStats stats={stats} />
            </div>
          </div>
        ) : (
          /* Welcome Screen */
          <div className="text-center py-16">
            {isLoading ? (
              <div className="flex justify-center items-center">
                <div className="spinner w-10 h-10"></div>
                <span className="ml-3 text-gray-500">Loading tenants...</span>
              </div>
            ) : (
              <div className="max-w-md mx-auto">
                <svg className="mx-auto h-16 w-16 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                <h2 className="mt-4 text-xl font-medium text-gray-900">Select a Tenant</h2>
                <p className="mt-2 text-gray-500">Choose a tenant from the dropdown above to view their events in real-time.</p>
                <button
                  onClick={() => setShowCreateTenant(true)}
                  className="mt-6 inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
                >
                  Create Your First Tenant
                </button>
              </div>
            )}
          </div>
        )}
      </main>

      {/* Modals */}
      {showCreateTenant && (
        <CreateTenantForm
          onClose={() => setShowCreateTenant(false)}
          onCreated={handleTenantCreated}
        />
      )}

      {showIngestForm && selectedTenant && (
        <EventIngestForm
          tenantId={selectedTenant.id}
          tenantApiKey={selectedTenant.api_key}
          onClose={() => setShowIngestForm(false)}
          onEventCreated={() => fetchStats()}
        />
      )}
    </div>
  )
}

export default App
