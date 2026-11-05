---
updatedAt: 2026-05-12T14:12:41.000Z
---

# Okta: Managing IdP user group sync

This guide covers ongoing management of your Okta user group synchronization with Atlas, including user lifecycle operations, group management, and troubleshooting.

# Before you begin

Ensure that:

* You have SSO Group Sync already configured between Okta and Atlas
* You have administrative access to your Okta organization
* You have the **Owner** role in your Atlas organization

# Managing users in synced groups

## Adding users to synced groups

To add a user to a synchronized group:

1. In your Okta Admin Console, navigate to **Directory** > **Groups**
2. Select the group that is synchronized with Atlas

   <Image border={false} src="https://files.readme.io/c34e2898272c79fca1a5f113d90615a2cabac4478458d4e66b03e4a6577fe21c-image.png" />
3. Click the **Assign people** button
4. Search for and select the user you want to add
5. Click **+** to add them

   <Image border={false} src="https://files.readme.io/bc1002b28d176721be64fe67c3433df808efedfb09021b85bbc47a56a4bebbde-image.png" />

The newly assigned user will appear in Atlas after synchronization completes. This typically takes up to a few minutes. You can always see the status of a pending group push in the **Push Groups** tab of your Okta application.

## Updating user information

To update a user's information that syncs to Atlas:

1. In your Okta Admin Console, navigate to **Applications** > **Applications**
2. Select your SCIM application for Atlas
3. Go to the **Assignments** tab
4. Find the user whose information you want to update
5. Select **Edit**

   <Image border={false} src="https://files.readme.io/79b5614a0a42c1dc7a06a63c2fc6b82fb1417abc23f7e151eb8443eaabd18e3a-image.png" />
6. Modify the relevant fields (e.g., display name)
7. Click **Save**

The updated information will sync to Atlas.

> 📘 Note
>
> Not all specific user attributes sync to Atlas, beyond things, such asdisplay name, email, and similar. A person's title or contact information, for example, would not sync to Atlas, because those mappings and fields do not exist in the Atlas user system.

## Removing users from synced groups

To remove a user from a synchronized group:

1. In your Okta Admin Console, navigate to **Directory** > **Groups**
2. Select the synced group
3. Find the user you want to remove
4. Click the **X** button next to the user

<Image border={false} src="https://files.readme.io/dd1755f900a1d7247b5158749e0a5426f7bc129805639ecb3792fc45651893e5-image.png" />

Once removed from the group, the user will lose access associated with that group in Atlas.

### Deactivating users

When you deactivate a user in Okta, they lose all access to Atlas:

1. In your Okta Admin Console, navigate to **Directory** > **People**
2. Find the user you want to deactivate
3. Click on the user's name
4. Click **More Actions** > **Deactivate**
5. Confirm the deactivation

<Image border={false} src="https://files.readme.io/1868e6e6600dbfe7b6fcd67d05a3873c47f6bcaf8e3743811f8b7ce1c8a4c4fd-image.png" />

As a result, on Atlas:

* The user is unable to authenticate via SSO
* Any active sessions will be terminated on the next action attempt
* The user's status in Atlas will be marked as inactive

<Image border={false} src="https://files.readme.io/e038374fa356efb8e261b15caea3e2b5d8f3b78014faad54ce664521cd692f3b-image.png" />

# Managing synced groups

## Adding groups to synchronization

To add additional Okta groups to your existing sync:

1. In your SCIM application, go to the **Assignments** tab
2. Click **Assign** > **Assign to Groups**
3. Select the new groups you want to sync
4. Click **Assign** for each group, then **Done**
5. Navigate to the **Push Groups** tab
6. Click **Push Groups** and select **Find groups by name**
7. Search for and select the newly assigned groups
8. For each group, select **Create Group**
9. Click **Save**

The new groups will appear in Atlas after the sync completes.

For more comprehensive instructions, follow the relevant step from <Anchor label="Okta: Set Up IdP User Group Sync" target="_blank" href="doc:how-to-sso-group-sync-okta#/step-6-assign-groups-to-the-application">Okta: Set Up IdP User Group Sync</Anchor>.

## Removing groups from synchronization

To stop synchronizing a group:

1. In your SCIM application, go to the **Push Groups** tab
2. Find the group you want to stop syncing
3. Click on the **Push status** and select the appropriate way to unlink that group
4. Choose to either **maintain** or **delete** the group on Atlas and confirm the action

<Image border={false} src="https://files.readme.io/929d8da4ffb70be453df9cec6f088b9ef2906cfe4ef0e542185f695014f11072-image.png" />

## Updating group permissions in Atlas

While group membership is managed in Okta, permissions are assigned in Atlas.

For synced user groups, the available permissions changes can be made in several different ways, depending on the scope and intended result.

### Updating the default role of new joiners

Permissions granted to new joiners (newly synced users) by default can be changed under the established SSO connection powering the sync.

1. In the Atlas console, navigate to **SSO**
2. Click on the SSO connection
3. In the SSO configuration drawer, under **Default role assigned to new joiners**, select the appropriate role
4. Save your changes

> 📘 Note
>
> In order to make any changes to an existing SSO configuration, you will need to provide the Client secret from Okta each time before saving.

#### Configuring roles and resource access for synced groups

Beyond the default role for new joiners, you can assign specific roles and resource access to synced groups in Atlas. This allows you to grant different permission levels to different groups based on their responsibilities.

To configure permissions for a synced group:

1. In the Atlas console, navigate to **Access control** > **User groups**
2. Find and click on the synced group you want to configure
3. Assign roles and resource access following the same process as manual groups

For detailed instructions on configuring group permissions, see:

* <Anchor label="Creating and configuring user groups" target="_blank" href="doc:how-to-creating-and-configuring-user-groups">Creating and configuring user groups</Anchor>
* <Anchor label="Managing existing user groups" target="_blank" href="doc:how-to-managing-existing-user-groups">Managing existing user groups</Anchor>

> 📘 Note
>
> While group membership is controlled by Okta through synchronization, roles and resource access are managed in Atlas. Changes you make to permissions in Atlas will not be overwritten by sync operations.

# Disabling group synchronization

To stop synchronizing groups from Okta while keeping your SSO connection active:

1. In the Atlas console, navigate to **Access control** > **SSO**
2. Click on your Okta SSO connection to open the configuration drawer
3. Uncheck the **Synchronize user groups from the IAM system to Atlas** box
4. Enter your **Client secret** from Okta (required to save any SSO configuration changes)
5. Click **Save changes**

<Image border={false} src="https://files.readme.io/7adace9359bc178aa43df8efc5ff275fc9d4c5b552f141de2a82d7879261a31d-image.png" />

> 📘 Note
>
> Disabling sync converts existing synced users and groups to regular Atlas users and groups.
> To completely stop synchronization, also remove the Atlas app from your identity provider.