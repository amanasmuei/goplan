"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import type { ActionItem, ImmediateActionContent } from "@/lib/types";

interface ImmediateActionProps {
  content: ImmediateActionContent;
}

const PRIORITY_COLORS: Record<string, string> = {
  high: "bg-red-100 text-red-800",
  medium: "bg-yellow-100 text-yellow-800",
  low: "bg-green-100 text-green-800",
  critical: "bg-red-200 text-red-900",
};

const URGENCY_COLORS: Record<string, string> = {
  high: "bg-red-100 text-red-800",
  medium: "bg-yellow-100 text-yellow-800",
  low: "bg-green-100 text-green-800",
};

function getPriorityColor(priority: string): string {
  return PRIORITY_COLORS[priority.toLowerCase()] ?? "bg-gray-100 text-gray-800";
}

function getUrgencyColor(urgency: string): string {
  return URGENCY_COLORS[urgency.toLowerCase()] ?? "bg-gray-100 text-gray-800";
}

function ActionItemRow({
  item,
  index,
}: {
  item: ActionItem;
  index: number;
}) {
  return (
    <div className="flex items-start gap-3 text-sm">
      <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-navy-100 text-xs font-semibold text-navy-700">
        {index + 1}
      </span>
      <div className="min-w-0 flex-1 space-y-1">
        <div className="flex items-start justify-between gap-2">
          <span className="font-medium">{item.title}</span>
          <Badge
            variant="secondary"
            className={cn("shrink-0 text-[11px]", getPriorityColor(item.priority))}
          >
            {item.priority}
          </Badge>
        </div>
        {item.description && (
          <p className="text-muted-foreground leading-relaxed">
            {item.description}
          </p>
        )}
        <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
          {item.owner && <span>Owner: {item.owner}</span>}
          {item.deadline && (
            <>
              <span className="text-border">|</span>
              <span>Due: {item.deadline}</span>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

export function ImmediateAction({ content }: ImmediateActionProps) {
  return (
    <div className="space-y-8">
      {/* Time Horizon */}
      {content.time_horizon && (
        <Badge variant="secondary" className="bg-brand/10 text-brand text-sm">
          {content.time_horizon}
        </Badge>
      )}

      {/* Quick Wins */}
      {content.quick_wins && content.quick_wins.length > 0 && (
        <div className="space-y-4">
          <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Quick Wins
          </h4>
          <div className="space-y-3">
            {content.quick_wins.map((item, i) => (
              <ActionItemRow key={item.title} item={item} index={i} />
            ))}
          </div>
        </div>
      )}

      {/* Critical Path */}
      {content.critical_path && content.critical_path.length > 0 && (
        <>
          <Separator />
          <div className="space-y-4">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Critical Path
            </h4>
            <div className="space-y-3 rounded-lg border-l-4 border-brand pl-4">
              {content.critical_path.map((item, i) => (
                <ActionItemRow key={item.title} item={item} index={i} />
              ))}
            </div>
          </div>
        </>
      )}

      {/* Resource Needs */}
      {content.resource_needs && content.resource_needs.length > 0 && (
        <>
          <Separator />
          <div className="space-y-4">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Resource Needs
            </h4>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left">
                    <th className="pb-2 pr-4 font-medium text-muted-foreground">
                      Resource
                    </th>
                    <th className="pb-2 pr-4 font-medium text-muted-foreground">
                      Quantity
                    </th>
                    <th className="pb-2 font-medium text-muted-foreground">
                      Urgency
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {content.resource_needs.map((need) => (
                    <tr key={need.resource} className="border-b last:border-0">
                      <td className="py-2 pr-4 font-medium">
                        {need.resource}
                      </td>
                      <td className="py-2 pr-4 text-muted-foreground">
                        {need.quantity}
                      </td>
                      <td className="py-2">
                        <Badge
                          variant="secondary"
                          className={cn(
                            "text-[11px]",
                            getUrgencyColor(need.urgency),
                          )}
                        >
                          {need.urgency}
                        </Badge>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {/* Week by Week */}
      {content.week_by_week && content.week_by_week.length > 0 && (
        <>
          <Separator />
          <div className="space-y-4">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Week by Week
            </h4>
            <div className="grid gap-4 sm:grid-cols-2">
              {content.week_by_week.map((week) => (
                <Card key={week.week} className="gap-3 py-4">
                  <CardHeader className="pb-0">
                    <CardTitle className="text-sm">
                      <Badge variant="secondary" className="mr-2 text-xs">
                        Week {week.week}
                      </Badge>
                      {week.focus}
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    {week.tasks && week.tasks.length > 0 && (
                      <ul className="space-y-1.5">
                        {week.tasks.map((task) => (
                          <li
                            key={task.title}
                            className="flex items-start gap-2 text-sm"
                          >
                            <span className="mt-1.5 block size-1.5 shrink-0 rounded-full bg-brand" />
                            <span className="text-muted-foreground">
                              {task.title}
                            </span>
                          </li>
                        ))}
                      </ul>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
