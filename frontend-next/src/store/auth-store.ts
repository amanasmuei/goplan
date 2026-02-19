import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { User } from "@/lib/types";

interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  expiresAt: number | null;
}

interface AuthActions {
  setAuth: (token: string, user: User, expiresAt: number) => void;
  logout: () => void;
  isTokenExpired: () => boolean;
}

export const useAuthStore = create<AuthState & AuthActions>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      isAuthenticated: false,
      expiresAt: null,

      setAuth: (token, user, expiresAt) =>
        set({ token, user, isAuthenticated: true, expiresAt }),

      logout: () =>
        set({ token: null, user: null, isAuthenticated: false, expiresAt: null }),

      isTokenExpired: () => {
        const { expiresAt } = get();
        if (!expiresAt) return true;
        return Date.now() >= expiresAt * 1000;
      },
    }),
    {
      name: "goplan-auth",
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    },
  ),
);
