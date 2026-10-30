# Architecture

How resourcerer is put together and what happens under the hood: the components, the `Recommendation` resource,
how a recommendation is computed, priced, checked and applied, and how the UI gets its data.

## The big picture

resourcerer is one binary that runs as three components, each its own Deployment in the chart. They never talk to
each other directly: they coordinate through the Kubernetes API, around one custom resource per workload, the
`Recommendation`. Usage history comes from a Prometheus-compatible API that already exists in the cluster
(Prometheus, Thanos, Mimir…); resourcerer keeps no database.

```mermaid
flowchart LR
    subgraph sources[Data sources]
        prom[(Prometheus / Thanos<br/>cAdvisor, kube-state-metrics,<br/>Karpenter, Istio…)]
    end

    subgraph k8s[Kubernetes API]
        wl[Workloads<br/>Deployment · StatefulSet<br/>DaemonSet · Rollout]
        hpa[HPAs]
        cm[defaults ConfigMap<br/>settings · cost · reliability]
        rec[Recommendation<br/>one per workload]
        ann[namespace / workload<br/>annotations]
    end

    controller[controller<br/>recommender]
    applier[applier]
    server[server<br/>HTTP API + UI]
    ui[Browser UI]

    prom -->|usage, requests, prices| controller
    prom -->|live checks, charts| server
    wl --> controller
    hpa --> controller
    cm --> controller
    ann --> controller
    controller -->|writes status| rec
    rec --> applier
    applier -->|server-side apply<br/>resources, HPA target| wl
    applier -->|history, result| rec
    rec --> server
    server -->|spec.applyRequest<br/>spec.regenerateRequestedAt| rec
    server -->|settings annotations| ann
    ui <-->|/api/v1| server
    controller -.->|resourcerer_recommendation<br/>resourcerer_optimization_score| prom
```

| Component | Watches | Writes | Leader election |
|---|---|---|---|
| **controller** (`internal/recommender`) | Workloads, their HPAs, namespaces, the defaults ConfigMap, Recommendations | `Recommendation` objects and their `status`; Prometheus metrics | yes |
| **applier** (`internal/applier`) | Recommendations | Workload pod templates and HPA targets (server-side apply, field manager `resourcerer`); `status.history`, `status.lastApplyResult` | yes |
| **server** (`internal/server`) | Recommendations, workloads, namespaces, pods (through the cache) | `spec.applyRequest` and `spec.regenerateRequestedAt` on Recommendations; settings annotations on namespaces and workloads | no: every replica serves |

`resourcerer all` runs the three in one process (`make run`). The applier is the only component that changes
workloads, so a read-only install leaves it out and starts the server with `--read-only`.

## The `Recommendation` resource

One per managed workload, named `<kind>-<name>` in the workload's namespace and owned by it (deleted with it).

```mermaid
flowchart LR
    subgraph spec[spec: requests]
        t[targetRef<br/>apiVersion, kind, name]
        ar[applyRequest<br/>id, action, hash, requestedBy]
        rg[regenerateRequestedAt]
    end
    subgraph status[status: answers]
        ph[phase, message, hash]
        ct[containers<br/>current, recommended, basis]
        hp[hpa target]
        st[settings<br/>value and source per key]
        im[impact<br/>totals, monthly cost]
        pr[pricing<br/>allocation, factor, coverage]
        ev[evidence<br/>replay of usage]
        rl[reliability<br/>built-in and declared signals]
        hi[history<br/>last 10 applies and reverts]
        lr[lastApplyResult]
    end
    controller[controller] --> ph & ct & hp & st & im & pr & ev & rl
    applier[applier] --> hi & lr
    server[server] --> ar & rg
```

- **`spec`** is the request side: what to optimize, and requests from the UI (apply, revert, regenerate).
- **`status`** is the answer, written by the controller (recommendation, pricing, evidence, reliability) and the
  applier (history, result).
- **`status.hash`** identifies the recommended values. An apply carries the hash the user reviewed; the applier
  refuses it if the recommendation changed since.

## Which workloads are managed

- `--selection all` (default): every Deployment, StatefulSet, DaemonSet and Argo Rollout, except in excluded
  namespaces and those labeled `resourcerer.io/enabled=false`. `--selection label`: only those labeled
  `resourcerer.io/enabled=true` (on the workload or its namespace).
- A Deployment used by a Rollout through `spec.workloadRef` gets no Recommendation of its own: it is optimized
  through the Rollout, and applies patch that Deployment's template (`status.templateRef`).
- Turning a workload off deletes its Recommendation.

## Settings

Every behaviour that can vary per workload is a setting (`internal/settings`): optimization goals, learning period,
schedule, auto-deploy and its threshold, limits policies and ratios, minimum requests, containers to skip, HPA
handling. They resolve per key, the most specific level winning:

```mermaid
flowchart LR
    builtin[Built-in default] --> values[Chart values<br/>defaults ConfigMap]
    values --> ns[Namespace annotation<br/>resourcerer.io/KEY]
    ns --> wl[Workload annotation<br/>resourcerer.io/KEY]
    wl --> eff[Effective setting<br/>value + source in status]
```

An invalid value is skipped with a warning in the status, so the previous level wins. A change to any level
regenerates the affected recommendations (the controller watches namespaces and the ConfigMap).

## How a recommendation is computed

The controller reconciles each workload kind with its own controller. A reconcile is cheap until the workload is
**due**: its schedule slot passed, a regeneration was requested, its settings or current requests changed, or the
last run failed. Slots are spread over the schedule period by hashing the workload's name, so a cluster does not
query Prometheus all at once.

```mermaid
sequenceDiagram
    autonumber
    participant K as Kubernetes API
    participant C as controller
    participant P as Prometheus
    participant E as engine (pure)

    K->>C: workload / namespace / ConfigMap / Recommendation event
    C->>K: read workload, namespace, Recommendation, defaults ConfigMap
    C->>C: resolve settings, resolve cost pricing
    alt not due
        C-->>C: requeue at the next slot
    else due
        C->>P: busiest pod's CPU and memory per container (learning period)
        C->>P: pod count, CPU throttling, OOM and HPA context
        C->>E: Recommend(settings, containers, HPA, history)
        E-->>C: phase, requests, limits, HPA target, basis
        C->>E: Summarize (totals, monthly cost, score)
        opt cost.allocation: node
            C->>P: where the pods ran, at which node prices
        end
        C->>E: Replay (usage against the recommendation)
        C->>P: declared reliability signals
        C->>K: patch status
        C-->>C: export metrics, event if the hash changed, requeue at the next slot
    end
```

### The engine (`internal/engine`)

Pure functions over usage series, so every rule is tested with synthetic data.

- **Usage** is the busiest pod per container at each step: the pod that would be throttled or OOM-killed first.
- **Requests** are a percentile of that usage plus a margin, per optimization goal:

  | Goal | CPU | Memory |
  |---|---|---|
  | Savings | p85 + 5% | p99 + 10% |
  | Balanced | p95 + 10% | max + 15% |
  | Reliability | p99 + 20% | max + 30% |

  then raised to the minimum (`cpu.requests.min`, `memory.requests.min`) and rounded.
- **OOM kills** cap the observed memory at the old limit, so the memory recommendation is raised to at least 1.2×
  that limit.
- **Limits** follow the limits policy: keep them (`RequestsOnly`), raise them only if the request would pass them
  (`RequestsRaiseLimitsIfNeeded`), set them at a ratio of the request (`RequestsAndLimits`), or leave the resource
  alone (`DoNotOptimize`). A CPU limit throttling at least 25% of CFS periods is raised by 1.5×, since throttling
  happens in 100ms periods that sampled usage does not show.
- **HPAs**: an HPA scaling on both CPU and memory makes the workload `Unsupported` (changing requests would change
  replica counts). With one on CPU, the target utilization is adjusted so the HPA keeps scaling at the same absolute
  CPU per pod (`hpa.preserve-scale-point`), clamped to 50–90%.
- **Phase**:

```mermaid
stateDiagram-v2
    [*] --> Pending
    Pending --> Learning: enough samples
    Learning --> Ready: history covers 95% of the learning period
    Pending --> Unsupported: HPA on CPU and memory
    Learning --> Warning: scaled to zero
    Ready --> Warning: scaled to zero
    Warning --> Ready: pods back
    Pending --> Error: metrics unavailable
    Learning --> Error: metrics unavailable
    Ready --> Error: metrics unavailable
    Error --> Ready: next run succeeds
```

  Only `Ready` and `Learning` recommendations can be applied by hand; auto-deploy needs `Ready`.
- **Pending changes**: the recommendation counts as pending when a request moves by more than
  `auto-deploy.threshold` (and a minimum absolute amount), so tiny drifts do not trigger rollouts.

## How costs are priced

Cost is requests × average replicas × hourly rates × 730. The rates come from the `cost` values and the
`Pricer` (`internal/cluster/pricing.go`), and with node allocation each workload gets its own rates.

```mermaid
flowchart TD
    rates[cost.rates<br/>CPU core-hour, GiB-hour] --> src{cost.source}
    src -->|static| static[rates as configured]
    src -->|karpenter / auto| pricer[Pricer: calibrate to Karpenter<br/>node price ÷ capacity at the rates]
    pricer --> cluster[cluster rates<br/>cluster factor k]
    static --> cluster
    cluster --> alloc{cost.allocation}
    alloc -->|cluster| wrates[workload rates = cluster rates]
    alloc -->|node| place[node factor per node<br/>weighted by time and requests<br/>over the learning period]
    place --> wrates2[workload rates = configured × F<br/>unpriced requests at k]
    wrates --> monthly[monthly cost now and recommended]
    wrates2 --> monthly
```

- **Calibration** scales the configured rates so the priced Karpenter nodes cost what Karpenter's price estimates
  say, keeping the configured CPU-to-memory ratio (node prices do not say how much of them is CPU). It refreshes every
  `cost.karpenter.refreshInterval`.

```mermaid
stateDiagram-v2
    [*] --> ok: calibration succeeds
    [*] --> fallback: auto without Karpenter prices
    [*] --> error: karpenter without prices
    ok --> stale: refresh fails, last calibration kept
    stale --> ok: refresh succeeds (retried after 1m)
    stale --> error: karpenter, 3 failures in a row
```

- **Node allocation** values each pod's requests at the configured rates, multiplies them by the factor of the node
  it ran on (node price ÷ the node's capacity at the same rates; reserved capacity priced as on-demand), and averages
  over the learning period (at most 7 days, 336 points). Requests on nodes without a price count at the cluster
  factor. The result, with the capacity mix and top instance types, is stored in `status.pricing`, and the same
  factor prices both the current and the recommended requests, so savings stay comparable.
- **Placement** (`GET .../placement`) reuses the same queries to show where the workload ran over time, by capacity
  type and instance type, computed on request and cached for 5 minutes.

## Reliability

Signals that requests or limits are too tight, or that the workload is unhealthy for other reasons.

- **Built-in**: CPU throttling, memory peak against the limit, OOM kills, restarts, crash loops, HPA at its maximum,
  BestEffort pods. Computed from Prometheus and pod status.
- **Declared**: PromQL templates in the chart values, evaluated per workload over its learning period
  (`internal/signals`, `docs/reliability-signals.md`). A signal with no data does not apply to the workload; one that
  fails is skipped and never blocks the recommendation.

The controller stores the learning-period view in `status.reliability`, which drives the workload list and the
verdict. The server checks the same signals live (`GET .../reliability?window=15m|1h|6h|1d|learning`), so a fix shows
up as resolved within minutes instead of when old samples leave the learning period.

## Evidence and the verdict

At each regeneration the engine replays the usage history against the recommended requests (`engine.Replay`): the
busiest pod's peak and the share of samples above the recommendation, per resource, and the p95 usage × replicas
priced like the requests (what the workload actually needs). It is stored in `status.evidence`.

The Impact tab turns phase, history coverage, evidence and open reliability signals into a verdict
(`web/src/verdict.ts`): "Safe to apply", "Review before applying", "Apply to fix reliability", "Already
right-sized", or a waiting or error state. Signals the recommendation itself fixes (OOM kills, CPU throttling from
25%, where it raises the limit) count for applying; others ask for a review.

## How a recommendation is applied

```mermaid
sequenceDiagram
    autonumber
    actor U as User
    participant S as server
    participant K as Kubernetes API
    participant A as applier
    participant W as Deployment / StatefulSet / Rollout controller

    U->>S: POST .../apply {hash}
    S->>K: set spec.applyRequest {id, Apply, hash, requestedBy}
    K-->>A: Recommendation changed
    A->>A: request id not handled yet?
    alt hash differs from status.hash
        A->>K: status.lastApplyResult = Rejected
    else phase not Ready or Learning
        A->>K: status.lastApplyResult = Rejected
    else
        A->>K: read live resources (API server, not cache)
        A->>K: server-side apply pod template resources<br/>field manager "resourcerer", force
        opt HPA on CPU
            A->>K: patch HPA target utilization
        end
        A->>K: status.history += {Apply, from, to}, lastApplyResult = Succeeded, event
        K-->>W: template changed
        W->>W: rolling update
    end
```

- **Auto-deploy**: with `auto-deploy: Enabled`, the applier applies a `Ready`, pending recommendation as
  `auto-deploy`, once per hash; a hash already in the history (applied, or reverted by someone) is never re-applied.
- **Revert** applies the `from` values of the last apply, through the same path.
- **Server-side apply** with the `resourcerer` field manager owns only the resources fields it sets, so GitOps tools
  can ignore them by manager instead of fighting over them.
- The applier remembers handled request IDs and auto-applied hashes in memory, because right after a status patch
  the cache can still serve the old object.

## The server and the UI

```mermaid
flowchart LR
    subgraph browser[Browser]
        rq[React Query<br/>polling per view]
        pages[Overview · Workloads · Workload tabs · Settings]
    end
    subgraph srv[server]
        api[/api/v1 handlers/]
        cache[(in-memory cache<br/>per key and TTL)]
    end
    pages --> rq --> api
    api -->|lists, details| k8scache[(controller-runtime cache)]
    api -->|charts, live checks, placement| cache --> prom[(Prometheus)]
    api -->|apply, revert, regenerate, settings| kapi[Kubernetes API]
```

- **Endpoints**: `info` (cluster, settings defaults, cost and reliability configuration, Prometheus status),
  `workloads` (list), `workloads/{ns}/{kind}/{name}` (detail) and its `series`, `impact`, `reliability` and
  `placement`, `cluster/series`, and the writes `apply`, `revert`, `regenerate` and `settings`.
- **Lists and details** come from the Recommendations in the cache: no Prometheus query, so they stay fast with
  thousands of workloads.
- **Charts and live checks** query Prometheus on request, cached per key for a short time; concurrent identical
  requests share one query.
- **Writes** go through `s.write`, which refuses them in read-only mode. Authentication is anonymous for now: keep
  the Service internal or put an authenticating proxy in front.
- **The UI** is embedded in the binary (`web.FS()`) and served as a single-page app. React Query polls the workload
  every 5s, the list every 15s and charts every 30s; expensive views fetch only when shown.

## Deployment

```mermaid
flowchart TB
    subgraph ns[namespace resourcerer]
        dc[Deployment controller<br/>leader-elected]
        da[Deployment applier<br/>leader-elected]
        ds[Deployment server<br/>all replicas serve]
        svc[Service → http :8080]
        cmd[ConfigMap defaults]
        sm[ServiceMonitor :8081]
    end
    crd[CRD recommendations.resourcerer.io]
    ds --- svc
    dc --- cmd
    sm -.-> dc
```

- **RBAC** follows the split: the controller reads workloads and writes Recommendations; the applier patches
  workload templates and HPAs; the server reads, and writes Recommendation specs and settings annotations unless
  read-only.
- **Cache footprint**: only the defaults ConfigMap's namespace is cached for ConfigMaps, and managed fields are
  stripped from cached objects, since caching every ConfigMap in a large cluster costs hundreds of MiB.
- **Metrics**: `resourcerer_recommendation{namespace, workload_kind, workload, container, resource, type}` and
  `resourcerer_optimization_score`, scraped through the ServiceMonitor; the charts read recommendation history from
  them.

## Where this is going

The roadmap is in the design issues: reliability signals as custom resources (#28), `ApplyPolicy` for windows,
freezes and verified auto-deploy (#29), `OptimizationPolicy` for settings by selector (#30), `CostProfile` for Savings
Plans and Reserved Instances (#27), and decision models (Jev / decider) behind the verdict (#26).
