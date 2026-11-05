# Import recommendation data into APM tools

> Prometheus scrape job for importing recommendation data into APM tools


By creating a Prometheus scrape job, Orion exports recommendation data as Prometheus metrics, which you can import into other application performance management (APM) tools.

To define a scrape job to get recommendation data:  
1. Edit the `prometheus-server` ConfigMap in the `prometheus` namespace: 
    ```sh
    kubectl edit configmap -n prometheus prometheus-server
    ```
 
1. Add the following scrape job to the `scrape_configs:` section of the ConfigMap.
    - Replace SF_CLIENT_ID with the clientID value in your cluster credential file.
    - Replace SF_CLIENT_SECRET with the clientSecret value in your cluster credential file.\
    Alternatively, a Prometheus administrator can put the clientSecret in a file and leverage the `client_secret_file` parameter and specify the absolute path (for example, `client_secret: /mnt/client_secret_file`).

        ```yaml
        - job_name: orion-recommendations
          scrape_interval: 90s 
          scrape_timeout: 15s
          honor_labels: true
          proxy_from_environment: true
          scheme: https
          oauth2:
            client_id: 'SF_CLIENT_ID'
            client_secret: 'SF_CLIENT_SECRET'
            token_url: 'https://api.orion.example/oauth/token'
            endpoint_params:
                audience: 'https://api.orion.example/'
          metrics_path: /v3/workload-metrics
          static_configs:
          - targets: ["api.orion.example"]
        ```
