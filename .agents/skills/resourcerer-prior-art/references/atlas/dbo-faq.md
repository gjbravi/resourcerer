---
updatedAt: 2026-10-08T09:12:31.000Z
---

# FAQ

## General questions

### What is Atlas Database Optimizer (DBO)?

DBO is a database optimization product built on three pillars: the DBO proxy, which sits between your application and your databases to cache repeated read queries and pool connections; the Performance Advisor, which analyzes your workload and recommends concrete improvements such as missing indexes and query rewrites; and an observability layer that ties it together with dashboards and insights for every instance and database. See [How does it work?](dbo-how-does-it-work.md) for the architecture.

### How does DBO differ from traditional caching solutions like Redis?

Redis-style caching requires development effort: choosing what to cache, writing cache logic, implementing invalidation. DBO is autonomous, with zero code changes and zero configuration: it uses ML to decide what to cache and for how long.

### What databases does DBO currently support?

PostgreSQL 13 through 18 and MySQL 5.7 through 9, including Amazon Aurora, with MariaDB and Microsoft SQL Server on the roadmap. See [Supported platforms](dbo-supported-platforms.md) for the current list.

### Can I use DBO with any cloud provider?

Yes. Caching works with any PostgreSQL and MySQL deployment regardless of where it's hosted; performance insights are currently available for AWS RDS and Google Cloud SQL.

## Performance and cost benefits

### How much can DBO reduce my database costs?

DBO's first job is performance: faster query responses, lower database load, and validated recommendations from the Performance Advisor. Because DBO typically offloads a large share of reads from your database, lower infrastructure cost is a common outcome, for example when the freed-up headroom lets you downsize instances, but the exact savings depend on your workload.

### What performance improvements can I expect?

DBO typically delivers query responses in under 1ms, versus hundreds of milliseconds for direct database queries. Customers have seen up to 400x faster responses and up to 10x less compute required.

### How quickly will I see benefits after deploying DBO?

Immediately. The proxy starts learning from the first query, and full benefits typically arrive within hours.

### Does DBO improve all types of queries?

The proxy caches deterministic read queries (SELECT and similar): queries whose results depend only on stable data, not on time-volatile values such as `now()`. Gains are biggest on frequent, expensive, or large-dataset reads; writes pass through to the origin database.

## Technical details

### How does the deployment process work?

From the organization **Overview** page, click **+ Connect instance** and follow the guided flow: connect the database server, deploy DBO's components into Kubernetes or Amazon ECS, and point your application at DBO. Once deployed, **Advisor agent** and **Database proxy** show **Active** and the features show **Enabled**. See [Getting started](dbo-getting-started.md).

### Does DBO require changes to my application code?

No. The proxy operates at the database protocol level and only requires a change to your database connection string, with no changes to application logic, data models, or caching strategies.

### How does DBO maintain data consistency?

The proxy maintains consistency via Smart Invalidation: SQL query parsing to analyze which rows are read or written, table-level invalidation for complex queries, and automated detection of unexplained data changes. The proxy also implements read-after-write semantics.

### Will DBO work with my existing failover or high availability setup?

Yes, and in most cases you don't need to configure anything. The DBO proxy load-balances connections across your database instances round-robin and handles failover itself: health-check probes remove unhealthy instances from the load balancer until they recover. The proxy runs as a Kubernetes ClusterIP service, so its endpoint address stays stable while proxy pods restart or reschedule. See [Application failover configuration](dbo-application-failover-configuration.md).

### Does DBO provide connection pooling?

Yes. Connection pooling is a shipped DBO feature. It reuses connections instead of opening new ones (PgDog for PostgreSQL, ProxySQL for MySQL). The instance status row shows **CONNECTION POOLING: Enabled**, and the **Connections** views report the reduction. See [Connection Pooling](dbo-connection-pooling-overview.md).

### Does DBO store my data outside my infrastructure?

No. Cached data stays entirely within your infrastructure. Only anonymized query-pattern metadata is sent to the Atlas control plane; no PII or sensitive data leaves your environment.

## Implementation and management

### How long does it take to implement DBO?

Minutes: connect the instance and deploy DBO's components, update the connection string, then test and verify.

### How does DBO handle database schema changes or application updates?

Automatically. The proxy operates at the protocol level and makes caching decisions in real time, so there's no need to update caching rules when your application evolves.

### Can I monitor DBO's performance and effectiveness?

Yes, in the console at three levels: the organization **Overview** page, an instance's **Caches**, **Connections**, and **Performance** tabs, and each database's **Caching** views. Cache hit rates, query performance, and estimated cost savings are tracked over time: projected before you enable caching, actual after.

### How does DBO handle database upgrades or migrations?

DBO handles migrations with no action from you. When the migration completes, update the DBO connection configuration to point at the new database, and DBO continues working with it.

### What happens if the DBO cache fails?

DBO relies on Kubernetes' built-in high availability: the proxy's ClusterIP service keeps the endpoint reachable while proxy pods restart or reschedule. If you want your applications to fall back to a direct database connection when the proxy endpoint itself is unreachable, you can set that up with standard load-balancer or DNS failover patterns (for example, an NLB with Route 53 failover routing). See [Application failover configuration](dbo-application-failover-configuration.md).

## Getting started

### How can I try DBO for my databases?

Sign up for a Atlas account, click **+ Connect instance** on the organization **Overview** page, and follow the [Getting started](dbo-getting-started.md) guide.

### What permissions does DBO need to access my database?

None beyond ordinary network access to connect like any client application, from your DBO deployment with network connectivity to the database. The permissions described in [Access requirements and permissions](dbo-access-requirements-and-permissions.md) are for the optional Performance Insights feature on AWS RDS and Google Cloud SQL (cloud account permissions) and the optional Performance Advisor agent (a dedicated read-only database user). Caching itself needs no special database permissions.

### Can I customize caching behavior for specific queries?

Yes. Caching runs in **CACHE MODE: Auto** with per-query **TTL MODE: Dynamic**, needing no configuration. Advanced users can add override rules, including fixed TTLs (**TTL MODE: Fixed TTL**) for specific queries, patterns, or tables, or switch a database's cache mode to **Off**. See [Auto vs. Manual mode and override rules](dbo-auto-vs-manual-mode-and-override-rules.md).

### What if I need help implementing or troubleshooting DBO?

Atlas provides support for implementation, optimization, and any questions about the product.

## Next steps

<Cards columns={2}>
  <Card title="Getting started" href="https://docs.atlas.example/docs/dbo-getting-started" icon="fa-play">
    Set up DBO on your first database.
  </Card>

  <Card title="Security and compliance" href="https://docs.atlas.example/docs/dbo-security-and-compliance" icon="fa-shield-halved">
    Where your data lives and how DBO handles it.
  </Card>
</Cards>