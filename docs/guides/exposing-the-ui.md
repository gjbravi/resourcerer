---
title: Exposing the UI
description: Reach the UI with a port-forward, or expose it to your team behind an authenticating proxy.
---

!!! danger "No login yet"
    Authentication is `anonymous` for now: anyone who reaches the UI or the API can apply, revert and change settings,
    with the server's permissions. Never expose the Service without an authenticating proxy in front, and consider
    `readOnly: true` for a wide audience.


## Port-forward

The simplest and safest way, for one person:

```bash
kubectl -n resourcerer port-forward svc/resourcerer 8080:80
```

Access is then whatever your Kubernetes RBAC allows: only people who can port-forward in the `resourcerer` namespace
reach the UI.

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

## Read-only for a wide audience

With `readOnly: true` the server refuses every change and its RBAC drops the `patch` verbs, so the UI can be shared
widely as a report, and changes go through Git from the **Patches** tab. `readOnly` applies to the whole release: run
one resourcerer per cluster.

## The API

The UI's HTTP API is served on the same Service, under `/api/v1`, with the same (lack of) authentication. See
[HTTP API](../reference/http-api.md).

## What is coming

OIDC login, a trusted-header mode for proxies that pass the user's identity, and per-user authorization are designed in
[#33](https://github.com/gjbravi/resourcerer/issues/33). The history will then record who applied, and apply policies
will be able to let some people override a window.
