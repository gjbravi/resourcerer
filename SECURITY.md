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
  Ingress is enabled without `readOnly`. `auth.strategy: oidc` signs callers in at an OpenID Connect provider, and
  three rules decide what they may do:
  - `roles.anyAuthenticatedIsViewer` (default `true`) gives an admitted user read access;
  - `roles.editorGroups` gives apply, revert and settings changes. It defaults empty, so nobody may change anything;
  - `allowedGroups`, when set, is the perimeter: a user outside it has no access at all, a member of `editorGroups`
    included. With it empty, `editorGroups` alone decides who may write; with `anyAuthenticatedIsViewer` at its
    default, `allowedGroups` alone decides who may read.

  `allowedGroups` and `anyAuthenticatedIsViewer` default open, so **an `oidc` install with the defaults lets every
  user the provider authenticates read**, and nobody write. Set `allowedGroups` when not everyone who can sign in may
  read it, and `editorGroups` to say who may change things. A user the rules leave with no role is refused: an error
  on sign-in, and `403 forbidden` on every call but `GET /api/v1/me`.
- **`readOnly: true`** makes the API refuse every change, and the server's ClusterRole drops its `patch` verbs.
- **The authentication configuration is the trust boundary.** Whoever can write the OIDC ConfigMap can point `issuer`
  at a provider they run and sign in as an editor; whoever can read the session-key Secret can mint a cookie for any
  user. Write access to ConfigMaps and Secrets in the release namespace is admin access to resourcerer.
- **A session is a sealed cookie, not server state**: the identity, encrypted with AES-GCM under the session key, in a
  cookie that is `HttpOnly`, `SameSite=Lax` and `Path=/`, and carries the `__Host-` prefix where it is `Secure`. No
  replica holds sessions, and nothing can be revoked early: a leaked cookie or a departed employee stays usable until
  `auth.oidc.sessionTTL` passes, which is what that value trades. Replacing the session key, or the strategy, signs
  everyone out. Signing out is `GET /auth/logout`, which answers every other method with `405`: it carries no token and
  clears the cookie whatever the request brought with it, so any page that gets the browser to navigate there — a link,
  or a form that submits by GET — signs the user out. That is a nuisance and nothing more: the route reads nothing,
  changes nothing else, and signing in again brings the session back. What it ends is resourcerer's session only. The
  provider's own is still open, so the signed-out screen the browser lands on (`/?signed-out`) is one click from a new
  session and no password, and signing out of resourcerer on a shared machine is not signing out of the provider:
  whoever needs this machine to forget them signs out there too.
- **Changes have to come from the UI.** Under `oidc`, a mutation also needs the header the UI attaches and a body
  typed as JSON, and is refused with `403 ajax-required` without them. That is the CSRF guard: a page on the same site
  — a sibling subdomain — can make the browser send the session cookie with a form, and `SameSite=Lax` is what stops a
  cross-site page doing the same. Neither can attach a header this API grants to no origin. An `anonymous` install
  skips the guard — there is no session to ride, and the API is meant to be scripted there.
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
