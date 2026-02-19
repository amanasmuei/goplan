// Plan status
export type PlanStatus = "draft" | "generating" | "complete" | "archived";

// Plan category
export type PlanCategory =
  | "business"
  | "saas"
  | "event"
  | "nonprofit"
  | "personal"
  | "education"
  | "real_estate"
  | "generic";

// Section type
export type SectionType =
  | "executive_brief"
  | "strategic_context"
  | "recommended_approach"
  | "phased_execution"
  | "immediate_action";

// Subscription tier
export type SubscriptionTier = "free" | "pro" | "pro_plus";

// User role
export type UserRole = "admin" | "team_lead" | "member";

// --- Core domain interfaces ---

export interface StrategicPlan {
  id: string;
  user_id: string;
  organization_id: string;
  title: string;
  original_input: string;
  category: PlanCategory;
  sub_category?: string;
  complexity?: string;
  status: PlanStatus;
  current_version: number;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface PlanSection {
  id: string;
  plan_id: string;
  section_type: SectionType;
  section_order: number;
  title: string;
  content: unknown;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface PlanResponse {
  plan: StrategicPlan;
  sections: PlanSection[];
}

export interface PlanListResponse {
  plans: StrategicPlan[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// --- Section content types ---

export interface ExecutiveBriefContent {
  summary: string;
  objective: string;
  scope: string;
  expected_outcome: string;
  key_stakeholders: string[];
  timeline_overview: string;
}

export interface ContextItem {
  title: string;
  description: string;
  impact: "high" | "medium" | "low";
}

export interface StrategicContextContent {
  industry_analysis: string;
  market_conditions: string;
  competitive_landscape: string;
  opportunities: ContextItem[];
  threats: ContextItem[];
  assumptions: string[];
}

export interface StrategyPillar {
  name: string;
  description: string;
  kpis: string[];
}

export interface Alternative {
  name: string;
  pros: string;
  cons: string;
  why_not_chosen: string;
}

export interface RiskItem {
  risk: string;
  likelihood: string;
  impact: string;
  mitigation: string;
}

export interface RecommendedApproachContent {
  core_strategy: string;
  rationale: string;
  key_pillars: StrategyPillar[];
  alternatives_considered: Alternative[];
  risk_mitigation: RiskItem[];
}

export interface Milestone {
  name: string;
  target: string;
  metric: string;
}

export interface Action {
  title: string;
  owner: string;
  timeline: string;
  priority: string;
}

export interface ExecutionPhase {
  name: string;
  duration: string;
  objective: string;
  milestones: Milestone[];
  actions: Action[];
  dependencies: string;
  success_criteria: string;
}

export interface PhasedExecutionContent {
  phases: ExecutionPhase[];
}

export interface ActionItem {
  title: string;
  description: string;
  owner: string;
  deadline: string;
  priority: string;
}

export interface ResourceNeed {
  resource: string;
  quantity: string;
  urgency: string;
}

export interface WeekPlan {
  week: number;
  focus: string;
  tasks: ActionItem[];
}

export interface ImmediateActionContent {
  time_horizon: string;
  quick_wins: ActionItem[];
  critical_path: ActionItem[];
  resource_needs: ResourceNeed[];
  week_by_week: WeekPlan[];
}

// --- Subscription ---

export interface Subscription {
  id: string;
  user_id: string;
  tier: SubscriptionTier;
  max_plans: number;
  max_regenerations_per_day: number;
  can_export: boolean;
  can_version_history: boolean;
  can_refine: boolean;
  current_period_start?: string;
  current_period_end?: string;
  created_at: string;
  updated_at: string;
}

// --- Versions ---

export interface SectionVersion {
  id: string;
  section_id: string;
  plan_id: string;
  version: number;
  content: unknown;
  refinement_context?: string;
  generated_by: string;
  token_usage?: unknown;
  created_at: string;
}

export interface PlanVersion {
  id: string;
  plan_id: string;
  version: number;
  snapshot: unknown;
  change_summary?: string;
  created_at: string;
}

export interface GenerationLog {
  id: string;
  plan_id?: string;
  user_id: string;
  action: string;
  section_type?: SectionType;
  status: string;
  prompt_tokens?: number;
  completion_tokens?: number;
  model?: string;
  duration_ms?: number;
  error_message?: string;
  metadata?: unknown;
  created_at: string;
}

// --- Request types ---

export interface CreatePlanRequest {
  input: string;
  category?: PlanCategory;
}

export interface RegenerateSectionRequest {
  additional_context?: string;
  depth?: "standard" | "deep";
}

export interface RefineSectionRequest {
  refinement_prompt: string;
}

// --- Auth types ---

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  organization_id: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface AuthResponse {
  token: string;
  user: User;
  expires_at: number;
}

// --- Constants ---

export const SECTION_ORDER: SectionType[] = [
  "executive_brief",
  "strategic_context",
  "recommended_approach",
  "phased_execution",
  "immediate_action",
];

export const SECTION_LABELS: Record<SectionType, string> = {
  executive_brief: "Executive Brief",
  strategic_context: "Strategic Context",
  recommended_approach: "Recommended Approach",
  phased_execution: "Phased Execution",
  immediate_action: "Immediate Action",
};

export const CATEGORY_LABELS: Record<PlanCategory, string> = {
  business: "Business",
  saas: "SaaS",
  event: "Event",
  nonprofit: "Nonprofit",
  personal: "Personal",
  education: "Education",
  real_estate: "Real Estate",
  generic: "Generic",
};
