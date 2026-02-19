import { useQuery } from "@tanstack/react-query";
import { subscriptionApi } from "@/lib/api";

export function useSubscription() {
  return useQuery({
    queryKey: ["subscription"],
    queryFn: () => subscriptionApi.get(),
  });
}
