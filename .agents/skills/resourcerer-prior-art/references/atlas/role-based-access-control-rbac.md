---
updatedAt: 2026-09-22T13:03:06.000Z
---

# Role-Based Access Control (RBAC)

# Understanding RBAC in Atlas

Role-based access control (RBAC) in Atlas enables fine-grained access management for your organization's resources. It allows you to control who can access what within your organization and clusters, ensuring security and operational efficiency.

## Roles and permissions

Atlas uses predefined roles to streamline access management. Each role comes with a specific set of permissions designed for common use cases:

* **Owner** role has full administrative access to all organization and cluster resources. This role is typically assigned to platform team members who manage the entire organization.

* **Member** role can perform all cluster-related operations but has limited access to organizational features like billing (read-only) or adding more users. This role suits DevOps engineers and cluster administrators.

* **Viewer** role provides read-only access across all features, making it perfect for team members who need visibility but shouldn't make changes.

* **Analyst** role specializes in cost monitoring, with full access to cost-related features while maintaining read-only access elsewhere. This role is designed for FinOps teams and financial analysts.

* **Schedule Manager** role can manage rebalancing and hibernation schedules across the organization, with no other permissions. This role suits users or service accounts that automate cost-optimization schedules (for example, via the API, Terraform, or Crossplane) without full organization Owner access. See [Schedule Manager scope and coverage](#schedule-manager-scope-and-coverage).

| Capability                       | Owner | Member    | Analyst   | Viewer    | Schedule Manager |
| -------------------------------- | ----- | --------- | --------- | --------- | ---------------- |
| Connecting clusters              | ✓     | -         | -         | -         | -                |
| Cluster management               | ✓     | ✓         | View only | View only | -                |
| Hibernation schedule management  | ✓     | -         | -         | -         | ✓                |
| Rebalancing schedule management  | ✓     | -         | -         | -         | ✓                |
| Rebalancing job management       | ✓     | -         | -         | -         | ✓                |
| Manual rebalancing (plans)       | ✓     | -         | -         | -         | ✓                |
| Billing access                   | ✓     | View only | View only | View only | -                |
| Cost monitoring                  | ✓     | ✓         | ✓         | View only | -                |
| Organization management          | ✓     | -         | -         | -         | -                |
| Invite/remove members            | ✓     | -         | -         | -         | -                |
| Service account API key creation | ✓     | -         | -         | -         | -                |

### Schedule Manager scope and coverage

The Schedule Manager role's permissions, as shown in the Console under **View permissions** when configuring a role binding:

```text
clusterOptimization.hibernationSchedules.create
clusterOptimization.hibernationSchedules.delete
clusterOptimization.hibernationSchedules.read
clusterOptimization.hibernationSchedules.write
clusterOptimization.rebalancing.create
clusterOptimization.rebalancing.delete
clusterOptimization.rebalancing.read
clusterOptimization.rebalancing.write
```

These are the role's only permissions. What they cover depends on the [scope](#resource-scope) of the role binding.

**Organization scope** — required for managing schedules. Rebalancing and hibernation schedules are organization-scoped objects that can be assigned to multiple clusters (see [Scheduled rebalancing](scheduled-rebalancing.md#setup)), so no cluster-scoped binding can manage them, even with the Owner role:

* Create, read, update, and delete rebalancing schedules (see [Scheduled rebalancing](scheduled-rebalancing.md))
* Create, read, update, and delete hibernation schedules (see [Cluster hibernation](autoscaling-cluster-hibernation.md))

**Cluster scope** — for the assigned clusters only:

* Rebalancing jobs: link a rebalancing schedule to a cluster so the schedule runs on it, and view or remove those links
* Manual rebalancing: generate, execute, and cancel rebalancing plans (see [Rebalancing](rebalancing.md))

**Not covered at any scope:** hibernation pause/resume operations (the separate `clusterOptimization.clusterHibernationJobs` permission resource, held by Owner).

### Enterprise roles

If your organization uses [Enterprise organizations](enterprise-organizations.md), two additional roles are available that span the parent Enterprise and all child organizations:

* **Enterprise Owner** — full access to all product features, including organization management for the parent Enterprise and all child organizations.
* **Enterprise Viewer** — read-only access to all product features, including organization management for the parent Enterprise and all child organizations.

Enterprise roles are assigned at the Enterprise level. Standard organization roles (Owner, Member, Viewer, Analyst) still apply within individual child organizations. A user can hold an Enterprise role and a child organization role simultaneously.

<Callout icon="💡" theme="default">
  ### Tip

  Need to find out who in your organization has a specific role? Go to **View Org → Access control → Resources → Organization** to see all users and their current role assignments.
</Callout>

## Resource scope

When assigning roles, you can control their scope of access:

* **Organization scope** grants access to **all clusters** and organization-wide features like billing and SSO configuration. This broad access is suitable for platform teams and administrators managing multiple clusters.

* **Cluster scope** restricts access to specific clusters. This granular control allows you to limit team members' access to only the clusters they need to work with, following the principle of least privilege.

* **Enterprise scope** grants access across all child organizations' resources. This scope is available only with Enterprise roles (Enterprise Owner, Enterprise Viewer) and is assigned at the Enterprise level. See [Enterprise organizations](enterprise-organizations.md).

## User Groups

User Groups simplify access management at scale by allowing you to:

* Group users with similar access needs under a single policy
* Manage permissions at the group level instead of individual users
* Automatically handle access through your Identity Provider (IdP) groups (**Coming soon!**)
* Apply consistent access policies across teams

For example, you might create separate groups for your platform team, application developers, and financial analysts, each with appropriate roles and resource access.

## Service Accounts

Service Accounts enable automated access to Atlas resources, which is particularly useful for:

* CI/CD pipelines requiring programmatic access
* Infrastructure as Code tools like Terraform
* Automation scripts and system integrations

Each Service Account can:

* Have its own set of roles and permissions
* Generate API keys for secure authentication
* Be managed independently of user accounts

Service Accounts follow the same RBAC model as user accounts, ensuring consistent security policies across both human and automated access.

> 📘 Note
>
> When you delete a user from your organization, all API keys associated with that user get deleted as well. Employing Service Accounts and using their API keys instead to ensure consistent, uninterrupted functionality allows you to circumvent many potential issues.

## Managing RBAC programmatically

Atlas's RBAC system can be managed not only through the console but also programmatically through our API and Terraform provider.

### API Management

All RBAC operations are available through the Atlas API, enabling you to:

* Create and manage users and service accounts
* Configure user groups and their permissions
* Assign and modify role bindings

For detailed API documentation, visit our [API Reference](https://docs.atlas.example/reference/rbacserviceapi_creategroup).

### Terraform Provider

The Atlas Terraform provider offers comprehensive RBAC management capabilities, allowing you to:

* Define and maintain RBAC configurations as code
* Automate user and group management
* Ensure consistent access patterns across organizations

Example Terraform configuration for creating a service account:

```shell
resource "atlas_service_account" "service_account" {
  organization_id = organization.id
  name            = "example-service-account"
  description     = "service account description"
}

resource "atlas_service_account_key" "service_account_key" {
  organization_id    = data.atlas_organization.test.id
  service_account_id = atlas_service_account.service_account.id
  name               = "example-key"
  active             = true
  expires_at         = "2025-01-01T00:00:00Z"
}

output "service_account_key" {
  value = atlas_service_account_key.service_account_key.token
}
```

For more information and examples, see our [Terraform Provider Documentation](terraform-provider.md).

# Next steps

[Configure Service Accounts →](service-accounts.md)

[Manage User Groups →](user-groups.md)

[Learn about SSO Integration →](sso.md)