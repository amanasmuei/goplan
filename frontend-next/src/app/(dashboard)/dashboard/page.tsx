"use client";

import Link from "next/link";
import { PlusCircle, FileText, RefreshCw, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { PlanCard } from "@/components/strategy/PlanCard";
import { useStrategies } from "@/hooks/useStrategy";
import { useSubscription } from "@/hooks/useSubscription";

export default function DashboardPage() {
  const { data, isLoading, isError, refetch } = useStrategies();
  const { data: subscription } = useSubscription();

  const plans = data?.plans ?? [];
  const total = data?.total ?? 0;
  const isFree = !subscription || subscription.tier === "free";
  const maxPlans = subscription?.max_plans ?? 3;
  const atLimit = isFree && total >= maxPlans;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            My Strategic Plans
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {isFree
              ? `${total} of ${maxPlans} plans used`
              : `${total} plan${total !== 1 ? "s" : ""}`}
          </p>
        </div>
        <Button asChild>
          <Link href="/strategies/new">
            <PlusCircle />
            New Strategy
          </Link>
        </Button>
      </div>

      {/* Plan limit warning */}
      {atLimit && (
        <div className="flex items-center gap-3 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          <AlertTriangle className="size-4 shrink-0" />
          <span>
            You&apos;ve reached the free plan limit.{" "}
            <Link
              href="/settings/billing"
              className="font-medium underline underline-offset-2 hover:text-amber-900"
            >
              Upgrade to Pro
            </Link>{" "}
            for more plans.
          </span>
        </div>
      )}

      {/* Loading */}
      {isLoading && (
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="space-y-3 rounded-xl border p-6">
              <Skeleton className="h-5 w-3/4" />
              <div className="flex gap-2">
                <Skeleton className="h-5 w-16 rounded-full" />
                <Skeleton className="h-5 w-14 rounded-full" />
              </div>
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-2/3" />
              <Skeleton className="h-3 w-24" />
            </div>
          ))}
        </div>
      )}

      {/* Error */}
      {isError && (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <p className="text-muted-foreground mb-4">
            Failed to load your plans. Please try again.
          </p>
          <Button variant="outline" onClick={() => refetch()}>
            <RefreshCw />
            Retry
          </Button>
        </div>
      )}

      {/* Empty state */}
      {!isLoading && !isError && plans.length === 0 && (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="mb-4 flex size-16 items-center justify-center rounded-full bg-muted">
            <FileText className="size-8 text-muted-foreground" />
          </div>
          <h2 className="text-lg font-semibold">No strategic plans yet</h2>
          <p className="mt-1 max-w-sm text-sm text-muted-foreground">
            Describe your idea or goal and let AI craft a comprehensive strategy
            for you.
          </p>
          <Button asChild className="mt-6">
            <Link href="/strategies/new">
              <PlusCircle />
              Create your first strategy
            </Link>
          </Button>
        </div>
      )}

      {/* Plan grid */}
      {!isLoading && !isError && plans.length > 0 && (
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          {plans.map((plan) => (
            <PlanCard key={plan.id} plan={plan} />
          ))}
        </div>
      )}
    </div>
  );
}
