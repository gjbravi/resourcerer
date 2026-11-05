---
updatedAt: 2026-05-22T13:13:52.000Z
---

# IdP user group sync

IdP Group Sync automates the synchronization of users and groups between your identity provider (IdP) and Atlas. This feature eliminates the need to manually manage user access across multiple systems by keeping your Atlas user groups automatically synchronized with your IdP groups.

## The challenge of manual user management

As organizations scale their cloud-native infrastructure, managing access control becomes time-consuming and error-prone. Manual user and group management creates challenges that impact both security and productivity:

**Security gaps** – Delays in revoking access for departing employees or updating permissions when team members change roles can create security vulnerabilities.

**Wasted engineering time** – Platform engineers and administrators spend valuable time managing permissions instead of focusing on infrastructure optimization and innovation.

**Inconsistency** – Permissions drift out of sync with your organization's identity provider, making audits difficult and compliance harder to maintain.

## Your IdP as the single source of truth

IdP User Group Sync solves these challenges by integrating directly with your identity provider. Your IdP becomes the single, authoritative source for managing who has access to your organization, clusters, and optimization settings within Atlas.

<Image align="center" caption="This diagram illustrates how Atlas mirrors your IdP's group structure, maintaining consistent access control across your organization." src="https://files.readme.io/a2b6589561456277b30f592b84ca289fd739d879a125238b9042e57f3a40c8aa-image.png" />

> 📘 Learn More
>
> For a business-focused overview of how IdP User Group Sync transforms access management, including real-world scenarios and organizational benefits, see our <Anchor label="blog post on user management automation" target="_blank" href="https://atlas.example/blog/user-management-automation-idp-sync/">blog post on user management automation</Anchor>.

# Overview

## What User Group Sync does

User Group Sync connects your existing identity provider with Atlas to:

* Automatically create and update user groups based on your IdP groups
* Provision new users when they're added to synced groups in your IdP
* Remove user access when they're removed from groups or deactivated in your IdP
* Keep group memberships synchronized between your IdP and Atlas

## Why use User Group Sync

User Group Sync delivers four key benefits that transform how you manage access control:

### Automation: eliminate manual steps

Say goodbye to recreating groups and adding users one by one in Atlas. The sync process runs continuously in the background, freeing your security and engineering teams to focus on higher-value tasks.

### Security: enforce corporate access policies

Access to sensitive cloud optimization and management controls always aligns with your company's established security rules. Automated sync prevents human error and significantly reduces the window for potential security threats caused by delayed permission updates.

### Consistency: never drift out of sync

Your IdP is your source of truth. By mirroring groups directly, you ensure that permission structures within Atlas remain consistent with changes across your entire organization, making audits and compliance significantly easier.

### Scalability: ready for enterprise adoption

Whether you're onboarding one new team or managing hundreds of engineers across multiple departments, IdP User Group Sync makes scaling access permissions simple and immediate.

## Supported identity providers

Atlas supports user group synchronization with any IdP that has SCIM 2.0 **client** capability. Contact your Technical Account Manager for compatibility and next steps.

For the IdPs below, Atlas offers documented step-by-step guidance:

* **Okta** – Syncs groups and users through SCIM protocol
* **Azure AD (Entra ID)** – Syncs groups and users through SCIM protocol

# Prerequisites

Before setting up User Group Sync:

* You must have an existing SSO connection configured between your IdP and Atlas. For instructions on how to do so, see [Single Sign-On (SSO)](sso.md)
* You need administrative access to your identity provider to configure group synchronization.
* Your Atlas account must have the **Owner** role to enable and configure group sync.

# How it works

The IdP User Groups Sync feature establishes a secure connection between your IdP and the Atlas platform using the SCIM 2.0 protocol. This connection enforces near real-time consistency for all access controls.

<Image align="center" caption="This diagram shows how identities and groups are pushed from your IdP through the SCIM protocol to your Atlas organization." src="https://files.readme.io/55e029969cf9c0985cac251de39b5bb74ff1a6fc004fa3b5ff2b9e46e6268012-image.png" />

## Setup process

1. **Enable sync**: You enable group sync in Atlas and configure your IdP to send group information
2. **Group discovery**: Atlas receives your IdP groups and displays them for configuration
3. **Group mapping**: You choose how to handle each IdP group and assign roles
4. **Ongoing synchronization**: Changes in your IdP groups automatically update Atlas

## Group handling options

When configuring sync, you choose how to handle each IdP group:

**Create new groups**: Atlas creates new groups based on your IdP groups. These groups are automatically managed and updated in response to changes in the IdP. This ensures your Atlas group structure matches your IdP exactly.

**Map to existing groups** (Okta only): Connect an IdP group to an existing Atlas group. The IdP group membership will override the existing group membership. Use this when you want to maintain your current Atlas group structure while populating it with IdP users.

> 📘 Note
>
> Azure AD only supports creating new groups. Group mapping to existing Atlas groups is not available for Azure AD.

## User management

**Selective syncing**: You can choose which user groups from your IdP are synchronized to Atlas, ensuring that only relevant teams are granted access.

**New users**: When users are added to synced groups in your IdP, they are automatically provisioned in Atlas with the appropriate group memberships and roles. Near real-time provisioning means minimal delay—typically only due to your IdP's defined sync cycle.

**Existing users**: Users who already have Atlas access will have their group memberships updated to match their IdP group assignments.

**User removal**: When users are removed from synced groups or deactivated in your IdP, they lose the associated access in Atlas. Deactivating a user also automatically revokes any API keys they created, eliminating potential backdoors and significantly tightening security around infrastructure access.

## Sync timing

* **Initial sync**: Timing depends on your identity provider and the volume of users and groups being synchronized. Okta typically completes initial sync quickly, while Azure AD can take up to 40 minutes
* **Ongoing updates**: Atlas processes changes immediately as it receives them from your IdP
* **Provider-initiated sync**: You can trigger synchronization from your IdP's SCIM provisioning application when needed

# Key concepts

## IdP as single source of truth

When group sync is enabled, your identity provider becomes the authoritative source for group membership, user status, and synced group structure. Atlas automatically reflects changes made in your IdP. Direct changes to synced group membership in Atlas are restricted to maintain consistency with your IdP.

## Permission inheritance

When users join Atlas through synced groups, they receive the default role defined in your SSO connection configuration. Beyond this default, users can be assigned additional permissions through:

* Membership in synced groups with configured roles and resource access
* Individual role assignments made in Atlas

When assignments provide access to the same resource, the highest permission level takes precedence. Organization-level permissions override cluster-level permissions.

## Monitoring and visibility

You can monitor synchronization through your identity provider's SCIM application, which provides:

* **Sync health**: Whether synchronization is working correctly
* **Last sync time**: When the most recent sync occurred
* **Sync logs**: Detailed information about sync operations and any issues

In Atlas, you can view which groups are synced by navigating to **Access control** > **User groups**, where synced groups are displayed alongside manually created groups.

> 📘 Note
>
> Additional sync status information, such as active sync status and member counts, will be available in the Atlas console in a future update.

# Limitations

## Sync timing

User Group Sync operates in near real-time, but not instantly:

* **Initial sync**: Can take several minutes to complete for organizations with many users and groups
* **Ongoing updates**: Changes typically reflect within minutes, but may take longer
* **Large organizations**: Sync times may be extended due to provider rate limits and processing queues

## Provider-specific limitations

**Okta**:

* Requires explicit "Push Groups" configuration - groups must be manually selected for synchronization
* Rate limits apply (typically 600 requests per minute for SCIM operations)

**Azure AD**:

* Users and groups must be explicitly assigned to the Atlas application before they can be synchronized
* Attribute mapping may require customization to work with Atlas's expected format
* Sync behavior can vary based on tenant size and license type

# Getting started

## Prerequisites checklist

Before proceeding with setup:

* [ ] SSO connection between your IdP and Atlas is working
* [ ] You have Owner permissions in your Atlas organization
* [ ] You have administrative access to your identity provider
* [ ] You've identified which IdP groups should be synced to Atlas

## Next steps

Choose your identity provider to begin setup:

* [Set up User Group Sync with Okta](how-to-sso-group-sync-okta.md)

Once configured, learn how to manage your synchronized groups:

* [Managing SSO Group Sync](how-to-managing-sso-group-sync.md)

## Frequently asked questions (FAQ)

### User deactivation and reactivation

**Q: If I deactivate a user in the IdP, will it automatically deactivate them in Atlas?**

Yes. User status is synced automatically—once deactivated in your identity provider, the user will also be deactivated in the Atlas console.

**Q: What happens to API keys created by a deactivated user?**

All API keys created by that user will expire and can no longer be used across the platform.

**Q: If we deactivate a user and later reactivate them in our IdP, what happens in Atlas?**

The same user identity will be reactivated in Atlas, with all previously assigned permissions restored.

### Login methods and SSO

**Q: Will we still be able to use other login types if IdP User Sync is enabled?**

User identities that were created before IdP User Sync enablement (such as via email/password or Google login) will remain in the Atlas console and continue to work as before. However, if a new user attempts to log in with the same email address using a non-SSO login method, they will not be connected to your organization. Instead, a new blank organization will be created for them.

### Azure AD configuration

**Q: What permissions are needed to enable IdP User Group Sync for Azure if we already have an Enterprise Application?**

You'll need an administrator who can create an Enterprise Application on the Azure side to enable this feature. Additionally, a paid Azure AD (Entra) plan is required if you want to sync only specific groups instead of all users and groups within your directory.

### Troubleshooting sync issues

**Q: I'm part of a synced group with a role assigned within the Atlas console, but when signing in, I see the error "User is not assigned to the client application." What could be missing?**

You also need to be assigned to the initial Okta web application—the one that handles the SSO login, not the SCIM integration app. To verify this, open the SSO web app in Okta and check the **Assignments** tab to ensure your user is listed there:

![](https://files.readme.io/859935cab92cb58d5b2347b0bf28c35d6a835e9ad8c14a609215a763712e0eb0-image.png)

**Q: We set up IdP User Group Sync, but it still doesn't automatically push everyone through IdP. What could be wrong?**

The most common issue is using incorrect Atlas endpoints during configuration. Make sure you're using the correct subdomain for your region:

* US organizations: `https://console.atlas.example/`
* EU organizations: `https://console.eu.atlas.example/`
* India organizations: `https://console.india.atlas.example/`

# External resources

* <Anchor label="Blog: User Management Automation With Atlas" target="_blank" href="https://atlas.example/blog/user-management-automation-idp-sync/">Blog: User Management Automation With Atlas</Anchor>