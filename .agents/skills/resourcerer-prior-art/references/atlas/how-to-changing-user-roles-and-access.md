---
updatedAt: 2026-05-12T14:12:41.000Z
---

# How-to: Changing user roles and access

Changing a user's role allows you to modify their access level and permissions within your organization. This guide walks you through changing roles and explains important considerations.

# Before you begin

Before changing a user's role, ensure that:

* You have **Owner** permissions in your organization
* You've reviewed any API keys associated with the user
* You understand the implications of the new role's permissions

> 🚧 Warning: API Key Impact
>
> Changing a user's role will automatically invalidate all API keys associated with that user. This can lead to:
>
> * Service disruptions and cluster downtime
> * Break automated processes and CI/CD pipelines
> * Disconnect third-party integrations
>
> Always review and replace affected API keys before proceeding with role changes.

# Changing a role

To change a user's role:

1. **Navigate to the organization settings**

   * Select "Manage organization" at the top of the Atlas console:

     ![](https://files.readme.io/62c8d8460ec41b2dacd7162b09576f5f692aef11b30ee36eb1228a15e676c468-image.png)

2. **Find the user**

   * In the organization profile view, locate the user in the list. For the full list and search capabilities, go to "Access control" --> "Users"  in the navigation sidebar on the left-hand side:

     ![](https://files.readme.io/0914f2a038cb9ffe2a8be2eae64389076dec080cef0ea665d3b1923d2fdc13da-image.png)
   * Each user's current access to resources is displayed in the same row as their name.

3. **Change the role**

   * Click on the name of the user whose role you wish to change
   * Select the new role from the available options:

     ![](https://files.readme.io/57117804a45686b2c3050c5fafdfe13ddd5c3df4d0deca716e91315881b8da23-image.png)

The change takes effect immediately.

## After changing a role

After changing a user's role:

* Inform the user about their new permissions
* Ask them to set up new API keys if needed
* Verify their access to required resources