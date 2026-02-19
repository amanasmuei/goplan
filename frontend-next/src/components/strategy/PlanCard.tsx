"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Trash2 } from "lucide-react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardAction,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn, formatDate, truncate } from "@/lib/utils";
import { CATEGORY_COLORS, STATUS_COLORS } from "@/lib/constants";
import { CATEGORY_LABELS } from "@/lib/types";
import { useArchiveStrategy } from "@/hooks/useStrategy";
import type { StrategicPlan } from "@/lib/types";

interface PlanCardProps {
  plan: StrategicPlan;
}

export function PlanCard({ plan }: PlanCardProps) {
  const router = useRouter();
  const [showConfirm, setShowConfirm] = useState(false);
  const archiveMutation = useArchiveStrategy();

  const handleArchive = (e: React.MouseEvent) => {
    e.stopPropagation();
    setShowConfirm(true);
  };

  const confirmArchive = () => {
    archiveMutation.mutate(plan.id, {
      onSettled: () => setShowConfirm(false),
    });
  };

  const isGenerating = plan.status === "generating";

  return (
    <>
      <Card
        className={cn(
          "group relative cursor-pointer transition-all hover:shadow-md hover:scale-[1.01]",
          isGenerating && "border-blue-200",
        )}
        onClick={() => router.push(`/strategies/${plan.id}`)}
      >
        <CardHeader className="pb-0">
          <CardTitle className="line-clamp-2 text-base">
            {plan.title}
          </CardTitle>
          <CardAction>
            <Button
              variant="ghost"
              size="icon-xs"
              className="opacity-0 transition-opacity group-hover:opacity-100 text-muted-foreground hover:text-destructive"
              onClick={handleArchive}
            >
              <Trash2 />
            </Button>
          </CardAction>
        </CardHeader>

        <CardContent className="space-y-3 pt-0">
          <div className="flex flex-wrap items-center gap-1.5">
            <Badge
              variant="secondary"
              className={cn("text-[11px]", CATEGORY_COLORS[plan.category])}
            >
              {CATEGORY_LABELS[plan.category]}
            </Badge>
            <Badge
              variant="secondary"
              className={cn(
                "text-[11px]",
                STATUS_COLORS[plan.status],
                isGenerating && "animate-pulse",
              )}
            >
              {plan.status}
            </Badge>
          </div>

          {plan.original_input && (
            <p className="text-sm text-muted-foreground leading-relaxed">
              {truncate(plan.original_input, 100)}
            </p>
          )}

          <p className="text-xs text-muted-foreground/70">
            {formatDate(plan.created_at)}
          </p>
        </CardContent>
      </Card>

      <Dialog open={showConfirm} onOpenChange={setShowConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Archive this plan?</DialogTitle>
            <DialogDescription>
              &quot;{truncate(plan.title, 50)}&quot; will be archived and hidden
              from your dashboard.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowConfirm(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmArchive}
              disabled={archiveMutation.isPending}
            >
              {archiveMutation.isPending ? "Archiving..." : "Archive"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
