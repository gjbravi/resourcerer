---
updatedAt: 2026-08-27T12:24:59.000Z
---

# Organization management

Learn how to manage your organization settings, user roles, and access controls safely in Atlas.

Managing your organization settings and user access in Atlas requires careful consideration of security implications and potential operational impacts. This guide covers important aspects of organization management and best practices for maintaining secure, uninterrupted cluster operations.

# Enterprise organizations

If your organization uses a multi-organization structure, Atlas's Enterprise model lets you manage multiple child organizations from a centralized workspace. Users, roles, SSO, and user groups can be managed at the Enterprise level or within each child organization. Each child organization retains its own clusters and billing.

<Cards columns={2}>
  <Card title="Enterprise organizations" href="https://docs.atlas.example/docs/enterprise-organizations" icon="fa-building">
    Learn how the Enterprise hierarchy, roles, and API access work.
  </Card>

  <Card title="Manage child organizations" href="https://docs.atlas.example/docs/how-to-manage-child-organizations" icon="fa-sitemap">
    Create, edit, and switch between child organizations.
  </Card>

  <Card title="Manage Enterprise users" href="https://docs.atlas.example/docs/how-to-manage-enterprise-users" icon="fa-users">
    Invite users, assign roles, and manage user groups across the Enterprise.
  </Card>
</Cards>

# User roles

Atlas uses role-based access control (RBAC) to manage permissions within your organization. Each user can be assigned specific roles that determine what actions they can perform and which resources they can access.

Learn more about:

* [Understanding RBAC in Atlas](role-based-access-control-rbac.md)
* [Managing users and roles](users.md)
* [Creating and managing user groups](user-groups.md)
* [Working with service accounts](service-accounts.md)

# Add new members

Growing your team? Atlas provides multiple ways to add and manage users in your organization, from direct invitations to automated provisioning through SSO.

Find detailed guides on:

* [Inviting users](how-to-invite-users-to-organization.md)
* [Setting up user groups](how-to-creating-and-configuring-user-groups.md)

# Managing API Keys

API keys enable programmatic access to Atlas resources. Proper API key management is crucial for security when setting up automation tools, CI/CD pipelines, or custom integrations.

Explore our guides on:

* [Creating service accounts](how-to-creating-service-accounts.md)
* [Managing API key access](how-to-managing-service-accounts.md)

# Additional organization settings

Beyond user and access management, Atlas offers several organization-wide configurations:

* [Configure SSO integration](sso.md)
* [Manage notification preferences](notifications.md)

Each of these areas has dedicated documentation with detailed instructions and best practices.