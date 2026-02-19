"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import type { ContextItem, StrategicContextContent } from "@/lib/types";

interface StrategicContextProps {
  content: StrategicContextContent;
}

const IMPACT_COLORS: Record<string, string> = {
  high: "bg-red-100 text-red-800",
  medium: "bg-yellow-100 text-yellow-800",
  low: "bg-green-100 text-green-800",
};

function ContextCard({
  item,
  variant,
}: {
  item: ContextItem;
  variant: "opportunity" | "threat";
}) {
  return (
    <Card
      className={cn(
        "gap-3 py-4",
        variant === "threat" && "border-red-200",
      )}
    >
      <CardHeader className="pb-0">
        <CardTitle className="flex items-center justify-between text-sm">
          <span>{item.title}</span>
          <Badge
            variant="secondary"
            className={cn("text-[11px]", IMPACT_COLORS[item.impact])}
          >
            {item.impact}
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground leading-relaxed">
          {item.description}
        </p>
      </CardContent>
    </Card>
  );
}

export function StrategicContext({ content }: StrategicContextProps) {
  return (
    <div className="space-y-8">
      {/* Analysis sections */}
      <div className="space-y-6">
        {content.industry_analysis && (
          <div className="space-y-2">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Industry Analysis
            </h4>
            <p className="text-sm leading-relaxed">
              {content.industry_analysis}
            </p>
          </div>
        )}
        {content.market_conditions && (
          <div className="space-y-2">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Market Conditions
            </h4>
            <p className="text-sm leading-relaxed">
              {content.market_conditions}
            </p>
          </div>
        )}
        {content.competitive_landscape && (
          <div className="space-y-2">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Competitive Landscape
            </h4>
            <p className="text-sm leading-relaxed">
              {content.competitive_landscape}
            </p>
          </div>
        )}
      </div>

      <Separator />

      {/* Opportunities */}
      {content.opportunities && content.opportunities.length > 0 && (
        <div className="space-y-4">
          <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Opportunities
          </h4>
          <div className="grid gap-4 sm:grid-cols-2">
            {content.opportunities.map((item) => (
              <ContextCard key={item.title} item={item} variant="opportunity" />
            ))}
          </div>
        </div>
      )}

      {/* Threats */}
      {content.threats && content.threats.length > 0 && (
        <div className="space-y-4">
          <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Threats
          </h4>
          <div className="grid gap-4 sm:grid-cols-2">
            {content.threats.map((item) => (
              <ContextCard key={item.title} item={item} variant="threat" />
            ))}
          </div>
        </div>
      )}

      {/* Assumptions */}
      {content.assumptions && content.assumptions.length > 0 && (
        <>
          <Separator />
          <div className="space-y-3">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Assumptions
            </h4>
            <ul className="space-y-2">
              {content.assumptions.map((assumption, i) => (
                <li
                  key={i}
                  className="flex items-start gap-2 text-sm leading-relaxed"
                >
                  <span className="mt-1.5 block size-1.5 shrink-0 rounded-full bg-muted-foreground/40" />
                  {assumption}
                </li>
              ))}
            </ul>
          </div>
        </>
      )}
    </div>
  );
}
