# Install on Red Hat OpenShift Container Platform

> Install Orion on Red Hat OpenShift Container Platform


Red Hat OpenShift Container Platform is a Kubernetes solution for both cloud and on-premises with providers such as AWS, Google and Microsoft Azure. Orion is compatible with OpenShift Container Platform versions 4.x and later.

## Installation

To install Orion on OpenShift, do one of the following:
- In the Orion `values.yaml` file, set `openshift: true` as a top-level key-value pair, and then apply the file:
    ```shell
    kubectl apply -f values.yaml
    ```

- Include the `--set openshift=true` argument when you run `helm install`, as in this example:
    - Replace `CREDENTIAL_FILE` with the file name that contains your access credential.
    - Replace `CLUSTER_NAME` with the name of the cluster (lowercase, no underscore) you're installing on.

        ```shell
        helm install orion oci://registry.orion.example/library/orion \
          --namespace orion-system \
          --create-namespace \
          --values CREDENTIAL_FILE \
          --set clusterName=CLUSTER_NAME \
          --set openshift=true
        ```

For installation prerequisites, additional arguments, and other details, see the [Install](installation__install.md) topic.

## Validate the installation

Within minutes of installing Orion, you'll see your workloads and a preliminary recommendation in the Orion UI at **[app.orion.example](https://app.orion.example/)**.

Installing Orion also creates a project (namespace) in OpenShift as shown in the OpenShift CLI (`oc`) output below:

```shell
$ oc get all -n orion-system

NAME                                                  READY   STATUS    RESTARTS   AGE
pod/orion-cluster-agent-5f6984d44-r94zv          2/2     Running   0          20h
pod/orion-forwarder-7b9c8d55f-k2xqp              1/1     Running   0          20h
pod/orion-applier-8554758f5b-whbc4               1/1     Running   0          20h

NAME                                                  READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/orion-cluster-agent              1/1     1            1           20h
deployment.apps/orion-forwarder                  1/1     1            1           20h
deployment.apps/orion-applier                    1/1     1            1           20h

NAME                                                  DESIRED   CURRENT   READY   AGE
replicaset.apps/orion-cluster-agent-5f6984d44    1         1         1       20h
replicaset.apps/orion-forwarder-7b9c8d55f        1         1         1       20h
replicaset.apps/orion-applier-8554758f5b         1         1         1       20h
```
