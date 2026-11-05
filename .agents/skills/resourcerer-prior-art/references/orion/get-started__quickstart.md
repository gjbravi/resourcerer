# Quickstart

> Get recommendations in under 10 minutes


This quickstart installs Orion on your Kubernetes cluster with a single Helm command. Orion starts generating preliminary recommendations within 10 minutes. Complete, actionable recommendations are ready after a 7-day learning period.

## What you'll need

- A Kubernetes cluster (EKS, AKS, GKE, minikube, kind, or any other distribution)
- A Orion account — sign up for free at **[app.orion.example/signup](https://app.orion.example/signup)**
- `helm` installed locally

## Install Orion

1. Log in to **[app.orion.example](https://app.orion.example/)**.

    If this is your first time, you'll be taken directly to the **Add Cluster** page.

1. Click **Clusters** in the left navigation, then click **Add Cluster**.

1. Give your cluster a name and click **Continue**.

    The cluster name can't exceed 63 characters, can't start or end with a dash (`-`), and must follow the [RFC 1123 DNS label standard][rfc-1123].

1. Copy and run the generated `helm install` command. It looks like this (with your credentials in place of `CLIENT_ID` and `CLIENT_SECRET`):

    ```shell
    helm install orion oci://registry.orion.example/library/orion \
      --create-namespace --namespace orion-system \
      --set clusterName="my-cluster" \
      --set authorization.clientID="CLIENT_ID" \
      --set authorization.clientSecret="CLIENT_SECRET"
    ```

## Verify the installation

After a minute or two, confirm all pods are running:

```shell
kubectl get pods -n orion-system
```

You should see three pods:

```shell
NAME                                        READY   STATUS    RESTARTS   AGE
orion-cluster-agent-xxxxxxxxxx-xxxxx   2/2     Running   0          2m
orion-forwarder-xxxxxxxxxx-xxxxx       1/1     Running   0          2m
orion-applier-xxxxxxxxxx-xxxxx         1/1     Running   0          2m
```

Back in the Orion UI, click **Check Connection**. You'll be redirected to the Overview page with a confirmation that the Agent is connected.

## What happens next

Orion is now collecting workload metrics.

- **In ~10 minutes:** Preliminary recommendations appear. These are previews — don't apply them yet.
- **Over 7 days:** Orion collects metrics and refines its model.
- **After 7 days:** Complete recommendations are generated daily. These are ready to review and apply.

## Apply your first recommendation

Once a complete recommendation is available:

1. In the Orion UI, click **Workloads**.
1. Select a workload, then click its name to view the recommendation details.
1. Click **Apply Now** to apply the recommendation immediately.

## Next steps

- **[Enable automatic deployment](configure__settings__apply-settings.md#auto-deploy)** — let Orion apply recommendations on a schedule
- **[Configure](configure.md)** — tune optimization goals, schedules, and how recommendations are applied

[rfc-1123]: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
