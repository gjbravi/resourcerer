# Security policy

## Supported versions

Only the latest minor release receives security fixes.

## Reporting a vulnerability

Report vulnerabilities privately through
[GitHub security advisories](https://github.com/gjbravi/resourcerer/security/advisories/new).
You will get an acknowledgement within a week.

## Deployment notes

- The UI and API have no authentication. Expose them only behind an authenticating proxy, or keep the Service
  internal and use `kubectl port-forward`.
- With `ui.readOnly: true` the API cannot change workloads, and the ClusterRole drops the `patch` verb on them.
- The controller can create and delete VPAs in every namespace that is not excluded.
