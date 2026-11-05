---
updatedAt: 2026-10-08T09:47:55.000Z
---

# Pause DBO for troubleshooting

When something looks wrong (latency crept up, results look stale, errors appeared), the fastest way to tell whether the DBO proxy's caching is involved is to switch that database's cache mode to **Off**. In **Off** mode the proxy forwards every query straight to the database: traffic flows through uncached, with no caching analysis. Collect what you need, then set the mode back to **Auto** to resume caching.

**Before you start:** make sure your application is connected through the DBO proxy, and that you have access to your database's native monitoring tools for comparison.

## Switch the cache mode to Off

You change a database's cache mode from two console views:

* **Instance Caches tab**: the per-database table's **CACHE** control shows each database's current mode and changes it without leaving the table.
* **Cache configuration →**: the button at the top right of the database's **Caching** pages opens the database's cache configuration, including its mode.

With the mode set to **Off**, the proxy passes every query through to the database unchanged: nothing is served from or written to the cache, so you can compare performance with and without the cache in the path.

## Set the mode back to Auto

Switch the mode back to **Auto** from the same place. The proxy resumes its caching decisions, and traffic flows through the cache again as your application queries the database.

## Troubleshooting workflow

1. Observe the issue: latency, errors, or unexpected behavior.
2. Check DBO metrics for anomalies in cache hit rate or query patterns.
3. Set the database's cache mode to **Off** so traffic flows straight to the database, and compare performance. If the issue persists, it likely originates from the database or application layer; if it resolves, investigate the caching configuration or cache behavior.
4. Set the mode back to **Auto** once done, then act on your findings: tune what the proxy caches with override rules (see [Auto vs. Manual mode and override rules](dbo-auto-vs-manual-mode-and-override-rules.md)), or address the underlying database issue.

For high availability of the DBO proxy endpoint itself, see [Application failover configuration](dbo-application-failover-configuration.md).

## Next steps

<Cards columns={2}>
  <Card title="Auto vs. Manual mode and override rules" href="https://docs.atlas.example/docs/dbo-auto-vs-manual-mode-and-override-rules" icon="fa-sliders">
    Re-tune caching after ruling it in or out.
  </Card>

  <Card title="Application failover configuration" href="https://docs.atlas.example/docs/dbo-application-failover-configuration" icon="fa-shield-halved">
    High availability for the proxy endpoint itself.
  </Card>
</Cards>