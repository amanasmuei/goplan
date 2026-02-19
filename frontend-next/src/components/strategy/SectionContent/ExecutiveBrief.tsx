"use client";

import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import type { ExecutiveBriefContent } from "@/lib/types";

interface ExecutiveBriefProps {
  content: ExecutiveBriefContent;
}

export function ExecutiveBrief({ content }: ExecutiveBriefProps) {
  return (
    <div className="space-y-6">
      {/* Summary */}
      <p className="text-lg font-medium leading-relaxed text-foreground/90">
        {content.summary}
      </p>

      <Separator />

      {/* Objective, Scope, Expected Outcome */}
      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        <div className="space-y-2">
          <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Objective
          </h4>
          <p className="text-sm leading-relaxed">{content.objective}</p>
        </div>
        <div className="space-y-2">
          <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Scope
          </h4>
          <p className="text-sm leading-relaxed">{content.scope}</p>
        </div>
        <div className="space-y-2">
          <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Expected Outcome
          </h4>
          <p className="text-sm leading-relaxed">{content.expected_outcome}</p>
        </div>
      </div>

      {/* Key Stakeholders */}
      {content.key_stakeholders && content.key_stakeholders.length > 0 && (
        <div className="space-y-3">
          <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Key Stakeholders
          </h4>
          <div className="flex flex-wrap gap-2">
            {content.key_stakeholders.map((stakeholder) => (
              <Badge key={stakeholder} variant="secondary">
                {stakeholder}
              </Badge>
            ))}
          </div>
        </div>
      )}

      {/* Timeline Overview */}
      {content.timeline_overview && (
        <div className="space-y-2">
          <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Timeline Overview
          </h4>
          <p className="text-sm leading-relaxed">{content.timeline_overview}</p>
        </div>
      )}
    </div>
  );
}
