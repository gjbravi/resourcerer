#!/usr/bin/env bash
# Checks that the docs for contributors and coding agents still match the repository: paths and make targets they
# mention exist, every skill is well formed and listed in AGENTS.md, and Mermaid diagrams render (MERMAID=1).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

docs=(AGENTS.md CONTRIBUTING.md README.md docs/*.md .agents/skills/*/SKILL.md)
errors=0
err() {
  echo "lint-docs: $*" >&2
  errors=$((errors + 1))
}

# Paths in backticks under the repository's top-level directories must exist. Placeholders (..., <x>, *) are skipped.
for f in "${docs[@]}"; do
  grep -oE '`[^` ]+`' "$f" | tr -d '`' | sort -u | while read -r p; do
    p="${p%%:*}"
    p="${p%[.,;)]}"
    [[ "$p" =~ ^(cmd|api|internal|web|charts|hack|test|docs|\.github|\.agents)/ ]] || continue
    [[ "$p" == *"..."* || "$p" == *"<"* || "$p" == *"*"* || "$p" == *"{"* ]] && continue
    [[ -e "$p" ]] || echo "$f: $p"
  done
done > /tmp/lint-docs-paths.$$ || true
while read -r line; do err "missing path: $line"; done < /tmp/lint-docs-paths.$$
rm -f /tmp/lint-docs-paths.$$

# make targets in commands (code blocks and inline code) must exist.
targets=$(grep -oE '^[a-zA-Z0-9_-]+:' Makefile | tr -d ':' | sort -u)
for f in "${docs[@]}"; do
  for t in $(grep -oE '(^|`)make( [a-z][a-z0-9-]*)+' "$f" | sed -E 's/^`?make //' | tr ' ' '\n' | sort -u); do
    grep -qx "$t" <<<"$targets" || err "$f: unknown make target '$t'"
  done
done

# Skills: name matches the directory, a description, and a row in AGENTS.md; and the other way round.
for s in .agents/skills/*/SKILL.md; do
  dir=$(basename "$(dirname "$s")")
  name=$(sed -n 's/^name: *//p' "$s" | head -1)
  desc=$(sed -n 's/^description: *//p' "$s" | head -1)
  [[ "$name" == "$dir" ]] || err "$s: name '$name' does not match its directory"
  [[ -n "$desc" && ${#desc} -le 1024 ]] || err "$s: description missing or longer than 1024 characters"
  grep -q "| \`$dir\` |" AGENTS.md || err "AGENTS.md: skill $dir is not listed in the skills table"
done
for listed in $(grep -oE '^\| `resourcerer-[a-z-]+` \|' AGENTS.md | grep -oE 'resourcerer-[a-z-]+'); do
  [[ -f ".agents/skills/$listed/SKILL.md" ]] || err "AGENTS.md lists skill $listed, which does not exist"
done
[[ "$(readlink .claude/skills)" == "../.agents/skills" ]] || err ".claude/skills must link to ../.agents/skills"

# Mermaid diagrams render. Needs Node and Chrome; CI sets MERMAID=1.
if [[ "${MERMAID:-0}" == 1 ]]; then
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT
  echo "{\"executablePath\":\"${CHROME:-/usr/bin/google-chrome}\",\"args\":[\"--no-sandbox\"]}" > "$tmp/puppeteer.json"
  n=0
  for f in "${docs[@]}"; do
    awk -v dir="$tmp" -v base="$(echo "$f" | tr '/.' '__')" '
      /^```mermaid/ { inside = 1; n++; file = dir "/" base "-" n ".mmd"; next }
      /^```/ && inside { inside = 0; next }
      inside { print > file }
    ' "$f"
  done
  for d in "$tmp"/*.mmd; do
    [[ -e "$d" ]] || continue
    n=$((n + 1))
    PUPPETEER_SKIP_DOWNLOAD=1 npx -y @mermaid-js/mermaid-cli@11.17.1 -q -p "$tmp/puppeteer.json" -i "$d" -o "${d%.mmd}.svg" \
      >/dev/null 2>"${d%.mmd}.err" || err "Mermaid diagram does not render: $(basename "$d"): $(head -3 "${d%.mmd}.err")"
  done
  echo "lint-docs: $n Mermaid diagrams checked"
fi

if ((errors > 0)); then
  echo "lint-docs: $errors problem(s)" >&2
  exit 1
fi
echo "lint-docs: ok"
