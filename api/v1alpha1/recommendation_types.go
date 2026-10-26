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
	// OOM kills seen during the learning period.
	// +optional
	OOMKills int32 `json:"oomKills,omitempty"`
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
