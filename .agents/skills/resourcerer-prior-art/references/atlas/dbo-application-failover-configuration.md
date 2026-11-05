---
updatedAt: 2026-10-08T09:12:31.000Z
---

# Application failover configuration

Your application needs no failover configuration of its own: the DBO [proxy](dbo-how-does-it-work.md) handles failover and load balancing for you. Your application connects to one of the proxy's two endpoints (read-write or read-only), and the proxy spreads its connections across the database instances behind it, round-robin. Health-check probes run continuously: an instance that fails its checks is removed from the load balancer until it is healthy again, so new connections only ever land on instances that can serve them.

The proxy endpoint itself is resilient too. The proxy runs as a Kubernetes ClusterIP service, which provides stable IP routing to healthy proxy pods: while pods restart or are rescheduled, the service address stays the same and Kubernetes automatically routes your connections to available proxy pods at the IP level.

For the proxy hostname and port to use in your connection string, see [Connecting client applications](dbo-connecting-client-applications.md).

## Optional: falling back to a direct database connection

Teams that want an application-level safety net, reconnecting directly to the database if the proxy endpoint itself becomes unreachable, can build one with standard load-balancer or DNS failover patterns, for example an NLB in front of the proxy with Route 53 failover routing to the direct database endpoint.

## Next steps

<Cards columns={2}>
  <Card title="Connecting client applications" href="https://docs.atlas.example/docs/dbo-connecting-client-applications" icon="fa-plug">
    The proxy hostname, port, and connection string.
  </Card>

  <Card title="Pause DBO for troubleshooting" href="https://docs.atlas.example/docs/dbo-pause-dbo-for-troubleshooting" icon="fa-wrench">
    Rule the cache in or out while you investigate an issue.
  </Card>
</Cards>