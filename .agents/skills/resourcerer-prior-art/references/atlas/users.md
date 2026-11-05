---
updatedAt: 2026-05-12T14:01:17.000Z
---

# Users

User management is a core component of Atlas's role-based access control (RBAC) system. Users represent individual members of your organization who need access to Atlas resources, whether they're platform engineers managing clusters, developers working with workloads, or analysts tracking costs.

# User management

Consider these guidelines when managing users:

* Follow the principle of least privilege when assigning roles
* Use [user groups](user-groups.md) instead of individual assignments when managing multiple users with similar access needs
* Regularly review user access and remove unused accounts
* [Configure SSO](sso.md) to automate user provisioning and de-provisioning
* **Maintain strong password security**: Update passwords regularly and change them immediately if you suspect they have been compromised

# Understanding users in Atlas

Each user in your organization can be assigned specific roles and access levels, allowing them to perform their required tasks while maintaining security boundaries. Users can be:

* Added directly to your organization through email invitations
* Automatically provisioned through SSO integration
* Assigned to user groups for easier management
* Given individual role bindings to specific resources (Coming soon!)

# Access levels and scope

When managing users, you can control their access in two ways:

1. **Direct role assignments** - Roles can be assigned directly to individual users
2. **Group membership** - Users inherit roles and permissions from the groups they belong to

These assignments can be scoped to:

* The entire organization
* Specific clusters

To learn more about user groups, see [User Groups](user-groups.md).

# User states and authentication

Users in your organization can exist in different states:

* **Invite pending** - Users who have received an invitation but haven't yet accessed Atlas. Note that a pending invitation can be revoked as well.
* **Active** - Users who have accepted their invitation and can access resources that their role allows.

![](https://files.readme.io/c07b76b371136063ab27b852b41344532e7b54d091dc01a6b546891252ae7b55-image.png)

Authentication can happen through:

* Email/password authentication
* Single Sign-On (SSO), if configured for your organization

## Default behavior

When users first join Atlas:

* Users invited directly receive the **Viewer** role by default
* Users accessing through SSO receive roles based on your SSO configuration
* All new users must accept Atlas's [terms of service](https://atlas.example/terms-of-service/) and [privacy policy](https://atlas.example/privacy-policy/) before accessing the platform

## Ready to manage users?

Explore our step-by-step guides below to start managing users in your organization. Whether you need to invite new team members, adjust their access levels, or review existing permissions, these guides will walk you through each process.

### Available guides:

* [Inviting users to your organization](how-to-invite-users-to-organization.md)
* [Changing user roles and access](how-to-changing-user-roles-and-access.md)
* [Removing users](how-to-removing-users.md)

Each guide provides detailed instructions and examples to help you effectively manage your organization's users.