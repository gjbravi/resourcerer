// Command resourcerer manages Vertical Pod Autoscalers for Kubernetes workloads
// and serves a UI to review their recommendations.
package main

import (
	"context"
	"flag"
	"os"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/gjbravi/resourcerer/internal/config"
	"github.com/gjbravi/resourcerer/internal/controller"
	"github.com/gjbravi/resourcerer/internal/scheme"
	"github.com/gjbravi/resourcerer/internal/server"
	"github.com/gjbravi/resourcerer/web"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	var (
		selection   string
		defaultMode string
		excluded    string
		httpAddr    string
		metricsAddr string
		probeAddr   string
		leaderElect bool
		readOnly    bool
		disableUI   bool
	)
	flag.StringVar(&selection, "selection", string(config.SelectionAll),
		`which workloads get a VPA: "all" (opt out with label `+config.LabelEnabled+`=false) or "label" (opt in with `+config.LabelEnabled+`=true on the workload or namespace)`)
	flag.StringVar(&defaultMode, "default-mode", string(config.ModeOff), `VPA mode when no `+config.AnnotationMode+` annotation is set: "Off" or "Auto"`)
	flag.StringVar(&excluded, "excluded-namespaces", "kube-system,kube-public,kube-node-lease", "comma-separated namespaces that are never managed")
	flag.StringVar(&httpAddr, "http-bind-address", ":8080", "address for the API and UI; empty disables")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8081", "address for Prometheus metrics; 0 disables")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8082", "address for liveness and readiness probes")
	flag.BoolVar(&leaderElect, "leader-elect", false, "enable leader election for the controller")
	flag.BoolVar(&readOnly, "read-only", false, "serve the UI without mode changes")
	flag.BoolVar(&disableUI, "disable-ui", false, "serve the API without the embedded UI")
	zapOpts := zap.Options{}
	zapOpts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zapOpts)))
	logger := ctrl.Log.WithName("setup")

	sel, err := config.ParseSelection(selection)
	if err != nil {
		logger.Error(err, "invalid --selection")
		os.Exit(2)
	}
	mode, err := config.ParseMode(defaultMode)
	if err != nil {
		logger.Error(err, "invalid --default-mode")
		os.Exit(2)
	}
	opts := config.Options{Selection: sel, DefaultMode: mode, ExcludedNamespaces: splitList(excluded)}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme.New(),
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         leaderElect,
		LeaderElectionID:       "resourcerer.resourcerer.io",
	})
	if err != nil {
		logger.Error(err, "create manager")
		os.Exit(1)
	}
	if err := controller.Setup(mgr, opts); err != nil {
		logger.Error(err, "set up controllers")
		os.Exit(1)
	}
	if httpAddr != "" {
		srv := &server.Server{Client: mgr.GetClient(), Options: opts, Version: version, ReadOnly: readOnly}
		if !disableUI {
			srv.UI = web.FS()
		}
		if err := mgr.Add(everyReplica(srv.Start(httpAddr))); err != nil {
			logger.Error(err, "add HTTP server")
			os.Exit(1)
		}
	}
	_ = mgr.AddHealthzCheck("healthz", healthz.Ping)
	_ = mgr.AddReadyzCheck("readyz", healthz.Ping)

	logger.Info("starting", "version", version, "selection", sel, "defaultMode", mode, "excluded", opts.ExcludedNamespaces)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error(err, "run manager")
		os.Exit(1)
	}
}

// everyReplica runs a function on all replicas, not only the elected leader.
type everyReplica func(context.Context) error

func (f everyReplica) Start(ctx context.Context) error { return f(ctx) }
func (everyReplica) NeedLeaderElection() bool          { return false }

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
