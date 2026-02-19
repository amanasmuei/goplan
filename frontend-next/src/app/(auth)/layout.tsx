"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth-store";

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const { isAuthenticated, isTokenExpired } = useAuthStore();

  useEffect(() => {
    if (isAuthenticated && !isTokenExpired()) {
      router.replace("/dashboard");
    }
  }, [isAuthenticated, isTokenExpired, router]);

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-navy-50 px-4">
      <div className="mb-8 text-center">
        <h1 className="text-3xl font-bold tracking-tight text-navy-900">
          GoPlan
        </h1>
        <p className="mt-1 text-sm text-navy-500">
          AI-Powered Strategy Consultant
        </p>
      </div>
      {children}
    </div>
  );
}
