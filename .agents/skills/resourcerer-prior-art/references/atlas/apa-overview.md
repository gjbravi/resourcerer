---
updatedAt: 2026-08-21T06:56:27.000Z
---

# Overview 

<Callout icon="🔐">
  ### Private Preview Feature

  This feature is in private preview and available to select customers. To request access, provide feedback or book a demo, contact your Atlas Account Manager.
</Callout>

Application Performance Automation (APA) transforms Atlas observability data into production-ready pull requests. Using agentic AI, APA identifies issues in your Kubernetes environment, plans remediation strategies, and generates PRs that you can review and merge, reducing manual effort for teams while keeping *you* in control.

# What is APA?

APA extends Atlas beyond observability by introducing automation that acts on detected issues. Rather than reporting problems such as container vulnerabilities or configuration drift, APA agents analyze your environment, test potential fixes, and deliver tested solutions directly to your Git repositories.

The system follows a human-in-the-loop model: APA handles analysis, planning, and implementation, while you retain full control through familiar code review and merge processes.

# How it works

APA operates through runbooks—pre-configured automation workflows that target specific optimization or security goals. Each runbook follows a similar pattern:

1. **Detection**: APA identifies issues using Atlas observability data, such as vulnerable container images or resource configuration drift.

2. **Analysis**: AI agents examine your codebase to understand how the affected components are built and configured.

3. **Remediation**: Agents generate a fix tailored to your environment and create a pull request for your review.

Once you merge the PR, your existing CI/CD pipelines handle the rest—building, testing, and deploying the updated code.

# Architecture

APA runs as an agent deployed to a single cluster in your organization. From there, it operates across your entire org: fixing images, syncing recommendations, and creating PRs regardless of which cluster the affected workloads run on.

You don't need to enable APA on multiple clusters. Select one cluster to host the APA agent, and it handles automation org-wide.

# Available runbooks

APA provides runbooks targeting different aspects of your Kubernetes environment:

| Runbook                                               | Description                                                                                                                                                                                    | Status    |
| :---------------------------------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :-------- |
| **Fix container image vulnerabilities**               | Automatically remediates vulnerabilities in container images by updating base images, patching packages, or upgrading dependencies. Creates PRs to update Dockerfiles and related build files. | Available |
| **Synchronize workload optimization recommendations** | Detects drift between Workload Autoscaler recommendations and your Git manifests. Generates PRs to align your infrastructure-as-code with optimal resource configurations.                     | Available |
| **Optimize database performance**                     | Analyzes database performance using Atlas's Database Optimizer and generates PRs to add or remove indexes.                                                                                   | Available |
| **Enforce Kubernetes compliance**                     | Scans for compliance violations and security misconfigurations, then generates PRs to fix RBAC settings, control plane configurations, and workload security issues.                           | Available |

# Components and integrations

APA runbooks share two common integrations configured at the organization level:

**AI providers**

APA agents use AI models to analyze code and generate fixes. Register at least one provider before configuring runbooks. Supported providers:

* OpenAI
* Anthropic
* Google Gemini

**Git credentials**

APA agents access your Git repositories to read source files and create pull requests. GitHub and GitLab are supported.

Required token permissions:

| Provider | Permissions                           |
| :------- | :------------------------------------ |
| GitHub   | `read:discussion`, `read:org`, `repo` |
| GitLab   | `api`, `write_repository`             |

For setup instructions, see <Anchor label="Getting started" target="_blank" href="doc:apa-getting-started">Getting started</Anchor>.

# Accessing APA

To access APA in the Atlas console:

1. Log in to the Atlas console.
2. Select **APA** from the sidebar.
3. Choose **Runbooks** to view and configure available runbooks, or **Integrations** to set up AI providers and Git credentials.

# Related resources

* <Anchor label="APA Getting Started" target="_blank" href="doc:apa-getting-started">APA Getting Started</Anchor> — Configure AI providers and Git credentials
* <Anchor label="Fix container image vulnerabilities" target="_blank" href="doc:apa-runbook-fix-image-vulnerabilities">Fix container image vulnerabilities</Anchor> — Learn how APA remediates container image vulnerabilities