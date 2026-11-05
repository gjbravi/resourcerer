---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Compliance

Assess your Kubernetes cluster configuration against Atlas security best practices.

The Compliance page assesses the security posture of your Kubernetes clusters against the Atlas security benchmark. It evaluates your resources and identifies misconfigurations, helping you maintain a strong compliance standing.

Atlas employs a transparent scoring system that prioritizes issues based on severity, enabling you to allocate resources effectively. The severity score and level of each issue are determined using the [CVSS v3.1 (Common Vulnerability Scoring System)](https://www.first.org/cvss/) standard.

## Access the report

To access the Compliance page:

1. Log in to the Atlas console.
2. Select a cluster, then click **Compliance** in the sidebar.

> 📘 Enable compliance checks
>
> To receive a full assessment of your cluster, enable **Compliance checks** under **Settings** > **Security insights** in your cluster view. No additional installation is required. Sentinel is already installed with cost monitoring features.

## Compliance scanning frequency

Atlas continuously monitors your cluster configuration for compliance with security best practices:

* **Scan frequency**: Compliance checks are performed every 60 seconds by default.
* **Real-time updates**: The compliance report reflects near real-time changes in your cluster configuration.
* **Configuration**: The scan interval can be adjusted using [Sentinel configuration](sec-configuring-sentinel.md) parameters.

This frequent scanning ensures that any configuration changes that impact your security posture are quickly identified and reported.

## Inside the report

The Compliance report comprehensively lists findings related to best practice violations in your Kubernetes environment. Each check in the report includes detailed information to help you understand and address the issue:

* **Check type**: Indicates whether the check is **Manual** or **Automated**.
  * Manual checks require human intervention to determine if the system's configuration meets the expected state.
  * Automated checks can be evaluated automatically against the recommended state using configuration assessment tools, i.e., a pass/fail assessment result can be automatically achieved.

* **Description of the check**: Offers a clear explanation of what the check is looking for and why it's important for your cluster's security and performance.

* **Severity level**: Based on the Common Vulnerability Scoring System (CVSS), this indicates the potential impact of the issue. Levels range from Low to Critical, helping you prioritize your remediation efforts.

* **Number of affected resources**: Provides a count of how many resources in your cluster are impacted by this particular best practice violation, giving you an idea of the issue's scope.

* **Relevant standards and configurations**: Lists the standards and specific configurations related to this check, helping you understand its broader context and importance.

* **Remediation procedure**: Offers step-by-step instructions on addressing the issue and bringing your cluster into compliance with best practices.

### Filter resources

To focus on specific resources, use the filters located above the list.

<Image src="https://files.readme.io/38a6bcf-Screenshot_2024-06-10_at_15.20.39.png" align="center" border={true} />

The filters allow you to narrow down the view based on various criteria, such as:

* Resource location (cluster or namespace)
* Resource labels
* Severity level
* Compliance standard

By applying the appropriate filters, you can quickly identify the resources of interest and assess their compliance with best practices.

### Access detailed compliance check information

Click on the check name for more information about a specific failed check.

![](https://files.readme.io/931bf12f58c6fab993c8236bab73f68b275b2ca60835b6e63078463fa21b7cec-image.png)

This action will open a detailed view containing the following:

* Comprehensive description of the problem
* Severity level of the issue
* Remediation steps to address the violation
* Number of affected resources
* Resources associated with the check
* Other relevant details

### Manage exceptions

Sometimes, you may have resources that do not comply with best practices but are considered acceptable risks. The Compliance report provides an exception feature that allows you to exclude such resources from the reporting.

#### Add an exception

To exclude specific resources from a compliance check, follow these steps:

1. Click on the check you want to manage.
2. Click the **Exceptions** button in the drawer's top right corner.
3. In the new drawer that appears, define the exception rules by specifying the location, name, and kind of resource you want to exclude. You can create multiple exception rules as needed.

<Image src="https://files.readme.io/d118e53-Screenshot_2024-06-10_at_15.41.35.png" align="center" width="75%" border={true} />

4. Review the impacted resources in the table.
5. If you are satisfied with the result, click the **Apply** button to save the exceptions.

To view the excluded resources for each check, select the `Excepted` value in the **Resources** filter in the **Resources** tab.

#### Remove or modify an exception

If you need to remove previously applied exceptions, follow these steps:

1. Click on the check you want to manage.
2. Click the **Exceptions** button in the drawer's top right corner.
3. In the exception rules drawer, modify or remove the rules as needed.
4. Review the dynamically updated list of excepted resources in the table.
5. If you are satisfied with the changes, click the **Apply** button to save the modifications.

<Cards columns={2}>
  <Card title="Vulnerabilities" href="https://docs.atlas.example/docs/image-security" icon="fa-shield-halved">
    Scan container images for known CVEs.
  </Card>
  <Card title="Configuring Sentinel features" href="https://docs.atlas.example/docs/sec-configuring-sentinel" icon="fa-gear">
    Adjust scan intervals and enable additional monitoring.
  </Card>
</Cards>