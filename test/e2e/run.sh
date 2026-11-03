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
  for r in payments/deployment-checkout demo/deployment-api pricing/deployment-on-demand-app pricing/deployment-spot-app; do
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
MPORT="${E2E_METRICS_PORT:-18091}"
PF=""
MPF=""
trap 'kill $PF $MPF 2>/dev/null || true' EXIT
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

# The applier's counters are the only place an apply, a hold or a slot shows up as a number.
if curl -s -o /dev/null --max-time 2 "http://localhost:$MPORT/metrics"; then
  fail "port $MPORT already answers: set E2E_METRICS_PORT to a free port"
fi
k -n resourcerer port-forward svc/resourcerer-applier-metrics "$MPORT:8081" >/dev/null 2>&1 &
MPF=$!
metric() { curl -fs "http://localhost:$MPORT/metrics" | grep -qE "$1"; }
wait_for "the applier's metrics endpoint answers" 60 curl -fs "http://localhost:$MPORT/metrics"

# Start from the manifests, also when an earlier run stopped halfway.
k apply --server-side --force-conflicts -f "$(dirname "$0")/../../hack/demo.yaml" >/dev/null
k delete cap e2e-closed-window e2e-slots e2e-audit e2e-bad-zone --ignore-not-found >/dev/null 2>&1 || true
k label ns demo resourcerer.io/apply-policy- >/dev/null 2>&1 || true
k -n demo annotate deploy api resourcerer.io/auto-deploy- resourcerer.io/cpu.optimization-goal- >/dev/null 2>&1 || true
k -n demo annotate deploy worker resourcerer.io/memory.optimization-goal- >/dev/null 2>&1 || true
k -n payments annotate deploy checkout resourcerer.io/learning-period- >/dev/null 2>&1 || true

rec_field() { k -n "$1" get rec "$2" -o jsonpath="$3"; }
api_cpu() { k -n demo get deploy api -o jsonpath='{.spec.template.spec.containers[?(@.name=="api")].resources.requests.cpu}'; }

wait_for "creates a Recommendation per workload" 60 k -n demo get rec deployment-api deployment-worker statefulset-cache daemonset-log-agent deployment-web deployment-indexer
wait_for "recommendations leave Pending once history exists" 360 bash -c \
  "[[ \$(kubectl --context $CTX -n demo get rec deployment-api -o jsonpath='{.status.phase}') =~ ^(Learning|Ready)$ ]]"
wait_for "istio-proxy is not optimized" 60 bash -c \
  "[[ \$(kubectl --context $CTX -n demo get rec deployment-api -o jsonpath='{.status.containers[?(@.name==\"istio-proxy\")].skipped}') == true ]]"
wait_for "HPA target is recommended for web" 360 bash -c \
  "[[ -n \$(kubectl --context $CTX -n demo get rec deployment-web -o jsonpath='{.status.hpa.recommended}') ]]"

# hack/fake-karpenter.sh prices one worker on-demand (m7g.xlarge, factor 0.84) and the rest spot (t3.xlarge, 0.26).
wait_for "costs are calibrated to Karpenter prices" 120 bash -c \
  "curl -fs $API/info | jq -e '.cost.source == \"karpenter\" and .cost.state == \"ok\" and .cost.allocation == \"node\"'"
wait_for "node allocation finds on-demand-app on the on-demand node" 180 bash -c \
  "[[ \$(kubectl --context $CTX -n pricing get rec deployment-on-demand-app -o json | jq -r '.status.pricing.capacity[\"on-demand\"] // 0') -ge 90 ]]"
wait_for "node allocation finds spot-app on spot nodes" 180 bash -c \
  "[[ \$(kubectl --context $CTX -n pricing get rec deployment-spot-app -o json | jq -r '.status.pricing.capacity.spot // 0') -ge 90 ]]"
od=$(rec_field pricing deployment-on-demand-app '{.status.impact.current.monthlyCost}')
spot=$(rec_field pricing deployment-spot-app '{.status.impact.current.monthlyCost}')
jq -en --argjson od "$od" --argjson spot "$spot" '$od > 2 * $spot' >/dev/null \
  && pass "the same requests cost more on on-demand (\$$od) than on spot (\$$spot)" \
  || fail "the same requests cost more on on-demand (\$$od) than on spot (\$$spot)"

curl -fs "$API/workloads/pricing/Deployment/on-demand-app/placement" \
  | jq -e '.available and .capacity["on-demand"].requests > 0.9 and (.types[0].name == "m7g.xlarge") and (.timeline | length) > 0' >/dev/null \
  && pass "placement shows where on-demand-app ran" || fail "placement shows where on-demand-app ran"
wait_for "the recommendation records evidence from its history" 120 bash -c \
  "kubectl --context $CTX -n demo get rec deployment-api -o json | jq -e '.status.evidence.historySeconds > 0 and (.status.evidence.used.cpu | tonumber) > 0'"

# hack/resourcerer-values.yaml declares cpu-hot and memory-over-request.
curl -fs "$API/info" | jq -e '[.reliability.signals[] | select(.error == null)] | length == 2' >/dev/null \
  && pass "declared reliability signals are valid" || fail "declared reliability signals are valid"
wait_for "declared signals are evaluated per workload" 120 bash -c \
  "kubectl --context $CTX -n demo get rec deployment-api -o json | jq -e '[.status.reliability.signals[]?.id] | index(\"memory-over-request\")'"
curl -fs "$API/workloads/demo/Deployment/api/reliability?window=1h" \
  | jq -e '[.checks[] | select(.id == "signal" and .signal == "memory-over-request" and .container != null)] | length > 0' >/dev/null \
  && pass "the Reliability tab checks declared signals per container" || fail "the Reliability tab checks declared signals per container"

code=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$API/workloads/demo/Deployment/api/apply" -d '{"hash":"stale"}')
[[ "$code" == 409 ]] && pass "apply with a stale hash is refused" || fail "apply with a stale hash is refused (got $code)"

before=$(api_cpu)
hash=$(rec_field demo deployment-api '{.status.hash}')
want=$(rec_field demo deployment-api '{.status.containers[?(@.name=="api")].recommended.requests.cpu}')
curl -fs -X POST "$API/workloads/demo/Deployment/api/apply" -d "{\"hash\":\"$hash\"}" >/dev/null
wait_for "Apply patches the workload with the recommendation ($before → $want)" 60 bash -c "[[ \$(kubectl --context $CTX -n demo get deploy api -o jsonpath='{.spec.template.spec.containers[?(@.name==\"api\")].resources.requests.cpu}') == '$want' ]]"
# The applier patches the workload first and records history right after.
wait_for "apply is recorded in history" 30 bash -c \
  "[[ \$(kubectl --context $CTX -n demo get rec deployment-api -o jsonpath='{.status.history[0].action}') == Apply ]]"
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

# Argo Rollouts: one with its own template, one taking it from a Deployment via spec.workloadRef.
wait_for "creates Recommendations for Rollouts" 60 k -n demo get rec rollout-storefront rollout-billing
wait_for "a workloadRef Deployment gets no Recommendation of its own" 30 bash -c "! kubectl --context $CTX -n demo get rec deployment-billing-template"
[[ "$(rec_field demo rollout-billing '{.status.templateRef.name}')" == "billing-template" ]] \
  && pass "workloadRef is recorded as the template to patch" || fail "workloadRef is recorded as the template to patch"
apply_and_check() { # <recommendation> <jsonpath to the patched cpu request> <resource to read it from>
  local rec="$1" path="$2" obj="$3" hash want
  wait_for "$rec leaves Pending" 360 bash -c \
    "[[ \$(kubectl --context $CTX -n demo get rec $rec -o jsonpath='{.status.phase}') =~ ^(Learning|Ready)$ ]]"
  hash=$(rec_field demo "$rec" '{.status.hash}')
  want=$(rec_field demo "$rec" '{.status.containers[0].recommended.requests.cpu}')
  curl -fs -X POST "$API/workloads/demo/Rollout/${rec#rollout-}/apply" -d "{\"hash\":\"$hash\"}" >/dev/null
  wait_for "Apply on $rec patches $obj ($want)" 60 bash -c \
    "[[ \$(kubectl --context $CTX -n demo get $obj -o jsonpath='$path') == '$want' ]]"
}
apply_and_check rollout-storefront '{.spec.template.spec.containers[0].resources.requests.cpu}' rollout/storefront
apply_and_check rollout-billing '{.spec.template.spec.containers[0].resources.requests.cpu}' deploy/billing-template

k -n demo label deploy worker resourcerer.io/enabled=false --overwrite >/dev/null
wait_for "opt-out deletes the Recommendation" 60 bash -c "! kubectl --context $CTX -n demo get rec deployment-worker"
k -n demo label deploy worker resourcerer.io/enabled- >/dev/null
wait_for "removing the opt-out recreates it" 60 k -n demo get rec deployment-worker

# Apply policies: when an apply may start. hack/resourcerer-values.yaml ships one through the chart, which
# selects namespaces labelled resourcerer.io/apply-policy=strict.
policy() { k apply -f - >/dev/null; }
decision() { # <namespace> <recommendation> <jq filter over the Recommendation>
  k -n "$1" get rec "$2" -o json | jq -e "$3" >/dev/null
}
policy_status() { # <policy> <jq filter over .status>
  k get cap "$1" -o json | jq -e ".status | $2" >/dev/null
}
wait_for "the chart renders its ClusterApplyPolicy" 120 policy_status strict-namespace-window '.workloads == 0'
k label ns demo resourcerer.io/apply-policy=strict >/dev/null
wait_for "a policy selects the workloads of a labelled namespace" 120 policy_status strict-namespace-window '.workloads > 0'
k label ns demo resourcerer.io/apply-policy- >/dev/null

# A window open one minute a week: closed for every run of this script, to the minute.
window_policy() { # <name> <app label>
  policy <<EOF
apiVersion: resourcerer.io/v1alpha1
kind: ClusterApplyPolicy
metadata:
  name: $1
spec:
  namespaceSelector:
    matchLabels: { kubernetes.io/metadata.name: demo }
  workloadSelector:
    matchLabels: { app: $2 }
  appliesTo: [Auto, Manual]
  windows:
    - days: [Mon]
      from: "00:00"
      to: "00:01"
EOF
}

# A click during a closed window is refused and named, and nothing changes.
window_policy e2e-closed-window api
wait_for "a closed window holds both kinds of apply, and says when it opens" 120 decision demo deployment-api \
  '.status.applyPolicy.auto.state == "Blocked" and .status.applyPolicy.auto.blockedBy == "windows"
   and .status.applyPolicy.auto.policy == "ClusterApplyPolicy/e2e-closed-window"
   and .status.applyPolicy.auto.nextAllowed != null and .status.applyPolicy.manual.state == "Blocked"'
held=$(curl -s -o /tmp/rz-held.json -w '%{http_code}' -X POST "$API/workloads/demo/Deployment/api/apply" \
  -d "{\"hash\":\"$(rec_field demo deployment-api '{.status.hash}')\"}")
[[ "$held" == 409 ]] && pass "a held apply is refused with 409" || fail "a held apply is refused with 409 (got $held)"
jq -e '.error | test("e2e-closed-window")' /tmp/rz-held.json >/dev/null \
  && pass "the refusal names the policy" || fail "the refusal names the policy"
k delete cap e2e-closed-window >/dev/null

# An automatic apply that is due while the window is closed is deferred, not dropped, and starts by itself once
# nothing holds it. worker is the workload here: its Recommendation is pending, and no run of this script applies it.
worker_cpu=$(k -n demo get deploy worker -o jsonpath='{.spec.template.spec.containers[?(@.name=="worker")].resources.requests.cpu}')
window_policy e2e-auto-window worker
# A short learning period, so the recommendation is Ready and the applier wants to apply it in a cluster this young.
k -n demo annotate deploy worker resourcerer.io/learning-period=PT2M --overwrite >/dev/null
k -n demo annotate deploy worker resourcerer.io/auto-deploy=Enabled --overwrite >/dev/null
wait_for "a due automatic apply under the closed window" 300 decision demo deployment-worker \
  '.status.phase == "Ready" and .status.pending == true
   and .status.settings["auto-deploy"].value == "Enabled" and .status.applyPolicy.auto.blockedBy == "windows"'
sleep 5
[[ "$(k -n demo get deploy worker -o jsonpath='{.spec.template.spec.containers[?(@.name=="worker")].resources.requests.cpu}')" == "$worker_cpu" ]] \
  && pass "a held automatic apply leaves the workload alone ($worker_cpu)" \
  || fail "a held automatic apply leaves the workload alone (want $worker_cpu)"
k delete cap e2e-auto-window >/dev/null
wait_for "the deferred apply goes through once nothing holds it" 120 decision demo deployment-worker \
  '.status.applyPolicy.auto.state == "Allowed"'
wait_for "and the applier did it, not a person" 120 decision demo deployment-worker \
  '.status.history[0].by == "auto-deploy"'
k -n demo annotate deploy worker resourcerer.io/auto-deploy=Disabled --overwrite >/dev/null

# A revert is never held: it is how a bad apply is undone, and the window that would hold an apply is exactly
# when it has to stay possible.
api_before=$(api_cpu)
curl -fs -X POST "$API/workloads/demo/Deployment/api/apply" -d "{\"hash\":\"$(rec_field demo deployment-api '{.status.hash}')\"}" >/dev/null
wait_for "an apply goes through while no window holds it" 60 bash -c \
  "[[ \$(kubectl --context $CTX -n demo get deploy api -o jsonpath='{.spec.template.spec.containers[?(@.name==\"api\")].resources.requests.cpu}') != '$api_before' ]]"
window_policy e2e-closed-window api
wait_for "the window holds the workload again" 120 decision demo deployment-api \
  '.status.applyPolicy.manual.state == "Blocked"'
curl -fs -X POST "$API/workloads/demo/Deployment/api/revert" >/dev/null
wait_for "a revert is never held: it restores $api_before during a closed window" 60 bash -c \
  "[[ \$(kubectl --context $CTX -n demo get deploy api -o jsonpath='{.spec.template.spec.containers[?(@.name==\"api\")].resources.requests.cpu}') == '$api_before' ]]"
k delete cap e2e-closed-window >/dev/null

# A rate of zero holds every automatic apply in the namespace, and no manual one: a person is not a queue.
policy <<'EOF'
apiVersion: resourcerer.io/v1alpha1
kind: ClusterApplyPolicy
metadata:
  name: e2e-slots
spec:
  namespaceSelector:
    matchLabels: { kubernetes.io/metadata.name: payments }
  rate: { maxConcurrent: 0 }
EOF
wait_for "a rate of zero holds the automatic apply and leaves the manual one free" 120 decision payments deployment-checkout \
  '.status.applyPolicy.auto.state == "Blocked" and .status.applyPolicy.auto.blockedBy == "rate.maxConcurrent"
   and .status.applyPolicy.manual.state == "Allowed"'
k delete cap e2e-slots >/dev/null

# Audit records what it would block, and blocks nothing.
policy <<'EOF'
apiVersion: resourcerer.io/v1alpha1
kind: ClusterApplyPolicy
metadata:
  name: e2e-audit
spec:
  namespaceSelector:
    matchLabels: { kubernetes.io/metadata.name: pricing }
  mode: Audit
  windows:
    - days: [Mon]
      from: "00:00"
      to: "00:01"
EOF
wait_for "an Audit policy records what it would block without blocking" 120 decision pricing deployment-on-demand-app \
  '.status.applyPolicy.auto.state == "Allowed" and (.status.applyPolicy.auto.wouldBlock | length) == 1
   and (.status.applyPolicy.auto.wouldBlock[0] | test("e2e-audit"))'
k delete cap e2e-audit >/dev/null

# A policy that cannot be read holds instead of opening the door, and its status says why.
policy <<'EOF'
apiVersion: resourcerer.io/v1alpha1
kind: ClusterApplyPolicy
metadata:
  name: e2e-bad-zone
spec:
  namespaceSelector:
    matchLabels: { kubernetes.io/metadata.name: demo }
  windows:
    - days: [Mon]
      from: "09:00"
      to: "17:00"
      timeZone: Not/AZone
EOF
wait_for "a policy with an unknown time zone holds the workloads it selects" 120 decision demo deployment-worker \
  '.status.applyPolicy.auto.state == "Blocked" and .status.applyPolicy.auto.blockedBy == "policy"'
wait_for "the policy reports itself invalid" 120 policy_status e2e-bad-zone \
  '(.conditions[0].type == "Valid" and .conditions[0].status == "False")'
k delete cap e2e-bad-zone >/dev/null
wait_for "with every policy gone, applies are allowed again" 120 decision demo deployment-worker \
  '.status.applyPolicy.auto.state == "Allowed" and .status.applyPolicy.manual.state == "Allowed"'

# The metrics the issue asks for: applies by result, holds by rule, reverts, and slots left per policy.
wait_for "the applier exports the apply counters" 60 metric '^resourcerer_apply_total\{result="Succeeded"\} [1-9]'
wait_for "the applier counts the holds by the rule that held them" 60 metric '^resourcerer_apply_deferred_total\{reason="window"\} [1-9]'
wait_for "the applier counts the reverts" 60 metric '^resourcerer_apply_reverted_total\{reason="manual"\} [1-9]'
wait_for "a policy exports the slots it has left" 60 metric '^resourcerer_applypolicy_slots_available\{policy="ClusterApplyPolicy/e2e-slots"\}'

echo "all e2e checks passed"
