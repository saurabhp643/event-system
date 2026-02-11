/**
 * Error handling utilities for the frontend
 * Provides structured error handling, retry logic, and user-friendly messages
 */

// Error types
export interface ApiError {
  code: string
  message: string
  details?: string
}

export interface AppError {
  error: ApiError
}

export type ErrorSeverity = 'low' | 'medium' | 'high' | 'critical'

// Error category for better handling
export type ErrorCategory = 
  | 'network' 
  | 'authentication' 
  | 'authorization'
  | 'validation'
  | 'not_found'
  | 'rate_limit'
  | 'server'
  | 'client'
  | 'unknown'

// Map error codes to categories
const errorCategoryMap: Record<string, ErrorCategory> = {
  // Network errors
  'Failed to fetch': 'network',
  'NetworkError': 'network',
  'TypeError': 'network',
  
  // Authentication
  'unauthorized': 'authentication',
  'invalid_api_key': 'authentication',
  'expired_token': 'authentication',
  'missing_authentication': 'authentication',
  
  // Authorization
  'forbidden': 'authorization',
  
  // Validation
  'invalid_request': 'validation',
  'invalid_tenant_id': 'validation',
  'invalid_event_type': 'validation',
  'invalid_timestamp': 'validation',
  'invalid_metadata': 'validation',
  
  // Not found
  'tenant_not_found': 'not_found',
  'event_not_found': 'not_found',
  'not_found': 'not_found',
  
  // Rate limit
  'rate_limit_exceeded': 'rate_limit',
  'too_many_requests': 'rate_limit',
  
  // Server errors
  'internal_error': 'server',
  'database_error': 'server',
  'websocket_error': 'server',
}

// Map error codes to user-friendly messages
const errorMessages: Record<string, { title: string; message: string }> = {
  invalid_request: {
    title: 'Invalid Request',
    message: 'The data you provided is not valid. Please check and try again.',
  },
  invalid_tenant_id: {
    title: 'Invalid Tenant',
    message: 'The tenant ID is not valid. Please refresh and try again.',
  },
  invalid_event_type: {
    title: 'Invalid Event Type',
    message: 'Event type can only contain letters, numbers, underscores, hyphens, and dots.',
  },
  invalid_timestamp: {
    title: 'Invalid Timestamp',
    message: 'Timestamp must be in ISO8601 format (e.g., 2026-02-10T19:07:41Z).',
  },
  invalid_metadata: {
    title: 'Invalid Metadata',
    message: 'Metadata must be a valid JSON object.',
  },
  unauthorized: {
    title: 'Unauthorized',
    message: 'You are not authorized to perform this action. Please log in again.',
  },
  invalid_api_key: {
    title: 'Invalid API Key',
    message: 'Your API key is not valid. Please refresh the page or select a different tenant.',
  },
  expired_token: {
    title: 'Session Expired',
    message: 'Your session has expired. Please refresh the page to continue.',
  },
  missing_authentication: {
    title: 'Authentication Required',
    message: 'Please select a tenant to access this feature.',
  },
  tenant_not_found: {
    title: 'Tenant Not Found',
    message: 'The selected tenant no longer exists. Please select a different tenant.',
  },
  event_not_found: {
    title: 'Event Not Found',
    message: 'The requested event could not be found.',
  },
  tenant_exists: {
    title: 'Tenant Exists',
    message: 'A tenant with this name already exists. Please choose a different name.',
  },
  rate_limit_exceeded: {
    title: 'Too Many Requests',
    message: 'You are making requests too quickly. Please wait a moment before trying again.',
  },
  internal_error: {
    title: 'Server Error',
    message: 'Something went wrong on our end. Please try again later.',
  },
  database_error: {
    title: 'Database Error',
    message: 'Unable to save your data. Please try again.',
  },
  websocket_error: {
    title: 'Connection Error',
    message: 'Lost connection to the server. Attempting to reconnect...',
  },
}

/**
 * Categorize an error based on its code or message
 */
export function categorizeError(error: unknown): ErrorCategory {
  if (error instanceof TypeError) {
    return 'network'
  }
  
  if (error instanceof Error) {
    const message = error.message
    
    // Check if it's a network error
    if (message.includes('Failed to fetch') || message.includes('NetworkError')) {
      return 'network'
    }
    
    // Check error code map
    for (const [code, category] of Object.entries(errorCategoryMap)) {
      if (message.includes(code)) {
        return category
      }
    }
  }
  
  return 'unknown'
}

/**
 * Get user-friendly error information
 */
export function getErrorInfo(error: unknown): { 
  title: string
  message: string
  category: ErrorCategory
  severity: ErrorSeverity
  code?: string
  retryable: boolean
} {
  let title = 'Something went wrong'
  let message = 'An unexpected error occurred. Please try again.'
  let category: ErrorCategory = 'unknown'
  let severity: ErrorSeverity = 'medium'
  let code: string | undefined
  let retryable = true

  // Handle API errors
  if (isApiError(error)) {
    code = error.error.code
    category = categorizeErrorByCode(code)
    
    const info = errorMessages[code]
    if (info) {
      title = info.title
      message = info.message
    } else {
      message = error.error.message || message
    }
    
    // Set retryability based on error type
    retryable = canRetry(category)
    severity = getSeverity(category)
  } 
  // Handle network errors
  else if (error instanceof TypeError && error.message.includes('Failed to fetch')) {
    title = 'Connection Error'
    message = 'Unable to connect to the server. Please check your internet connection.'
    category = 'network'
    severity = 'high'
    retryable = true
  }
  // Handle other errors
  else if (error instanceof Error) {
    message = error.message || message
    category = categorizeError(error)
    severity = getSeverity(category)
    retryable = canRetry(category)
  }

  return { title, message, category, severity, code, retryable }
}

/**
 * Check if an error is an API error
 */
function isApiError(error: unknown): error is AppError {
  if (error && typeof error === 'object') {
    const e = error as Record<string, unknown>
    return (
      'error' in e &&
      e.error !== null &&
      typeof e.error === 'object' &&
      'code' in (e.error as Record<string, unknown>) &&
      'message' in (e.error as Record<string, unknown>)
    )
  }
  return false
}

/**
 * Categorize error by code
 */
function categorizeErrorByCode(code: string): ErrorCategory {
  return errorCategoryMap[code] || 'unknown'
}

/**
 * Determine if an error category is retryable
 */
function canRetry(category: ErrorCategory): boolean {
  switch (category) {
    case 'network':
    case 'server':
    case 'rate_limit':
      return true
    case 'authentication':
    case 'authorization':
    case 'validation':
    case 'not_found':
      return false
    default:
      return true
  }
}

/**
 * Get severity level for an error category
 */
function getSeverity(category: ErrorCategory): ErrorSeverity {
  switch (category) {
    case 'network':
      return 'high'
    case 'server':
      return 'high'
    case 'rate_limit':
      return 'medium'
    case 'authentication':
    case 'authorization':
      return 'high'
    case 'validation':
      return 'low'
    case 'not_found':
      return 'medium'
    default:
      return 'medium'
  }
}

/**
 * Create a retryable function wrapper
 */
export function withRetry<T>(
  fn: () => Promise<T>,
  options: {
    maxRetries?: number
    delay?: number
    onRetry?: (attempt: number, error: Error) => void
    onFinalError?: (error: Error) => void
  } = {}
): Promise<T> {
  const maxRetries = options.maxRetries ?? 3
  const delay = options.delay ?? 1000
  let attempt = 0

  const execute = async (): Promise<T> => {
    try {
      return await fn()
    } catch (error) {
      const info = getErrorInfo(error)
      
      if (attempt < maxRetries && info.retryable) {
        attempt++
        const waitTime = delay * Math.pow(2, attempt - 1) // Exponential backoff
        
        if (options.onRetry) {
          options.onRetry(attempt, error as Error)
        }
        
        await new Promise(resolve => setTimeout(resolve, waitTime))
        return execute()
      }
      
      if (options.onFinalError) {
        options.onFinalError(error as Error)
      }
      throw error
    }
  }

  return execute()
}

/**
 * Safe async function wrapper with error handling
 */
export async function safeFetch<T>(
  fetchFn: () => Promise<T>,
  options: {
    onError?: (error: unknown) => void
    onSuccess?: (data: T) => void
  } = {}
): Promise<{ data?: T; error?: unknown }> {
  try {
    const data = await fetchFn()
    if (options.onSuccess) {
      options.onSuccess(data)
    }
    return { data }
  } catch (error) {
    if (options.onError) {
      options.onError(error)
    }
    return { error }
  }
}

/**
 * Parse API error from response
 */
export function parseApiError(response: Response): Promise<AppError> {
  return response.json().then((data) => {
    if (data.error) {
      return data as AppError
    }
    return {
      error: {
        code: response.status.toString(),
        message: response.statusText,
      },
    }
  })
}

/**
 * Log error to console with appropriate level
 */
export function logError(error: unknown, context?: string): void {
  const info = getErrorInfo(error)
  
  switch (info.severity) {
    case 'critical':
    case 'high':
      console.error(`[ERROR${context ? ` - ${context}` : ''}]`, info.title, info.message, error)
      break
    case 'medium':
      console.warn(`[WARN${context ? ` - ${context}` : ''}]`, info.title, info.message)
      break
    default:
      console.log(`[INFO${context ? ` - ${context}` : ''}]`, info.message)
  }
}
