---
updatedAt: 2026-08-11T10:26:34.000Z
---

# API access

## Overview

We build our services API-first; anything you can do in our console UI is available via [REST API](https://docs.atlas.example/reference/). You can use either your preferred way to call REST services directly or leverage our [Terraform-provider](terraform-provider.md) to automate your infrastructure provisioning.

## Authentication

You will need an API key before you can use our API, either with your preferred REST client or via Terraform.

## Obtaining an API access key

From the top-right corner of the Atlas console, click your user menu and select [**API keys**](https://console.atlas.example/user/api-access).

<Image align="center" border={false} src="https://files.readme.io/b085c5374e4a4755db91f603b4fbf51dc3c4548a947c6a1697cb736efab5a202-api-keys-navigation.png" />

This takes you to the **API access keys** page. Click **Create access key**.

<Image align="center" border={false} src="https://files.readme.io/3f0961c3d38947c02d2b50747ee392f0e95d728ecc7484ec10b6f417b36e24cf-api-access-keys-page.png" />

Name your key and select the organization to scope it to. The key will inherit the permissions from your role bindings within the selected organization.

<Image align="center" border={false} width="500px" src="https://files.readme.io/a95ab1acbf665114b648a0a620e79e7457da3a87193f1389aa485caf89540903-create-access-keys.png" />

We advise using a descriptive name for your intended purpose. If you add more keys later, it will be easier to distinguish which key is used for which integration.

## Saving a key

When the key is created — save it, because **you will not be able to view the key again after this window is closed**.

<Image align="center" border={false} width="500px" src="https://files.readme.io/fb08385642552a3cd13ba36c67d58288f6d08da216b3cb0b875b23d571d7ed9d-access-key-created-dialog.png" />

The reason the API key value is visible only at the time of creation is that we do not store the key in plain text on our system. For security reasons, Atlas "forgets" the key value after giving it to you, and later is only able to verify if the key is valid, but not to re-retrieve the value for you.

If you lose your key, the only solution is to create a new key.

## Using a key in API calls

To authenticate, provide the key in the `X-API-Key` HTTP header. For example, with `curl`:

```shell
curl -X GET "https://api.atlas.example/v1/kubernetes/external-clusters" -H "X-API-Key: your-api-key-here" | jq
```

> 📘 Regional API endpoints
>
> Use the API endpoint that matches your Atlas environment:
>
> * **US region**: `https://api.atlas.example`
> * **EU region**: `https://api.eu.atlas.example`
> * **India region**: `https://api.india.atlas.example`

## Enterprise API keys

Enterprise Owners and Viewers can use an enterprise API key to access resources in child organizations. In addition to the standard `X-API-Key` header, you need to include the **`X-Atlas-Organization-Id`** header with the target child organization ID for the request to target the child organization.

```
curl -X GET "https://api.atlas.example/v1/kubernetes/external-clusters" \
  -H "X-API-Key: your-enterprise-api-key" \
  -H "X-Atlas-Organization-Id: child-organization-id"
```

## Atlas Swagger setup

You can test your key directly in our [API specification](https://api.atlas.example/spec/).

Visit <https://api.atlas.example/spec/>, click **Authorize** and enter your key in the `X-API-Key` field.

After setting this up, you are ready to use the **Try it out** button available for each endpoint.