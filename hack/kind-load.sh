#!/usr/bin/env bash
# Loads local images into kind nodes. `kind load docker-image` fails on node
# images that use containerd config v4, so this imports with ctr directly.
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-resourcerer}"
for image in "$@"; do
  for node in $(kind get nodes --name "$CLUSTER_NAME"); do
    docker save "$image" | docker exec -i "$node" ctr --namespace=k8s.io images import --all-platforms - >/dev/null
  done
  echo "loaded $image"
done
