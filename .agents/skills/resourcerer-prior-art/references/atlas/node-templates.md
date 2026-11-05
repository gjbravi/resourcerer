---
updatedAt: 2026-10-13T09:01:48.000Z
---

# Node templates

# What are node templates?

Node templates are a key part of the Autoscaler component. They allow users to define virtual buckets of constraints and properties for nodes to be added to the cluster during upscaling. Users can:

* Specify various constraints to limit the inventory of instance types to be used.
* Select a preferred instance resource offering (e.g., build a spot-only cluster without the need for specific tolerations or selectors).
* Define how Atlas should manage Spot Instance interruptions.
* Specify labels and taints to be applied to the nodes.
* Link a [Node configuration](node-configuration.md) to be applied when creating new nodes.
* Enable [Container Live Migration](clm-overview.md) for zero-downtime workload optimization.

When a user enables Atlas, the default Node template is created in the cluster. This template is essential and serves as the default Autoscaler behavior. In addition to the default template, users can set up multiple templates to suit their specific use cases.

<Table align={["left","left","left"]}>
  <thead>
    <tr>
      <th>
        Feature
      </th>

      <th>
        Default Node Template
      </th>

      <th>
        Custom Node Template
      </th>
    </tr>
  </thead>

  <tbody>
    <tr>
      <td>
        Name
      </td>

      <td>
        `default-by-atlas`.
      </td>

      <td>
        Custom (user-defined).
      </td>
    </tr>

    <tr>
      <td>
        Creation
      </td>

      <td>
        Mandatory. Created by Atlas during cluster onboarding.
      </td>

      <td>
        Optional. Created by the user.
      </td>
    </tr>

    <tr>
      <td>
        Deletion
      </td>

      <td>
        Can't be deleted or otherwise removed.
      </td>

      <td>
        Can be deleted.
      </td>
    </tr>

    <tr>
      <td>
        Activation
      </td>

      <td>
        On by default. Can be disabled.
      </td>

      <td>
        Can be disabled.
      </td>
    </tr>

    <tr>
      <td>
        Usage in Autoscaling
      </td>

      <td>
        Used when workloads are not explicitly designated to trigger cluster upscaling through other Node templates.
      </td>

      <td>
        Activated by workloads utilizing autoscaling labels that are also defined in the Node template.
      </td>
    </tr>

    <tr>
      <td>
        Renaming
      </td>

      <td>
        Can't be renamed.
      </td>

      <td>
        Can't be renamed after creation.
      </td>
    </tr>

    <tr>
      <td>
        Setting as default
      </td>

      <td>
        Is the **default by design**. Cannot be changed to non-default.
        Disabling it is recommended if you no longer wish to use it.
      </td>

      <td>
        Cannot be set as default. Attempts to set `isDefault=true` via API will not update the template.
      </td>
    </tr>
  </tbody>
</Table>

> 📘 Note
>
> For the Atlas autoscaler to function, the Unscheduled Pods policy and at least one Node template (default or not) must be enabled.

# Main attributes

The table below outlines the Node template attributes and their availability across different cloud providers.

<Table align={["left","left","left"]}>
  <thead>
    <tr>
      <th>
        Attribute
      </th>

      <th>
        Description
      </th>

      <th>
        Availability per Cloud provider
      </th>
    </tr>
  </thead>

  <tbody>
    <tr>
      <td>
        **Custom labels and taints**
      </td>

      <td>
        Choose this if custom labels or taints should be applied on the nodes created by Atlas.

        Based on the selected properties, `nodeSelector` and toleration must be applied to the workload to trigger the autoscaling.
      </td>

      <td>
        All
      </td>
    </tr>

    <tr>
      <td>
        **Node configuration link**
      </td>

      <td>
        By default, Node templates mostly focus on the type of resources that need to be scheduled. To learn how to configure the scheduling of your resources and how to use other kinds of attributes on provisioned nodes, see

        [Node configuration](node-configuration.md)
      </td>

      <td>
        All
      </td>
    </tr>

    <tr>
      <td>
        **Target capacity reservations**
      </td>

      <td>
        Select one or more AWS capacity reservations (ODCRs or Capacity Blocks) for the autoscaler to target when provisioning nodes. Selecting a reservation locks the instance type and AZ to match the reservation's constraints. Nodes are tainted with `reserved=only:NoSchedule`. See [Target capacity reservations](#target-capacity-reservations).
      </td>

      <td>
        AWS EKS only
      </td>
    </tr>

    <tr>
      <td>
        **Container Live Migration**
      </td>

      <td>
        Enable zero-downtime workload optimization through container live migration. When enabled, this feature automatically installs the live migration controller and configures nodes with specialized VPC CNI for TCP preservation.

        * **Important**: Configure your node template to include only compatible instance types from the same generation family. Migration will fail between different generation families (e.g., c3 and c5 are incompatible, but c5 and r5 are compatible). For complete requirements, see [Container Live Migration Requirements](clm-requirements-and-limitations.md).
      </td>

      <td>
        AWS EKS only
      </td>
    </tr>

    <tr>
      <td>
        **Stuck pod resize reconciliation**
      </td>

      <td>
        When enabled, Atlas resolves in-place pod resizes stuck on a node without enough unallocated resources. It provisions a new node and evicts other pods onto it, using [Container Live Migration](clm-overview.md) when possible. If nothing else can make room, Atlas evicts the stuck pod itself. Disabled by default. Set via the API (see [In-Place Pod Resizing](woop-in-place-resizing.md#podresizepending)).
      </td>

      <td>
        All
      </td>
    </tr>

    <tr>
      <td>
        **Resource offering**
      </td>

      <td>
        This attribute tells the Autoscaler which node resource offering (spot or on-demand) it should create. There is an option to configure both offerings, in which case workloads targeted to run on spot nodes must be marked accordingly. Check this [guide](spot.md).

        Additional configuration settings available when you want the Autoscaler to use **Spot** nodes:

        * **Spot fallback**: Atlas will create temporary on-demand fallback nodes when spot capacity is unavailable in the cloud.
        * **Interruption prediction model** — Atlas can react to its own Machine Learning prediction model signals (AWS, GCP, Azure) to predict Spot interruptions and proactively rebalance affected nodes. See this [guide](spot.md#interruption-prediction-model) for more details.
        * **Spot Reliability** — Uses machine learning to select more reliable Spot Instances with longer average runtime before interruption. Configure a maximum price increase tolerance (0-100%) to balance reliability with cost optimization.  Available via API, Terraform, and the console UI. [Read more](spot.md#spot-instance-reliability).

        **Diversified Spot Instances** — By default, Atlas seeks the most cost-effective instances without assessing your cluster's current composition. To limit the impact of a potential mass spot reclaim, you can instruct the Autoscaler to evaluate and enhance the diversity of spot nodes in the cluster, but this may increase your costs. [Read more](spot.md#spot-diversity). **Note: This feature is being deprecated in favor of Spot Reliability.**
      </td>

      <td>
        All
      </td>
    </tr>

    <tr>
      <td>
        **Processor architecture**
      </td>

      <td>
        You can select `x86_64`, `ARM64`, or both architecture nodes to be created by Atlas. When using a multi-architecture Node template, also use the `nodeSelector` `kubernetes.io/arch: "arm64"` to ensure that the Pod lands on an ARM node.

        On Azure AKS, ARM64 VMs must be available in the selected region, and your subscription needs quota for ARM64 VMs.
      </td>

      <td>
        All
      </td>
    </tr>

    <tr>
      <td>
        **GPU-enabled instances**
      </td>

      <td>
        Choose this attribute to run workloads only on GPU-enabled instances. Once you select it, **Instance constraints** get enhanced with GPU-related properties.

        If you are unable to select this attribute in the console, there likely are no GPUs in the region of your cluster. See your CSP's documentation for reconciling what GPUs are available in which regions:

        * [GCP documentation](https://cloud.google.com/compute/docs/gpus/gpu-regions-zones)
        * [AWS documentation](https://docs.aws.amazon.com/ec2/latest/instancetypes/ec2-instance-regions.html)

<Callout icon="📘" theme="info">
  ### Note

  Custom extended resources (for example, resources exposed by a customer-owned device plugin) cannot drive Atlas autoscaling. See [GPU instances](gpu.md) for supported alternatives and how to contact support.
</Callout>
      </td>

      <td>
        AWS, GCP
      </td>
    </tr>

    <tr>
      <td>
        **GPU sharing**
      </td>

      <td>
        Enable GPU time-slicing to allow multiple workloads to share physical GPUs. Configure the number of shared clients per GPU (1-48). Currently supported on AWS EKS with Bottlerocket AMI only.

        For complete GPU setup and sharing configuration, see [GPU instances](gpu.md).  
        For GPU time-slicing configuration details, see [GPU sharing with time-slicing](gpu-sharing-time-slicing.md).
      </td>

      <td>
        AWS, GCP
      </td>
    </tr>

    <tr>
      <td>
        **Apply Instance constraints**
      </td>

      <td>
        Apply additional constraints on instances to be selected, such as:

        * Instance Family
        * Min/Max CPU
        * Min/Max Memory
        * Compute-optimized
        * Storage-optimized
        * GPU manufacturer, name, count
        * Bare metal. AWS-only feature
        * Burstable instances
        * Fractional GPUs
        * The list of AZ names to consider for the node template. If the list is empty or not set, all AZs are considered. When the subnets defined in the Node Configuration are zonal (e.g., in AWS or Azure), the effective AZs are determined as the intersection of the subnet AZs and the AZs specified in the node template. These AZ restrictions also limit where the autoscaler may create a node for a pod that sets a hard zone topology spread constraint; see [Topology spread constraints and node provisioning](autoscaler.md#topology-spread-constraints-and-node-provisioning).
        * OS selection (Windows 2019/2022/2025 or Linux). Azure-only feature. Ensure that the quota for `Standard_DS2_v2` VMs is available. Contact the support team, as this feature has to be enabled for the organization.
        * Customer-specific. Azure-only feature. When enabled, the inventory will include customer-specific (preview) instances; otherwise, they will be excluded. Contact the support team, as this feature must be enabled for the organization.
      </td>

      <td>
        All
      </td>
    </tr>

    <tr>
      <td>
        **Set up instance family prioritization**
      </td>

      <td>
        Custom priority order can be defined in cases where it is preferred that the Autoscaler prioritize certain instance families when creating new nodes.  
        The priority queue is organized into tiers, each representing a set of instance families. The Autoscaler will attempt to find an instance type from the highest available tier before moving down to the next one. Instance families within the same priority tier are treated equally.  
        Instance family prioritization takes precedence over [commitment](commitments.md)-based instance selection. If a priority tier specifies a family that is not covered by a commitment, the Autoscaler still selects from that tier first. To maximize commitment utilization and align with priority preferences, place commitment-covered families in the highest-priority tier.  
        When no instance families from the priority queue are available, the Autoscaler will fall back to its default behavior, which includes prioritizing commitment-covered instances.
      </td>

      <td>
        All
      </td>
    </tr>

    <tr>
      <td>
        **Custom instances**
      </td>

      <td>
        The Autoscaler will consider <Anchor label="GCP custom VM types" target="_blank" href="https://cloud.google.com/compute/docs/instances/creating-instance-with-custom-machine-type">GCP custom VM types</Anchor> in addition to predefined machines.
        The <Anchor label="extended memory" target="_blank" href="https://cloud.google.com/compute/docs/instances/creating-instance-with-custom-machine-type#extendedmemory">extended memory</Anchor> setting will allow the Autoscaler to determine whether a custom VM with increased memory per vCPU would be optimal.
      </td>

      <td>
        GCP
      </td>
    </tr>

    <tr>
      <td>
        **Set CPU limit**
      </td>

      <td>
        Controls the total CPU cores for nodes provisioned with this template.
      </td>

      <td>
        All
      </td>
    </tr>
  </tbody>
</Table>

## Container Live Migration configuration

When you enable Container Live Migration in a node template, Atlas automatically:

* Configures nodes with the necessary runtime components
* Sets up the specialized VPC CNI for TCP preservation

### Important considerations for CLM

**Processor architecture**: Configure your node template for a single processor architecture (either AMD64 or ARM64, not "Any"). Live migration cannot occur between nodes with different architectures. Set the processor architecture to a specific value rather than selecting "Any."

**Instance type compatibility**: Configure your node template to include only instance types from the same generation family. For example:

* **Compatible**: c5 with r5, m5 (same generation)
* **Incompatible**: c3 with c5 (different generations)

**Infrastructure requirements**: Ensure your cluster meets all <Anchor label="Container Live Migration requirements" target="_blank" href="doc:clm-requirements-and-limitations">Container Live Migration requirements</Anchor> before enabling this feature.

**Integration with Evictor**: When both CLM and <Anchor label="Evictor" target="_blank" href="doc:evictor">Evictor</Anchor> are enabled, Evictor automatically attempts live migration for eligible workloads before falling back to traditional eviction.

## Processor Architecture Support

Atlas supports different processor architectures across cloud providers through node templates. You can create architecture-specific node templates or multi-architecture templates that include both x86 and ARM options.

| Cloud Provider | x86-64 (AMD64) | ARM64 | Notes                                          |
| -------------- | -------------- | ----- | ---------------------------------------------- |
| AWS EKS        | ✓              | ✓     | ARM instances use AWS Graviton processors      |
| GCP GKE        | ✓              | ✓     | ARM support is available through T2A instances |
| Azure AKS      | ✓              | ✓     |                                                |

### Using ARM on AKS

ARM support on Azure AKS works like on other cloud providers — select **ARM64** (or **Any** for multi-architecture) as the processor architecture in your node template.

Before using ARM on AKS, ensure:

1. ARM64 VMs are available in your selected Azure region (see the [Azure VM sizes documentation](https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/overview) for availability by region)
2. Your Azure subscription has quota for ARM64 VMs
3. Your workloads are built for ARM64, or use multi-architecture images

# Create a Node template

You have the following options to create Node templates:

* <Anchor label="Create Node template through the API" target="_blank" href="ref:nodetemplatesapi_createnodetemplate">Create Node template through the API</Anchor>
* Create Node template through the UI (**Cluster** → **Autoscaler** → **Node templates**)
* Generate templates automatically – Use Atlas's auto-generation feature
* <Anchor label="Terraform" target="_blank" href="https://registry.terraform.io/namespaces/atlas">Terraform</Anchor>

When you create a Node template, you may want to associate it with a Node configuration for provisioning nodes. You can achieve this by linking the template with a custom or default <Anchor label="Node configuration" target="_blank" href="doc:node-configuration">Node configuration</Anchor>.

## Auto-generating Node templates

Atlas can automatically generate optimized node templates based on your cluster's workloads and requirements.

### How auto-generation works

1. **Analysis**: Atlas analyzes your current cluster workloads
2. **Generation**: Multiple template suggestions are created
3. **Review**: You can review each suggested template and its configuration
4. **Selection**: Choose which templates to add to your cluster
5. **Customization**: Edit template names before finalizing; settings can only be edited after adding the templates to your cluster

### Using auto-generation

To use the node template auto-generation:

1. Navigate to **Cluster** → **Autoscaler** → **Node templates**
2. Click **Create template** → **Generate templates**

   ![](https://files.readme.io/ae570d951469930e2755d147a63550ce60474b11a3f49e33a3003e20ef5c69a6-image.png)
3. Review the suggested templates in the preview and select the ones you want to add to your cluster. If necessary, customize template names.

   ![](https://files.readme.io/4f302aaef0f2aaea5f6c179ec13a853446809a86b3053d8bbe9efe0b5f1a643e-image.png)
4. Click **Add templates** to create the selected templates

Auto-generated templates are created in a disabled state, to allow you to configure them before enabling.

### API access

You can also generate templates programmatically using the <Anchor label="Generate Node Templates API" target="_blank" href="https://docs.atlas.example/reference/nodetemplatesapi_generatenodetemplates#/">Generate Node Templates API</Anchor>.

## Using the `shouldTaint` flag

While creating a Node template, you can choose whether the nodes created by the Atlas Autoscaler should be tainted. This is controlled through `shouldTaint` property in the API payload.

> 🚧 When shouldTaint is set to false
>
> Since no taints will be applied on the nodes created by Atlas, any pods being deployed to the cluster, even the ones without `nodeSelector`, might get scheduled on these nodes. This effect might not always be desirable.

### When `shouldTaint` is set to `true`

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox-sleep
spec:
  nodeSelector:
    scheduling.atlas.example/node-template: spark-jobs
  tolerations:
    - key: scheduling.atlas.example/node-template
      value: spark-jobs
      operator: Equal
      effect: NoSchedule
  containers:
    - name: busybox
      image: busybox:1.28
      args:
        - sleep
        - "1200"
```

### When `shouldTaint` is set to `false`

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox-sleep
spec:
  nodeSelector:
    scheduling.atlas.example/node-template: spark-jobs
  containers:
    - name: busybox
      image: busybox:1.28
      args:
        - sleep
        - "1200"
```

## Using `nodeSelector`

You can use `nodeSelector` to schedule pods on the nodes created using the template. By default, you construct `nodeSelector` using the template name. However, you can opt for a custom label if it suits your use case better.

#### Using a Node template name in `nodeSelector`

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox-sleep
spec:
  nodeSelector:
    scheduling.atlas.example/node-template: spark-jobs
  containers:
    - name: busybox
      image: busybox:1.28
      args:
        - sleep
        - "1200"
```

> 📘 Recommended: use `scheduling.atlas.example/node-template` for explicit template selection
>
> When a pod specifies `scheduling.atlas.example/node-template: <name>` in its `nodeSelector` or `nodeAffinity`, the autoscaler uses that as a **direct, unambiguous template selection**. It does not evaluate the template's custom labels at all, and the `Partial node template matching` setting has no effect.
>
> This is the most reliable way to bind a workload to a specific node template. It avoids any ambiguity that can arise when matching is based solely on custom labels — for example, when only a subset of the template's custom labels is specified on the pod.
>
> Custom labels defined on the template will still be applied to the provisioned node, and the pod's remaining `nodeSelector` entries will be satisfied by the Kubernetes scheduler as usual.

### Using multiple custom labels in `nodeSelector`

In case you have a Node template with multiple custom labels `custom-label-key-1=custom-label-value-1` and `custom-label-key-2=custom-label-value-2`. You can schedule your pods on a node created using that Node template by providing `nodeSelector` with all the custom labels as described below:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox-sleep
spec:
  nodeSelector:
    custom-label-key-1: custom-label-value-1
    custom-label-key-2: custom-label-value-2
  containers:
    - name: busybox
      image: busybox:1.28
      args:
        - sleep
        - "1200"
```

## Using `nodeAffinity`

You can use `nodeAffinity` to schedule pods on the nodes created using the template. By default, you construct `nodeAffinity` using the template name. However, you may use a custom label to better fit your use case. The only supported `nodeAffinity` operator is `In`.

#### Using the Node template name in `nodeAffinity`

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox-sleep
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - matchExpressions:
            - key: scheduling.atlas.example/node-template
              operator: In
              values:
                - "spark-jobs"
  containers:
    - name: busybox
      image: busybox:1.28
      args:
        - sleep
        - "1200"
```

### Using a Node template's custom labels in `nodeAffinity`

In case you have a Node template with multiple custom labels `custom-label-key-1=custom-label-value-1` and `custom-label-key-2=custom-label-value-2`. You can schedule your pods on a node created using that Node template by providing `nodeAffinity` with all the custom labels as described below:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox-sleep
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - matchExpressions:
            - key: custom-label-key-1
              operator: In
              values:
                - "custom-label-value-1"
            - key: custom-label-key-2
              operator: In
              values:
                - "custom-label-value-2"
  containers:
    - name: busybox
      image: busybox:1.28
      args:
        - sleep
        - "1200"
```

### Using a mix of `nodeAffinity `and `nodeSelector`

In case you have a Node template with multiple custom labels `custom-label-key-1=custom-label-value-1` and `custom-label-key-2=custom-label-value-2`, you can schedule your pods on a node created using that Node template by providing `nodeAffinity` and `nodeSelector` with all the custom labels as described below:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox-sleep
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - matchExpressions:
            - key: custom-label-key-1
              operator: In
              values:
                - "custom-label-value-1"
  nodeSelector:
    custom-label-key-2: custom-label-value-2
  containers:
    - name: busybox
      image: busybox:1.28
      args:
        - sleep
        - "1200"
```

# More Specific Requirements/Constraints

Templates also support further instance type selection refinement via additional `nodeAffinity`/ `nodeSelector` values. This can be done if the additional constraints don't conflict with the template constraints.

Here's an example with some assumptions:

* The template has no constraints;
* The template is named my-template;
* The template has custom labels enabled, and they are as follows:

```Text YAML
product: "my-product"
team: "my-team"
```

Here's an example deployment that would further specify what a pod supports/needs:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  labels:
    app: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: "team"
                operator: In
                values: ["my-team"]
                # Pick only nodes that are compute-optimized
              - key: "scheduling.atlas.example/compute-optimized"
                operator: In
                values: ["true"]
                # Pick only nodes that are storage-optimized
              - key: "scheduling.atlas.example/storage-optimized"
                operator: In
                values: ["true"]
      nodeSelector:
        # template selector (can also be in affinities)
        product: "my-product"
        team: "my-team"
        # Storage optimized nodes will also have a taint, so we need to tolerate it.
      tolerations:
        - key: "scheduling.atlas.example/storage-optimized"
          operator: Exists
          # toleration for the template
        - key: "scheduling.atlas.example/node-template"
          value: "my-template"
          operator: "Equal"
        effect: "NoSchedule"
      containers:
      - name: nginx
        image: k8s.gcr.io/nginx-slim:0.8
        ports:
        - containerPort: 80
        resources:
          requests:
            cpu: 700m
            memory: 700Mi
          limits:
            memory: 700Mi
```

Using this example, you'd get a storage-optimized, compute-optimized node on a template that doesn't have those requirements for the whole pool.

Note: If a template has one of those constraints (on the template itself), there is currently no ability to loosen requirements for some pods based on their `nodeAffinity`/ `nodeSelector` values.

> 📘 Per-workload hardware constraints without a custom template
>
> If you only need to restrict hardware for a single workload — not an entire node pool — you can use Atlas's hardware scheduling labels (`scheduling.atlas.example/cpu-manufacturer`, `scheduling.atlas.example/instance-family`, `scheduling.atlas.example/instance-cpu`, and others) directly in the workload's `nodeSelector` or `nodeAffinity`. Atlas will provision a matching node using the default template. See [Pod placement — scheduling by hardware characteristics](pod-placement.md#scheduling-by-hardware-characteristics).

# Node template matching

The autoscaler uses two distinct paths to match a pending pod to a node template:

**1. Explicit template selector (recommended)**

If the pod's `nodeSelector` or `nodeAffinity` contains the key `scheduling.atlas.example/node-template`, the autoscaler performs a direct name lookup — it matches the pod to the template whose name equals the selector value. The template's custom labels are not evaluated, and the `Partial node template matching` setting has no effect on this path. This is the most predictable matching behavior and is the recommended approach for binding workloads to a specific template.

**2. Custom label matching**

If the pod does not use `scheduling.atlas.example/node-template`, the autoscaler attempts to match the pod to a template based on the template's `Custom labels`. By default, workloads must specify node selectors or affinity that match **all terms** defined in the Node template's `Custom labels`. If at least one term is not matched, then the Node template is not considered.

A workload can match multiple Node templates. In such cases, the templates are scored primarily by the number of terms that the Node template contains (there are additional scoring criteria). The greater the number of terms, the higher the score.

## Partial Node template matching

Partial matching allows workloads to specify **at least one** of all terms of the Node template for it to be considered when provisioning a node for the workload. To enable partial Node template matching, go to **Autoscaler settings**, navigate to **Unscheduled pods policy**, and select **Partial node template matching**.

![](https://files.readme.io/80422aa-image.png)

When partial matching is enabled, the `Unscheduled Pods Policy Applied` event will include information on how the Node template was selected. This can be used to troubleshoot autoscaler decisions.

### Node template selection

When multiple Node templates are matched, they are sorted, and the first one is selected to create a node for the workload. The sorting criteria, in order of their priority, are:

* Node templates where all terms were matched are sorted higher.
* The count of matched terms, in descending order.
* The average price per CPU of instance types within the Node template, in descending order.
* The name of the template, in alphabetical order.

# Node template CPU limit

You can enable and configure CPU limits on node templates to control the total number of CPU cores provisioned during cluster upscaling and rebalancing. When configured, the total number of CPU cores of all nodes provisioned using that template will not exceed the specified limit.

> 📘 Cluster-wide CPU limits are separate
>
> Node template CPU limits apply only to nodes provisioned from a specific template. They are separate from the [cluster-wide CPU limits policy](autoscaler-settings.md#cluster-cpu-limits-policy), which applies to the total CPU capacity of all worker nodes in the cluster.

<Image align="center" width="75% " src="https://files.readme.io/7cebb9dad67c21a9536c06c2f988b2c15585c74145e4d2158772deb0632f3206-image.png" />

> 🚧 Caution
>
> When using partial node template matching, it's important to note that if CPU limits are reached on the best-matching template, the pods will remain pending. This can impact your cluster's ability to schedule workloads effectively.

To set a CPU limit:

1. Enable the CPU limit toggle
2. Enter the maximum number of CPU cores allowed to provision for nodes under this template

## Impact on Autoscaler

When a CPU limit is configured, the autoscaler will only provision nodes up to the specified CPU limit when responding to unscheduled pods. If pods cannot be scheduled due to the CPU limit, they will receive a pod event indicating that autoscaling failed due to the node template CPU limit.

## Impact on Rebalancing

During rebalancing, the total CPU count may temporarily exceed the configured limit. However, after rebalancing, the final cluster state will respect the configured CPU limits.

# Spot Reliability Configuration

Spot Reliability uses machine learning to identify and select more reliable Spot Instances, reducing interruption frequency for your workloads. This feature analyzes historical Spot Instance data to predict which instance types are likely to run longer before being interrupted.

## Configuration

Configure Spot Reliability when creating or updating a node template by including these parameters in the constraints object:

```json
{
    "name": "reliable-spot-template",
    "isEnabled": true,
    "constraints": {
        "spot": true,
        "enableSpotReliability": true,
        "spotReliabilityPriceIncreaseLimitPercent": 20
    }
}
```

### Parameters

| Parameter                                  | Type    | Required | Description                                        | Range      |
| ------------------------------------------ | ------- | -------- | -------------------------------------------------- | ---------- |
| `enableSpotReliability`                    | boolean | Yes      | Enables ML-based, reliable Spot Instance selection | true/false |
| `spotReliabilityPriceIncreaseLimitPercent` | integer | Yes      | Maximum price increase tolerance for reliability   | 0-100      |

# Support of Dedicated (Sole tenant) nodes

> 📘 This is a new feature currently supported only on GKE clusters
>
> Please note that the sole tenancy node group must be shared with the GKE cluster's project or organization.

Dedicated nodes (also known as Sole tenant nodes) are dedicated hosts in the customer's cloud account that can host VMs that are used in the Kubernetes clusters. The setup for GKE clusters can be found in the Google Cloud <Anchor label="documentation" target="_blank" href="https://cloud.google.com/kubernetes-engine/docs/how-to/sole-tenancy">documentation</Anchor>.

Atlas Autoscaler can be instructed to prefer dedicated hosts until capacity is available when the following parameters are set in the <Anchor label="Node template" target="_blank" href="https://docs.atlas.example/reference/nodetemplatesapi_updatenodetemplate">Node template</Anchor>:

<Table align={["left","left"]}>
  <thead>
    <tr>
      <th>
        Attribute
      </th>

      <th>
        Description
      </th>
    </tr>
  </thead>

  <tbody>
    <tr>
      <td>
        Affinity
      </td>

      <td>
        The affinity rules required for choosing the node dedicated note group. It is constructed using:
        **Key**
        **Value**

        * _Operator_*:
          `IN`: In values
          `NotIn`: Not in values
          `Exists`: Just exist
          `DoesNotExist`: Values do not exist
          `Gt`: Greater then
          `Lt`: Lower then
      </td>
    </tr>

    <tr>
      <td>
        azName
      </td>

      <td>
        The availability zone of the dedicated node group
      </td>
    </tr>

    <tr>
      <td>
        InstanceTypes
      </td>

      <td>
        Instance types of the dedicated node group
      </td>
    </tr>

    <tr>
      <td>
        cpusPerGpu
      </td>

      <td>
        Target number of CPUs to allocate per GPU on this dedicated/sole-tenant node group. This allows configuring nodes so that when all GPUs are in use, all CPUs are also utilized, maximizing resource efficiency. For a node with N GPUs and M CPUs, set `cpusPerGpu = M / N`. If not set, no ratio-based sizing is applied. Must be greater than 0.
      </td>
    </tr>

    <tr>
      <td>
        minGpusPerNode
      </td>

      <td>
        Minimum number of GPUs per node for this dedicated/sole-tenant node group. Only node types with at least this many GPUs will be selected. If not set, no minimum GPU count filtering is applied. Must be greater than or equal to 0.
      </td>
    </tr>

    <tr>
      <td>
        maxCpu
      </td>

      <td>
        Maximum number of CPUs that can be provisioned from this dedicated/sole-tenant node group across all nodes for the specific node template. Once the CPU budget is exhausted, the autoscaler falls back to non-sole-tenant nodes. 
        If not set, no CPU cap is applied. Must be greater than 0.
        When `maxCpu` is set, the affinity entry must include a `name`, and names must be unique among entries that have `maxCpu` set.
        If `cpusPerGpu` is set, `maxCpu` must be a multiple of `cpusPerGpu`. If both `cpusPerGpu` and `minGpusPerNode` are set (and `minGpusPerNode` is greater than 0), `maxCpu` must be a multiple of `cpusPerGpu * minGpusPerNode`, since the smallest valid node has `minGpusPerNode * cpusPerGpu` CPUs.
      </td>
    </tr>
  </tbody>
</Table>

Example of configuration setup in the API response:

```json
    "dedicatedNodeAffinity": [
      {
        "instanceTypes": [
          "n1-node-96-624"
        ],
        "name": "test-nodegroup",
        "affinity": [
          {
            "operator": "IN",
            "key": "stnodegroup",
            "values": [
              "test-group"
            ]
          },
          {
            "operator": "IN",
            "key": "compute.googleapis.com/project",
            "values": [
              "test-123"
            ]
          }
        ],
        "azName": "europe-west1-b",
        "cpusPerGpu": 16,
        "minGpusPerNode": 4,
        "maxCpu": 64
      }
    ]
```

Please note that all constraints configured in the node template will be respected (e.g., the minimum CPU `minCPU` set to `16`).

Once the dedicated host capacity is saturated, Atlas will not add additional dedicated hosts to the node group. Instead, it will return to multi-tenant instances configured in the Node template. During the fallback to multi-tenant instances, the same availability zone will be selected as configured in the `dedicatedNodeAffinity`.

> 🚧 Notice
>
> Currently, the Rebalancer does not simulate the deletion of instances running on dedicated nodes; therefore, rebalancing the whole cluster might produce some suboptimal results.

# Target capacity reservations

> 📘 AWS only
>
> Capacity reservation targeting is currently available for AWS EKS clusters only.

You can configure a Node Template to target one or more AWS capacity reservations: [On-Demand Capacity Reservations (ODCRs)](aws-capacity-reservations.md#working-with-odcrs) or [EC2 Capacity Blocks for ML](aws-capacity-reservations.md#working-with-capacity-blocks). When the autoscaler detects pending pods matching a Node Template with capacity reservation targeting, it provisions nodes into the reserved capacity first, up to the reserved instance count.

To target capacity reservations:

1. Create or edit a Node Template.
2. In the **Target capacity reservations** section, click **+ Add**.
3. Select one or more reservations from the **Capacity reservation** dropdown. Each entry shows the reservation name and its current status (Active or Expired).
4. Click **Save changes**.

<Image align="center" caption="Target capacity reservations section in a Node Template" src="https://files.readme.io/c55e5b65f6ab1c737763a4839c0e7acf9cf1c6d985ac458acb718e7a16bb4345-image.png" />

## Auto-populated constraints

When you select a capacity reservation, Atlas automatically populates and locks the following constraints from the reservation:

**Instance type** — Set to the reservation's instance type (for example, `p5.48xlarge`). This field cannot be changed while the reservation is targeted.

**Availability Zone** — Set to the reservation's AZ. This field cannot be changed while the reservation is targeted.

These inherited constraints ensure that the Node Template only provisions nodes compatible with the reservation. To change these constraints, remove the reservation targeting first.

You can still configure all other Node Template settings, including Node Configuration, custom labels, taints, and fallback purchase options (On-Demand or Spot).

## Taints and tolerations

Nodes provisioned from a capacity reservation are automatically tainted with `reserved=only:NoSchedule`. Your workloads must include a matching toleration to be scheduled on these nodes:

```yaml
tolerations:
  - key: "reserved"
    value: "only"
    operator: "Equal"
    effect: "NoSchedule"
```

This taint ensures only workloads explicitly configured for reserved capacity are placed on these nodes.

## Fallback behavior

If the targeted reservation is fully utilized or unavailable, the autoscaler evaluates the fallback options configured on the Node Template:

**If On-Demand or Spot is enabled**, the autoscaler provisions using those options as usual, following the standard candidate ordering.

**If no fallback options are configured**, pods remain pending until capacity becomes available in the reservation.

You can target multiple capacity reservations on a single Node Template. If one reservation is exhausted, the autoscaler moves to the next before considering non-reserved options.

## Multiple Node Templates targeting the same reservation

Multiple Node Templates can target the same capacity reservation. Atlas tracks utilization across all Node Templates and clusters. Since AWS enforces capacity limits at the reservation level, over-assignment may cause provisioning attempts to fail when the reservation is already fully utilized.

For complete documentation on capacity reservations, see [AWS capacity reservations](aws-capacity-reservations.md).

# High availability

## Maintaining minimum nodes per template

While Atlas doesn't directly support setting a minimum node count per template, you can maintain a fixed number of nodes using a helper deployment. This is useful for maintaining a fixed set of nodes at all times for specific workloads.

### Implementation

You can maintain a minimum number of nodes by creating a helper deployment. This deployment uses minimal resources to keep nodes active without affecting your actual workloads:

* Places **one pod per node** using `podAntiAffinity`
* Uses the desired node template via `nodeSelector`
* Prevents node removal using Atlas's removal-disabled label:
  ```yaml
  autoscaling.atlas.example/removal-disabled: "true"
  ```

Use this helper deployment with the desired node template for which you always wish to have a fixed number of nodes.

#### Helper deployment

This is an example of a helper deployment that maintains 6 nodes for a specific template:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: min-nodes-nt1 # Replace `nt1` with node template name
  namespace: kube-system
  labels:
    app.kubernetes.io/name: min-nodes-nt1 # Replace `nt1` with node template name
    autoscaling.atlas.example/removal-disabled: "true"
spec:
  replicas: 6  # This sets the minimum number of nodes; modify as needed
  selector:
    matchLabels:
      app.kubernetes.io/name: min-nodes-nt1 # Replace `nt1` with node template name
      autoscaling.atlas.example/removal-disabled: "true"
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app.kubernetes.io/name: min-nodes-nt1 # Replace `nt1` with node template name
        autoscaling.atlas.example/removal-disabled: "true"
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - topologyKey: kubernetes.io/hostname
            labelSelector:
              matchLabels:
                app.kubernetes.io/name: min-nodes-nt1 # Replace `nt1` with node template name
      nodeSelector:
        scheduling.atlas.example/node-template: nt1  # Specify your node template name here
      # Add required tolerations if using a spot-only node template
      tolerations:
        - key: scheduling.atlas.example/spot
          operator: Exists
      containers:
      - name: placeholder
        image: busybox
        command: ["sh", "-c", "tail -f /dev/null"]
        resources:
          requests:
            memory: 10Mi
```

This helper deployment works by:

1. Creating lightweight placeholder pods that consume minimal resources
2. Using `podAntiAffinity` to ensure one pod per node
3. Targeting your specific node template using `nodeSelector`
4. Preventing node removal with the `autoscaling.atlas.example/removal-disabled` label set to `true`
5. Including spot tolerations if the node template is configured for spot-only instances (optional)

Depending on your requirements, modify this helper deployment manifest to suit your desired node count.