#!/usr/bin/env bash
# End-to-end checks against a cluster with resourcerer installed (make kind-up kind-deploy).
set -euo pipefail

CTX="${KUBE_CONTEXT:-kind-resourcerer}"
k() { kubectl --context "$CTX" "$@"; }
pass() { printf '\033[32mPASS\033[0m %s\n' "$*"; }
fail() { printf '\033[31mFAIL\033[0m %s\n' "$*"; exit 1; }

wait_for() {
  local desc="$1"; shift
  for _ in $(seq 1 30); do
    if "$@" >/dev/null 2>&1; then pass "$desc"; return; fi
    sleep 2
  done
  fail "$desc"
}

vpa_mode() { k -n demo get vpa "$1" -o jsonpath='{.spec.updatePolicy.updateMode}'; }

wait_for "creates a VPA per workload" k -n demo get vpa deployment-api deployment-worker statefulset-cache daemonset-agent
[[ "$(k -n demo get vpa deployment-api -o jsonpath='{.metadata.labels.resourcerer\.io/managed}')" == "true" ]] \
  && pass "VPA carries the managed label" || fail "VPA carries the managed label"

k -n demo annotate deploy api resourcerer.io/mode=Auto --overwrite >/dev/null
wait_for "mode annotation sets InPlaceOrRecreate" bash -c "[[ \$(kubectl --context $CTX -n demo get vpa deployment-api -o jsonpath='{.spec.updatePolicy.updateMode}') == InPlaceOrRecreate ]]"
k -n demo annotate deploy api resourcerer.io/mode- >/dev/null
wait_for "removing the annotation restores Off" bash -c "[[ \$(kubectl --context $CTX -n demo get vpa deployment-api -o jsonpath='{.spec.updatePolicy.updateMode}') == Off ]]"

k -n demo label deploy worker resourcerer.io/enabled=false --overwrite >/dev/null
wait_for "workload opt-out deletes the VPA" bash -c "! kubectl --context $CTX -n demo get vpa deployment-worker"
k -n demo label deploy worker resourcerer.io/enabled- >/dev/null
wait_for "removing the opt-out recreates the VPA" k -n demo get vpa deployment-worker

k -n demo label namespace demo resourcerer.io/enabled=false --overwrite >/dev/null
wait_for "namespace opt-out deletes all VPAs" bash -c "[[ -z \$(kubectl --context $CTX -n demo get vpa -l resourcerer.io/managed=true -o name) ]]"
k -n demo label namespace demo resourcerer.io/enabled- >/dev/null
wait_for "namespace opt-in again recreates VPAs" k -n demo get vpa deployment-api

k -n demo create deployment e2e-tmp --image=registry.k8s.io/pause:3.10 >/dev/null
wait_for "new workload gets a VPA" k -n demo get vpa deployment-e2e-tmp
k -n demo delete deployment e2e-tmp >/dev/null
wait_for "deleting the workload garbage-collects the VPA" bash -c "! kubectl --context $CTX -n demo get vpa deployment-e2e-tmp"

echo "all e2e checks passed"
