# Manage secrets with a third-party credential manager

> Orion Agent secrets managed by a third-party credential manager


By default, Orion uses Helm to create and update its credentials Secret. However, if you would like to use a third-party credential manager (for example, to rotate secrets), you can choose to disable Helm management of the credentials secret.

When Helm management of the credentials secret is disabled, you must:

- Create the `orion-system` namespace before you install the agent.
- Create the `orion-auth` secret, and populate it with the required credentials, before you install the agent.
- Maintain the `orion-auth` secret yourself, manually or with a third-party credential manager.

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Without a credential manager</div>


If you are **not** using a third-party credential management tool, you do **not** need to follow the steps in this guide. Instead, follow the steps in the [Install](installation__install.md) topic. You still have to create an access credential (as described in the Install topic), but Orion will create and manage the secret for you.
</div>


Before you begin, make sure that the Orion CLI tool is [installed](installation__install-adv.md#prepare-to-install).

## Steps

1. Log in to your Orion account.

    ```sh
    orion login
    ```

1. Generate an access credential and save the output to a file.
    - Replace CREDENTIAL_NAME with a name that will help you identify the credential (such as the name of the cluster you're installing on).
    - Replace CREDENTIAL_FILE with a filename that will help you identify the file that contains the credential.

    ```sh
    orion auth create CREDENTIAL_NAME > CREDENTIAL_FILE
    ```

    The credential file will look something like this:

    ```yaml
    orion:
       address: https://api.orion.example/
    authorization:
       issuer: https://api.orion.example/
       clientID: CREDENTIAL_NAME 
       clientSecret: CREDENTIAL_SECRET
    ```

1. Create a Orion Agent secret and save it to a file.  
    - Replace CLUSTER_NAME with the name of the cluster you're installing the Agent on.
    - Replace CREDENTIAL_FILE with the filename from the previous step.
    - Replace SECRET_FILE with a filename that will help you to identify the secret (for example, `orion-auth`).

        ```sh
        helm template orion oci://registry.orion.example/library/orion \
          -n orion-system \
          -s templates/auth-secret.yaml \
          -f CREDENTIAL_FILE \
          --set clusterName=CLUSTER_NAME \
        | grep -vi -e app.kubernetes.io/managed-by -e app.kubernetes.io/version -e helm.sh/chart \
        > SECRET_FILE
        ```

    The secret file will look something like this:

    ```yaml
    ---
    # Source: orion/templates/common/auth-secret.yaml
    apiVersion: v1
    kind: Secret
    type: Opaque
    metadata:
      name: "orion-auth"
      namespace: "orion-system"
      labels:
        app.kubernetes.io/name: orion
        app.kubernetes.io/instance: orion
    data:
      ORION_CLIENT_ID: "**********"
      ORION_CLIENT_SECRET: "**********"
      ORION_ISSUER: "**********"
      ORION_SERVER: "**********"
    ```

1. Create the `orion-system` namespace and then apply the secret.
    - Replace SECRET_FILE with the name of the file that contains the secret that you generated in the previous step.

    ```sh
    kubectl create namespace orion-system
    kubectl apply -f SECRET_FILE -n orion-system
    ```

1. When you're ready to install Orion, include `--set createAuthSecret=false` to indicate to Orion that you or a third-party credential management tool will manage the secret.

    - Replace CLUSTER_NAME with the name of the cluster (lowercase, no underscore) you're installing on.

    ```sh
    helm install orion oci://registry.orion.example/library/orion \
      --namespace orion-system \
      --set clusterName=CLUSTER_NAME \
      --set createAuthSecret=false
    ```

## Related topics

- [`orion auth`](orion-cli__reference.md#orion-auth) command (CLI reference)
