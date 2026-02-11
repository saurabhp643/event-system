import { useState } from 'react'

interface EventIngestFormProps {
  tenantId: string
  tenantApiKey: string
  onClose: () => void
  onEventCreated: () => void
}

export default function EventIngestForm({ tenantId, tenantApiKey, onClose, onEventCreated }: EventIngestFormProps) {
  const [eventType, setEventType] = useState('')
  const [metadata, setMetadata] = useState('{}')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const API_BASE = import.meta.env.VITE_API_URL || '/api/v1'

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!eventType.trim()) return

    setLoading(true)
    setError('')

    try {
      let parsedMetadata = {}
      try {
        parsedMetadata = JSON.parse(metadata)
      } catch {
        throw new Error('Invalid JSON in metadata field')
      }

      const response = await fetch(`${API_BASE}/events`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': tenantApiKey,
        },
        body: JSON.stringify({
          tenant_id: tenantId,
          event_type: eventType,
          timestamp: new Date().toISOString(),
          metadata: parsedMetadata,
        }),
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.message || 'Failed to ingest event')
      }

      onEventCreated()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Ingest Event</h3>
        </div>
        <form onSubmit={handleSubmit} className="px-6 py-4">
          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
              {error}
            </div>
          )}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Event Type
            </label>
            <input
              type="text"
              value={eventType}
              onChange={(e) => setEventType(e.target.value)}
              placeholder="e.g., page_view, click, purchase"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-indigo-500"
              required
            />
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Metadata (JSON)
            </label>
            <textarea
              value={metadata}
              onChange={(e) => setMetadata(e.target.value)}
              rows={5}
              className="w-full px-3 py-2 border border-gray-300 rounded-md font-mono text-sm focus:outline-none focus:ring-1 focus:ring-indigo-500"
              placeholder='{"key": "value"}'
            />
          </div>
          <div className="mt-6 flex justify-end space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 bg-indigo-600 rounded-md text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
            >
              {loading ? 'Ingesting...' : 'Ingest Event'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
