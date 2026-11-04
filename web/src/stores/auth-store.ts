import { create } from 'zustand'

import type { AuthUser } from '@/features/auth/types'

interface AuthState {
  user: AuthUser | null
  initialized: boolean
  setUser: (user: AuthUser | null) => void
  setInitialized: (initialized: boolean) => void
  clear: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  initialized: false,
  setUser: (user) => set({ user }),
  setInitialized: (initialized) => set({ initialized }),
  clear: () => set({ user: null, initialized: true }),
}))
