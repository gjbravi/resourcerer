package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DecisionModelDefault is the name of the DecisionModel the server asks.
const DecisionModelDefault = "default"

// DecisionProvider is the kind of server behind a DecisionModel's endpoint.
// +kubebuilder:validation:Enum=decider;typesafe
type DecisionProvider string

const (
	// DecisionProviderDecider is a self-hosted decider server: no credentials, no data leaving the cluster.
	DecisionProviderDecider DecisionProvider = "decider"
	// DecisionProviderTypeSafe is TypeSafe's managed Jev API, authenticated with a token from a Secret.
	DecisionProviderTypeSafe DecisionProvider = "typesafe"
)

// SecretKeyRef names one key of a Secret in resourcerer's namespace.
type SecretKeyRef struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`
}

// +kubebuilder:validation:XValidation:rule="self.provider != 'typesafe' || has(self.credentialsSecretRef)",message="typesafe needs credentialsSecretRef"
type DecisionModelSpec struct {
	Provider DecisionProvider `json:"provider"`
	// Endpoint is the URL of the model's POST /v1/systemone.
	// +kubebuilder:validation:Pattern=`^https?://`
	Endpoint string `json:"endpoint"`
	// Model names the model behind the endpoint, such as decider-4b: sent with every request, for a gateway that routes
	// by it, and shown with every answer.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	Model string `json:"model"`
	// CredentialsSecretRef holds the API token: required for typesafe, and for a decider behind a gateway that wants one.
	// +optional
	CredentialsSecretRef *SecretKeyRef `json:"credentialsSecretRef,omitempty"`
	// TimeoutSeconds bounds one request to the model.
	// +optional
	// +kubebuilder:default=30
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=300
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`
}

// DecisionModel is the model a user can ask, from the workload page, whether a recommendation is safe to apply. Its
// answer helps the user decide; it never applies anything. Only the one named "default" is used.
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=dm,scope=Cluster
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
// +kubebuilder:printcolumn:name="Model",type=string,JSONPath=`.spec.model`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type DecisionModel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DecisionModelSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
type DecisionModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DecisionModel `json:"items"`
}
