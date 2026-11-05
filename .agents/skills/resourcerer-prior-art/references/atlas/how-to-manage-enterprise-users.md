---
updatedAt: 2026-08-28T06:44:25.000Z
---

# Manage Enterprise users

Invite users, assign roles, manage access, and configure user groups across your Atlas Enterprise and its child organizations.

## Overview

This guide explains how to manage users across a Atlas Enterprise. From the Enterprise, you can invite users, assign roles at the Enterprise or child organization level, and manage user groups — all from a single interface. Child organization Owners can also manage users, roles, and user groups within their own organization.

## Before you start

* You must have the **Enterprise Owner** role to invite users, assign roles, and revoke access. Enterprise Viewers can view user lists but cannot make changes.
* Open **Settings** by selecting the Enterprise in the organization switcher, or by expanding the left sidebar and selecting **Settings** at the bottom.
* In the **View by** dropdown at the top of the left sidebar, select the Enterprise.

## View Enterprise users

1. In the left sidebar, select **Users**.

   The paginated table displays all users with access to the Enterprise or any of its child organizations, including their authentication method, email, organization tags, and last activity:

   | Column        | Description                                         |
   | ------------- | --------------------------------------------------- |
   | Name          | User's display name                                 |
   | Email         | User's email address                                |
   | Auth          | Authentication method (SSO, email/password)         |
   | Organization  | Organizations the user has access to, shown as tags |
   | Last activity | Date of the user's most recent activity             |

   ![](https://files.readme.io/9effb298bc88f6cf18b79b940ecce58a3064698a75c98c19765298d2e11efe33-image.png)
2. Click a user's name to view their detailed profile, including:
   * Roles assigned per child organization
   * User groups the user belongs to

![](https://files.readme.io/38d5671a5f38df25c2f12fac3787000167c4d5f3febae17e6e89e46cc63b2567-image.png)

## Invite users

1. In the left sidebar, select **Users**.

2. Click **Invite users**.

3. Select the scope for the invitation — either Enterprise or child organization(s):

   * **Enterprise** — grants the user a role that spans the Enterprise and all child organizations. Selecting the Enterprise automatically grants access to all child organizations.
   * **Child organization(s)** — grants the user a role within one or more specific child organizations.

<Image src="https://files.readme.io/29fdd45220f4bb014dbb85dea752b78c80ecbe6e43b1fda23281d2f6c625ee36-image.png" align="center" />

<Image src="https://files.readme.io/347a92152d48bfdae2622f1cbdb912521f454dc5c0c7ba78d65deebaefed5583-image.png" align="center" />

4. Select the role to assign:

   **Enterprise roles:**

   | Role                  | Access level                                                                                    |
   | --------------------- | ----------------------------------------------------------------------------------------------- |
   | **Enterprise Owner**  | Full access to all features and organization management for parent and child organizations      |
   | **Enterprise Viewer** | Read-only access to all features and organization management for parent and child organizations |

   **Child organization roles:**

   | Role        | Access level                                                             |
   | ----------- | ------------------------------------------------------------------------ |
   | **Owner**   | Full access to clusters, billing, and organization management            |
   | **Member**  | Access to clusters, read-only access to billing                          |
   | **Viewer**  | Read-only access to clusters and billing                                 |
   | **Analyst** | Full access to cost monitoring, read-only access to clusters and billing |

![Enterprise role selection showing Enterprise Owner and Enterprise Viewer options](https://files.readme.io/0cf0ccddb25a55497e31e18b3bb9852235825c3f1ab6eea322f5b3b43c755882-image.png)

![Child organization role selection showing Owner, Member, Viewer, and Analyst options](https://files.readme.io/22bfd3da5610fdb8ea44468b65834e4219f0672792cfb7fd5d7aab183b2826d6-image.png)

5. Enter the user's email address (or multiple addresses separated by commas).

6. Click **Invite**.

## Cancel or resend invitations

If a user has not yet accepted their invitation, you can cancel or resend it:

1. In the left sidebar, select **Users**.

2. Find the user with a pending invitation.

3. Click **Resend invite** to send the invitation again, or **Cancel invite** to revoke the pending invitation.

![Pending invitation with Resend invite and Cancel invite action buttons](https://files.readme.io/0ba2d39afd11bce62eb29f6de8848615e5e0a9f7b502588b6227ed9fa7c5a89c-image.png)

## Revoke Enterprise access

1. In the left sidebar, select **Users**.

2. Click **Remove access** next to the user whose access you want to revoke.

3. Choose one of two options:

   * **Revoke all access** — removes the user's access from the Enterprise and all child organizations. Use the button at the bottom-left corner of the dialog.
   * **Revoke Enterprise access only** — removes Enterprise-level roles while allowing you to set child-organization-specific roles. Select new roles from the dropdown for each child organization, then click **Apply changes**. Organizations where no role is displayed mean the user does not currently have an organization-level role assigned.

![Revoke Enterprise access dialog with options to revoke all access or set child organization roles](https://files.readme.io/0d18bf370c753ea28fca36d88fec5b1a300b84e8298fb2ee11e31d08d0aa6036-image.png)

## Manage user groups

User groups let you define access policies for sets of users with similar needs. In an Enterprise, user groups can be scoped to the Enterprise level or to individual child organizations.

### View user groups

1. In the left sidebar, select **User groups**.

   The table displays all user groups with their associated organization and member count.

![](https://files.readme.io/69804ceefde5730dc5fa5ac90655ab55da5dd76f4dffa819a741b223d0a11781-image.png)

### Create a user group

1. In the left sidebar, select **User groups**.

2. Click **Create user group**.

3. Select the organization scope — the Enterprise or a specific child organization.

![Create user group dialog with organization scope selection](https://files.readme.io/0037feccfbf9ded7e57dc499d24001a84d34faf4e939c54c768d8987b7023bb8-image.png)

4. Configure the group name, role bindings, and optionally add users.

5. Click **Create**.

### Edit a user group

1. In the left sidebar, select **User groups**.

2. Click the user group name to open its details drawer.

   The drawer shows:

   * **Role bindings** — the roles and resource scopes assigned to the group.
   * **Organization** — the organization the group is scoped to.
   * **Users** — the users assigned to the group. You can add or remove users from this tab.

   ![](https://files.readme.io/07edf896fa8055f9f11ed255e04714e365e71f88e05256cc34770d71dbe5da35-image.png)

3. Click a role binding row to view its details, including the resource type, scope, and assigned role.

![Role binding details panel showing resource type, access scope, and assigned role](https://files.readme.io/487fdf4a36aa82980329671951569e43e696b94ac3a19c12e918c4a9640c54ed-image.png)

For more details on user group configuration and management, see [User Groups](user-groups.md).

## See also

<Cards columns="2">
  <Card title="Enterprise organizations" href="https://docs.atlas.example/docs/enterprise-organizations" icon="fa-building">
    Learn how the Enterprise hierarchy, roles, and API access work.
  </Card>

  <Card title="User Groups" href="https://docs.atlas.example/docs/user-groups" icon="fa-users">
    Create and configure user groups with role bindings and resource scopes.
  </Card>
</Cards>