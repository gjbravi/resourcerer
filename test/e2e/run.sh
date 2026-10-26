#!/usr/bin/env bash
# End-to-end checks against the kind environment (make kind-up kind-deploy).
set -euo pipefail

CTX="${KUBE_CONTEXT:-kind-resourcerer}"
PORT="${E2E_PORT:-18080}"
k() { kubectl --context "$CTX" "$@"; }
pass() { printf '\033[32mPASS\033[0m %s\n' "$*"; }
fail() {
  printf '\033[31mFAIL\033[0m %s\n' "$*"
  echo "--- recommendations"
  k get recs -A -o wide || true
  for r in payments/deployment-checkout demo/deployment-api; do
    echo "--- $r status"
    k -n "${r%%/*}" get rec "${r##*/}" -o json | jq '.status | del(.containers[]?.basis)' || true
  done
  for c in controller applier; do
    echo "--- $c logs"
    k -n resourcerer logs "deploy/resourcerer-$c" --tail=40 || true
  done
  exit 1
}

# wait_for <description> <seconds> <command...>
wait_for() {
  local desc="$1" secs="$2"; shift 2
  for _ in $(seq 1 $((secs / 3))); do
    if "$@" >/dev/null 2>&1; then pass "$desc"; return; fi
    sleep 3
  done
  fail "$desc"
}

API="http://localhost:$PORT/api/v1"
PF=""
trap 'kill $PF 2>/dev/null || true' EXIT
# Right after a rollout the forward can land on a terminating pod and exit; reopen it until the API answers.
k -n resourcerer rollout status deploy/resourcerer-server --timeout=120s >/dev/null
for _ in $(seq 1 20); do
  if [[ -z "$PF" ]] || ! kill -0 "$PF" 2>/dev/null; then
    k -n resourcerer port-forward svc/resourcerer "$PORT:80" >/dev/null 2>&1 &
    PF=$!
  fi
  curl -fs "$API/info" >/dev/null 2>&1 && break
  sleep 3
done
wait_for "server answers" 30 curl -fs "$API/info"

# Start from the manifests, also when an earlier run stopped halfway.
k apply --server-side --force-conflicts -f "$(dirname "$0")/../../hack/demo.yaml" >/dev/null
k -n demo annotate deploy worker resourcerer.io/memory.optimization-goal- >/dev/null 2>&1 || true
k -n payments annotate deploy checkout resourcerer.io/learning-period- >/dev/null 2>&1 || true

rec_field() { k -n "$1" get rec "$2" -o jsonpath="$3"; }
api_cpu() { k -n demo get deploy api -o jsonpath='{.spec.template.spec.containers[?(@.name=="api")].resources.requests.cpu}'; }

wait_for "creates a Recommendation per workload" 60 k -n demo get rec deployment-api deployment-worker statefulset-cache daemonset-log-agent deployment-web deployment-indexer
wait_for "recommendations leave Pending once history exists" 360 bash -c \
  "[[ \$(kubectl --context $CTX -n demo get rec deployment-api -o jsonpath='{.status.phase}') =~ ^(Learning|Ready)$ ]]"
[[ "$(rec_field demo deployment-api '{.status.containers[?(@.name=="istio-proxy")].skipped}')" == "true" ]] \
  && pass "istio-proxy is not optimized" || fail "istio-proxy is not optimized"
[[ -n "$(rec_field demo deployment-web '{.status.hpa.recommended}')" ]] \
  && pass "HPA target is recommended for web" || fail "HPA target is recommended for web"

code=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API/workloads/demo/Deployment/api/apply" -d '{"hash":"stale"}')
[[ "$code" == 409 ]] && pass "apply with a stale hash is refused" || fail "apply with a stale hash is refused (got $code)"

before=$(api_cpu)
hash=$(rec_field demo deployment-api '{.status.hash}')
want=$(rec_field demo deployment-api '{.status.containers[?(@.name=="api")].recommended.requests.cpu}')
curl -fs -X POST "$API/workloads/demo/Deployment/api/apply" -d "{\"hash\":\"$hash\"}" >/dev/null
wait_for "Apply patches the workload with the recommendation ($before → $want)" 60 bash -c "[[ \$(kubectl --context $CTX -n demo get deploy api -o jsonpath='{.spec.template.spec.containers[?(@.name==\"api\")].resources.requests.cpu}') == '$want' ]]"
[[ "$(rec_field demo deployment-api '{.status.history[0].action}')" == "Apply" ]] \
  && pass "apply is recorded in history" || fail "apply is recorded in history"
[[ "$(k -n demo get deploy api -o json --show-managed-fields | jq -r '[.metadata.managedFields[] | select(.manager=="resourcerer")] | length')" -ge 1 ]] \
  && pass "field manager is resourcerer" || fail "field manager is resourcerer"

curl -fs -X POST "$API/workloads/demo/Deployment/api/revert" >/dev/null
wait_for "Revert restores the previous requests ($before)" 60 bash -c "[[ \$(kubectl --context $CTX -n demo get deploy api -o jsonpath='{.spec.template.spec.containers[?(@.name==\"api\")].resources.requests.cpu}') == '$before' ]]"

curl -fs -X PUT "$API/workloads/demo/Deployment/worker/settings" -d '{"key":"memory.optimization-goal","value":"Savings"}' >/dev/null
wait_for "a workload setting is stored as an annotation and resolved" 60 bash -c \
  "[[ \$(kubectl --context $CTX -n demo get rec deployment-worker -o jsonpath='{.status.settings.memory\.optimization-goal.source}') == workload ]]"
code=$(curl -s -o /dev/null -w '%{http_code}' -X PUT "$API/workloads/demo/Deployment/worker/settings" -d '{"key":"memory.optimization-goal","value":"Cheap"}')
[[ "$code" == 400 ]] && pass "invalid setting values are refused" || fail "invalid setting values are refused (got $code)"

# payments has auto-deploy enabled on the namespace; shorten checkout's learning period so it becomes Ready quickly.
k -n payments annotate deploy checkout resourcerer.io/learning-period=PT2M --overwrite >/dev/null
wait_for "auto-deploy applies a Ready recommendation" 300 bash -c \
  "[[ \$(kubectl --context $CTX -n payments get rec deployment-checkout -o jsonpath='{.status.history[0].by}') == auto-deploy ]]"
[[ -z "$(rec_field payments deployment-ledger '{.status.history}')" ]] \
  && pass "workload annotation turns auto-deploy off for ledger" || fail "workload annotation turns auto-deploy off for ledger"

k -n demo label deploy worker resourcerer.io/enabled=false --overwrite >/dev/null
wait_for "opt-out deletes the Recommendation" 60 bash -c "! kubectl --context $CTX -n demo get rec deployment-worker"
k -n demo label deploy worker resourcerer.io/enabled- >/dev/null
wait_for "removing the opt-out recreates it" 60 k -n demo get rec deployment-worker

echo "all e2e checks passed"
