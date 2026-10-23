package config

import "testing"

func TestDecide(t *testing.T) {
	all := Options{Selection: SelectionAll, DefaultMode: ModeOff, ExcludedNamespaces: []string{"kube-system"}}
	label := Options{Selection: SelectionLabel, DefaultMode: ModeOff}
	on := map[string]string{LabelEnabled: "true"}
	off := map[string]string{LabelEnabled: "false"}
	auto := map[string]string{AnnotationMode: "Auto"}

	cases := []struct {
		name          string
		opts          Options
		ns            string
		nsL, nsA, wlL map[string]string
		wlA           map[string]string
		managed       bool
		mode          Mode
		warning       bool
	}{
		{name: "all: default managed", opts: all, ns: "shop", managed: true, mode: ModeOff},
		{name: "all: excluded namespace", opts: all, ns: "kube-system", wlL: on, managed: false},
		{name: "all: workload opt-out", opts: all, ns: "shop", wlL: off, managed: false},
		{name: "all: namespace opt-out", opts: all, ns: "shop", nsL: off, managed: false},
		{name: "all: workload overrides namespace opt-out", opts: all, ns: "shop", nsL: off, wlL: on, managed: true, mode: ModeOff},
		{name: "label: default unmanaged", opts: label, ns: "shop", managed: false},
		{name: "label: namespace opt-in", opts: label, ns: "shop", nsL: on, managed: true, mode: ModeOff},
		{name: "label: workload opt-in", opts: label, ns: "shop", wlL: on, managed: true, mode: ModeOff},
		{name: "label: workload opt-out wins over namespace", opts: label, ns: "shop", nsL: on, wlL: off, managed: false},
		{name: "namespace mode default", opts: all, ns: "shop", nsA: auto, managed: true, mode: ModeAuto},
		{name: "workload mode wins", opts: all, ns: "shop", nsA: auto, wlA: map[string]string{AnnotationMode: "off"}, managed: true, mode: ModeOff},
		{name: "invalid mode ignored", opts: all, ns: "shop", wlA: map[string]string{AnnotationMode: "Recreate"}, managed: true, mode: ModeOff, warning: true},
		{name: "invalid label ignored", opts: label, ns: "shop", wlL: map[string]string{LabelEnabled: "yes"}, managed: false, warning: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := Decide(c.opts, c.ns, c.nsL, c.nsA, c.wlL, c.wlA)
			if d.Managed != c.managed || d.Mode != c.mode || (d.Warning != "") != c.warning {
				t.Errorf("got %+v, want managed=%v mode=%q warning=%v", d, c.managed, c.mode, c.warning)
			}
		})
	}
}

func TestParse(t *testing.T) {
	if _, err := ParseSelection("some"); err == nil {
		t.Error("ParseSelection should reject unknown values")
	}
	if m, err := ParseMode("auto"); err != nil || m != ModeAuto {
		t.Errorf("ParseMode(auto) = %q, %v", m, err)
	}
}
