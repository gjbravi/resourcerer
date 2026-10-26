// Command resourcerer recommends CPU and memory requests for Kubernetes workloads from
// Prometheus usage history, applies them on request or automatically, and serves a UI.
//
// One binary runs one or more components:
//
//	resourcerer controller   discovery + recommender (leader-elected)
//	resourcerer applier      writes workloads: Apply, Revert, auto-deploy (leader-elected)
//	resourcerer server       REST API + UI (every replica)
//	resourcerer all          everything in one process, for local runs
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/gjbravi/resourcerer/internal/applier"
	"github.com/gjbravi/resourcerer/internal/cluster"
	"github.com/gjbravi/resourcerer/internal/prom"
	"github.com/gjbravi/resourcerer/internal/recommender"
	"github.com/gjbravi/resourcerer/internal/scheme"
	"github.com/gjbravi/resourcerer/internal/server"
	"github.com/gjbravi/resourcerer/web"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

var components = []string{"controller", "applier", "server"}

func main() {
	if len(os.Args) < 2 || !slices.Contains(append(components, "all"), os.Args[1]) {
		fmt.Fprintf(os.Stderr, "usage: resourcerer <controller|applier|server|all> [flags]\n")
		os.Exit(2)
	}
	role := os.Args[1]
	run := map[string]bool{}
	for _, c := range components {
		run[c] = role == "all" || role == c
	}

	fs := flag.NewFlagSet(role, flag.ExitOnError)
	var (
		clusterName = fs.String("cluster-name", "kubernetes", "name shown in the UI")
		selection   = fs.String("selection", "all", `"all" (opt out with resourcerer.io/enabled=false) or "label" (opt in with resourcerer.io/enabled=true)`)
		excluded    = fs.String("excluded-namespaces", "kube-system,kube-public,kube-node-lease", "comma-separated namespaces never managed")
		defaultsCM  = fs.String("defaults-configmap", "", "namespace/name of the ConfigMap with cluster defaults and cost rates")
		promURL     = fs.String("prometheus-url", "http://prometheus-server.monitoring.svc", "Prometheus-compatible API base URL")
		rateWindow  = fs.String("rate-window", "5m", "rate() window for CPU usage; at least 4× the scrape interval")
		minStep     = fs.Duration("min-step", 30*time.Second, "finest query resolution; about the scrape interval")
		httpAddr    = fs.String("http-bind-address", ":8080", "address for the API and UI")
		metricsAddr = fs.String("metrics-bind-address", ":8081", "address for Prometheus metrics; 0 disables")
		probeAddr   = fs.String("health-probe-bind-address", ":8082", "address for liveness and readiness probes")
		leaderElect = fs.Bool("leader-elect", false, "enable leader election for controller and applier")
		readOnly    = fs.Bool("read-only", false, "server: refuse every change")
		disableUI   = fs.Bool("disable-ui", false, "server: serve the API without the embedded UI")
	)
	zapOpts := zap.Options{}
	zapOpts.BindFlags(fs)
	_ = fs.Parse(os.Args[2:])

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zapOpts)))
	logger := ctrl.Log.WithName("setup")
	fail := func(err error, msg string) {
		logger.Error(err, msg)
		os.Exit(1)
	}

	cfg := cluster.Config{Name: *clusterName, Selection: cluster.Selection(*selection), ExcludedNamespaces: splitList(*excluded)}
	if cfg.Selection != cluster.SelectionAll && cfg.Selection != cluster.SelectionLabel {
		fail(fmt.Errorf("got %q", *selection), "--selection must be all or label")
	}
	if *defaultsCM != "" {
		ns, name, ok := strings.Cut(*defaultsCM, "/")
		if !ok {
			fail(fmt.Errorf("got %q", *defaultsCM), "--defaults-configmap must be namespace/name")
		}
		cfg.DefaultsConfigMap = types.NamespacedName{Namespace: ns, Name: name}
	}
	p, err := prom.New(*promURL, *rateWindow)
	if err != nil {
		fail(err, "create Prometheus client")
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme.New(),
		Metrics:                metricsserver.Options{BindAddress: *metricsAddr},
		HealthProbeBindAddress: *probeAddr,
		LeaderElection:         *leaderElect && (run["controller"] || run["applier"]),
		LeaderElectionID:       role + ".resourcerer.io",
	})
	if err != nil {
		fail(err, "create manager")
	}
	if run["controller"] {
		if err := recommender.Setup(mgr, cfg, p, *minStep); err != nil {
			fail(err, "set up recommender")
		}
	}
	if run["applier"] {
		if err := applier.Setup(mgr); err != nil {
			fail(err, "set up applier")
		}
	}
	if run["server"] {
		srv := &server.Server{Client: mgr.GetClient(), APIReader: mgr.GetAPIReader(), Cluster: cfg, Prom: p, PromURL: *promURL,
			MinStep: *minStep, Version: version, ReadOnly: *readOnly}
		if !*disableUI {
			srv.UI = web.FS()
		}
		if err := mgr.Add(everyReplica(srv.Start(*httpAddr))); err != nil {
			fail(err, "add HTTP server")
		}
	}
	_ = mgr.AddHealthzCheck("healthz", healthz.Ping)
	_ = mgr.AddReadyzCheck("readyz", healthz.Ping)

	logger.Info("starting", "role", role, "version", version, "cluster", cfg.Name, "selection", cfg.Selection,
		"prometheus", *promURL, "readOnly", *readOnly)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		fail(err, "run manager")
	}
}

// everyReplica runs on all replicas, not only the elected leader.
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
