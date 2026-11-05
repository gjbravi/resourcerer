---
updatedAt: 2026-10-08T09:20:53.000Z
---

# How does it work?

Atlas Database Optimizer (DBO) is a database optimization product organized around your database instances. Its three pillars, the proxy layer, the Performance Advisor, and the observability layer, all build on the same per-instance view: the same instances and the same understanding of your workload.

## What runs in your environment

DBO is container-based and runs in your Kubernetes cluster or on Amazon ECS. The Helm and script flows shown in the docs deploy the Kubernetes path, into the `atlas-db-optimizer` namespace in your cluster (see [Getting started](dbo-getting-started.md)). Caching is handled by a single component:

**Database proxy** is a caching database proxy, written in Rust, that speaks the native PostgreSQL and MySQL wire protocols, caches read queries in in-cluster memory, and serves repeated queries in sub-millisecond time. Its ML-driven caching decisions happen inside the proxy; there is no separate intelligence component.

**Connection pooling** is delivered by an additional pod, the pooler: **PgDog** for PostgreSQL and **ProxySQL** for MySQL. When you enable connection pooling, the DBO components are redeployed with the pooler added; without it, the proxy runs on its own.

**Performance Advisor agent** runs the [Performance Advisor](dbo-performance-advisor.md)'s workload analysis. It connects to your databases with read-only credentials and analyzes query patterns from PostgreSQL's `pg_stat_statements` and MySQL's `performance_schema`, ranking index recommendations by database time reclaimed.

Cached query responses are held in memory inside your infrastructure and never leave it. The Atlas control plane, which manages and monitors DBO, receives only anonymized query metadata (see [Security and compliance](dbo-security-and-compliance.md)).

## What you see in the console

The instance's **Settings** page shows the deployed components and their state:

* **Performance advisor**: status **Active**. A lightweight agent that monitors your query workload and recommends improvements.
* **Database proxy**: status **Active**. The caching component that sits between your applications and the database, with two nested capabilities:
  * **Cache manager**: status **Enabled**. Intercepts traffic between your app and database to serve repeated queries from memory.
  * **Connection pooling**: status **Enabled**. Reuses connections instead of opening new ones.
* **Endpoints**: the upstream database endpoints used by the proxy, with **Go to configuration →** to manage them.

<Image
  src="https://files.readme.io/c76baae960a804bdf3a5dbd65946c194f171bddf6ca41ae41ab9d710bd1c88d6-how-does-it-work-01.png"
  alt="Instance Settings page showing the Performance advisor, Database proxy, Cache manager, Connection pooling, and Endpoints cards with their status badges"
  caption="The instance Settings page, showing the Performance advisor, Database proxy, Cache manager, Connection pooling, and Endpoints cards."
/>

## How data flows

Your application points at the proxy through an updated connection string: a configuration change, no code changes (see [Connecting client applications](dbo-connecting-client-applications.md)). [Caching](dbo-caching.md) and [connection pooling](dbo-connection-pooling-overview.md) apply to that traffic. The Performance Advisor works in parallel, without proxy routing: once the agent is deployed, it starts analyzing your workload, and you can begin with recommendations before enabling caching or pooling. First recommendations typically appear within about 30 minutes. For queries an index cannot fix, the advisor uses an LLM and various heuristics to propose a rewritten query; only anonymized query templates, with parameter values and identifiers replaced by placeholders, are shared with Atlas.

## Reads and writes

**Reads (SELECT):** when the proxy observes a SELECT, it computes a hash of the query and a hash of the response returned by the database. When the same query hash and the same response hash are observed a second time, caching begins; the second occurrence of the query is normally the first one served from the in-memory cache, in sub-millisecond time.

**Writes (INSERT, UPDATE, DELETE, and all other traffic):** pass through to the origin database untouched. The proxy observes these writes over the network and uses that signal to drive **Smart Invalidation**, which removes stale cache data in real time, typically in sub-millisecond time. Because invalidation is based on wire-traffic analysis, with no privileged access to database internals such as WAL, it also works with managed services like Amazon RDS and Cloud SQL.

Invalidation works in layers: row-level parsing where possible, falling back to table-level invalidation when a query is too complex to analyze precisely. As a safety net, it disables caching on queries affected by unexplained data changes.

## Caching behavior

**Time To First Hit (TTFH):** the first cache hit for a new query normally arrives on its second occurrence. TTFH is later only when an observation arrives while the connection still owes a response upstream, for example with pipelining or an unanswered `BEGIN`, because such an observation is neither served from nor stored to the cache; each skipped observation pushes the first hit out by one (TTFH = 2 + the number of skipped observations).

**Cache uniqueness:** a cache identifies a single database server. Cache uniqueness is guaranteed per combination of database name and database user, so different databases or users on the same server each get their own cache view.

**Wire-protocol compatible:** DBO works with any database driver in any programming language, with no data migration required. PostgreSQL and MySQL are supported today.

## Next steps

<Cards columns={2}>
  <Card title="Caching" href="https://docs.atlas.example/docs/dbo-caching" icon="fa-bolt">
    The in-line query cache Smart Invalidation feeds.
  </Card>

  <Card title="Connection pooling" href="https://docs.atlas.example/docs/dbo-connection-pooling-overview" icon="fa-server">
    The pool the proxy's other side maintains.
  </Card>
</Cards>