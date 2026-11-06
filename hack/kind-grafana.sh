#!/usr/bin/env bash
# Grafana on the kind cluster, with the example dashboard the docs publish (docs/assets/grafana): run it again after
# make dashboard to load a changed dashboard. Optional: make kind-up does not need it, and e2e does not check it.
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-resourcerer}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CTX="kind-$CLUSTER_NAME"
GRAFANA_CHART_VERSION="${GRAFANA_CHART_VERSION:-10.5.15}"

helm repo add grafana https://grafana.github.io/helm-charts --force-update >/dev/null
helm repo update grafana >/dev/null
helm upgrade --install grafana grafana/grafana --kube-context "$CTX" --namespace monitoring --create-namespace \
  --version "$GRAFANA_CHART_VERSION" -f "$ROOT/hack/grafana-values.yaml" --wait --timeout 5m >/dev/null
kubectl --context "$CTX" -n monitoring create configmap grafana-dashboard-resourcerer \
  --from-file=resourcerer.json="$ROOT/docs/assets/grafana/resourcerer-dashboard.json" --dry-run=client -o yaml |
  kubectl --context "$CTX" label --local -f - grafana_dashboard=1 -o yaml |
  kubectl --context "$CTX" apply -f - >/dev/null
echo "kubectl --context $CTX -n monitoring port-forward svc/grafana 3000:80, then http://localhost:3000/d/resourcerer"
