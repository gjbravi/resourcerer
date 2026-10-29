#!/usr/bin/env bash
# Stands in for Karpenter's metrics in kind, so cost calibration and node allocation are tested end to end.
# The first worker is labeled on-demand (m7g.xlarge), every other node spot (t3.xlarge); prices are fixed.
set -euo pipefail

CTX="${1:?kube context}"
CPU=4
MEMORY=$((16 * 1024 * 1024 * 1024))

nodes=$(kubectl --context "$CTX" get nodes -o name | cut -d/ -f2)
first_worker=$(kubectl --context "$CTX" get nodes -l '!node-role.kubernetes.io/control-plane' -o name | cut -d/ -f2 | sort | head -1)

metrics="# TYPE karpenter_nodes_allocatable gauge
"
while read -r node; do
  [[ -z "$node" ]] && continue
  if [[ "$node" == "$first_worker" ]]; then cap=on-demand type=m7g.xlarge; else cap=spot type=t3.xlarge; fi
  kubectl --context "$CTX" label node "$node" "karpenter.sh/capacity-type=$cap" --overwrite >/dev/null
  for r in cpu memory; do
    v=$CPU; [[ $r == memory ]] && v=$MEMORY
    metrics+="karpenter_nodes_allocatable{node_name=\"$node\",instance_type=\"$type\",zone=\"kind-a\",capacity_type=\"$cap\",resource_type=\"$r\"} $v
"
  done
done <<<"$nodes"
metrics+='# TYPE karpenter_cloudprovider_instance_type_offering_price_estimate gauge
karpenter_cloudprovider_instance_type_offering_price_estimate{instance_type="t3.xlarge",zone="kind-a",capacity_type="spot"} 0.0503
karpenter_cloudprovider_instance_type_offering_price_estimate{instance_type="t3.xlarge",zone="kind-a",capacity_type="on-demand"} 0.1664
karpenter_cloudprovider_instance_type_offering_price_estimate{instance_type="m7g.xlarge",zone="kind-a",capacity_type="spot"} 0.0612
karpenter_cloudprovider_instance_type_offering_price_estimate{instance_type="m7g.xlarge",zone="kind-a",capacity_type="on-demand"} 0.1632
'

kubectl --context "$CTX" create namespace karpenter --dry-run=client -o yaml | kubectl --context "$CTX" apply -f - >/dev/null
kubectl --context "$CTX" -n karpenter create configmap fake-karpenter-metrics --from-literal=metrics.txt="$metrics" \
  --dry-run=client -o yaml | kubectl --context "$CTX" apply -f - >/dev/null
kubectl --context "$CTX" apply -f - <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fake-karpenter
  namespace: karpenter
spec:
  replicas: 1
  selector:
    matchLabels: { app: fake-karpenter }
  template:
    metadata:
      labels: { app: fake-karpenter }
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        # .txt so busybox serves text/plain, which Prometheus 3 requires.
        prometheus.io/path: /metrics.txt
    spec:
      containers:
        - name: httpd
          image: busybox:1.37
          command: ["httpd", "-f", "-p", "8080", "-h", "/www"]
          ports: [{ containerPort: 8080 }]
          resources: { requests: { cpu: 5m, memory: 8Mi } }
          volumeMounts: [{ name: metrics, mountPath: /www }]
      volumes:
        - name: metrics
          configMap: { name: fake-karpenter-metrics }
YAML
kubectl --context "$CTX" -n karpenter rollout restart deploy/fake-karpenter >/dev/null
kubectl --context "$CTX" -n karpenter rollout status deploy/fake-karpenter --timeout=120s >/dev/null
echo "fake Karpenter: $first_worker on-demand, other nodes spot"
