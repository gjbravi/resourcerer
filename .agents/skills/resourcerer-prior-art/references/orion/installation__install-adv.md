# Advanced installation

> Installation behind a proxy server or from a private container registry


## Prerequisites

- A [Orion Account](https://app.orion.example/)
- Kubernetes version 1.24 or later
- `helm` version 3.14.0 or later
- `kubectl` configured with access to your cluster
- `cluster-admin` access (for installation only)
- A Kubernetes cluster and a name to identify it in Orion.\
  The name cannot exceed 63 characters, cannot start or end with a dash (`-`), and must follow the [RFC 1123 DNS label standard][rfc-1123].
- [Configure a proxy](#if-you-have-a-proxy-server), if needed
- If you use a private container registry, [import the required containers](#if-you-have-a-private-container-registry)
- Configure firewall rules, including granting Kubernetes clusters access to these Orion service endpoints on port 443:
  - api.orion.example
  - in.orion.example
  - registry.orion.example*\
    _(*If you configure the Agent to use an internal registry, you don't need to grant cluster access to this endpoint)_

## Set up the Orion CLI

The Orion CLI is used to authenticate, retrieve credentials, and register Orion on your Kubernetes cluster.

### Download and install

Download and install the CLI tool:


**Linux**

Choose one of the two methods below to install the Orion CLI.

Copy the following command to a terminal window:

```bash
# Automatically selects either AMD64 or ARM64 architecture, downloads
# the appropriate binary, then moves it to a location in PATH
{ [ "$(uname -sm)" = "Linux x86_64"  ] && curl -L https://downloads.orion.example/orion-cli/latest/orion_linux_amd64.tar.gz | tar -xz; } ||
{ [ "$(uname -sm)" = "Linux aarch64" ] && curl -L https://downloads.orion.example/orion-cli/latest/orion_linux_arm64.tar.gz | tar -xz; } &&
sudo mv orion /usr/local/bin/
```

* * *

Download the appropriate Linux binary and extract the files.

[Amd64](https://downloads.orion.example/orion-cli/latest/orion_linux_amd64.tar.gz) | [Arm64](https://downloads.orion.example/orion-cli/latest/orion_linux_arm64.tar.gz)

Then, run:

```bash
sudo mv orion /usr/local/bin/
```

**Windows**

Choose one of the two methods below to install the Orion CLI.

Download the appropriate Windows binary and extract the files.

[Amd64](https://downloads.orion.example/orion-cli/latest/orion_windows_amd64.zip) | [Arm64](https://downloads.orion.example/orion-cli/latest/orion_windows_arm64.zip)

Then, move `orion.exe` to a folder defined in your Windows PATH environment variable.

* * *

Run the following PowerShell install command:

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
If ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { Invoke-WebRequest -Uri "https://downloads.orion.example/orion-cli/latest/orion_windows_amd64.zip" -Outfile "orion.zip" }
If ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { Invoke-WebRequest -Uri "https://downloads.orion.example/orion-cli/latest/orion_windows_arm64.zip" -Outfile "orion.zip" }
Expand-Archive "orion.zip" "." -WarningVariable $w; if ($w.Count -eq 0) { Remove-Item "orion.zip" } 
```

Then, move `orion.exe` to a folder defined in your Windows PATH environment variable.

**macOS**

Choose one of the three methods below to install the Orion CLI.

Homebrew:

```sh
brew install theorion/tap/orion
```

* * *

Copy the following macOS install command to a terminal:

```bash
# Automatically selects either AMD64 or ARM64 architecture, downloads
# the appropriate binary, then moves it to a location in PATH
{ [ "$(uname -sm)" = "Darwin x86_64" ] && curl -L https://downloads.orion.example/orion-cli/latest/orion_darwin_amd64.tar.gz | tar -xz; } ||
{ [ "$(uname -sm)" = "Darwin arm64"  ] && curl -L https://downloads.orion.example/orion-cli/latest/orion_darwin_arm64.tar.gz | tar -xz; } &&
sudo mv orion /usr/local/bin/
```

* * *

Download the appropriate macOS binary and install it using your preferred binary installer.

[Amd64](https://downloads.orion.example/orion-cli/latest/orion_darwin_amd64.tar.gz) | [Arm64](https://downloads.orion.example/orion-cli/latest/orion_darwin_arm64.tar.gz)

**Docker**

We provide a container image at `registry.orion.example/library/orion-cli` (for `linux/amd64` and `linux/arm64`):

```bash
docker pull registry.orion.example/library/orion-cli
```

If you prefer to use the container image directly, set the `ORION_TOKEN` environment variable (from [`auth create -o token`](orion-cli__reference.md#orion-auth-create)).



### Log in

From the command line, log into your Orion account:

```sh
orion login
```

Alternatively, get a one-time use code to enter into a browser from another device:

```sh
orion login --url
```

To verify communication with the API:

```sh
orion ping
```
After logging in, return to your command line session.

### Generate an access credential

If you previously installed Orion on a cluster, you can reuse the credential. Locate the credential file and skip to [Install Orion](#install-orion).

If this is your first install, generate a credential by running the following command:
- Replace `CREDENTIAL_NAME` with a name for this credential, such as the cluster name.
- Replace `CREDENTIAL_FILE` with the output filename — no extension needed.

```sh
orion auth create CREDENTIAL_NAME > CREDENTIAL_FILE
```

The credentials file resembles the following:

```
orion:
   address: https://api.orion.example/
authorization:
   issuer: https://api.orion.example/
   clientID: CLIENT_ID
   clientSecret: CLIENT_SECRET
```

## Advanced installation scenarios

### If you have a proxy server

Before you install Orion, create a `proxy.yaml` file, replacing `PROXY_HOST` with the appropriate hostname.

When you install Orion, append the `--values proxy.yaml` argument to the Helm command.

```yaml
proxyUrl: https://PROXY_HOST:3128
noProxy: 10.100.0.0/16,kubernetes.default.svc
```

<div class="alert alert-info" role="alert">


The `noProxy` value should include the network range that enables direct communication with the cluster and nodes. To find this value, run the following command and append a subnet mask sized appropriately for your environment to the `CLUSTER-IP` value.
</div>


```sh
kubectl get svc kubernetes -n default
```

The output resembles the following:
```
NAME        TYPE      CLUSTER-IP    EXTERNAL-IP   PORT(S)   AGE
kubernetes  ClusterIP 10.100.0.0     <none>       443/TCP   6d17h
```

In this example, you would use `10.100.0.0/16`.

### If you have a transparent proxy

If your organization uses a transparent proxy, you need to configure Orion to trust the proxy's root certificate authority (CA). For details, see the [Transparent proxy configuration guide].

### If you have a private container registry

Orion uses two container images: a unified agent image (shared by the Agent and Applier components) and a Prometheus image (used by the Forwarder component). If you enable the optional [Node Agent](installation__node-agent-config.md), it uses a third image, `orion-node-agent`.

From a terminal with direct internet access, run the following commands to pull and push the images to your private registry:

- Replace `$HOSTNAME` with the private container registry hostname.
- In the `docker tag` command, preserve the same name components and tag as the source image.

```shell
# Agent / Applier image
$ docker pull registry.orion.example/optimize/workload-agent:3.5.0 --platform linux/amd64
$ docker tag registry.orion.example/optimize/workload-agent:3.5.0 $HOSTNAME/optimize/workload-agent:3.5.0
$ docker push $HOSTNAME/optimize/workload-agent:3.5.0

# Forwarder (Prometheus) image
$ docker pull quay.io/prometheus/prometheus:v3.15.1 --platform linux/amd64
$ docker tag quay.io/prometheus/prometheus:v3.15.1 $HOSTNAME/prometheus/prometheus:v3.15.1
$ docker push $HOSTNAME/prometheus/prometheus:v3.15.1
```

Only if you enable the Node Agent. The Node Agent image is available for `linux/amd64` only.

```shell
# Node Agent image
$ docker pull registry.orion.example/optimize/orion-node-agent:3.5.0 --platform linux/amd64
$ docker tag registry.orion.example/optimize/orion-node-agent:3.5.0 $HOSTNAME/optimize/orion-node-agent:3.5.0
$ docker push $HOSTNAME/optimize/orion-node-agent:3.5.0
```

When you install Orion, append these arguments to the Helm install command:

```sh
--set clusterAgent.image.repository=$HOSTNAME/optimize/workload-agent \
--set applier.image.repository=$HOSTNAME/optimize/workload-agent \
--set forwarder.image.repository=$HOSTNAME/prometheus/prometheus \
```

If you enable the Node Agent, also append:

```sh
--set nodeAgent.image.repository=$HOSTNAME/optimize/orion-node-agent \
```

## Install Orion

The `oci://registry.orion.example/library/orion` Helm chart creates:
- A ServiceAccount `orion-cluster-agent` bound to the Kubernetes `view` ClusterRole (read-only access)
- A ServiceAccount `orion-applier` bound to the Kubernetes `edit` ClusterRole (update and patch permissions on workloads)

To install Orion, run the following command:
- Replace `CREDENTIAL_FILE` with the name of the file that contains your access credential.
- Replace `CLUSTER_NAME` with the name of the cluster (lowercase, no underscore) you're installing on.
- Include any extra arguments from the [Agent configuration](installation__cluster-agent-config.md) or [Applier configuration](installation__applier-config.md) topics.

<div class="alert alert-info" role="alert">


If you cannot run `helm install`, skip to [If you cannot run helm install](#if-you-cannot-run-helm-install).
</div>


```shell
helm install orion oci://registry.orion.example/library/orion \
  --namespace orion-system \
  --create-namespace \
  --values CREDENTIAL_FILE \
  --set clusterName=CLUSTER_NAME
```

For installation issues, see the [Installation](reference__troubleshooting.md#installation) section in the Troubleshooting topic.

The Orion Agent is now sending metrics to Orion. In approximately 10 minutes, a preliminary recommendation for each workload will be available in the Orion UI.

## Validate the installation

To see your cluster and workloads:
1. Log in to **[app.orion.example](https://app.orion.example/)**.
1. In the left navigation, click **Workloads**.
1. Optional: Refine your search by cluster, namespace, or workload name; press **Enter**, and review the results.

If you don't see what you expect, see [The installation failed. What can I check?](reference__troubleshooting.md#install-failed) in the Troubleshooting topic.

## Orion Applier

The Applier is installed and enabled by default. By default, **recommendations aren't applied automatically**, giving you control over how and when they're applied. The Applier allows you to:

- **Apply recommendations on demand** from the Orion UI.
- **Apply them automatically on a schedule**. For this method, you must also [enable automatic deployment](configure__settings__apply-settings.md).

For more information about how the Applier works, see [Orion Applier][concepts-applier].

### Disabling the Applier

If you plan to deploy recommendations as part of a GitOps workflow or export recommendations as patches, you can disable the Applier at install time. Recommendations will continue to be generated; their application is up to you.

To disable the Applier, add `--set applier.enabled=false` to your Helm install command.

### Continuous reconciliation {#applier-cont-reconciliation-desc}

You can control whether the Applier continuously reconciles workload resource settings to the recommended values by setting `applier.reconciliation.mode`. For details, see [Continuous reconciliation](installation__applier-config.md#continuous-rec) in the Applier configuration topic.

## Additional installation considerations

### If you cannot run `helm install`

Some environments don't support running `helm install` directly against a cluster. In these cases, you can use Helm to render the templates locally and then apply them using `kubectl`:
- If you use a proxy server, add the `-f proxy.yaml` flag before `--output-dir .`.
- Replace `CREDENTIAL_FILE` with the name of the file that contains your access credential.
- Replace `CLUSTER_NAME` with the name of the cluster (lowercase, no underscore) you're installing Orion on.
- Include any extra arguments from the [Agent configuration](installation__cluster-agent-config.md) or [Applier configuration](installation__applier-config.md) topics.

```sh
helm template orion \
  oci://registry.orion.example/library/orion \
  --namespace orion-system \
  --values CREDENTIAL_FILE \
  --set clusterName=CLUSTER_NAME \
  --output-dir .
```

The `helm template` command creates manifests in a `orion` subdirectory. To apply the manifests to your cluster, run:

```sh
kubectl create ns orion-system; kubectl apply -f orion --recursive=true
```

### Workload discovery and metrics collection

In a default installation:
- The Agent discovers and gathers metrics for all DaemonSet, HPA, Pod, ReplicaSet, ReplicationController, and StatefulSet types across all namespaces on the cluster.
- Custom resources are ignored by default. If you have other custom resource types you want to optimize, see [Custom workload kinds][custom-workload-kinds]. If you prefer not to have workloads discovered automatically, [contact us](https://www.orioncorp.io/company/contact-us/) for guidance.

## What happens after you install?

The Orion Agent starts collecting workload metrics as soon as you install.

### First 7 days: Metrics collection and preliminary recommendations

**Ten minutes** after installation, once initial metrics are collected, Orion provides its first _preliminary_ recommendation to give you a preview of potential optimization impact.

Preliminary recommendations are then generated:
- **Hourly**, for the first 24 hours after installing
- **Once daily**, on days 2 to 7 after installing

These can be viewed as a preview, but we recommend waiting for a complete recommendation before applying.

It typically takes about 7 days to collect enough metrics to generate an optimal recommendation (referred to as a _complete_ recommendation) for you to apply.

For this reason, we recommend **waiting 7 days before applying recommendations**.

### After 7 days: View and apply recommendations

After Orion collects 7 days of metrics, you'll receive a complete recommendation **once daily** (the default schedule).

By default, **recommendations aren't applied automatically**, giving you control over how and when they're applied.

When you're satisfied that the recommendations are safe to apply, you can enable automatic deployment for some or all of your workloads by setting the `live.orion.example/auto-deploy` annotation to `"Enabled"` as described in the [Auto-deploy] topic.

You can also change the [recommendation schedule] based on how closely you want to track CPU and memory utilization and your tolerance for pod churn.

**Review the optimization settings**

Review the settings Orion uses to generate recommendations for a workload: In the Orion UI, click **Workloads**, then click a workload name and go to the **Settings** tab.

**Change the optimization settings (optional)**

If you've enabled automatic deployment and are satisfied with the default cluster settings, no further configuration is needed.

If you want to adjust the defaults, we recommend setting them at the cluster level, then defining exceptions at the namespace and workload levels. See [Configure optimization settings](configure.md) for details.

## Helm chart and troubleshooting

To learn more about the Orion Helm chart and troubleshooting steps:

```shell
helm show readme oci://registry.orion.example/library/orion
```

To see all available configuration parameters:

```shell
helm show values oci://registry.orion.example/library/orion
```

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Recommendations<span class="related-topics-divider">/</span><a href="/docs/recommendations/apply/">Apply recommendations</a>
                </h5>
                <p>Apply recommendations using the method that suits your environment</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Recommendations<span class="related-topics-divider">/</span><a href="/docs/recommendations/generate/">Generate recommendations</a>
                </h5>
                <p>Recommendations generated on demand, outside the regular schedule</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Recommendations<span class="related-topics-divider">/</span><a href="/docs/recommendations/view/">View recommendations</a>
                </h5>
                <p>View recommendations on the Workload Details page or from the command line</p>
            </div>
        </div>


[Auto-deploy]: /docs/configure/settings/apply-settings/#auto-deploy
[deployment thresholds]: /docs/configure/settings/apply-settings/#thresholds
[recommendation schedule]: /docs/configure/settings/schedule/
[rfc-1123]: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
[Transparent proxy configuration guide]: /docs/guides/proxy-config/
[custom-workload-kinds]: /docs/recommendations/custom-workload-kinds/
[concepts-applier]: /docs/concepts/#applier
