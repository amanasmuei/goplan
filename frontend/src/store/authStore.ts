import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '../types'

interface AuthState {
  token: string | null
  expiresAt: string | null
  user: User | null
  isAuthenticated: boolean
  login: (token: string, user: User, expiresAt?: string) => void
  logout: () => void
  isTokenExpired: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      expiresAt: null,
      user: null,
      isAuthenticated: false,
      login: (token: string, user: User, expiresAt?: string) =>
        set({ token, user, isAuthenticated: true, expiresAt: expiresAt || null }),
      logout: () =>
        set({ token: null, user: null, isAuthenticated: false, expiresAt: null }),
      isTokenExpired: () => {
        const { expiresAt } = get()
        if (!expiresAt) return false
        return new Date().getTime() > new Date(expiresAt).getTime()
      },
    }),
    {
      name: 'goplan-auth',
      onRehydrateStorage: () => (state) => {
        if (state && state.expiresAt) {
          const isExpired = new Date().getTime() > new Date(state.expiresAt).getTime()
          if (isExpired) {
            state.logout()
          }
        }
      },
    }
  )
)
