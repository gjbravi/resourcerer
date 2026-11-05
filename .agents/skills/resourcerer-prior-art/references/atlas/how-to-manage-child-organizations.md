---
updatedAt: 2026-08-27T12:24:59.000Z
---

# Manage child organizations

Create, edit, and navigate between child organizations within your Atlas Enterprise.

## Overview

This guide explains how to manage child organizations within a Atlas Enterprise. Child organizations are independent workspaces where clusters are connected and managed. As an Enterprise Owner, you can create new child organizations, edit existing ones, and switch between them.

## Before you start

* You must have the **Enterprise Owner** role. Enterprise Viewers can view the child organizations list but cannot create or edit organizations.
* Open **Settings** by selecting the Enterprise in the organization switcher, or by expanding the left sidebar and selecting **Settings** at the bottom.
* In the **View by** dropdown at the top of the left sidebar, select the Enterprise.

## View child organizations

1. In the left sidebar, select **Organizations**.

   The table lists all child organizations in the Enterprise:

   | Column             | Description                                              |
   | ------------------ | -------------------------------------------------------- |
   | Child organization | Organization name                                        |
   | Users              | Number of users with access to the organization          |
   | User groups        | Number of user groups configured for the organization    |
   | Role bindings      | Number of role bindings assigned within the organization |
   | Clusters           | Number of clusters connected to the organization         |
   | Created            | Date the organization was created                        |

2. Click an organization name or a value in the Users, User groups, Role bindings, or Clusters column to open the corresponding page in that child organization.

## Create a child organization

1. In the left sidebar, select **Organizations**.

2. Click **Create child organization**.

3. Enter the organization name and configure any required settings.

4. Click **Create**.

You can also create a child organization from the organization switcher:

1. Click the organization switcher in the top navigation bar.

2. Click **Create organization**.

3. Choose **Create child organization** for the current Enterprise. The other option creates a standalone organization outside the Enterprise.

4. Enter the organization name and configure any required settings.

5. Click **Create**.

## Edit a child organization

1. In the left sidebar, select **Organizations**.

2. Click **⋯** next to the child organization you want to modify.

3. Update the organization name or settings as needed.

4. Save your changes.

## Navigate to a child organization

Use either of these methods:

* **Organization switcher** — click the organization switcher in the top navigation bar and select the child organization. You stay on the same page type (for example, switching while on the Cluster List opens the Cluster List of the selected organization).
* **View by** — in Settings, open the **View by** dropdown at the top of the left sidebar and select the child organization.

To return to the Enterprise, select the Enterprise in the organization switcher or in **View by**.

## See also

<Cards columns={2}>
  <Card title="Enterprise organizations" href="https://docs.atlas.example/docs/enterprise-organizations" icon="fa-building">
    Learn how the Enterprise hierarchy, roles, and API access work.
  </Card>

  <Card title="Role-Based Access Control" href="https://docs.atlas.example/docs/role-based-access-control-rbac" icon="fa-shield-alt">
    Understand Enterprise and organization roles, permissions, and resource scopes.
  </Card>
</Cards>

<br />