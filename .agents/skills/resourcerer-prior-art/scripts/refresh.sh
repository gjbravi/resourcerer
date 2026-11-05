#!/usr/bin/env bash
# Re-download both docs sites and rebuild references/ with brand names replaced by codenames.
# This directory (scripts/) is the only place that names the real products; references/ never does.
set -euo pipefail

skill="$(cd "$(dirname "$0")/.." && pwd)"
scripts="$skill/scripts"
raw="$(mktemp -d)"
trap 'rm -rf "$raw"' EXIT
mkdir -p "$raw/a" "$raw/b"

# Atlas
a_host="https://docs.cast.ai"
curl -sfL "$a_host/sitemap.xml" | grep -o '<loc>[^<]*' | sed 's/<loc>//' | grep "^$a_host/docs/" |
  while read -r u; do
    curl -sfL "$u.md" -o "$raw/a/${u#"$a_host"/docs/}.md" || echo "skip $u" >&2
  done
curl -sfL "$a_host/docs/llms.txt" -o "$raw/a.llms.txt"

# Orion
b_host="https://docs.stormforge.io"
curl -sfL "$b_host/sitemap.xml" | grep -o '<loc>[^<]*' | sed 's/<loc>//' | grep '^/docs/' | grep -v '/previous/' |
  while read -r p; do
    name="$(echo "$p" | sed 's#^/docs/##; s#/$##; s#/#__#g')"
    curl -sfL "$b_host${p}index.md" -o "$raw/b/${name:-home}.md" || echo "skip $p" >&2
  done
curl -sfL "$b_host/llms.txt" -o "$raw/b.llms.txt"

rm -rf "$skill/references"
python3 -I -B "$scripts/sanitize.py" "$raw/a" "$raw/b" "$skill/references"
python3 -I -B "$scripts/index.py" "$raw/a.llms.txt" "$raw/b.llms.txt" "$skill/references"

if grep -rliE 'cast ?ai|storm ?forge|cloudbolt' "$skill/references"; then
  echo "brand names left in the files above; extend the substitutions in sanitize.py" >&2
  exit 1
fi
date -u +%Y-%m-%d > "$skill/references/FETCHED"
