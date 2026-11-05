---
updatedAt: 2026-10-08T09:20:53.000Z
---

# Supported platforms

Atlas Database Optimizer (DBO) focuses on providing exceptional database optimization for PostgreSQL and MySQL environments, and is committed to delivering the best possible experience for these databases before expanding to additional platforms.

## Deployment

DBO is container-based and runs in your Kubernetes cluster or on Amazon ECS, entirely within your infrastructure. The Helm and script flows shown in the docs deploy the Kubernetes path, into the `atlas-db-optimizer` namespace.

## Database engines

| Engine                   | Versions | Support |
| ------------------------ | -------- | ------- |
| PostgreSQL               | 13–18    | Full    |
| Amazon Aurora PostgreSQL | 13–18    | Full    |
| MySQL                    | 5.7–9    | Full    |
| Amazon Aurora MySQL      | 5.7–9    | Full    |

All PostgreSQL deployments within these version ranges are supported, and [Performance Advisor](dbo-performance-advisor.md) recommendations are validated against PostgreSQL 18.

PostgreSQL and MySQL caching is supported for any deployment: self-hosted, cloud-hosted on any provider, and managed services. The console handles mixed fleets naturally: the database server selector in the left nav lists all your connected instances side by side, so a self-managed PostgreSQL server, an AWS RDS instance, a Cloud SQL instance, and a MySQL server can all be managed from the same organization.

<Image
  src="https://files.readme.io/597dc705ec2d7e04a90f368bb30143ca37e98ed5bf8909ccdb7c678c61f5af1b-supported-platforms-01.png"
  alt="Select database server picker listing self-managed, AWS RDS, Cloud SQL, and MySQL instances side by side"
  caption="The Select database server picker listing self-managed, AWS RDS, Cloud SQL, and MySQL instances side by side."
/>

## Cloud provider discovery

* **AWS (RDS and Aurora):** connect your AWS account via cloud connect; the connect script creates a read-only, scoped IAM role for you, and DBO auto-discovers your RDS instances and Aurora clusters. Cloud connect also collects infrastructure metrics and unlocks Performance Insights, the deepest available metrics and cost analysis (see the [AWS RDS & Aurora quick start](dbo-aws-rds-and-aurora-quick-start.md)).
* **Google Cloud (Cloud SQL):** connect your Google Cloud project via cloud connect; the connect script creates the service account for you, and DBO discovers your Cloud SQL instances and collects infrastructure metrics (see the [Cloud SQL Proxy quick start](dbo-cloud-sql-proxy-quick-start.md)).
* **Self-managed and other deployments:** connect the instance manually with **+ Connect instance** (see [Getting started](dbo-getting-started.md#manual-any-postgresql-or-mysql)). Caching requires only network reachability from your DBO deployment (Kubernetes cluster or Amazon ECS) to the database; no provider-specific integration is needed.

Atlas does not access the data inside your databases at any point; see [Security and compliance](dbo-security-and-compliance.md).

## Roadmap

Additional database platforms, including MariaDB and MS SQL, are being evaluated based on customer demand and technical requirements. If your database is not currently supported, contact us to discuss your needs and learn about the roadmap.

## Next steps

<Cards columns={2}>
  <Card title="Getting started" href="https://docs.atlas.example/docs/dbo-getting-started" icon="fa-play">
    Connect your first supported database.
  </Card>

  <Card title="Security and compliance" href="https://docs.atlas.example/docs/dbo-security-and-compliance" icon="fa-shield-halved">
    What DBO collects from any supported platform.
  </Card>
</Cards>