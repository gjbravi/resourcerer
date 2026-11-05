---
updatedAt: 2026-10-08T09:20:53.000Z
---

# Introduction

Atlas Database Optimizer (DBO) is a database optimization product for PostgreSQL and MySQL, organized around your database instances rather than cache groups. Its three pillars, the proxy layer, the Performance Advisor, and the observability layer, all build on the same per-instance view: the same instances and the same understanding of your workload.

DBO addresses the optimization work that has no natural owner in most organizations: queries that become full table scans as traffic grows, connections exhausted the moment autoscaling kicks in, missing indexes that quietly inflate latency and cost. DBO performs this continuous work for you. The goal is not to replace a DBA; it is to make that work no longer depend on one person's institutional knowledge.

## Key capabilities

### Performance Advisor

The [Performance Advisor](dbo-performance-advisor.md) is a lightweight agent with read-only database credentials. It analyzes query patterns from PostgreSQL's `pg_stat_statements` and MySQL's `performance_schema` and ranks index recommendations by database time reclaimed, the change most likely to improve overall latency. For queries an index cannot fix, it uses an LLM and various heuristics to propose a rewritten query with a predicted improvement and a side-by-side execution plan comparison. The advisor is read-only: it never changes your schema, and every recommendation requires human review. It works immediately after deployment, with no proxy routing required; first recommendations typically appear within about 30 minutes.

### Caching

The DBO proxy [caches](dbo-caching.md) reads between your application and your database over the native wire protocols, with no code changes and sub-millisecond cached responses. Invalidation is driven by wire-traffic analysis: the proxy observes writes over the network and invalidates affected entries as data changes, with no privileged access to database internals such as WAL, so it also works with managed services like Amazon RDS and Cloud SQL. The proxy operates in three modes (see [Auto vs. Manual mode and override rules](dbo-auto-vs-manual-mode-and-override-rules.md)):

* **Passthrough mode**: observational only. DBO learns your query patterns and builds the baseline for Auto mode, caching nothing.
* **Auto mode**: the ML model decides what to cache; read-heavy queries with stable result sets become candidates, while write-path queries stay untouched.
* **Off mode**: requests go straight to the database without query analysis, and nothing is cached.

On read-heavy workloads with over 70% cache hits, cache-hit queries are up to 400x faster, the database consumes up to 10x less compute, and infrastructure costs typically drop by 30% or more.

### Connection pooling

The DBO proxy multiplexes connections: your application opens many connections to the proxy, which keeps a controlled pool to the database. [Connection pooling](dbo-connection-pooling-overview.md) means the database sees a predictable connection count regardless of pod autoscaling, preventing the exhaustion that occurs when autoscaling runs into default limits such as PostgreSQL's 100 or MySQL's 151 `max_connections`. PostgreSQL is proxied through [PgDog](dbo-postgresql.md) and MySQL through [ProxySQL](dbo-mysql.md).

### Observability

The observability layer ties the pillars together: health metrics, query details, connection stats, cache hit rates, and advisor recommendations in one per-instance view. The console mirrors this: the Database optimization [Overview](dbo-organization-overview.md) page lists onboarded and not-onboarded instances with **Top 5 recommendations**, **Cache hits**, and **Connection pooling** cards, and the [instance dashboard](dbo-instance-dashboard.md) and [database dashboard](dbo-database-dashboard.md) cover **Dashboard**, **Caches**, **Connections**, **Performance**, **Performance advisor**, **Caching**, **Audit log**, and **Settings**; component status reads **Active** or **Enabled**.

<Image
  src="https://files.readme.io/f617b7b5b69b837692540fb8cda20c01197df54dd987b00566d02cc8b5338811-introduction-01.png"
  alt="Organization Overview page with the Onboarded and Not onboarded tabs, the Top 5 recommendations, Cache hits, and Connection pooling cards, and the instances table"
  caption="The organization Overview page, with the Onboarded and Not onboarded tabs, the Top 5 recommendations, Cache hits, and Connection pooling cards, and the instances table."
/>

## How it works at a glance

The DBO agent and proxy run in your Kubernetes cluster or on Amazon ECS, container-based and entirely within your infrastructure.

Your application points at the proxy through an updated connection string: a configuration change, no code changes.

The Performance Advisor analyzes your workload in parallel and does not require proxy routing, so you can start with recommendations before enabling caching or pooling.

Cached data stays in in-cluster memory and never leaves your environment; for query rewrite analysis, only anonymized query templates, with parameter values and identifiers replaced with placeholders, are shared with Atlas (see [Security and compliance](dbo-security-and-compliance.md)).

Discovery of your first instances completes within about 10 minutes, and you can enable Performance Advisor, caching, and connection pooling one at a time, in any order.

## Next steps

<Cards columns={2}>
  <Card title="Getting started" href="https://docs.atlas.example/docs/dbo-getting-started" icon="fa-play">
    Setup paths to connect your first database instance.
  </Card>

  <Card title="Supported platforms" href="https://docs.atlas.example/docs/dbo-supported-platforms" icon="fa-server">
    Engines, versions, and cloud providers DBO supports.
  </Card>
</Cards>