"use client";

import { useCallback } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth-store";
import { authApi } from "@/lib/api";

export function useAuth() {
  const router = useRouter();
  const { user, isAuthenticated, token, setAuth, logout: clearAuth, isTokenExpired } = useAuthStore();

  const login = useCallback(
    async (email: string, password: string) => {
      const res = await authApi.login({ email, password });
      setAuth(res.token, res.user, res.expires_at);
      return res.user;
    },
    [setAuth],
  );

  const register = useCallback(
    async (name: string, email: string, password: string) => {
      await authApi.register({ name, email, password });
      const res = await authApi.login({ email, password });
      setAuth(res.token, res.user, res.expires_at);
      return res.user;
    },
    [setAuth],
  );

  const logout = useCallback(() => {
    clearAuth();
    router.push("/login");
  }, [clearAuth, router]);

  return {
    user,
    token,
    isAuthenticated: isAuthenticated && !isTokenExpired(),
    login,
    register,
    logout,
  };
}
