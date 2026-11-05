---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Private image scanning

Set up access to private or managed container image registries.

This guide explains how to configure Sentinel to scan images from private registries, providing you with comprehensive vulnerability assessments across all container images in your environment.

## Overview

By default, the Sentinel agent can scan both public and private images running on nodes managed by the Atlas cluster controller. You'll need to configure access to your private registries to get a complete picture of vulnerabilities across your entire environment.

Sentinel supports scanning images from any private registry using image pull secrets, as well as directly integrating with managed registry services from major cloud providers:

* Amazon Elastic Container Registry (ECR) for AWS EKS clusters
* Google Container Registry (GCR) and Artifact Registry for GKE clusters
* Azure Container Registry (ACR) for AKS clusters

## Configuration methods

### Private registries with image pull secret

To enable scanning of images from a private registry using an image pull secret:

1. Create an image pull secret in the `atlas-agent` namespace:

```shell
kubectl -n atlas-agent create secret docker-registry [secret-name] \
  --docker-server=[registry-server] \
  --docker-username=[registry-username] \
  --docker-password=[registry-password]
```

The `[registry-server]` can be specified in one of these formats:

* `{registry}` (e.g., `docker.io`)
* `{registry}/{namespace}` (e.g., `docker.io/atlas`)
* `{registry}/{namespace}/{repository}` (e.g., `docker.io/atlas/agent`)

Example for GitLab private images:

```shell
kubectl -n atlas-agent create secret docker-registry [secret-name] \
  --docker-server=https://registry.gitlab.com \
  --docker-username=registry-user \
  --docker-password=registry-password
```

<Callout icon="📘" theme="info">
  ### Which upgrade method to use

  The Helm commands on this page use the **umbrella chart** (`atlas-helm/atlas`) by default. If you need to use a different method:

  - **[atlasctl](connect-with-atlasctl.md)**: To upgrade all Atlas components at once without managing Helm flags:
    ```shell
    atlasctl atlasware upgrade
    ```
    This preserves your existing configuration. See the [atlasctl documentation](connect-with-atlasctl.md) for installation and authentication instructions.
  - **Individual charts**: If you installed each component as a separate Helm release (e.g., for ArgoCD or custom GitOps), replace the release name and chart reference with the component-specific ones (e.g., `atlas-workload-autoscaler` and `atlas-helm/atlas-workload-autoscaler`) and remove the `autoscaler.atlas-workload-autoscaler.` value prefix.

  Not sure which method you used? Run `helm list -n atlas-agent`. A single release named `atlas` means umbrella chart; separate releases like `atlas-workload-autoscaler` mean individual charts.
</Callout>

<br />

2. Configure Sentinel to use this image pull secret:

> 📘 Configuration-only changes
>
> This command upgrades all components in the umbrella chart, not just Sentinel. To change configuration without upgrading component versions, pin the chart to your current version by adding `--version <your-current-chart-version>` to the command. Run `helm list -n atlas-agent` to check your current version.

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values --set autoscaler.atlas-sentinel.controller.extraArgs.image-private-registry-pull-secret=[secret-name]
```

### Working with multiple registries

You can configure a single image pull secret to scan images from multiple registries:

```shell
cat <<EOF | kubectl apply -n atlas-agent -f -
apiVersion: v1
kind: Secret
metadata:
  name: <secret-name>
type: kubernetes.io/dockerconfigjson
stringData:
  .dockerconfigjson: |
    {
      "auths":
        {
          "docker.io": {
            "username":"user1",
            "password":"password1",
            "auth":"auth1=="
          },
          "docker.io/my-repository": {
            "username":"user2",
            "password":"password2",
            "auth":"auth2=="
          },
          "registry.gitlab.com": {
            "username":"user3",
            "password":"password3",
            "auth":"auth3=="
          }
        }
    }
EOF
```

To generate a configuration for individual registries, use:

```shell
kubectl -n atlas-agent create secret docker-registry [secret-name] \
  --docker-server=[registry-server] \
  --docker-username=[registry-username] \
  --docker-password=[registry-password] \
  --dry-run=client \
  --output go-template='{{ index .data ".dockerconfigjson" | base64decode }}' | jq '.auths'
```

## Cloud provider registry integration

### Amazon Elastic Container Registry (ECR)

For AWS environments, Sentinel can authenticate directly with ECR:

1. Enable OIDC provider for your EKS cluster:

```shell
eksctl utils associate-iam-oidc-provider \
  --cluster <cluster_name> \
  --approve
```

2. Create a service account:

```shell
kubectl create serviceaccount atlas-sentinel-ecr -n atlas-agent
```

3. Attach an IAM policy for ECR access:

```shell
eksctl create iamserviceaccount \
  --name atlas-sentinel-ecr \
  --namespace atlas-agent \
  --cluster <cluster_name> \
  --attach-policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly \
  --approve \
  --override-existing-serviceaccounts
```

4. Configure Sentinel to use this service account:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-service-account=atlas-sentinel-ecr
```

### Microsoft Azure Container Registry (ACR)

For Azure environments:

1. Enable OIDC provider and Workload Identity:

```shell
az aks update -g <resource_group> -n <cluster_name> --enable-oidc-issuer --enable-workload-identity
```

2. Create a Managed Identity:

```shell
az identity create --name <identity_name> --resource-group <resource_group> --location <location> --subscription <subscription>
export IDENTITY_CLIENT_ID="$(az identity show --resource-group <resource_group> --name <identity_name> --query 'clientId' -o tsv)"
```

3. Assign ACR permissions:

```shell
ACR_ID=$(az acr show --name <acr_name> --resource-group <resource_group> --query "id" --output tsv)
az role assignment create --assignee $IDENTITY_CLIENT_ID --role "AcrPull" --scope $ACR_ID
```

4. Create a Kubernetes service account linked to this identity:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    azure.workload.identity/client-id: "${IDENTITY_CLIENT_ID}"
  name: atlas-sentinel-aks
  namespace: atlas-agent
EOF
```

5. Create a federated identity credential:

```shell
export AKS_OIDC_ISSUER="$(az aks show -n <cluster_name> -g <resource_group> --query "oidcIssuerProfile.issuerUrl" -o tsv)"

az identity federated-credential create --name <federated_identity_name> \
  --identity-name <identity_name> \
  --resource-group <resource_group> \
  --issuer "${AKS_OIDC_ISSUER}" \
  --subject system:serviceaccount:atlas-agent:atlas-sentinel-aks \
  --audience api://AzureADTokenExchange
```

6. Configure Sentinel to use this service account:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-service-account=atlas-sentinel-aks
```

### Google Container Registry/Artifact Registry

No additional configuration is required for GKE clusters to scan images from Google Container Registry or Artifact Registry.

## Troubleshooting

If you encounter issues with private image scanning:

### Check for authentication errors in the Sentinel logs

```shell
kubectl logs -l app.kubernetes.io/name=atlas-sentinel-controller -n atlas-agent | grep -i "authentication\|auth\|registry"
```

### Verify your secrets

Ensure your image pull secrets contain the correct credentials:

```shell
kubectl get secret <secret-name> -n atlas-agent -o jsonpath='{.data.\.dockerconfigjson}' | base64 --decode
```

### Check service account configuration

For cloud provider registry integration, verify that the service account is correctly configured:

```shell
kubectl get serviceaccount <service-account-name> -n atlas-agent -o yaml
```

### Next steps

After configuring private image scanning:

* View the [Vulnerabilities](image-security.md) page to see scan results
* Explore other [Sentinel features](sec-configuring-sentinel.md) to configure monitoring and scanning

<Cards columns={2}>
  <Card title="Configuring Sentinel" href="https://docs.atlas.example/docs/sec-configuring-sentinel" icon="fa-gear">
    Configure scanning frequencies and image scanning options.
  </Card>
  <Card title="Image security" href="https://docs.atlas.example/docs/image-security" icon="fa-shield-halved">
    View vulnerability scan results in the security insights report.
  </Card>
</Cards>