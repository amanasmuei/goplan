"use client";

import { useState } from "react";
import { Check, ChevronDown, X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import type { RecommendedApproachContent } from "@/lib/types";

interface RecommendedApproachProps {
  content: RecommendedApproachContent;
}

const LIKELIHOOD_COLORS: Record<string, string> = {
  high: "bg-red-100 text-red-800",
  medium: "bg-yellow-100 text-yellow-800",
  low: "bg-green-100 text-green-800",
};

function getLikelihoodColor(value: string): string {
  const lower = value.toLowerCase();
  return LIKELIHOOD_COLORS[lower] ?? "bg-gray-100 text-gray-800";
}

export function RecommendedApproach({ content }: RecommendedApproachProps) {
  const [expandedAlt, setExpandedAlt] = useState<number | null>(null);

  return (
    <div className="space-y-8">
      {/* Core Strategy */}
      <div className="rounded-lg border-l-4 border-brand bg-navy-50 p-5">
        <h4 className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Core Strategy
        </h4>
        <p className="text-sm font-medium leading-relaxed">
          {content.core_strategy}
        </p>
      </div>

      {/* Rationale */}
      {content.rationale && (
        <div className="space-y-2">
          <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Rationale
          </h4>
          <p className="text-sm leading-relaxed">{content.rationale}</p>
        </div>
      )}

      <Separator />

      {/* Key Pillars */}
      {content.key_pillars && content.key_pillars.length > 0 && (
        <div className="space-y-4">
          <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Key Pillars
          </h4>
          <div className="grid gap-4 sm:grid-cols-2">
            {content.key_pillars.map((pillar) => (
              <Card key={pillar.name} className="gap-3 py-4">
                <CardHeader className="pb-0">
                  <CardTitle className="text-sm">{pillar.name}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <p className="text-sm text-muted-foreground leading-relaxed">
                    {pillar.description}
                  </p>
                  {pillar.kpis && pillar.kpis.length > 0 && (
                    <div className="flex flex-wrap gap-1.5">
                      {pillar.kpis.map((kpi) => (
                        <Badge
                          key={kpi}
                          variant="outline"
                          className="text-[11px]"
                        >
                          {kpi}
                        </Badge>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* Alternatives Considered */}
      {content.alternatives_considered &&
        content.alternatives_considered.length > 0 && (
          <>
            <Separator />
            <div className="space-y-4">
              <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                Alternatives Considered
              </h4>
              <div className="space-y-3">
                {content.alternatives_considered.map((alt, i) => (
                  <div key={alt.name} className="rounded-lg border">
                    <button
                      type="button"
                      onClick={() =>
                        setExpandedAlt(expandedAlt === i ? null : i)
                      }
                      className="flex w-full items-center justify-between p-4 text-left"
                    >
                      <span className="text-sm font-medium">{alt.name}</span>
                      <ChevronDown
                        className={cn(
                          "size-4 text-muted-foreground transition-transform",
                          expandedAlt === i && "rotate-180",
                        )}
                      />
                    </button>
                    {expandedAlt === i && (
                      <div className="space-y-3 border-t px-4 pb-4 pt-3">
                        <div className="flex items-start gap-2 text-sm">
                          <Check className="mt-0.5 size-4 shrink-0 text-green-600" />
                          <span className="leading-relaxed">{alt.pros}</span>
                        </div>
                        <div className="flex items-start gap-2 text-sm">
                          <X className="mt-0.5 size-4 shrink-0 text-red-500" />
                          <span className="leading-relaxed">{alt.cons}</span>
                        </div>
                        <div className="rounded bg-muted/50 p-3">
                          <p className="text-xs font-medium text-muted-foreground">
                            Why not chosen
                          </p>
                          <p className="mt-1 text-sm leading-relaxed">
                            {alt.why_not_chosen}
                          </p>
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </>
        )}

      {/* Risk Mitigation */}
      {content.risk_mitigation && content.risk_mitigation.length > 0 && (
        <>
          <Separator />
          <div className="space-y-4">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Risk Mitigation
            </h4>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left">
                    <th className="pb-3 pr-4 font-medium text-muted-foreground">
                      Risk
                    </th>
                    <th className="pb-3 pr-4 font-medium text-muted-foreground">
                      Likelihood
                    </th>
                    <th className="pb-3 pr-4 font-medium text-muted-foreground">
                      Impact
                    </th>
                    <th className="pb-3 font-medium text-muted-foreground">
                      Mitigation
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {content.risk_mitigation.map((risk) => (
                    <tr key={risk.risk} className="border-b last:border-0">
                      <td className="py-3 pr-4 font-medium">{risk.risk}</td>
                      <td className="py-3 pr-4">
                        <Badge
                          variant="secondary"
                          className={cn(
                            "text-[11px]",
                            getLikelihoodColor(risk.likelihood),
                          )}
                        >
                          {risk.likelihood}
                        </Badge>
                      </td>
                      <td className="py-3 pr-4">
                        <Badge
                          variant="secondary"
                          className={cn(
                            "text-[11px]",
                            getLikelihoodColor(risk.impact),
                          )}
                        >
                          {risk.impact}
                        </Badge>
                      </td>
                      <td className="py-3 text-muted-foreground">
                        {risk.mitigation}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
