"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, ChevronDown, ChevronRight, Clock, Lock } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import { formatDate } from "@/lib/utils";
import {
  SECTION_LABELS,
  SECTION_ORDER,
  type SectionType,
  type PlanVersion,
  type SectionVersion,
} from "@/lib/types";
import { useStrategy, usePlanVersions, useSectionVersions } from "@/hooks/useStrategy";
import { useSubscription } from "@/hooks/useSubscription";

function SectionVersionList({ planId, sectionType }: { planId: string; sectionType: SectionType }) {
  const { data: versions, isLoading } = useSectionVersions(planId, sectionType);

  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-20 w-full" />
        ))}
      </div>
    );
  }

  if (!versions || versions.length === 0) {
    return (
      <p className="py-8 text-center text-sm text-muted-foreground">
        No versions for this section yet.
      </p>
    );
  }

  const sorted = [...versions].sort((a, b) => b.version - a.version);

  return (
    <div className="space-y-3">
      {sorted.map((sv: SectionVersion) => (
        <Card key={sv.id} className="py-4">
          <CardContent className="px-4 py-0">
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <Badge variant="outline" className="text-xs">
                    v{sv.version}
                  </Badge>
                  <span className="text-xs text-muted-foreground">
                    {formatDate(sv.created_at)}
                  </span>
                </div>
                {sv.refinement_context && (
                  <p className="mt-2 text-xs text-muted-foreground italic">
                    Refinement: &quot;{sv.refinement_context}&quot;
                  </p>
                )}
                <p className="mt-2 text-sm text-muted-foreground line-clamp-3">
                  {contentPreview(sv.content)}
                </p>
              </div>
              <Badge variant="secondary" className="shrink-0 text-[10px]">
                {sv.generated_by}
              </Badge>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

function contentPreview(content: unknown): string {
  if (typeof content === "string") {
    return content.slice(0, 200);
  }
  if (content && typeof content === "object") {
    const str = JSON.stringify(content);
    // Try to extract a readable summary from the JSON
    const plain = str.replace(/[{}"[\]]/g, " ").replace(/\s+/g, " ").trim();
    return plain.slice(0, 200);
  }
  return "";
}

function TimelineEntry({ version }: { version: PlanVersion }) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="relative flex gap-4 pb-8 last:pb-0">
      {/* Timeline connector */}
      <div className="flex flex-col items-center">
        <div className="flex size-8 shrink-0 items-center justify-center rounded-full border-2 border-primary bg-background text-xs font-semibold text-primary">
          v{version.version}
        </div>
        <div className="w-px flex-1 bg-border" />
      </div>

      {/* Content */}
      <div className="flex-1 pb-2">
        <div className="flex items-center gap-3">
          <span className="text-sm font-medium">Version {version.version}</span>
          <span className="flex items-center gap-1 text-xs text-muted-foreground">
            <Clock className="size-3" />
            {formatDate(version.created_at)}
          </span>
        </div>

        {version.change_summary && (
          <p className="mt-1 text-sm text-muted-foreground">
            {version.change_summary}
          </p>
        )}

        <Button
          variant="ghost"
          size="sm"
          className="mt-2 h-7 gap-1 px-2 text-xs"
          onClick={() => setExpanded(!expanded)}
        >
          {expanded ? (
            <ChevronDown className="size-3" />
          ) : (
            <ChevronRight className="size-3" />
          )}
          {expanded ? "Hide snapshot" : "View snapshot"}
        </Button>

        {expanded && version.snapshot != null && (
          <Card className="mt-2 py-3">
            <CardContent className="px-4 py-0">
              <pre className="max-h-80 overflow-auto whitespace-pre-wrap text-xs text-muted-foreground">
                {typeof version.snapshot === "string"
                  ? version.snapshot
                  : JSON.stringify(version.snapshot, null, 2)}
              </pre>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}

function LoadingSkeleton() {
  return (
    <div className="space-y-6">
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="flex gap-4">
          <Skeleton className="size-8 shrink-0 rounded-full" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-4 w-48" />
            <Skeleton className="h-3 w-72" />
          </div>
        </div>
      ))}
    </div>
  );
}

export default function VersionHistoryPage() {
  const params = useParams<{ id: string }>();
  const planId = params.id;

  const { data: planData, isLoading: planLoading } = useStrategy(planId);
  const { data: versions, isLoading: versionsLoading } = usePlanVersions(planId);
  const { data: subscription } = useSubscription();

  const isPro = subscription?.tier === "pro" || subscription?.tier === "pro_plus";
  const isLoading = planLoading || versionsLoading;

  return (
    <div className="mx-auto max-w-4xl px-6 py-8">
      {/* Header */}
      <div className="mb-8">
        <Link
          href={`/strategies/${planId}`}
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          Back to strategy
        </Link>

        <h1 className="text-2xl font-bold tracking-tight">Version History</h1>
        {planData?.plan && (
          <p className="mt-1 text-sm text-muted-foreground">
            {planData.plan.title}
          </p>
        )}
      </div>

      {/* Pro gate */}
      {!isPro && !isLoading && (
        <Card className="py-6">
          <CardContent className="flex flex-col items-center gap-4 text-center">
            <div className="flex size-12 items-center justify-center rounded-full bg-muted">
              <Lock className="size-6 text-muted-foreground" />
            </div>
            <div>
              <h3 className="font-semibold">Version History is a Pro feature</h3>
              <p className="mt-1 text-sm text-muted-foreground">
                Upgrade to Pro to access full version history, track changes over
                time, and compare section revisions.
              </p>
            </div>
            <Button asChild>
              <Link href="/settings">Upgrade to Pro</Link>
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Content for Pro users */}
      {isPro && (
        <div className="space-y-10">
          {/* Plan-level timeline */}
          <section>
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Plan Versions</CardTitle>
              </CardHeader>
              <CardContent>
                {isLoading ? (
                  <LoadingSkeleton />
                ) : !versions || versions.length === 0 ? (
                  <p className="py-8 text-center text-sm text-muted-foreground">
                    No version history yet. Versions are created when sections are
                    regenerated or refined.
                  </p>
                ) : (
                  <div>
                    {[...versions]
                      .sort((a, b) => b.version - a.version)
                      .map((v: PlanVersion) => (
                        <TimelineEntry key={v.id} version={v} />
                      ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </section>

          <Separator />

          {/* Section-level versions */}
          <section>
            <h2 className="mb-4 text-lg font-semibold">Section Versions</h2>
            <Tabs defaultValue={SECTION_ORDER[0]}>
              <TabsList className="w-full flex-wrap">
                {SECTION_ORDER.map((st) => (
                  <TabsTrigger key={st} value={st} className="text-xs">
                    {SECTION_LABELS[st]}
                  </TabsTrigger>
                ))}
              </TabsList>
              {SECTION_ORDER.map((st) => (
                <TabsContent key={st} value={st} className="mt-4">
                  <SectionVersionList planId={planId} sectionType={st} />
                </TabsContent>
              ))}
            </Tabs>
          </section>
        </div>
      )}
    </div>
  );
}
