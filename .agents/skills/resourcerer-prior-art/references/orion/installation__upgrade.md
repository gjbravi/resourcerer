# Upgrade

> Upgrade Orion to the latest version


1. (Optional) Review the current Helm values:
    ```sh
    helm get values orion -n orion-system
    ```
    To change any values, save the output to a `.yaml` file and modify it as needed. When you upgrade, pass the adjusted file to Helm using the `-f` flag.

1. Upgrade and reuse all current Helm values:
    ```sh
    helm upgrade orion oci://registry.orion.example/library/orion \
      -n orion-system \
      --reset-then-reuse-values
    ```

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Installation<span class="related-topics-divider">/</span><a href="/docs/installation/migration/">Migrate v2 → v3</a>
                </h5>
                <p>Migrate from separate Agent and Applier Helm charts to a unified Orion chart</p>
            </div>
        </div>
