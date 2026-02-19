"use client";

import { AuthGuard } from "@/components/auth/AuthGuard";
import { QueryProvider } from "@/providers/query-provider";
import { Sidebar } from "@/components/layout/Sidebar";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <QueryProvider>
      <AuthGuard>
        <div className="flex min-h-screen">
          <Sidebar />
          <main className="flex-1 overflow-auto p-6 md:p-8">{children}</main>
        </div>
      </AuthGuard>
    </QueryProvider>
  );
}
