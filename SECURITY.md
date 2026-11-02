# Security policy

## Supported versions

resourcerer is in beta. Only the latest release receives security fixes.

## Reporting a vulnerability

Report vulnerabilities privately through
[GitHub security advisories](https://github.com/gjbravi/resourcerer/security/advisories/new).
Please do not open public issues for security problems. You will get an acknowledgement within a week.

## Security model

- **Two authentication strategies.** `auth.strategy: anonymous` (the default) identifies nobody: anyone who can reach
  the server's Service uses the UI and API with the server's ServiceAccount, which is what lets a port-forward install
  work without setup. Keep the Service internal, or put an authenticating proxy in front; the chart warns when an
  Ingress is enabled without `readOnly`. `auth.strategy: oidc` signs callers in at an OpenID Connect provider, and the
  groups in the token decide what they may do: `allowedGroups` admits them at all, `roles.editorGroups` lets them
  apply, revert and change settings. Someone signed in outside both gets an error, not a read view.
- **`readOnly: true`** makes the API refuse every change, and the server's ClusterRole drops its `patch` verbs.
- **The authentication configuration is the trust boundary.** Whoever can write the OIDC ConfigMap can point `issuer`
  at a provider they run and sign in as an editor; whoever can read the session-key Secret can mint a cookie for any
  user. Write access to ConfigMaps and Secrets in the release namespace is admin access to resourcerer.
- **A session is a sealed cookie, not server state**: the identity, encrypted with AES-GCM under the session key, in a
  cookie that is `HttpOnly`, `SameSite=Lax` and `Path=/`, and carries the `__Host-` prefix where it is `Secure`. No
  replica holds sessions, and nothing can be revoked early: a leaked cookie or a departed employee stays usable until
  `auth.oidc.sessionTTL` passes, which is what that value trades. Replacing the session key, or the strategy, signs
  everyone out.
- **Changes have to come from the UI.** Under `oidc`, a mutation also needs the header the UI attaches and a body
  typed as JSON, and is refused with `403 ajax-required` without them. That is the CSRF guard: a page on another origin
  can make the browser send a session cookie with a form, but it cannot attach a header this API grants to no origin.
  An `anonymous` install skips the guard — there is no session to ride, and the API is meant to be scripted there.
- **Separate identities per component:**
  - `controller` reads workloads, pods, HPAs and Prometheus, and writes only `Recommendation` objects and Events.
  - `applier` is the only component that patches workloads and HPA targets. Disable it with
    `applier.enabled: false` for a recommend-only install.
  - `server` reads, records apply requests on `Recommendation` objects and sets `resourcerer.io/*` annotations.
- Every apply is recorded in the `Recommendation` status (who, when, from and to) and as a Kubernetes Event. Under
  `oidc` that name is the signed-in user, taken from the session; no request can claim to be someone else.

## Verifying releases

Images and charts are signed with [cosign](https://github.com/sigstore/cosign) keyless signing from the release
workflow, and images carry an SBOM and SLSA provenance. Releases from 0.3.0 are signed by the workflow on `main`,
earlier ones by the workflow on their tag:

```bash
cosign verify ghcr.io/gjbravi/resourcerer:<version> \
  --certificate-identity-regexp '^https://github.com/gjbravi/resourcerer/\.github/workflows/release\.yaml@refs/(heads/main|tags/v.+)$' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com

cosign verify ghcr.io/gjbravi/charts/resourcerer:<version> \
  --certificate-identity-regexp '^https://github.com/gjbravi/resourcerer/\.github/workflows/release\.yaml@refs/(heads/main|tags/v.+)$' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```
