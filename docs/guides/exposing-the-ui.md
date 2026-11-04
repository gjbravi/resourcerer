---
title: Exposing the UI
description: Reach the UI with a port-forward, sign users in with OIDC, or expose it to your team behind an authenticating proxy.
---

!!! danger "No login unless you turn it on"
    `auth.strategy: anonymous`, the default, identifies nobody: anyone who reaches the UI or the API can apply, revert
    and change settings, with the server's permissions. Never expose the Service in that mode without an
    authenticating proxy in front, and consider `readOnly: true` for a wide audience. With `auth.strategy: oidc` the
    server signs users in itself — see [Sign-in with OIDC](#sign-in-with-oidc).


## Port-forward

The simplest and safest way, for one person:

```bash
kubectl -n resourcerer port-forward svc/resourcerer 8080:80
```

Access is then whatever your Kubernetes RBAC allows: only people who can port-forward in the `resourcerer` namespace
reach the UI.

## Sign-in with OIDC

`auth.strategy: oidc` makes the server its own gatekeeper, so the Service can be exposed as it is: the user signs in
at your provider (authorization code with PKCE), the server keeps the identity in a sealed cookie, and the groups from
the token decide the role. Nothing is stored server-side, so no replica holds a session.

```yaml title="values.yaml"
auth:
  strategy: oidc
  oidc:
    issuer: https://sso.example.com/realms/example
    clientID: resourcerer
    clientSecretSecretRef:
      name: resourcerer-oidc
      key: clientSecret
    allowedGroups: [resourcerer-users]
    roles:
      editorGroups: [resourcerer-editors]
  session:
    keySecretRef:
      name: resourcerer-session
      key: key
```

The provider needs a **confidential client** using the authorization code flow, with the callback
`https://<host>/auth/callback` registered. The chart derives that URL from the request, so every way of reaching the
server — the Ingress host, a port-forward — needs its own callback; `auth.oidc.redirectURL` sets it instead, for a
proxy that rewrites the `Host`.

Two Secrets hold the material, never the values file:

```bash
kubectl -n resourcerer create secret generic resourcerer-oidc --from-literal=clientSecret=<secret>
kubectl -n resourcerer create secret generic resourcerer-session \
  --from-literal=key="$(openssl rand -base64 32)"
```

`editorGroups` empty means nobody may change anything, and `allowedGroups` empty admits every user the provider
authenticates as a reader. Both default empty, so set at least `editorGroups` before exposing the UI. A user outside
`allowedGroups` is refused at the callback; one whose groups or configuration changed after signing in lands on a
page that names the two values to check. Groups are re-read on every request, so an edit to `editorGroups` takes
effect on the next call, with no new sign-in.

Sessions last `auth.oidc.sessionTTL` (12h by default) and nothing revokes one earlier: signing out clears the cookie
in that browser and leaves the provider's own session alone, so signing in again usually asks for nothing. Behind a
TLS-terminating proxy set `auth.cookieSecure: true`, so the cookie is always `Secure`.

The rules that decide who may read and who may write, and what the configuration itself is trusted to do, are in the
[security model](https://github.com/gjbravi/resourcerer/blob/main/SECURITY.md).

## Behind an authenticating proxy

To share the UI with a team, put a proxy that authenticates users in front of the Service, such as
[oauth2-proxy](https://oauth2-proxy.github.io/oauth2-proxy/) with your identity provider, or your ingress controller's
external authentication. A minimal Ingress with NGINX and oauth2-proxy:

```yaml title="values.yaml"
server:
  ingress:
    enabled: true
    className: nginx
    annotations:
      nginx.ingress.kubernetes.io/auth-url: "https://oauth2.example.com/oauth2/auth"
      nginx.ingress.kubernetes.io/auth-signin: "https://oauth2.example.com/oauth2/start?rd=$scheme://$host$escaped_request_uri"
    hosts:
      - host: resourcerer.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: resourcerer-tls
        hosts: [resourcerer.example.com]
```

Restrict the proxy to the people who should be able to change workloads. Everyone it lets in has the same rights.
Under `oidc` the server knows who each caller is, so a proxy in front of it is for what the server does not do —
TLS, a company-wide sign-in the provider does not offer — not for authorization.

## Read-only for a wide audience

With `readOnly: true` the server refuses every change, editors included, and its RBAC drops the `patch` verbs, so
the UI can be shared widely as a report, and changes go through Git from the **Patches** tab. `readOnly` applies to
the whole release: run one resourcerer per cluster.

## The API

The UI's HTTP API is served on the same Service, under `/api/v1`, with the same authentication: `anonymous`, or the
`oidc` session cookie, plus the `X-Resourcerer-Request: 1` header and a JSON body on mutations. See
[HTTP API](../reference/http-api.md).

## What is coming

A trusted-header mode for proxies that pass the user's identity, per-user authorization, and an override that lets
an authorized caller past an advisory apply window are designed in
[#33](https://github.com/gjbravi/resourcerer/issues/33). Under `oidc` the history already records who applied.
