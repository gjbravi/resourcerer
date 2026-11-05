---
updatedAt: 2026-08-28T07:05:39.000Z
---

# Single Sign-On (SSO)

Single Sign-On (SSO) allows users easy access to Atlas through their existing SSO provider.

# Getting started

To start using SSO, you can use Self-Serve Single Sign-On and configure the integration.

This guide describes establishing trust between Atlas and your identity provider.

> 📘 Regional endpoints
>
> Atlas operates in multiple regions with different endpoints. Ensure you use the endpoints that match your Atlas environment. Using incorrect regional URLs will prevent successful authentication and cause connection errors.
>
> See [API access](api-access.md) for the full list of regional endpoints.

## OpenID providers

Atlas now supports any SSO provider that implements the OpenID Connect standard. This gives you the flexibility to use your preferred identity provider.

### Prerequisites

* An SSO provider that supports OpenID Connect
* Admin access to configure your SSO provider

### Instructions

1. In your SSO provider, create a new OpenID Connect application

2. Configure the application with the following settings based on your Atlas region:

   Use the endpoints matching your Atlas region (see [API access](api-access.md)):

   * **Sign-in redirect URI**: `https://login.<region>.atlas.example/login/callback` (for US: `https://login.atlas.example/login/callback`)
   * **Sign-out redirect URI**: `$ATLAS_API_URL/v1/auth/logout`

3. Note the following details you'll need to provide to Atlas:
   * **Client ID**
   * **Client secret**
   * **Issuer URL** (the base URL for your SSO provider)

4. In the Atlas console, go to **Organizational profile** -> **SSO**

5. Click **Create SSO Connection** and select **OpenID** as the ID Provider

<Image align="center" border={true} src="https://files.readme.io/b822a84a6b19e1ff9ec657e7020ace1fe8d6d618b1825d9f35940874c77f6c6f-image.png" className="border" />

6. Fill in the required fields:
   * **Name**: Give your SSO connection a name
   * **Email domain**: Your organization's email domain
   * **Type**: Choose the appropriate type based on your setup:
     * **Back channel**: Select this if you're using the back channel type of OIDC connection
     * **Front channel**: Select this if you're using the front channel type of OIDC connection
   * **Issuer URL**: The base URL for your SSO provider
   * **Client ID**: From step 3
   * **Client secret**: From step 3. Required for **Back channel** OIDC connection type only.

7. Click **Connect** to establish the SSO connection

Your OpenID SSO connection should now be configured and ready to use.

> 📘 Note
>
> **The correct Issuer URL for Google OIDC authentication is:**
> `https://accounts.google.com/.well-known/openid-configuration`
>
> **All OIDC Issuer URLs should end with:**
> `/.well-known/openid-configuration`

> 📘 Note
>
> The type you choose (back channel or front channel) should match the configuration in your OpenID provider. If you're unsure which type to use, consult your SSO provider's documentation or your IT administrator.

## Azure AD

This section describes setting up an Azure AD application and defining parameters for calling the [Create organization SSO connection](https://docs.atlas.example/reference/ssoapi_createssoconnection) to establish SSO trust between Atlas and the customer-created Azure AD application.

### Prerequisites

Admin access to your Azure AD subscription.

### Instructions

#### Create an Azure AD application

In your Azure portal, navigate to **Microsoft Entra ID** → **App Registrations** → **New Registration**:

<Image align="center" border={true} src="https://files.readme.io/3db3d55-image.png" className="border" />

#### Register a new AD Application

1. Give your preferred name to an app registration
2. Set **Supported account type** to **Accounts in this organizational directory only (Single tenant)**
3. Set **Redirect URI** type to **Web**, and as a callback value, use:
   * US region: `https://login.atlas.example/login/callback`
   * EU region: `https://login.eu.atlas.example/login/callback`
   * India region: `https://login.india.atlas.example/login/callback`
4. Click **Register**

<Image align="center" border={true} src="https://files.readme.io/6e5593d-image.png" className="border" />

#### Configure the new AD Application

1. Under API permissions, make sure that your application has these permissions:

<Image align="center" border={true} src="https://files.readme.io/75e8edb-image.png" className="border" />

2. Make a note of the created app registration Application (client) ID:

<Image align="center" border={true} src="https://files.readme.io/939aeca-image.png" className="border" />

##### Add a Client secret for the created app registration

1. Under the **Certificates & secrets** section in the previously created app registration window, select **New client secret**
2. Set the description for the secret and set the expiration date for two years.
3. Note and securely store the **Secret Value** as it is shown only during creation.

<Image align="center" border={true} src="https://files.readme.io/9d7db32-image.png" className="border" />

Now, you can set up a connection using Atlas console. Go to **Organizational profile** - > **SSO** in the Atlas console.

> 📘 Office 365 App
>
> If your organization is using the [Office 365 App](https://learn.microsoft.com/en-us/microsoft-365/admin/manage/customize-the-app-launcher?view=o365-worldwide), configure your app **Initiate login URI** within the App Registration Home Page URL based on your Atlas region:
>
> * **US region**: `https://console.atlas.example/api/sso?auth0returnTo=https%3A%2F%2Fconsole.atlas.example%2Fapi%2Fsso&domain=org-email-domain`
> * **EU region**: `https://console.eu.atlas.example/api/sso?auth0returnTo=https%3A%2F%2Fconsole.eu.atlas.example%2Fapi%2Fsso&domain=org-email-domain`
> * **India region**: `https://console.india.atlas.example/api/sso?auth0returnTo=https%3A%2F%2Fconsole.india.atlas.example%2Fapi%2Fsso&domain=org-email-domain`
>
> If a URL contains another URL as a query parameter, the parameter must be encoded using percent encoding.

### Terraform

If you manage your infrastructure with Terraform, you can check [this example](https://github.com/atlas/terraform-provider-atlas/tree/master/examples/sso_connection) of creating an SSO connection with the Atlas Terraform provider.

## Okta Workforce Enterprise Connection

This section describes setting up an Okta OIDC app integration and defining parameters for calling [Create organization SSO connection](https://docs.atlas.example/reference/ssoapi_createssoconnection) to establish SSO trust between Atlas and the customer-created Okta OIDC app integration.

### Prerequisites

Access to the OKTA admin console.

### Instructions

#### Create a new Okta OIDC app integration

1. In your admin console, navigate to **Applications** → **Applications** and click on **Create App Integration**:

<Image align="center" border={true} src="https://files.readme.io/16dc70f-image.png" className="border" />

2. Set **Sign-in method** and **Application type**:

| Option               | Type                      |
| :------------------- | :------------------------ |
| **Sign-in method**   | **OIDC - OpenID Connect** |
| **Application type** | **Web Application**       |

3. Click **Next**

<Image align="center" border={true} src="https://files.readme.io/87e6e98-image.png" className="border" />

#### Configure the OKTA app integration

1. Give your preferred name to the app integration
2. Set **Sign-in redirect URIs** based on your Atlas region (see [API access](api-access.md)):
   * `https://login.<region>.atlas.example/login/callback` (for US: `https://login.atlas.example/login/callback`)
3. Set **Sign-out redirect URIs**:
   * `$ATLAS_API_URL/v1/auth/logout`
4. Set proper **Assignments**
5. Click **Save**

<Image align="center" border={true} src="https://files.readme.io/6843688-image.png" className="border" />

Now, you can set up a connection using the Atlas console.

> 📘 Okta App Tile
>
> If your organization is using the [Okta App Tile](https://help.okta.com/oie/en-us/content/topics/settings/new-end-user-dashboard.htm), configure your app **Initiate login URI** based on your Atlas region:
>
> * **US region**: `https://console.atlas.example/api/sso?auth0returnTo=https://console.atlas.example/api/sso&domain=email-domain`
> * **EU region**: `https://console.eu.atlas.example/api/sso?auth0returnTo=https://console.eu.atlas.example/api/sso&domain=email-domain`
> * **India region**: `https://console.india.atlas.example/api/sso?auth0returnTo=https://console.india.atlas.example/api/sso&domain=email-domain`

#### Atlas Console Okta SSO Configuration:

Navigate to **Org Icon** - > **Manage** - > **Organizational profile** - > **SSO**.

<Image align="right" border={false} src="https://files.readme.io/c8ed49b-Screenshot_2024-02-01_at_11.54.41_AM.png" />

<Image align="center" border={true} src="https://files.readme.io/65d7d52-Screenshot_2024-02-01_at_11.54.57_AM.png" className="border" />

<Image align="center" border={true} src="https://files.readme.io/5cab2b3-Screenshot_2024-02-01_at_11.55.32_AM.png" className="border" />

To create an SSO Connection using Okta in the Atlas console, follow these steps:

1. Gather the following information:
   * Connection Name
   * Email Domain (e.g., `your-organization.com`)
   * (Optional) Additional Email Domains (e.g., `your-organization.co.uk`, your-org`anization-service.eu`)
   * Identity Provider (Okta)

2. Prepare the Okta Configuration details:
   * Okta Domain (e.g., `dev-12345678.okta.com`)
   * Client ID (generated during Okta app creation)
   * Client Secret (generated during Okta app creation)

3. Enter the information in the SSO Connection form:

   <Image align="center" border={true} src="https://files.readme.io/63c1920-Screenshot_2024-07-12_at_11.15.58.png" className="border" />

   > 📘 Note
   >
   > Ensure you use your organization's email domain(-s) and the correct Okta domain.

4. After filling in all required fields, click the **Connect** button.

5. If all information is correct, your Okta SSO Connection should be successfully established.

Remember to double-check all entered information before submitting to avoid connection issues. If you encounter any problems, verify your Okta configuration and try again.

## Multiple Email Domains Support

Atlas supports configuring multiple email domains for your SSO connection. This feature is available through three methods.

### Atlas Console

You can add multiple email domains in the configuration form when setting up your SSO connection in the Atlas console:

1. Enter your primary email domain.
2. Click "Add another email domain" to add additional domains, which are optional.
3. Repeat step 2 for each additional domain.

<Image align="center" border={true} src="https://files.readme.io/83c8ac5-Screenshot_2024-07-12_at_11.15.58.png" className="border" />

### API

To configure multiple email domains using the Atlas API, use the [Create organization SSO connection](https://docs.atlas.example/reference/ssoapi_createssoconnection) endpoint. Include all additional email domains in the `additionalEmailDomains` array of the request body:

```json
{
  "name": "My SSO Connection",
  "emailDomain": "company.com",
  "additionalEmailDomains": [
    "company.co.uk",
    "company.eu"
  ]
}
```

### Terraform

If you manage your infrastructure with Terraform, you can configure multiple email domains using the Atlas Terraform provider. Update your Terraform configuration to include all desired email domains.

## Expired credentials

While Atlas doesn't actively monitor SSO credential expiration dates, the platform uses a mechanism to detect when credentials have expired based on provider responses during login attempts. When this happens:

1. Users will see an error message during login indicating expired credentials:

<Image align="center" border={true} width="50% " src="https://files.readme.io/8fe5bf54a733bb7e521154e9bba441ae6f590a54b99c80df0da06eb311dedaec-image.png" className="border" />

2. Organization administrators will receive a detailed notification about the SSO connection issue through the notifications hub. See [Notifications](notifications.md) for more information.

To resolve this, administrators should:

1. Review the notification details in the Atlas console
2. Update the expired credentials in your SSO provider
3. Verify the SSO connection is working by testing the login

> 📘 Note
>
> We recommend tracking SSO credential expiration dates in your identity provider to prevent disruption to user access.

## SSO in Enterprise organizations

If your organization uses [Enterprise organizations](enterprise-organizations.md), SSO is configured at the Enterprise level by default, and child organizations inherit that connection. When multi-tenant SSO is enabled, a child organization can configure its own SSO; until it does, it keeps inheriting the Enterprise connection.

**Key behavior:**

* By default, only one SSO connection can be configured per Enterprise. If subsidiaries or departments use separate identity providers, multi-tenant SSO can be enabled so child organizations can configure their own. Contact your Atlas account representative to enable this — it is not self-service.
* Users authenticate with the IdP for the organization they sign into. Access is still determined by [RBAC](role-based-access-control-rbac.md) — a separate IdP does not grant access to additional organizations.
* When a user logs in via SSO for the first time, Atlas automatically creates a new user entity in the corresponding organization. By default, this user has **no access** — the Enterprise Owner must assign roles on the **Users** page in Settings. See [Manage Enterprise users](how-to-manage-enterprise-users.md).
* To pre-provision access before a user's first SSO login, add them to a [user group](user-groups.md) that already has role bindings configured. The user inherits the group's roles automatically when they first authenticate.
* To eliminate manual role assignment at scale, enable **SSO group sync** to import user groups and their members directly from your identity provider (Okta, Azure Entra ID, or Google Workspace). Group sync is available on the Enterprise only — child organizations with their own SSO connection cannot use it. Contact your Atlas account representative to enable SSO group sync.
* SSO login creates a **separate user entity** from email/password login. If you previously logged in with email/password and then log in via SSO, you will see two user entries with your name. User entity merging is planned but not yet available. To work around this, log in with your existing email/password account and grant your new SSO user the same role your email/password account has — only assign Enterprise Owner if that is the level of access you actually need.

**Troubleshooting:**

If a user receives an **"Access denied"** error with the message "User is not assigned to the client application" when logging in via SSO for the first time, the issue is on the SSO provider side. Ensure the user is assigned to the Atlas application in your identity provider settings.

## Managing SSO connections

### Restricting user access to SSO only

When SSO is enabled, you can prevent manual user invitations via email to enforce that all users authenticate through your identity provider:

1. In Atlas console, navigate to **Access control** > **SSO**
2. Select **User connection only via SSO**
3. The invitation button will be disabled in **Access control** > **Users**

   <Image border={false} src="https://files.readme.io/e7273bbfd3dfbdfc0a913b50a4e7288ff4ad984a9e93f0bbefeb6da216c835be-image.png" />

> 📘 Note
>
> This setting only works when an active SSO connection exists.

#### Revoking pending invitations

If invitations were sent before restricting to SSO-only access, they remain valid until manually revoked:

1. In Atlas console, navigate to **Access control** > **Users**
2. Find users with **Invite pending** status
3. Click the ellipsis (⋯) and select **Revoke invite**

   <Image border={false} src="https://files.readme.io/964ef0f085f6ec8abc81c56f52bc29d68669e31906bcb4c7d3000e8b2973f422-image.png" />

### Deleting an SSO connection

> ⚠️ Warning
>
> Deleting your SSO connection will permanently remove all users who were provisioned through SSO from Atlas. If you have SSO Group Sync enabled, all synced groups will also be removed.

To remove an SSO connection:

1. In Atlas console, navigate to **Access control** > **SSO**
2. Find your SSO connection and click on it
3. In the configuration drawer, click the ellipsis in the top right and select **Delete configuration**
4. Confirm the deletion

   <Image border={false} src="https://files.readme.io/39138f679f99dd1ba0dd28e7d523320b939c9b98f3a168fb93dc792307bab195-image.png" />

**Impact:**

* All users provisioned through SSO will be deleted from Atlas
* If Group Sync is enabled, synced groups will be removed
* Any SCIM connections will be severed