// Package cluster holds cluster-wide configuration shared by all components:
// which workloads are managed, the settings defaults and cost rates.
package cluster

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/settings"
	"github.com/gjbravi/resourcerer/internal/signals"
)

type Selection string

const (
	// SelectionAll manages every workload unless opted out, like Goldilocks --on-by-default.
	SelectionAll Selection = "all"
	// SelectionLabel manages only workloads or namespaces labeled resourcerer.io/enabled=true.
	SelectionLabel Selection = "label"
)

// Config is set by flags; defaults and cost rates come from a ConfigMap rendered by the chart.
type Config struct {
	Name               string
	Selection          Selection
	ExcludedNamespaces []string
	DefaultsConfigMap  types.NamespacedName
	// Pricer resolves the cost rates; nil means the configured rates.
	Pricer *Pricer
}

const (
	costCPUKey     = "cost.cpu-core-hour"
	costMemoryKey  = "cost.memory-gib-hour"
	costSourceKey  = "cost.source"
	costAllocKey   = "cost.allocation"
	costRefreshKey = "cost.karpenter.refresh-interval"
)

var DefaultRates = engine.Rates{CPUCoreHour: 0.0316, MemoryGiBHour: 0.0042}

func (c Config) Excluded(ns string) bool { return slices.Contains(c.ExcludedNamespaces, ns) }

// Managed decides whether a workload gets a Recommendation: workload label, then namespace label, then selection.
func (c Config) Managed(ns *corev1.Namespace, wlLabels map[string]string) bool {
	if c.Excluded(ns.Name) {
		return false
	}
	if v, ok := parseBool(wlLabels[settings.LabelEnabled]); ok {
		return v
	}
	if v, ok := parseBool(ns.Labels[settings.LabelEnabled]); ok {
		return v
	}
	return c.Selection == SelectionAll
}

// Defaults reads settings defaults and the cost pricing; a missing ConfigMap means built-in defaults.
func (c Config) Defaults(ctx context.Context, r client.Reader) (map[string]string, Pricing) {
	var cm corev1.ConfigMap
	rates, source := DefaultRates, CostAuto
	if c.DefaultsConfigMap.Name == "" || r.Get(ctx, c.DefaultsConfigMap, &cm) != nil {
		return nil, withAllocation(c.Pricer.Pricing(ctx, rates, source, DefaultRefresh), rates, "")
	}
	if v, err := strconv.ParseFloat(cm.Data[costCPUKey], 64); err == nil {
		rates.CPUCoreHour = v
	}
	if v, err := strconv.ParseFloat(cm.Data[costMemoryKey], 64); err == nil {
		rates.MemoryGiBHour = v
	}
	if v := cm.Data[costSourceKey]; v != "" {
		source = v
	}
	return cm.Data, withAllocation(c.Pricer.Pricing(ctx, rates, source, refreshInterval(cm.Data[costRefreshKey])), rates, cm.Data[costAllocKey])
}

// withAllocation records the configured rates and allocation; anything but node allocates at the cluster rates.
func withAllocation(p Pricing, static engine.Rates, allocation string) Pricing {
	p.Static, p.Allocation = static, AllocationCluster
	if allocation == AllocationNode {
		p.Allocation = AllocationNode
	}
	return p
}

// Reliability reads the reliability signals configuration from the defaults ConfigMap data.
func Reliability(data map[string]string) signals.Config {
	return signals.Parse(data[signals.ConfigKey])
}

// refreshInterval parses the configured refresh, keeping it at or above MinRefresh; invalid values use the default.
func refreshInterval(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return DefaultRefresh
	}
	return max(d, MinRefresh)
}

func parseBool(s string) (value, ok bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true":
		return true, true
	case "false":
		return false, true
	}
	return false, false
}
