package server

// Mode as shown in the UI; Disabled means no VPA.
type Mode string

const (
	ModeOff      Mode = "Off"
	ModeAuto     Mode = "Auto"
	ModeDisabled Mode = "Disabled"
)

// Status summarizes the VPA state for display.
type Status string

const (
	StatusPending       Status = "Pending"
	StatusReady         Status = "Ready"
	StatusLowConfidence Status = "LowConfidence"
	StatusConflict      Status = "Conflict"
	StatusError         Status = "Error"
	StatusNone          Status = "None"
)

type Resources struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type Container struct {
	Name     string    `json:"name"`
	Requests Resources `json:"requests"`
	Limits   Resources `json:"limits"`
	// Recommendation fields come from VPA status and are empty until VPA has data.
	Target     *Resources `json:"target,omitempty"`
	LowerBound *Resources `json:"lowerBound,omitempty"`
	UpperBound *Resources `json:"upperBound,omitempty"`
	Uncapped   *Resources `json:"uncappedTarget,omitempty"`
}

type Workload struct {
	Namespace  string      `json:"namespace"`
	Kind       string      `json:"kind"`
	Name       string      `json:"name"`
	Replicas   int32       `json:"replicas"`
	Mode       Mode        `json:"mode"`
	Status     Status      `json:"status"`
	Message    string      `json:"message,omitempty"`
	VPA        string      `json:"vpa,omitempty"`
	Containers []Container `json:"containers"`
	// ModeSource says whether the mode comes from the workload, the namespace or the default.
	ModeSource string `json:"modeSource"`
}

type Pod struct {
	Name       string      `json:"name"`
	Containers []Container `json:"containers"`
}

type Event struct {
	Type    string `json:"type"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

type WorkloadDetail struct {
	Workload
	Pods   []Pod   `json:"pods"`
	Events []Event `json:"events"`
}

type ModeRequest struct {
	Mode Mode `json:"mode"`
}

type Info struct {
	Version     string `json:"version"`
	Selection   string `json:"selection"`
	DefaultMode string `json:"defaultMode"`
	ReadOnly    bool   `json:"readOnly"`
}
