---
updatedAt: 2026-08-28T07:09:04.000Z
---

# Service Accounts

Service accounts in Atlas enable automated, non-user access to platform resources. Unlike regular user accounts that represent human users, service accounts are designed for machine-to-machine communication, making them ideal for automation tools, CI/CD pipelines, and third-party integrations.

# Understanding service accounts

Service accounts are special identities in Atlas RBAC designed for programmatic access to the platform. Unlike regular user accounts that represent human team members, service accounts operate independently and aren't tied to any individual person.

Each service account functions as a distinct identity that can generate its own API keys for secure authentication. Regarding permissions and access control, service accounts work just like regular users - they can be assigned specific roles and even added to user groups to streamline permission management. This flexibility allows you to treat service accounts as full members of your organization while maintaining clear separation from human users.

This design makes service accounts ideal for scenarios where automated systems need to interact with Atlas, whether through infrastructure automation tools, CI/CD pipelines, or other integrations.

## Common use cases

Service accounts are typically used for:

**Infrastructure as Code**\
Create service accounts for Terraform or other IaC tools to automate cluster management and configuration.

**CI/CD Pipelines**\
Enable automated deployments and cluster operations from your CI/CD systems.

**Monitoring and Analytics**\
Allow external monitoring tools to collect data from Atlas.

When you use a service account and its associated API key for any of the aforementioned use cases, you can be sure that changes in your team will not negatively impact any integrations, which must remain operational at all times.

**Using service accounts for cluster onboarding**\
Service accounts with appropriate permissions can be used to onboard clusters and manage Atlas components.

To use a service account for cluster onboarding:

1. Create a service account with the **Owner** role at the organization level
2. Generate an API key for the service account
3. Use this API key when configuring Atlas components like:
   * The Atlas agent
   * Cluster controller
   * Other Atlas services that require authentication

> 📘 Note
>
> Using service accounts for cluster onboarding ensures that API access remains uninterrupted even when individual users leave the organization or change roles.

## Service accounts in Enterprise organizations

If your organization uses [Enterprise organizations](enterprise-organizations.md), a service account can be scoped to the entire Enterprise or to one or more child organizations.

Enterprise service accounts accept **one role only**. That role is either read or write across every resource the account can access. You cannot mix permission levels on the same service account — for example, read access to one child organization and write access to another. This differs from organization-level service accounts, where a single account can have different roles on different clusters (read on Cluster A, write on Cluster B).

If you need different permission levels, create separate service accounts.

### API keys in Enterprise organizations

API key scoping determines which resources a key can access:

* **Child organization API key** — works without any additional configuration. The key is already scoped to the correct child organization.
* **Parent organization API key** — to access resources in a child organization, include the `X-AtlasAi-Organization-Id` header with the child organization's ID. Without this header, the API resolves the request against the parent organization and returns `404` for child organization resources.

```shell
curl --request GET \
  --url https://api.atlas.example/v1/<endpoint> \
  --header 'X-API-Key: <parent-org-api-key>' \
  --header 'X-AtlasAi-Organization-Id: <child-org-id>' \
  --header 'accept: application/json'
```

This applies to all API endpoints — it is not specific to a single endpoint. Any request that targets a resource (cluster, workload, security data, etc.) in a child organization requires the header when authenticating with a parent organization API key.

## Security considerations

When working with service accounts:

* Follow the principle of least privilege when assigning roles
* Remove unused service accounts and revoke unnecessary access
* Consider using user groups to manage service account permissions

# Ready to manage service accounts?

Explore our guides to learn how to:

* [Create service accounts](how-to-creating-service-accounts.md)
* [Manage existing service accounts](how-to-managing-service-accounts.md)