import type { PlanCategory, PlanStatus, SectionType, SubscriptionTier } from "@/lib/types";

// Icon names from lucide-react mapped to each section type
export const SECTION_ICONS: Record<SectionType, string> = {
  executive_brief: "FileText",
  strategic_context: "Globe",
  recommended_approach: "Compass",
  phased_execution: "Layers",
  immediate_action: "Zap",
};

// Tailwind color classes per plan category
export const CATEGORY_COLORS: Record<PlanCategory, string> = {
  business: "bg-blue-100 text-blue-800",
  saas: "bg-violet-100 text-violet-800",
  event: "bg-amber-100 text-amber-800",
  nonprofit: "bg-emerald-100 text-emerald-800",
  personal: "bg-rose-100 text-rose-800",
  education: "bg-cyan-100 text-cyan-800",
  real_estate: "bg-orange-100 text-orange-800",
  generic: "bg-gray-100 text-gray-800",
};

// Tailwind color classes per plan status
export const STATUS_COLORS: Record<PlanStatus, string> = {
  draft: "bg-yellow-100 text-yellow-800",
  generating: "bg-blue-100 text-blue-800",
  complete: "bg-green-100 text-green-800",
  archived: "bg-gray-100 text-gray-500",
};

// Features available per subscription tier
export const TIER_FEATURES: Record<SubscriptionTier, string[]> = {
  free: [
    "Up to 3 strategic plans",
    "5 regenerations per day",
    "All 5 plan sections",
    "Basic AI analysis",
  ],
  pro: [
    "Up to 25 strategic plans",
    "20 regenerations per day",
    "Section refinement",
    "Version history",
    "Export to PDF/DOCX",
    "Priority AI processing",
  ],
  pro_plus: [
    "Up to 100 strategic plans",
    "50 regenerations per day",
    "Section refinement",
    "Full version history",
    "Export to all formats",
    "Semantic similarity search",
    "Priority support",
  ],
};
