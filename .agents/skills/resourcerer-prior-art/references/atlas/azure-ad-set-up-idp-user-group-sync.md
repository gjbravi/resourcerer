---
updatedAt: 2026-09-22T08:46:05.000Z
---

# Azure AD: Set up IdP user group sync

This guide walks you through setting up automated group synchronization between Microsoft Entra ID (Azure AD) and Atlas. The process involves configuration steps in both Atlas and the Microsoft Entra console.

# Before you begin

Ensure that:

* You have an existing SSO connection between Azure AD and Atlas configured and working. If you haven't set this up yet, see <Anchor target="_blank" href="https://docs.atlas.example/docs/sso#azure-ad">Setting up SSO with Azure AD</Anchor>
* You have administrative access to your Microsoft Entra ID (Azure AD) tenant, including the ability to create and configure Enterprise Applications
* You have the **Owner** role in your Atlas organization
* You've identified which Azure AD groups should be synchronized to Atlas

<Callout icon="⚠️" theme="warn">
  ### Azure AD licensing requirement

  A paid Azure AD (Entra ID) plan — such as Microsoft Entra ID P1 or P2 — is required to sync only specific groups rather than all users and groups in your directory. Without a paid plan, you can still configure provisioning, but you will not be able to selectively assign groups to the application.
</Callout>

<Callout icon="📘" theme="info">
  ### Note

  This setup requires coordination between Atlas console configuration and Microsoft Entra console changes. Plan to complete both parts in the same session.
</Callout>

# Part 1: Enable group sync in Atlas

## Step 1: Navigate to SSO settings

1. In the Atlas console, go to **Organization settings**
2. Navigate to **Access control** > **Single sign-on** in the left sidebar
3. Locate your existing Azure connection in the table

![](https://files.readme.io/ce9d5657cc2d8463676b78d66e6d00854ad1be5a9f258eed225cb25e765f3737-image.png)

## Step 2: Enable synchronization

1. Click your Azure AD connection, then check the **Synchronize user groups from the IAM system to Atlas** box in the drawer that opens to enable sync

![](https://files.readme.io/94d84ae1fb92c7ce9db2c049b77807c8046a0ea6fcb4bceb73ef7f6783959329-image.png)

2. Atlas will generate an SCIM endpoint URL and bearer token for your organization once you press the **Save changes** button
3. **Copy and save both the URL and token** – you'll need these for Azure AD configuration

![](https://files.readme.io/9e7d70f6c39e63b3617b57fce1785407d594553678a84df15762451459ba9235-image.png)

<Callout icon="❗️" theme="error">
  ### Important

  Keep the SCIM URL and token secure. You'll use these in Azure to establish the connection.
</Callout>

<Callout icon="📘" theme="info">
  ### Note

  This step can also be completed when setting up the SSO connection for the first time at Atlas.
</Callout>

# Part 2: Configure SCIM provisioning in Microsoft Entra ID

## Step 1: Create an Enterprise Application

1. Sign in to the [Microsoft Entra admin center](https://entra.microsoft.com) as an administrator
2. Navigate to **Applications** > **Enterprise applications**
3. Click **New application**
4. Click **Create your own application**
5. Give the application a descriptive name (e.g., "Atlas SCIM Sync" or "Atlas Group Provisioning") to help other administrators identify its purpose
6. Under **What are you looking to do with your application?**, select **Integrate any other application you don't find in the gallery (Non-gallery)**
7. Click **Create**

<Callout icon="📘" theme="info">
  ### Note

  Use a descriptive name rather than the default. This prevents accidental deletion by other administrators who may not recognize a generic application name.
</Callout>

<Columns layout="auto">
  <Column>
    ![](https://files.readme.io/643f1bff03932a24db86c2d5b38fdd3a60a9f4c274640817d0fdf545e300de1b-image.png)
  </Column>
</Columns>

## Step 2: Configure provisioning credentials

1. In your newly created Enterprise Application, click **Provisioning** in the left sidebar

![](https://files.readme.io/8f9783a504ce353d2a7e25a1d7c67d7aee1d0f082ee5e45523113fe9f4d7f85f-image.png)

2. On the new view, click the **New Configuration** button

![](https://files.readme.io/f57463ffcb5a1ef8fe1949052a805207f3bab7af21fbb78a159509fc15f01cfa-image.png)

3. Under **Admin Credentials**, enter the following:
   * **Tenant URL**: Paste the SCIM URL from Atlas
   * **Secret Token**: Paste the bearer token from Atlas
4. Click **Test Connection** to verify that Microsoft Entra ID can reach the Atlas SCIM endpoint

<Callout icon="❗️" theme="error">
  ### Important

  If the test connection fails, verify that:

  * The Tenant URL exactly matches the SCIM endpoint from Atlas
  * The Secret Token exactly matches the bearer token from Atlas
</Callout>

5. After the connection is validated, click **Create**

![](https://files.readme.io/cf61254fcc15c675150e5fb85f644b27efd6fcdac5c3e0eb11a51ff954bb200f-image.png)

## Step 3: Review attribute mappings

Microsoft Entra ID maps user and group attributes to the SCIM standard. Review these mappings to ensure they align with Atlas's expected format.

1. Under **Provisioning**, click **Mappings** and select **Provision Azure Active Directory Groups**

2. Ensure the mapping for Groups is as the picture below.

![](https://files.readme.io/b3173cc19f5c4694440567895c9f9d4dcc29771c6bc3eeb070b58bcbc8b06135-image.png)

3. Click **Save**
4. Go to the **Users** mapping tab
5. Configure the following user attribute mappings. For each mapping, click the attribute row to edit it:<br />

   | Azure AD source attribute | SCIM target attribute | Mapping type |
   | ------------------------- | --------------------- | ------------ |
   | `objectId`                | `externalId`          | Direct       |
   | `mail`                    | `userName`            | Direct       |
   | `active`                  | `active`              | Direct       |

   ![](https://files.readme.io/0d669a1dcbc9e8f537d488c7227c356606c8ab621d8229f3ab7250ffd157d00a-image.png)

   ![](https://files.readme.io/c6924cb4363dcacf01346297d236911b553461723e9c835844d8223ed03a9d36-image.png)

   ![](https://files.readme.io/7fb83e12a2044aa3ffcf6b156f9fb7b53a8d269abd435fad4626ff923cbba4eb-image.png)

<Callout icon="❗️" theme="error">
  ### Critical mapping requirements

  The following mappings are mandatory for Atlas to correctly provision user identities. If any of these are missing or misconfigured, the Atlas SCIM server will reject the user identity:

  * `objectId`**&#x20;→&#x20;**`externalId`: Atlas uses the Azure AD object ID to uniquely identify and track users across sync cycles. This mapping is **not** included in the default Entra ID SCIM template and you must add it manually.
  * `userPrincipalName`**&#x20;→&#x20;**`emails[type eq "work"].value`: The `emails` field **must** contain a value in valid email format. Map it from the `userPrincipalName` attribute, or some other field that is always an email address. In order to decide which field might work for you, you can navigate to your `Users` section and click on one of your user's **Properties** tab and look for a field that's always an email.<br />

    ![](https://files.readme.io/a325bd9ad0a4bc98eaea28023e593745168e282886ab7d2a642a80b6ad599409-image.png)

    If the server receives a `emails[type eq "work"].value` that is not a valid email, it will reject the identity.
  * `active`**&#x20;→&#x20;**`active`: Set the mapping type to **Direct** (not Expression). Ensure the mapping is **Enabled**. This allows Atlas to correctly reflect the user's enabled or disabled state from Azure AD.
</Callout>

6. If any of the mappings listed above do not already exist, click **Add New Mapping**, set the **Mapping type** to **Direct**, select the appropriate source and target attributes, and click **OK**
7. Click **Save**

# Part 3: Assign groups and start provisioning

## Step 1: Assign groups to the application

1. In your Enterprise Application, click **Users and groups** in the left sidebar
2. Click **Add user/group**

   ![](https://files.readme.io/cf6cc067012342f305c7125311c75a9836187a8e1e9076d5aae376960446348f-image.png)
3. Under **Users**, leave the selection as **None** (group membership will drive user provisioning)
4. Under **Groups**, search for and select the Azure AD groups you want to synchronize with Atlas
5. Click **Select**
6. Click **Assign**

<Callout icon="📘" theme="info">
  ### Note

  Only users and groups explicitly assigned to the Enterprise Application will be synchronized to Atlas. This is why a paid Entra ID plan (P1 or P2) is required — it enables group-based assignment rather than syncing the entire directory.
</Callout>

## Step 2: Start provisioning

1. Navigate back to **Provisioning** in your Enterprise Application
2. Under **Settings**, set **Provisioning Status** to **On**
3. Optionally, set a **Notification Email** to receive alerts about provisioning failures

![](https://files.readme.io/c4054c6311efe9084921449ebede9f15c6c8820ed36fafc2c1f77c5b2f3f4ba5-image.png)

4. Click **Save**
5. Click **Start provisioning** to initiate the first synchronization cycle

![](https://files.readme.io/ffaeac479b3f40a5d4a592fa432df3c6549344c20d7c71e0a80dd467056c7176-image.png)

<Callout icon="📘" theme="info">
  ### Note

  The initial sync for Azure AD can take up to **40 minutes** to complete, depending on the number of users and groups being synchronized. Subsequent incremental syncs are typically faster.
</Callout>

## Step 3: Monitor provisioning progress

1. Click **View audit logs** (or navigate to **Provisioning logs**) for detailed information about:
   * Which groups and users were created or updated in Atlas
   * Any errors or skipped records
2. Wait for the status to show **Cycle completed** before verifying in Atlas

   ![](https://files.readme.io/a6773d4d7961f01af61aed4b271c0a2634c95686931b662c7dccc7cf55f2ed8b-image.png)

<Callout icon="⚠️" theme="warn">
  ### Troubleshooting provisioning errors

  If provisioning fails for specific users or groups:

  * Check the **Provisioning logs** for error details — filter by **Status: Failure**
  * Ensure the affected users have all required attributes populated in Entra ID (especially `emails` and `externalId`)
  * Verify that the group is assigned under **Users and groups**
  * For recurring failures, try clicking **Restart provisioning** to force a full sync cycle
</Callout>

# Part 4: Confirm groups in Atlas

1. Return to the Atlas console and navigate to **Access control** > **User groups**
2. You should see the new groups created from Azure AD
3. Click on a user group and go to the **Users** tab to see the users that match your Azure AD configuration:

   ![](https://files.readme.io/c52a875cf0b7c5b0020c92873ac256e9b364c885eacf3cfe9a37807876864c91-image.png)

# Ongoing synchronization

From this point forward, any changes you make to the synchronized groups in Microsoft Entra ID will automatically be reflected in Atlas with the next provisioning cycle. This includes:

* **Adding users**: When you add users to synced groups in Entra ID, they will appear in Atlas
* **Removing users**: When you remove users from synced groups or deactivate them in Entra ID, they will lose their Atlas access
* **Updating user information**: Changes to user attributes like display names will sync to Atlas

Keep in mind that synchronization is not instantaneous and may occur in apparent stages. You may notice that:

* A newly added user appears in the Atlas **Users** list first
* The same user may take some additional time to appear in their assigned **User groups**
* Updates typically complete within several minutes, though timing can vary based on group size, the scope of changes, and the Entra ID provisioning cycle

## Managing group permissions in Atlas

While group membership is managed in Microsoft Entra ID through synchronization, permissions and roles are assigned in Atlas. To configure permissions for a synced group:

1. In the Atlas console, navigate to **Access control** > **User groups**
2. Find and click on the synced group you want to configure
3. Assign roles and resource access following the same process as manual groups

For detailed instructions on configuring group permissions, see:

* <Anchor target="_blank" href="doc:how-to-creating-and-configuring-user-groups">Creating and configuring user groups</Anchor>
* <Anchor target="_blank" href="doc:how-to-managing-existing-user-groups">Managing existing user groups</Anchor>

<Callout icon="📘" theme="info">
  ### Note

  While group membership is controlled by Microsoft Entra ID through synchronization, roles and resource access are managed in Atlas. Changes you make to permissions in Atlas will not be overwritten by sync operations.
</Callout>

# Disabling group synchronization

To stop synchronizing groups from Azure while keeping your SSO connection active:

1. In the Atlas console, navigate to **Access control** > **SSO**
2. Click on your Azure SSO connection to open the configuration drawer
3. Uncheck the **Synchronize user groups from the IAM system to Atlas** box
4. Enter your **Client secret** from Azure (required to save any SSO configuration changes)
5. Click **Save changes**

![](https://files.readme.io/cbb28e3b22e70dc3705c919f9354a7f495cffd53a7ac5d065305d8438b38af02-image.png)

<Callout icon="📘" theme="info">
  ### Note

  Disabling sync converts existing synced users and groups to regular Atlas users and groups.
  To completely stop synchronization, also remove the Atlas app from your identity provider.
</Callout>

# For additional help

* <Anchor target="_blank" href="https://learn.microsoft.com/en-us/entra/identity/app-provisioning/use-scim-to-provision-users-and-groups">Microsoft Entra ID SCIM provisioning documentation</Anchor>
* <Anchor target="_blank" href="https://learn.microsoft.com/en-us/entra/identity/app-provisioning/tutorial-scim-provisioning">Tutorial: Develop and plan provisioning for a SCIM endpoint</Anchor>
* <Anchor target="_blank" href="https://learn.microsoft.com/en-us/entra/identity/monitoring-health/concept-provisioning-logs">Microsoft Entra ID provisioning logs</Anchor>