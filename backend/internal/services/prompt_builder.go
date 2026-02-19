package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goplan/backend/internal/models"
)

// BuildStrategyPrompt constructs the system and user prompts for initial strategy generation.
func BuildStrategyPrompt(input string, category models.PlanCategory, depth string) (systemPrompt string, userPrompt string) {
	categoryOverlay := getCategoryOverlay(category)

	depthInstruction := ""
	if depth == "deep" {
		depthInstruction = `
DEPTH LEVEL: DEEP ANALYSIS
- Provide exhaustive detail in every section
- Include 5+ phases in the execution plan
- Provide 4+ alternatives considered with detailed trade-off analysis
- Include week-by-week breakdown for at least the first 8 weeks
- Add comprehensive risk mitigation with likelihood and impact scoring
- Provide detailed KPIs with specific numerical targets where possible`
	}

	systemPrompt = fmt.Sprintf(`You are a world-class strategy consultant with expertise across industries including technology, finance, healthcare, real estate, education, and non-profit sectors. You have decades of experience helping organizations and individuals turn ambiguous goals into actionable, phased strategic plans.

Your task is to analyze the user's input and produce a comprehensive strategic plan as a JSON object. You must:

1. CLASSIFY the input: Determine the category, sub-category, and complexity level.
2. GENERATE a compelling title for the plan.
3. PRODUCE exactly 5 sections with structured content matching the schemas below.

You MUST return ONLY valid JSON matching this exact structure (no markdown, no explanation, no code blocks):

{
  "classification": {
    "category": "business|saas|event|nonprofit|personal|education|real_estate|generic",
    "sub_category": "a specific sub-category relevant to the input",
    "complexity": "low|medium|high"
  },
  "title": "A compelling, specific title for this strategic plan",
  "sections": {
    "executive_brief": {
      "summary": "A concise executive summary of the entire plan",
      "objective": "The primary objective this plan aims to achieve",
      "scope": "What is included and excluded from this plan",
      "expected_outcome": "The expected end-state after successful execution",
      "key_stakeholders": ["stakeholder1", "stakeholder2", "stakeholder3"],
      "timeline_overview": "High-level timeline from start to completion"
    },
    "strategic_context": {
      "industry_analysis": "Analysis of the relevant industry or domain",
      "market_conditions": "Current market conditions and trends affecting this plan",
      "competitive_landscape": "Overview of competition or comparable efforts",
      "opportunities": [
        {"title": "Opportunity name", "description": "Details", "impact": "high|medium|low"}
      ],
      "threats": [
        {"title": "Threat name", "description": "Details", "impact": "high|medium|low"}
      ],
      "assumptions": ["assumption1", "assumption2"]
    },
    "recommended_approach": {
      "core_strategy": "The central strategic approach being recommended",
      "rationale": "Why this approach is the best path forward",
      "key_pillars": [
        {"name": "Pillar name", "description": "Details", "kpis": ["KPI 1", "KPI 2"]}
      ],
      "alternatives_considered": [
        {"name": "Alternative name", "pros": "Advantages", "cons": "Disadvantages", "why_not_chosen": "Reason"}
      ],
      "risk_mitigation": [
        {"risk": "Risk description", "likelihood": "high|medium|low", "impact": "high|medium|low", "mitigation": "Mitigation strategy"}
      ]
    },
    "phased_execution": {
      "phases": [
        {
          "name": "Phase name",
          "duration": "e.g., 2 weeks, 1 month",
          "objective": "What this phase aims to accomplish",
          "milestones": [
            {"name": "Milestone", "target": "Target date or condition", "metric": "How to measure"}
          ],
          "actions": [
            {"title": "Action item", "owner": "Who is responsible", "timeline": "When", "priority": "high|medium|low"}
          ],
          "dependencies": "What must be in place before this phase",
          "success_criteria": "How to know this phase is complete"
        }
      ]
    },
    "immediate_action": {
      "time_horizon": "e.g., First 30 days",
      "quick_wins": [
        {"title": "Quick win", "description": "Details", "owner": "Who", "deadline": "When", "priority": "high|medium|low"}
      ],
      "critical_path": [
        {"title": "Critical item", "description": "Details", "owner": "Who", "deadline": "When", "priority": "high|medium|low"}
      ],
      "resource_needs": [
        {"resource": "What is needed", "quantity": "How much", "urgency": "high|medium|low"}
      ],
      "week_by_week": [
        {"week": 1, "focus": "Focus area", "tasks": [
          {"title": "Task", "description": "Details", "owner": "Who", "deadline": "When", "priority": "high|medium|low"}
        ]}
      ]
    }
  }
}

DOMAIN-SPECIFIC GUIDANCE:
%s
%s

IMPORTANT RULES:
- Return ONLY the JSON object, no surrounding text or markdown formatting
- All string fields must be non-empty
- All array fields must contain at least one item
- The phased_execution must contain at least 3 phases
- The immediate_action week_by_week must cover at least 4 weeks
- Provide realistic, actionable content — not generic boilerplate
- Tailor every section to the specific input provided`, categoryOverlay, depthInstruction)

	userPrompt = input
	return systemPrompt, userPrompt
}

// BuildSectionRegenerationPrompt constructs prompts to regenerate a single section with full plan context.
func BuildSectionRegenerationPrompt(
	originalInput string,
	category models.PlanCategory,
	sectionType models.SectionType,
	currentContent any,
	additionalContext string,
	otherSections []models.PlanSection,
) (systemPrompt string, userPrompt string) {
	sectionSchema := getSectionSchema(sectionType)

	// Build context from other sections
	var contextParts []string
	for _, s := range otherSections {
		if s.SectionType == sectionType {
			continue
		}
		contentJSON, _ := json.Marshal(s.Content)
		contextParts = append(contextParts, fmt.Sprintf("--- %s ---\n%s", models.DefaultSectionTitles[s.SectionType], string(contentJSON)))
	}
	otherSectionsContext := strings.Join(contextParts, "\n\n")

	currentContentJSON, _ := json.Marshal(currentContent)

	additionalCtxBlock := ""
	if additionalContext != "" {
		additionalCtxBlock = fmt.Sprintf("\nADDITIONAL CONTEXT FROM USER:\n%s", additionalContext)
	}

	categoryOverlay := getCategoryOverlay(category)

	systemPrompt = fmt.Sprintf(`You are a world-class strategy consultant. You are regenerating the "%s" section of a strategic plan.

The original user request was: "%s"
Plan category: %s

DOMAIN GUIDANCE:
%s

Here are the other sections of the plan for context:
%s

The current version of this section (which needs to be regenerated with fresh perspective):
%s
%s
You MUST return ONLY valid JSON matching this exact structure for the "%s" section (no markdown, no explanation, no code blocks):

%s

IMPORTANT:
- Maintain consistency with the other sections
- Provide fresh, improved content — do not simply copy the current version
- All fields must be non-empty and all arrays must have at least one item
- Return ONLY the JSON object`, models.DefaultSectionTitles[sectionType], originalInput, string(category), categoryOverlay, otherSectionsContext, string(currentContentJSON), additionalCtxBlock, string(sectionType), sectionSchema)

	userPrompt = fmt.Sprintf("Regenerate the %s section with a fresh perspective, maintaining consistency with the rest of the plan.", models.DefaultSectionTitles[sectionType])
	return systemPrompt, userPrompt
}

// BuildRefinementPrompt constructs prompts to refine an existing section based on user feedback.
func BuildRefinementPrompt(
	sectionType models.SectionType,
	currentContent any,
	refinementPrompt string,
	planContext string,
) (systemPrompt string, userPrompt string) {
	sectionSchema := getSectionSchema(sectionType)
	currentContentJSON, _ := json.Marshal(currentContent)

	planContextBlock := ""
	if planContext != "" {
		planContextBlock = fmt.Sprintf("\nPLAN CONTEXT:\n%s", planContext)
	}

	systemPrompt = fmt.Sprintf(`You are a world-class strategy consultant. You are refining the "%s" section of a strategic plan based on specific user feedback.

Current content of this section:
%s
%s
You MUST return ONLY valid JSON matching this exact structure for the "%s" section (no markdown, no explanation, no code blocks):

%s

IMPORTANT:
- Apply the user's feedback precisely
- Preserve parts of the content that the user did not ask to change
- All fields must be non-empty and all arrays must have at least one item
- Return ONLY the JSON object`, models.DefaultSectionTitles[sectionType], string(currentContentJSON), planContextBlock, string(sectionType), sectionSchema)

	userPrompt = refinementPrompt
	return systemPrompt, userPrompt
}

// getCategoryOverlay returns domain-specific prompt additions for a given plan category.
func getCategoryOverlay(category models.PlanCategory) string {
	switch category {
	case models.PlanCategoryBusiness:
		return `BUSINESS STRATEGY FOCUS:
- Include revenue model analysis with projected income streams
- Apply market validation frameworks (TAM/SAM/SOM where relevant)
- Break down capital expenditure vs operating expenditure
- Analyze competitive positioning using Porter's Five Forces or similar
- Include unit economics and break-even analysis
- Consider regulatory and compliance requirements`

	case models.PlanCategorySaaS:
		return `SaaS STRATEGY FOCUS:
- Prioritize MVP feature set with must-have vs nice-to-have distinction
- Include tech stack recommendations with scalability considerations
- Define user acquisition channels and CAC targets
- Project MRR/ARR with growth assumptions
- Include churn reduction strategies and retention metrics
- Plan for infrastructure scaling milestones
- Consider pricing model (freemium, tiered, usage-based)`

	case models.PlanCategoryEvent:
		return `EVENT STRATEGY FOCUS:
- Include vendor management plan with backup options
- Build timeline with critical path and buffer periods
- Allocate budget by category (venue, catering, marketing, tech, staffing)
- Plan risk contingencies for weather, attendance, vendor failure
- Include day-of logistics and run-of-show planning
- Define attendee experience journey and touchpoints
- Consider sponsorship and partnership opportunities`

	case models.PlanCategoryNonprofit:
		return `NONPROFIT STRATEGY FOCUS:
- Develop fundraising strategy with diversified revenue sources (grants, donations, earned income)
- Create impact measurement framework with quantifiable outcomes
- Build stakeholder engagement plan (board, donors, beneficiaries, community)
- Include volunteer recruitment and management strategy
- Plan for regulatory compliance and reporting requirements
- Consider strategic partnerships with other organizations
- Define theory of change and logic model`

	case models.PlanCategoryPersonal:
		return `PERSONAL STRATEGY FOCUS:
- Apply goal decomposition into measurable sub-goals
- Include habit-building frameworks (e.g., implementation intentions, habit stacking)
- Create personal resource inventory (time, money, skills, network)
- Build accountability mechanisms and progress tracking
- Consider work-life balance and sustainability
- Include contingency plans for common obstacles
- Define personal success metrics and reflection points`

	case models.PlanCategoryEducation:
		return `EDUCATION STRATEGY FOCUS:
- Apply curriculum design principles with clear learning progressions
- Create assessment rubrics aligned with learning objectives
- Map learning outcomes to activities and assessments (constructive alignment)
- Include differentiated instruction strategies
- Plan for technology integration and accessibility
- Consider stakeholder engagement (students, parents, administrators)
- Define program evaluation criteria and continuous improvement cycle`

	case models.PlanCategoryRealEstate:
		return `REAL ESTATE STRATEGY FOCUS:
- Include market comparable analysis with recent transaction data
- Evaluate financing options (conventional, FHA, private, commercial)
- Calculate renovation ROI and value-add potential
- Build timeline to close with contingency buffers
- Assess zoning, permitting, and regulatory requirements
- Include property management and maintenance planning
- Consider tax implications and depreciation strategies
- Evaluate insurance requirements and risk coverage`

	default:
		return `GENERAL STRATEGY FOCUS:
- Apply balanced strategic framework covering vision, execution, and measurement
- Include stakeholder analysis and communication plan
- Build resource allocation model with budget considerations
- Define clear success metrics and evaluation criteria
- Consider both short-term wins and long-term sustainability
- Include change management and adoption strategies`
	}
}

// getSectionSchema returns the JSON schema example for a specific section type.
func getSectionSchema(sectionType models.SectionType) string {
	switch sectionType {
	case models.SectionTypeExecutiveBrief:
		return `{
  "summary": "...",
  "objective": "...",
  "scope": "...",
  "expected_outcome": "...",
  "key_stakeholders": ["..."],
  "timeline_overview": "..."
}`

	case models.SectionTypeStrategicContext:
		return `{
  "industry_analysis": "...",
  "market_conditions": "...",
  "competitive_landscape": "...",
  "opportunities": [{"title": "...", "description": "...", "impact": "high|medium|low"}],
  "threats": [{"title": "...", "description": "...", "impact": "high|medium|low"}],
  "assumptions": ["..."]
}`

	case models.SectionTypeRecommendedApproach:
		return `{
  "core_strategy": "...",
  "rationale": "...",
  "key_pillars": [{"name": "...", "description": "...", "kpis": ["..."]}],
  "alternatives_considered": [{"name": "...", "pros": "...", "cons": "...", "why_not_chosen": "..."}],
  "risk_mitigation": [{"risk": "...", "likelihood": "high|medium|low", "impact": "high|medium|low", "mitigation": "..."}]
}`

	case models.SectionTypePhasedExecution:
		return `{
  "phases": [{
    "name": "...",
    "duration": "...",
    "objective": "...",
    "milestones": [{"name": "...", "target": "...", "metric": "..."}],
    "actions": [{"title": "...", "owner": "...", "timeline": "...", "priority": "high|medium|low"}],
    "dependencies": "...",
    "success_criteria": "..."
  }]
}`

	case models.SectionTypeImmediateAction:
		return `{
  "time_horizon": "...",
  "quick_wins": [{"title": "...", "description": "...", "owner": "...", "deadline": "...", "priority": "high|medium|low"}],
  "critical_path": [{"title": "...", "description": "...", "owner": "...", "deadline": "...", "priority": "high|medium|low"}],
  "resource_needs": [{"resource": "...", "quantity": "...", "urgency": "high|medium|low"}],
  "week_by_week": [{"week": 1, "focus": "...", "tasks": [{"title": "...", "description": "...", "owner": "...", "deadline": "...", "priority": "high|medium|low"}]}]
}`

	default:
		return `{}`
	}
}
