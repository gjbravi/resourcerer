# Transparent proxy configuration

> Orion configuration for a transparent proxy


If you use a transparent proxy, you need to configure Orion to trust the proxy's root certificate authority (CA). To do this, mount the proxy's root CA certificate using a volume, then use `additionalEnvVars` to configure Orion to use the mounted bundle.

## Prerequisites

- Helm version 3.14.0 or later (required for the `--reset-then-reuse` parameter)

## Steps

1. Copy the YAML below into a new file named, for example, `transparent-proxy-values.yaml`. This example assumes the required root CA is available in a ConfigMap, but any `core/v1 Volumes` type can be used (see the related link below).

    ```yaml
    ---
    # Configures additional volumes that will be added to all pods
    additionalVolumes:
    - name: corporate-root-ca
      configMap:
        name: corporate-root-ca
        items:
        - key: ca.crt
          path: ca.crt

    # Configures additional volume mounts that will be added to all containers
    additionalVolumeMounts:
    - name: 
      mountPath: /etc/ssl/certs

    # Configures additional environment variables that will be added to all containers
    additionalEnvVars: 
    - name: SSL_CERT_FILE  # or SSL_CERT_DIR. Configures Orion apps to use the mounted bundle.
      value: /etc/ssl/certs/ca.crt
    ```

1. Run `helm upgrade` to apply the new settings:

    ```sh
    helm upgrade orion oci://registry.orion.example/library/orion \
        -n orion-system \
        --reset-then-reuse-values \
        -f transparent-proxy-values.yaml
    ```

## Related topics

- [Additional volumes and mounts][volumes]

[volumes]: /docs/installation/cluster-agent-config/#pods-mounts-volumes
