---
updatedAt: 2026-05-12T14:12:41.000Z
---

# Okta: Set up IdP user group sync

This guide walks you through setting up automated group synchronization between Okta and Atlas. The process involves configuration steps in both Atlas and Okta.

# Before you begin

Ensure that:

* You have an existing SSO connection between Okta and Atlas configured and working. If you haven't set this up yet, see <Anchor label="Setting up SSO with Okta" target="_blank" href="https://docs.atlas.example/docs/sso#okta-workforce-enterprise-connection">Setting up SSO with Okta</Anchor>
* You have administrative access to your Okta organization
* You have the **Owner** role in your Atlas organization
* You've identified which Okta groups should be synchronized to Atlas

> 📘 Note
>
> This setup requires coordination between Atlas console configuration and Okta admin interface changes. Plan to complete both parts in the same session.

# Part 1: Enable group sync in Atlas

## Step 1: Navigate to SSO settings

1. In the Atlas console, go to organization settings
2. Navigate to **Access control** > **SSO** in the sidebar
3. Locate your existing Okta connection

<Image border={false} src="https://files.readme.io/fc70a6a7f633c8df31e4a41a01f45dd8edb11c241e03e6841bca069c4126a74c-image.png" />

## Step 2: Enable synchronization

1. Click on your Okta connection and check the **Synchronize user groups from the IAM system to Atlas** box in the drawer that opens to enable sync

   <Image border={false} src="https://files.readme.io/de404d33108863e40a1951695d40966ecc6d7b8cfaa75d5d84b24e5c612aa9db-image.png" />
2. Atlas will generate a SCIM endpoint URL and bearer token for your organization once you press the **Save changes** button
3. **Copy and save both the URL and token** – you'll need these for Okta configuration

   <Image border={false} src="https://files.readme.io/5d9c309980b3ced7afd1e5de1e50ef6cb600a42692ce5669b665ec910f1fc412-image.png" />

> ❗️ Important
>
> Keep the SCIM URL and token secure. You'll use these in Okta to establish the connection.

> 📘 Note
>
> This step can also be completed when setting up the SSO connection for the first time at Atlas.

# Part 2: Configure SCIM integration in Okta

## Step 1: Add SCIM 2.0 integration

<Columns layout="auto">
  <Column>
    1. In your Okta Admin Console, go to **Applications** > **Browse App Catalog**
    2. Search for "SCIM 2.0"
    3. Select **SCIM 2.0 Test App (OAuth Bearer Token)** and click **Add Integration**
  </Column>

  <Column>
    <Image align="center" src="https://files.readme.io/0fda4fbfbca4240d91d79a22ebfa7037b36eebad2b66f6ecbcdb2f902fedd19b-image.png" width="100% " />
  </Column>
</Columns>

## Step 2: Configure SCIM provisioning

1. In the newly added SCIM application, give it a descriptive name (e.g., "Atlas SCIM Sync" or "Atlas Group Provisioning") to help other administrators identify its purpose

   > 📘 Note
   >
   > The default name "SCIM 2.0 Test App (OAuth Bearer Token)" may not clearly indicate this application's purpose. Using a descriptive name prevents accidental deletion by other administrators.
2. Make any other preferred changes in **General Settings** and click **Next**

   <Image border={false} src="https://files.readme.io/2676ae83c98d08e39c88c87e4cb8eefc78b5e9e8cceb8b5a7e665a6f5808a913-image.png" />
3. In the **Single-Sign On** tab, click **Done**
4. You will be taken to your app's **Assignments** tab; switch to the **Provisioning** tab instead
5. Click **Configure API Integration**

   <Image border={false} src="https://files.readme.io/8bea0a1ed32a459fab5ccd10f47f4f62b3812aff49bc5c09abe68b8169f4936d-image.png" />
6. Check **Enable API integration**
7. Enter the following:
   * **Base URL**: \[Paste the SCIM URL from Atlas]
   * **API Token**: \[Paste the bearer token from Atlas]
8. Click **Test API Credentials** to verify the connection
9. Click **Save**

<Image border={false} src="https://files.readme.io/0069272f77d17adadb1b477fd062f4e3a268f756870e5714b6a8a2d2dc9ee3b0-image.png" />

## Step 3: Enable provisioning features

1. Still in the **Provisioning** tab, under the **Provisioning to App** settings, click **Edit**
2. Enable the following features:
   * **Create Users**
   * **Update User Attributes**
   * **Deactivate Users**
3. Click **Save**

<Image border={false} src="https://files.readme.io/8b710ffc3364d3d48bea7918b5b6949728efee87bba8b1bafaaf19fb3ef65405-image.png" />

# Part 3: Configure group and user assignments

## Step 1: Assign groups to the application

1. Go to the **Assignments** tab in your SCIM application
2. Click **Assign** > **Assign to Groups**

   <Image border={false} src="https://files.readme.io/e4c95a86153cc9d0521f94cfece7e21740caa1f8f5166d2c13d20a677acaf3da-image.png" />
3. Select the Okta groups you want to synchronize with Atlas
4. Click **Assign** for each group, then **Done**

   <Image border={false} src="https://files.readme.io/6c0e523790956d1a931533de6dc8e55d4260b20e9b6f7052d6b1739c530fe554-image.png" />

After this step, if you go back to the Atlas console, under **Users**, you will notice the users that match the ones under **People** in the assignments on Okta.

**Okta:**

<Image border={false} src="https://files.readme.io/957d4e20b51822d674a830d3b34e3ac185cea8a873a335779e10a30a269b9944-image.png" />

**Atlas:**

<Image border={false} src="https://files.readme.io/5fd9b4aec18742eabcb2bb176ab6c1f17af55eb5cfa6ec4cea1fb3fedab5f736-image.png" />

## Step 2: Configure group pushing

Follow the steps outlined below to configure the mapping of user groups between Okta and Atlas.

1. In the SCIM application, go to the **Push Groups** tab
2. Click **Push Groups** and select **Find groups by name** (or by rule, depending on how you have that configured on Okta)
3. Search for and select the groups you assigned in **Step 6**
4. For each group, select **Create Group** (this will create the group in Atlas)
5. Click **Save**

<Image border={false} src="https://files.readme.io/97333b6b9664c0cb015a19095e89119777cd3c14817331b8f329be786e1c11fa-image.png" />

> 📘 Note
>
> You may use the option to **Link Group** (rather than **Create Group**) on Okta if you have gone ahead and created user groups on Atlas side that match your Okta groups already (by name). Although for a less error-prone experience, we recommend that you allow Okta to create groups on Atlas from scratch.

## Step 3: Initiate group sync

1. After configuring push groups, if you hadn't checked the box to **Push group memberships immediately** in the step before, initiate the push action now to start the initial synchronization and mapping
2. Wait for the sync to complete – this may take several minutes depending on group size

<Image border={false} src="https://files.readme.io/c992048f49c570c032db1ea936954568211c0772caaf3ea89589619a152fd71f-image.png" />

**Monitoring sync progress:**

* Check the **Push Status** column to track synchronization progress
* If any errors occur during sync, hover over the blue tooltip icon next to the status to view error details
* After sync completes, any errors will also be visible in the **Errors** tab under **Pushed Groups**

# Part 4: Confirm groups in Atlas

1. Return to the Atlas console and navigate to **Access control** > **User groups**
2. You should see the new groups created from Okta
3. Click on the user group and go to the **Users** tab to see the users that match your Okta configuration:

   <Image border={false} src="https://files.readme.io/c52a875cf0b7c5b0020c92873ac256e9b364c885eacf3cfe9a37807876864c91-image.png" />

# Ongoing synchronization

From this point forward, any changes you make to the synchronized groups in Okta will automatically be reflected in Atlas. This includes:

* **Adding users**: When you add users to synced groups in Okta, they will appear in Atlas
* **Removing users**: When you remove users from synced groups or deactivate them in Okta, they will lose their Atlas access
* **Updating user information**: Changes to user attributes like display names will sync to Atlas

Keep in mind that synchronization is not instantaneous and may occur in apparent stages. You may notice that:

* A newly added user appears in the Atlas **Users** list first
* The same user may take some additional time to appear in their assigned **User groups**
* Updates typically complete within several minutes, though timing can vary based on group size and the sope of changes applied and pushed from Okta

# For additional help

* <Anchor label="Okta SCIM provisioning documentation" target="_blank" href="https://developer.okta.com/docs/guides/scim-provisioning-integration-connect/main/">Okta SCIM provisioning documentation</Anchor>
* <Anchor label="Okta Push Groups guide" target="_blank" href="https://help.okta.com/en-us/content/topics/users-groups-profiles/usgp-enable-group-push.htm">Okta Push Groups guide</Anchor>