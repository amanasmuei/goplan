"use client";

import {
  FileText,
  Globe,
  Compass,
  Layers,
  Zap,
  type LucideIcon,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import { CATEGORY_COLORS, STATUS_COLORS } from "@/lib/constants";
import {
  SECTION_ORDER,
  SECTION_LABELS,
  CATEGORY_LABELS,
  type SectionType,
  type StrategicPlan,
} from "@/lib/types";

interface SectionNavProps {
  plan: StrategicPlan;
  activeSection: SectionType;
  onSectionClick: (section: SectionType) => void;
}

const SECTION_ICON_MAP: Record<SectionType, LucideIcon> = {
  executive_brief: FileText,
  strategic_context: Globe,
  recommended_approach: Compass,
  phased_execution: Layers,
  immediate_action: Zap,
};

export function SectionNav({
  plan,
  activeSection,
  onSectionClick,
}: SectionNavProps) {
  return (
    <aside className="sticky top-0 flex h-full w-60 shrink-0 flex-col border-r bg-white">
      {/* Plan title */}
      <div className="space-y-3 p-5">
        <h2 className="line-clamp-2 font-semibold leading-snug">
          {plan.title}
        </h2>
        <div className="flex flex-wrap items-center gap-1.5">
          <Badge
            variant="secondary"
            className={cn("text-[11px]", CATEGORY_COLORS[plan.category])}
          >
            {CATEGORY_LABELS[plan.category]}
          </Badge>
          <Badge
            variant="secondary"
            className={cn("text-[11px]", STATUS_COLORS[plan.status])}
          >
            {plan.status}
          </Badge>
        </div>
      </div>

      <Separator />

      {/* Section list */}
      <nav className="flex-1 space-y-0.5 p-3">
        {SECTION_ORDER.map((sectionType) => {
          const Icon = SECTION_ICON_MAP[sectionType];
          const isActive = activeSection === sectionType;

          return (
            <button
              key={sectionType}
              type="button"
              onClick={() => onSectionClick(sectionType)}
              className={cn(
                "flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
                isActive
                  ? "border-l-2 border-brand bg-navy-50 text-navy-900"
                  : "text-muted-foreground hover:bg-gray-50",
              )}
            >
              <Icon className="size-4 shrink-0" />
              <span className="truncate">{SECTION_LABELS[sectionType]}</span>
            </button>
          );
        })}
      </nav>
    </aside>
  );
}
