---
updatedAt: 2026-07-27T07:16:36.000Z
---

# How-to: Creating service accounts

Learn how to create and configure service accounts for programmatic access to your resources on Atlas.

# Before you begin

Ensure that:

* You have the **Owner** role in your organization
* You've determined what level of access the service account needs
* You understand which resources the service account will access

# Creating a service account

1. **Navigate to Service Accounts**

   * Select "Manage organization" at the top of the Atlas console:  ![](https://files.readme.io/62c8d8460ec41b2dacd7162b09576f5f692aef11b30ee36eb1228a15e676c468-image.png)

     * In the organization profile view, go to "Access control" > "Service Accounts" in the navigation sidebar on the left-hand side:

     <Image border={false} src="https://files.readme.io/1ab1ceaf69065df305fdc886bc2e503441ce6ec9647a7b403a33c6ab87ba85a1-image.png" />

2. **Start service account creation**

   * Click "Create service account" button in the top right
   * Configure basic settings

     * Enter a name for the service account
     * Add a description (optional but recommended)

     <Image align="center" border={false} width="50% " src="https://files.readme.io/47b8dc281e03380b29455cfbfbde4923578e2ca72347bb9785e3b3f0696dc007-image.png" />

3. **Choose access method**\
   You can configure access in one of two ways:

   ### Through user groups

   Service accounts can inherit access configurations from existing user groups, making it easier to maintain consistent permissions and reducing management overhead.

   * Select "Through user group"
   * Choose one or more existing user groups

     <Image align="center" border={false} width="50% " src="https://files.readme.io/6da8c5deac0f9ce6fa272fa347fe338dcc80f11dd5fcf032f18e0a3685c08422-image.png" />

     The service account will inherit all permissions from selected groups, and any changes to group permissions will automatically apply to it.

   ### Direct access

   Direct access gives you more granular control over the service account's permissions, allowing you to set specific access levels for individual resources. This is best when you need a service account with unique access to a resource or when you want to manage its permissions independently of your user groups.

   * Select "Direct access"
   * Select access scope:

     * Organization-wide access
     * Specific cluster access (select one or more clusters)
   * Select role:

     <Image align="center" border={false} width="50% " src="https://files.readme.io/0d6ecf6a4b38498a0c837f8dff4de7c322abc2be5075bab7c1f9d857e9750e3f-image.png" />

4. **Generate API key**

   * Choose to generate an API key and give it a name:

     <Image align="center" border={false} width="50% " src="https://files.readme.io/c09fded16758ef8fae7783b3df449782f8712a51868351b178e6f2a28c02e5c8-image.png" />
   * Copy and securely store the key, as it won't be shown again:

     <Image align="center" border={false} width="50% " src="https://files.readme.io/96eab2a99c38834ce23b56b18baa6ec15872b693bb955b2926e49c51c934cbe3-image.png" />

     You can generate **additional** keys for a service account later.

> 🚧 Warning
>
> Generated API keys give **full access** at the service account's role and permission level.