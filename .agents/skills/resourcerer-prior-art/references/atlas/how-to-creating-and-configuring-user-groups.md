---
updatedAt: 2026-05-12T14:12:41.000Z
---

# How-to: Creating and configuring user groups

Learn how to create and set up user groups to manage access for multiple users in your organization efficiently.

# Before you begin

Ensure that:

* You have the **Owner** role in your organization

# Creating a user group

To create a new user group:

1. **Navigate to the organization settings**

   * Select "Manage organization" at the top of the Atlas console:  ![](https://files.readme.io/62c8d8460ec41b2dacd7162b09576f5f692aef11b30ee36eb1228a15e676c468-image.png)

2. **Go to User groups**

   * In the organization profile view, go to "Access control" > "User groups" in the navigation sidebar on the left-hand side:

     ![](https://files.readme.io/ba455d90729dbecc6869e77c3b0f0a086097f8a7180ba3f154989da275cd46d9-image.png)
   * Click the "Create user group" button in the top right.

3. **Configure basic settings**

   * Enter a name for the group
   * Add a description (optional but recommended)

     <Image align="center" width="50% " src="https://files.readme.io/eb7082d0e6cc9bd9d9546ad1f2208037f385d5be8cfcfb756eadfabeffeeb44d-image.png" />

4. **Add members**

   * Select users to add to the group:

     <Image align="center" width="50% " src="https://files.readme.io/7a0f251b52d0d5f3600d76975631c548e778c0e5003336ac60753fa55b460d16-image.png" />

5. **Set access permissions**

   * Choose resource access level:
     * **Organization**: Group members will have access to all resources
     * **Specific clusters**: Select individual clusters members can access

   * Select a role for the group:

     <Image align="center" width="50% " src="https://files.readme.io/d7a170f32e6e3f7b21a88aa9d8f21b3f49a27afc7db5aa2678134491030e065e-image.png" />

   * Click "Give access" to create the user group.

> 📘 Note
>
> Both adding members and assigning resource access can be skipped during initial group creation. You can leave these settings empty and configure them later when you're ready. This is useful when:
>
> * Setting up groups before all team members are onboarded
> * Waiting for new clusters to be created