---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Security insights

Detect vulnerabilities in container images and assess cluster compliance with security best practices.

Atlas provides two security insight features that help you maintain a strong security posture across your Kubernetes clusters. Both features are powered by [Sentinel](sentinel.md) and can be enabled under **Settings** > **Security insights** in your cluster view. No additional installation is required.

## Vulnerabilities

The [Vulnerabilities](image-security.md) page identifies all container images currently running in your cluster and the CVEs they may expose. It consolidates data from individual image registries, third-party Helm charts, and other deployment methods into a single view. Images are scanned automatically when new running images are detected, by default every 30 seconds.

Each image is scored by severity using CVSS, and you can filter by cluster, namespace, severity level, or resource label to focus on what matters. Detected vulnerabilities can be remediated automatically using the [Fix container image vulnerabilities](apa-runbook-fix-image-vulnerabilities.md) APA runbook.

## Compliance

The [Compliance](best-practices.md) page evaluates your Kubernetes environment against the Atlas security benchmark. Checks run continuously (every 60 seconds by default) and cover misconfigurations, RBAC issues, and kubelet/system-component settings validated via kube-bench.

Each finding includes a severity score (CVSS v3.1), a description of the check, the number of affected resources, and step-by-step remediation instructions. You can manage exceptions for resources that represent accepted risks.

<Cards columns={2}>
  <Card title="Vulnerabilities" href="https://docs.atlas.example/docs/image-security" icon="fa-shield-halved">
    Scan container images for known CVEs.
  </Card>
  <Card title="Compliance" href="https://docs.atlas.example/docs/best-practices" icon="fa-clipboard-check">
    Assess your cluster against security best practices.
  </Card>
</Cards>