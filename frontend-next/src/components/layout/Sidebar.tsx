"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  LayoutDashboard,
  PlusCircle,
  LogOut,
  Target,
} from "lucide-react";
import { useAuthStore } from "@/store/auth-store";
import { useSubscription } from "@/hooks/useSubscription";
import { cn, truncate } from "@/lib/utils";
import type { SubscriptionTier } from "@/lib/types";

const navItems = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/strategies/new", label: "New Strategy", icon: PlusCircle },
];

const tierLabels: Record<SubscriptionTier, string> = {
  free: "Free",
  pro: "Pro",
  pro_plus: "Pro+",
};

const tierColors: Record<SubscriptionTier, string> = {
  free: "bg-navy-700 text-navy-200",
  pro: "bg-brand/20 text-brand-light",
  pro_plus: "bg-amber-500/20 text-amber-300",
};

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout } = useAuthStore();
  const { data: subscription } = useSubscription();

  const tier = subscription?.tier ?? "free";

  const handleLogout = () => {
    logout();
    router.push("/login");
  };

  return (
    <aside className="sticky top-0 flex h-screen w-64 flex-col bg-navy-950 text-white">
      {/* Logo */}
      <div className="flex items-center gap-2 px-6 py-5">
        <Target className="size-6 text-brand-light" />
        <span className="text-xl font-bold tracking-tight">GoPlan</span>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 px-3 py-4">
        {navItems.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
                isActive
                  ? "border-l-2 border-brand bg-navy-800 text-white"
                  : "text-navy-300 hover:bg-navy-800/60 hover:text-white",
              )}
            >
              <item.icon className="size-5 shrink-0" />
              {item.label}
            </Link>
          );
        })}
      </nav>

      {/* Bottom section */}
      <div className="border-t border-navy-800 px-4 py-4 space-y-3">
        {/* Subscription badge */}
        <div className="flex items-center justify-between">
          <span
            className={cn(
              "rounded-full px-2.5 py-0.5 text-xs font-medium",
              tierColors[tier],
            )}
          >
            {tierLabels[tier]}
          </span>
        </div>

        {/* User info */}
        {user && (
          <div className="flex items-center gap-3">
            <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-brand font-semibold text-sm text-white">
              {user.name.charAt(0).toUpperCase()}
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-white truncate">
                {user.name}
              </p>
              <p className="text-xs text-navy-400 truncate">
                {truncate(user.email, 22)}
              </p>
            </div>
          </div>
        )}

        {/* Logout */}
        <button
          onClick={handleLogout}
          className="flex w-full items-center gap-2 rounded-lg px-2 py-2 text-sm text-navy-400 transition-colors hover:bg-navy-800/60 hover:text-white"
        >
          <LogOut className="size-4" />
          Log out
        </button>
      </div>
    </aside>
  );
}
