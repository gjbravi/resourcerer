---
updatedAt: 2026-08-28T07:05:39.000Z
---

# Permissions, users, Orgs, and SSO

### Can a different user be specified than expected during Atlas onboarding?

It was a conscious design decision not to reuse the same role/serviceAccount/servicePrinciple as best practice to avoid operational and security risks. The role gets narrow permissions for this specific VPC and EKS with a specific tag-based IAM policy, which is the bare minimum needed for Atlas to do autoscaling.

Having one role spanning multiple clusters would mean giving Atlas permissions not limited to scope. It also brings operational risk: the same role could be accidentally deleted when removing a single cluster, effectively killing Atlas in other clusters.

***

### Is it possible to transfer clusters between organizations?

Currently, there is no built-in method to transfer clusters between organizations. To achieve this, you would need to disconnect the cluster from organization A and connect it to organization B.

***

### Why does Atlas require adding, deleting, or assigning a keypair? Does it require any ability to SSH?

Atlas doesn't depend on or need SSH access.

However, as Atlas takes over creating VMs, some customers still want to retain SSH access to their nodes. In this scenario, customers can specify in "NodeConfiguration" which keypair to inject into the new VM so that SSH access will work.

Atlas doesn't check KeyPair permission during credential validation.

<Image align="center" src="https://files.readme.io/edd83e3-Screenshot_2023-12-19_at_2.06.05_PM.png" />

***

### Can Atlas SSO use claims in the JWT to get user data without checking the directory?

The Atlas system follows the current API permissions in our documentation. Atlas doesn't process user data; it's verified by our identity provider, Auth0, to confirm the user's identity. We don't store these details, and Auth0 manages them securely during integration setup.

***

### Can I transfer clusters between child organizations in an Enterprise?

No. Clusters cannot be transferred directly between child organizations. To move a cluster, disconnect it from one child organization and connect it to another. See [Enterprise organizations](enterprise-organizations.md) for more details.

***

### Can existing organizations become child organizations in an Enterprise?

Yes. Atlas can attach existing organizations as children of an Enterprise. Users from those organizations are moved to the Enterprise. Clusters stay in the child organization and are not transferred automatically. Contact your Atlas account representative to attach existing organizations. See [Set up an Enterprise](enterprise-organizations.md#set-up-an-enterprise).

***

### What happens to SSO when existing organizations join an Enterprise?

By default an Enterprise supports one SSO connection, which child organizations inherit.

* If only one of the existing organizations has SSO, that connection can be moved to the Enterprise.
* If several have SSO, choose which connection to move to the Enterprise. The others inherit it unless multi-tenant SSO is enabled so a child can keep or configure its own.
* If none have SSO, confirm whether all users share one identity provider. Different IdPs require multi-tenant SSO.

Contact your Atlas account representative to move SSO connections or enable multi-tenant SSO. See [SSO in Enterprise organizations](sso.md#sso-in-enterprise-organizations).

***

### How do I use a parent org API key to access child organization resources?

When using a parent organization API key to call an API endpoint for a resource in a child organization, include the `X-AtlasAi-Organization-Id` header with the child organization's ID:

```shell
curl --request GET \
  --url https://api.atlas.example/v1/<endpoint> \
  --header 'X-API-Key: <parent-org-api-key>' \
  --header 'X-AtlasAi-Organization-Id: <child-org-id>' \
  --header 'accept: application/json'
```

Without this header, the API resolves the request against the parent organization and returns `404`. API keys scoped to a specific child organization work without the header. See [API keys in Enterprise organizations](service-accounts.md#api-keys-in-enterprise-organizations).

***

### Which role are SSO users assigned when they initially sign in?

In a standard organization, Atlas SSO assigns the Viewer role by default. In an [Enterprise](enterprise-organizations.md), the first SSO login creates a user with **no access** until an Enterprise Owner assigns a role. See [SSO in Enterprise organizations](sso.md#sso-in-enterprise-organizations).