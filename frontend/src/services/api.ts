import axios from 'axios'
import { useAuthStore } from '../store/authStore'
import type {
  Task,
  TaskResponse,
  TaskListResponse,
  TaskLink,
  TaskJustification,
  TaskBlocker,
  TaskReview,
  SimilarTask,
  CreateTaskRequest,
  CreateLinkRequest,
  CreateJustificationRequest,
  CreateBlockerRequest,
  CreateReviewRequest,
  AcknowledgmentRequest,
  TeamResponse,
  TeamListResponse,
  TeamMemberListResponse,
  CreateTeamRequest,
  UpdateTeamRequest,
  AddTeamMemberRequest,
  UpdateMemberRoleRequest,
  ProjectResponse,
  ProjectListResponse,
  CreateProjectRequest,
  UpdateProjectRequest,
  AssignTeamsRequest,
  User,
} from '../types'

// Auth types
export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
}

export interface AuthResponse {
  token: string
  expires_at: string
  user: User
}

const API_BASE = import.meta.env.VITE_API_URL || '/api/v1'

const api = axios.create({
  baseURL: API_BASE,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth token to requests
api.interceptors.request.use((config) => {
  const { token, isTokenExpired, logout } = useAuthStore.getState()
  if (token && !isTokenExpired()) {
    config.headers.Authorization = `Bearer ${token}`
  } else if (token) {
    logout()
    window.location.href = '/login'
  }
  return config
})

// Handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Task API
export const taskApi = {
  create: async (data: CreateTaskRequest): Promise<TaskResponse> => {
    const response = await api.post<TaskResponse>('/tasks', data)
    return response.data
  },

  get: async (id: string): Promise<Task> => {
    const response = await api.get<Task>(`/tasks/${id}`)
    return response.data
  },

  update: async (id: string, data: Partial<Task>): Promise<Task> => {
    const response = await api.put<Task>(`/tasks/${id}`, data)
    return response.data
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/tasks/${id}`)
  },

  list: async (params?: {
    project_id?: string
    status?: string
    assigned_to?: string
    search?: string
    page?: number
    page_size?: number
  }): Promise<TaskListResponse> => {
    const response = await api.get<TaskListResponse>('/tasks', { params })
    return response.data
  },

  getSimilar: async (id: string): Promise<{ similar_tasks: SimilarTask[] }> => {
    const response = await api.get<{ similar_tasks: SimilarTask[] }>(
      `/tasks/${id}/similar`
    )
    return response.data
  },

  acknowledge: async (id: string, data?: AcknowledgmentRequest): Promise<Task> => {
    const response = await api.post<Task>(`/tasks/${id}/acknowledge`, data)
    return response.data
  },

  start: async (id: string): Promise<Task> => {
    const response = await api.post<Task>(`/tasks/${id}/start`)
    return response.data
  },

  complete: async (id: string): Promise<Task> => {
    const response = await api.post<Task>(`/tasks/${id}/complete`)
    return response.data
  },
}

// Link API
export const linkApi = {
  create: async (taskId: string, data: CreateLinkRequest): Promise<TaskLink> => {
    const response = await api.post<TaskLink>(`/tasks/${taskId}/links`, data)
    return response.data
  },

  list: async (taskId: string): Promise<{ links: TaskLink[] }> => {
    const response = await api.get<{ links: TaskLink[] }>(`/tasks/${taskId}/links`)
    return response.data
  },

  delete: async (taskId: string, linkId: string): Promise<void> => {
    await api.delete(`/tasks/${taskId}/links/${linkId}`)
  },
}

// Justification API
export const justificationApi = {
  create: async (
    taskId: string,
    data: CreateJustificationRequest
  ): Promise<TaskJustification> => {
    const response = await api.post<TaskJustification>(
      `/tasks/${taskId}/justify`,
      data
    )
    return response.data
  },

  get: async (taskId: string): Promise<TaskJustification | null> => {
    try {
      const response = await api.get<TaskJustification>(`/tasks/${taskId}/justify`)
      return response.data
    } catch {
      return null
    }
  },
}

// Blocker API
export const blockerApi = {
  create: async (
    taskId: string,
    data: CreateBlockerRequest
  ): Promise<TaskBlocker> => {
    const response = await api.post<TaskBlocker>(
      `/tasks/${taskId}/blockers`,
      data
    )
    return response.data
  },

  list: async (taskId: string): Promise<{ blockers: TaskBlocker[] }> => {
    const response = await api.get<{ blockers: TaskBlocker[] }>(
      `/tasks/${taskId}/blockers`
    )
    return response.data
  },

  resolve: async (blockerId: string, daysBlocked: number): Promise<TaskBlocker> => {
    const response = await api.put<TaskBlocker>(`/blockers/${blockerId}/resolve`, {
      days_blocked: daysBlocked,
    })
    return response.data
  },
}

// Review API
export const reviewApi = {
  create: async (
    taskId: string,
    data: CreateReviewRequest
  ): Promise<TaskReview> => {
    const response = await api.post<TaskReview>(`/tasks/${taskId}/review`, data)
    return response.data
  },

  get: async (taskId: string): Promise<TaskReview | null> => {
    try {
      const response = await api.get<TaskReview>(`/tasks/${taskId}/review`)
      return response.data
    } catch {
      return null
    }
  },
}

// Team API
export const teamApi = {
  create: async (data: CreateTeamRequest): Promise<TeamResponse> => {
    const response = await api.post<TeamResponse>('/teams', data)
    return response.data
  },

  get: async (id: string): Promise<TeamResponse> => {
    const response = await api.get<TeamResponse>(`/teams/${id}`)
    return response.data
  },

  update: async (id: string, data: UpdateTeamRequest): Promise<TeamResponse> => {
    const response = await api.put<TeamResponse>(`/teams/${id}`, data)
    return response.data
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/teams/${id}`)
  },

  list: async (): Promise<TeamListResponse> => {
    const response = await api.get<TeamListResponse>('/teams')
    return response.data
  },

  addMember: async (teamId: string, data: AddTeamMemberRequest): Promise<void> => {
    await api.post(`/teams/${teamId}/members`, data)
  },

  listMembers: async (teamId: string): Promise<TeamMemberListResponse> => {
    const response = await api.get<TeamMemberListResponse>(`/teams/${teamId}/members`)
    return response.data
  },

  updateMemberRole: async (teamId: string, userId: string, data: UpdateMemberRoleRequest): Promise<void> => {
    await api.put(`/teams/${teamId}/members/${userId}`, data)
  },

  removeMember: async (teamId: string, userId: string): Promise<void> => {
    await api.delete(`/teams/${teamId}/members/${userId}`)
  },

  listProjects: async (teamId: string): Promise<ProjectListResponse> => {
    const response = await api.get<ProjectListResponse>(`/teams/${teamId}/projects`)
    return response.data
  },
}

// Project API
export const projectApi = {
  create: async (data: CreateProjectRequest): Promise<ProjectResponse> => {
    const response = await api.post<ProjectResponse>('/projects', data)
    return response.data
  },

  get: async (id: string): Promise<ProjectResponse> => {
    const response = await api.get<ProjectResponse>(`/projects/${id}`)
    return response.data
  },

  update: async (id: string, data: UpdateProjectRequest): Promise<ProjectResponse> => {
    const response = await api.put<ProjectResponse>(`/projects/${id}`, data)
    return response.data
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/projects/${id}`)
  },

  list: async (params?: {
    status?: string
    team_id?: string
    search?: string
    page?: number
    page_size?: number
  }): Promise<ProjectListResponse> => {
    const response = await api.get<ProjectListResponse>('/projects', { params })
    return response.data
  },

  archive: async (id: string): Promise<ProjectResponse> => {
    const response = await api.post<ProjectResponse>(`/projects/${id}/archive`)
    return response.data
  },

  assignTeams: async (id: string, data: AssignTeamsRequest): Promise<ProjectResponse> => {
    const response = await api.post<ProjectResponse>(`/projects/${id}/teams`, data)
    return response.data
  },

  removeTeam: async (projectId: string, teamId: string): Promise<void> => {
    await api.delete(`/projects/${projectId}/teams/${teamId}`)
  },

  getTeams: async (id: string): Promise<TeamListResponse> => {
    const response = await api.get<TeamListResponse>(`/projects/${id}/teams`)
    return response.data
  },
}

// Auth API
export const authApi = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await api.post<AuthResponse>('/auth/login', data)
    return response.data
  },

  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    const response = await api.post<AuthResponse>('/auth/register', data)
    return response.data
  },

  getMe: async (): Promise<User> => {
    const response = await api.get<User>('/auth/me')
    return response.data
  },
}

// Users API
export const usersApi = {
  list: async (): Promise<{ users: User[]; total: number }> => {
    const response = await api.get<{ users: User[]; total: number }>('/users')
    return response.data
  },
}

export default api
