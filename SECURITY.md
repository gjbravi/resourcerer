# Security policy

## Supported versions

resourcerer is in beta. Only the latest release receives security fixes.

## Reporting a vulnerability

Report vulnerabilities privately through
[GitHub security advisories](https://github.com/gjbravi/resourcerer/security/advisories/new).
Please do not open public issues for security problems. You will get an acknowledgement within a week.

## Security model

- **Authentication is anonymous.** Anyone who can reach the server's Service can use the UI and API with the
  server's ServiceAccount. Keep the Service internal and use `kubectl port-forward`, or put an authenticating
  proxy in front. The chart warns when an Ingress is enabled without `readOnly`.
- **`readOnly: true`** makes the API refuse every change, and the server's ClusterRole drops its `patch` verbs.
- **Separate identities per component:**
  - `controller` reads workloads, pods, HPAs and Prometheus, and writes only `Recommendation` objects and Events.
  - `applier` is the only component that patches workloads and HPA targets. Disable it with
    `applier.enabled: false` for a recommend-only install.
  - `server` reads, records apply requests on `Recommendation` objects and sets `resourcerer.io/*` annotations.
- Every apply is recorded in the `Recommendation` status (who, when, from and to) and as a Kubernetes Event.

## Verifying releases

Images and charts are signed with [cosign](https://github.com/sigstore/cosign) keyless signing from the release
workflow, and images carry an SBOM and SLSA provenance:

```bash
cosign verify ghcr.io/gjbravi/resourcerer:<version> \
  --certificate-identity-regexp '^https://github.com/gjbravi/resourcerer/\.github/workflows/release\.yaml@refs/tags/v' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com

cosign verify ghcr.io/gjbravi/charts/resourcerer:<version> \
  --certificate-identity-regexp '^https://github.com/gjbravi/resourcerer/\.github/workflows/release\.yaml@refs/tags/v' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```
