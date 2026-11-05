---
updatedAt: 2026-09-10T11:37:49.000Z
---

# Enterprise organizations

Centrally manage multiple child organizations, users, and access controls from a single Enterprise workspace in Atlas.

An Enterprise in Atlas introduces a hierarchical structure where a parent workspace — the Enterprise — manages multiple child organizations. Each child organization operates independently with its own clusters, billing, and cost reports. Users, roles, SSO, and user groups can be managed centrally at the Enterprise level, and each child organization can also manage them within its own scope.

* **Organization** — a single workspace where clusters and users are managed directly.
* **Enterprise** — a parent workspace that contains multiple **child organizations**, each with its own clusters. Access can be managed from the Enterprise, from within a child organization, or both.

The Enterprise itself does not have clusters and does not expose organization-level features such as cluster management or cost monitoring. All clusters live within child organizations.

![Enterprise architecture diagram](https://files.readme.io/3f8acce2860bb5628639106bb6ed0f2b0e731dae6909c13526023c1c6caec90f-enterprise-architecture-diagram.png)

## How Enterprise organizations work

An Enterprise enables centralized visibility and control while preserving autonomy and isolation for individual business units:

* **Centralized access management** — users, roles, SSO connections, and user groups can be configured at the Enterprise level and cascade to child organizations. You can also [synchronize user groups from your identity provider](idp-user-group-sync.md) so group membership and access stay in sync with your IdP. IdP user and group sync is also available for child organizations if multi-tenant SSO is enabled
* **Child organization autonomy** — each child organization maintains its own clusters, billing data, and cost reports. Teams work independently and can manage access within their own scope: remove users, change a user's role in that child organization, and create user groups. When multi-tenant SSO is enabled, a child organization can also configure its own SSO connection.
* **Consolidated reporting** — features like [Platform usage](administration-platform-usage.md) and [Enterprise commitments](enterprise-commitments.md) aggregate data across all child organizations.

## Enterprise navigation

### Organization switcher

The organization switcher in the top navigation lists the Enterprise and its child organizations:

* Click the Enterprise name to open **Settings** for the Enterprise. The left sidebar then includes **Organizations**, which only the Enterprise has.
* Click a child organization to switch to it. You stay on the same page type (for example, switching while on the Cluster List opens the Cluster List of the selected organization).
* Click **Create organization** to create a standalone organization (outside the Enterprise) or a child organization within the Enterprise.
* **Manage Enterprise** appears only when you are in the Enterprise or one of its child organizations, and only for users with the Enterprise Owner or Enterprise Viewer role.

<Callout icon="📘" theme="info">
  ### Note

  Child organization Owners and Viewers see **Manage Org** / **View Org** instead of **Manage Enterprise**, and cannot access Enterprise-level data.
</Callout>

### Settings

You can also open Settings by expanding the left sidebar and selecting **Settings** at the bottom.

At the top of the Settings sidebar, the **View by** dropdown shows which organization you are managing. With the Enterprise selected, the sidebar lists:

* **Overview**
* **Organizations** — child organizations in this Enterprise
* **Users**
* **User groups**
* **Single sign-on**
* **Platform usage**

Select a child organization in **View by** to manage that organization instead.

## Enterprise roles

Enterprise organizations use two additional roles that span the parent and all child organizations:

| Role                  | Access level                                                                                                    |
| --------------------- | --------------------------------------------------------------------------------------------------------------- |
| **Enterprise Owner**  | Full access to all product features, including organization management for parent and child organizations.      |
| **Enterprise Viewer** | Read-only access to all product features, including organization management for parent and child organizations. |

Enterprise roles are assigned at the Enterprise level. Standard organization roles (Owner, Member, Viewer, Analyst) still apply within individual child organizations. For the full permissions matrix, see [Role-Based Access Control (RBAC)](role-based-access-control-rbac.md).

## API access in Enterprise setups

When using a parent organization API key to access resources that live in a child organization, include the `X-AtlasAi-Organization-Id` header with the child organization's ID:

```shell
curl --request GET \
  --url https://api.atlas.example/v1/<endpoint> \
  --header 'X-API-Key: <parent-org-api-key>' \
  --header 'X-AtlasAi-Organization-Id: <child-org-id>' \
  --header 'accept: application/json'
```

Without this header, the API resolves the request against the parent organization and returns `404` for resources that belong to a child organization.

API keys scoped to a specific child organization work without the header — they are already scoped to the correct organization.

For more details on API keys and service accounts, see [Service Accounts](service-accounts.md).

## Set up an Enterprise

Enterprise organizations cannot be created in the Atlas console. Contact your Atlas account representative.

To set up a new Enterprise, create a regular organization in the console first. Atlas converts that organization into an Enterprise. The person who created it becomes **Enterprise Owner**. You can rename the organization later.

Atlas can also attach your existing organizations as children. When an organization is attached:

* Its users are moved to the Enterprise.
* Its clusters stay in that organization. Clusters cannot be transferred automatically.

If those organizations already have SSO:

* **One organization with SSO** — that connection can be moved to the Enterprise and used as the centralized SSO.
* **Several organizations with SSO** — by default the Enterprise supports one SSO connection. Choose which connection to move to the Enterprise. The others inherit it unless multi-tenant SSO is enabled.
* **No SSO** — confirm whether all users share one identity provider. If subsidiaries or departments use different IdPs, ask Atlas to enable multi-tenant SSO.

After the Enterprise exists, you can create child organizations, invite users, and configure access. See [Manage child organizations](how-to-manage-child-organizations.md) and [Manage Enterprise users](how-to-manage-enterprise-users.md).

## Limitations

* Clusters cannot be transferred between child organizations. To move a cluster, disconnect it from one child organization and connect it to another.
* By default, only one SSO connection can be configured per Enterprise. Child organizations inherit the Enterprise connection until a child configures its own, which requires multi-tenant SSO. Contact your Atlas account representative to enable this — it is not self-service. See [SSO in Enterprise organizations](sso.md#sso-in-enterprise-organizations).
* Enterprise Owner and Viewer roles grant access to **all** child organizations — there is no built-in way to restrict an enterprise-scoped user to a subset of child orgs. To limit a user's visibility to specific child organizations only, do not assign them an Enterprise-level role. Instead, invite them directly to each child organization they should access.
* Enterprise organizations are not intended for separating development, staging, and production environments. Use [user groups](user-groups.md) with cluster-scoped roles to control access by environment instead.

## Related resources

<Cards columns={2}>
  <Card title="Role-Based Access Control" href="https://docs.atlas.example/docs/role-based-access-control-rbac" icon="fa-shield-alt">
    Understand Enterprise and organization roles, permissions, and resource scopes.
  </Card>

  <Card title="Single Sign-On (SSO)" href="https://docs.atlas.example/docs/sso" icon="fa-key">
    Configure SSO at the Enterprise level, or on child organizations when multi-tenant SSO is enabled.
  </Card>

  <Card title="IdP user group sync" href="https://docs.atlas.example/docs/idp-user-group-sync" icon="fa-sync">
    Synchronize user groups from your identity provider for automated access control.
  </Card>
</Cards>

<br />