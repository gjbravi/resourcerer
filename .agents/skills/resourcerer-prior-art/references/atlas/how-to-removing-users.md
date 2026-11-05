---
updatedAt: 2026-05-12T14:01:17.000Z
---

# How-to: Removing users

When team members no longer need access to Atlas, you can remove them from your organization. This guide explains how to remove users safely.

# Before you begin

Before removing a user, ensure that:

* You have **Owner** permissions in your organization
* You've reviewed any resources or configurations owned by the user
* You've backed up or transferred any necessary API keys

> 🚧 Warning
>
> Removing a user:
>
> * Immediately revokes all their access to Atlas
> * Invalidates all their API keys
> * Cannot be undone - they'll need a new invitation to regain access
> * Does not affect resources or configurations they created

# Removing a user

1. **Navigate to the organization settings**

   * Select "Manage organization" at the top of the Atlas console:

     ![](https://files.readme.io/62c8d8460ec41b2dacd7162b09576f5f692aef11b30ee36eb1228a15e676c468-image.png)

2. **Find the user you want to remove**

   * In the organization profile view, locate the user in the list. For the full list and search capabilities, go to "Access control" --> "Users"  in the navigation sidebar on the left-hand side:

     ![](https://files.readme.io/0914f2a038cb9ffe2a8be2eae64389076dec080cef0ea665d3b1923d2fdc13da-image.png)
   * Each user's current access to resources is displayed in the same row as their name.

3. **Remove the user**
   * Click the three dots (⋮) menu at the end of the user's row
   * Select "Remove user" from the dropdown menu
   * Confirm the removal when prompted

4. **Verify removal**
   * The user will disappear from your "Users" list