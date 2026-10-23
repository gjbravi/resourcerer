// Package config defines how resourcerer decides which workloads get a VPA and how
// that VPA behaves, based on labels and annotations on workloads and namespaces.
package config

import (
	"fmt"
	"slices"
	"strings"
)

const Domain = "resourcerer.io"

const (
	// LabelEnabled opts a namespace or workload in ("true") or out ("false").
	LabelEnabled = Domain + "/enabled"
	// AnnotationMode sets the VPA update mode for a workload or, as default, for a namespace.
	AnnotationMode = Domain + "/mode"
	// LabelManaged marks VPAs created by resourcerer; others are never touched.
	LabelManaged = Domain + "/managed"
)

// Selection controls which workloads get a VPA.
type Selection string

const (
	// SelectionAll creates a VPA for every workload unless opted out, like Goldilocks --on-by-default.
	SelectionAll Selection = "all"
	// SelectionLabel creates a VPA only when the workload or its namespace is labeled enabled=true.
	SelectionLabel Selection = "label"
)

func ParseSelection(s string) (Selection, error) {
	switch Selection(s) {
	case SelectionAll, SelectionLabel:
		return Selection(s), nil
	}
	return "", fmt.Errorf("invalid selection %q: want %q or %q", s, SelectionAll, SelectionLabel)
}

// Mode is the user-facing management mode, mapped onto the VPA update mode.
type Mode string

const (
	// ModeOff: VPA computes recommendations only (updateMode Off).
	ModeOff Mode = "Off"
	// ModeAuto: VPA resizes pods in place when possible, otherwise recreates them (updateMode InPlaceOrRecreate).
	ModeAuto Mode = "Auto"
)

func ParseMode(s string) (Mode, error) {
	for _, m := range []Mode{ModeOff, ModeAuto} {
		if strings.EqualFold(s, string(m)) {
			return m, nil
		}
	}
	return "", fmt.Errorf("invalid mode %q: want %q or %q", s, ModeOff, ModeAuto)
}

// Options are cluster-wide settings from flags.
type Options struct {
	Selection          Selection
	DefaultMode        Mode
	ExcludedNamespaces []string
}

func (o Options) Excluded(ns string) bool { return slices.Contains(o.ExcludedNamespaces, ns) }

// Decision is the effective configuration for one workload.
type Decision struct {
	Managed bool
	Mode    Mode
	// Reason explains Managed=false, or which level set the mode.
	Reason string
	// Warning reports an invalid label or annotation value that was ignored.
	Warning string
}

// Decide applies, in order: excluded namespace, workload label, namespace label, selection default.
// The mode comes from the workload annotation, then the namespace annotation, then the default.
func Decide(o Options, namespace string, nsLabels, nsAnnotations, wlLabels, wlAnnotations map[string]string) Decision {
	if o.Excluded(namespace) {
		return Decision{Reason: "namespace is excluded"}
	}
	d := Decision{}
	wl, wlSet := parseBool(wlLabels[LabelEnabled])
	ns, nsSet := parseBool(nsLabels[LabelEnabled])
	switch {
	case wlSet:
		d.Managed, d.Reason = wl, "workload label"
	case nsSet:
		d.Managed, d.Reason = ns, "namespace label"
	default:
		d.Managed, d.Reason = o.Selection == SelectionAll, "selection "+string(o.Selection)
	}
	for _, v := range []struct{ raw, where string }{{wlLabels[LabelEnabled], "workload"}, {nsLabels[LabelEnabled], "namespace"}} {
		if _, ok := parseBool(v.raw); !ok && v.raw != "" {
			d.Warning = fmt.Sprintf("ignoring %s label %s=%q: want true or false", v.where, LabelEnabled, v.raw)
		}
	}
	if !d.Managed {
		return d
	}

	d.Mode = o.DefaultMode
	for _, v := range []struct{ raw, where string }{{nsAnnotations[AnnotationMode], "namespace"}, {wlAnnotations[AnnotationMode], "workload"}} {
		if v.raw == "" {
			continue
		}
		m, err := ParseMode(v.raw)
		if err != nil {
			d.Warning = fmt.Sprintf("ignoring %s annotation %s: %v", v.where, AnnotationMode, err)
			continue
		}
		d.Mode = m
	}
	return d
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
