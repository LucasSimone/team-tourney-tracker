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

  // Default to current origin's API endpoint with /api prefix
  // In development, this will be localhost:5173/api (routes to backend via docker-compose)
  // In production, this will be yourdomain.com/api (routes via Nginx reverse proxy)
  const protocol = typeof window !== 'undefined' ? window.location.protocol : 'http:'
  const host = typeof window !== 'undefined' ? window.location.host : 'localhost:5173'
  
  // Always use /api prefix for consistency between dev and prod
  // Dev: http://localhost:5173/api (docker-compose routes to backend)
  // Prod: https://yourdomain.com/api (Nginx reverse proxy routes to backend)
  return `${protocol}//${host}/api`
}

export const API_URL = getApiUrl()

export default API_URL
