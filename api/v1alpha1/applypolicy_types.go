package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ApplyPolicyMode says whether a policy blocks the applies it governs or only reports them.
// +kubebuilder:validation:Enum=Enforce;Audit
type ApplyPolicyMode string

const (
	// ApplyPolicyEnforce blocks the applies that break the policy.
	ApplyPolicyEnforce ApplyPolicyMode = "Enforce"
	// ApplyPolicyAudit lets them through and records them in wouldBlock, to try a policy first.
	ApplyPolicyAudit ApplyPolicyMode = "Audit"
)

// ApplyKind is a kind of apply a policy can govern.
// +kubebuilder:validation:Enum=Auto;Manual
type ApplyKind string

const (
	// ApplyKindAuto is an apply by auto-deploy.
	ApplyKindAuto ApplyKind = "Auto"
	// ApplyKindManual is an apply a person asked for, from the UI or the API.
	ApplyKindManual ApplyKind = "Manual"
)

// Weekday names a day a window opens on.
// +kubebuilder:validation:Enum=Mon;Tue;Wed;Thu;Fri;Sat;Sun
type Weekday string

const (
	Monday    Weekday = "Mon"
	Tuesday   Weekday = "Tue"
	Wednesday Weekday = "Wed"
	Thursday  Weekday = "Thu"
	Friday    Weekday = "Fri"
	Saturday  Weekday = "Sat"
	Sunday    Weekday = "Sun"
)

// ApplyWindow is a recurring period, in a time zone, an apply may start in.
// +kubebuilder:validation:XValidation:rule="self.from != self.to",message="to must differ from from"
type ApplyWindow struct {
	// Days are the weekdays the window opens on. Empty means every day. A window whose To is earlier
	// than its From crosses midnight and closes on the day after the one it opened on.
	// +optional
	// +kubebuilder:validation:MaxItems=7
	Days []Weekday `json:"days,omitempty"`
	// From is the local time the window opens, HH:MM.
	// +kubebuilder:validation:MaxLength=5
	// +kubebuilder:validation:Pattern=`^([01][0-9]|2[0-3]):[0-5][0-9]$`
	From string `json:"from"`
	// To is the local time the window closes, HH:MM, exclusive.
	// +kubebuilder:validation:MaxLength=5
	// +kubebuilder:validation:Pattern=`^([01][0-9]|2[0-3]):[0-5][0-9]$`
	To string `json:"to"`
	// TimeZone is an IANA name such as America/Sao_Paulo. Empty means UTC. With it, the window is open
	// while the local clock shows a time in [from, to): a time the clock skips is the instant it
	// resumes, a time it repeats is its first occurrence, so a window never opens early.
	// +optional
	// +kubebuilder:validation:MaxLength=64
	TimeZone string `json:"timeZone,omitempty"`
}

// ApplyFreeze is a period nothing is applied in, windows or not.
// +kubebuilder:validation:XValidation:rule="self.from < self.to",message="to must be after from"
type ApplyFreeze struct {
	From metav1.Time `json:"from"`
	// To is exclusive.
	To metav1.Time `json:"to"`
	// Reason explains the freeze in the decision and the UI.
	// +optional
	// +kubebuilder:validation:MaxLength=256
	Reason string `json:"reason,omitempty"`
}

// ApplyRate bounds how many applies happen, and how close together.
type ApplyRate struct {
	// MaxConcurrent is the applies in progress at once among the workloads the policy selects: the
	// rollout has not finished yet. A whole number, or a percentage of those workloads with
	// auto-deploy enabled, rounded up. 0 blocks every automatic apply.
	// +optional
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern=`^([0-9]{1,4}|[0-9]{1,3}%)$`
	MaxConcurrent *intstr.IntOrString `json:"maxConcurrent,omitempty"`
	// MaxPerHour is the applies allowed in any hour among the workloads the policy selects. 0 blocks them.
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxPerHour *int32 `json:"maxPerHour,omitempty"`
	// PerNamespace bounds the applies in progress in one namespace, the workload's, counting rollouts
	// resourcerer did not start too.
	// +optional
	PerNamespace *ApplyNamespaceRate `json:"perNamespace,omitempty"`
	// Cooldown is the minimum time between two changes to the same workload, an apply or a revert,
	// as a Go duration such as 24h.
	// +optional
	Cooldown *metav1.Duration `json:"cooldown,omitempty"`
	// Spread delays each held apply to a stable offset in [0, spread) after the window opens, so a herd
	// does not start every rollout at once. Capped at the window's length.
	// +optional
	Spread *metav1.Duration `json:"spread,omitempty"`
}

// ApplyNamespaceRate bounds the applies in progress in a namespace.
type ApplyNamespaceRate struct {
	// MaxConcurrent is the applies in progress in the namespace, resourcerer's and anyone else's
	// rollouts alike. A whole number, or a percentage of the managed workloads in the namespace,
	// rounded up. 1 allows one rollout at a time in the namespace.
	// +optional
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern=`^([0-9]{1,4}|[0-9]{1,3}%)$`
	MaxConcurrent *intstr.IntOrString `json:"maxConcurrent,omitempty"`
}

// ApplyRules are the rules a policy sets, shared by ApplyPolicy and ClusterApplyPolicy.
type ApplyRules struct {
	// WorkloadSelector narrows the workloads the policy governs, matched against the workload's labels
	// and its pod template's. Empty governs every workload in scope.
	// +optional
	WorkloadSelector *metav1.LabelSelector `json:"workloadSelector,omitempty"`
	// AppliesTo is the kinds of apply the policy governs: Auto, Manual, or both. Empty means Auto, so a
	// policy never holds a person back unless it says so.
	// +optional
	// +kubebuilder:validation:MaxItems=2
	AppliesTo []ApplyKind `json:"appliesTo,omitempty"`
	// Mode is Enforce, which blocks, or Audit, which only records what it would have blocked.
	// +kubebuilder:default=Enforce
	// +optional
	Mode ApplyPolicyMode `json:"mode,omitempty"`
	// Windows are the periods an apply may start in. Empty means any time. They are a union: inside any
	// of them the apply may start, and the freezes still hold.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	Windows []ApplyWindow `json:"windows,omitempty"`
	// Freezes are the periods nothing is applied in, windows or not.
	// +optional
	// +kubebuilder:validation:MaxItems=64
	Freezes []ApplyFreeze `json:"freezes,omitempty"`
	// Rate bounds how many applies happen, and how close together.
	// +optional
	Rate *ApplyRate `json:"rate,omitempty"`
}

// ApplyPolicySpec governs the applies of the workloads in its own namespace.
type ApplyPolicySpec struct {
	ApplyRules `json:",inline"`
}

// ClusterApplyPolicySpec governs the applies of the workloads in the namespaces it selects.
type ClusterApplyPolicySpec struct {
	// NamespaceSelector narrows the namespaces the policy governs. Empty governs every managed namespace.
	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`
	ApplyRules        `json:",inline"`
}

// ApplyDecisionState is whether an apply may start now.
// +kubebuilder:validation:Enum=Allowed;Blocked
type ApplyDecisionState string

const (
	ApplyAllowed ApplyDecisionState = "Allowed"
	ApplyBlocked ApplyDecisionState = "Blocked"
)

// ApplyDecision is what the policies that govern a workload decided for one kind of apply.
type ApplyDecision struct {
	State ApplyDecisionState `json:"state"`
	// Policy is the policy that blocks, Kind/name (or Kind/namespace/name), cluster policies first and
	// then by name, so the same rules always name the same policy. Empty when allowed.
	// +optional
	Policy string `json:"policy,omitempty"`
	// BlockedBy is the rule that blocks, as the spec field path: windows, freezes, rate.cooldown,
	// rate.maxConcurrent, rate.maxPerHour, rate.perNamespace, rate.spread, rollout (the workload's own
	// rollout is still running) or policy (a policy could not be read).
	// +optional
	BlockedBy string `json:"blockedBy,omitempty"`
	// Reason says the same for people.
	// +optional
	Reason string `json:"reason,omitempty"`
	// NextAllowed is when an apply may start. Absent while waiting for a slot, which depends on other
	// rollouts finishing.
	// +optional
	NextAllowed *metav1.Time `json:"nextAllowed,omitempty"`
	// WouldBlock lists what Audit-mode policies would have blocked, as "<policy>: <reason>".
	// +optional
	WouldBlock []string `json:"wouldBlock,omitempty"`
}

// ApplyPolicyStatus is what the policies decided for one workload's two kinds of apply. They are decided
// apart, because "auto-deploy is held, a person may still apply" is a common state.
type ApplyPolicyStatus struct {
	// +optional
	Auto ApplyDecision `json:"auto,omitempty"`
	// +optional
	Manual ApplyDecision `json:"manual,omitempty"`
	// EvaluatedAt is when this decision last changed, not when it was last computed.
	// +optional
	EvaluatedAt *metav1.Time `json:"evaluatedAt,omitempty"`
}

// ApplyPolicy governs when the workloads in its namespace are applied: a team writes its own rules
// without touching the platform's configuration, and can only tighten what the cluster policies set.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ap
// +kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.spec.mode`
// +kubebuilder:printcolumn:name="Applies to",type=string,JSONPath=`.spec.appliesTo`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type ApplyPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplyPolicySpec   `json:"spec,omitempty"`
	Status ApplyPolicyHealth `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ApplyPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplyPolicy `json:"items"`
}

// ClusterApplyPolicy governs when the workloads in the namespaces it selects are applied. A namespaced
// ApplyPolicy can tighten it and never loosen it.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=cap,scope=Cluster
// +kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.spec.mode`
// +kubebuilder:printcolumn:name="Applies to",type=string,JSONPath=`.spec.appliesTo`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type ClusterApplyPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterApplyPolicySpec `json:"spec,omitempty"`
	Status ApplyPolicyHealth      `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ClusterApplyPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterApplyPolicy `json:"items"`
}

// ApplyPolicyHealth says how many managed workloads a policy governs, and whether resourcerer can
// interpret it. A policy it cannot read holds the applies it governs, so the condition says why.
type ApplyPolicyHealth struct {
	// Workloads is how many managed workloads the policy currently governs. Nil until resourcerer has
	// looked; zero says the policy selects nothing, which is worth seeing.
	// +optional
	Workloads *int32 `json:"workloads,omitempty"`
	// Conditions hold Valid: False with the error when the policy cannot be interpreted.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
