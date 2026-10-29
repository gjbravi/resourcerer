package server

import (
	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/prom"
)

// Resources are cores for CPU and bytes for memory; absent keys are unset.
type Resources map[string]float64

type Info struct {
	Version            string        `json:"version"`
	Cluster            string        `json:"cluster"`
	Auth               string        `json:"auth"`
	ReadOnly           bool          `json:"readOnly"`
	Selection          string        `json:"selection"`
	ExcludedNamespaces []string      `json:"excludedNamespaces"`
	Defaults           []SettingInfo `json:"defaults"`
	Cost               Cost          `json:"cost"`
	Prometheus         prom.Status   `json:"prometheus"`
}

type Cost struct {
	CPUCoreHour   float64 `json:"cpuCoreHour"`
	MemoryGiBHour float64 `json:"memoryGiBHour"`
	// Source is "karpenter" when the rates are calibrated to node prices, "static" for the configured rates.
	Source             string  `json:"source"`
	Reason             string  `json:"reason,omitempty"`
	Nodes              int     `json:"nodes,omitempty"`
	Spot               int     `json:"spot,omitempty"`
	OnDemand           int     `json:"onDemand,omitempty"`
	HourlyCost         float64 `json:"hourlyCost,omitempty"`
	OnDemandHourlyCost float64 `json:"onDemandHourlyCost,omitempty"`
	UpdatedAt          string  `json:"updatedAt,omitempty"`
	Refresh            string  `json:"refresh,omitempty"`
}

type SettingInfo struct {
	Key        string   `json:"key"`
	Value      string   `json:"value"`
	Source     string   `json:"source"`
	Annotation string   `json:"annotation"`
	Options    []string `json:"options,omitempty"`
}

type Totals struct {
	CPU         float64 `json:"cpu"`
	Memory      float64 `json:"memory"`
	MonthlyCost float64 `json:"monthlyCost"`
}

type Score struct {
	Overall int32 `json:"overall"`
	CPU     int32 `json:"cpu"`
	Memory  int32 `json:"memory"`
}

type Summary struct {
	Namespace    string                `json:"namespace"`
	Kind         string                `json:"kind"`
	Name         string                `json:"name"`
	Replicas     int32                 `json:"replicas"`
	AvgReplicas  float64               `json:"avgReplicas"`
	Phase        string                `json:"phase"`
	Message      string                `json:"message,omitempty"`
	Score        *Score                `json:"score,omitempty"`
	Current      Totals                `json:"current"`
	Recommended  Totals                `json:"recommended"`
	Pending      bool                  `json:"pending"`
	AutoDeploy   bool                  `json:"autoDeploy"`
	OOMKills     int32                 `json:"oomKills"`
	Reliability  *v1alpha1.Reliability `json:"reliability,omitempty"`
	NoRequests   bool                  `json:"noRequests"`
	Under        bool                  `json:"under"`
	HPA          []string              `json:"hpa,omitempty"`
	GeneratedAt  string                `json:"generatedAt,omitempty"`
	LastApplied  string                `json:"lastAppliedAt,omitempty"`
	Hash         string                `json:"hash,omitempty"`
	NamespaceSet bool                  `json:"namespaceAnnotated"`
}

type Container struct {
	Name               string            `json:"name"`
	Skipped            bool              `json:"skipped"`
	CurrentRequests    Resources         `json:"currentRequests"`
	CurrentLimits      Resources         `json:"currentLimits"`
	RecommendedRequest Resources         `json:"recommendedRequests"`
	RecommendedLimits  Resources         `json:"recommendedLimits"`
	Basis              map[string]string `json:"basis,omitempty"`
}

type HPA struct {
	Name        string `json:"name"`
	Current     int32  `json:"current"`
	Recommended int32  `json:"recommended"`
	Clamped     bool   `json:"clamped"`
}

type Change struct {
	Name         string    `json:"name"`
	FromRequests Resources `json:"fromRequests"`
	ToRequests   Resources `json:"toRequests"`
	FromLimits   Resources `json:"fromLimits"`
	ToLimits     Resources `json:"toLimits"`
}

type History struct {
	Time       string   `json:"time"`
	Action     string   `json:"action"`
	By         string   `json:"by"`
	Hash       string   `json:"hash,omitempty"`
	Containers []Change `json:"containers"`
}

type Event struct {
	Type    string `json:"type"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

type ApplyResult struct {
	Time    string `json:"time"`
	Result  string `json:"result"`
	Message string `json:"message,omitempty"`
}

type Detail struct {
	Summary
	// Capacity counts the workload's pods per capacity type of their Karpenter node (spot, on-demand).
	Capacity        map[string]int `json:"capacity,omitempty"`
	NextScheduleAt  string         `json:"nextScheduleAt,omitempty"`
	Containers      []Container    `json:"containers"`
	HPARec          *HPA           `json:"hpaRecommendation,omitempty"`
	Settings        []SettingInfo  `json:"settings"`
	Warnings        []string       `json:"warnings,omitempty"`
	History         []History      `json:"history"`
	Events          []Event        `json:"events"`
	LastApplyResult *ApplyResult   `json:"lastApplyResult,omitempty"`
	ApplyMethod     string         `json:"applyMethod"`
	// TemplateKind and TemplateName name where resources are applied, when not the workload itself.
	TemplateKind string `json:"templateKind,omitempty"`
	TemplateName string `json:"templateName,omitempty"`
}

type Band struct {
	T   float64 `json:"t"`
	P50 float64 `json:"p50"`
	P75 float64 `json:"p75"`
	P95 float64 `json:"p95"`
}

type Series struct {
	From  float64                 `json:"from"`
	To    float64                 `json:"to"`
	Bands []Band                  `json:"bands"`
	Lines map[string][]prom.Point `json:"lines"`
}

// Check is one reliability check over a window; Value is a share for "ratio" checks and a number for "count".
type Check struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Container string       `json:"container,omitempty"`
	Pod       string       `json:"pod,omitempty"`
	Status    string       `json:"status"`
	Unit      string       `json:"unit"`
	Value     float64      `json:"value"`
	Display   string       `json:"display"`
	Threshold float64      `json:"threshold,omitempty"`
	Detail    string       `json:"detail"`
	Source    string       `json:"source"`
	Query     string       `json:"query,omitempty"`
	Series    []prom.Point `json:"series,omitempty"`
}

type ReliabilityReport struct {
	Window      string  `json:"window"`
	From        float64 `json:"from"`
	To          float64 `json:"to"`
	Step        float64 `json:"step"`
	GeneratedAt string  `json:"generatedAt"`
	Checks      []Check `json:"checks"`
}
