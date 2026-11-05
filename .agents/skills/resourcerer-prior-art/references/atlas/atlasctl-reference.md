---
updatedAt: 2026-10-28T08:54:20.000Z
---

# atlasctl command reference

Every atlasctl command and flag, in one place, for connecting and disconnecting clusters, migrating standalone charts to the umbrella chart, importing existing node configuration, and connecting AI tools to Atlas MCP.

Every `atlasctl` command and its flags, for lookup.

Every command also accepts `atlasctl <command> --help` for the flags currently supported by your installed version.

## atlasctl auth login

Authenticates `atlasctl` with your Atlas account and stores the token in `~/.config/atlasctl/config.yaml`. Takes no flags. Run it again to switch accounts or organizations.

## atlasctl cluster connect

| Flag                            | Description                                                                                                                                                                                                        |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `--dry-run`                     | Resolve and print the configuration as YAML without registering the cluster or making changes.                                                                                                                     |
| `--non-interactive`             | Skip prompts; combine with feature-toggle flags for CI.                                                                                                                                                            |
| `--provider anywhere`           | Connect a cluster outside EKS, GKE, or AKS via [Atlas Anywhere](atlas-anywhere-overview.md).                                                                                                                  |
| `--umbrella-values-file <path>` | Merge a YAML file of Helm values into the umbrella chart install.                                                                                                                                                  |
| `--import-config`               | After a successful connect, run node-group discovery and import it as node configurations and node templates. See [Import and apply cluster configuration with atlasctl](atlasctl-import-cluster-configuration.md). |

Feature-toggle flags (for example `--cluster-optimization`, `--workload-autoscaler`) and cloud-credential flags vary by provider. Run `atlasctl cluster connect --help` for the complete, current list.

See [Connect using the atlasctl CLI](connect-with-atlasctl.md).

## atlasctl cluster disconnect

No flags are required for the common case: `atlasctl` reads the cluster ID from the in-cluster configuration. Run `atlasctl cluster disconnect --help` for flags that skip cloud cleanup or the Kubernetes uninstall.

See [Connect using the atlasctl CLI](connect-with-atlasctl.md#disconnect-a-cluster).

## atlasctl cluster migrate

| Flag                           | Description                                                                                                                             |
| ------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| `--yes`                        | Skip the interactive confirmation prompt.                                                                                               |
| `--umbrella-version <version>` | Install a specific umbrella chart version instead of the latest stable.                                                                 |
| `--values-dump-path <path>`    | On failure, write the preserved standalone release values to this path instead of the default `atlas-migrate-values-<timestamp>.yaml`. |
| `--values-file <path>`         | Recovery mode: skip standalone detection and uninstall, and install the umbrella chart using the values in this file.                   |

Three subcommands support migrating clusters managed by GitOps or IaC (available in `atlasctl` 0.15.0 and later):

### atlasctl cluster migrate generate

Read-only: scans the cluster for standalone Atlas releases (including ArgoCD-managed components) and writes an umbrella chart values file, a tool-specific manifest, and a step-by-step `MIGRATION-STEPS.md` checklist.

| Flag                     | Description                                                                                                                                                                            |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--output`, `-o <dir>`   | Directory to write the generated files. Defaults to the current directory.                                                                                                             |
| `--format <format>`      | Output format: `generic` (default), `argocd`, `flux`, or `terraform` (also valid for OpenTofu).                                                                                        |
| `--profile <profile>`    | Override the auto-detected profile: `autoscaler` or `autoscaler-anywhere` for [Atlas Anywhere](atlas-anywhere-overview.md) clusters.                                              |
| `--repo-url <url>`       | GitOps repository URL that will host the generated `values.yaml`. Recommended for `--format argocd`; without it the generated Application contains a `<YOUR-GITOPS-REPO>` placeholder. |
| `--namespace`, `-n <ns>` | Namespace where the standalone Atlas components are installed. Defaults to `atlas-agent`.                                                                                           |

### atlasctl cluster migrate pre-umbrella-install

Prepares the cluster for umbrella adoption without installing the umbrella chart: preserves Workload Autoscaler CRDs, deletes the standalone Helm release secrets (cluster resources survive), and orphan-deletes Deployments and DaemonSets with immutable selectors. Pods keep running.

| Flag                     | Description                                                                                                                                     |
| ------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `--strategy <strategy>`  | Adoption strategy: `helm-secrets` or `server-side-apply`. Auto-detected if not set (`server-side-apply` for kubectl-applied manifest installs). |
| `--dry-run`              | Show what would be done without making changes.                                                                                                 |
| `--yes`                  | Skip the interactive confirmation prompt.                                                                                                       |
| `--namespace`, `-n <ns>` | Namespace where the standalone Atlas components are installed. Defaults to `atlas-agent`.                                                    |

### atlasctl cluster migrate post-umbrella-install

Cleans up after the umbrella chart is installed and its pods are Ready. It deletes stale ReplicaSets and orphaned pods from the standalone releases. On ArgoCD, it also sets `preDeleteHook.enabled=false` on Applications running the standalone workload-autoscaler chart, so Recommendation CR finalizers are preserved.

| Flag                     | Description                                                                                  |
| ------------------------ | -------------------------------------------------------------------------------------------- |
| `--dry-run`              | Show what would be done without making changes.                                              |
| `--yes`                  | Skip the interactive confirmation prompt.                                                    |
| `--namespace`, `-n <ns>` | Namespace where the standalone Atlas components are installed. Defaults to `atlas-agent`. |

See [Migrate standalone Helm charts with atlasctl](atlasctl-migrate-standalone-helm-charts.md) for the imperative path and [Migrate GitOps and IaC installations with atlasctl](atlasctl-migrate-gitops.md) for the phased ArgoCD, Flux, and Terraform flow.

## atlasctl cluster import-config

| Flag                | Description                                                                                |
| ------------------- | ------------------------------------------------------------------------------------------ |
| `--cluster-id <id>` | Atlas cluster ID. Auto-detected from the cluster in your kubeconfig if not provided.     |
| `--file <path>`     | Write the discovered plan to a file instead of applying it.                                |
| `--output <format>` | Plan serialization when `--file` is set: `yaml` (default) or `tf` (apply-ready Terraform). |
| `--non-interactive` | Skip the interactive review and apply the discovered plan directly.                        |

See [Import and apply cluster configuration with atlasctl](atlasctl-import-cluster-configuration.md).

## atlasctl cluster apply-config

| Flag            | Description                                                                                                            |
| --------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `--file <path>` | Plan file to apply. Required. `apiVersion`, `kind`, and `metadata.clusterId` are validated before anything is applied. |

See [Apply a plan from a file](atlasctl-import-cluster-configuration.md#apply-a-plan-from-a-file).

## atlasctl mcp

Configures AI clients to connect to [Atlas MCP](atlaspilot-mcp.md) for read-only access to your Atlas data from AI tools.

### atlasctl mcp add

Detects supported AI clients installed on your machine and configures the Atlas MCP connection for them. Clients it cannot configure automatically, such as Claude Desktop, are listed with manual instructions and a link to the vendor's setup guide. Complete the browser sign-in when your client prompts for it; the timing depends on the client. For the full setup flow, see [Connect AI tools to Atlas MCP](atlaspilot-mcp.md).

| Flag                    | Description                                                |
| ----------------------- | ---------------------------------------------------------- |
| `--api-region <region>` | Atlas region for the connection: `us`, `eu`, or `india`. |

## atlasctl version

Prints the installed `atlasctl` version. Takes no flags.

## See also

<Cards columns="2">
  <Card title="Connect using the atlasctl CLI" href="https://docs.atlas.example/docs/connect-with-atlasctl" icon="fa-terminal">
    Install atlasctl, log in, and connect a cluster.
  </Card>

  <Card title="Import and apply cluster configuration with atlasctl" href="https://docs.atlas.example/docs/atlasctl-import-cluster-configuration" icon="fa-download">
    The recommended path to prepare a connected cluster for Node Autoscaling.
  </Card>
</Cards>