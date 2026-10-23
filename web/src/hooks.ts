import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, type Mode, type WorkloadRef } from "./api";

const keys = {
  info: ["info"] as const,
  workloads: ["workloads"] as const,
  workload: (w: WorkloadRef) => ["workload", w.namespace, w.kind, w.name] as const,
};

export const useInfo = () => useQuery({ queryKey: keys.info, queryFn: api.info, staleTime: Infinity });

export const useWorkloads = () => useQuery({ queryKey: keys.workloads, queryFn: api.list, refetchInterval: 15_000 });

export const useWorkload = (w: WorkloadRef) =>
  useQuery({ queryKey: keys.workload(w), queryFn: () => api.get(w), refetchInterval: 5_000 });

export function useSetMode(w: WorkloadRef) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (mode: Mode) => api.setMode(w, mode),
    onSuccess: (data) => {
      qc.setQueryData(keys.workload(w), data);
      qc.invalidateQueries({ queryKey: keys.workloads });
    },
  });
}
