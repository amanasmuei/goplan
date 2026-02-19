"use client";

import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  ExecutiveBrief,
  StrategicContext,
  RecommendedApproach,
  PhasedExecution,
  ImmediateAction,
} from "@/components/strategy/SectionContent";
import {
  SECTION_ORDER,
  SECTION_LABELS,
  type PlanSection,
  type SectionType,
  type ExecutiveBriefContent,
  type StrategicContextContent,
  type RecommendedApproachContent,
  type PhasedExecutionContent,
  type ImmediateActionContent,
} from "@/lib/types";

interface SectionViewerProps {
  sections: PlanSection[];
}

function renderSectionContent(section: PlanSection) {
  const content = section.content;
  if (!content) return null;

  switch (section.section_type) {
    case "executive_brief":
      return <ExecutiveBrief content={content as ExecutiveBriefContent} />;
    case "strategic_context":
      return <StrategicContext content={content as StrategicContextContent} />;
    case "recommended_approach":
      return (
        <RecommendedApproach content={content as RecommendedApproachContent} />
      );
    case "phased_execution":
      return <PhasedExecution content={content as PhasedExecutionContent} />;
    case "immediate_action":
      return <ImmediateAction content={content as ImmediateActionContent} />;
    default:
      return null;
  }
}

export function SectionViewer({ sections }: SectionViewerProps) {
  // Build a map of section_type -> PlanSection for ordered rendering
  const sectionMap = new Map<SectionType, PlanSection>();
  for (const section of sections) {
    sectionMap.set(section.section_type, section);
  }

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto max-w-4xl space-y-10 px-8 py-8">
        {SECTION_ORDER.map((sectionType, index) => {
          const section = sectionMap.get(sectionType);
          if (!section) return null;

          return (
            <section
              key={sectionType}
              id={`section-${sectionType}`}
              className="scroll-mt-8"
            >
              {/* Section header */}
              <div className="mb-6 flex items-center gap-3">
                <h2 className="text-xl font-semibold">
                  {section.title || SECTION_LABELS[sectionType]}
                </h2>
                {section.version > 1 && (
                  <Badge variant="outline" className="text-xs">
                    v{section.version}
                  </Badge>
                )}
              </div>

              {/* Section content */}
              {renderSectionContent(section)}

              {/* Separator between sections */}
              {index < SECTION_ORDER.length - 1 && (
                <Separator className="mt-10" />
              )}
            </section>
          );
        })}
      </div>
    </div>
  );
}
