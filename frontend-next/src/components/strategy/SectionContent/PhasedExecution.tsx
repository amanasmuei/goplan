"use client";

import { CheckCircle } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import type { PhasedExecutionContent } from "@/lib/types";

interface PhasedExecutionProps {
  content: PhasedExecutionContent;
}

const PRIORITY_COLORS: Record<string, string> = {
  high: "bg-red-100 text-red-800",
  medium: "bg-yellow-100 text-yellow-800",
  low: "bg-green-100 text-green-800",
  critical: "bg-red-200 text-red-900",
};

function getPriorityColor(priority: string): string {
  return PRIORITY_COLORS[priority.toLowerCase()] ?? "bg-gray-100 text-gray-800";
}

export function PhasedExecution({ content }: PhasedExecutionProps) {
  if (!content.phases || content.phases.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        No execution phases defined.
      </p>
    );
  }

  return (
    <div className="space-y-6">
      {content.phases.map((phase, index) => (
        <div key={phase.name} className="relative">
          {/* Timeline connector */}
          {index < content.phases.length - 1 && (
            <div className="absolute left-5 top-16 bottom-0 w-px bg-border" />
          )}

          <Card className="gap-4 py-5">
            <CardHeader className="pb-0">
              <CardTitle className="flex items-center gap-3 text-base">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-brand/10 text-sm font-bold text-brand">
                  {index + 1}
                </span>
                <span className="flex-1">{phase.name}</span>
                <Badge variant="secondary" className="text-xs">
                  {phase.duration}
                </Badge>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-5">
              {/* Objective */}
              <p className="text-sm leading-relaxed text-muted-foreground">
                {phase.objective}
              </p>

              {/* Milestones */}
              {phase.milestones && phase.milestones.length > 0 && (
                <div className="space-y-2">
                  <h5 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Milestones
                  </h5>
                  <div className="space-y-2">
                    {phase.milestones.map((milestone) => (
                      <div
                        key={milestone.name}
                        className="flex items-start gap-2 text-sm"
                      >
                        <CheckCircle className="mt-0.5 size-4 shrink-0 text-brand" />
                        <div>
                          <span className="font-medium">{milestone.name}</span>
                          <span className="text-muted-foreground">
                            {" "}
                            — {milestone.target}
                          </span>
                          {milestone.metric && (
                            <span className="text-muted-foreground/70">
                              {" "}
                              ({milestone.metric})
                            </span>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <Separator />

              {/* Actions */}
              {phase.actions && phase.actions.length > 0 && (
                <div className="space-y-2">
                  <h5 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Actions
                  </h5>
                  <div className="space-y-2">
                    {phase.actions.map((action) => (
                      <div
                        key={action.title}
                        className="flex items-start justify-between gap-4 rounded-md bg-muted/30 px-3 py-2 text-sm"
                      >
                        <div className="min-w-0 flex-1">
                          <span className="font-medium">{action.title}</span>
                          <div className="mt-1 flex flex-wrap items-center gap-2">
                            <Badge variant="outline" className="text-[11px]">
                              {action.owner}
                            </Badge>
                            <span className="text-xs text-muted-foreground">
                              {action.timeline}
                            </span>
                          </div>
                        </div>
                        <Badge
                          variant="secondary"
                          className={getPriorityColor(action.priority)}
                        >
                          {action.priority}
                        </Badge>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Dependencies & Success Criteria */}
              <div className="flex flex-wrap gap-4">
                {phase.dependencies && (
                  <div className="space-y-1">
                    <h5 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                      Dependencies
                    </h5>
                    <p className="text-sm text-muted-foreground">
                      {phase.dependencies}
                    </p>
                  </div>
                )}
                {phase.success_criteria && (
                  <div className="space-y-1">
                    <h5 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                      Success Criteria
                    </h5>
                    <p className="text-sm text-muted-foreground">
                      {phase.success_criteria}
                    </p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      ))}
    </div>
  );
}
