---
updatedAt: 2026-05-12T14:12:41.000Z
---

# How-to: Managing existing user groups

Learn how to modify user group settings, manage access controls, and delete groups.

# Before you begin

Ensure that:

* You have the **Owner** role in your organization
* You understand how group changes will affect user access
* You've planned any necessary adjustments to API keys or automated processes

# Adding and removing members

All of the operations described below take place on the "User groups" page in the Atlas console:

1. **Navigate to the organization settings**

   * Select "Manage organization" at the top of the Atlas console:  ![](https://files.readme.io/62c8d8460ec41b2dacd7162b09576f5f692aef11b30ee36eb1228a15e676c468-image.png)

2. **Go to User groups**

   * In the organization profile view, go to "Access control" > "User groups" in the navigation sidebar on the left-hand side:

     ![](https://files.readme.io/ba455d90729dbecc6869e77c3b0f0a086097f8a7180ba3f154989da275cd46d9-image.png)

## Add members to a group

1. **Select the group you want to modify**:

   ![](https://files.readme.io/6ffa4233875659f446ebd914365b4c258e7850c29936bcd7ac90d97dd8351c93-image.png)

   * This will open a user group configuration drawer.

2. **Go to the "Users" tab**
   * Click "+ Add users" :

![](https://files.readme.io/bc4622e12d595487e6732e703a6f172b6be48aa278cca8cf6831f82d113473f3-image.png)

3. **Select the users**
   * Select the users and/or service accounts you wish to add from the list from the list:

<Image align="center" width="50% " src="https://files.readme.io/d0e7864bf792cd3bdb0bd6322ddd9325a072805309beafe1ccdec98918b49f2d-image.png" />

4. **Confirm selection**
   * Click "Add user" to confirm the addition of the selected users and/or service accounts.

The user group will populate with the newly added members immediately.

## Remove members from a group

1. **Select the group you want to modify**:

   ![](https://files.readme.io/6ffa4233875659f446ebd914365b4c258e7850c29936bcd7ac90d97dd8351c93-image.png)

   * This will open a user group configuration drawer.

2. **Go to the "Users" tab**

3. **Select the users to remove**\
   You can either remove a single user/service account from the list or select multiple ones and remove them in bulk.

   * To remove a single user or service account from the list, click the (`⋮`) symbol next to its name and select 'Remove from group":

     ![](https://files.readme.io/634198fdd16fc79d4d7af298b0193aec9c7f48253eeba52b638f46fe11ac080b-image.png)

   * To remove users or service accounts in bulk, select the ones you wish to remove using the checkboxes and click "Remove users from group":

     ![](https://files.readme.io/9ff2ab86f6301d3084a860dab3fbccaa699bc031872d84cb099a1e237df17de6-image.png)

4. When prompted to confirm, click "Remove".

# Adjusting resource access

Resource access can be configured at both organization and resource (cluster) levels through separate role assignments.

> 🚧 Warning
>
> Changing a group's role affects all members simultaneously. This might impact:
>
> * Access to resources
> * API key functionality
> * Automated processes

## Change organization-level access

To change a user group's access at the organizational level, do the following:

1. **Select the group you want to modify**:

   ![](https://files.readme.io/6ffa4233875659f446ebd914365b4c258e7850c29936bcd7ac90d97dd8351c93-image.png)

   * This will open a user group configuration drawer.
2. **Select role:**

   * At the top of the resource access view, set an organization-wide role from the dropdown:

     ![](https://files.readme.io/5384b76c539f5d86da8205ba1220b2176892b1dbda219ebcb44981d75c16a8d3-image.png)

This sets the base access level for all resources at the organization level.

### Revoke access

To revoke a user group's organizational-level access:

1. Click the (`⋮`) symbol in the organizational-level permissions card and select "Revoke access":

   ![](https://files.readme.io/43de6b7b1077fea620342fe736142a3f0eee71085d06f835a1a1383176216b1d-image.png)

> 🚧 Warning
>
> Revoking a user group's organization-level access will leave the group without any access to resources.

## Change cluster-level access

Below the organization-level role, you can:

* See all available clusters for that user group
* Configure cluster-specific roles using individual dropdowns if they are not inherited from the organization-level access of the user group
* Set different roles than the organization-level access

To modify a direct resource access role:

1. **Select the group you want to modify**:

   ![](https://files.readme.io/6ffa4233875659f446ebd914365b4c258e7850c29936bcd7ac90d97dd8351c93-image.png)

   * This will open a user group configuration drawer.
2. **Select role:**

   * In the cluster list, set a cluster-wide role from the dropdown for each cluster whose role you wish to modify:

     ![](https://files.readme.io/3f000504570c7cde790763d23262a99f5a75c24b449377baf9241ee78cfdef2e-image.png)

This sets the user group's access level for that particular resource.

> 📘 Note
>
> When clusters inherit organization-level access, their individual role dropdowns will be disabled. To change cluster-specific roles, you must first configure "Direct" access to clusters.

### Revoke access

To revoke a user group's organizational-level access:

1. Click the (`⋮`) symbol in the cluster row card and select "Revoke access":

   ![](https://files.readme.io/38579a0c4cc59eeeb9ec988a04bf057065902209ade09f183d59ddb9adb3bb93-image.png)

> 🚧 Warning
>
> If a user group contains service accounts, revoking a user group's cluster-level access will leave the API keys of those service accounts unable to access the cluster, which can break existing integrations.

## Give access to a resource

Access to more resources can be granted to a user group at any time.

To grant additional access:

1. **Select the group you want to modify**:

   ![](https://files.readme.io/6ffa4233875659f446ebd914365b4c258e7850c29936bcd7ac90d97dd8351c93-image.png)

   * This will open a user group configuration drawer.

2. **Click "+ Give access"** in the resource list area:

   ![](https://files.readme.io/e325b322172fc26f5991ede20eaa4cff60899d7195a067c5299d0988ea719e80-image.png)

3. **Configure new access:**

   * Choose the resource type (Cluster or Organization)
   * For cluster access:
     * Select specific clusters from the dropdown
     * Choose the role (Owner, Member, Analyst, or Viewer)
   * For organization access:

     * Select the role (Owner, Member, Analyst, or Viewer):

     <Image align="center" width="50% " src="https://files.readme.io/29391bfecc4033942e0204511e08cadde3fff8ebf71ae5431247ec2fe77b2fab-image.png" />

4. **Click "Give access"** to confirm the new access settings. The new bindings will be built momentarily.

> 📘 Note
>
> Access can be granted at either organization or cluster level. Organization-level access applies to all current and future clusters in the organization, while cluster-level access is specific to individual clusters.

> 📘 Note
>
> When adding multiple clusters at once, they will all be assigned the same role that you select. For example, if you select three clusters and the **Viewer** role, all three clusters will be added with **Viewer** access.
>
> If you need to give different roles for different clusters (like **Viewer** access to one cluster and **Member** access to another), you'll need to add them separately through multiple "Give access" operations - one for each role level you want to assign.