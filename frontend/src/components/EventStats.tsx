interface EventStatsProps {
  stats: Record<string, number> | null
}

export default function EventStats({ stats }: EventStatsProps) {
  if (!stats || Object.keys(stats).length === 0) {
    return (
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-lg font-medium text-gray-900 mb-4">Event Statistics</h2>
        <p className="text-gray-500">No statistics available yet.</p>
      </div>
    )
  }

  const total = stats.total || 0
  const eventTypes = Object.entries(stats).filter(([key]) => key !== 'total')

  return (
    <div className="bg-white shadow rounded-lg">
      <div className="px-6 py-4 border-b border-gray-200">
        <h2 className="text-lg font-medium text-gray-900">Event Statistics</h2>
      </div>
      <div className="px-6 py-4">
        <div className="mb-6">
          <p className="text-sm text-gray-500">Total Events</p>
          <p className="text-3xl font-bold text-gray-900">{total.toLocaleString()}</p>
        </div>
        <div className="space-y-4">
          <h3 className="text-sm font-medium text-gray-700">By Event Type</h3>
          {eventTypes.map(([type, count]) => (
            <div key={type}>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm text-gray-600">{type}</span>
                <span className="text-sm font-medium text-gray-900">{count}</span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-indigo-600 h-2 rounded-full"
                  style={{ width: `${total > 0 ? (count / total) * 100 : 0}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
