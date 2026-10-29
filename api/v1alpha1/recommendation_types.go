package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Phase summarizes whether a recommendation can be trusted and applied.
// +kubebuilder:validation:Enum=Pending;Learning;Ready;Warning;Unsupported;Error
type Phase string

const (
	// PhasePending: not enough history for any recommendation yet.
	PhasePending Phase = "Pending"
	// PhaseLearning: preliminary recommendation; the learning period is not covered yet.
	PhaseLearning Phase = "Learning"
	PhaseReady    Phase = "Ready"
	// PhaseWarning: a recommendation exists but cannot be applied now, e.g. scaled to zero.
	PhaseWarning Phase = "Warning"
	// PhaseUnsupported: the workload cannot be optimized, e.g. its HPA scales on CPU and memory.
	PhaseUnsupported Phase = "Unsupported"
	// PhaseError: metrics could not be read.
	PhaseError Phase = "Error"
)

// +kubebuilder:validation:Enum=Apply;Revert
type Action string

const (
	ActionApply  Action = "Apply"
	ActionRevert Action = "Revert"
)

type TargetRef struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

// ApplyRequest asks the applier to act. It is handled once per ID.
type ApplyRequest struct {
	// ID is unique per request; timestamps alone cannot tell apart two requests in the same second.
	ID     string `json:"id"`
	Action Action `json:"action"`
	// Hash of the recommendation the user reviewed; Apply is refused if it changed since.
	// +optional
	Hash        string      `json:"hash,omitempty"`
	RequestedBy string      `json:"requestedBy"`
	RequestedAt metav1.Time `json:"requestedAt"`
}

type RecommendationSpec struct {
	TargetRef TargetRef `json:"targetRef"`
	// +optional
	ApplyRequest *ApplyRequest `json:"applyRequest,omitempty"`
	// Regenerate now instead of waiting for the schedule, when newer than status.generatedAt.
	// +optional
	RegenerateRequestedAt *metav1.Time `json:"regenerateRequestedAt,omitempty"`
}

// Setting is the effective value of one setting and the level that set it.
type Setting struct {
	Value string `json:"value"`
	// +kubebuilder:validation:Enum=cluster;namespace;workload
	Source string `json:"source"`
}

type ContainerRecommendation struct {
	Name string `json:"name"`
	// Skipped containers keep their resources (containers.skip or DoNotOptimize).
	// +optional
	Skipped bool `json:"skipped,omitempty"`
	// +optional
	Current corev1.ResourceRequirements `json:"current,omitempty"`
	// +optional
	Recommended corev1.ResourceRequirements `json:"recommended,omitempty"`
	// Basis explains each value, e.g. "p95 + 10% over 7d".
	// +optional
	Basis map[string]string `json:"basis,omitempty"`
}

type HPARecommendation struct {
	Name    string `json:"name"`
	Current int32  `json:"current"`
	// Target utilization that keeps the HPA scaling at the same absolute CPU per pod.
	Recommended int32 `json:"recommended"`
	// +optional
	Clamped bool `json:"clamped,omitempty"`
}

type Score struct {
	Overall int32 `json:"overall"`
	CPU     int32 `json:"cpu"`
	Memory  int32 `json:"memory"`
}

// Totals are requests summed over containers and multiplied by average replicas.
type Totals struct {
	CPU         string `json:"cpu"`
	Memory      string `json:"memory"`
	MonthlyCost string `json:"monthlyCost"`
}

type Impact struct {
	AvgReplicas string `json:"avgReplicas"`
	Current     Totals `json:"current"`
	Recommended Totals `json:"recommended"`
}

// Pricing records how the costs in Impact were priced.
type Pricing struct {
	// Allocation is cluster (the cluster's average rates) or node (the rates of the nodes the pods ran on).
	// +kubebuilder:validation:Enum=cluster;node
	Allocation string `json:"allocation"`
	// Factor multiplies the configured rates for this workload; ClusterFactor is the cluster's average.
	Factor        string `json:"factor"`
	ClusterFactor string `json:"clusterFactor"`
	// Share of the workload's requests that ran on nodes with a price, in percent.
	// +optional
	CoveragePercent int32 `json:"coveragePercent,omitempty"`
	// Share of the priced requests per capacity type (spot, on-demand, reserved), in percent.
	// +optional
	Capacity map[string]int32 `json:"capacity,omitempty"`
	// Instance types that carried most of the priced requests, with their share in percent.
	// +optional
	InstanceTypes []InstanceTypeShare `json:"instanceTypes,omitempty"`
	// Reason explains why node allocation fell back to the cluster's rates.
	// +optional
	Reason string `json:"reason,omitempty"`
}

type InstanceTypeShare struct {
	Name    string `json:"name"`
	Percent int32  `json:"percent"`
}

// Reliability holds signals from the learning period that requests and limits may be too tight.
type Reliability struct {
	// +optional
	Containers []ContainerReliability `json:"containers,omitempty"`
	// Restarts other than OOM kills, approximated like OOMKills.
	// +optional
	Restarts int32 `json:"restarts,omitempty"`
	// Containers in CrashLoopBackOff now.
	// +optional
	CrashLooping []string `json:"crashLooping,omitempty"`
	// Pods with no requests or limits at all, the first evicted under node pressure.
	// +optional
	BestEffortPods int32 `json:"bestEffortPods,omitempty"`
	// The HPA runs at its maximum replicas, so it cannot absorb more load.
	// +optional
	HPAAtMax bool `json:"hpaAtMax,omitempty"`
}

type ContainerReliability struct {
	Name string `json:"name"`
	// Share of CFS periods throttled by the CPU limit, in percent.
	// +optional
	CPUThrottledPercent int32 `json:"cpuThrottledPercent,omitempty"`
	// Peak memory of the busiest pod, in percent of the memory limit.
	// +optional
	MemoryPeakPercent int32 `json:"memoryPeakPercent,omitempty"`
}

type ContainerChange struct {
	Name string                      `json:"name"`
	From corev1.ResourceRequirements `json:"from"`
	To   corev1.ResourceRequirements `json:"to"`
}

type HistoryEntry struct {
	Time       metav1.Time       `json:"time"`
	Action     Action            `json:"action"`
	By         string            `json:"by"`
	Hash       string            `json:"hash,omitempty"`
	Containers []ContainerChange `json:"containers"`
}

type ApplyResult struct {
	Time metav1.Time `json:"time"`
	// +kubebuilder:validation:Enum=Succeeded;Failed;Rejected
	Result  string `json:"result"`
	Message string `json:"message,omitempty"`
}

type RecommendationStatus struct {
	// +optional
	Phase Phase `json:"phase,omitempty"`
	// TemplateRef is the object whose pod template is changed on apply, when it is not the target
	// itself: the Deployment an Argo Rollout references with spec.workloadRef.
	// +optional
	TemplateRef *TargetRef `json:"templateRef,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
	// +optional
	GeneratedAt *metav1.Time `json:"generatedAt,omitempty"`
	// +optional
	NextScheduleAt *metav1.Time `json:"nextScheduleAt,omitempty"`
	// Hash identifies the recommended values.
	// +optional
	Hash string `json:"hash,omitempty"`
	// Pending is true when recommended requests differ from current ones by more than the auto-deploy threshold.
	// +optional
	Pending bool `json:"pending,omitempty"`
	// +optional
	Settings map[string]Setting `json:"settings,omitempty"`
	// +optional
	Containers []ContainerRecommendation `json:"containers,omitempty"`
	// +optional
	HPA *HPARecommendation `json:"hpa,omitempty"`
	// +optional
	Score *Score `json:"score,omitempty"`
	// +optional
	Impact *Impact `json:"impact,omitempty"`
	// +optional
	Pricing *Pricing `json:"pricing,omitempty"`
	// OOM kills seen during the learning period.
	// +optional
	OOMKills int32 `json:"oomKills,omitempty"`
	// +optional
	Reliability *Reliability `json:"reliability,omitempty"`
	// +optional
	Warnings []string `json:"warnings,omitempty"`

	// Fields below are written by the applier.

	// Newest first, at most 10 entries.
	// +optional
	History []HistoryEntry `json:"history,omitempty"`
	// +optional
	LastHandledRequestID string `json:"lastHandledRequestID,omitempty"`
	// +optional
	LastHandledRequestAt *metav1.Time `json:"lastHandledRequestAt,omitempty"`
	// +optional
	LastApplyResult *ApplyResult `json:"lastApplyResult,omitempty"`
}

// Recommendation holds the resource recommendation for one workload.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rec;recs
// +kubebuilder:printcolumn:name="Kind",type=string,JSONPath=`.spec.targetRef.kind`
// +kubebuilder:printcolumn:name="Target",type=string,JSONPath=`.spec.targetRef.name`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Score",type=integer,JSONPath=`.status.score.overall`
// +kubebuilder:printcolumn:name="Cost/mo",type=string,JSONPath=`.status.impact.current.monthlyCost`
// +kubebuilder:printcolumn:name="Recommended/mo",type=string,JSONPath=`.status.impact.recommended.monthlyCost`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Recommendation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RecommendationSpec   `json:"spec,omitempty"`
	Status RecommendationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type RecommendationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Recommendation `json:"items"`
}
