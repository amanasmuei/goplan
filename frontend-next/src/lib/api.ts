import axios from "axios";
import { useAuthStore } from "@/store/auth-store";
import type {
  AuthResponse,
  CreatePlanRequest,
  LoginRequest,
  PlanListResponse,
  PlanResponse,
  PlanVersion,
  RegisterRequest,
  RegenerateSectionRequest,
  RefineSectionRequest,
  SectionType,
  SectionVersion,
  StrategicPlan,
  Subscription,
  User,
} from "@/lib/types";

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1",
  headers: { "Content-Type": "application/json" },
});

// Attach JWT token to every request
api.interceptors.request.use((config) => {
  const { token, isTokenExpired } = useAuthStore.getState();
  if (token && !isTokenExpired()) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401 responses
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout();
    }
    return Promise.reject(error);
  },
);

// --- Auth API ---

export const authApi = {
  login: (data: LoginRequest) =>
    api.post<AuthResponse>("/auth/login", data).then((r) => r.data),

  register: (data: RegisterRequest) =>
    api.post<AuthResponse>("/auth/register", data).then((r) => r.data),

  getMe: () => api.get<User>("/auth/me").then((r) => r.data),
};

// --- Strategy API ---

export const strategyApi = {
  create: (data: CreatePlanRequest) =>
    api.post<PlanResponse>("/strategies", data).then((r) => r.data),

  list: (params?: { page?: number; page_size?: number; status?: string; category?: string; search?: string }) =>
    api.get<PlanListResponse>("/strategies", { params }).then((r) => r.data),

  getById: (id: string) =>
    api.get<PlanResponse>(`/strategies/${id}`).then((r) => r.data),

  archive: (id: string) =>
    api.delete(`/strategies/${id}`).then((r) => r.data),

  regenerateSection: (id: string, type: SectionType, data?: RegenerateSectionRequest) =>
    api.post<PlanResponse>(`/strategies/${id}/sections/${type}/regenerate`, data).then((r) => r.data),

  refineSection: (id: string, type: SectionType, data: RefineSectionRequest) =>
    api.post<PlanResponse>(`/strategies/${id}/sections/${type}/refine`, data).then((r) => r.data),

  listVersions: (id: string) =>
    api.get<PlanVersion[]>(`/strategies/${id}/versions`).then((r) => r.data),

  getVersion: (id: string, version: number) =>
    api.get<PlanResponse>(`/strategies/${id}/versions/${version}`).then((r) => r.data),

  listSectionVersions: (id: string, type: SectionType) =>
    api.get<SectionVersion[]>(`/strategies/${id}/sections/${type}/versions`).then((r) => r.data),

  getSimilar: (id: string) =>
    api.get<StrategicPlan[]>(`/strategies/${id}/similar`).then((r) => r.data),

  exportPlan: (id: string, format: "pdf" | "docx" | "markdown") =>
    api.get(`/strategies/${id}/export`, { params: { format }, responseType: "blob" }).then((r) => r.data),
};

// --- Subscription API ---

export const subscriptionApi = {
  get: () =>
    api.get<Subscription>("/subscription").then((r) => r.data),

  upgrade: (tier: string) =>
    api.post<Subscription>("/subscription/upgrade", { tier }).then((r) => r.data),
};

export default api;
