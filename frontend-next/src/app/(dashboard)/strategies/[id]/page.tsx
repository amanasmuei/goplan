"use client";

import { useState, useCallback } from "react";
import { useParams } from "next/navigation";
import { useStrategy } from "@/hooks/useStrategy";
import { SectionNav } from "@/components/strategy/SectionNav";
import { SectionViewer } from "@/components/strategy/SectionViewer";
import { ActionPanel } from "@/components/strategy/ActionPanel";
import { GeneratingState } from "@/components/strategy/GeneratingState";
import { SECTION_ORDER, type SectionType } from "@/lib/types";

export default function StrategyViewerPage() {
  const params = useParams<{ id: string }>();
  const [activeSection, setActiveSection] = useState<SectionType>(
    SECTION_ORDER[0],
  );

  const { data, isLoading, isError } = useStrategy(params.id, {
    refetchInterval: (query) => {
      const plan = query.state.data?.plan;
      return plan?.status === "generating" ? 3000 : false;
    },
  });

  const handleSectionClick = useCallback((section: SectionType) => {
    setActiveSection(section);
    const el = document.getElementById(`section-${section}`);
    el?.scrollIntoView({ behavior: "smooth", block: "start" });
  }, []);

  if (isLoading) {
    return <GeneratingState />;
  }

  if (isError || !data) {
    return (
      <div className="flex h-[calc(100vh-2rem)] items-center justify-center">
        <div className="text-center">
          <h2 className="text-lg font-semibold">Strategy not found</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            The strategy you are looking for does not exist or you do not have
            access to it.
          </p>
        </div>
      </div>
    );
  }

  const { plan, sections } = data;

  if (plan.status === "generating") {
    return <GeneratingState />;
  }

  return (
    <div className="flex h-[calc(100vh-2rem)]">
      <SectionNav
        plan={plan}
        activeSection={activeSection}
        onSectionClick={handleSectionClick}
      />
      <SectionViewer sections={sections} />
      <ActionPanel plan={plan} activeSection={activeSection} />
    </div>
  );
}
