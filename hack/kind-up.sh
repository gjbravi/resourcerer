#!/usr/bin/env bash
# Creates a kind cluster with metrics-server, the Vertical Pod Autoscaler and demo workloads.
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-resourcerer}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CTX="kind-$CLUSTER_NAME"
METRICS_SERVER_CHART_VERSION="${METRICS_SERVER_CHART_VERSION:-3.14.0}"
VPA_CHART_VERSION="${VPA_CHART_VERSION:-0.15.0}"

log() { printf '\033[1;34m==>\033[0m %s\n' "$*"; }

if ! kind get clusters 2>/dev/null | grep -qx "$CLUSTER_NAME"; then
  log "creating kind cluster $CLUSTER_NAME"
  kind create cluster --name "$CLUSTER_NAME" --config "$ROOT/hack/kind-config.yaml" --wait 120s
fi

# Behind a TLS-intercepting proxy, nodes need its root CA to pull images.
if [[ -n "${EXTRA_CA_CERT:-}" ]]; then
  log "adding $EXTRA_CA_CERT to node trust stores"
  for node in $(kind get nodes --name "$CLUSTER_NAME"); do
    docker cp "$EXTRA_CA_CERT" "$node:/usr/local/share/ca-certificates/extra-ca.crt"
    docker exec "$node" sh -c 'update-ca-certificates >/dev/null && systemctl restart containerd'
  done
fi

log "helm repositories"
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/ --force-update >/dev/null
helm repo add autoscaler https://kubernetes.github.io/autoscaler --force-update >/dev/null
helm repo update metrics-server autoscaler >/dev/null

log "metrics-server"
helm upgrade --install metrics-server metrics-server/metrics-server \
  --kube-context "$CTX" --namespace kube-system --version "$METRICS_SERVER_CHART_VERSION" \
  --set 'args={--kubelet-insecure-tls}' --wait

log "vertical-pod-autoscaler"
helm upgrade --install vpa autoscaler/vertical-pod-autoscaler \
  --kube-context "$CTX" --namespace vpa-system --create-namespace --version "$VPA_CHART_VERSION" \
  -f "$ROOT/hack/vpa-values.yaml" --wait --timeout 5m

log "demo workloads"
kubectl --context "$CTX" apply -f "$ROOT/hack/demo.yaml"

log "ready: run 'make kind-deploy' to install resourcerer"
