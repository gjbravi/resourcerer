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

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/applier"
	"github.com/gjbravi/resourcerer/internal/cluster"
	"github.com/gjbravi/resourcerer/internal/prom"
	"github.com/gjbravi/resourcerer/internal/recommender"
	"github.com/gjbravi/resourcerer/internal/scheme"
	"github.com/gjbravi/resourcerer/internal/server"
	"github.com/gjbravi/resourcerer/internal/server/auth"
	"github.com/gjbravi/resourcerer/web"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

var buildInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "resourcerer_build_info",
	Help: "1 for the version and component this process runs.",
}, []string{"version", "component"})

func init() {
	metrics.Registry.MustRegister(buildInfo)
}

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
		promTimeout = fs.Duration("prometheus-timeout", 2*time.Minute, "longest a Prometheus query may take, response included; 0 means no limit")
		httpAddr    = fs.String("http-bind-address", ":8080", "address for the API and UI")
		metricsAddr = fs.String("metrics-bind-address", ":8081", "address for Prometheus metrics; 0 disables")
		probeAddr   = fs.String("health-probe-bind-address", ":8082", "address for liveness and readiness probes")
		leaderElect = fs.Bool("leader-elect", false, "enable leader election for controller and applier")
		readOnly    = fs.Bool("read-only", false, "server: refuse every change")
		disableUI   = fs.Bool("disable-ui", false, "server: serve the API without the embedded UI")
		authStrat   = fs.String("auth-strategy", auth.StrategyAnonymous, `server: "anonymous" (anyone who reaches the server may change anything) or "oidc"`)
		authCM      = fs.String("auth-config-configmap", "", "server: namespace/name of the ConfigMap with the OIDC configuration; required with --auth-strategy=oidc")
		authSecure  = fs.String("auth-cookie-secure", string(auth.CookieSecureAuto), `server: Secure attribute of auth cookies: "auto" (when the request is HTTPS), "true" or "false"; set "true" behind a TLS-terminating proxy`)
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

	if *authStrat != auth.StrategyAnonymous && *authStrat != auth.StrategyOIDC {
		fail(fmt.Errorf("got %q", *authStrat), "--auth-strategy must be anonymous or oidc")
	}
	cookieSecure, err := auth.ParseCookieSecure(*authSecure)
	if err != nil {
		fail(err, "--auth-cookie-secure must be auto, true or false")
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
	p, err := prom.New(*promURL, *rateWindow, prom.Options{Timeout: *promTimeout, UserAgent: fmt.Sprintf("resourcerer/%s (%s)", version, role)})
	if err != nil {
		fail(err, "create Prometheus client")
	}
	cfg.Pricer = &cluster.Pricer{Prom: p}

	// Only the defaults ConfigMap is ever read; caching every ConfigMap in a large cluster costs
	// hundreds of MiB.
	cacheOpts := cache.Options{DefaultTransform: cache.TransformStripManagedFields()}
	if ns := cfg.DefaultsConfigMap.Namespace; ns != "" {
		cacheOpts.ByObject = map[client.Object]cache.ByObject{
			&corev1.ConfigMap{}: {Namespaces: map[string]cache.Config{ns: {}}},
		}
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme.New(),
		Cache:  cacheOpts,
		// Argo Rollouts are read as unstructured; cache them like typed objects instead of hitting the API server.
		Client:                 client.Options{Cache: &client.CacheOptions{Unstructured: true}},
		Metrics:                metricsserver.Options{BindAddress: *metricsAddr},
		HealthProbeBindAddress: *probeAddr,
		LeaderElection:         *leaderElect && (run["controller"] || run["applier"]),
		LeaderElectionID:       role + ".resourcerer.io",
	})
	if err != nil {
		fail(err, "create manager")
	}
	buildInfo.WithLabelValues(version, role).Set(1)
	if run["controller"] {
		if err := recommender.Setup(mgr, cfg, p, *minStep); err != nil {
			fail(err, "set up recommender")
		}
		c := mgr.GetClient()
		metrics.Registry.MustRegister(recommender.FleetCollector{Elected: mgr.Elected(), List: func(ctx context.Context) ([]v1alpha1.Recommendation, error) {
			var l v1alpha1.RecommendationList
			err := c.List(ctx, &l)
			return l.Items, err
		}})
	}
	if run["applier"] {
		if err := applier.Setup(mgr); err != nil {
			fail(err, "set up applier")
		}
	}
	if run["server"] {
		srv := &server.Server{Client: mgr.GetClient(), APIReader: mgr.GetAPIReader(), Cluster: cfg, Prom: p, PromURL: *promURL,
			MinStep: *minStep, Version: version, ReadOnly: *readOnly, Auth: auth.Anonymous{}}
		if !*disableUI {
			srv.UI = web.FS()
		}
		if *authStrat == auth.StrategyOIDC {
			ap, err := oidcProvider(mgr.GetAPIReader(), *authCM, cookieSecure)
			if err != nil {
				fail(err, "set up OIDC authentication")
			}
			if err := mgr.Add(everyReplica(ap.Run)); err != nil {
				fail(err, "add OIDC configuration reload")
			}
			srv.Auth = ap
		}
		if err := mgr.Add(everyReplica(srv.Start(*httpAddr))); err != nil {
			fail(err, "add HTTP server")
		}
	}
	_ = mgr.AddHealthzCheck("healthz", healthz.Ping)
	_ = mgr.AddReadyzCheck("readyz", healthz.Ping)

	logger.Info("starting", "role", role, "version", version, "cluster", cfg.Name, "selection", cfg.Selection,
		"prometheus", *promURL, "readOnly", *readOnly, "auth", *authStrat)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		fail(err, "run manager")
	}
}

// oidcProvider reads secrets from the environment, where the chart injects them from Secrets, and the rest from
// the ConfigMap. Any of them missing stops the server: it never falls back to anonymous.
func oidcProvider(r client.Reader, configMap string, secure auth.CookieSecure) (*auth.OIDC, error) {
	ns, name, ok := strings.Cut(configMap, "/")
	if !ok || ns == "" || name == "" {
		return nil, fmt.Errorf("--auth-config-configmap must be namespace/name, got %q", configMap)
	}
	for _, env := range []string{"RESOURCERER_OIDC_CLIENT_SECRET", "RESOURCERER_SESSION_KEY"} {
		if os.Getenv(env) == "" {
			return nil, fmt.Errorf("%s is not set", env)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return auth.NewOIDC(ctx, auth.OIDCOptions{
		Reader:             r,
		ConfigMap:          types.NamespacedName{Namespace: ns, Name: name},
		ClientSecret:       os.Getenv("RESOURCERER_OIDC_CLIENT_SECRET"),
		SessionKey:         os.Getenv("RESOURCERER_SESSION_KEY"),
		PreviousSessionKey: os.Getenv("RESOURCERER_SESSION_KEY_PREVIOUS"),
		CookieSecure:       secure,
	})
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
