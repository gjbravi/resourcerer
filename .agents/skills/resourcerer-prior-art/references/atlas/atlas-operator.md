---
updatedAt: 2026-09-22T11:40:29.000Z
---

# Atlas Operator

Automated lifecycle management for Atlas components

The Atlas Operator (`atlasware-operator`) automates the installation, configuration, and updating of Atlas components in your Kubernetes cluster. The Operator reduces manual effort by managing component lifecycles through a single, self-updating control plane component.

# What the Operator manages

The Atlas Operator manages the following components:

**Phase 1 components:**

* `atlas-agent`: Cluster monitoring and data collection
* `spot-handler`: Spot Instance interruption monitoring and reporting

**Phase 2 components:**

* `cluster-controller`: Executes cluster management operations and Container Live Migration

Additional components are coming soon: `evictor`, `pod-mutator`, `pod-pinner`, and other Atlas components will be managed by the Operator in future releases.

The Operator automatically:

* Installs the latest versions of managed components
* Applies updates through the Atlas console with a single click
* Self-upgrades when new Operator versions are released
* Migrates existing component installations when the Operator is added to a cluster

# Architecture

The Operator runs as a single pod in the `atlas-agent` namespace and uses minimal resources:

* **Memory**: Typically less than 256 MB
* **CPU**: Typically less than 0.5 CPU

```shell
kubectl get pods -n atlas-agent | grep -i operator
NAME                                 READY   STATUS    RESTARTS   AGE
atlasware-operator-8667b6744d-kmqnn   1/1     Running   0          10m
```

The Operator uses Kubernetes Custom Resources (CRs) to manage configuration:

* **Cluster CR**: Global configuration shared across components (API URLs, cloud provider, cluster metadata)
* **Component CRs**: Individual component configuration (version, Helm values, enablement status)

## Permissions

The Atlas Operator requires Kubernetes permissions to install and manage Atlas components. The Operator's permissions are structured in two levels:

**Phase 1 permissions**: Required to install and manage Phase 1 components (`atlas-agent` and `spot-handler`). These permissions include:

* Read-only cluster-scoped permissions to generate cost and optimization recommendations
* Write permissions within the `atlas-agent` namespace to create and manage component resources
* DaemonSet management permissions to deploy and update `spot-handler`
* Install and update the `atlas-agent` and `spot-handler` components

**Extended permissions (Phase 2)**: Additional permissions required to install and manage components that modify cluster behavior. When a cluster moves from read-only to Phase 2, the Operator receives extended permissions that allow it to:

* Install and manage `cluster-controller` for cluster operations and Container Live Migration
* Create, update, and delete cluster-scoped resources required by Phase 2 components
* Manage additional Phase 2 components as they become Operator-supported

The extended permissions are requested through the `extendedPermissions:true` flag in the Phase 2 onboarding script.

The Operator follows Kubernetes privilege escalation prevention mechanisms: it can only create roles and bindings with permissions it already has, preventing unauthorized privilege escalation.

For a detailed breakdown of specific permissions, see the <Anchor label="Operator permissions reference" target="_blank" href="doc:kubernetes-permissions">Operator permissions reference</Anchor>.

# Installation

<br />

## New cluster onboarding

### Phase 1 (read-only mode)

For new clusters onboarded via the Atlas Console, the Operator is installed automatically during Phase 1 onboarding:

1. The onboarding script installs `atlasware-operator`
2. The Operator pod restarts once after generating certificates (this is expected behavior)
3. The Operator then installs the latest versions of `atlas-agent` and `spot-handler`
4. The cluster remains in `Connecting...` status until all components are successfully installed:

```shell
   kubectl get pods -n atlas-agent -w
   NAME                                 READY   STATUS    RESTARTS   AGE
   atlasware-operator-8667b6744d-kmqnn   1/1     Running   0          93s
   atlas-agent-74989f5596-knfrs        2/2     Running   0          38s
   atlas-agent-74989f5596-vnqvk        2/2     Running   0          38s
   atlas-agent-cpvpa-964fc94b6-n5dkq   1/1     Running   0          38s
   atlas-spot-handler-xk7pm            1/1     Running   0          38s
```

> 📘 Note
>
> Some onboarding journeys (such as Cloud Connect) do not include the Operator by default. In these cases, components are installed using the traditional method, but you can enable the Operator later via <Anchor label="Component Control" target="_blank" href="doc:component-control">Component Control</Anchor>.

## Existing clusters

For clusters already onboarded to Atlas, you can enable the Operator in three ways:

### Option 1: Component Control (recommended)

1. In the Atlas console, select **Manage Organization** in the top right

2. Navigate to **Component control** in the left menu

   <Image border={false} src="https://files.readme.io/54c7d98777a09e1830e11c689daaa4c6ca0849729061e2d2f2a9a27f4bb7040b-image.png" />

3. Click **Enable Now →** in the Operator widget

   <Image border={false} src="https://files.readme.io/f77cd8f930c72c1812668bf512448717b3f88041e506e6fb20ad61e7f10c5a96-image.png" />

4. Select the cluster where you want to install the Operator and click **Enable**

   <Image border={false} src="https://files.readme.io/4cd14ce812f17a2562d0a72bed6f6702555f85568e7fc488d05ff4cc40bd09e0-image.png" />

5. Copy and run the provided installation script

The console-provided Helm script handles cluster registration and automatic migration of existing `atlas-agent` installations.

### Option 2: Manual installation via Helm

For manual installation or automation scenarios:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update atlas-helm

helm upgrade --install atlasware-operator atlas-helm/atlasware-operator \
  --namespace atlas-agent \
  --create-namespace \
  --set apiKeySecret.apiKey="<your-api-key>" \
  --set defaultCluster.provider="<eks|gke|aks>" \
  --set defaultCluster.api.apiUrl="$ATLAS_API_URL" \
  --set defaultComponents.enabled=false \
  --set defaultCluster.terraform=false \
  --atomic \
  --wait
```

**Parameters:**

* `apiKeySecret.apiKey`: Your Atlas Full Access API key (required)
* `defaultCluster.provider`: Cloud provider - `eks`, `gke`, or `aks` (required)
* `defaultCluster.api.apiUrl`: Atlas API endpoint. Defaults to the US endpoint — see [Communication requirements](communication-requirements.md) for regional endpoints.
* `defaultComponents.enabled`: Automatically install `atlas-agent` after Operator installation (defaults to `true`; set to `false` for existing clusters with `atlas-agent` installed already)
* `defaultCluster.terraform`: Set to `false` when installing via Helm or console (defaults to `true`; only set to `true` when using Terraform-based installation)
* `defaultCluster.migrationMode`: How existing components should be migrated - `write` (migrate and manage), `autoUpgrade` (migrate, manage, and upgrade), or `read` (detect only, no management). Defaults to `write`

**Migrating existing `atlas-agent` installations:**

When you install the Operator on a cluster with an existing `atlas-agent`, the Operator automatically detects and migrates the agent based on the `migrationMode` setting:

* **`write` (default)**: Migrates existing components to Operator management, preserving current versions
* **`autoUpgrade`**: Migrates components and immediately upgrades them to the latest versions
* **`read`**: Detects existing components without taking over management (read-only monitoring)

The Operator handles both installation types:

* **Helm-installed agents**: Creates a Component CR matching the current version and takes over management
* **YAML manifest agents**: Extracts configuration, creates a Component CR, and converts to Helm-managed

The migration process:

1. Detects the existing `atlas-agent` deployment
2. Extracts configuration (environment variables, resource settings, version)
3. Creates a Component CR with detected configuration
4. Updates the agent's API key secret reference to use the Operator-managed secret
5. Takes over lifecycle management

<Callout icon="💡" theme="default">
  ### Migration impact

  During migration, the agent may restart once, causing 1-2 cluster snapshots to be skipped. This is normal and does not affect cluster functionality.
</Callout>

### Option 3: During `atlas-agent` update

When updating `atlas-agent`, the console prompts you to install the Operator:

1. Open the `atlas-agent` update drawer
2. Copy the update script (Operator installation is included by default)
3. Run the script to install the Operator, which will in turn upgrade `atlas-agent`

> 📘 Opt-out option
>
> The update drawer includes an option to opt out of Operator installation if needed.

## Terraform installation

Starting from version `0.0.26`, the Atlas Operator can be onboarded and managed via Terraform.

### Required permissions for Terraform installations

When installing the Operator via Terraform, the `atlasware-components` chart (which contains the Custom Resources that the Operator uses) requires a RoleBinding for a pre-installation job. This job verifies that the Operator deployment and at least one pod are running before creating Custom Resources (CRs). The RoleBinding grants these permissions:

```yaml
apiGroups: ["apps"]
resources: ["deployments"]
verbs: ["get", "list"]

apiGroups: [""]
resources: ["pods"]
verbs: ["get", "list", "watch"]
```

For the complete RBAC configuration, see the Operator templates in GitHub: <Anchor label="[https://github.com/atlas/atlasware-operator/blob/main/charts/atlas-atlasware-components/templates/check-operator-job-rbac.yaml](https://github.com/atlas/atlasware-operator/blob/main/charts/atlas-atlasware-components/templates/check-operator-job-rbac.yaml)" target="_blank" href="https://github.com/atlas/atlasware-operator/blob/main/charts/atlas-atlasware-components/templates/check-operator-job-rbac.yaml"><https://github.com/atlas/atlasware-operator/blob/main/charts/atlas-atlasware-components/templates/check-operator-job-rbac.yaml></Anchor>

### Configuration and examples

For configuration details and up-to-date examples, see the Atlas Terraform Provider repository: <Anchor label="[https://github.com/atlas/terraform-provider-atlas/tree/master/examples/atlasware-operator](https://github.com/atlas/terraform-provider-atlas/tree/master/examples/atlasware-operator)" target="_blank" href="https://github.com/atlas/terraform-provider-atlas/tree/master/examples/atlasware-operator"><https://github.com/atlas/terraform-provider-atlas/tree/master/examples/atlasware-operator></Anchor>

## Configuring cluster metadata when IMDS is unavailable

The Atlas Operator requires access to cloud provider metadata to identify cluster properties. By default, the Operator uses the cloud provider's Instance Metadata Service (IMDS). However, in environments where IMDS is restricted or unavailable (such as private networks, air-gapped clusters, or security-hardened environments), you must provide cluster metadata through environment variables.

> 📘 Also applies to the standalone atlas-agent Helm chart
>
> The same `additionalEnv` mechanism applies to the standalone `atlas-agent` Helm chart, not just the Atlas Operator (`atlasware-operator`). When deploying `atlas-agent` directly via Helm on EKS, these env vars are **required**. atlasctl and the Terraform `atlas/eks-cluster/atlas` module set these automatically. The Operator examples below show the correct key names (`additionalEnv.EKS_ACCOUNT_ID`, `additionalEnv.EKS_REGION`, `additionalEnv.EKS_CLUSTER_NAME`) and both values-file and inline `--set` methods.

<Accordion title="When to use this configuration" icon="fa-info-circle">
  Configure environment variables when:

  * IMDS endpoints are blocked by network policies or firewall rules
  * Running in environments without access to cloud provider metadata services
  * Security policies prevent metadata service access
  * The Operator fails with region detection or account identification errors
</Accordion>

<Accordion title="Required environment variables by cloud provider" icon="fa-info-circle">
  <Tabs>
    <Tab title="AWS EKS">
      ```yaml
          - name: EKS_ACCOUNT_ID
            value: "<your-aws-account-id>"    # your AWS account ID
          - name: EKS_REGION
            value: "<your-eks-region>"        # your EKS cluster region
          - name: EKS_CLUSTER_NAME
            value: "<your-cluster-name>"      # your EKS cluster name
      ```
    </Tab>

    <Tab title="GCP GKE">
      ```yaml
          - name: GKE_PROJECT_ID
            value: "<your-gcp-project-id>"    # your GCP project ID
          - name: GKE_CLUSTER_NAME
            value: "<your-cluster-name>"      # your GKE cluster name
          - name: GKE_REGION
            value: "<your-gke-region>"        # your cluster region
          - name: GKE_LOCATION
            value: "<your-gke-location>"      # your cluster zone (optional)
      ```
    </Tab>

    <Tab title="Azure AKS">
      ```yaml
          - name: AKS_LOCATION
            value: "<your-aks-location>"                      # your cluster location
          - name: AKS_SUBSCRIPTION_ID
            value: "<your-azure-subscription-id>"             # your subscription ID
          - name: AKS_NODE_RESOURCE_GROUP
            value: "<your-aks-node-resource-group>"           # your node resource group
      ```
    </Tab>
  </Tabs>
</Accordion>

<Accordion title="Configuration methods" icon="fa-info-circle">
  #### Using Helm values file

  Create a `values.yaml` file with the required environment variables for your cloud provider:

  <Tabs>
    <Tab title="AWS EKS">
      ```yaml
          # values.yaml
          additionalEnv:
            EKS_ACCOUNT_ID: "<your-aws-account-id>"
            EKS_REGION: "<your-eks-region>"
            EKS_CLUSTER_NAME: "<your-cluster-name>"
      ```

      Install or upgrade the Operator with the values file:

      ```shell
          helm upgrade --install atlasware-operator atlas-helm/atlasware-operator \
            --namespace atlas-agent \
            --create-namespace \
            --set apiKeySecret.apiKey="<your-api-key>" \
            --set defaultCluster.provider="eks" \
            --set defaultCluster.api.apiUrl="$ATLAS_API_URL" \
            -f values.yaml
      ```
    </Tab>

    <Tab title="GCP GKE">
      ```yaml
          # values.yaml
          additionalEnv:
            GKE_PROJECT_ID: "<your-gcp-project-id>"
            GKE_CLUSTER_NAME: "<your-cluster-name>"
            GKE_REGION: "<your-gke-region>"
            GKE_LOCATION: "<your-gke-location>" # optional
      ```

      Install or upgrade the Operator with the values file:

      ```shell
          helm upgrade --install atlasware-operator atlas-helm/atlasware-operator \
            --namespace atlas-agent \
            --create-namespace \
            --set apiKeySecret.apiKey="<your-api-key>" \
            --set defaultCluster.provider="gke" \
            --set defaultCluster.api.apiUrl="$ATLAS_API_URL" \
            -f values.yaml
      ```
    </Tab>

    <Tab title="Azure AKS">
      ```yaml
          # values.yaml
          additionalEnv:
            AKS_LOCATION: "<your-aks-location>"
            AKS_SUBSCRIPTION_ID: "<your-azure-subscription-id>"
            AKS_NODE_RESOURCE_GROUP: "<your-aks-node-resource-group>"
      ```

      Install or upgrade the Operator with the values file:

      ```shell
          helm upgrade --install atlasware-operator atlas-helm/atlasware-operator \
            --namespace atlas-agent \
            --create-namespace \
            --set apiKeySecret.apiKey="<your-api-key>" \
            --set defaultCluster.provider="aks" \
            --set defaultCluster.api.apiUrl="$ATLAS_API_URL" \
            -f values.yaml
      ```
    </Tab>
  </Tabs>

  #### Using inline Helm parameters

  <Tabs>
    <Tab title="AWS EKS">
      ```shell
          helm upgrade --install atlasware-operator atlas-helm/atlasware-operator \
            --namespace atlas-agent \
            --create-namespace \
            --set apiKeySecret.apiKey="<your-api-key>" \
            --set defaultCluster.provider="eks" \
            --set defaultCluster.api.apiUrl="$ATLAS_API_URL" \
            --set additionalEnv.EKS_ACCOUNT_ID="<your-aws-account-id>" \
            --set additionalEnv.EKS_REGION="<your-eks-region>" \
            --set additionalEnv.EKS_CLUSTER_NAME="<your-cluster-name>"
      ```
    </Tab>

    <Tab title="GCP GKE">
      ```shell
          helm upgrade --install atlasware-operator atlas-helm/atlasware-operator \
            --namespace atlas-agent \
            --create-namespace \
            --set apiKeySecret.apiKey="<your-api-key>" \
            --set defaultCluster.provider="gke" \
            --set defaultCluster.api.apiUrl="$ATLAS_API_URL" \
            --set additionalEnv.GKE_PROJECT_ID="<your-gcp-project-id>" \
            --set additionalEnv.GKE_CLUSTER_NAME="<your-cluster-name>" \
            --set additionalEnv.GKE_REGION="<your-gke-region>" \
            --set additionalEnv.GKE_LOCATION="<your-gke-location>" # optional
      ```
    </Tab>

    <Tab title="Azure AKS">
      ```shell
          helm upgrade --install atlasware-operator atlas-helm/atlasware-operator \
            --namespace atlas-agent \
            --create-namespace \
            --set apiKeySecret.apiKey="<your-api-key>" \
            --set defaultCluster.provider="aks" \
            --set defaultCluster.api.apiUrl="$ATLAS_API_URL" \
            --set additionalEnv.AKS_LOCATION="<your-aks-location>" \
            --set additionalEnv.AKS_SUBSCRIPTION_ID="<your-azure-subscription-id>" \
            --set additionalEnv.AKS_NODE_RESOURCE_GROUP="<your-aks-node-resource-group>"
      ```
    </Tab>
  </Tabs>

  #### Updating existing Operator installations

  If the Operator is already installed and experiencing IMDS-related failures, update the deployment:

  <Tabs>
    <Tab title="AWS EKS">
      ```shell
          helm upgrade atlasware-operator atlas-helm/atlasware-operator \
            --namespace atlas-agent \
            --reset-then-reuse-values \
            --set additionalEnv.EKS_ACCOUNT_ID="<your-aws-account-id>" \
            --set additionalEnv.EKS_REGION="<your-eks-region>" \
            --set additionalEnv.EKS_CLUSTER_NAME="<your-cluster-name>"
      ```
    </Tab>

    <Tab title="GCP GKE">
      ```shell
          helm upgrade atlasware-operator atlas-helm/atlasware-operator \
            --namespace atlas-agent \
            --reset-then-reuse-values \
            --set additionalEnv.GKE_PROJECT_ID="<your-gcp-project-id>" \
            --set additionalEnv.GKE_CLUSTER_NAME="<your-cluster-name>" \
            --set additionalEnv.GKE_REGION="<your-gke-region>" \
            --set additionalEnv.GKE_LOCATION="<your-gke-location>"
      ```
    </Tab>

    <Tab title="Azure AKS">
      ```shell
          helm upgrade atlasware-operator atlas-helm/atlasware-operator \
            --namespace atlas-agent \
            --reset-then-reuse-values \
            --set additionalEnv.AKS_LOCATION="<your-aks-location>" \
            --set additionalEnv.AKS_SUBSCRIPTION_ID="<your-azure-subscription-id>" \
            --set additionalEnv.AKS_NODE_RESOURCE_GROUP="<your-aks-node-resource-group>"
      ```
    </Tab>
  </Tabs>
</Accordion>

<Accordion title="Verifying the configuration" icon="fa-info-circle">
  After applying the environment variables, verify the Operator is running correctly:

  ```shell
  kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlasware-operator
  ```

  Check the Operator logs for successful startup without metadata-related errors:

  ```shell
  kubectl logs -n atlas-agent -l app.kubernetes.io/name=atlasware-operator
  ```
</Accordion>

# Updating components

## Semi-automated updates

Once the Operator is installed, you can update both the Operator and `atlas-agent` through the Atlas console without running additional scripts.

### Updating the Operator

1. Navigate to **Component Control**
2. Open the **Atlas Operator** widget
3. Click **Update** when a new version is available
4. The Operator performs a self-upgrade

### Updating `atlas-agent`

1. Navigate to **Component Control**
2. Locate `atlas-agent` in the component list
3. Click **Update**
4. The Operator handles the upgrade process automatically

<Callout icon="💡" theme="default">
  ### Best practice

  Always update both the Operator and managed components to the latest verified versions to ensure compatibility and access to the newest features.
</Callout>

## Automatic CRD upgrades

Starting with version 0.0.26, the Operator automatically upgrades Custom Resource Definitions (CRDs) during installation and upgrades. This ensures that CRD schemas stay synchronized with the Operator version.

### How it works

The Operator uses a Helm pre-install/pre-upgrade hook to update CRDs before the main Operator deployment starts. This process:

1. Creates a dedicated ServiceAccount with minimal CRD management permissions
2. Runs a temporary job that applies updated CRD definitions for `clusters.atlasware.atlas.example` and `components.atlasware.atlas.example`
3. Completes the upgrade and cleans up automatically

The CRD upgrade job uses a separate ServiceAccount with permissions scoped only to CRD management, following the principle of least privilege. The entire process completes in seconds during normal operation.

### Expected behavior on first upgrade

When upgrading to version 0.0.26+ from an earlier version, you may see this warning in the upgrade job logs:

```
Warning: resource customresourcedefinitions/clusters.atlasware.atlas.example is missing the
kubectl.kubernetes.io/last-applied-configuration annotation which is required by kubectl apply.
```

**This warning is normal and expected.** It appears because:

* The CRDs were originally installed via Helm's `crds/` directory
* Helm doesn't add the `last-applied-configuration` annotation
* The upgrade job adds this annotation automatically
* Future upgrades won't show this warning

As long as you see "configured" for both CRDs and "CRDs applied successfully", the upgrade worked correctly.

## Future capabilities

The Operator will soon manage additional Atlas components. Some enhancements may require additional permissions:

* When clicking **Update** for the Operator, you may be prompted to run a short script to grant required permissions
* These permissions enable the Operator to handle installation and upgrades for other Atlas components

See <Anchor label="Kubernetes permissions" target="_blank" href="doc:kubernetes-permissions">Kubernetes permissions</Anchor> for more information.

# Removing the Operator

## Before you remove

Removing the Operator stops automated component management. After removal:

* `atlas-agent` and other managed components continue running normally
* You'll need to use manual Helm commands or console scripts to update components
* Custom Resources (CRs) and Custom Resource Definitions (CRDs) are removed from the cluster

## Uninstall via Helm

To completely remove the Operator and its associated resources:

```shell
# Uninstall the Operator
helm uninstall atlasware-operator -n atlas-agent
```

> ⚠️ Important
>
> Uninstalling the Operator does **not** remove `atlas-agent` or other Atlas components. These components continue operating independently. If you want to fully remove all Atlas components, see the <Anchor label="cluster disconnection documentation" target="_blank" href="doc:disconnect-your-cluster">cluster disconnection documentation</Anchor>.

## Console behavior after removal

After manually uninstalling the Operator:

* The console shows the Operator as enabled until the next sync cycle (\~15 minutes)
* Component update prompts may still reference the Operator temporarily
* Use manual Helm commands, Terraform, or console-provided scripts to manage components

> 📘 Note
>
> UI-based Operator disconnection and re-enablement is planned for future releases.

# Troubleshooting

## Cluster stuck in Connecting status

**Symptom**: Cluster remains in `Connecting...` status for more than 10 minutes after running the onboarding script.

**Cause**: The Operator may have installed successfully, but `atlas-agent` failed to deploy.

**Solution**:

1. Check the status of both the Operator and the agent:

   ```shell
   kubectl get pods -n atlas-agent
   ```

2. Check Operator logs for errors:

   ```shell
   kubectl logs -l app.kubernetes.io/name=atlasware-operator -n atlas-agent
   ```

3. If `atlas-agent` is not running, check the Component CR status:

   ```shell
   kubectl get component atlas-agent -n atlas-agent -o yaml
   ```

4. If the status doesn't change after 10 minutes, try the **Enable** button in **Component Control**; alternatively, if that fails, rerun the installation script

5. If the issue persists, contact Atlas Customer Success

## Component updates failing

**Symptom**: Clicking **Update** for a component results in an error, or the update doesn't complete.

**Cause**: The component may require additional permissions that the Operator doesn't have, or there may be a configuration conflict.

**Solution**:

1. Check the Component CR status for error messages:

   ```shell
   kubectl get component <component-name> -n atlas-agent -o yaml
   ```

2. Review Operator logs during the update attempt:

   ```shell
   kubectl logs -l app.kubernetes.io/name=atlasware-operator -n atlas-agent --tail=50
   ```

3. If permission errors appear, the Operator may need additional RBAC permissions. Contact Atlas Customer Success

## Migration from existing agent not working

**Symptom**: After installing the Operator, the existing `atlas-agent` is not migrated automatically.

**Possible causes**:

* The agent deployment doesn't have the required labels or annotations (due to modifications)
* The agent was installed using a non-standard method
* The Operator doesn't have sufficient permissions to modify the existing deployment

**Solution**:

1. Check if the agent has Helm labels:

   ```shell
   kubectl get deployment atlas-agent -n atlas-agent -o jsonpath='{.metadata.labels}'
   ```

2. Check if a Component CR was created:

   ```shell
   kubectl get component atlas-agent -n atlas-agent
   ```

3. If migration fails repeatedly, contact Atlas Customer Success with the agent deployment details

## Failed Operator uninstall

**Symptom**: Running `helm uninstall atlasware-operator` fails or hangs, preventing removal of the Operator.

**Cause**: Helm hooks may be blocking the uninstall process, or there may be finalizers preventing resource deletion.

**Solution**:

1. Uninstall without running Helm hooks:

   ```shell
   helm -n atlas-agent uninstall atlasware-operator --no-hooks
   ```

2. Manually remove Custom Resources:

   ```shell
   kubectl delete clusters.atlasware.atlas.example --all -n atlas-agent
   kubectl delete components.atlasware.atlas.example --all -n atlas-agent
   ```

3. Remove Custom Resource Definitions:

   ```shell
   kubectl delete crd clusters.atlasware.atlas.example
   kubectl delete crd components.atlasware.atlas.example
   ```

4. Verify all resources are removed:

   ```shell
   kubectl get all -n atlas-agent | grep atlasware-operator
   ```

**Cause**: The dedicated CRD upgrade ServiceAccount lacks necessary permissions, or the ClusterRole/ClusterRoleBinding wasn't created properly.

**Solution**:

1. Verify the CRD upgrade ServiceAccount exists:

   ```shell
   kubectl get serviceaccount atlasware-operator-crd-upgrade -n atlas-agent
   ```

2. Verify the CRD upgrade ClusterRole exists and has correct permissions:

   ```shell
   kubectl get clusterrole atlasware-operator-crd-upgrade -o yaml
   ```

3. Check the ClusterRoleBinding:

   ```shell
   kubectl get clusterrolebinding atlasware-operator-crd-upgrade -o yaml
   ```

4. If any resources are missing or incorrect, reinstall the Operator which will recreate these resources:

   ```shell
   helm upgrade --install atlasware-operator atlas-helm/atlasware-operator \
     --namespace atlas-agent \
     --set apiKeySecret.apiKey="your-api-key"
   ```

# Related documentation

* <Anchor label="Component Control" target="_blank" href="doc:component-control">Component Control</Anchor>: View and manage all Atlas components
* <Anchor label="Helm Charts" target="_blank" href="doc:helm-charts">Helm Charts</Anchor>: Manual component installation using Helm
* <Anchor label="Connecting Your Cluster" target="_blank" href="doc:connecting-your-cluster">Connecting Your Cluster</Anchor>: How to connect a cluster and what happens after