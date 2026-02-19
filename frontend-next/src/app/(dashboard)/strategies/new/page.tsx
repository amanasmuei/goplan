"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Loader2, Lightbulb, AlertTriangle } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { useCreateStrategy, useStrategies } from "@/hooks/useStrategy";
import { useSubscription } from "@/hooks/useSubscription";
import { CATEGORY_LABELS, type PlanCategory } from "@/lib/types";

const CATEGORIES: Array<{ value: PlanCategory | "auto"; label: string }> = [
  { value: "auto", label: "Auto-detect" },
  ...Object.entries(CATEGORY_LABELS).map(([value, label]) => ({
    value: value as PlanCategory,
    label,
  })),
];

const MAX_LENGTH = 5000;
const MIN_LENGTH = 20;

export default function NewStrategyPage() {
  const router = useRouter();
  const [input, setInput] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<
    PlanCategory | "auto"
  >("auto");

  const createStrategy = useCreateStrategy();
  const { data: subscription } = useSubscription();
  const { data: strategiesData } = useStrategies({ page: 1, page_size: 1 });

  const totalPlans = strategiesData?.total ?? 0;
  const maxPlans = subscription?.max_plans ?? 3;
  const atLimit = maxPlans !== -1 && totalPlans >= maxPlans;

  const charCount = input.length;
  const isValid = charCount >= MIN_LENGTH && charCount <= MAX_LENGTH;
  const canSubmit = isValid && !createStrategy.isPending && !atLimit;

  async function handleSubmit() {
    if (!canSubmit) return;

    try {
      const result = await createStrategy.mutateAsync({
        input,
        category:
          selectedCategory === "auto" ? undefined : selectedCategory,
      });
      router.push(`/strategies/${result.plan.id}`);
    } catch {
      toast.error("Failed to generate strategy. Please try again.");
    }
  }

  return (
    <div className="mx-auto max-w-2xl py-8">
      <div className="space-y-2 mb-8">
        <h1 className="text-2xl font-bold tracking-tight">
          What&apos;s your strategic goal?
        </h1>
        <p className="text-muted-foreground">
          Describe your idea, project, or initiative. Be as specific as
          possible for better results.
        </p>
      </div>

      <div className="space-y-6">
        {/* Textarea */}
        <div className="space-y-2">
          <Textarea
            value={input}
            onChange={(e) =>
              setInput(e.target.value.slice(0, MAX_LENGTH))
            }
            placeholder="e.g., I want to launch a specialty coffee shop in downtown Austin targeting remote workers..."
            rows={6}
            className="min-h-[160px] resize-y text-base leading-relaxed"
            disabled={createStrategy.isPending}
          />
          <p
            className={`text-sm text-right ${
              charCount > 0 && charCount < MIN_LENGTH
                ? "text-red-500"
                : "text-muted-foreground"
            }`}
          >
            {charCount.toLocaleString()} / {MAX_LENGTH.toLocaleString()}
          </p>
        </div>

        {/* Category selector */}
        <div className="space-y-3">
          <p className="text-sm font-medium">Category</p>
          <div className="flex flex-wrap gap-2">
            {CATEGORIES.map((cat) => (
              <button
                key={cat.value}
                type="button"
                onClick={() => setSelectedCategory(cat.value)}
                disabled={createStrategy.isPending}
              >
                <Badge
                  variant={
                    selectedCategory === cat.value ? "default" : "outline"
                  }
                  className="cursor-pointer px-3 py-1 text-sm"
                >
                  {cat.label}
                </Badge>
              </button>
            ))}
          </div>
        </div>

        {/* Free tier warning */}
        {atLimit && (
          <div className="flex items-center gap-3 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
            <AlertTriangle className="size-4 shrink-0" />
            <span>
              You&apos;ve reached your free plan limit ({maxPlans} plans).{" "}
              <a
                href="/settings/billing"
                className="font-medium underline underline-offset-2 hover:text-amber-900"
              >
                Upgrade to Pro
              </a>{" "}
              for unlimited strategies.
            </span>
          </div>
        )}

        {/* Generate button */}
        <Button
          onClick={handleSubmit}
          disabled={!canSubmit}
          className="w-full"
          size="lg"
        >
          {createStrategy.isPending ? (
            <>
              <Loader2 className="animate-spin" />
              Generating your strategic plan...
            </>
          ) : (
            "Generate Strategy"
          )}
        </Button>

        {/* Tips */}
        <div className="rounded-lg border bg-muted/40 px-5 py-4 space-y-2">
          <div className="flex items-center gap-2 text-sm font-medium">
            <Lightbulb className="size-4 text-muted-foreground" />
            Tips for better results
          </div>
          <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1 pl-6">
            <li>Be specific about your industry and target audience</li>
            <li>Include budget constraints or timeline if relevant</li>
            <li>Mention any existing resources or constraints</li>
            <li>Describe the desired outcome or success metrics</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
