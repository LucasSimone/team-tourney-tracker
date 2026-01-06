/**
 * API Configuration
 * Dynamically determines the API URL based on environment
 */

interface ImportMetaEnv {
  readonly VITE_API_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

function getApiUrl(): string {
  // Try to use environment variable first (populated during build)
  const envUrl = (import.meta as any).env?.VITE_API_URL
  if (envUrl) {
    return envUrl
  }

  // Try window config (for runtime configuration)
  if (typeof window !== 'undefined' && (window as any).__API_URL__) {
    return (window as any).__API_URL__
  }

  // Default to current origin's API endpoint
  // In development, this will be localhost:8080
  // In production, this will match the frontend's domain
  const protocol = typeof window !== 'undefined' ? window.location.protocol : 'http:'
  const host = typeof window !== 'undefined' ? window.location.host : 'localhost:8080'
  
  // If running on port 5173 (dev), assume backend is on 8080
  if (host.includes('5173')) {
    return 'http://localhost:8080'
  }
  
  // Otherwise, assume backend is on same host as frontend
  // with /api prefix (Nginx routes /api/* to backend)
  return `${protocol}//${host}/api`
}

export const API_URL = getApiUrl()

export default API_URL
