---
updatedAt: 2026-05-12T14:12:41.000Z
---

# How-to: Inviting users to organization

Adding new members to your Atlas organization enables them to access and manage resources based on their assigned roles. This guide explains how to invite new users and manage the invitation process.

# Before you begin

Ensure that:

* You have **Owner** permissions in your organization
* You have the email addresses of the users you want to invite
* You understand the [available roles and their permissions](role-based-access-control-rbac.md)

# Sending invitations

To change a user's role:

1. **Navigate to the organization settings**

   * Select "Manage organization" at the top of the Atlas console:

     ![](https://files.readme.io/62c8d8460ec41b2dacd7162b09576f5f692aef11b30ee36eb1228a15e676c468-image.png)

2. **Navigate to the "Users" page**

   * In the organization profile view, go to "Access control" --> "Users"  in the navigation sidebar on the left-hand side:

     ![](https://files.readme.io/0914f2a038cb9ffe2a8be2eae64389076dec080cef0ea665d3b1923d2fdc13da-image.png)
   * Click the "Invite users" button in the top right corner.

3. **Configure the invitation**

   * Enter one or more email addresses (separate multiple addresses with commas)
   * Click "Send invite"

   Invitees will receive an email with instructions on how to join. The **Viewer** role will automatically apply to all invited users.

   <Image align="center" width="50% " src="https://files.readme.io/442cdbcf651791aaba87db01bdef00ca6c374777826009539385fb901aa413c2-image.png" />

# Managing pending invitations

After sending invitations:

* Pending invitations appear in the "Users" list with a "Pending invite" status
* Invitations expire after 7 days
* You can resend or revoke pending invitations

> 📘 Note
>
> New users receive the Viewer role by default when joining. If you need different default permissions, configure them in your [SSO settings](sso.md).