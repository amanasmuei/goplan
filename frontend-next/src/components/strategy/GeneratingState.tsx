"use client";

import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";

export function GeneratingState() {
  return (
    <div className="flex h-[calc(100vh-2rem)]">
      {/* Left panel skeleton */}
      <div className="flex w-60 shrink-0 flex-col border-r bg-white p-5">
        <Skeleton className="mb-3 h-5 w-3/4" />
        <div className="mb-4 flex gap-2">
          <Skeleton className="h-5 w-16 rounded-full" />
          <Skeleton className="h-5 w-14 rounded-full" />
        </div>
        <Separator className="mb-4" />
        <div className="space-y-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-9 w-full rounded-lg" />
          ))}
        </div>
      </div>

      {/* Center panel skeleton */}
      <div className="flex-1 overflow-hidden">
        <div className="mx-auto max-w-4xl space-y-10 px-8 py-8">
          <div className="space-y-4">
            <p className="animate-pulse text-center text-sm font-medium text-brand">
              Generating your strategic plan...
            </p>
          </div>

          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="space-y-4">
              <Skeleton className="h-7 w-48" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-5/6" />
              <Skeleton className="h-4 w-4/6" />
              <div className="grid grid-cols-3 gap-4 pt-2">
                <Skeleton className="h-20 rounded-lg" />
                <Skeleton className="h-20 rounded-lg" />
                <Skeleton className="h-20 rounded-lg" />
              </div>
              {i < 2 && <Separator className="mt-8" />}
            </div>
          ))}
        </div>
      </div>

      {/* Right panel skeleton */}
      <div className="flex w-80 shrink-0 flex-col border-l bg-gray-50 p-4">
        <Skeleton className="mb-2 h-5 w-3/4" />
        <Skeleton className="mb-1 h-5 w-16 rounded-full" />
        <Skeleton className="mb-4 h-3 w-32" />
        <Separator className="mb-4" />
        <div className="space-y-2">
          <Skeleton className="h-9 w-full rounded-md" />
          <Skeleton className="h-9 w-full rounded-md" />
        </div>
        <Separator className="my-4" />
        <div className="space-y-2">
          <Skeleton className="h-9 w-full rounded-md" />
          <Skeleton className="h-9 w-full rounded-md" />
        </div>
      </div>
    </div>
  );
}
