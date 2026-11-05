---
updatedAt: 2026-09-22T11:09:14.000Z
---

# Disconnect your cluster

Learn how to disconnect your cluster from Atlas and remove all related data.

# Remove Atlas resources from your cluster

Choose the disconnect method that matches how you connected the cluster.

## Disconnect with atlasctl (recommended)

If you connected the cluster with `atlasctl` or from the Atlas console, the fastest way to disconnect is:

```shell
atlasctl cluster disconnect
```

`atlasctl` uninstalls the Atlas Helm releases, removes the IAM resources created during the connection, and deletes the cluster record from Atlas.

For the full walkthrough, including partial disconnect options and troubleshooting, see [Disconnect a cluster](connect-with-atlasctl.md#disconnect-a-cluster).

> 🚧 Not for Terraform-managed clusters
>
> Do **not** use `atlasctl cluster disconnect` for clusters connected with the [Terraform provider](terraform-provider.md). Use `terraform destroy` instead (see below).

## Migrate from standalone Helm charts

If your cluster was connected using individual Atlas Helm releases (one chart per component), use `atlasctl cluster migrate` to consolidate them into the unified umbrella chart without disconnecting. This preserves the cluster ID, carries forward your custom Helm values, and installs only the components that were already running.

```shell
atlasctl cluster migrate
```

> ⚠️ Brief downtime during migration
>
> Standalone Helm releases are uninstalled before the umbrella chart is installed. Atlas components are absent from the cluster during this window.

If the umbrella install fails, `atlasctl` preserves the captured standalone release values to a local file so you can retry without repeating the uninstall step:

```shell
atlasctl cluster migrate --values-file atlas-migrate-values-<timestamp>.yaml
```

For the full walkthrough, including all flags, the `--values-dump-path` option, and failure recovery, see [Migrate standalone Helm charts with atlasctl](atlasctl-migrate-standalone-helm-charts.md).

## Terraform-managed clusters

Run the [terraform destroy](https://developer.hashicorp.com/terraform/cli/commands/destroy) command to remove the resources. Terraform will destroy all remote objects managed by a particular configuration.

## Clusters managed in the console (legacy script)

### Read-only clusters

To disconnect your cluster from Atlas:

#### Step 1

Click the *Disconnect cluster* button on the list of clusters and follow the on-screen instructions.

![](https://files.readme.io/71805a3-image.png)

Alternatively, run the following command in your terminal to access the cluster:

```shell
kubectl delete namespace atlas-agent
```

#### Step 2

Additionally, delete the following Kubernetes objects related to the Atlas agent:

```shell
kubectl delete clusterrole.rbac.authorization.k8s.io/atlas-agent
kubectl delete clusterrolebinding.rbac.authorization.k8s.io/atlas-agent
```

Once the cluster is disconnected, its status will update to `Disconnected`.

#### Step 3

Remove the cluster from the console by pressing the *Remove cluster* button.

<Image align="center" className="border" border={true} src="https://files.readme.io/0a4be4a-image.png" />

> 📘 Note
>
> The cluster will continue to run as normal, as deleting the cluster in Atlas only removes it from the console.

### Managed service clusters

#### Step 1

In the console, choose the *Disconnect cluster* option.

<Image align="center" className="border" border={true} src="https://files.readme.io/d17eb63-image.png" />

#### Step 2

Enter the cluster name in the form and check the box to delete all nodes Atlas created. Due to reduced cluster capacity, this might cause downtime.

<Image align="center" className="border" width="50% " border={true} src="https://files.readme.io/5e2f315-image.png" />

When the cluster disconnects from Atlas, its status label changes from `Disconnecting` to `Disconnected`.

<Image align="center" className="border" border={true} src="https://files.readme.io/7a6dcf8-image.png" />

#### Step 3

Press the *Remove cluster* button to get a script that cleans up cloud resources like Roles, Bindings, and ServiceAccounts. Run the provided script in your cloud shell or terminal.

<Image align="center" className="border" border={true} src="https://files.readme.io/0a4be4a-image.png" />

> 📘 Note
>
> The cleanup script removes cloud-level resources but does not remove in-cluster Kubernetes objects. See [Step 4 — Remove in-cluster resources](#step-4--remove-in-cluster-resources) below.

## How to remove Atlas credentials and other resources

Follow these steps to ensure that you remove Atlas from your cluster altogether:

### Step 1

In the Atlas console, go to the Autoscaler page and switch the toggle to *OFF* to disable all policies.

<Image align="center" className="border" border={true} src="https://files.readme.io/e205640-image.png" />

### Step 2

Go back to the cluster list and disconnect the cluster by clicking the *Disconnect cluster* button and following the instructions on the screen.

<Image align="center" className="border" border={true} src="https://files.readme.io/d17eb63-image.png" />

Once disconnecting has finished and the cluster status has changed to *Disconnected*, move on to the next step.

### Step 3

Click the *Remove cluster* button and execute the provided offboarding script in your cloud shell or terminal to remove Atlas-created resources from your cloud account.

<Image align="center" className="border" border={true} src="https://files.readme.io/0a4be4a-image.png" />

### Step 4 — Remove in-cluster resources

The cleanup script from Step 3 removes cloud-level resources (IAM roles, service accounts) but does not remove Kubernetes objects that Atlas installed in your cluster. Run the following commands to remove them.

Delete Atlas namespaces and all resources inside them:

```bash
kubectl delete ns atlas-agent atlas-omni liqo-storage --ignore-not-found
```

Wait for the namespaces to fully terminate before proceeding. Controllers running inside these namespaces will recreate cluster-scoped resources if you delete them too early.

```bash
kubectl wait --for=delete ns/atlas-agent ns/atlas-omni ns/liqo-storage --timeout=300s 2>/dev/null; echo "Done"
```

> ⚠️ Warning
>
> If a namespace becomes stuck in the `Terminating` state, see [Removing finalizers from a namespace stuck in the terminating state](#removing-finalizers-from-a-namespace-stuck-in-the-terminating-state) below.

Delete webhook configurations:

```bash
kubectl delete mutatingwebhookconfigurations atlas-mutate-endpointslice atlas-pod-mutator atlas-pod-pinner atlas-workload-autoscaler liqo-webhook --ignore-not-found
kubectl delete validatingwebhookconfigurations atlas-validate-migration vpodmutation.atlas.example liqo-webhook --ignore-not-found
```

Delete ClusterRoleBindings and ClusterRoles:

```bash
kubectl get clusterrolebindings -o name | grep -iE "atlas|liqo" | xargs kubectl delete --ignore-not-found
kubectl get clusterroles -o name | grep -iE "atlas|liqo" | xargs kubectl delete --ignore-not-found
```

Delete CRDs. Remove their finalizers first, otherwise the deletion can hang indefinitely:

```bash
kubectl get crds -o name | grep -E '\.atlas\.example|\.liqo\.io' | xargs -I{} kubectl patch {} -p '{"metadata":{"finalizers":[]}}' --type=merge
kubectl get crds -o name | grep -E '\.atlas\.example|\.liqo\.io' | xargs kubectl delete --ignore-not-found
```

## Troubleshooting

### Removing finalizers from a namespace stuck in the `terminating` state

When disconnecting your cluster or re-onboarding to Atlas, you may encounter a situation where a namespace becomes stuck in the `Terminating` state. If this occurs, follow these steps to forcefully delete the namespace:

1. Verify the namespace status:
   ```shell
   kubectl get ns <namespace-name>
   ```

2. Remove the finalizers:\
   a. Start the proxy in one terminal:

   ```shell
   kubectl proxy
   ```

   b. In another terminal, execute:

   ```shell
   kubectl get namespace atlas-agent -o json | jq '.spec.finalizers = []' > tmp.json && curl -k -H "Content-Type: application/json" -X PUT --data-binary @tmp.json http://127.0.0.1:8001/api/v1/namespaces/atlas-agent/finalize
   ```

3. Confirm namespace deletion:
   ```shell
   kubectl get ns <namespace-name>
   ```

> 🚧 Notice
>
> This process forcefully removes finalizers, allowing namespace deletion. Be aware that this method may leave orphaned resources.

By following these steps, you can successfully disconnect your cluster from Atlas and remove all associated resources.