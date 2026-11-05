---
updatedAt: 2026-09-29T08:33:20.000Z
---

# Data collection and storage

Data collection and storage practices at Atlas.

Atlas takes the confidentiality and integrity of its customer data very seriously and strives to ensure that data is protected from unauthorized access and is available when needed.

# Certifications and audits

Atlas maintains the following industry-standard certifications:

* **ISO 27001**: This certification demonstrates our commitment to information security management. It ensures we have a comprehensive system in place to manage and protect information assets.
* **SOC 2 Type II**: This attestation verifies that our service commitments and system requirements are based on the trust services criteria relevant to security, availability, processing integrity, confidentiality, and privacy.

To ensure ongoing compliance and security, we undergo third-party audits three times a year.

# Data security overview

* **No sensitive data leaves your cluster**: Atlas cannot access sensitive user data such as Kubernetes Secrets or ConfigMaps.
* **In-cluster data scrubbing**: The agent removes sensitive environment variables from workload manifests inside your cluster, before any data is sent to Atlas. Atlas never receives these values. See [How the agent removes sensitive environment variables](about-the-read-only-agent.md#how-the-agent-removes-sensitive-environment-variables).
* **Limited data collection**: The most sensitive information collected is workload names. We do not collect or process any Personally Identifiable Information (PII), Payment Card Industry (PCI) data, or Health Care (HIPAA) data.

# Customer data protection

## Data storage and encryption

* **Cloud storage**: Customer data is stored on Google Cloud Platform (GCP). The storage location depends on your account's region:
  * **US region (default)**: Data is stored in the `us-east4` (North Virginia) region.
  * **EU region**: Data is stored in the `europe-west3` (Frankfurt, Germany) region.
* **Encryption at rest**: All production data is stored on encrypted disks, enforced by our cloud service provider's encryption policy.
* **Encryption in transit**: All data in flight is encrypted with a minimum of TLS 1.2.

## Authentication

* **Third-party authentication**: User login data (emails, passwords, SSO IDs) is handled by Auth0 (Okta), a secure third-party authentication provider.
* **API token security**: We do not store user API tokens; we only store secure hashes for validation.

## Customer data usage policy

Atlas is committed to maintaining strict data privacy standards:

* **No customer data for AI/ML training**: We do not use customer data to train machine learning (ML) models that may be shared across customers. This includes:
  * No customer-provided data (e.g., application logs, content, user information, usage patterns)
  * No metadata about customer instances (e.g., cluster names, node names, workload names)
  * Complete isolation of customer data from ML training workflows
* **ML model training sources**: Our models are trained exclusively on:
  * System-generated metadata (instance types, pricing, availability)
  * Public or internal non-customer-specific datasets
  * Infrastructure-related metadata that cannot be traced to specific customers
* **Multi-tenant safety**: Our data pipelines are designed to isolate and segregate customer data to prevent leakage or inadvertent inclusion in any ML workflow
* **No PII processing**: Our ML systems do not process or store any Personally Identifiable Information (PII)
* **Service improvement only**: All customer data is used solely for providing and improving our core Kubernetes optimization services

## AI and Machine Learning models in use

Atlas uses ML techniques to improve system performance and automation:

* **Gradient Boosted Trees (GBTs)**:
  * Purpose: Predictive tasks such as Spot Instance interruption prediction and predictive autoscaling
  * Scope: Trained on infrastructure-related metadata only—never on customer-specific workloads or behavioral data
* **Foundation Models/Generative AI**:
  * Purpose: Resource utilization prediction for Workload Optimization tasks
  * Scope: Trained on synthetic and publicly available datasets; no customer data is used

Our architecture and data handling practices are designed to be compliant with industry regulations, including GDPR and SOC 2.

For more information on customer data and AI/ML usage at Atlas, please visit our [Trust security portal](https://trust.atlas.example/).

# Data retention

Atlas maintains different retention periods for various types of data:

* **Kubernetes metadata** is retained for at least 10 years to support audit requirements and compliance standards such as SOC2 and ISO27001.
* **Cluster snapshots** are retained for 21 days, providing information about the cluster state and configuration.
* **Audit logs** are retained and accessible in the console for 90 days, after which they are archived (see Audit Log Retention Policy for details).
* Data from **inactive customer accounts** is marked accordingly but never deleted to ensure data integrity and potential future access.
* **Security insights**:
  * Maximum of 5000 unique image repositories.
  * Last 20 image versions for each image repository.
  * 30 days of netflows.

These retention policies ensure that you have access to historical security data while allowing Atlas to maintain compliance with industry regulations and standards.

## Audit Log Retention Policy

Atlas maintains the following retention policy for audit logs:

* Audit logs are retained and directly accessible in the Atlas console for 90 days.
* After 90 days, audit logs are archived and no longer available through the console interface.
* Access to logs older than 90 days is available upon request through our customer support team. We provide a raw archive of historical logs when needed.
* Archived logs are retained indefinitely per our unlimited retention policy (subject to fair use).

<Callout icon="💡" theme="default">
  ### Tip

  We recommend using our open source [Audit Log Exporter](audit-log-exporter.md) to ensure continuous access to audit log data beyond the 90-day retention period. This tool allows users to store logs in their own systems indefinitely, meeting specific compliance or data retention requirements.
</Callout>

# Examples of collected data

To provide transparency, here are examples of the types of metadata we collect:

### Node metadata (extract)

```yaml
labels:
  addon.gke.io/node-local-dns-ds-ready: "true"
  beta.kubernetes.io/arch: "amd64"
  beta.kubernetes.io/instance-type: "e2-custom-4-16896"
  beta.kubernetes.io/os: "linux"
  cloud.google.com/gke-boot-disk: "pd-standard"
  cloud.google.com/gke-container-runtime: "docker2"
  cloud.google.com/gke-cpu-scaling-level: "2"
  cloud.google.com/gke-max-pods-per-node: "110"
  cloud.google.com/gke-netd-ready: "true"
  cloud.google.com/gke-os-distribution: "cos"
  failure-domain.beta.kubernetes.io/region: "us-east4"
  failure-domain.beta.kubernetes.io/zone: "us-east4-b"
  iam.gke.io/gke-metadata-server-enabled: "true"
  kubernetes.io/arch: "amd64"
  kubernetes.io/hostname: "gke-dev-master-atlas-pool-c19ff18f"
  kubernetes.io/os: "linux"
  node.kubernetes.io/instance-type: "e2-custom-4-16896"
  node.kubernetes.io/masq-agent-ds-ready: "true"
  projectcalico.org/ds-ready: "true"

```

### Pod replica metadata (extract)

```yaml
▾ metadata:
  name: "dashboard-metrics-scraper-c45b7869d"
  namespace: "kubernetes-dashboard"
  resourceVersion: "637593368"
  generation: 1
  creation Timestamp: "2022-08-16T12:10:33Z"
  ► labels: { ... }
  ► annotations: { ... }
  ► ownerReferences: { ... }
▾ spec:
  replicas: 1
  ▾ selector:
    ▾ matchLabels:
      k8s-app: "dashboard-metrics-scraper"
      pod-template-hash: "c45b7869d"
▾ template:
  ▾ metadata:
    creation Timestamp: null
  ▾ labels:
    k8s-app: "dashboard-metrics-scraper"
    pod-template-hash: "dashboard-metrics-scraper"
```

# Commitment to privacy and security

Atlas is dedicated to maintaining the highest data protection and privacy standards. We continuously update our security measures to align with industry best practices and regulatory requirements.

To learn more about our security policies and compliance, head over to the [Trust security portal](https://trust.atlas.example/).