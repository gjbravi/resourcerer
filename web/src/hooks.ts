import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState, useSyncExternalStore } from "react";
import { api, refKey, type Range, type Ref, type ReliabilityWindow, type Resource } from "./api";

export const useInfo = () => useQuery({ queryKey: ["info"], queryFn: api.info, refetchInterval: 60_000 });
export const useWorkloads = () => useQuery({ queryKey: ["workloads"], queryFn: api.list, refetchInterval: 15_000 });
export const useWorkload = (w: Ref) =>
  useQuery({ queryKey: ["workload", refKey(w)], queryFn: () => api.get(w), refetchInterval: 5_000 });
export const useSeries = (w: Ref, container: string, resource: Resource, range: Range) =>
  useQuery({
    queryKey: ["series", refKey(w), container, resource, range],
    queryFn: () => api.series(w, container, resource, range),
    refetchInterval: 30_000,
    placeholderData: keepPreviousData,
  });
export const useImpact = (w: Ref, resource: Resource, range: Range) =>
  useQuery({
    queryKey: ["impact", refKey(w), resource, range],
    queryFn: () => api.impact(w, resource, range),
    refetchInterval: 30_000,
    placeholderData: keepPreviousData,
  });
/** The placement covers the whole learning period and the server caches it for 5 minutes. */
export const usePlacement = (w: Ref, enabled = true) =>
  useQuery({ queryKey: ["placement", refKey(w)], queryFn: () => api.placement(w), staleTime: 300_000, enabled });
/** Live mode refreshes every 30s, about the scrape interval, so a fix shows up as soon as Prometheus has it. */
export const useReliability = (w: Ref, window: ReliabilityWindow, live: boolean) =>
  useQuery({
    queryKey: ["reliability", refKey(w), window],
    queryFn: () => api.reliability(w, window),
    refetchInterval: live ? 30_000 : false,
    placeholderData: keepPreviousData,
  });

/** The current time, ticking every second, for "updated 12s ago" labels. */
export function useNow(): number {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const t = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(t);
  }, []);
  return now;
}

export const useClusterSeries = (resource: Resource, range: Range) =>
  useQuery({
    queryKey: ["cluster-series", resource, range],
    queryFn: () => api.clusterSeries(resource, range),
    refetchInterval: 60_000,
    placeholderData: keepPreviousData,
  });

/** Mutations refresh the workload and the list; results show up as the controller and applier act. */
export function useAction<T>(w: Ref, fn: (arg: T) => Promise<unknown>) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: fn,
    onSettled: () => {
      qc.invalidateQueries({ queryKey: ["workload", refKey(w)] });
      qc.invalidateQueries({ queryKey: ["workloads"] });
    },
  });
}

export function useHash(): string {
  const [hash, setHash] = useState(() => window.location.hash || "#/overview");
  useEffect(() => {
    // Switching tabs keeps the page where it is; only a different page starts at the top.
    const page = (h: string) => h.replace(/^(#\/w\/[^/]+\/[^/]+\/[^/?]+)\/[^/?]*/, "$1");
    let prev = window.location.hash || "#/overview";
    const on = () => {
      const next = window.location.hash || "#/overview";
      // The browser's scroll anchoring would otherwise follow the cards below the tab content.
      const y = page(next) === page(prev) ? window.scrollY : 0;
      prev = next;
      setHash(next);
      requestAnimationFrame(() => window.scrollTo(0, y));
    };
    window.addEventListener("hashchange", on);
    return () => window.removeEventListener("hashchange", on);
  }, []);
  return hash;
}

export const navigate = (h: string) => {
  window.location.hash = h;
};

export const workloadHash = (w: Ref, tab = "recommendation") =>
  `#/w/${encodeURIComponent(w.namespace)}/${w.kind}/${encodeURIComponent(w.name)}/${tab}`;

/* Toasts: a tiny external store so any component can raise one. */
export interface Toast {
  id: number;
  kind: "ok" | "info" | "error";
  title: string;
  body?: string;
  action?: { label: string; run: () => void };
}
let toasts: Toast[] = [];
const listeners = new Set<() => void>();
const emit = () => listeners.forEach((l) => l());
export function toast(t: Omit<Toast, "id">) {
  const id = Date.now() + Math.random();
  toasts = [...toasts, { ...t, id }];
  emit();
  setTimeout(() => dismiss(id), 7000);
}
export function dismiss(id: number) {
  toasts = toasts.filter((t) => t.id !== id);
  emit();
}
export const useToasts = () =>
  useSyncExternalStore(
    (l) => {
      listeners.add(l);
      return () => listeners.delete(l);
    },
    () => toasts,
  );
