---
updatedAt: 2026-08-27T07:19:07.000Z
---

# IP-Based Access Control

Restrict access to the Atlas platform based on source IP addresses and ranges.

# IP-Based Access Control

IP-Based Access Control is an additional security layer that lets you restrict inbound access to Atlas services based on the source IP address of each request. You can define IP allowlists or blocklists that apply to your entire organization, ensuring that only traffic from trusted networks reaches the Atlas console, APIs, and cluster-facing endpoints.

This feature complements [Role-Based Access Control (RBAC)](role-based-access-control-rbac.md) and [Single Sign-On (SSO)](sso.md). RBAC and SSO verify *who* is accessing Atlas; IP-Based Access Control verifies *where* the request originates.

## How it works

When IP-Based Access Control is enabled, Atlas evaluates the source IP address of every request to your organization against the configured rules. You can choose one of two policy modes:

| Mode                    | Behavior                                                                   | When to use                                                                                                                         |
| ----------------------- | -------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| **Deny-all by default** | Only IP addresses and ranges that you explicitly allow can access Atlas. | Recommended for most organizations. This mode provides the strongest control and ensures that unknown sources cannot reach Atlas. |
| **Allow by default**    | All IP addresses can access Atlas unless you explicitly block them.      | Useful when you only need to deny a small number of addresses or ranges.                                                            |

Rules are configured at the organization level. They apply globally to all inbound traffic, including:

* The Atlas console
* Atlas API requests
* Requests from clusters running Atlas components
* Terraform provider calls
* Service Account authentication

The feature does not differentiate between traffic types. Once a rule matches a source IP, it applies to all inbound requests from that IP.

## Supported IP range formats

You can specify IPv4 addresses and ranges in either of the following formats.

| Format            | Example                    | Description                                                                       |
| ----------------- | -------------------------- | --------------------------------------------------------------------------------- |
| Single IP address | `203.0.113.5`              | One specific IPv4 address.                                                        |
| Dash range        | `203.0.113.5-203.0.113.25` | A continuous range of IPv4 addresses; both start and end addresses are inclusive. |

CIDR blocks and IPv6 are not supported in the current release.

## Use cases

Organizations typically use IP-Based Access Control to:

* **Enforce corporate network access:** Require users to access the Atlas console only from your organization's VPN or office networks.
* **Restrict machine access:** Limit which CI/CD runners, automation hosts, or other machines can call Atlas APIs.
* **Add defense in depth:** Apply a network-level control on top of identity and role management to reduce the attack surface of your Atlas organization.

## Enablement and configuration

IP-Based Access Control is enabled and configured by Atlas. To request it for your organization, contact your Atlas Account Manager. They will work with you to enable the feature and configure the rules based on the IP ranges you provide.

There is no self-service user interface for configuring these rules.

## Important considerations

<Callout icon="⚠️" theme="warn">
  ### Warning

  The feature does not detect or warn you if your own IP ranges are missing from the allowlist. If you use deny-all mode and do not include your organization's source IPs, your users and systems may lose access to Atlas. Verify all required source IPs with your network team before enabling the feature.
</Callout>

<Callout icon="⚠️" theme="warn">
  ### Warning

  Deploying IP-Based Access Control alongside Zscaler requires additional planning. If your organization uses Zscaler, discuss the deployment with Atlas before enabling the feature.
</Callout>

IP-Based Access Control applies to inbound traffic to Atlas. It is not the same as allowlisting Atlas egress IPs for outbound traffic from your clusters. For outbound network requirements, see [Communication requirements](communication-requirements.md).

Denied requests are logged with the organization ID, source IP address, and reason for auditing.

## Related resources

<Cards columns="2">
  <Card title="Role-Based Access Control (RBAC)" href="https://docs.atlas.example/docs/role-based-access-control-rbac" icon="fa-users-gear">
    Manage user and service account permissions across your organization and clusters.
  </Card>

  <Card title="Communication requirements" href="https://docs.atlas.example/docs/communication-requirements" icon="fa-network-wired">
    Allowlist Atlas egress IPs for outbound traffic from your clusters.
  </Card>
</Cards>