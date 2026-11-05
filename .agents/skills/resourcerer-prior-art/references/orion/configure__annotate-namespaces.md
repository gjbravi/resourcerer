# Configure optimization defaults for namespaces

> Default optimization settings for every workload in a namespace, set with annotations


The purpose of this topic is to describe how to use annotations to configure optimization settings.

**Topics:**

- [Configure default optimization settings for all workloads in a namespace](#annot-ns)
  - [Using YAML manifests](#annot-ns-yaml)
  - [Using `kubectl annotate`](#annot-ns-kubectl-annotate)

## Configure default optimization settings for all workloads in a namespace {#annot-ns}

To configure settings for all workloads in a namespace, edit the Namespace definition or run `kubectl annotate`.

Namespace-level defaults:

- Define the default optimization settings for all workloads in a namespace
- Override any equivalent cluster-level defaults
- Are specified using annotations on the namespace
- Fill in any settings not provided by annotations on the workloads themselves

### Using YAML manifests {#annot-ns-yaml}

1. Open the namespace definition in your favorite editor. This is commonly done using commands like these (remember to replace NAMESPACE and FILENAME with your own values):

    ```shell
    kubectl get namespace NAMESPACE -o yaml > FILENAME && $EDITOR FILENAME
    ```

    or:

    ```shell
    kubectl edit namespace NAMESPACE
    ```

1. Edit the `metadata.annotations` section to include the annotations you need. For a complete list of annotations for Orion settings, see the [Optimization settings and descriptions](configure__settings.md) topic.

    Your file will look something like this:

    ```yaml
    apiVersion: v1
    kind: Namespace
    metadata:
      name: example-namespace-name
      annotations:
        live.orion.example/schedule: "@daily"
        live.orion.example/auto-deploy: "Enabled"
    ...
    ```

    <div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Tip</div>


    To save time, you can copy the complete list of annotations from the sample Namespace YAML excerpt on the **Namespace** tab in the [Copy a template](configure.md#quickstart-templates) section of the configuration topic.
    </div>


1. Save your changes.

1. If you chose to run the `kubectl get namespace NAMESPACE -o yaml > FILENAME ...` command in step 1, apply your changes.

    ```shell
    kubectl apply -f FILENAME
    ```

1. Optional: Confirm your changes by running the following command and then reviewing the annotations in the `metadata.annotations` section of the output:

    ```shell
    kubectl get namespace NAMESPACE -o yaml
    ```

### Using the `kubectl annotate` command {#annot-ns-kubectl-annotate}

1. Run the `kubectl annotate` command and include the annotation:value pairs you want to set.

    As an example, a `kubectl annotate` might look something like this (remember to replace NAMESPACE with your appropriate value):

    ```shell
    kubectl annotate namespace NAMESPACE \
      live.orion.example/schedule="P1D" \
      live.orion.example/auto-deploy="Disabled" \
      live.orion.example/containers.cpu.requests.min="20m" \
      live.orion.example/containers.cpu.requests.max="16000m"
    ```

    <div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Tip</div>


    To save time, you can copy the complete list of annotations from the sample Namespace YAML excerpt on the **Namespace** tab in the [Copy a template](configure.md#quickstart-templates) section of the configuration topic. Be sure to replace the `: ` separator from the template with `=` as shown in the sample command above
    </div>


1. Optional: To confirm your changes, run the following command and review the annotations in the `.metadata.annotations` section of the output:

    ```shell
    kubectl get namespace NAMESPACE -o yaml
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
                    
                    Configure<span class="related-topics-divider">/</span><a href="/docs/configure/cluster-defaults/">Configure clusters</a>
                </h5>
                <p>Configure default optimization settings for the cluster</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Configure<span class="related-topics-divider">/</span><a href="/docs/configure/annotate-workloads/">Configure workloads</a>
                </h5>
                <p>Optimization settings for a single workload, set with annotations</p>
            </div>
        </div>
