---
updatedAt: 2026-09-28T07:53:29.000Z
---

# Vulnerabilities

Identify container image vulnerabilities across your Kubernetes cluster and remediate them automatically.

The Vulnerabilities page shows all container images currently running in your cluster and the vulnerabilities they may expose. It consolidates information from individual image registries, third-party Helm charts, and other deployment methods, giving you a complete picture of your vulnerability posture.

> 📘 Enable vulnerability detection
>
> To use this feature, enable **Vulnerability detection** under **Settings** > **Security insights** in your cluster view. No additional installation is required. Sentinel is already installed with cost monitoring features.

## Scanning frequency and behavior

Atlas automatically scans container images for vulnerabilities with the following behavior:

* **Trigger**: Scans are initiated when new running images are detected in your cluster.
* **Detection interval**: Atlas checks for new images every 30 seconds by default.
* **Concurrent scans**: One image is scanned at a time by default to minimize resource impact.
* **Configuration**: These intervals can be adjusted using [Sentinel](sentinel.md) configuration parameters:
  * `--image-scan-interval=30s` (detection frequency)
  * `--image-concurrent-scans=1` (number of simultaneous scans)

For more information on configuring scan behavior, see [Configuring Sentinel features](sec-configuring-sentinel.md).

## What's inside

To access the Vulnerabilities page:

1. Log in to the Atlas console.
2. Select a cluster, then click **Vulnerabilities** in the sidebar.

Upon opening the Vulnerability report, you will see a list of images used by your organization across all connected clusters. The report provides various statistics and insights related to the discovered images.

> 📘 Note
>
> If all nodes are managed by the Atlas Cluster Controller, there is no need for additional action to view a complete list of images running on your clusters. In other cases, you need to enable the [Private image scanning](image-scanning.md) functionality.

### Filter images

To focus on specific images, you can use the filter located above the list. The filter allows you to narrow down the view based on criteria such as:

* Image location (cluster or namespace)
* Resource labels
* Image status
* Scan status

When applying a filter, the statistics view will update to reflect the selected scope.

### Image repositories list

The image repositories list presents a grouped view of the images used by your organization across all clusters. Images are grouped into **repositories** based on their name, with each repository containing images that share the same name but may have different tags and/or digests.

The list provides details such as vulnerability information, the number of clusters and running resources where the image is found, and the tag of the latest scanned image.

### Image details

To view detailed information about a specific image, click on the repository name and select the desired image digest from the drop-down menu. This allows you to navigate between different versions of the image.

Under the selected image digest, you can find all associated tags, details, and status. There is an `Image hierarchy` section and three tabs below the digest: `Vulnerabilities`, `Packages`, and `Affected resources`.

#### Image hierarchy

The **Image hierarchy** section displays the base image used to create the selected image, along with the layers and commands used to build each layer. This information is valuable for identifying the layer that introduced vulnerabilities and determining the necessary steps to address them.

By default, all layers are selected. You can filter out vulnerabilities or problematic packages specific to a particular layer by selecting that layer.

> 📘 Note
>
> Depending on your image build process, the Image hierarchy section may not include commands used to build certain layers. To fully leverage this feature, it is recommended to preserve image metadata during the image build process.

#### Vulnerabilities

The **Vulnerabilities** tab lists all detected vulnerabilities in the selected image, along with available fixes. The vulnerabilities are sorted by their CVSS (Common Vulnerability Scoring System) score, highlighting the most critical issues.

<Image align="center" border={true} src="https://files.readme.io/0327836c6629d2028a62f29b9aa270473ecb2016f2124f425972aaf2fa8be95a-image.png" className="border" />

#### Packages

The **Packages** tab provides information about the operating system and application-specific packages present in the image. It identifies any vulnerabilities associated with these packages and suggests available fixes.

<Image align="center" border={true} src="https://files.readme.io/55c60f1b925acd7fe578abbc8244a6941519446793c788b46ce1ae7f6c6019d2-image.png" className="border" />

#### Affected resources

The **Affected Resources** tab shows where the selected image is being used within your clusters and indicates when the deployment using the image was first created. This information helps you understand the impact and scope of any vulnerabilities.

<Image align="center" border={true} src="https://files.readme.io/054a9e0-Screenshot_2024-06-10_at_20.39.32.png" className="border" />

### Manage exceptions

The vulnerability exceptions feature allows you to remove any vulnerabilities and image repositories from the report that you deem to be false positives.

#### Add image repository to exceptions

To exclude an image repository from the list view, follow these steps:

1. Select the checkbox next to the image repository name. You can select multiple repositories at once.
2. Click on the `Except repository` button that appears above the table of repositories.
3. Provide a reason for the exception in the newly opened drawer and click the `Except repositories` button.

Alternatively, you can exclude a repository through the **Image Details** view:

1. Click on the `Except repository` button in the upper right corner of the screen.
2. Enter the reason for the exception in the drawer and click the `Except repositories` button.

#### Find an exception

When a repository is marked as an exception, it is moved to the **Excepted** view. The repository will no longer appear in the main image repository list, and the statistics section will be updated accordingly. The **Excepted** view contains a separate list of excepted repositories.

The main difference between the **Active** and **Excepted** image views is that the **Excepted** image details contain two additional fields: `REASON` and `EXCEPTION SET ON`.

#### Remove an exception

To remove exceptions, navigate to the **Excepted** repositories list, which can be found in the upper right corner of the screen, and follow these steps:

1. Select the checkbox next to the image repository name. You can select multiple repositories at once.
2. Click on the `Cancel exception` button that appears right above the table of exceptions.
3. Provide a reason for canceling the exception in the drawer and click the exception button.

You can also cancel the repository exception from the **Image details** view.

1. Select the `Cancel exception` button in the upper right corner of the screen.
2. Review the list of selected images in the newly opened drawer and press the `Cancel exception` button.

## Automated remediation

Atlas can automatically fix detected vulnerabilities using APA runbooks. The [Fix container image vulnerabilities](apa-runbook-fix-image-vulnerabilities.md) runbook detects vulnerabilities in your container images, generates fixes, and opens pull requests for your review.

The Vulnerabilities page shows fixed vulnerability counts and the latest runbook activity, including agent status, timestamps, and PR status for each image that the runbook has processed.

<Cards columns={2}>
  <Card title="Compliance" href="https://docs.atlas.example/docs/best-practices" icon="fa-clipboard-check">
    Assess your cluster against security best practices.
  </Card>
  <Card title="Private image scanning" href="https://docs.atlas.example/docs/image-scanning" icon="fa-lock">
    Configure scanning for private container registries.
  </Card>
</Cards>