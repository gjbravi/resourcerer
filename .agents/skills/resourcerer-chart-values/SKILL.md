---
name: resourcerer-chart-values
description: Add or change a value in the resourcerer Helm chart and wire it to the binary. Use when a feature needs configuration, a default changes, or a value is renamed or removed.
---

# Chart values

## Steps

1. **`charts/resourcerer/values.yaml`:** add the value with a `# --` comment that says what it does, the options and
   the trade-off, in full sentences. Commented examples are welcome for lists.
2. **`charts/resourcerer/values.schema.json`:** validate it. Objects use `additionalProperties: false`, so typos and
   old keys fail at install instead of being ignored. Use `enum`, `pattern`, `minimum`, `maxItems`, `oneOf` and
   shared `definitions` as the existing entries do.
3. **Wire it:**
   - settings and cost keys go into the defaults ConfigMap (`templates/configmap.yaml`) and are read in
     `internal/cluster` (`Defaults`), so changes apply without restarting the pods;
   - structured configuration goes in as a YAML document under one key (as `reliability.yaml`) and is parsed strictly
     and cached by content (`signals.Parse`);
   - process-level options become flags in the Deployment templates.
4. **`charts/resourcerer/README.md`:** add a row to the values table for anything a user would set.
5. **`hack/resourcerer-values.yaml`:** set it when the kind environment or e2e should exercise it.
6. **Docs:** update `README.md` or `docs/` when the value changes behaviour users see.

## Check

```bash
helm lint charts/resourcerer
helm template rz charts/resourcerer | grep -A3 '<the key>'
helm template rz charts/resourcerer --set <bad value>    # must fail with a schema error
helm lint charts/resourcerer -f hack/resourcerer-values.yaml
```

## Compatibility

Renaming or removing a value breaks installs that set it (the schema rejects unknown keys). That is acceptable in
beta when the PR description says so, so release-please puts it in the changelog. Installs that declare every value
explicitly need the new keys added when they upgrade.
