package settings

import (
	"testing"
	"time"
)

func TestResolveCascade(t *testing.T) {
	e := Resolve(
		map[string]string{CPUGoal: "Savings"},
		map[string]string{Domain + "/" + CPUGoal: "Reliability", Domain + "/" + AutoDeploy: "Enabled"},
		map[string]string{Domain + "/" + AutoDeploy: "Disabled", Domain + "/" + MemoryGoal: "Bogus"},
	)
	cases := map[string]Value{
		CPUGoal:    {"Reliability", SourceNamespace},
		AutoDeploy: {"Disabled", SourceWorkload},
		MemoryGoal: {"Balanced", SourceCluster},
	}
	for k, want := range cases {
		if got := e.Values[k]; got != want {
			t.Errorf("%s = %+v, want %+v", k, got, want)
		}
	}
	if len(e.Warnings) != 1 {
		t.Errorf("warnings = %v, want one for the invalid memory goal", e.Warnings)
	}
}

func TestParseDuration(t *testing.T) {
	for in, want := range map[string]time.Duration{
		"P7D":    7 * 24 * time.Hour,
		"PT10M":  10 * time.Minute,
		"P1DT6H": 30 * time.Hour,
		"P2W":    14 * 24 * time.Hour,
	} {
		got, err := ParseDuration(in)
		if err != nil || got != want {
			t.Errorf("%s = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, bad := range []string{"7d", "P", "PT", "P1H"} {
		if _, err := ParseDuration(bad); err == nil {
			t.Errorf("%s should fail", bad)
		}
	}
}

func TestParseSchedule(t *testing.T) {
	if d, err := ParseSchedule("@every 2m"); err != nil || d != 2*time.Minute {
		t.Errorf("@every 2m = %v, %v", d, err)
	}
	if _, err := ParseSchedule("@every 10s"); err == nil {
		t.Error("periods under a minute should fail")
	}
}
