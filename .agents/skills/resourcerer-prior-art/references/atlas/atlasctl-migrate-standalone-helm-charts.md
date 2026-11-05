---
updatedAt: 2026-09-24T14:11:46.000Z
---

# Migrate standalone Helm charts with atlasctl

Consolidate individually-installed Atlas Helm releases into the unified umbrella chart using atlasctl cluster migrate.

If your cluster was previously connected using individual Atlas Helm releases (one chart per component), `atlasctl cluster migrate` consolidates them into the unified umbrella chart. This preserves your cluster's Atlas ID, carries forward any custom Helm values you supplied, and installs only the components that were already running.

Use this imperative flow if you installed the standalone charts directly with `helm` or the Atlas onboarding script. If ArgoCD, Flux, or Terraform manages the releases, the next sync or apply would undo manual changes. Use the phased [Migrate GitOps and IaC installations with atlasctl](atlasctl-migrate-gitops.md) flow instead. (Clusters using the Atlas Terraform cluster modules migrate through the module instead: see the [module umbrella migration guide](https://github.com/atlas/terraform-provider-atlas/blob/master/docs/umbrella-migration/README.md).)

The command also handles clusters where the Atlas agent was installed via raw `kubectl apply` rather than Helm.

The `atlas-agent` namespace and any non-chart-owned resources (including the `atlas-agent-metadata` ConfigMap that holds the cluster ID) are not deleted.

<Callout icon="⚠️" theme="warn">
  ### Brief downtime during migration

  The migration uninstalls all standalone Helm releases before installing the umbrella chart. Atlas components are absent from the cluster during this window. Node Autoscaling and Workload Autoscaling are paused until the umbrella chart install completes.
</Callout>

<Callout icon="📘" theme="info">
  ### Not for Atlas Operator clusters

  This flow does not apply to clusters managed via the Atlas Operator. The Operator installs and manages its own components. A migration path off the Operator is coming in a future release.
</Callout>

## Run the migration

1. Switch your kubeconfig to the cluster you want to migrate.

2. Run:

   ```shell
   atlasctl cluster migrate
   ```

   `atlasctl` reads the currently installed standalone releases, captures their values, uninstalls them, and then installs the umbrella chart with the reconstructed configuration.

3. Review the confirmation prompt and proceed. Use `--yes` to skip the prompt in scripts or CI.

4. On success, `atlasctl` confirms that all components are running under the unified chart.

For the full list of flags, including version pinning and recovery options, see the [atlasctl command reference](atlasctl-reference.md).

## If the umbrella install fails

When the umbrella chart installation fails (for example, because an immutable field on an existing Deployment cannot be patched), `atlasctl` rolls back what it can and writes the captured standalone release values to a local file:

```
! Preserved standalone release values written to atlas-migrate-values-<timestamp>.yaml
  Pass this file to `atlasctl cluster migrate --values-file <path>` or
  `helm upgrade atlas atlas/atlas -f <path>` to reapply the custom configuration.
```

A debug snapshot ZIP is also saved to help diagnose the failure.

Fix the underlying issue (for example, delete the conflicting Deployment and let Helm recreate it), then pass the saved values file to retry. In `--values-file` mode, `atlasctl` skips standalone detection and uninstall entirely, running only the umbrella chart installation:

```shell
atlasctl cluster migrate --values-file atlas-migrate-values-<timestamp>.yaml
```

Alternatively, apply the values directly with Helm:

```shell
helm upgrade atlas atlas/atlas -f atlas-migrate-values-<timestamp>.yaml
```

The dump file contains a top-level `releases:` map with per-release values and the cluster ID, so both approaches restore the original configuration.

## See also

<Cards columns="2">
  <Card title="Migrate GitOps and IaC installations with atlasctl" href="https://docs.atlas.example/docs/atlasctl-migrate-gitops" icon="fa-code-branch">
    The phased migration path for ArgoCD, Flux, and Terraform-managed releases.
  </Card>

  <Card title="atlasctl command reference" href="https://docs.atlas.example/docs/atlasctl-reference" icon="fa-book">
    Every atlasctl command and flag, in one place.
  </Card>
</Cards>