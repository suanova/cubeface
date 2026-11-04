import axios from 'axios'

import { api, type ApiEnvelope } from '@/lib/api'
import type {
  AuthUser,
  LoginInput,
  RegisterInput,
  StatusInfo,
} from './types'

export async function getSelf(): Promise<AuthUser | null> {
  try {
    const { data } = await api.get<ApiEnvelope<{ user: AuthUser }>>('/user/self')
    return data.data?.user ?? null
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return null
    }
    throw error
  }
}

export async function loginRequest(
  input: LoginInput
): Promise<ApiEnvelope<{ user: AuthUser; expires_at: string }>> {
  const { data } = await api.post<ApiEnvelope<{ user: AuthUser; expires_at: string }>>(
    '/user/login',
    input
  )
  return data
}

export async function registerRequest(
  input: RegisterInput
): Promise<ApiEnvelope<{ user: AuthUser }>> {
  const { data } = await api.post<ApiEnvelope<{ user: AuthUser }>>('/user/register', input)
  return data
}

export async function logoutRequest(): Promise<void> {
  await api.post('/user/auth/logout')
}

export async function sendVerificationCode(email: string): Promise<ApiEnvelope<{ sent: boolean }>> {
  const { data } = await api.post<ApiEnvelope<{ sent: boolean }>>('/verification/send', {
    email,
    purpose: 'register',
  })
  return data
}

export async function getStatus(): Promise<StatusInfo> {
  const { data } = await api.get<ApiEnvelope<StatusInfo>>('/status')
  if (!data.success || !data.data) {
    throw new Error(data.message || 'Failed to load server status')
  }
  return data.data
}
