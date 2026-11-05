---
updatedAt: 2026-10-14T10:21:11.000Z
---

# Migrate GitOps and IaC installations with atlasctl

Move ArgoCD-, Flux-, or Terraform-managed Atlas standalone releases to the unified umbrella chart.

Older Atlas installations deployed one Helm chart per component, with each component running as a separate release, upgraded and configured independently. The recommended installation today is the **umbrella chart** (`atlas-helm/atlas`): a single Helm release that bundles every Atlas component.&#x20;

With one release, a single upgrade moves all components to mutually compatible versions. Helm tags control which components the umbrella installs, and one values file passes custom settings down to each sub-component. New clusters connect with it, and existing standalone installations should move to it.

**If your standalone releases are managed by GitOps (ArgoCD, Flux) or Terraform, you can't migrate them imperatively.** The next sync or `terraform apply` would undo any manual Helm change. Instead, `atlasctl cluster migrate` provides three subcommands that split the migration into phases. Each change is committed to your source of truth and applied by your own tooling:

* `generate` (read-only): inspects the cluster and writes an umbrella chart values file, a manifest for your tool, and a step-by-step `MIGRATION-STEPS.md` tailored to what it found.
* `pre-umbrella-install`: prepares the cluster for adoption. Preserves Workload Autoscaler CRDs and strips standalone Helm ownership without touching running workloads.
* `post-umbrella-install`: once the umbrella is healthy, it cleans up stale ReplicaSets and orphaned pods. On ArgoCD, it also disarms the Workload Autoscaler pre-delete hook, so a later Application delete can't strip finalizers from Recommendation CRs.

**The phase split is deliberate**: you can pause between phases, check the cluster state, and roll back before the final cleanup runs.

<Callout icon="📘" theme="info">
  ### Which migration path applies to you?

  * :x: Cluster installed directly with `helm` or the Atlas onboarding script: use the imperative [Migrate standalone Helm charts with atlasctl](atlasctl-migrate-standalone-helm-charts.md) instead.
  * :x: Cluster managed with the **Atlas Terraform cluster modules** (`atlas-eks-cluster`, `atlas-gke-cluster`, `atlas-aks`): this flow does not apply, because the module manages the standalone releases itself. Follow the module's [umbrella migration guide](https://github.com/atlas/terraform-provider-atlas/blob/master/docs/umbrella-migration/README.md) instead (set `umbrella_enabled=true`; requires module version EKS 15.x, GKE 11.x, or AKS 12.x).
  * :x: Cluster managed via the **Atlas Operator**: this flow does not apply, because the Operator installs and manages its own components. A migration path off the Operator is coming in a future release.
  * :x: Cluster running Karpenter Enterprise Suite with the `atlas-keeper` controller: this flow is not supported. The controller is not migrated.
  * :white_check_mark: ArgoCD, Flux, or plain Terraform `helm_release` resources: you're in the right place.
</Callout>

## Before you begin

* `atlasctl` **0.15.0 or later**, [updated](connect-with-atlasctl.md). Check with:

  ```shell
  atlasctl version
  ```

* `atlasctl` authenticated with your Atlas account:

  ```shell
  atlasctl auth login
  ```

  The CLI asks how to authenticate: a browser window, or an [API key](api-access.md) you paste. It then asks you to select an organization. Select the one where the cluster appears in the [Atlas console](https://console.atlas.example); migrate commands run only against clusters in the selected organization.

* Your kubeconfig context points at the cluster you're migrating. Check with:

  ```shell
  kubectl config current-context
  ```

* Cluster-admin or equivalent RBAC on that cluster. `pre-umbrella-install` deletes Helm release secrets and orphan-deletes Deployments and DaemonSets in the Atlas namespace (default `atlas-agent`).

* Access to commit to the GitOps repository or apply the Terraform code that manages your Atlas releases.

* **Helm 3.17.0 or later**, if you might reinstall a standalone chart with `--take-ownership` (used in the rollback steps below). Check with:

  ```shell
  helm version
  ```

## How the migration works

### Phase 1: Generate

Run `atlasctl cluster migrate generate --format <your tool>`. It only reads from the cluster, so nothing changes and there is nothing to undo.

### Phase 2: Prepare

Run `atlasctl cluster migrate pre-umbrella-install`. It protects Workload Autoscaler CRDs, deletes the Helm release secrets, and orphan-deletes the old Deployments and DaemonSets. **Pods keep running.** To roll back here, reinstall the standalone charts with `--take-ownership`. Use the values from the step 0 backup in `MIGRATION-STEPS.md`.

### Phase 3: Install

Sync or apply the umbrella chart through your tool. Umbrella pods start alongside the old generation, which stays as warm standby, and only the new pods receive traffic. To roll back, uninstall the umbrella; the old generation is still running.

### Phase 4: Clean up

Run `atlasctl cluster migrate post-umbrella-install`. It deletes the old-generation ReplicaSets and pods, and on ArgoCD it also disarms the standalone Workload Autoscaler's pre-delete hook. After this, the old pods are gone, so rolling back means a full restart of Atlas pods.

Because pods are never deleted before the last phase, the migration has no downtime window: Atlas components keep running throughout the process.

## Generate the migration plan

1. Run `generate` with the format matching your tool. The command only reads from the cluster, so it's safe to run at any time and repeat:

   ```shell
   atlasctl cluster migrate generate \
     --output atlas-migration \
     --format <argocd|flux|terraform|generic> \
     --repo-url <https://github.com/org/gitops-repo.git>
   ```

   Only `--format argocd` needs `--repo-url`. Without it, the generated Application contains a `<YOUR-GITOPS-REPO>` placeholder. For [Atlas Anywhere](atlas-anywhere-overview.md) clusters, add `--profile autoscaler-anywhere`.

   `generate` detects standalone Helm releases via their Helm records and ArgoCD-managed components by scanning ArgoCD Applications cluster-wide. For ArgoCD components, it reconstructs their values from each Application's `spec.source.helm` section. If the umbrella chart is already installed, the command stops and tells you there's nothing to migrate.

2. Review the generated files in the output directory:

   * `values.yaml`: umbrella chart values, including your preserved per-component custom values, the mode tag covering your detected components, and Workload Autoscaler CRD protection (`crds.keep=true`) where applicable. The API key is referenced via `apiKeySecretRef`, never embedded.
   * `standalone-patch.yaml`: a patch that protects Workload Autoscaler CRDs and scales the controller down before adoption. Only generated when a Workload Autoscaler installation is detected, and skipped for Flux (Flux setups patch the HelmRelease instead).
   * `argocd-application.yaml`, `flux-helmrelease.yaml`, or `atlas-helm-release.tf`: the umbrella manifest for your tool, with the current umbrella chart version pinned.
   * `MIGRATION-STEPS.md`: a step-by-step checklist for your exact setup, including the step 0 backup, per-tool commands, and rollback instructions.

3. Follow `MIGRATION-STEPS.md`, starting with the step 0 backup of your current values and the `atlas-agent-metadata` ConfigMap. That file is the authoritative checklist for your cluster. This guide covers the same phases next: [Prepare the cluster for adoption](#prepare-the-cluster-for-adoption), [Install the umbrella chart through your tool](#install-the-umbrella-chart-through-your-tool), and [Clean up after the install](#clean-up-after-the-install).

## Prepare the cluster for adoption

1. Preview what the preparation phase will do:

   ```shell
   atlasctl cluster migrate pre-umbrella-install --dry-run
   ```

   Check that the output lists only your standalone Atlas releases: Helm release secrets to delete, Deployments and DaemonSets to orphan-delete, and Workload Autoscaler CRD preservation if you have Workload Autoscaler.

2. Run the preparation:

   ```shell
   atlasctl cluster migrate pre-umbrella-install
   ```

   Type `yes` at the confirmation prompt, or pass `--yes` in automation.

   The command protects Workload Autoscaler Recommendation CRDs by scaling the controller to 0 and pinning `crds.keep=true`. It deletes the standalone releases' Helm secrets, so your GitOps tool no longer fights Helm ownership. It also orphan-deletes the Deployments and DaemonSets that the umbrella will re-create. **Pods are not deleted and keep running.**

If you're migrating a kubectl-applied manifest installation rather than Helm releases, the command automatically uses the server-side-apply strategy instead of deleting Helm secrets.

## Install the umbrella chart through your tool

Commit the generated artifacts and let your own tooling install the umbrella chart. The exact steps are in `MIGRATION-STEPS.md`; in short:

<Tabs>
  <Tab title="ArgoCD">
    1. Commit `values.yaml` to the root of your GitOps repo and push.

    2. Apply `argocd-application.yaml`, then review and sync manually:

       ```shell
       kubectl apply -f argocd-application.yaml
       argocd app diff atlas
       argocd app sync atlas
       ```

    3. Keep auto-sync disabled until the migration is complete. `MIGRATION-STEPS.md` walks through pausing your existing standalone Applications first.
  </Tab>

  <Tab title="Flux">
    1. Suspend the standalone HelmReleases, and patch the Workload Autoscaler HelmRelease for CRD preservation as shown in `MIGRATION-STEPS.md`.

    2. Commit `flux-helmrelease.yaml` (the umbrella values are inlined in `spec.values`) and remove the standalone HelmRelease manifests from Git.

    3. Push, or force reconciliation:

       ```shell
       flux reconcile helmrelease atlas -n atlas-agent
       ```
  </Tab>

  <Tab title="Terraform">
    1. Replace the standalone `helm_release` resources with the single umbrella resource from `atlas-helm-release.tf`.

    2. Remove the standalone releases from state. Use `terraform state rm` only; never `destroy`, or you'll delete the resources being adopted:

       ```shell
       terraform state rm helm_release.<standalone_release>
       ```

    3. Apply. This installs the umbrella with `take_ownership = true`. It requires the `hashicorp/helm` provider 3.1.0 or later. `MIGRATION-STEPS.md` documents an out-of-band fallback for older providers:

       ```shell
       terraform apply
       ```

    The generated code is also valid for OpenTofu, unchanged.
  </Tab>
</Tabs>

Because `pre-umbrella-install` stripped the old Helm ownership, the umbrella release adopts the surviving resources. Its pods start as a **new generation** next to the old one. The umbrella's Services select only the new pods. The old pods remain running as your rollback safety net.

While both generations run, the namespace temporarily carries about twice the usual pod count. Before installing the umbrella, make sure the cluster has spare capacity for that: on a tightly packed cluster, the new pods can stay Pending until `post-umbrella-install` removes the old generation. If new-generation pods stay Pending, check why:

```shell
kubectl describe pod <pod-name> -n atlas-agent
```

Scheduling pressure clears itself once `post-umbrella-install` removes the old generation. On a full cluster, you can also delete a few old-generation pods for that component; the new ones then schedule immediately.

## Clean up after the install

1. Wait until the umbrella pods are Ready:

   ```shell
   kubectl rollout status deployment,daemonset -n atlas-agent
   ```

2. Run the cleanup (preview first with `--dry-run` if you want):

   ```shell
   atlasctl cluster migrate post-umbrella-install
   ```

   This deletes the stale ReplicaSets and orphaned pods of the old generation. On ArgoCD, it also sets `preDeleteHook.enabled=false` on any Application that still runs the standalone workload-autoscaler chart. This stops a later Application delete from stripping finalizers from your Recommendation CRs and deleting the CRD.

3. Finish the tool-specific cleanup from `MIGRATION-STEPS.md`:

   <Tabs>
     <Tab title="ArgoCD">
       Delete each old standalone Application with `--cascade=orphan`. This deletes only the resources the umbrella didn't adopt. Then re-enable auto-sync on the umbrella:

       ```shell
       argocd app delete <standalone-app> --cascade=orphan
       # repeat for every standalone Application generate detected, or list them first: argocd app list
       argocd app set atlas --sync-policy automated --auto-prune --self-heal
       ```

       Older `argocd` CLI versions use the deprecated `--cascade=false` instead.
     </Tab>

     <Tab title="Flux">
       Check that only the umbrella HelmRelease remains; the standalone HelmReleases were already removed from Git in the install step:

       ```shell
       flux get helmrelease -n atlas-agent
       ```
     </Tab>

     <Tab title="Terraform">
       Check that only the umbrella resource remains in state; the standalone `helm_release` resources were already removed in the install step:

       ```shell
       terraform state list | grep helm_release
       ```
     </Tab>
   </Tabs>

## Verify the migration

```shell
kubectl get pods -n atlas-agent
kubectl rollout status deployment,daemonset -n atlas-agent
kubectl get cm atlas-agent-metadata -n atlas-agent -o jsonpath='{.data.CLUSTER_ID}'
```

All pods should be running under the umbrella release, and the cluster ID should match the one from before the migration. The Atlas console shows your components as managed by the unified chart.

## Roll back if something goes wrong

Before `post-umbrella-install` runs, the old pod generation is still alive, so rollback is quick:

1. Delete the umbrella release, or the Application/HelmRelease/`helm_release` that installed it. For plain Helm:

   ```shell
   helm uninstall atlas -n atlas-agent
   ```

   The Workload Autoscaler and its exporter CRDs survive because `generate` pins `crds.keep=true` in the generated values.

2. Re-create the standalone releases from your source of truth (restore the blocks you removed). Pass `--take-ownership` on every one; this requires Helm 3.17.0 or later. The surviving Workload Autoscaler and exporter CRDs still carry the umbrella release's ownership metadata. Their standalone charts need `--take-ownership` to adopt them; otherwise the install fails on a metadata conflict. Passing the flag on the other components too is harmless. `MIGRATION-STEPS.md` has the exact command for each release, in the form:

   ```shell
   helm upgrade --install <release> atlas-helm/<chart> -n atlas-agent -f backup-<release>.yaml --take-ownership
   ```

If you already ran `post-umbrella-install`, the old pods are gone. The same steps still work, but every Atlas pod is recreated, so expect a full restart. The tool-specific rollback checklists (including how to revert the Workload Autoscaler scale-down) are in the **Rollback & Recovery** section of your generated `MIGRATION-STEPS.md`.

## Limitations and notes

* This migration covers standalone releases managed by ArgoCD, Flux, or plain Terraform `helm_release` resources. It does not apply to clusters managed via the Atlas Operator, which installs and manages its components itself.
* Only components owned by the umbrella chart are migrated. Standalone releases the umbrella doesn't manage are left untouched and reported with a warning during the prepare and cleanup phases. This currently covers `atlas-egressd`, OMNI, and Database Optimizer components.
* The Workload Autoscaler and exporter CRDs (`recommendations`, `gpurecommendations`, `custommetricsexporterconfigs`) are pinned to survive an uninstall, on both the standalone and umbrella charts. Other component CRDs are not pinned; for example, the ones for Atlas Live Migration and Pod Mutations. An uninstall deletes those, and the next install recreates them. Any custom resources of that kind are lost during that window.
* `atlas-agent-cpvpa`-style helper workloads from older installs are preserved automatically.
* Clusters using the Atlas Terraform cluster modules (`atlas-eks-cluster`, `atlas-gke-cluster`, `atlas-aks`) follow the module's own flow instead: see <Anchor target="_blank" href="https://github.com/atlas/terraform-provider-atlas/blob/master/docs/umbrella-migration/README.md">Migrating to the umbrella chart with the Terraform modules</Anchor>. On module versions with umbrella support (EKS 15.x, GKE 11.x, AKS 12.x), an `apply` with `umbrella_enabled=true` performs the migration. A future major version will remove standalone chart support, so migrate while on these versions.
* `generate` never mutates the cluster, so you can re-run it freely. For example, run it again with a different `--format` for another tool (OpenTofu and Spacelift use `--format terraform`; OpenShift GitOps and ApplicationSets use `--format argocd`).

## See also

<Cards columns="2">
  <Card title="Migrate standalone Helm charts with atlasctl" href="https://docs.atlas.example/docs/atlasctl-migrate-standalone-helm-charts" icon="fa-arrows-rotate">
    The imperative migration path for direct Helm and onboarding-script installs.
  </Card>

  <Card title="Migrate with the Terraform cluster modules" href="https://github.com/atlas/terraform-provider-atlas/blob/master/docs/umbrella-migration/README.md" icon="fa-cloud">
    The umbrella migration path for clusters managed by the Atlas Terraform cluster modules.
  </Card>

  <Card title="atlasctl command reference" href="https://docs.atlas.example/docs/atlasctl-reference" icon="fa-book">
    Every atlasctl command and flag, including the migrate subcommands.
  </Card>
</Cards>