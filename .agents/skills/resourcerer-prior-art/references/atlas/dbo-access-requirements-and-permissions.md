---
updatedAt: 2026-10-19T09:18:38.000Z
---

# Access requirements and permissions

What you need depends on how you connect your databases to DBO. Proxying your traffic always requires only network connectivity; the Performance Advisor agent uses its own dedicated read-only database user, and cloud permissions are created for you by the connect script when you connect a cloud account.

## All setups

* A Kubernetes cluster or Amazon ECS environment where the DBO components run. The Kubernetes flow shown in the docs lands in the `atlas-db-optimizer` namespace, and you need enough cluster access to run the DBO deployment script or Helm chart there.
* Network connectivity between that environment and your database, at the database's host and port (for example, 5432 for PostgreSQL, 3306 for MySQL), with the same level of access any application would need.
* Nothing beyond that on the database side: your application connects through the proxy as it would connect to any database endpoint, and no special or additional database permissions are needed for the proxy itself.
* One application change at runtime: point your application's connection string at the DBO proxy instead of at the database directly.

## Automated discovery on AWS

To enable auto-discovery, infrastructure metrics, and Performance Insights for Amazon RDS or Aurora, connect your AWS account. Run the script the console provides, in AWS CloudShell or a local CLI; for the full flow, see the [AWS RDS & Aurora quick start](dbo-aws-rds-and-aurora-quick-start.md).

The script creates a cross-account IAM role in your account with these read-only permissions: the role cannot modify your databases or their configuration.

| Permission                       | What it does                                                                                                          |
| -------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `rds:DescribeDBInstances`        | Lists your RDS instances and their details (engine, size, status, configuration) for auto-discovery and metrics       |
| `rds:DescribeDBClusters`         | Lists your Aurora clusters and their details (for multi-instance cluster setups)                                      |
| `rds:DescribeDBClusterEndpoints` | Reads cluster connection endpoints (writer and reader URLs) so DBO can tell primary from replicas                     |
| `cloudwatch:GetMetricData`       | Reads CloudWatch performance metrics (CPU, IOPS, connections, latency) that power the Dashboard and Performance views |
| `pi:GetResourceMetrics`          | Reads detailed database-internal metrics from Performance Insights (query-level load); powers the DBO **Health** tab  |
| `pi:DescribeDimensionKeys`       | Reads the top contributors to database load in Performance Insights (top SQL queries, top wait events)                |
| `account:ListRegions`            | Lists the regions enabled for your account, so the connect flow can locate your databases                             |
| `ec2:DescribeAvailabilityZones`  | Lists availability zones, so DBO knows where your instances and databases run                                         |

The script attaches this IAM policy statement:

```json
{
  "Effect": "Allow",
  "Action": [
    "ec2:DescribeAvailabilityZones",
    "rds:DescribeDBInstances",
    "pi:GetResourceMetrics",
    "rds:DescribeDBClusters",
    "rds:DescribeDBClusterEndpoints",
    "pi:DescribeDimensionKeys",
    "cloudwatch:GetMetricData",
    "account:ListRegions"
  ],
  "Resource": "*"
}
```

AWS requires `"Resource": "*"` because these describe and list APIs do not support resource-level scoping.

## Automated discovery on Google Cloud

To discover your Cloud SQL instances, connect your Google Cloud project or organization: run the script the console provides. For the full flow, see the [Cloud SQL Proxy quick start](dbo-cloud-sql-proxy-quick-start.md). Connect an organization to discover instances across all its projects, or a project to limit discovery to that project.

The script creates a read-only service account and grants it the matching permission set: the service account cannot modify your databases or their configuration.

### Organization scope

| Permission                      | What it does                                                              |
| ------------------------------- | ------------------------------------------------------------------------- |
| `resourcemanager.projects.list` | Lists all projects in the organization                                    |
| `cloudsql.instances.list`       | Lists all Cloud SQL instances (PostgreSQL, MySQL) across every project    |
| `cloudsql.databases.list`       | Lists the databases inside each Cloud SQL instance                        |
| `monitoring.timeSeries.list`    | Reads monitoring metrics (CPU, memory, connections) from Cloud Monitoring |

### Project scope

| Permission                   | What it does                                   |
| ---------------------------- | ---------------------------------------------- |
| `cloudsql.instances.list`    | Lists Cloud SQL instances in that project only |
| `cloudsql.databases.list`    | Lists databases inside those instances         |
| `monitoring.timeSeries.list` | Reads monitoring metrics for that project only |

At organization scope, the script grants:

```text
resourcemanager.projects.list
cloudsql.instances.list
cloudsql.databases.list
monitoring.timeSeries.list
```

At project scope:

```text
cloudsql.instances.list
cloudsql.databases.list
monitoring.timeSeries.list
```

## Adding database instances manually

Add an instance with **+ Connect instance** in the console: self-hosted servers, other cloud providers, or any instance outside cloud discovery. The requirement is network reachability from your DBO deployment to the database's host, port, and engine type. For the supported cloud providers, see [Supported platforms](dbo-supported-platforms.md).

## Performance Advisor agent permissions

The Performance Advisor agent connects with a dedicated, read-only database user, separate from your application credentials:

* **PostgreSQL:** the agent user needs the `pg_read_all_stats` role (`GRANT pg_read_all_stats TO performance_advisor_agent;`) for read-only access to the query statistics views, and the `pg_stat_statements` extension must be enabled on the database.
* **MySQL:** grant `SELECT` on `performance_schema.*` for query statistics, plus `REFERENCES` on each database you want analyzed. `REFERENCES` is used only for `information_schema` lookups and never allows reading, modifying, or deleting table data.

Full setup steps are in [Installing the agent for PostgreSQL](dbo-installing-the-agent-for-postgresql.md) and [Installing the agent for MySQL](dbo-installing-the-agent-for-mysql.md).

## What Atlas can access

All cloud credentials created by the connect script (the AWS IAM role and the Google Cloud service account) are read-only and scoped to what each integration needs, and Atlas does not access the data inside your databases. For caching, all data remains within your infrastructure; only anonymized metadata about query patterns is sent to the Atlas control plane. See [Security and compliance](dbo-security-and-compliance.md) for details.

## Related resources

<Cards columns={2}>
  <Card title="Security and compliance" href="https://docs.atlas.example/docs/dbo-security-and-compliance" icon="fa-shield-halved">
    Where your data lives and how DBO handles what it collects.
  </Card>

  <Card title="Supported platforms" href="https://docs.atlas.example/docs/dbo-supported-platforms" icon="fa-server">
    Engines, versions, and cloud providers DBO supports.
  </Card>
</Cards>