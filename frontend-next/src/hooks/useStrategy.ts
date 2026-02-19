import { useMutation, useQuery, useQueryClient, type UseQueryOptions } from "@tanstack/react-query";
import { strategyApi } from "@/lib/api";
import type {
  CreatePlanRequest,
  PlanResponse,
  RefineSectionRequest,
  RegenerateSectionRequest,
  SectionType,
} from "@/lib/types";

export function useStrategies(params?: {
  page?: number;
  page_size?: number;
  status?: string;
  category?: string;
  search?: string;
}) {
  return useQuery({
    queryKey: ["strategies", params],
    queryFn: () => strategyApi.list(params),
  });
}

export function useStrategy(
  id: string,
  options?: Partial<UseQueryOptions<PlanResponse>>,
) {
  return useQuery({
    queryKey: ["strategy", id],
    queryFn: () => strategyApi.getById(id),
    enabled: !!id,
    ...options,
  });
}

export function useCreateStrategy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreatePlanRequest) => strategyApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["strategies"] });
    },
  });
}

export function useArchiveStrategy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => strategyApi.archive(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["strategies"] });
    },
  });
}

export function useRegenerateSection(planId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      sectionType,
      data,
    }: {
      sectionType: SectionType;
      data?: RegenerateSectionRequest;
    }) => strategyApi.regenerateSection(planId, sectionType, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["strategy", planId] });
    },
  });
}

export function useRefineSection(planId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      sectionType,
      data,
    }: {
      sectionType: SectionType;
      data: RefineSectionRequest;
    }) => strategyApi.refineSection(planId, sectionType, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["strategy", planId] });
    },
  });
}

export function useExportStrategy() {
  return async (id: string, format: "pdf" | "docx" | "markdown") => {
    const blob = await strategyApi.exportPlan(id, format);
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `strategy.${format === "markdown" ? "md" : format}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };
}

export function useSimilarStrategies(id: string) {
  return useQuery({
    queryKey: ["strategy", id, "similar"],
    queryFn: () => strategyApi.getSimilar(id),
    enabled: !!id,
  });
}

export function usePlanVersions(id: string) {
  return useQuery({
    queryKey: ["strategy", id, "versions"],
    queryFn: () => strategyApi.listVersions(id),
    enabled: !!id,
  });
}

export function useSectionVersions(planId: string, sectionType: SectionType) {
  return useQuery({
    queryKey: ["strategy", planId, "sections", sectionType, "versions"],
    queryFn: () => strategyApi.listSectionVersions(planId, sectionType),
    enabled: !!planId && !!sectionType,
  });
}
