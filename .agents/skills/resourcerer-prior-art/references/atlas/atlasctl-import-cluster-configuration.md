---
updatedAt: 2026-09-24T12:41:36.000Z
---

# Import and apply cluster configuration with atlasctl

Discover existing node pool configurations from your cloud provider, then import or apply them into Atlas as node configurations and node templates using atlasctl.

`atlasctl cluster import-config` reads your cluster's node group configuration from your cloud provider and translates it into Atlas [node configurations](node-configuration.md) and [node templates](node-templates.md). Nodes provisioned by Atlas then match the ones they replace, with the same subnets, images, security groups, labels, and taints.

Discovery is **read-only** against your cloud account: it never modifies or deletes your existing node groups. The import only creates the corresponding objects inside Atlas.

## What gets imported

Each cloud node group backing your cluster is converted into:

* A [node configuration](node-configuration.md): the cloud-level settings used to build instances (subnets, image, disk, security groups, network placement, and provider-specific options).
* A [node template](node-templates.md): the Kubernetes-level placement rules for the group, such as labels, custom taints, Spot/On-Demand capacity, CPU architecture, OS, GPU-only, and instance-family constraints. Templates that describe the same node group are merged into one, with their constraints combined.

What discovery reads and captures per provider:

| Provider | Read from                                                                                 | Captured settings                                                                                                                                                                                                                                               |
| -------- | ----------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **EKS**  | Auto Scaling groups, EC2 launch templates, EKS managed node groups, IAM instance profiles | Subnets, custom AMIs and image family (AL2, AL2023, Bottlerocket), security groups, instance profile, disk settings (type, size, IOPS, throughput, KMS key), IMDS settings, kubelet and bootstrap configuration parsed from node user data, target groups, tags |
| **GKE**  | GKE node pools, managed instance groups, backend services                                 | Max pods per node, network tags, disk type, zones, min CPU platform, on-host maintenance, pod secondary IP range, local SSD storage, load balancer backend pools, unmanaged instance groups                                                                     |
| **AKS**  | AKS agent pools (VM scale sets), load balancers, virtual networks                         | Max pods per node, image family, encryption at host, pod subnet, OS disk type, accelerated networking, network security group, ephemeral OS disk placement, public IP prefix, load balancer backend pools                                                       |

## How it works

1. `atlasctl` reads the node group configuration for the cluster from your cloud provider.
2. Each node group is translated into a node configuration and a node template.
3. You review the result and confirm. `atlasctl` creates whatever is missing in Atlas, leaving any node configuration or node template that already exists untouched.

<Callout icon="📘" theme="info">
  ### Safe to re-run

  The import only creates what's missing. It never updates, overwrites, or deletes existing node configurations or node templates in Atlas.
</Callout>

## Run the import

With your kubeconfig pointing at the connected cluster, run:

```shell
atlasctl cluster import-config
```

`atlasctl` detects the Atlas cluster from your kubeconfig (or use `--cluster-id` to select it explicitly), runs discovery, prints a summary such as:

```
Based on your cluster config we will create 3 node configuration(s) and 4 node template(s).
```

and opens an interactive review where you can inspect the discovered settings before confirming the apply.

To import without any prompts (useful in CI):

```shell
atlasctl cluster import-config --non-interactive
```

For the full list of `import-config` flags, see the [atlasctl command reference](atlasctl-reference.md).

## Import during connect

You can run the import as part of connecting the cluster, instead of as a separate step:

```shell
atlasctl cluster connect --import-config
```

After a successful connect, the existing cluster configuration is discovered and imported automatically. This needs the same cloud read credentials as standalone `import-config` (see [Requirements](#requirements)); a connect can succeed while the import step still fails if those credentials are missing. If the import fails, the cluster **stays connected**. `atlasctl` prints a warning, and you can retry at any time with `atlasctl cluster import-config`. The import step is skipped when `--dry-run` is set.

See [Connect using the atlasctl CLI](connect-with-atlasctl.md) for the full connect flow.

## The plan file

With `--file`, the discovered plan is written to disk instead of being applied. This lets you review the full plan, check it into git, apply it later, or hand it to another team.

```shell
atlasctl cluster import-config --file plan.yaml
```

The file is a `ClusterBootstrapPlan`, a plain YAML document:

```yaml
apiVersion: atlasctl.atlas.example/v1
kind: ClusterBootstrapPlan
metadata:
  clusterId: 741c0e8e-6b67-4d48-a52a-2a4cbb1e0f61
nodeConfigurations:
  - name: prod-cluster-general-4f2a
    subnets:
      - subnet-0a1b2c3d4e5f6a7b8
    minDiskSize: 50
    containerRuntime: containerd
    tags:
      team: platform
    eks:
      instanceProfileArn: arn:aws:iam::123456789012:instance-profile/atlas-prod-cluster-eks-741c0e8e
      securityGroups:
        - sg-0123456789abcdef0
      imageFamily: FAMILY_AL2023
      volumeType: gp3
nodeTemplates:
  - name: prod-cluster-general-spot-4f2a
    ncName: prod-cluster-general-4f2a
    isDefault: true
    isEnabled: true
    labels:
      workload: general
    constraints:
      spot: true
      onDemand: true
      architectures:
        - amd64
```

The document contains:

* `metadata.clusterId`: the Atlas cluster the plan belongs to.
* `nodeConfigurations`: shared fields (`name`, `subnets`, `image`, `minDiskSize`, `tags`, `containerRuntime`, `initScript`, `kubeletConfig`) plus exactly one provider block (`eks`, `gke`, or `aks`) holding the provider-specific settings listed in the table above.
* `nodeTemplates`: each template binds to a configuration via `ncName` and carries `isDefault`, `isEnabled`, `shouldTaint`, `labels`, `customTaints`, and scheduling `constraints`.

The example above is illustrative. Field names and values follow the plan schema; your file will contain your cluster's actual discovered settings.

## Apply a plan from a file

To apply a plan that was written to a file earlier (for example, after reviewing or editing it, or on a different machine), use:

```shell
atlasctl cluster apply-config --file plan.yaml
```

`apply-config` validates the plan before applying it:

* `apiVersion` must be `atlasctl.atlas.example/v1` and `kind` must be `ClusterBootstrapPlan`.
* `metadata.clusterId` must match the cluster the command runs against. A plan produced for one cluster cannot be applied to another.

Apply then creates the node configurations and node templates described in the file, skipping anything that already exists.

## Export to Terraform

If you manage your cluster with the [Atlas Terraform provider](terraform-provider.md), export the discovered plan as apply-ready Terraform instead of YAML:

```shell
atlasctl cluster import-config --file main.tf --output tf
```

The generated HCL declares the `atlas` provider, a sensitive `atlas_api_token` variable, and a `cluster_id` variable, and contains one `atlas_node_configuration` resource per node configuration and one `atlas_node_template` resource per node template. When the plan references objects that already exist in Atlas, the file includes `import` blocks so Terraform adopts them into state instead of attempting to create duplicates.

Apply it with:

```shell
terraform init
terraform apply
```

`terraform apply` prompts for `atlas_api_token` interactively unless you supply it another way. Avoid typing the raw token into a `-var` flag or an `export` command: both land in your shell history and process list. Instead, populate `TF_VAR_atlas_api_token` from a source that doesn't echo the value, such as your secret manager's CLI:

```shell
export TF_VAR_atlas_api_token="<value from your secret manager>"
terraform apply
```

In CI, set `TF_VAR_atlas_api_token` as a pipeline secret rather than a shell command.

## Requirements

* The cluster is already [connected to Atlas](connect-with-atlasctl.md) with automation capabilities (not read-only).
* `kubectl` is configured for the cluster (used to auto-detect the Atlas cluster ID), or pass `--cluster-id`.
* Cloud credentials with read access to the node group inventory:
  * **EKS**: the default AWS credential chain (SSO, environment variables, or shared profile) for the account and region the cluster runs in.
  * **GKE**: Application Default Credentials (for example, `gcloud auth application-default login`).
  * **AKS**: Azure credentials (for example, `az login`).

`import-config` and `apply-config` are supported for EKS, GKE, and AKS clusters.

## See also

<Cards columns="2">
  <Card title="Node configuration" href="https://docs.atlas.example/docs/node-configuration" icon="fa-server">
    The cloud-level settings discovery imports as node configurations.
  </Card>

  <Card title="Node templates" href="https://docs.atlas.example/docs/node-templates" icon="fa-layer-group">
    The scheduling constraints and placement rules discovery imports as node templates.
  </Card>
</Cards>