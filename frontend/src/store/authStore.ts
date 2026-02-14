import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '../types'

interface AuthState {
  // In-memory only (not persisted) - prevents XSS from stealing access token
  token: string | null
  expiresAt: string | null
  // Persisted to localStorage
  refreshToken: string | null
  user: User | null
  isAuthenticated: boolean
  login: (token: string, user: User, expiresAt?: string, refreshToken?: string) => void
  logout: () => void
  setToken: (token: string, expiresAt?: string) => void
  isTokenExpired: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      expiresAt: null,
      refreshToken: null,
      user: null,
      isAuthenticated: false,
      login: (token: string, user: User, expiresAt?: string, refreshToken?: string) =>
        set({
          token,
          user,
          isAuthenticated: true,
          expiresAt: expiresAt || null,
          refreshToken: refreshToken || token,
        }),
      logout: () =>
        set({
          token: null,
          user: null,
          isAuthenticated: false,
          expiresAt: null,
          refreshToken: null,
        }),
      setToken: (token: string, expiresAt?: string) =>
        set({
          token,
          expiresAt: expiresAt || null,
        }),
      isTokenExpired: () => {
        const { expiresAt } = get()
        if (!expiresAt) return false
        return new Date().getTime() > new Date(expiresAt).getTime()
      },
    }),
    {
      name: 'goplan-auth',
      // Only persist refreshToken, user, and isAuthenticated to localStorage.
      // The access token (token, expiresAt) stays in-memory only,
      // preventing XSS attacks from stealing the active session token.
      partialize: (state) => ({
        refreshToken: state.refreshToken,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        if (state) {
          // After rehydration from localStorage, token and expiresAt will be null
          // (since they are not persisted). The token refresh interceptor in api.ts
          // will handle obtaining a new access token using the persisted refreshToken.
          if (!state.refreshToken) {
            state.logout()
          }
        }
      },
    }
  )
)
