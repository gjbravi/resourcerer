// Package prom reads usage history from a Prometheus-compatible API.
//
// Queries join cAdvisor usage to the owning workload through kube-state-metrics
// (kube_pod_owner and kube_replicaset_owner), so pods from earlier ReplicaSets count too.
package prom

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Client struct {
	API promv1.API
	// RateWindow is the rate() range for CPU usage; at least 4× the scrape interval.
	RateWindow string
}

func New(url, rateWindow string) (*Client, error) {
	c, err := api.NewClient(api.Config{Address: url})
	if err != nil {
		return nil, err
	}
	return &Client{API: promv1.NewAPI(c), RateWindow: rateWindow}, nil
}

type Ref struct {
	Namespace, Kind, Name string
}

// Point is one sample: t in unix seconds.
type Point struct {
	T float64 `json:"t"`
	V float64 `json:"v"`
}

func quote(s string) string { return strings.ReplaceAll(s, `"`, `\"`) }

// owner selects the pods of a workload as a vector of 1s keyed by namespace and pod.
func owner(r Ref) string {
	ns := quote(r.Namespace)
	name := quote(r.Name)
	// Deployments and Argo Rollouts both own their pods through ReplicaSets.
	if r.Kind == "Deployment" || r.Kind == "Rollout" {
		return fmt.Sprintf(`max by (namespace, pod) (
  label_replace(kube_pod_owner{namespace="%s", owner_kind="ReplicaSet"}, "replicaset", "$1", "owner_name", "(.*)")
  * on (namespace, replicaset) group_left ()
  max by (namespace, replicaset) (kube_replicaset_owner{namespace="%s", owner_kind="%s", owner_name="%s"})
)`, ns, ns, r.Kind, name)
	}
	return fmt.Sprintf(`max by (namespace, pod) (kube_pod_owner{namespace="%s", owner_kind="%s", owner_name="%s"})`, ns, quote(r.Kind), name)
}

// perPod returns usage per pod and container, restricted to the workload's pods.
func (c *Client) perPod(r Ref, resource string) string {
	ns := quote(r.Namespace)
	var usage string
	if resource == "cpu" {
		usage = fmt.Sprintf(`sum by (namespace, pod, container) (rate(container_cpu_usage_seconds_total{namespace="%s", container!="", container!="POD"}[%s]))`, ns, c.RateWindow)
	} else {
		usage = fmt.Sprintf(`max by (namespace, pod, container) (container_memory_working_set_bytes{namespace="%s", container!="", container!="POD"})`, ns)
	}
	return fmt.Sprintf("(%s)\n* on (namespace, pod) group_left ()\n%s", usage, owner(r))
}

// BusiestPod is usage of the busiest pod per container at each step: the one that throttles or gets OOM-killed.
func (c *Client) BusiestPod(r Ref, resource string) string {
	return fmt.Sprintf("max by (container) (%s)", c.perPod(r, resource))
}

// Total is usage summed over all pods and containers of the workload.
func (c *Client) Total(r Ref, resource string) string {
	return fmt.Sprintf("sum(%s)", c.perPod(r, resource))
}

// Pods counts running pods of the workload.
func (c *Client) Pods(r Ref) string {
	return fmt.Sprintf(`count(max by (namespace, pod) (container_memory_working_set_bytes{namespace="%s", container!="", container!="POD"}) * on (namespace, pod) group_left () %s)`,
		quote(r.Namespace), owner(r))
}

// Requests is the configured request ("requests") or limit ("limits") per container over time, from kube-state-metrics.
func (c *Client) Requests(r Ref, kind, resource string) string {
	unit := map[string]string{"cpu": "cpu", "memory": "memory"}[resource]
	return fmt.Sprintf(`max by (container) (kube_pod_container_resource_%s{namespace="%s", resource="%s"} * on (namespace, pod) group_left () %s)`,
		kind, quote(r.Namespace), unit, owner(r))
}

// Throttled is the share of CFS periods each container of the workload spent throttled over window.
func (c *Client) Throttled(r Ref, window time.Duration) string {
	ns := quote(r.Namespace)
	part := func(m string) string {
		return fmt.Sprintf("sum by (container) (increase(%s{namespace=\"%s\", container!=\"\", container!=\"POD\"}[%s]) * on (namespace, pod) group_left () %s)",
			m, ns, model.Duration(window), owner(r))
	}
	return part("container_cpu_cfs_throttled_periods_total") + "\n/\n" + part("container_cpu_cfs_periods_total")
}

// Instant runs an instant query and returns one value per value of label.
func (c *Client) Instant(ctx context.Context, q string, label string) (map[string]float64, error) {
	v, _, err := c.API.Query(ctx, q, time.Now())
	if err != nil {
		return nil, err
	}
	vec, ok := v.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected result type %s", v.Type())
	}
	out := map[string]float64{}
	for _, s := range vec {
		if f := float64(s.Value); !math.IsNaN(f) && !math.IsInf(f, 0) {
			out[string(s.Metric[model.LabelName(label)])] = f
		}
	}
	return out, nil
}

// ClusterAllocatable is what the nodes offer to pods, from kube-state-metrics.
func ClusterAllocatable(resource string) string {
	return fmt.Sprintf(`sum(kube_node_status_allocatable{resource="%s"})`, resource)
}

// ClusterRequested sums the requests of running pods; finished pods keep their series in kube-state-metrics.
func ClusterRequested(resource string) string {
	return fmt.Sprintf(`sum(kube_pod_container_resource_requests{resource="%s"} * on (namespace, pod) group_left () (max by (namespace, pod) (kube_pod_status_phase{phase="Running"}) == 1))`, resource)
}

// ClusterUsage sums usage over every container in the cluster.
func (c *Client) ClusterUsage(resource string) string {
	if resource == "cpu" {
		return fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{container!="", container!="POD"}[%s]))`, c.RateWindow)
	}
	return `sum(container_memory_working_set_bytes{container!="", container!="POD"})`
}

// Recommended is the history of recommendations exported by the controller.
func Recommended(r Ref, typ, resource string) string {
	return fmt.Sprintf(`max by (container) (resourcerer_recommendation{namespace="%s", workload_kind="%s", workload="%s", resource="%s", type="%s"})`,
		quote(r.Namespace), quote(r.Kind), quote(r.Name), resource, typ)
}

// Step picks a resolution that stays under Prometheus' 11000 points per series.
func Step(window time.Duration, minStep time.Duration) time.Duration {
	s := window / 8000
	if s < minStep {
		s = minStep
	}
	return s.Truncate(time.Second)
}

// Range runs a range query and returns one series per value of label (empty label: a single series under "").
func (c *Client) Range(ctx context.Context, q string, start, end time.Time, step time.Duration, label string) (map[string][]Point, error) {
	v, _, err := c.API.QueryRange(ctx, q, promv1.Range{Start: start, End: end, Step: step})
	if err != nil {
		return nil, err
	}
	m, ok := v.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected result type %s", v.Type())
	}
	out := map[string][]Point{}
	for _, s := range m {
		key := string(s.Metric[model.LabelName(label)])
		pts := make([]Point, 0, len(s.Values))
		for _, p := range s.Values {
			f := float64(p.Value)
			if math.IsNaN(f) || math.IsInf(f, 0) {
				continue
			}
			pts = append(pts, Point{T: float64(p.Timestamp.Unix()), V: f})
		}
		out[key] = pts
	}
	return out, nil
}

func Values(pts []Point) []float64 {
	out := make([]float64, len(pts))
	for i, p := range pts {
		out[i] = p.V
	}
	return out
}

func Mean(pts []Point) float64 {
	if len(pts) == 0 {
		return 0
	}
	s := 0.0
	for _, p := range pts {
		s += p.V
	}
	return s / float64(len(pts))
}

// Status describes the metrics source for the Settings page.
type Status struct {
	URL       string   `json:"url"`
	Connected bool     `json:"connected"`
	Error     string   `json:"error,omitempty"`
	Version   string   `json:"version,omitempty"`
	Retention string   `json:"retention,omitempty"`
	Oldest    string   `json:"oldest,omitempty"`
	Checks    []Check  `json:"checks"`
	Missing   []string `json:"missing,omitempty"`
}

type Check struct {
	Name   string `json:"name"`
	Metric string `json:"metric"`
	OK     bool   `json:"ok"`
}

func (c *Client) Status(ctx context.Context, url string) Status {
	st := Status{URL: url}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	b, err := c.API.Buildinfo(ctx)
	if err != nil {
		st.Error = err.Error()
		return st
	}
	st.Connected = true
	st.Version = b.Version
	if rt, err := c.API.Runtimeinfo(ctx); err == nil {
		st.Retention = rt.StorageRetention
	}
	checks := []Check{
		{Name: "cAdvisor CPU", Metric: "container_cpu_usage_seconds_total"},
		{Name: "cAdvisor memory", Metric: "container_memory_working_set_bytes"},
		{Name: "kube-state-metrics owners", Metric: "kube_pod_owner"},
		{Name: "kube-state-metrics requests", Metric: "kube_pod_container_resource_requests"},
		{Name: "resourcerer recommendations", Metric: "resourcerer_recommendation"},
	}
	for i, ch := range checks {
		v, _, err := c.API.Query(ctx, fmt.Sprintf("count(%s)", ch.Metric), time.Now())
		if vec, ok := v.(model.Vector); err == nil && ok && len(vec) > 0 && vec[0].Value > 0 {
			checks[i].OK = true
		} else {
			st.Missing = append(st.Missing, ch.Metric)
		}
	}
	st.Checks = checks
	// Oldest usable sample, approximated by the furthest point back that cAdvisor data exists at a coarse step.
	end := time.Now()
	start := end.Add(-90 * 24 * time.Hour)
	if v, _, err := c.API.QueryRange(ctx, `count(container_memory_working_set_bytes)`, promv1.Range{Start: start, End: end, Step: 6 * time.Hour}); err == nil {
		if m, ok := v.(model.Matrix); ok && len(m) > 0 && len(m[0].Values) > 0 {
			st.Oldest = m[0].Values[0].Timestamp.Time().UTC().Format(time.RFC3339)
		}
	}
	sort.Strings(st.Missing)
	return st
}
