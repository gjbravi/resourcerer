package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicyMode is whether a policy blocks an apply or only reports what it would have blocked.
// +kubebuilder:validation:Enum=Enforce;Audit
type PolicyMode string

const (
	// PolicyModeEnforce blocks applies that break the policy.
	PolicyModeEnforce PolicyMode = "Enforce"
	// PolicyModeAudit allows them and records them in the decision's wouldBlock list, to try a policy
	// before enforcing it.
	PolicyModeAudit PolicyMode = "Audit"
)

// AppliesTo is a kind of apply a policy governs.
// +kubebuilder:validation:Enum=Auto;Manual
type AppliesTo string

const (
	// AppliesToAuto is an apply by auto-deploy.
	AppliesToAuto AppliesTo = "Auto"
	// AppliesToManual is an apply requested from the UI or the API.
	AppliesToManual AppliesTo = "Manual"
)

// ApplyWhen says when an apply may start.
type ApplyWhen struct {
	// Windows are the periods an apply may start in. Empty: any time. Windows are a union: inside any
	// of them the apply may start, and freezes still apply.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	Windows []ApplyWindow `json:"windows,omitempty"`
	// Freezes are periods nothing may be applied in, windows or not.
	// +optional
	// +kubebuilder:validation:MaxItems=64
	Freezes []ApplyFreeze `json:"freezes,omitempty"`
}

// ApplyWindow is a recurring period, in a time zone, an apply may start in.
// +kubebuilder:validation:XValidation:rule="self.from != self.to",message="a window cannot open and close at the same time"
type ApplyWindow struct {
	// Days are the weekdays the window opens on, Mon to Sun, in TimeZone. Empty: every day. An overnight window
	// opens on its From day and closes on the next one.
	// +optional
	// +kubebuilder:validation:items:Enum=Mon;Tue;Wed;Thu;Fri;Sat;Sun
	Days []string `json:"days,omitempty"`
	// From is the local time the window opens, HH:MM.
	// +kubebuilder:validation:Pattern=`^([01][0-9]|2[0-3]):[0-5][0-9]$`
	// +kubebuilder:validation:MaxLength=5
	From string `json:"from"`
	// To is the local time the window closes, HH:MM, exclusive. Equal to From is invalid; earlier than
	// From crosses midnight.
	// +kubebuilder:validation:Pattern=`^([01][0-9]|2[0-3]):[0-5][0-9]$`
	// +kubebuilder:validation:MaxLength=5
	To string `json:"to"`
	// TimeZone is an IANA name such as America/Sao_Paulo; the window follows its daylight saving.
	// Empty: UTC.
	// +optional
	// +kubebuilder:validation:MaxLength=64
	TimeZone string `json:"timeZone,omitempty"`
}

// ApplyFreeze is a period nothing may be applied in.
// +kubebuilder:validation:XValidation:rule="self.from < self.to",message="to must be after from"
type ApplyFreeze struct {
	From metav1.Time `json:"from"`
	// To is exclusive.
	To metav1.Time `json:"to"`
	// Reason explains the freeze in the UI and in the decision.
	// +optional
	Reason string `json:"reason,omitempty"`
}

// ApplyLimits bounds how often applies happen.
type ApplyLimits struct {
	// Cooldown is the minimum time between two changes to the same workload, counted from the last
	// apply or revert of any hash, as a Go duration such as `24h`.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self.matches(r'^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$')",message="cooldown must be a duration such as 24h"
	Cooldown *metav1.Duration `json:"cooldown,omitempty"`
}

// ApplyRules is what a policy decides, shared by ApplyPolicy and ClusterApplyPolicy.
type ApplyRules struct {
	// WorkloadSelector narrows the workloads the policy governs, matched against the workload's labels
	// and its pod template's. Empty: every workload in scope.
	// +optional
	WorkloadSelector *metav1.LabelSelector `json:"workloadSelector,omitempty"`
	// Priority decides which policy is reported when more than one blocks: the highest, then cluster
	// policies before namespaced ones, then by name. It never makes a policy stronger or weaker.
	// +optional
	Priority int32 `json:"priority,omitempty"`
	// AppliesTo is the kinds of apply the policy governs. Empty: Auto, so a policy never blocks a
	// person unless it says so.
	// +optional
	AppliesTo []AppliesTo `json:"appliesTo,omitempty"`
	// Mode is Enforce, which blocks, or Audit, which only reports what it would block.
	// +kubebuilder:default=Enforce
	// +optional
	Mode PolicyMode `json:"mode,omitempty"`
	// When is when an apply may start.
	// +optional
	When *ApplyWhen `json:"when,omitempty"`
	// Limits bounds how often applies happen.
	// +optional
	Limits *ApplyLimits `json:"limits,omitempty"`
}

// ApplyPolicySpec governs applies for the workloads in its own namespace.
type ApplyPolicySpec struct {
	ApplyRules `json:",inline"`
}

// ClusterApplyPolicySpec governs applies for the workloads in the namespaces it selects.
type ClusterApplyPolicySpec struct {
	// NamespaceSelector narrows the namespaces the policy governs. Empty: every managed namespace.
	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`
	ApplyRules        `json:",inline"`
}

// PolicyState is the decision the applier reached for a workload.
// +kubebuilder:validation:Enum=Allowed;Blocked
type PolicyState string

const (
	PolicyStateAllowed PolicyState = "Allowed"
	PolicyStateBlocked PolicyState = "Blocked"
)

// ApplyPolicyStatus is what the apply policies decided for one workload. The two kinds of apply are
// decided apart, so the UI can say which one is held back: a policy commonly governs auto-deploy and
// leaves a person free, or the other way around.
type ApplyPolicyStatus struct {
	// +optional
	Auto ApplyDecision `json:"auto,omitempty"`
	// +optional
	Manual ApplyDecision `json:"manual,omitempty"`
	// EvaluatedAt is when the applier reached this decision.
	// +optional
	EvaluatedAt *metav1.Time `json:"evaluatedAt,omitempty"`
}

// ApplyDecision is what the ApplyPolicy objects that govern a workload decided for one kind of apply:
// when it may start, and which policy blocks it otherwise.
type ApplyDecision struct {
	// Policy is the object that blocked, Kind/namespace/name. Empty when nothing blocked.
	// +optional
	Policy string      `json:"policy,omitempty"`
	State  PolicyState `json:"state"`
	// BlockedBy is the field of the policy that blocks, such as when.windows or limits.cooldown, so
	// the UI can point at it instead of parsing a message.
	// +optional
	BlockedBy string `json:"blockedBy,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// NextAllowed is when an apply may start again; the applier requeues for it.
	// +optional
	NextAllowed *metav1.Time `json:"nextAllowed,omitempty"`
	// WouldBlock lists what Audit-mode policies would have blocked, in the order Priority reports them.
	// +optional
	WouldBlock []string `json:"wouldBlock,omitempty"`
}

// ApplyPolicy governs when the workloads in its namespace may be applied. A team declares its own
// rules without touching the platform's configuration.
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=ap
// +kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.spec.mode`
// +kubebuilder:printcolumn:name="Applies to",type=string,JSONPath=`.spec.appliesTo`
// +kubebuilder:printcolumn:name="Priority",type=integer,JSONPath=`.spec.priority`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type ApplyPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ApplyPolicySpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
type ApplyPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplyPolicy `json:"items"`
}

// ClusterApplyPolicy governs when the workloads in the namespaces it selects may be applied, and
// cannot be loosened by a namespaced ApplyPolicy.
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=cap,scope=Cluster
// +kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.spec.mode`
// +kubebuilder:printcolumn:name="Applies to",type=string,JSONPath=`.spec.appliesTo`
// +kubebuilder:printcolumn:name="Priority",type=integer,JSONPath=`.spec.priority`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type ClusterApplyPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterApplyPolicySpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
type ClusterApplyPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterApplyPolicy `json:"items"`
}
