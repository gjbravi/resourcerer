---
updatedAt: 2026-07-27T07:16:36.000Z
---

# Getting started

Enable Application Performance Automation and configure the required integrations.

<Callout icon="🔐">
  ### Private Preview Feature

  This feature is in private preview and available to select customers. To request access, provide feedback or book a demo, contact your Atlas Account Manager.
</Callout>

This guide walks you through enabling Application Performance Automation (APA) and configuring the integrations required to run your first runbook.

By the end of this guide, you will have:

* Enabled APA on at least one cluster
* Registered an AI provider
* Connected a Git repository

# Before you start

To complete this guide, you need:

* A cluster connected to Atlas
* An API key from a supported AI provider (OpenAI, Anthropic, or Google Gemini)
* A personal access token for GitHub or GitLab with the required permissions:

| Provider | Required permissions                  |
| :------- | :------------------------------------ |
| GitHub   | `read:discussion`, `read:org`, `repo` |
| GitLab   | `api`, `write_repository`             |

# Enable APA on a cluster

APA must be enabled on each cluster where you want to run runbooks.

1. In the Atlas console, navigate to **APA** > **Integrations** and select the **Clusters** tab.

2. Find the cluster you want to enable and click **Enable APA**, or click **Connect cluster** to onboard a new cluster.

   <Image border={false} src="https://files.readme.io/fac87e675f30d5c8304c832f5ae1abebc571076890ce4f450791662c191ea2f8-image.png" />

> 📘 Note
>
> APA runs on a single cluster but operates org-wide. The cluster you select hosts the APA agent, which can fix images and create PRs for workloads across all clusters in your organization. You only need to enable APA on one cluster.

<Columns layout="auto">
  <Column>
    3. In the feature selection dialog, ensure **Application Performance Automation (APA)** is selected.

    4. Run the provided script in your cluster.

    5. Once complete, the cluster appears with an **Enabled** status in the Clusters tab.
  </Column>

  <Column>
    <Image align="center" src="https://files.readme.io/249cd0e8b365d8707f23b1a42e350cac95b89762e7e3ed78acc4e1b130b6097a-image.png" width="100%" alt="Enable Atlas features dialog with APA selected" />
  </Column>
</Columns>

6. Verify that the `atlas-apa-runbook-operator` component is running in your cluster:

   ```shell
   kubectl get pods -n atlas-agent
   ```

   You should see the APA runbook operator among the Atlas components:

   ```shell
   NAME                                                  READY   STATUS    RESTARTS   AGE
   atlas-agent-77cf954b4-57p9n                          2/2     Running   0          3h11m
   atlas-apa-runbook-operator-778c68645d-6b6c5          1/1     Running   0          1h36m
   ...
   ```

<br />

# Register an AI provider

APA agents use AI models to analyze code and generate fixes. You must register at least one provider before configuring runbooks. Supported providers are OpenAI, Anthropic, and Google Gemini.

1. Navigate to **APA** > **Integrations** and select the **Integrations** tab.

2. In the **AI Providers** section, click **Register provider**.

   <Image border={false} src="https://files.readme.io/7dcf9f0fbeaf9c48d10cbfd4365218a1a35d8f5b6d8854af05e9145e2f3c45f8-image.png" />

<Columns layout="auto">
  <Column>
    3. Complete the registration form:

    | Field             | Description                                                                           |
    | :---------------- | :------------------------------------------------------------------------------------ |
    | **LLM provider**  | Select your AI provider.                                                              |
    | **Models**        | Select one or more models to make available to APA.                                   |
    | **Provider name** | Enter a unique name for this configuration (for example, `openai`).                   |
    | **API key**       | Enter your provider API key. Atlas stores only the last 4 characters for reporting. |
    | **Free credits**  | (Optional) If your account has free credits, enter the amount to track usage.         |

    4. Click **Register provider**.
  </Column>

  <Column>
    ![Register an LLM provider dialog](https://files.readme.io/72bb27d03976fd2eb95a406ac3398dd609f58ad0953326ba8f96781dbc2abd89-image.png)
  </Column>
</Columns>

The provider appears in the **Managed models** table. You can register multiple providers if needed.

# Connect a Git repository

APA agents need access to your Git repositories to read source files and create pull requests.

1. In the **Integrations** tab, locate the **Git credentials** section, then click **Configure Git access**.

   <Image border={false} src="https://files.readme.io/a08e3f37b124aa85e47758c89f719d9fef84eda2e59ab662997999e8e8fd0d42-image.png" />

<Columns layout="auto">
  <Column>
    2. Select your Git provider: **GitHub** or **GitLab**.

    3. Enter a **Credential name** to identify this connection (for example, `github-production`).

    4. Choose an authentication method:

    **Token authentication** — Provide a personal access token directly. The token is encrypted and stored securely.

    | Field                     | Description                                                                        |
    | :------------------------ | :--------------------------------------------------------------------------------- |
    | **Personal access token** | A token with the required permissions (see [Before you start](#before-you-start)). |

    **Kubernetes secret authentication** — Reference a secret stored in your cluster.

    | Field           | Description                                                            |
    | :-------------- | :--------------------------------------------------------------------- |
    | **Secret name** | Name of a secret in the `atlas-agent` namespace containing the token. |

    The secret must contain a key named `GITHUB_TOKEN` (for GitHub) or `GITLAB_TOKEN` (for GitLab).

    5. Click **Save**.
  </Column>

  <Column>
    ![Configure Git access dialog](https://files.readme.io/b4b49079a6f5cc8ef544aee53a9b99ca4f80f8d2d687ec3a2760c00e5af33e2a-image.png)
  </Column>
</Columns>

The repository appears in the credentials list with a **Connected** status.

# Next steps

With APA enabled and integrations configured, you're ready to set up your first runbook:

* <Anchor label="Fix container image vulnerabilities" target="_blank" href="doc:apa-runbook-fix-image-vulnerabilities">Fix container image vulnerabilities</Anchor> — Automatically remediate vulnerabilities in your container images