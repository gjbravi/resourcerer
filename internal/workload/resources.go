package workload

import (
	"math"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/gjbravi/resourcerer/internal/engine"
)

// ToValues converts a ResourceList into cores and bytes.
func ToValues(l corev1.ResourceList) engine.Values {
	v := engine.Values{}
	if q, ok := l[corev1.ResourceCPU]; ok {
		v[engine.CPU] = q.AsApproximateFloat64()
	}
	if q, ok := l[corev1.ResourceMemory]; ok {
		v[engine.Memory] = q.AsApproximateFloat64()
	}
	return v
}

// ToList converts cores and bytes into quantities in their usual units (m, Mi).
func ToList(v engine.Values) corev1.ResourceList {
	l := corev1.ResourceList{}
	if c, ok := v[engine.CPU]; ok {
		l[corev1.ResourceCPU] = *resource.NewMilliQuantity(int64(math.Round(c*1000)), resource.DecimalSI)
	}
	if m, ok := v[engine.Memory]; ok {
		l[corev1.ResourceMemory] = resource.MustParse(strconv.FormatInt(int64(math.Ceil(m/engine.Mi)), 10) + "Mi")
	}
	if len(l) == 0 {
		return nil
	}
	return l
}
