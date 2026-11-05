# Configure optimization defaults for the cluster

> Configure default optimization settings for the cluster


The purpose of this topic is to describe how to use the `cluster-defaults` ConfigMap to configure default optimization settings for a cluster.

**Topics:**

- [Configure default optimization settings for all workloads in the cluster](#configure-defaults)
  - [Using `kubectl apply`](#defaults-using-kubectl-apply)
  - [Using `kubectl edit`](#defaults-using-kubectl-edit)

## Configure default optimization settings for all workloads in the cluster {#configure-defaults}

Cluster-level defaults:

- Define the default optimization settings for all workloads in a cluster
- Are specified in a ConfigMap named `cluster-defaults` in the `orion-system` namespace
- Have the lowest configuration precedence and can be overridden by namespace annotations or by annotations on individual workload resources

### Setting cluster default values using `kubectl apply` {#defaults-using-kubectl-apply}

1. Create a YAML file that defines a ConfigMap called `cluster-defaults`, with a `cluster-defaults.yaml` data key, and the settings you want to set in annotation syntax.

    For a complete list of annotations for Orion settings, see the [Optimization settings and descriptions](configure__settings.md) topic.
  
    Your file will look something like this:

    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: cluster-defaults
      namespace: orion-system
    data:
      cluster-defaults.yaml: |
        live.orion.example/schedule: "P1D"
        live.orion.example/auto-deploy: "Disabled"
        live.orion.example/containers.cpu.requests.min: "10m"
        live.orion.example/containers.cpu.requests.max: "16000m"
    ...
    ```

1. Apply the file:

    ```shell
    kubectl apply -f FILENAME -n orion-system
    ```

1. Restart the Orion Agent:

    ```shell
    kubectl rollout restart deployment orion-cluster-agent -n orion-system
    ```

    To check the status of the restart, run:

    ```sh
    kubectl rollout status deployment orion-cluster-agent -n orion-system -w
    ```

    Any previous cluster-level default values are now removed, and your new defaults are in effect unless they are overridden at the namespace or workload level.

1. Optional: Confirm your changes by running the following command and then reviewing the `cluster-defaults.yaml` item in the **Data** section of the output:

    ```shell
    kubectl describe configmap cluster-defaults -n orion-system 
    ```

### Setting cluster default values using `kubectl edit` {#defaults-using-kubectl-edit}

If the `cluster-defaults` ConfigMap has already been created, you can edit the ConfigMap directly.

1. Run the following command:

    ```shell
    kubectl edit configmap cluster-defaults -n orion-system
    ```

    If you see a `Error from server (NotFound): configmaps "cluster-defaults" not found` message, no cluster-level defaults are set. You must first create and apply a cluster defaults YAML file as described [above](#defaults-using-kubectl-apply).

    If you don't see an error message, continue to the next step.

1. In the `.data.cluster-defaults.yaml` section, add, edit, or remove annotation strings as needed.

      As an example, your annotations in the `cluster-defaults` ConfigMap might look something like this excerpt when you're done:

    ```yaml
    # Please edit the object below. Lines beginning with a '#' will be ignored,
    # and an empty file will abort the edit. If an error occurs while saving this file will be
    # reopened with the relevant failures.
    #
    apiVersion: v1
    kind: ConfigMap
    data:
      cluster-defaults.yaml: |
        live.orion.example/auto-deploy: "Enabled"
        live.orion.example/cpu.optimization-goal: "Reliability"
        live.orion.example/memory.optimization-goal: "Reliability"
    ...
    ```

    <div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Tip</div>


    To save time, you can copy the complete list of annotations from the sample ConfigMap YAML excerpt on the **Cluster** tab in the [Copy a template](configure.md#quickstart-templates) section of the configuration topic.
    </div>


1. Save your changes.
1. Restart the Orion Agent:

    ```shell
    kubectl rollout restart deployment orion-cluster-agent -n orion-system
    ```

    To check the status of the restart, run:

    ```sh
    kubectl rollout status deployment orion-cluster-agent -n orion-system -w
    ```

    Any previous cluster-level default values are now removed, and your new defaults are in effect unless they are overridden at the namespace or workload level.

1. Optional: Confirm your changes by running the following command and then reviewing the annotations in the **Data** section of the output:

    ```shell
    kubectl describe configmap cluster-defaults -n orion-system 
    ```

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Configure<span class="related-topics-divider">/</span>Settings<span class="related-topics-divider">/</span><a href="/docs/configure/settings/annotations/">Annotations</a>
                </h5>
                <p>All available annotations for Orion configuration</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Configure<span class="related-topics-divider">/</span><a href="/docs/configure/annotate-namespaces/">Configure namespaces</a>
                </h5>
                <p>Default optimization settings for every workload in a namespace, set with annotations</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Configure<span class="related-topics-divider">/</span><a href="/docs/configure/annotate-workloads/">Configure workloads</a>
                </h5>
                <p>Optimization settings for a single workload, set with annotations</p>
            </div>
        </div>
