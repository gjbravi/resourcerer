---
updatedAt: 2026-10-28T08:41:16.000Z
---

# AtlasPilot Assistant

An AI-powered assistant that answers operational questions about your Kubernetes clusters, costs, and database optimization in plain English.

AtlasPilot Assistant is an AI-powered assistant embedded in the Atlas console. Ask operational questions about your Kubernetes clusters in plain English and get structured, actionable answers, without writing `kubectl` commands, navigating multiple dashboards, or running SQL queries.

Questions that previously required switching between dashboards and tools can be handled in a single interaction. AtlasPilot Assistant understands your live cluster state, cost and savings data, workload events, audit logs, cluster health scores, and Atlas documentation. It combines information from multiple data sources into a single response.

Beyond answering questions, AtlasPilot Assistant can generate and maintain Workload Autoscaler scaling policies for you. Describe your environment and reliability targets, and AtlasPilot Assistant analyzes your cluster and produces a complete set of policies and assignment rules tailored to what it finds. See [Generate scaling policies with AtlasPilot Assistant](woop-scaling-policies-atlaspilot-assistant.md).

You can also connect AI tools, such as Claude Code or Cursor, to your Atlas data with [Atlas MCP](atlaspilot-mcp.md), available in private preview.

## Access AtlasPilot Assistant

AtlasPilot Assistant is available to all Atlas customers directly in the console. No separate configuration, installation, or onboarding required.

1. Log in to the [Atlas console](https://console.atlas.example).
2. Locate the **AtlasPilot Assistant** chat interface in the console header. It is accessible from any page.
3. Type your question in the input field and press **Enter**.

AtlasPilot Assistant authenticates you automatically using your existing Atlas session.

![](https://files.readme.io/f1cc3750c97e01ef34b9698490023808fcb44c6f1b73e0f6fc49bb1bf8ce8b03-image.png)

## How AtlasPilot Assistant works

When you ask a question, AtlasPilot Assistant analyzes your intent and routes the query to the right data sources. It can query multiple sources in parallel and combine the results into a single, coherent response.

AtlasPilot Assistant covers four main data domains:

* **Live cluster state**: pods, nodes, services, deployments, statefulsets, daemonsets, HPAs, namespaces, workload events, and audit logs. AtlasPilot Assistant queries your real-time cluster data and applies filtering, counting, and analysis. When kubectl logs access is enabled, AtlasPilot Assistant can also retrieve live container logs.
* **Cost and savings analytics**: cluster costs, namespace-level cost breakdowns, workload-level costs, savings data, and optimization recommendations. Cost data updates at 5-minute intervals and can include trend charts.
* **Database optimization (DBO)**: DBO deployment health, cache hit rates, query-level performance, connection pooling status, and optimization activity. Available when DBO is enabled on your cluster.
* **Atlas documentation**: how-to guides, feature explanations, and configuration references from the Atlas docs. AtlasPilot Assistant retrieves the most relevant content and synthesizes a targeted answer with source citations.

For multi-topic questions like *"show me expensive pods with their resource requests,"* AtlasPilot Assistant queries both the cost analytics and cluster state data simultaneously, then merges the results into a combined response.

## What you can ask

The table below summarizes what AtlasPilot Assistant can answer and the kind of question each capability handles.

| Capability                       | What you can ask about                                                                                                                                                                                                                                                                                                     | Example                                                                                    |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| **Cluster state**                | Pods, nodes, deployments, statefulsets, daemonsets, services, HPAs, and namespaces. Filter by namespace, status, label, or restart count.                                                                                                                                                                                  | *"Show me all pods in CrashLoopBackOff in the payments namespace."*                        |
| **Container logs**               | Live container logs fetched directly from your cluster. Requires kubectl logs access to be enabled (see [Enable kubectl logs](#enable-kubectl-logs)).                                                                                                                                                                      | *"Show me the last 50 log lines from the api-gateway container."*                          |
| **Workload events & audit logs** | OOM kills, traffic surges, memory pressure, rightsizing activity, and policy changes. Time ranges work in natural language.                                                                                                                                                                                                | *"Which containers were OOM killed in the last 12 hours?"*                                 |
| **Cost and savings**             | Cluster, namespace, or workload-level costs; trend data over custom periods; savings opportunities. Cost data updates every 5 minutes.                                                                                                                                                                                     | *"Cost breakdown by namespace for the last 30 days, ordered by spend."*                    |
| **Health score & optimization**  | Score breakdown across version management, instance diversity, spot usage, resource requests, autoscaling, security, and cost efficiency.                                                                                                                                                                                  | *"Why is my health score low and what should I fix first?"*                                |
| **Database optimization (DBO)**  | DBO deployment status, cache hit rates, slow or low-hit-rate queries, connection pool usage, and optimization activity. Requires DBO to be enabled on your cluster.                                                                                                                                                        | *"What's my overall cache hit rate?"* or *"Is DBO working on my cluster?"*                 |
| **Documentation**                | Atlas feature explanations, how-to guidance, and configuration references, with source citations.                                                                                                                                                                                                                        | *"What annotations does the workload autoscaler support?"*                                 |
| **Scaling policy generation**    | Analyze your cluster and generate, update, and maintain Workload Autoscaler scaling policies and assignment rules based on your intent. See [Generate scaling policies with AtlasPilot Assistant](woop-scaling-policies-atlaspilot-assistant.md).                                                                           | *"Generate policies for this cluster prioritizing savings, for non-production workloads."* |
| **Recommendation explanations**  | Ask why Workload Autoscaler recommended specific resource values for a workload. AtlasPilot Assistant surfaces the contributing factors (traffic surges, OOM kills, CPU stall, or other adjustments that shaped the recommendation), so you don't have to correlate events manually. Name the workload you're asking about. | *"Why did the workload autoscaler recommend 2 CPU for the payments-api?"*                  |

AtlasPilot Assistant can also handle questions that span multiple areas simultaneously. *"Which nodes are underutilized and what would I save by removing them?"* draws on cluster state and savings data in a single response.

<Callout icon="📘" theme="info">
  ### Documentation index freshness

  The documentation index updates periodically. Very recent documentation changes may take up to 24 hours to appear in AtlasPilot Assistant answers.
</Callout>

## Work with cluster context

AtlasPilot Assistant determines which cluster your question is about automatically using these signals (in priority order):

1. **Explicit mention**: if you name a cluster in your question, AtlasPilot Assistant uses that cluster.
2. **"Current cluster" or "this cluster"**: uses the cluster on the console page you're currently viewing.
3. **Conversation context**: if you asked about a cluster earlier in the session, subsequent questions default to it.
4. **Page context**: for the first question in a session, the cluster dashboard you're viewing is used.

To switch clusters mid-conversation, mention the new cluster by name. All subsequent questions will use the new cluster until you switch again.

If you're viewing one cluster's dashboard but your conversation is about a different cluster, AtlasPilot Assistant displays a hint to prevent wrong-cluster mistakes. Minor typos in cluster names are corrected automatically with a confirmation note.

### Follow-up questions

AtlasPilot Assistant maintains conversation context within a session, so follow-up questions work naturally:

1. Ask an initial question: *"Show me OOM kill events in the last 24 hours."*
2. Follow up with: *"Which workloads were affected most?"*
3. Drill deeper: *"What are the memory limits on those workloads?"*

## Enable kubectl logs

By default, AtlasPilot Assistant does not have access to live container logs. To let AtlasPilot Assistant fetch logs directly from your cluster, enable the cluster proxy on Sentinel:

```shell
helm upgrade -i atlas-sentinel atlas/atlas-sentinel \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set controller.extraArgs.cluster-proxy-enabled=true
```

Once enabled, you can ask questions like *"Show me the last 100 log lines from the payments-api container"* and AtlasPilot Assistant will retrieve them live from your cluster.

<Callout icon="⚠️" theme="warn">
  ### Security note

  Enabling the cluster proxy allows AtlasPilot Assistant to read container logs on your behalf. All requests are authenticated using your Atlas session and scoped to your organization. Logs are not stored. They are retrieved on demand and returned within your session only.
</Callout>

## Interpret results

AtlasPilot Assistant returns structured responses designed to help you take action, not just read data. Results land in different formats depending on what you asked:

* **Tables** give you scannable lists of resources, cost breakdowns, and comparisons, sorted and filtered so the most relevant rows surface first.
* **Charts** render cost trends and time-series data inline in the console, giving you the context to understand whether a cost spike is an anomaly or a pattern.
* **Console links** take you directly to the relevant page to act on what AtlasPilot Assistant found (for example, to apply a recommendation or inspect a specific workload). For the Q\&A assistant, which is read-only, these links are how you get from answer to action.
* **Recommendations** surface optimization suggestions with estimated savings impact when your question is about cost or efficiency.
* **Source citations** accompany documentation-based answers with links to the relevant pages on [docs.atlas.example](https://docs.atlas.example).

## Use cases

### Incident investigation

When pods fail or latency increases, you typically need to check workload events, review recent changes in the audit log, and cross-reference with cluster state, a process that can take 15–20 minutes across multiple dashboards.

With AtlasPilot Assistant, ask a single question like *"What workload optimization events happened in the api-gateway namespace in the last 24 hours, and did any cause OOM kills?"* and get the full picture: the change that was made, the events that followed, and the correlation between them.

### Cost attribution

Engineering or finance teams often need cost breakdowns by namespace or workload to allocate cloud costs back to product teams. Instead of exporting data and building spreadsheets, ask AtlasPilot Assistant *"Give me a cost breakdown by namespace for the last 30 days across all clusters, ordered by total cost"* and get a formatted result immediately. Follow up with *"Which workload in the analytics namespace is most expensive?"* to drill deeper.

### Weekly reporting

Summarizing autoscaling activity or cost trends for stakeholders typically requires exporting event data, filtering, and building a summary. AtlasPilot Assistant handles this in a single question: *"How many autoscaling events happened last month and how many traffic surge events occurred?"*

### Onboarding new engineers

New team members can ask AtlasPilot Assistant about Atlas features and get targeted answers without leaving their workflow. Questions like *"How does spot instance management work?"* or *"What policies control autoscaling?"* return synthesized answers drawn from Atlas documentation, with source links for further reading.

### Understanding workload recommendations

When Workload Autoscaler recommends resource values that seem surprising (for example, a higher CPU request than you'd expect for a quiet workload), the reason often lies in events the autoscaler already responded to: a traffic surge, an OOM kill, CPU stall, or a startup failure. Correlating those events to the recommendation manually means digging through the [Event log](event-log.md) and reasoning across entries.

With AtlasPilot Assistant, ask about a specific workload and get the explanation directly, including the events that shaped the recommendation: *"Why did the workload autoscaler recommend 2 CPU for the payments-api?"* AtlasPilot Assistant identifies the contributing factors and explains how each one adjusted the recommendation, so you can confirm the behavior is expected or spot something to investigate further.

## Data security and privacy

AtlasPilot Assistant enforces strict data isolation and privacy:

* **Organization-level isolation**: every query is scoped to your organization. Data isolation is enforced at the database level, not just in the application layer. There is no path to access another organization's data.
* **Input validation**: all queries pass through multiple validation layers before processing, including pattern-based detection and machine-learning-based scanning for prompt injection attempts.
* **Response sanitization**: internal identifiers like cluster UUIDs are resolved to friendly names in responses. Internal system details are never surfaced.
* **No model training on your data**: your cluster data is never used to train AI models. AtlasPilot Assistant queries your data in real time and returns results within your session.

## Current limitations

| Limitation                  | Details                                                                                                                                                                                                                                                                                                                                                             |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Read-only assistant**     | AtlasPilot Assistant answers questions and surfaces recommendations but cannot apply changes or execute actions. The exception is guided scaling policy generation, where AtlasPilot Assistant creates and updates Workload Autoscaler scaling policies and assignment rules when you run it. For all other questions, use the provided console links to take action. |
| **Single organization**     | Cannot query across multiple organizations in one question.                                                                                                                                                                                                                                                                                                         |
| **Session timeout**         | Session context expires after 1 hour of inactivity. Start a new question with the full context (including the cluster name if needed) after a timeout.                                                                                                                                                                                                              |
| **Documentation freshness** | Very recent doc updates may take up to 24 hours to appear in answers.                                                                                                                                                                                                                                                                                               |
| **Event history window**    | Detailed workload events are available for the last 90 days. Audit log availability depends on retention settings.                                                                                                                                                                                                                                                  |
| **Chart export**            | Trend charts render in the console UI only. They cannot be exported or shared directly from AtlasPilot Assistant. Use screenshots as a workaround.                                                                                                                                                                                                                   |
| **No API access**           | AtlasPilot Assistant is available through the Atlas console only. There is no public API for programmatic access.                                                                                                                                                                                                                                                  |

## Troubleshooting

* **AtlasPilot Assistant can't find your cluster**: verify the cluster is connected and active in Atlas. Navigate to the cluster's dashboard and try again. Page context helps resolve the cluster. Check for typos in the cluster name.
* **Cost data looks different from the reports page**: cost data updates at 5-minute intervals. Very recent changes (last few minutes) may not be reflected yet. For historical data gaps, confirm the cluster was connected during the query period.
* **Container logs are unavailable**: kubectl logs access requires the Sentinel cluster proxy to be enabled. Run the Helm upgrade command in [Enable kubectl logs](#enable-kubectl-logs). After upgrading, allow a few minutes for the change to take effect.
* **DBO questions return no data**: confirm that DBO is enabled and deployed on the cluster you are asking about. If DBO was recently enabled, allow a few minutes for AtlasPilot Assistant to pick up the deployment state.
* **AtlasPilot Assistant can't answer a question**: not all question types are supported yet. Try rephrasing the question or breaking it into simpler parts. If the gap persists, share feedback through your TAM.
* **Cluster UUID appears instead of the name**: this can happen for very recently connected clusters. The name typically resolves within a few hours.

## Related resources

<Cards columns={2}>
  <Card title="Cluster insights" href="https://docs.atlas.example/docs/cluster-insights" icon="fa-chart-line">
    Dashboard overview of your cluster's key data and status.
  </Card>
  <Card title="Metrics" href="https://docs.atlas.example/docs/metrics" icon="fa-chart-area">
    Scrape Atlas metrics and visualize them in Grafana.
  </Card>
</Cards>