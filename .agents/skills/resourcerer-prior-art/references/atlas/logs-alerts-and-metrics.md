---
updatedAt: 2026-09-25T10:31:43.000Z
---

# Logs, alerts, and metrics

### What is the log level number for every severity ?

1: Fatal

2: Error

3: Warn

4: Info

5: Debug

<br />

***

### Is there a metric in Atlas I can use to check the number of empty node in the cluster?

Atlas deletes empty nodes, and all events are logged in the [cluster audit log](audit-log.md) (**Automation** > your cluster > **Logs**). You can pull them with an API if needed.

ListAuditEntries returns audit entries for a given cluster; see the [ListAuditEntries API reference](https://docs.atlas.example/reference/auditapi_listauditentries).

<br />

***

### I added a panel with compute hourly costs to our Grafana board but I'm not sure how it should be interpreted?

These prices are the hourly expenses. To get a typical monthly cost, you need to multiply this number by 730.

<br />

***

### I see the following log in the atlas-cluster-controller, can you explain what it means

Log:

```
{ message:I0524 06:00:28.976409 1 request.go:682] Waited for 1.037042311s due to client-side throttling, not priority and fairness, request: GET:https://172.20.0.1:443/apis/vpcresources.k8s.aws/v1beta1 }
```

request.go is part of the Kubernetes REST API package. Kubernetes uses a discovery cache to populate the names of all API resources in the cluster, and this generates potentially 100s of GET requests behind the scenes. This log is most often seen concerning this behavior. The throttling is used to prevent kubelet/kube-apiserver from being overloaded with requests and degrading performance.

This [pull request](https://github.com/kubernetes/kubernetes/pull/101634) explains the issue in detail.

<br />

***

### When I run Terraform, I get an error saying that role\_arn is an unsupported attribute?

Log:

```
Error: Unsupported attribute on atlas.tf line 18, in module "atlas-eks-cluster-use1-1":   18:   aws_assume_role_arn      = module.atlas-aws-iam-use1-1.role_arn  This object does not have an attribute named "role_arn".
```

role\_arn should work with the module "atlas-eks-iam-role". Similar to the usage in our examples:

```
aws_assume_role_arn = module.atlas-eks-role-iam.role_arn
```

Learn more: [Unsupported node affinity rules](https://github.com/atlas/terraform-provider-atlas/blob/master/examples/eks/eks_cluster_assumerole/atlas.tf)

<br />

***

### The pod wasn't scheduled due to nodeSelector, why does the log say it failed on node affinity?

Log:

```
[autoscaler.atlas.example  Unsupported node affinity rules]
```

Node affinity and node selectors are essentially the same concept, just that affinity allows for conditional matching while selectors are a strict match.

<br />

***

### If I apply Pod Disruption Budgets in a deployment, does Atlas generate a log if the policy is not met and it cannot drain the node?

In this scenario, the drain would fail, and Atlas would revert the node to active.

To prevent this issue, we recommend not using percentages for unavailable, but rather static numbers based on the min/max size of the deployment. For instance, with 10 replicas minimum having 3 unavailable would be acceptable. But in the case of having 2 or 3, 1 unavailable might be all it can tolerate.

<br />

***

### I’m trying to add a Grafana panel that will calculate a number of Atlas-managed spot instances over a period of time. Do those spot instances get terminated or does Atlas drain and gracefully shut them down in advance?

Typically, Atlas will drain and shut them down before the hard termination from AWS so long as the pods can move relatively quickly.

You can [scrape our metrics](https://docs.atlas.example/docs/metrics) which will show the Atlas spot nodes and other types of nodes.

<br />

***

### I am seeing a status of "WARNING" for my cluster.  How can I determine the reason behind that warning?

Usually, a warning means, "Atlas-managed cluster has encountered a transient error and is currently attempting to recover from it automatically. Autoscaling is not working."

Our system attempts an automatic recovery and only sends notifications if the recovery fails, resulting in a "Failed" status.

Learn more: [Cluster status](cluster-node-status.md#overview-of-cluster-status-values)

<br />

***

### Is it possible to send alert notification to email with Atlas?

Currently, we only support notifications via the Atlas console as a webhook for PagerDuty and Slack.

<br />

***

### Does Atlas rely in any shape or form on Datadog for scaling etc?

Atlas doesn't rely on Datadog for scaling.

<br />

***

### Sometimes, I have unscheduled pods. Is there a way to see those pods? What should I change in my Unscheduled pods policy to prevent that?

We recommend using `kubectl` to find the unscheduled pod and running `kubectl describe` to check the events for insight into why it's unscheduled. To see the unscheduled pods in your cluster, you can use the following command:

```
kubectl get pods --field-selector=status.phase=Pending
```

<br />

***

### Does Atlas take out pod spec and if so, which of its parts?

Atlas takes all pod configuration YAML content, but we remove env variables to avoid sending passwords/keys, etc. The Atlas agent running inside your cluster is in charge of doing this.

NodeSelectors, Affinity, Labels, CPU/Memory requests/limits, and PVC mounts - we use all that information either for reporting or making decisions, such as which AZ to create a node so that the pod gets scheduled.