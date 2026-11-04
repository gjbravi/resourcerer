#!/usr/bin/env bash
# Merges the per-kind files controller-gen writes into the single CRD file the resourcerer-crds chart ships.
# controller-gen emits no Helm annotations, so every CRD gets helm.sh/resource-policy: keep: uninstalling the
# chart must not delete every Recommendation and policy.
set -euo pipefail

src="${1:?source directory with the controller-gen output}"
out="${2:?output file}"

shopt -s nullglob
files=("$src"/*.yaml)
if ((${#files[@]} == 0)); then
  echo "merge-crds: no CRD files in $src" >&2
  exit 1
fi

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

mkdir -p "$(dirname "$out")"

first=1
while IFS= read -r file; do
  if ((first)); then
    first=0
  else
    printf -- '---\n'
  fi
  # controller-gen starts each file with a document separator. The file written here joins documents with exactly
  # "\n---\n", which templates/crds.yaml splits on.
  sed '1{/^---$/d;}' "$file" |
    awk '
      /^  annotations:$/ && !added {
        print
        print "    helm.sh/resource-policy: keep"
        added = 1
        next
      }
      { print }
      END { if (!added) exit 1 }
    ' || {
    echo "merge-crds: $file has no CRD metadata annotations block" >&2
    exit 1
  }
done < <(printf '%s\n' "${files[@]}" | sort) >"$tmp"

mv "$tmp" "$out"
chmod 644 "$out"
trap - EXIT
