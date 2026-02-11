import { useState, useEffect, useRef, useCallback } from 'react'

interface Event {
  id: number
  tenant_id: string
  event_type: string
  timestamp: string
  metadata: Record<string, unknown>
  created_at: string
}

interface EventFeedProps {
  tenantId: string
  tenantApiKey: string
  tenantName: string
  connectionStatus: 'connected' | 'disconnected' | 'connecting'
}

export default function EventFeed({ tenantId, tenantApiKey, tenantName, connectionStatus }: EventFeedProps) {
  const [events, setEvents] = useState<Event[]>([])
  const [filter, setFilter] = useState('')
  const [autoScroll, setAutoScroll] = useState(true)
  const eventsEndRef = useRef<HTMLDivElement>(null)
  const wsRef = useRef<WebSocket | null>(null)

  const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'
  const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/api/v1/ws'

  // WebSocket connection
  useEffect(() => {
    const connectWebSocket = () => {
      const wsUrl = `${WS_URL}?api_key=${tenantApiKey}`
      
      try {
        const ws = new WebSocket(wsUrl)
        wsRef.current = ws

        ws.onopen = () => {
          console.log('WebSocket connected')
        }

        ws.onmessage = (event) => {
          try {
            const newEvent = JSON.parse(event.data) as Event
            setEvents(prev => {
              // Check if event already exists to avoid duplicates
              if (prev.some(e => e.id === newEvent.id)) {
                return prev
              }
              return [newEvent, ...prev].slice(0, 100)
            })
          } catch (error) {
            console.error('Failed to parse event:', error)
          }
        }

        ws.onclose = () => {
          console.log('WebSocket disconnected, reconnecting...')
          setTimeout(connectWebSocket, 3000)
        }

        ws.onerror = (error) => {
          console.error('WebSocket error:', error)
        }
      } catch (error) {
        console.error('Failed to connect WebSocket:', error)
        setTimeout(connectWebSocket, 3000)
      }
    }

    connectWebSocket()

    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [tenantId, tenantApiKey])

  // Fetch historical events
  useEffect(() => {
    const fetchEvents = async () => {
      try {
        const response = await fetch(`${API_BASE}/events?limit=50`, {
          headers: {
            'X-API-Key': tenantApiKey,
          },
        })
        if (response.ok) {
          const data = await response.json()
          const newEvents = data.events || []
          setEvents(prev => {
            // Merge with existing events, avoiding duplicates by ID
            const existingIds = new Set(prev.map(e => e.id))
            const uniqueNewEvents = newEvents.filter((e: Event) => !existingIds.has(e.id))
            return [...uniqueNewEvents, ...prev].slice(0, 100)
          })
        }
      } catch (error) {
        console.error('Failed to fetch events:', error)
      }
    }

    fetchEvents()
  }, [tenantId, tenantApiKey, API_BASE])

  // Auto-scroll
  useEffect(() => {
    if (autoScroll && eventsEndRef.current) {
      eventsEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [events, autoScroll])

  const getEventColor = (eventType: string) => {
    const type = eventType.toLowerCase()
    if (type.includes('click')) return 'event-click'
    if (type.includes('view')) return 'event-view'
    if (type.includes('purchase')) return 'event-purchase'
    if (type.includes('error')) return 'event-error'
    if (type.includes('warning')) return 'event-warning'
    return 'event-info'
  }

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString()
  }

  const filteredEvents = events.filter(e => 
    e.event_type.toLowerCase().includes(filter.toLowerCase()) ||
    JSON.stringify(e.metadata).toLowerCase().includes(filter.toLowerCase())
  )

  return (
    <div className="bg-white shadow rounded-lg">
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-medium text-gray-900">Live Event Feed</h2>
            <p className="text-sm text-gray-500">Tenant: {tenantName}</p>
          </div>
          <div className="flex items-center space-x-4">
            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
              connectionStatus === 'connected'
                ? 'bg-green-100 text-green-800'
                : connectionStatus === 'connecting'
                ? 'bg-yellow-100 text-yellow-800'
                : 'bg-red-100 text-red-800'
            }`}>
              {connectionStatus === 'connected' ? '● Live' : 
               connectionStatus === 'connecting' ? '● Connecting' : '● Offline'}
            </span>
            <span className="text-sm text-gray-500">
              {filteredEvents.length} events
            </span>
          </div>
        </div>
        <div className="mt-4 flex items-center space-x-4">
          <input
            type="text"
            placeholder="Filter events..."
            value={filter}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFilter(e.target.value)}
            className="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
          <label className="flex items-center text-sm text-gray-600">
            <input
              type="checkbox"
              checked={autoScroll}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setAutoScroll(e.target.checked)}
              className="mr-2"
            />
            Auto-scroll
          </label>
        </div>
      </div>

      <div className="px-6 py-4 max-h-[600px] overflow-y-auto">
        {filteredEvents.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            No events yet. Ingest an event to see it here.
          </div>
        ) : (
          <div className="space-y-3">
            {filteredEvents.map((event) => (
              <div
                key={event.id}
                className={`p-4 rounded-lg border ${getEventColor(event.event_type)}`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center space-x-2">
                      <span className="font-medium">{event.event_type}</span>
                      <span className="text-xs opacity-75">
                        {formatTimestamp(event.timestamp)}
                      </span>
                    </div>
                    {Object.keys(event.metadata).length > 0 && (
                      <pre className="mt-2 text-xs overflow-x-auto">
                        {JSON.stringify(event.metadata, null, 2)}
                      </pre>
                    )}
                  </div>
                  <span className="text-xs font-mono opacity-75">
                    #{event.id}
                  </span>
                </div>
              </div>
            ))}
            <div ref={eventsEndRef} />
          </div>
        )}
      </div>
    </div>
  )
}
