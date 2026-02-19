package models

// ExecutiveBriefContent represents the typed content for the executive_brief section.
type ExecutiveBriefContent struct {
	Summary          string   `json:"summary"`
	Objective        string   `json:"objective"`
	Scope            string   `json:"scope"`
	ExpectedOutcome  string   `json:"expected_outcome"`
	KeyStakeholders  []string `json:"key_stakeholders"`
	TimelineOverview string   `json:"timeline_overview"`
}

// StrategicContextContent represents the typed content for the strategic_context section.
type StrategicContextContent struct {
	IndustryAnalysis     string        `json:"industry_analysis"`
	MarketConditions     string        `json:"market_conditions"`
	CompetitiveLandscape string        `json:"competitive_landscape"`
	Opportunities        []ContextItem `json:"opportunities"`
	Threats              []ContextItem `json:"threats"`
	Assumptions          []string      `json:"assumptions"`
}

// ContextItem represents an opportunity or threat with impact assessment.
type ContextItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact" validate:"omitempty,oneof=high medium low"`
}

// RecommendedApproachContent represents the typed content for the recommended_approach section.
type RecommendedApproachContent struct {
	CoreStrategy           string           `json:"core_strategy"`
	Rationale              string           `json:"rationale"`
	KeyPillars             []StrategyPillar `json:"key_pillars"`
	AlternativesConsidered []Alternative    `json:"alternatives_considered"`
	RiskMitigation         []RiskItem       `json:"risk_mitigation"`
}

// StrategyPillar represents a key pillar of the recommended strategy.
type StrategyPillar struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	KPIs        []string `json:"kpis"`
}

// Alternative represents an alternative approach that was considered.
type Alternative struct {
	Name         string `json:"name"`
	Pros         string `json:"pros"`
	Cons         string `json:"cons"`
	WhyNotChosen string `json:"why_not_chosen"`
}

// RiskItem represents a risk with likelihood, impact, and mitigation strategy.
type RiskItem struct {
	Risk       string `json:"risk"`
	Likelihood string `json:"likelihood"`
	Impact     string `json:"impact"`
	Mitigation string `json:"mitigation"`
}

// PhasedExecutionContent represents the typed content for the phased_execution section.
type PhasedExecutionContent struct {
	Phases []ExecutionPhase `json:"phases"`
}

// ExecutionPhase represents a single phase in the execution plan.
type ExecutionPhase struct {
	Name            string      `json:"name"`
	Duration        string      `json:"duration"`
	Objective       string      `json:"objective"`
	Milestones      []Milestone `json:"milestones"`
	Actions         []Action    `json:"actions"`
	Dependencies    string      `json:"dependencies"`
	SuccessCriteria string      `json:"success_criteria"`
}

// Milestone represents a milestone within an execution phase.
type Milestone struct {
	Name   string `json:"name"`
	Target string `json:"target"`
	Metric string `json:"metric"`
}

// Action represents an action item within an execution phase.
type Action struct {
	Title    string `json:"title"`
	Owner    string `json:"owner"`
	Timeline string `json:"timeline"`
	Priority string `json:"priority"`
}

// ImmediateActionContent represents the typed content for the immediate_action section.
type ImmediateActionContent struct {
	TimeHorizon  string         `json:"time_horizon"`
	QuickWins    []ActionItem   `json:"quick_wins"`
	CriticalPath []ActionItem   `json:"critical_path"`
	ResourceNeeds []ResourceNeed `json:"resource_needs"`
	WeekByWeek   []WeekPlan     `json:"week_by_week"`
}

// ActionItem represents an action with ownership and deadline.
type ActionItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
	Deadline    string `json:"deadline"`
	Priority    string `json:"priority"`
}

// ResourceNeed represents a resource requirement.
type ResourceNeed struct {
	Resource string `json:"resource"`
	Quantity string `json:"quantity"`
	Urgency  string `json:"urgency"`
}

// WeekPlan represents a week's plan within the immediate action section.
type WeekPlan struct {
	Week  int          `json:"week"`
	Focus string       `json:"focus"`
	Tasks []ActionItem `json:"tasks"`
}
