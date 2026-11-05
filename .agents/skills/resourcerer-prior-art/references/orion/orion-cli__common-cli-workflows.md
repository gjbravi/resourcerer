# Common Orion CLI workflows

> Common Orion CLI tasks, from authentication to managing recommendations


The Orion CLI lets you interact with Orion to inspect workloads, view optimization data, and manage recommendations.

This guide covers common workflows:

- Authenticate and configure the CLI
- Explore clusters and workloads
- Find workloads with auto-deploy enabled
- View and generate recommendations
- Query recommendation data
- Apply recommendations to workloads

Typical workflow:

```
Identify workload → Inspect recommendation → Apply recommendation
```

Common workflow commands:

```sh
orion get workloads --cluster <cluster>
orion get recommendations --cluster <cluster> --namespace <namespace> --name <workload>
orion apply --cluster <cluster> --namespace <namespace> --resource <resource-type> --name <workload>
```

---

## In this guide

- [Common commands](#common-commands)
- [Get started with the CLI](#get-started)
- [Explore clusters and workloads](#explore)
- [Find workloads with auto-deploy enabled](#find-auto-deploy)
- [View and generate recommendations](#view-recs)
- [Query recommendation data](#query-recs)
- [Apply recommendations](#apply-recs)

## Common commands {#common-commands}

| Task                  | Command                                                     |
| --------------------- | ----------------------------------------------------------- |
| Authenticate          | [`orion login`][orion-login]                      |
| List clusters         | [`orion get clusters`][orion-clusters]            |
| List workloads        | [`orion get workloads`][orion-get-workloads]      |
| View workloads        | [`orion view workloads`][orion-view-workloads]    |
| List recommendations  | [`orion get recommendations`][orion-get-recs]     |
| Create recommendation | [`orion create recommendation`][orion-create-rec] |
| Apply recommendations | [`orion apply`][orion-apply]                      |

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Using command help in the CLI</div>


Use the `--help` flag with any command for more information.
</div>


## Get started with the CLI {#get-started}

Authenticate and configure the Orion CLI.

### Authenticate

Command:

```sh
orion login
```

Example output:

```sh
You are now logged in.
```

Opens a browser window and authenticates using your Orion account.

### Get the CLI version

Command:

```sh
orion version
```

Example output:

```sh
orion version 6.1.1
```

Displays the installed Orion CLI version.

### Set up autocompletion

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Enabling CLI autocompletion</div>


If you installed the Orion CLI via `brew`, autocompletion is enabled by default.
</div>


Command:

```sh
orion completion [bash|zsh|fish|powershell] --help
```

Follow the instructions for your shell to enable command autocompletion.

Once configured, you can tab-complete Orion CLI commands.

## Explore clusters and workloads {#explore}

Retrieve cluster and workload information.

### List clusters

Command:

```sh
orion get clusters -o wide
```

Example output:

```sh
NAME        AGENT VERSION   APPLIER VERSION   KUBERNETES VERSION   STATE     AGE
<cluster>   2.27.2          2.12.0            1.34.0               Created   3 hours
```

The `wide` output option includes additional cluster metadata, such as the Orion Agent and Kubernetes versions.

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Available output formats</div>


Multiple output formats are supported, including `json` and `yaml`. Different output formats return varying levels of detail.
</div>


### View cluster details

Command:

```sh
orion get clusters <cluster> -o yaml
```

Example output:

```yaml
_metadata: ...
name: <cluster>
orion:
  products:
  - name: orion-agent
    version: 2.27.2
  - name: orion-applier
    version: 2.12.0
kubernetes:
  version: 1.34.0
nodes: ...
status: ...
```

The `yaml` output includes node details, Kubernetes version, and Orion cluster metadata.

### List workloads in a cluster

Command:

```sh
orion get workloads --cluster <cluster> -o wide
```

Example output:

```sh
CLUSTER     NAMESPACE     TYPE              NAME           AUTO DEPLOY   HPA ENABLED   GPU ENABLED   STATE     AGE
<cluster>   <namespace>   <resource-type>   <workload-1>   false         false         false         Created   3 hours
<cluster>   <namespace>   <resource-type>   <workload-2>   false         false         false         Created   3 hours
```

The `wide` output includes the `AUTO DEPLOY`, `HPA ENABLED`, and `GPU ENABLED` status columns for each workload.

<div class="alert alert-warning" role="alert"><div class="h4 alert-heading" role="heading">Cluster requirement for listing workloads</div>


Running `orion get workloads` without specifying a cluster returns an error.
</div>


### View detailed workload information

Command:

```sh
orion get workloads \
  --cluster <cluster> \
  --namespace <namespace> \
  --resource <resource-type> \
  --name <workload> \
  -o yaml
```

Example output:

```yaml
cluster: <cluster>
resource: <resource-type>
namespace: <namespace>
name: <workload>
labels: ...
annotations: ...
learningPeriod: P1D
recommendationHandlers: ...
patchTarget: ...
pods: ...
status: ...
```

Specifying `--cluster`, `--namespace`, `--resource`, and `--name` ensures the expected workload is returned. Workloads with the same name can exist across multiple resource types, namespaces, or clusters.

### View under- or over-provisioned workloads

Command:

```sh
orion view workloads --cluster <cluster> --over-provisioned
```

Example output:

```sh
WORKLOAD NAME   CLUSTER     NAMESPACE     TYPE              STATUS   OOM RATE   SCORE   NET IMPACT (CPU/MEMORY)
<workload>      <cluster>   <namespace>   <resource-type>   Ready    0          13      -$4.39/mo. (-0.180/-76MiB)
```

The `orion view` command returns recommendations ordered by impact. Filter using the `--over-provisioned` and `--under-provisioned` flags.

### View GPU workloads

Command:

```sh
orion view workloads --cluster <cluster> --gpu
```

The `--gpu` flag returns workloads that Orion identifies as using GPUs. It requires Orion CLI version 6.2.8 or later.

### Sort workloads by auto-deploy status

Command:

```sh
orion get workloads \
  --cluster <cluster> \
  --sort-by auto_deploy \
  -o wide
```

Example output:

```sh
CLUSTER     NAMESPACE       TYPE              NAME           AUTO DEPLOY   HPA ENABLED   GPU ENABLED   STATE     AGE
<cluster>   <namespace-1>   <resource-type>   <workload-1>   false         true          false         Created   3 weeks
<cluster>   <namespace-1>   <resource-type>   <workload-2>   false         false         false         Created   3 weeks
<cluster>   <namespace-2>   <resource-type>   <workload-3>   true          false         false         Created   3 weeks
```

Sorting by `auto_deploy` groups workloads by their auto-deploy status.

## Find workloads with auto-deploy enabled {#find-auto-deploy}

Identify workloads with auto-deploy enabled across your environment.

### List auto-deploy-enabled workloads in a cluster

Command:

```sh
orion view workloads --cluster <cluster> --auto-deploy
```

Example output:

```sh
WORKLOAD NAME   CLUSTER     NAMESPACE     TYPE              STATUS   OOM RATE   SCORE   NET IMPACT (CPU/MEMORY)
<workload-1>    <cluster>   <namespace>   <resource-type>   Ready    0          91      -$2.10/mo. (-0.095/-48MiB)
<workload-2>    <cluster>   <namespace>   <resource-type>   Ready    0          78      -$0.84/mo. (-0.040/-20MiB)
```

The `--auto-deploy` flag returns workloads configured to have recommendations applied automatically.

### Filter by namespace

Command:

```sh
orion view workloads \
  --cluster <cluster> \
  --namespace <namespace> \
  --auto-deploy
```

Combines `--namespace` and `--auto-deploy` flags to return results for a single namespace in a cluster.

### Filter with field-selector when piping to jq

Use `--field-selector auto_deploy=true` with `orion get workloads` to reduce the JSON blob before passing it to `jq` or other tools. The `--field-selector` flag filters on any sortable column (see `--sort-by` in the [CLI reference][orion-get-workloads]).

Command:

```sh
orion get workloads \
  --cluster <cluster> \
  --field-selector auto_deploy=true \
  -o json | jq '.results[].name'
```

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Client-side filtering</div>


When using `--field-selector`, filtering happens after the CLI fetches all workloads. For server-side filtering, use `orion view workloads --auto-deploy`.
</div>


## View and generate recommendations {#view-recs}

Inspect recommendation status and generate recommendations on demand.

### View recommendation status

Command:

```sh
orion get recommendations --cluster <cluster> --name <workload> -o wide
```

Example output:

```sh
CLUSTER      NAMESPACE    TYPE              NAME         TIMESTAMP    SCORE   NET IMPACT (CPU/MEMORY)    STATE       STATUS
<cluster>   <namespace>   <resource-type>   <workload>   1772819220   84      $1.80/mo. (0.087/-64MiB)   Generated   Learning
```

Returns recommendation data, including resource usage, requests and limits, and the net impact of applying a recommendation.

The net impact displays as a dollar amount determined by the configured cost per hour for CPU and RAM.

This generated recommendation is for a workload in the [learning period][learning]. The [optimization score][impact] of 84 indicates that the workload is close to optimal size.

### View recommendation details

Command:

```sh
orion get recommendations --cluster <cluster> --name <workload> -o yaml
```

<details>
<summary>View the example output</summary>

```yaml
name: "1772819220"
workload:
  cluster: <cluster>
  namespace: <namespace>
  name: <workload>
timestamp: "2026-03-06T17:47:00Z"
pods:
- observed:
    metrics:
      containerResources:
      - containerName: <container>
        usage:
          cpu: 0.11
          memory: 4012426
        requests:
          cpu: 0.5
          memory: 134217728
    values:
      containerResources:
      - containerName: <container>
        requests:
          cpu: 500m
          memory: 128Mi
        limits:
          cpu: 1
          memory: 128Mi
  recommended:
    values:
      containerResources:
      - containerName: <container>
        requests:
          cpu: 577m
          memory: 32Mi
        limits:
          cpu: 1154m
          memory: 128Mi
```

</details><br/>

The `yaml` output provides additional information, including observed usage, current requests and limits, and recommended requests and limits.

Focus on the `observed` and `recommended` sections to compare existing and optimized resources.

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Recommendations reflect point-in-time data</div>


A recommendation contains data from the time it was generated. Running `orion get recommendations` doesn't show the current workload usage.
</div>


### Locate recommendations that failed to generate

Command:

```sh
orion get recommendations --cluster <cluster> --state FailedToGenerate
```

Example output:

```sh
CLUSTER     NAMESPACE     TYPE              NAME         TIMESTAMP    SCORE   NET IMPACT (CPU/MEMORY)   STATE
<cluster>   <namespace>   <resource-type>   <workload>   1772794020           -- (--/--)                FailedToGenerate
```

Orion can fail to generate a recommendation if, for example, a workload lacks recent usage data. Use the `--state` flag to list recommendations in the `FailedToGenerate` state.

In this example, Orion was unable to generate a recommendation due to insufficient recent usage data.

### Generate a recommendation

Command:

```sh
orion create recommendation --cluster <cluster> --namespace <namespace> --resource <resource-type> --name <workload> --wait
```

Example output:

```sh
recommendation "<cluster>/<namespace>/<resource-type>/<workload>/<recommendation>" created.
```

You can generate a recommendation on demand using the CLI ([or the UI][ui]), rather than waiting for the next scheduled run.

Using the `--wait` flag forces the CLI to wait until the recommendation has been generated before returning a result.

## Query recommendation data {#query-recs}

Use the [jq] tool with JSON-formatted output to extract and analyze specific recommendation fields.

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Recommendations reflect point-in-time data</div>


Any `observed` data reflects usage at the time the recommendation was generated.
</div>


### Show observed usage for each container in a workload

Command:

```sh
orion get recommendations <cluster>/<namespace>/<resource-type>/<workload> -o json | jq '.results[].pods[].observed.metrics'
```

<details>
<summary>View the example output</summary>

```json
{
  "containerResources": [
    {
      "containerName": "<container-1>",
      "usage": {
        "cpu": 0,
        "memory": 124995
      }
    },
    {
      "containerName": "<container-2>",
      "usage": {
        "cpu": 0.10,
        "memory": 4209629
      },
      "requests": {
        "cpu": 0.5,
        "memory": 134217728
      }
    }
  ]
}
```

</details><br/>

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Filtering CLI output with paths and flags</div>


The Orion CLI supports filtering by providing a path, such as `<cluster>/<namespace>/<resource-type>/<workload>` in this example, as well as flags such as `--cluster <cluster> --namespace <namespace> --resource <resource-type> --name <workload>`.
</div>


### Show current request and limit values

Command:

```sh
orion get recommendations <cluster>/<namespace>/<resource-type>/<workload> -o json | jq '.results[].pods[].observed.values'
```

Example output:

```json
{
  "containerResources": [
    {
      "containerName": "<container>",
      "requests": {
        "cpu": "500m",
        "memory": "128Mi"
      },
      "limits": {
        "cpu": "1",
        "memory": "128Mi"
      }
    }
  ]
}
```

### Show recommended request and limit values

Command:

```sh
orion get recommendations <cluster>/<namespace>/<resource-type>/<workload> -o json | jq '.results[].pods[].recommended'
```

<details>
<summary>View the example output</summary>

```json
{
  "values": {
    "containerResources": [
      {
        "containerName": "<container>",
        "requests": {
          "cpu": "577m",
          "memory": "32Mi"
        },
        "limits": {
          "cpu": "1154m",
          "memory": "128Mi"
        }
      }
    ],
    "scheduling": {
      "categories": [
        "compute-optimized"
      ]
    }
  }
}
```

</details><br/>

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Scheduling and node affinity</div>


This example includes the `compute-optimized` scheduling category. See the [Pod scheduling documentation][scheduling] for details on node affinity.
</div>


### Show recommendations for all optimization goals

Command:

```sh
orion get recommendations <cluster>/<namespace>/<resource-type>/<workload> -o json | jq '.results[].pods[].effective'
```

<details>
<summary>View the example output</summary>

```json
[
  {
    "goal": "savings",
    "values": {
      "containerResources": [
        {
          "containerName": "<container>",
          "requests": {
            "cpu": "577m",
            "memory": "32Mi"
          },
          "limits": {
            "cpu": "1154m",
            "memory": "128Mi"
          }
        }
      ]
    }
  },
  {
    "goal": "balanced",
    "values": {
      "containerResources": [
        {
          "containerName": "<container>",
          "requests": {
            "cpu": "577m",
            "memory": "32Mi"
          },
          "limits": {
            "cpu": "1154m",
            "memory": "128Mi"
          }
        }
      ]
    }
  },
  {
    "goal": "reliability",
    "values": {
      "containerResources": [
        {
          "containerName": "<container>",
          "requests": {
            "cpu": "577m",
            "memory": "32Mi"
          },
          "limits": {
            "cpu": "1154m",
            "memory": "128Mi"
          }
        }
      ]
    }
  }
]
```

</details>
<br/>

A resource's [optimization goal][goal] indicates which of two competing criteria to maximize when computing recommended request values: cost savings or workload reliability.

Orion **always** produces values that satisfy both criteria, but goal-setting can shift the balance toward one or the other.

The available goals are `Balanced` (default), `Savings`, and `Reliability`. Per-container values display recommendations for all goals.

This can provide a helpful comparison to see what Orion would recommend for a different goal.

## Apply recommendations {#apply-recs}

The `orion create` and `orion apply` commands can both apply recommendations.

### Create and apply

Command:

```sh
orion create recommendation \
  --namespace <namespace> \
  --resource <resource-type> \
  --name <workload> \
  --auto-deploy \
  --wait
```

Example output:

```sh
recommendation "<cluster-1>/<namespace>/<resource-type>/<workload>/<recommendation>" applied.
recommendation "<cluster-2>/<namespace>/<resource-type>/<workload>/<recommendation>" applied.
```

Use the `orion create` command with the `--auto-deploy` flag to create and automatically apply a recommendation.

In this example, a cluster name isn't supplied. The recommendation is applied to matching workloads across all clusters.

### Apply the latest recommendation

First, check for recommendations.

Command:

```sh
orion get recommendations --cluster <cluster> --namespace <namespace> --state Generated
```

Then, apply the latest recommendation.

Command:

```sh
orion apply --cluster <cluster> --namespace <namespace> --resource <resource-type> --name <workload>
```

Example output:

```sh
recommendation "<cluster>/<namespace>/<resource-type>/<workload>/<recommendation>" applied.
```

Applies the most recently generated recommendation to the specified workload.

### Create with a validity period and apply

Command:

```sh
orion create recommendation \
  --cluster <cluster> \
  --namespace <namespace> \
  --resource <resource-type> \
  --name <workload> \
  --valid-for P7D \
  --auto-deploy \
  --wait
```

Example output:

```sh
recommendation "<cluster>/<namespace>/<resource-type>/<workload>/<recommendation>" applied.
```

Creates and automatically deploys a [recommendation that remains valid][validity] for 7 days.

## Next steps {#next-steps}

- View the full [Orion CLI command reference][orion-cli]

[goal]: /docs/configure/settings/goal/
[impact]: /docs/recommendations/view/#viewing-the-overall-projected-impact-of-a-recommendation
[jq]: https://jqlang.org/
[learning]: /docs/concepts/#learning-period
[scheduling]: /docs/recommendations/pod-scheduling/
[orion-apply]: /docs/orion-cli/reference/#orion-apply
[orion-cli]: /docs/orion-cli/reference/
[orion-clusters]: /docs/orion-cli/reference/#orion-get-clusters
[orion-login]: /docs/orion-cli/reference/#orion-login
[orion-create-rec]: /docs/orion-cli/reference/#orion-create-recommendation
[orion-get-recs]: /docs/orion-cli/reference/#orion-get-recommendations
[orion-get-workloads]: /docs/orion-cli/reference/#orion-get-workloads
[orion-view-workloads]: /docs/orion-cli/reference/#orion-view-workloads
[ui]: /docs/recommendations/generate/
[validity]: /docs/concepts/#validity-period
