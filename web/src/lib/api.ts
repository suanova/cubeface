import axios from 'axios'

import { basePath } from '@/lib/base-path'
import { useAuthStore } from '@/stores/auth-store'

export interface ApiEnvelope<T> {
  success: boolean
  message: string
  code?: string
  data?: T
}

export const api = axios.create({
  baseURL: `${basePath}/api`,
  withCredentials: true,
  headers: { 'Content-Type': 'application/json' },
  timeout: 20_000,
})

api.interceptors.response.use(
  (response) => response,
  (error: unknown) => {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      useAuthStore.getState().clear()
    }
    return Promise.reject(error)
  }
)

// apiErrorMessage extracts the server's user-facing message from a failed
// request, or returns null when the caller should show its own fallback.
export function apiErrorMessage(error: unknown): string | null {
  if (axios.isAxiosError(error)) {
    const payload = error.response?.data as { message?: unknown } | undefined
    if (payload && typeof payload.message === 'string' && payload.message) {
      return payload.message
    }
    if (!error.response) {
      return 'Network error. Please check your connection and try again.'
    }
  }
  return null
}
