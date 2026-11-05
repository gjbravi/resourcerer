---
updatedAt: 2026-08-24T12:48:48.000Z
---

# Fix container image vulnerabilities

Automatically remediate vulnerabilities in container images and generate pull requests for review.

<Callout icon="🔐">
  ### Private Preview Feature

  This feature is in private preview and available to select customers. To request access, provide feedback or book a demo, contact your Atlas Account Manager.
</Callout>

The Fix container image vulnerabilities runbook automatically detects vulnerabilities in your container images, generates fixes, and opens pull requests for your review. You stay in control. Nothing changes in your repositories until you review and merge each PR.

# The problem

Container images accumulate vulnerabilities over time. Even if you ship clean images today, new CVEs are discovered regularly, and within months, your images can have dozens of vulnerabilities. Fixing them manually requires developer time that could be better spent on feature work.

This runbook addresses the gap by continuously monitoring your running images, generating fixes, and delivering them as PRs.

# How it works

The runbook handles the complete remediation cycle:

1. **Detect** — Finds vulnerabilities in your container images.
2. **Fix** — Updates Dockerfiles to use patched base images or upgraded packages, then opens a pull request for your review.
3. **Rebuild** — After you merge the PR, your existing CI/CD pipeline rebuilds and pushes the secure image.
4. **Update** — Optionally updates Kubernetes manifests in your repository to reference the fixed image, preventing drift on future deployments.

The agent experiments with multiple remediation approaches, testing each for build success and vulnerability reduction. Then it selects the best solution based on effectiveness and minimal code changes.

# Set up the runbook

This section walks through configuring the runbook for the first time.

## Before you start

Ensure you have:

* A cluster with APA enabled (see <Anchor label="Getting started" target="_blank" href="doc:apa-getting-started">Getting started</Anchor>) and **Vulnerability detection** enabled under **Settings** > **Security insights**
* At least one AI provider registered
* Git credentials configured for your repository

## Step 1: Select a cluster

1. Navigate to **APA** > **Runbooks** and locate the **Fix container image vulnerabilities** card.

2. Click **Set up runbook**.

   <Image border={false} src="https://files.readme.io/698ccf219e7e58f1b1eb86be379502647f0fd4cb2d411653cc053c23ceb6b972-image.png" />

3. Select the cluster where you want to deploy the APA agent. The cluster must have both Security and APA enabled.

4. Click **Next: Integration**.

   <Image border={false} src="https://files.readme.io/b8df1fa2f504549084ce8377522351a74ce2b89785856fc19826c31fffaf3929-image.png" />

## Step 2: Configure integrations

The Integration step connects the runbook to your repositories, AI models, and container registries.

### Grant access to repositories

Select the Git credentials that allow APA to read your code and create pull requests. If you haven't configured credentials yet, click **Configure Git access** to add them.

### Select AI models

Choose the model family the agent uses to analyze code and generate fixes. Registered providers appear in the list with their available models.

### Connect container image registries

Connect your container registries so the agent can identify base images used in your containers. Click **Connect registry** to add registries. Supported registries include AWS ECR, GCR, ACR, Docker Hub, and private registries.

### Map images to repositories

Map container images to their source Git repositories. This tells the agent where to find Dockerfiles and create PRs.

| Field          | Description                                                           |
| :------------- | :-------------------------------------------------------------------- |
| **Image**      | Select one or more container images. Use exact names or wildcards.    |
| **Repository** | Select the Git repository containing the Dockerfile for these images. |

Click **Add** to create additional mappings as needed.

<Image border={false} src="https://files.readme.io/0a8ce303af593422e0b82e883416ed32efc9331543688dc6d780d35a42cefa98-image.png" />

Once all integrations are configured, click **Next: Configuration**.

## Step 3: Configure runbook settings

The Configuration step controls how the agent operates.

### Analytics collection

Choose whether to send analytics to Atlas. Analytics help improve fix accuracy and agent performance. The agent operates entirely within your cluster—only small, relevant code snippets (not full files) are sent when enabled.

| Option                | Description                                |
| :-------------------- | :----------------------------------------- |
| **Yes** (Recommended) | Send analytics to improve fix suggestions. |
| **No**                | Disable analytics collection.              |

### Severity threshold

Set the minimum severity level for vulnerabilities the agent will remediate. The agent fixes vulnerabilities at or above this level.

| Threshold    | Vulnerabilities addressed  |
| :----------- | :------------------------- |
| **Medium**   | Medium, High, and Critical |
| **High**     | High and Critical          |
| **Critical** | Critical only              |

<Image border={false} src="https://files.readme.io/e99e6a2871df1e68270200b40952d64d93d3f7e3b3cd2132545503d63171ec59-image.png" />

### Advanced settings

Expand **Advanced settings** to configure additional options:

| Setting                            | Description                                                                                                                                                                                                             |
| :--------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Container image prioritization** | Choose whether the agent considers base images from all container registries or only private registries when selecting less vulnerable alternatives.                                                                    |
| **Pull request frequency**         | Minimum interval between PRs. The agent won't create new PRs more frequently than this. Cannot be less than 10 minutes.                                                                                                 |
| **Automatic manifest updates**     | When enabled, the agent updates Kubernetes manifests in your repository to reference fixed images after PRs are merged. Disable this if you use GitOps tools (ArgoCD Image Updater, Flux) that handle manifest updates. |
| **Helm chart values**              | (Optional) Customize the APA agent deployment using a Helm values YAML file. Edit only if you intend to change the Helm chart configuration.                                                                            |

<Image border={false} src="https://files.readme.io/e5295902d3e1a26ab0056011721b1294b1dc4ac416fffc83e2b3f434045d25e6-image.png" />

Click **Save runbook** to finish setup and have the agents begin scanning your repositories.

# Monitor the runbook

After setup, you land on the runbook report page, where you can track agent activity and results.

<Image border={false} src="https://files.readme.io/8dbe5b41bdc29bdb943c003d313d6c22f3cee0ff57d133ddd544e4e740a23f45-image.png" />

The report shows:

| Metric                          | Description                                                                                           |
| :------------------------------ | :---------------------------------------------------------------------------------------------------- |
| **Total vulnerabilities fixed** | Cumulative count of vulnerabilities addressed, broken down by severity (Critical, High, Medium, Low). |
| **Pull requests open**          | Number of PRs awaiting your review.                                                                   |
| **Fixes pending**               | Fixes the agent is currently working on.                                                              |
| **Images rebuilt**              | Images that have been rebuilt with fixes.                                                             |
| **Workload manifests updated**  | Manifests updated to reference fixed images.                                                          |
| **Dev days saved**              | Estimated developer time saved by automated fixes.                                                    |

## Manage the runbook

Click **Actions** in the top right to:

| Action                 | Description                               |
| :--------------------- | :---------------------------------------- |
| **Edit configuration** | Modify runbook settings.                  |
| **Delete runbook**     | Remove the runbook and its configuration. |

<Image border={false} src="https://files.readme.io/6cf296e4d4bab299ef9e6edbe7d15f7cfe26472603294baed7f2037e21e3d976-image.png" />

# FAQ

**Do I stay in control of changes?**
Yes. APA creates pull requests for both image fixes and manifest updates. Nothing changes in your repositories until you review and approve each PR.

**What happens if a fix breaks something?**
APA uses your existing CI/CD pipelines to validate changes. If tests fail, the PR remains open for you to review. You can investigate the failure, adjust the fix, or close the PR.

**Do I need to change my workflows?**
No. Fixes and manifest updates flow through your standard Git process, so you keep your current way of working.

**Which registries are supported?**
APA works with any OCI-compatible registry, including AWS ECR, GCR, ACR, Docker Hub, and private registries.

**Can I control which vulnerabilities are fixed?**
Yes. During runbook setup, you can set a severity threshold to limit fixes to critical and high-severity issues only.

# Next steps

* <Anchor label="Synchronize workload optimization recommendations" target="_blank" href="doc:apa-runbook-sync-workload-autoscaler-recommendations">Synchronize workload optimization recommendations</Anchor> — Keep your Git manifests aligned with Workload Autoscaler recommendations
* <Anchor label="APA Overview" target="_blank" href="doc:apa-overview">APA Overview</Anchor> — Learn more about Application Performance Automation