"""Writes the resourcerer Grafana dashboard: how resourcerer is doing across the cluster.

make dashboard runs it into docs/assets/grafana/resourcerer-dashboard.json, an example the docs publish for people to
import and adapt (docs/guides/grafana-dashboard.md); make lint fails when the two differ. Edit this file, not the JSON.
"""
import json
import sys

DS = {"type": "prometheus", "uid": "${datasource}"}
W = '{namespace=~"$namespace", workload_kind=~"$kind"}'


def of(expr, component=None, by=""):
    """Keeps the series of resourcerer's own processes, and adds their component: matched by instance to
    resourcerer_build_info, so it holds whatever the release and the scrape job are called."""
    sel = f'{{component=~"{component}"}}' if component else ""
    return f"({expr}) * on (instance) group_left ({by}) max by (instance, component) (resourcerer_build_info{sel})"


CONTROLLER = "controller|all"
panels = []
y = 0
pid = 0


def nid():
    global pid
    pid += 1
    return pid


def row(title):
    global y
    panels.append({"type": "row", "title": title, "id": nid(), "collapsed": False,
                   "gridPos": {"h": 1, "w": 24, "x": 0, "y": y}, "panels": []})
    y += 1


def target(expr, legend="", instant=False, fmt=None):
    t = {"datasource": DS, "expr": expr, "legendFormat": legend or "__auto", "refId": chr(65 + 0)}
    if instant:
        t["instant"] = True
        t["range"] = False
    if fmt:
        t["format"] = fmt
    return t


def add(p, x, w, h):
    p["gridPos"] = {"h": h, "w": w, "x": x, "y": y}
    p["id"] = nid()
    p["datasource"] = DS
    for i, t in enumerate(p.get("targets", [])):
        t["refId"] = chr(65 + i)
    panels.append(p)


def stat(title, expr, unit="short", x=0, w=4, h=4, desc="", color="text", decimals=None, thresholds=None):
    p = {"type": "stat", "title": title, "description": desc, "targets": [target(expr, instant=True)],
         "fieldConfig": {"defaults": {"unit": unit, "color": {"mode": "thresholds" if thresholds else "fixed",
                                                                 "fixedColor": color},
                                      "thresholds": {"mode": "absolute", "steps": thresholds or [
                                          {"color": color, "value": None}]}},
                         "overrides": []},
         "options": {"reduceOptions": {"calcs": ["lastNotNull"]}, "colorMode": "value", "graphMode": "area",
                     "textMode": "value"}}
    if decimals is not None:
        p["fieldConfig"]["defaults"]["decimals"] = decimals
    add(p, x, w, h)


def ts(title, targets, unit="short", x=0, w=12, h=8, desc="", stack=False, fill=10, overrides=None, minv=None):
    p = {"type": "timeseries", "title": title, "description": desc,
         "targets": [target(*t) if isinstance(t, tuple) else target(t) for t in targets],
         "fieldConfig": {"defaults": {"unit": unit, "custom": {
             "fillOpacity": fill, "lineWidth": 2, "showPoints": "never",
             "stacking": {"mode": "normal" if stack else "none", "group": "A"}}},
             "overrides": overrides or []},
         "options": {"legend": {"displayMode": "list", "placement": "bottom"},
                     "tooltip": {"mode": "multi", "sort": "desc"}}}
    if minv is not None:
        p["fieldConfig"]["defaults"]["min"] = minv
    add(p, x, w, h)


def color(name, c):
    return {"matcher": {"id": "byName", "options": name},
            "properties": [{"id": "color", "value": {"mode": "fixed", "fixedColor": c}}]}


def table(title, expr, unit, x, w, h, desc="", value="Value"):
    p = {"type": "table", "title": title, "description": desc,
         "targets": [target(expr, instant=True, fmt="table")],
         "fieldConfig": {"defaults": {"unit": unit}, "overrides": [
             {"matcher": {"id": "byName", "options": "Value"},
              "properties": [{"id": "displayName", "value": value},
                             {"id": "custom.cellOptions", "value": {"type": "gauge", "mode": "lcd"}}]}]},
         "options": {"showHeader": True, "sortBy": [{"displayName": value, "desc": True}]},
         "transformations": [{"id": "organize", "options": {"excludeByName": {"Time": True}}}]}
    add(p, x, w, h)


cost = "resourcerer_workload_monthly_cost" + W[:-1]
cur = 'resourcerer_workload_monthly_cost{state="current", namespace=~"$namespace", workload_kind=~"$kind"}'
rec = 'resourcerer_workload_monthly_cost{state="recommended", namespace=~"$namespace", workload_kind=~"$kind"}'
reclaim = f"clamp_min({cur} - ignoring(state) {rec}, 0)"
wl = 'resourcerer_workloads{workload_kind=~"$kind"}'

# ── Summary ────────────────────────────────────────────────────────────────
row("Summary")
stat("Estimated savings / month", f"sum({cur}) - sum({rec})", "currencyUSD", 0, 4, color="green",
     desc="Current minus recommended monthly cost, over every workload: what applying everything would change. "
          "Negative when the recommendations need more capacity than they free.")
stat("Reclaimable / month", f"sum({reclaim})", "currencyUSD", 4, 4, color="green",
     desc="What applying every recommendation that lowers requests would save; workloads that need more are not "
          "subtracted. The Overview's Reclaimable.")
stat("Monthly cost now", f"sum({cur})", "currencyUSD", 8, 4, color="text",
     desc="Requests × average replicas × rates × 730, summed over the managed workloads. Not the node bill: idle "
          "capacity and excluded namespaces are not in it.")
stat("Managed workloads", f"sum({wl})", "short", 12, 3, color="purple")
stat("Ready", f'sum({wl[:-1]}, phase="Ready"}}) / sum({wl})', "percentunit", 15, 3, color="green",
     desc="Share of managed workloads whose recommendation can be applied with full history.")
stat("Pending an apply", f'sum({wl[:-1]}, pending="true"}})', "short", 18, 3, color="orange",
     desc="Recommendations that differ from the requests by more than auto-deploy.threshold.")
stat("Auto-deploy on", f'sum({wl[:-1]}, auto_deploy="true"}})', "short", 21, 3, color="blue")
y += 4

# ── Savings ────────────────────────────────────────────────────────────────
row("Savings")
ts("Monthly cost: now and with the recommendations",
   [(f"sum({cur})", "now"), (f"sum({rec})", "with recommendations")], "currencyUSD", 0, 12,
   overrides=[color("now", "blue"), color("with recommendations", "green")],
   desc="The distance between the lines is what applying everything would save.")
ts("Reclaimable and net savings / month",
   [(f"sum({reclaim})", "reclaimable"), (f"sum({cur}) - sum({rec})", "net")], "currencyUSD", 12, 12,
   overrides=[color("reclaimable", "green"), color("net", "yellow")])
y += 8
table("Reclaimable / month by namespace", f"sort_desc(sum by (namespace) ({reclaim}))", "currencyUSD", 0, 8, 9,
      value="Reclaimable")
table("Top workloads by reclaimable / month", f"topk(15, sum by (namespace, workload_kind, workload) ({reclaim}))",
      "currencyUSD", 8, 10, 9, value="Reclaimable")
table("Workloads that need more capacity", f"topk(10, sum by (namespace, workload_kind, workload) "
      f"(clamp_min({rec} - ignoring(state) {cur}, 0)) > 0)", "currencyUSD", 18, 6, 9, value="Extra / month",
      desc="Recommendations that raise cost, usually to fix throttling or OOM kills.")
y += 9

# ── Capacity ───────────────────────────────────────────────────────────────
row("Capacity")
req = 'resourcerer_workload_requests{namespace=~"$namespace", workload_kind=~"$kind", '
ts("CPU requested: now and with the recommendations",
   [(f'sum({req}resource="cpu", state="current"}})', "now"),
    (f'sum({req}resource="cpu", state="recommended"}})', "with recommendations"),
    ('sum(kube_node_status_allocatable{resource="cpu"})', "cluster allocatable")], "short", 0, 8,
   overrides=[color("now", "blue"), color("with recommendations", "green"), color("cluster allocatable", "#888888")],
   desc="Cores: requests × average replicas of the managed workloads, against what the nodes offer.")
ts("Memory requested: now and with the recommendations",
   [(f'sum({req}resource="memory", state="current"}})', "now"),
    (f'sum({req}resource="memory", state="recommended"}})', "with recommendations")], "bytes", 8, 8,
   overrides=[color("now", "blue"), color("with recommendations", "green")])
ts("Cost rates in use", [(f'max by (resource) ({of("resourcerer_cost_rate", CONTROLLER)})', "{{resource}}"),
                         (f'max({of("resourcerer_cost_cluster_factor", CONTROLLER)})', "cluster factor")], "short", 16, 8,
   overrides=[{"matcher": {"id": "byName", "options": "cluster factor"},
               "properties": [{"id": "custom.axisPlacement", "value": "right"}, {"id": "unit", "value": "none"}]}],
   desc="Per core-hour (cpu) and GiB-hour (memory). Cost moving while requests stay flat is a price change, not "
        "rightsizing.")
y += 8

# ── Workloads ──────────────────────────────────────────────────────────────
row("Workloads")
ts("By kind", [(f"sum by (workload_kind) ({wl})", "{{workload_kind}}")], "short", 0, 8, stack=True, fill=40)
ts("By phase", [(f"sum by (phase) ({wl})", "{{phase}}")], "short", 8, 8, stack=True, fill=40,
   overrides=[color("Ready", "green"), color("Learning", "blue"), color("Pending", "yellow"),
              color("Warning", "orange"), color("Error", "red"), color("Unsupported", "#888888")])
ts("Backlog and adoption", [(f'sum({wl[:-1]}, pending="true"}})', "pending an apply"),
                            (f'sum({wl[:-1]}, auto_deploy="true"}})', "auto-deploy on"),
                            (f'count(resourcerer_workload_drifted{W}) or vector(0)', "drifted")], "short", 16, 8,
   overrides=[color("pending an apply", "orange"), color("auto-deploy on", "blue"), color("drifted", "red")])
y += 8

# ── Applies ────────────────────────────────────────────────────────────────
row("Applies")
ts("Applies and reverts / hour",
   [('sum by (action, trigger, result) (increase(resourcerer_apply_total[1h]))', "{{action}} {{trigger}} {{result}}")],
   "short", 0, 8, stack=True, fill=40)
ts("Held by an apply policy / hour",
   [('sum by (reason) (increase(resourcerer_apply_deferred_total[1h]))', "{{reason}}")], "short", 8, 8, stack=True,
   fill=40)
ts("Applies undone by another manager",
   [('round(sum by (manager) (increase(resourcerer_apply_overwritten_total[1h])))', "{{manager}}")], "short", 16, 8,
   stack=True, fill=40, desc="An applied recommendation whose values a GitOps sync, Helm or a person changed back. "
                             "See the RecommendationOverwritten events.")
y += 8

# ── Prometheus load ────────────────────────────────────────────────────────
row("Load on Prometheus")
ts("Queries / s by purpose", [('sum by (purpose) (rate(resourcerer_prometheus_requests_total[5m]))', "{{purpose}}")],
   "reqps", 0, 8, stack=True, fill=40)
ts("Samples / s by purpose", [('sum by (purpose) (rate(resourcerer_prometheus_samples_total[5m]))', "{{purpose}}")],
   "short", 8, 8, stack=True, fill=40, desc="What the queries cost Prometheus to evaluate and send.")
ts("Query duration p95 by purpose",
   [('histogram_quantile(0.95, sum by (le, purpose) (rate(resourcerer_prometheus_request_duration_seconds_bucket[5m])))',
     "{{purpose}}")], "s", 16, 8)
y += 8
ts("Failed queries / s by code", [('sum by (code, purpose) (rate(resourcerer_prometheus_requests_total{code!~"2.."}[5m]))',
                                   "{{code}} {{purpose}}")], "reqps", 0, 8,
   desc="error: no response; timeout: past prometheus.timeout.")
ts("Response bytes / s by component",
   [(f'sum by (component) ({of("rate(resourcerer_prometheus_response_bytes_total[5m])", by="component")})',
     "{{component}}")], "Bps", 8, 8)
ts("Queries in flight by component",
   [(f'sum by (component) ({of("resourcerer_prometheus_requests_in_flight", by="component")})', "{{component}}")],
   "short", 16, 8)
y += 8

# ── Server ─────────────────────────────────────────────────────────────────
row("Server")
ts("Requests / s by route", [('sum by (route) (rate(resourcerer_http_requests_total{route!="ui"}[5m]))', "{{route}}")],
   "reqps", 0, 8, stack=True, fill=40)
ts("Latency p95 by route",
   [('histogram_quantile(0.95, sum by (le, route) (rate(resourcerer_http_request_duration_seconds_bucket{route!="ui"}[5m])))',
     "{{route}}")], "s", 8, 8)
ts("Errors / s by route and code",
   [('sum by (route, code) (rate(resourcerer_http_requests_total{code=~"5..|4.."}[5m]))', "{{route}} {{code}}")],
   "reqps", 16, 8)
y += 8
ts("Cache: share answered without querying",
   [('sum by (query) (rate(resourcerer_server_cache_requests_total{result!="miss"}[15m])) / '
     'sum by (query) (rate(resourcerer_server_cache_requests_total[15m]))', "{{query}}")], "percentunit", 0, 12,
   minv=0)
ts("Decision models", [('round(sum by (model, result) (increase(resourcerer_decision_requests_total[1h])))', "{{model}} {{result}} / h"),
                       ('histogram_quantile(0.95, sum by (le, model) (rate(resourcerer_decision_request_duration_seconds_bucket[1h])))',
                        "{{model}} p95 seconds")], "short", 12, 12)
y += 8

# ── Health ─────────────────────────────────────────────────────────────────
row("Health")
stat("Components up", f'count({of("up == 1")})', "short", 0, 3, color="green")
stat("Oldest recommendation", f"max(time() - resourcerer_recommendation_generated_timestamp_seconds{W})", "s", 3, 3,
     thresholds=[{"color": "green", "value": None}, {"color": "orange", "value": 129600},
                 {"color": "red", "value": 259200}],
     desc="Time since the stalest recommendation was computed. With the default @daily schedule, past a day and a "
          "half something is not refreshing.")
stat("Least history", f"min(resourcerer_recommendation_history_coverage_ratio{W})", "percentunit", 6, 3,
     thresholds=[{"color": "red", "value": None}, {"color": "orange", "value": 0.5}, {"color": "green", "value": 0.95}],
     desc="The recommendation with the least usage history over its learning period.")
stat("Cost calibration", f'max by (source, state) ({of("resourcerer_cost_calibration", CONTROLLER)})',
     "short", 9, 4, color="blue", desc="Where cost rates come from, and the state of the last calibration.")
panels[-1]["options"]["textMode"] = "name"
panels[-1]["targets"][0]["legendFormat"] = "{{source}} · {{state}}"
stat("Signal errors / h", "sum(increase(resourcerer_signal_errors_total[1h])) or vector(0)", "short", 13, 3,
     thresholds=[{"color": "green", "value": None}, {"color": "orange", "value": 1}])
stat("Settings warnings", f"count(resourcerer_settings_warnings{W}) or vector(0)", "short", 16, 3,
     thresholds=[{"color": "green", "value": None}, {"color": "orange", "value": 1}])
stat("Reconcile errors / h", f'sum({of("increase(controller_runtime_reconcile_errors_total[1h])")}) or vector(0)',
     "short", 19, 5, thresholds=[{"color": "green", "value": None}, {"color": "red", "value": 1}])
y += 4
ts("Recommendations computed / h by phase",
   [('sum by (phase) (increase(resourcerer_recommendation_generations_total[1h]))', "{{phase}}")], "short", 0, 8,
   stack=True, fill=40, overrides=[color("Ready", "green"), color("Learning", "blue"), color("Pending", "yellow"),
                                   color("Error", "red")],
   desc="Only the reconciles that query Prometheus. Pending and Error back off up to the schedule.")
ts("Time to compute one, p95 by kind",
   [('histogram_quantile(0.95, sum by (le, workload_kind) (rate(resourcerer_recommendation_generation_duration_seconds_bucket[15m])))',
     "{{workload_kind}}")], "s", 8, 8)
ts("Work queue depth", [(f'sum by (name) ({of("workqueue_depth")})', "{{name}}")], "short", 16, 8)
y += 8
table("Versions", 'max by (component, version) (resourcerer_build_info)', "short", 0, 24, 5, value="Up")

dashboard = {
    "title": "resourcerer",
    "uid": "resourcerer",
    "tags": ["resourcerer", "kubernetes", "rightsizing"],
    "timezone": "browser",
    "schemaVersion": 39,
    "refresh": "30s",
    "time": {"from": "now-3h", "to": "now"},
    "editable": True,
    "graphTooltip": 1,
    "description": "How resourcerer is doing across the cluster: savings, workloads, applies, load on Prometheus and health.",
    "templating": {"list": [
        {"name": "datasource", "type": "datasource", "query": "prometheus", "label": "Data source"},
        {"name": "namespace", "type": "query", "datasource": DS, "label": "Namespace",
         "query": {"query": "label_values(resourcerer_workload_monthly_cost, namespace)", "refId": "ns"},
         "definition": "label_values(resourcerer_workload_monthly_cost, namespace)",
         "includeAll": True, "multi": True, "allValue": ".*", "current": {"text": "All", "value": "$__all"},
         "refresh": 2, "sort": 1},
        {"name": "kind", "type": "query", "datasource": DS, "label": "Kind",
         "query": {"query": "label_values(resourcerer_workloads, workload_kind)", "refId": "kind"},
         "definition": "label_values(resourcerer_workloads, workload_kind)",
         "includeAll": True, "multi": True, "allValue": ".*", "current": {"text": "All", "value": "$__all"},
         "refresh": 2, "sort": 1},
    ]},
    "panels": panels,
}
json.dump(dashboard, sys.stdout, indent=1)
sys.stdout.write("\n")
