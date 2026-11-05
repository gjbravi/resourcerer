---
updatedAt: 2026-05-12T14:01:17.000Z
---

# User Groups

User groups in Atlas help you manage access for multiple users with similar roles and responsibilities. Instead of managing permissions individually, you can assign roles and resource access to groups, making it easier to maintain consistent access policies across your organization.

> 📘 Important
>
> Only users with the **Owner** role can create and manage user groups. This ensures centralized control over group-based access management.

# How user groups work

When you create a user group, you:

* Define what role the group will have
* Specify which resources (organization or specific clusters) members can access
* Add users to the group who need those permissions

**Users inherit all permissions from their groups**. A user can belong to multiple groups, and their effective permissions follow these principles:

* All group memberships contribute to the user's total access
* When groups provide access to the same resource, **the highest permission level takes precedence**
* Organization-level permissions override cluster-level permissions

For example, if a user belongs to two groups:

* "Production Viewers" group with Viewer access to production clusters
* "Platform Team" group with Owner access to all resources

The user will have Owner access to production clusters (and all other resources) because the higher-level permission from the Platform Team group overrides the Viewer role from the Production Viewers group.

# Ready to start managing groups?

Explore our guides to learn how to:

* [Create and configure user groups](how-to-creating-and-configuring-user-groups.md)
* [Manage existing user groups](how-to-managing-existing-user-groups.md)