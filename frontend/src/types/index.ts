export type TaskStatus =
  | 'draft'
  | 'pending_acknowledgment'
  | 'acknowledged'
  | 'active'
  | 'blocked'
  | 'pending_review'
  | 'completed'
  | 'cancelled'

export type LinkType = 'similar' | 'dependent' | 'retry' | 'related'

export type BlockerType =
  | 'approval'
  | 'external_team'
  | 'vendor'
  | 'technical'
  | 'resource'
  | 'requirements'

export interface Task {
  id: string
  title: string
  description: string
  status: TaskStatus
  created_by: string
  assigned_to?: string
  project_id: string
  organization_id: string
  estimated_days?: number
  predicted_days_low?: number
  predicted_days_high?: number
  prediction_confidence?: number
  planning_quality_score?: number
  acknowledged_at?: string
  started_at?: string
  completed_at?: string
  actual_days?: number
  tags?: string[]
  created_at: string
  updated_at: string
}

export interface SimilarTask {
  id: string
  title: string
  status: TaskStatus
  similarity_score: number
  actual_days?: number
  estimated_days?: number
  blockers_summary?: string
  lessons_learned_excerpt?: string
}

export interface BlockerRisk {
  type: string
  probability: number
  examples?: string[]
}

export interface Predictions {
  predicted_days_low: number
  predicted_days_high: number
  confidence: number
  blocker_risks: BlockerRisk[]
}

export interface Assessment {
  score: number
  breakdown: Record<string, number>
  suggestions: string[]
}

export interface TaskResponse {
  task: Task
  similar_tasks?: SimilarTask[]
  predictions?: Predictions
  planning_assessment?: Assessment
}

export interface TaskListResponse {
  tasks: Task[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface TaskLink {
  id: string
  source_task_id: string
  target_task_id: string
  link_type: LinkType
  created_by: string
  notes?: string
  created_at: string
}

export interface TaskJustification {
  id: string
  task_id: string
  checked_same_project: boolean
  checked_same_stakeholders: boolean
  checked_same_dependencies: boolean
  justification_text: string
  created_by: string
  created_at: string
}

export interface TaskBlocker {
  id: string
  task_id: string
  blocker_type: BlockerType
  description: string
  resolved_at?: string
  days_blocked?: number
  created_by: string
  created_at: string
}

export interface TaskReview {
  id: string
  task_id: string
  prediction_accuracy_rating: number
  prediction_feedback?: string
  lessons_learned?: string
  would_approach_differently?: string
  created_by: string
  created_at: string
}

export interface User {
  id: string
  email: string
  name: string
  role: 'admin' | 'team_lead' | 'member'
  organization_id: string
}

export type ProjectStatus = 'active' | 'archived'

export interface Project {
  id: string
  name: string
  description?: string
  status: ProjectStatus
  organization_id: string
  created_by?: string
  created_at: string
  updated_at: string
}

export interface ProjectResponse {
  project: Project
  teams?: Team[]
  task_count: number
}

export interface ProjectListResponse {
  projects: ProjectResponse[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface CreateProjectRequest {
  name: string
  description?: string
  team_ids?: string[]
}

export interface UpdateProjectRequest {
  name?: string
  description?: string
  status?: ProjectStatus
}

export type TeamRole = 'owner' | 'manager' | 'member' | 'viewer'

export interface Team {
  id: string
  name: string
  description?: string
  organization_id: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface TeamMember {
  id: string
  team_id: string
  user_id: string
  role: TeamRole
  joined_at: string
  user?: User
}

export interface TeamResponse {
  team: Team
  member_count: number
  members?: TeamMember[]
}

export interface TeamListResponse {
  teams: TeamResponse[]
  total: number
}

export interface TeamMemberListResponse {
  members: TeamMember[]
  total: number
}

export interface CreateTeamRequest {
  name: string
  description?: string
}

export interface UpdateTeamRequest {
  name?: string
  description?: string
}

export interface AddTeamMemberRequest {
  user_id: string
  role: TeamRole
}

export interface UpdateMemberRoleRequest {
  role: TeamRole
}

export interface AssignTeamsRequest {
  team_ids: string[]
}

export interface CreateTaskRequest {
  title: string
  description: string
  project_id: string
  estimated_days?: number
  assigned_to?: string
  tags?: string[]
}

export interface CreateLinkRequest {
  target_task_id: string
  link_type: LinkType
  notes?: string
}

export interface CreateJustificationRequest {
  checked_same_project: boolean
  checked_same_stakeholders: boolean
  checked_same_dependencies: boolean
  justification_text: string
}

export interface CreateBlockerRequest {
  blocker_type: BlockerType
  description: string
}

export interface CreateReviewRequest {
  prediction_accuracy_rating: number
  prediction_feedback?: string
  lessons_learned?: string
  would_approach_differently?: string
}

export type AcknowledgmentAction = 'accept' | 'modify' | 'disagree'

export interface AcknowledgmentRequest {
  action: AcknowledgmentAction
  modified_estimate?: number
  disagreement_notes?: string
}
