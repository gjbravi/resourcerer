---
updatedAt: 2026-09-22T08:46:05.000Z
---

# Component control

The **Component control** feature in Atlas provides a centralized hub for managing all Atlas components across your connected Kubernetes clusters. Instead of navigating through multiple dashboards or using various CLI commands, you can now monitor, update, and troubleshoot all components from a single interface.

# Overview

When you access the **Component control** dashboard, you'll be greeted with a comprehensive view of your Atlas ecosystem. The dashboard presents a health snapshot of your components, showing how many are healthy, which need attention, and which might be disconnected across your infrastructure.

This visibility is crucial when managing multiple clusters, as it helps identify potential issues before they affect your operations. The dashboard is your command center for maintaining optimal Atlas functionality throughout your organization.

# Accessing Component control

To access the Component management:

1. In the Atlas console, select **Manage Organization** in the top right
2. Navigate to **Component control** in the left menu. Once there, you'll have access to all component management capabilities across your organization.

   <Image border={false} src="https://files.readme.io/350fccd19a16026145fca209fdf514b73ed2108a7973f6df4c247ef4628ddf8c-image.png" />

# Understanding the dashboard

The Component management dashboard is organized into several key sections that provide high-level insights and detailed component information.

## Component status summary

At the top of the dashboard, you'll find a visual summary of component health across your organization. This section gives you an immediate understanding of your overall component health, broken down into three categories:

* **Healthy** components are running the latest version and functioning properly
* **Warning** status indicates components that need attention, typically requiring an update
* **Disconnected** components are not functioning or cannot be reached

<Image align="center" border={false} width="50% " src="https://files.readme.io/59e70053a414af3ce00790ded76e3f1c5ab19c3742964315e932f021cab8324e-image.png" />

This summary helps you quickly gauge the health of your infrastructure and prioritize which areas need attention.

## Atlas SaaS status

The Mothership Status indicator shows the connection health between your clusters and the Atlas control plane. This information includes:

* Connection status (Active/Inactive)
* Last 30 days' uptime percentage

<Image align="center" border={false} src="https://files.readme.io/fd3a50689ab98b312cd7d5b97a45465896bf1c3754cf0bb91542e6043958bc38-image.png" />

A healthy connection ensures that your clusters can communicate with Atlas services, which is essential for most features.

## Component list

The main section of the dashboard displays a comprehensive list of all available Atlas components. For each component, you'll see:

* The category it belongs to (Autoscaler, Security, Reporting, etc.)
* A brief description of its purpose and function
* The type of feature tier it belongs to (free or paid)
* Installation status across your clusters
* Health status across all installations

<Image border={false} src="https://files.readme.io/fc39c05073d1abd0b49bbc5cb08f417783caca9adc57eee383f99f0a344e9698-image.png" />

This list becomes your primary tool for managing the Atlas component lifecycle in your organization.

# Managing components

The **Component control** dashboard makes it easy to drill down into specific components and take action when needed.

## Viewing component details

When you need more information about a specific component, click on its row in the list. This takes you to a dedicated page showing:

* A detailed description of the component's purpose
* Current and latest available versions
* Installation status across all your clusters
* Actions you can take for each installation

<Image border={false} src="https://files.readme.io/267a1151a87dc7124e7ac4e3283669269af87db997590c7519c20104cce60f0d-image.png" />

## Finding what you need

As your infrastructure grows, you may have dozens of components across many clusters. The **Component control** dashboard includes the filtering and search capabilities you would expect in the Atlas Console:

* Use the search bar to find specific components by name or function quickly
* Filter by feature to focus on specific aspects of Atlas functionality in your infrastructure
* Filter by health status to prioritize components that need attention

<Image border={false} src="https://files.readme.io/9cb543953ba22112be36f05f61598952406134d2e2ed382fea904d7254a51fb1-image.png" />

These tools help you focus on what matters most, especially when managing large-scale deployments. The example above filters only those Autoscaler-related components that need attention.

## Installing and updating components

### Using Atlas Operator

For components managed by the <Anchor label="Atlas Operator" target="_blank" href="doc:atlas-operator">Atlas Operator</Anchor>, updates are applied with a single click:

1. Navigate to **Component Control**
2. Select the component (e.g., `atlas-agent`, `spot-handler`, or `cluster-controller`)
3. Click **Update** when a new version is available
4. The Operator handles the update automatically

No scripts are required for Operator-managed component updates.

### Using Helm commands

When a component needs to be installed or updated, Atlas provides the necessary commands right in the interface. Here's how it works:

1. Identify and select a component from the list that needs attention
2. Click the **Fix** or **Update** button next to that component

   <Image border={false} src="https://files.readme.io/fdf005faedf59d767531c50c0003ddc1650e19d26c805035e94fa0ca253e92ff-image.png" />
3. A drawer opens with detailed information about the required action
4. Copy the provided script or Helm command
5. Run the command in your terminal with access to your cluster

<Image border={false} src="https://files.readme.io/a2b6e1e01d172f5ba37150d77b155a489060c833d78de2b20d68e469af841580-image.png" />

> 📘 Note for Terraform-managed components
>
> For components that were installed using Terraform, the **Component control** dashboard cannot generate direct update commands. In these cases, the drawer will provide a link to the relevant documentation, and you'll need to update your Terraform configuration manually according to the documentation guidelines.

This workflow makes it easy to keep your components up to date without having to manually search for the correct commands or configurations.

# Component ecosystem

Atlas's automation functionality is delivered through a collection of specialized components, each serving a specific purpose in your Kubernetes optimization strategy. The **Component control** dashboard provides a unified interface to manage these components across all your clusters.

For detailed information about each component, including their functions and upgrade methods, see [Hosted components](hosted-components.md). The hosted components documentation provides:

* Complete descriptions of all Atlas components organized by operating mode
* A components-by-mode table showing what each umbrella chart tag installs
* Upgrade methods (automatic vs manual)
* Links to component-specific documentation

# Troubleshooting components

When components show warning or disconnected statuses, **Component control** makes troubleshooting straightforward:

1. Click on the component row to view detailed information
2. Check the specific error reported for each affected cluster
3. Use the **Fix** or **Update** button to get the appropriate remediation command
4. Run the command in your environment
5. Verify that the status returns to healthy after a few minutes

If you need more diagnostic information, you can check component logs directly in your cluster:

```shell
kubectl logs -l app.kubernetes.io/name=<component-name> -n atlas-agent
```

This direct access to logs helps you identify and resolve issues quickly.

# Infrastructure as Code support

For organizations that manage infrastructure as code, most Atlas components can be deployed and managed through Terraform. However, the **Component control** dashboard doesn't currently generate a Terraform update configuration. Instead, when a Terraform-managed component needs attention, the dashboard will:

1. Identify the component as Terraform-managed
2. Provide a link to the relevant documentation
3. Guide you to update your Terraform configuration manually

This approach ensures that your infrastructure-as-code workflow remains consistent while still benefiting from the monitoring capabilities of the **Component control** dashboard.

See <Anchor label="Terraform provider" target="_blank" href="https://docs.atlas.example/docs/terraform-provider">Terraform provider</Anchor> for more information on managing Atlas components with Terraform.