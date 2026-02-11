interface Tenant {
  id: string
  name: string
  api_key: string
  active: boolean
}

interface TenantSelectorProps {
  tenants: Tenant[]
  selectedTenant: Tenant | null
  onSelect: (tenant: Tenant) => void
  isLoading?: boolean
}

export default function TenantSelector({ tenants, selectedTenant, onSelect, isLoading = false }: TenantSelectorProps) {
  return (
    <div className="bg-white shadow rounded-lg p-6">
      <label className="block text-sm font-medium text-gray-700 mb-2">
        Select Tenant
      </label>
      <div className="relative">
        <select
          value={selectedTenant?.id || ''}
          onChange={(e) => {
            const tenant = tenants.find(t => t.id === e.target.value)
            if (tenant) onSelect(tenant)
          }}
          disabled={isLoading}
          className={`block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md border ${
            isLoading ? 'bg-gray-100 cursor-not-allowed' : ''
          }`}
        >
          <option value="">-- Select a tenant --</option>
          {tenants.map((tenant) => (
            <option key={tenant.id} value={tenant.id} disabled={!tenant.active}>
              {tenant.name} {!tenant.active && '(Inactive)'}
            </option>
          ))}
        </select>
        {isLoading && (
          <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
            <div className="spinner w-5 h-5"></div>
          </div>
        )}
      </div>
      {selectedTenant && (
        <div className="mt-4 p-4 bg-gray-50 rounded-md">
          <p className="text-sm text-gray-600">Selected tenant ID:</p>
          <p className="text-xs font-mono text-gray-800 break-all">{selectedTenant.id}</p>
        </div>
      )}
    </div>
  )
}
