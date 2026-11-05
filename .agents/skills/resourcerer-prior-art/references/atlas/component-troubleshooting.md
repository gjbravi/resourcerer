---
updatedAt: 2026-09-28T10:57:42.000Z
---

# Troubleshooting Atlas components

Solutions for resolving issues with Atlas agent, cluster controller, and other platform components.

## Upgrading the agent

### For clusters onboarded via console

To check the version of the agent running on your cluster, use the following command:

```shell
kubectl describe pod atlas-agent -n atlas-agent | grep "Image:"
```

To upgrade the Atlas agent version, please perform the following:

1. Go to [Connect cluster](https://console.atlas.example/external-clusters/new).
2. Select the correct cloud service provider.
3. Run the provided script.

In case of an error when upgrading the agent, e.g. `MatchExpressions:[]v1.LabelSelectorRequirement(nil)}: field is immutable` run the command `kubectl delete deployment -n atlas-agent atlas-agent` and repeat step 3.

The latest version of the Atlas agent is now deployed in your cluster.

### For clusters onboarded via Terraform

By default, Terraform modules do not specify the `atlas-agent` Helm chart version. As a result, the latest available atlas-agent Helm chart is installed when onboarding the cluster, but as new agent versions are released, re-running Terraform doesn't upgrade the agent.

We are looking to solve this, but a short-term fix would be to provide a specific agent version as TF variable [agent-version](https://github.com/atlas/terraform-atlas-aks/blob/559f4773f7dd1df4e87c1c624e846c3b7bd42c24/variables.tf#L105) and reapply the Terraform plan. For valid values of the `atlas-agent` Helm chart, see releases of Atlas [helm-charts](https://github.com/atlas/helm-charts/releases).

***

## Deleted agent

If you delete the Atlas agent deployment from the cluster, you can reinstall it by rerunning the script from the **Connect cluster** screen. Please ensure you choose the correct cloud service provider.

***

## Custom secret management

There are many technologies for managing Secrets in GitOps. Some store the encrypted secret data in a git repository and use a cluster add-on to decrypt it during deployment. Others use a reference to an external secret manager/vault.

The agent helm chart provides the parameter `apiKeySecretRef` to enable the use of Atlas with custom secret managers.

```yaml
# Name of secret with Token to be used for authorizing agent access to the API
# apiKey and apiKeySecretRef are mutually exclusive
# The referenced secret must provide the token in .data["API_KEY"]
apiKeySecretRef: ""
```

### An example of the Atlas agent

Here's an example of using a Atlas agent helm chart with a custom secret:

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-agent.apiKeySecretRef=<your-custom-secret>
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-agent.apiKeySecretRef=<your-custom-secret>
```

  </Tab>
</Tabs>

### An example of the Atlas cluster controller

An example of using the Atlas cluster controller helm chart with a custom secret:

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-cluster-controller.atlas.apiKeySecretRef=<your-custom-secret>
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-cluster-controller.atlas.apiKeySecretRef=<your-custom-secret>
```

  </Tab>
</Tabs>