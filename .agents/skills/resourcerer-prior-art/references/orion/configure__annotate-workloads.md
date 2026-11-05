# Configure optimization for individual workloads

> Optimization settings for a single workload, set with annotations


The purpose of this topic is to describe how to use annotations to configure optimization settings.

**Topics:**

- [Configure optimization settings for a specific workload](#annot-wl)
  - [Using YAML manifests](#annot-wl-yaml)
  - [Using `kubectl annotate`](#annot-wl-kubectl-annotate)

## Configure optimization settings for a specific workload {#annot-wl}

To configure settings an individual workload, edit the resource definition or run `kubectl annotate`.  

Workload optimization settings are defined by annotations in the `metadata.annotations` section of a Deployment, StatefulSet, or other supported workload resource type.

Workload optimization settings:

- Directly define configuration settings for a workload
- Are specified using annotations on workload resources, such as Deployments or StatefulSets
- Override equivalent values configured as defaults for the namespace or cluster

### Using YAML manifests {#annot-wl-yaml}

1. Open the workload definition in your favorite editor. This is commonly done using commands like these (remember to replace DEPLOYMENT and FILENAME with your own values):

    ```shell
    kubectl get deployment DEPLOYMENT -o yaml > FILENAME && $EDITOR FILENAME
    ```

    or:

    ```shell
    kubectl edit deployment DEPLOYMENT
    ```

1. Edit the `metadata.annotations` section to include the annotations you need. For a complete list of annotations for Orion settings, see the [Optimization settings and descriptions](configure__settings.md) topic.

    Your file will look something like this:

    ```yaml
    apiVersion: apps/v1 # ⎡ just an example; represents any  ⎤
    kind: Deployment    # ⎣ supported workload resource type ⎦
    metadata:
      name: example-workload-name
      annotations:
        live.orion.example/schedule: "@daily"
        live.orion.example/auto-deploy: "Enabled"
    ...
    ```

    <div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Tip</div>


    To save time, you can copy the complete list of annotations from the sample Deployment YAML excerpt on the **Workload** tab in the [Copy a template](configure.md#quickstart-templates) section of the configuration topic.
    </div>


1. Save your changes.

1. If you chose to run the `kubectl get deployment DEPLOYMENT -o yaml > FILENAME ...` command in step 1, apply your changes.

    ```shell
    kubectl apply -f FILENAME
    ```

1. Optional: Confirm your changes by running the following command and then reviewing the annotations in the `metadata.annotations` section of the output:

    ```shell
    kubectl get deployment DEPLOYMENT -o yaml
    ```

### Using the `kubectl annotate` command {#annot-wl-kubectl-annotate}

1. Run the `kubectl annotate` command and include the `annotation=value` pairs you want to set.

    The Kubernetes object you annotate should be the workload resource itself, such as a Deployment or StatefulSet.

    As an example, a `kubectl annotate` for a Deployment workload might look something like this (remember to replace NAME with your appropriate value):

    ```shell
    kubectl annotate deployment NAME \
      live.orion.example/schedule="@daily" \
      live.orion.example/auto-deploy="Enabled" \
      live.orion.example/containers.cpu.requests.min="20m" \
      live.orion.example/containers.cpu.requests.max="16000m"
    ```

    <div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Tip</div>


    To save time, you can copy the pairs you need from the sample Namespace YAML excerpt on the **Namespace** tab in the [Copy a template](configure.md#quickstart-templates) section of the configuration topic. Be sure to replace the `: ` separator from the template with `=` as shown in the sample command above.
    </div>


1. Optional: To confirm your changes, run the following command and review the annotations in the `.metadata.annotations` section of the output:

    ```shell
    kubectl get deployment NAME -o yaml
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
                    
                    Configure<span class="related-topics-divider">/</span><a href="/docs/configure/cluster-defaults/">Configure clusters</a>
                </h5>
                <p>Configure default optimization settings for the cluster</p>
            </div>
        </div>
