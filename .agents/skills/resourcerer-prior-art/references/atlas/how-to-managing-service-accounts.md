---
updatedAt: 2026-05-12T14:12:41.000Z
---

# How-to: Managing service accounts

Learn how to modify service account settings, manage API keys, and control resource access.

# Before you begin

Ensure that:

* You have the **Owner** role in your organization
* You understand how changes may affect automated processes

All of the operations described below take place on the "Service accounts" page in the Atlas console:

1. **Navigate to Service Accounts**

   * Select "Manage organization" at the top of the Atlas console:  ![](https://files.readme.io/62c8d8460ec41b2dacd7162b09576f5f692aef11b30ee36eb1228a15e676c468-image.png)

     * In the organization profile view, go to "Access control" > "Service Accounts" in the navigation sidebar on the left-hand side:

     ![](https://files.readme.io/1ab1ceaf69065df305fdc886bc2e503441ce6ec9647a7b403a33c6ab87ba85a1-image.png)

# Managing API keys

## Generate a new API key

1. Select the service account from the list
2. Navigate to the "API Keys" tab

   ![](https://files.readme.io/76ba87d13fc29f1187391e3f42fea8c286db5c20b1edc29b5efada686a7ba820-image.png)
3. **Click "Create new API key"**

   * Give it a name:

     <Image align="center" width="50% " src="https://files.readme.io/45ec77f41a80783b3fb114c07405f6072081e039dd2496e084494397ff9c91d1-image.png" />
   * Copy and securely store the key, as it won't be shown again:

     <Image align="center" width="50% " src="https://files.readme.io/96eab2a99c38834ce23b56b18baa6ec15872b693bb955b2926e49c51c934cbe3-image.png" />

     You can generate **additional** keys for the service account later.

## Delete an API key

1. Select the service account from the list
2. Navigate to the "API Keys" tab

   ![](https://files.readme.io/a68175d59a9e974ce8a1059fe4f9ed51ef16a95599c447e1fc5dd285d1963b46-image.png)
3. Find the key you want to delete
4. Click the three dots (`⋮`) menu next to the key and select "Delete access key"

   ![](https://files.readme.io/021cf0fda297410ceb78a09dc44babf79d390464f1441f6f7594b28c08b75fe6-image.png)
5. Confirm the deletion

> 🚧 Warning
>
> Deleting an API key:
>
> * Takes effect immediately after confirmation
> * Cannot be undone
> * Will break any integrations using that key

Alternatively, an API key can be toggled `on` or `off` to manage temporary access and functionality.

# Adjusting resource access

Resource access of a service account can be configured at both organization and cluster levels at any time.

> 🚧 Warning
>
> Modifying the resource access of a service account will automatically change the access of any associated API keys.

## Adjust access of a service account

1. Select the service account from the list
2. Go to the "Access tab"

   ![](https://files.readme.io/8ac8c3713a885deaf3898dbdf723f691acbe972a3b3b9c3c97b256df0c2b187e-image.png)

### Give access to a service account

1. Click "+ Give access" in the Access tab above the resource list

   ![](https://files.readme.io/060ff2fd23c41abfba92887f73e0ced7d17ddfc2dc84ea5e3f77ecada187b918-image.png)
2. **Set access permissions**

   * Choose resource access level:
     * **Organization**: The service account will have access to all resources
     * **Specific clusters**: Select individual clusters the service account can access

   * Select a role (access scope):

     <Image align="center" width="50% " src="https://files.readme.io/d7a170f32e6e3f7b21a88aa9d8f21b3f49a27afc7db5aa2678134491030e065e-image.png" />

   * Click "Give access".

### Modify existing access level

1. Use the role dropdown to modify the existing access level to any resources

   ![](https://files.readme.io/24ce3b205df8c5cc35a9871b6ab54eddb4617160265719b9bf4da8205b8225d1-image.png)

Changes take effect immediately.

> 📘 Note
>
> When clusters inherit organization-level access, their individual role dropdowns will be disabled. To set cluster-specific roles, you must first configure "Direct" access to clusters.

## Revoking access

To revoke a service account's access to specific resources:

1. Select the service account
2. Go to the "Access tab"

   ![](https://files.readme.io/8ac8c3713a885deaf3898dbdf723f691acbe972a3b3b9c3c97b256df0c2b187e-image.png)
3. Click the three dots (`⋮`) menu for the resource (whether it's a cluster or an organization)
4. Select "Revoke access"

   ![](https://files.readme.io/45dc1bb74e149d2b19c50f89f5b1a452c855eea276d032b4158b7cbdd7dcceeb-image.png)
5. Confirm the revocation

# Assigning to user groups

Service accounts can be managed through user groups just like regular users. You can add them to existing groups or remove them as needed.

## Add to user groups

1. Select the service account from the list
2. Navigate to the "User groups" tab

   ![](https://files.readme.io/55c7b2ec7ae5c022f8564f70d2a45d767759898420dee1e99356c3afca0cc7b4-image.png)
3. Click "+ Assign to user group"
4. Select one or more groups from the list

   <Image align="center" width="50% " src="https://files.readme.io/472ffa3819f7dd1cfc4221e67589876d24f8016f03943bbda2be9f6c49755203-image.png" />
5. Click "Assign" to confirm

   <Image align="center" width="50% " src="https://files.readme.io/82ec5ac6934203310f07947f59e62edf6e5a7c67cc5aaa5bbf61b3f0d362d096-image.png" />

> 📘 Note
>
> When adding a service account to user groups:
>
> * It will inherit all permissions from the groups
> * Any API keys associated with the service account will gain access based on group permissions
> * The service account can be a member of multiple groups simultaneously

## Remove from user groups

1. Select the service account from the list
2. Navigate to the "User groups" tab

   ![](https://files.readme.io/aa7d4fb17ec7cb8f982d389614d581f2206fb4bf868cada49e6e435345e5b680-image.png)
3. Find the group you want to remove the service account from
4. Click the three dots (`⋮`) menu for the user group and select "Remove from the group"
5. Confirm the removal

> 🚧 Warning
>
> Removing a service account from a user group:
>
> * Takes effect immediately
> * May affect access to resources if the account had no direct access configured
> * Will impact any API keys using permissions inherited from the group

# Removing a service account

> 🚧 Warning
>
> Removing a service account:
>
> * Immediately revokes all access
> * Invalidates all API keys
> * Cannot be undone
> * Will break any integrations using the account

To remove a service account:

1. Choose the service account from the list
2. Click the three dots (`⋮`) menu next to the chosen key
3. Select "Remove service account"

   ![](https://files.readme.io/6068eb09be889594c57d925e8520e863836c66b6adf4a958ad71c079e4c96426-image.png)
4. Confirm deletion