"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  Download,
  History,
  Lock,
  Pencil,
  RefreshCw,
  Trash2,
} from "lucide-react";
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";
import { cn, formatDate } from "@/lib/utils";
import { CATEGORY_COLORS } from "@/lib/constants";
import {
  SECTION_LABELS,
  CATEGORY_LABELS,
  type SectionType,
  type StrategicPlan,
} from "@/lib/types";
import {
  useArchiveStrategy,
  useExportStrategy,
  useRefineSection,
  useRegenerateSection,
  useSimilarStrategies,
} from "@/hooks/useStrategy";
import { useSubscription } from "@/hooks/useSubscription";

interface ActionPanelProps {
  plan: StrategicPlan;
  activeSection: SectionType;
}

export function ActionPanel({ plan, activeSection }: ActionPanelProps) {
  const router = useRouter();
  const { data: subscription } = useSubscription();
  const { data: similarPlans } = useSimilarStrategies(plan.id);
  const regenerateMutation = useRegenerateSection(plan.id);
  const refineMutation = useRefineSection(plan.id);
  const archiveMutation = useArchiveStrategy();
  const exportStrategy = useExportStrategy();

  const [showRegenerate, setShowRegenerate] = useState(false);
  const [showRefine, setShowRefine] = useState(false);
  const [showArchive, setShowArchive] = useState(false);
  const [regenerateContext, setRegenerateContext] = useState("");
  const [refinePrompt, setRefinePrompt] = useState("");

  const isPro = subscription?.tier === "pro" || subscription?.tier === "pro_plus";

  const handleRegenerate = () => {
    regenerateMutation.mutate(
      {
        sectionType: activeSection,
        data: regenerateContext
          ? { additional_context: regenerateContext }
          : undefined,
      },
      {
        onSuccess: () => {
          setShowRegenerate(false);
          setRegenerateContext("");
        },
      },
    );
  };

  const handleRefine = () => {
    refineMutation.mutate(
      {
        sectionType: activeSection,
        data: { refinement_prompt: refinePrompt },
      },
      {
        onSuccess: () => {
          setShowRefine(false);
          setRefinePrompt("");
        },
      },
    );
  };

  const handleArchive = () => {
    archiveMutation.mutate(plan.id, {
      onSuccess: () => {
        setShowArchive(false);
        router.push("/dashboard");
      },
    });
  };

  const handleExport = async (format: "markdown") => {
    await exportStrategy(plan.id, format);
  };

  return (
    <>
      <aside className="sticky top-0 flex h-full w-80 shrink-0 flex-col border-l bg-gray-50 p-4">
        {/* Plan Info */}
        <div className="space-y-3">
          <h3 className="line-clamp-2 text-sm font-semibold">{plan.title}</h3>
          <Badge
            variant="secondary"
            className={cn("text-[11px]", CATEGORY_COLORS[plan.category])}
          >
            {CATEGORY_LABELS[plan.category]}
          </Badge>
          <div className="flex items-center gap-4 text-xs text-muted-foreground">
            <span>Created {formatDate(plan.created_at)}</span>
            <span>Version {plan.current_version}</span>
          </div>
        </div>

        <Separator className="my-4" />

        {/* Regenerate Section */}
        <Button
          variant="outline"
          className="w-full justify-start gap-2"
          disabled={!isPro}
          onClick={() => isPro && setShowRegenerate(true)}
        >
          {isPro ? (
            <RefreshCw className="size-4" />
          ) : (
            <Lock className="size-4" />
          )}
          Regenerate Section
          {!isPro && (
            <span className="ml-auto text-[10px] text-muted-foreground">
              Pro
            </span>
          )}
        </Button>

        {/* Refine Section */}
        <Button
          variant="outline"
          className="mt-2 w-full justify-start gap-2"
          disabled={!isPro}
          onClick={() => isPro && setShowRefine(true)}
        >
          {isPro ? (
            <Pencil className="size-4" />
          ) : (
            <Lock className="size-4" />
          )}
          Refine Section
          {!isPro && (
            <span className="ml-auto text-[10px] text-muted-foreground">
              Pro
            </span>
          )}
        </Button>

        <Separator className="my-4" />

        {/* Version History */}
        <Button
          variant="ghost"
          className="w-full justify-start gap-2"
          disabled={!isPro}
          onClick={() =>
            isPro && router.push(`/strategies/${plan.id}/versions`)
          }
        >
          {isPro ? (
            <History className="size-4" />
          ) : (
            <Lock className="size-4" />
          )}
          Version History
          {!isPro && (
            <span className="ml-auto text-[10px] text-muted-foreground">
              Pro
            </span>
          )}
        </Button>

        {/* Export */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              className="w-full justify-start gap-2"
              disabled={!isPro}
            >
              {isPro ? (
                <Download className="size-4" />
              ) : (
                <Lock className="size-4" />
              )}
              Export
              {!isPro && (
                <span className="ml-auto text-[10px] text-muted-foreground">
                  Pro
                </span>
              )}
            </Button>
          </DropdownMenuTrigger>
          {isPro && (
            <DropdownMenuContent align="start">
              <DropdownMenuItem onClick={() => handleExport("markdown")}>
                Markdown (.md)
              </DropdownMenuItem>
            </DropdownMenuContent>
          )}
        </DropdownMenu>

        <Separator className="my-4" />

        {/* Similar Strategies */}
        <div className="space-y-3">
          <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Similar Strategies
          </h4>
          {similarPlans && similarPlans.length > 0 ? (
            <div className="space-y-2">
              {similarPlans.slice(0, 3).map((similar) => (
                <button
                  key={similar.id}
                  type="button"
                  onClick={() => router.push(`/strategies/${similar.id}`)}
                  className="w-full rounded-lg border bg-white p-3 text-left transition-colors hover:bg-gray-50"
                >
                  <p className="line-clamp-1 text-sm font-medium">
                    {similar.title}
                  </p>
                  <Badge
                    variant="secondary"
                    className={cn(
                      "mt-1.5 text-[10px]",
                      CATEGORY_COLORS[similar.category],
                    )}
                  >
                    {CATEGORY_LABELS[similar.category]}
                  </Badge>
                </button>
              ))}
            </div>
          ) : (
            <p className="text-xs text-muted-foreground">
              No similar strategies found.
            </p>
          )}
        </div>

        {/* Spacer */}
        <div className="flex-1" />

        <Separator className="my-4" />

        {/* Archive */}
        <Button
          variant="ghost"
          size="sm"
          className="w-full justify-start gap-2 text-destructive hover:text-destructive"
          onClick={() => setShowArchive(true)}
        >
          <Trash2 className="size-4" />
          Archive Plan
        </Button>
      </aside>

      {/* Regenerate Dialog */}
      <Dialog open={showRegenerate} onOpenChange={setShowRegenerate}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Regenerate Section</DialogTitle>
            <DialogDescription>
              Regenerate &quot;{SECTION_LABELS[activeSection]}&quot; with AI.
              Optionally add context to guide the output.
            </DialogDescription>
          </DialogHeader>
          <Textarea
            placeholder="Additional context (optional)..."
            value={regenerateContext}
            onChange={(e) => setRegenerateContext(e.target.value)}
            rows={4}
          />
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowRegenerate(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleRegenerate}
              disabled={regenerateMutation.isPending}
            >
              {regenerateMutation.isPending ? "Regenerating..." : "Regenerate"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Refine Dialog */}
      <Dialog open={showRefine} onOpenChange={setShowRefine}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Refine Section</DialogTitle>
            <DialogDescription>
              Describe what you&apos;d like to change in &quot;
              {SECTION_LABELS[activeSection]}&quot;.
            </DialogDescription>
          </DialogHeader>
          <Textarea
            placeholder="What would you like to change?"
            value={refinePrompt}
            onChange={(e) => setRefinePrompt(e.target.value)}
            rows={4}
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowRefine(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleRefine}
              disabled={refineMutation.isPending || !refinePrompt.trim()}
            >
              {refineMutation.isPending ? "Refining..." : "Refine"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Archive Dialog */}
      <Dialog open={showArchive} onOpenChange={setShowArchive}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Archive this plan?</DialogTitle>
            <DialogDescription>
              This plan will be archived and hidden from your dashboard.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowArchive(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleArchive}
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
