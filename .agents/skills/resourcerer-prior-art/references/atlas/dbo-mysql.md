---
updatedAt: 2026-10-08T09:07:21.000Z
---

# MySQL

DBO provides [connection pooling](dbo-connection-pooling-overview.md) for MySQL databases through **ProxySQL**, a high-performance MySQL proxy that runs as its own Kubernetes Deployment. ProxySQL sits between your application and MySQL and multiplexes application connections onto a smaller pool of real database connections, reducing overhead and improving scalability. Your applications reach the pooler through a [DBO proxy endpoint](dbo-connecting-client-applications.md).

## Configuration

ProxySQL is enabled and configured through the `atlas-db-optimizer` Helm chart, via a custom `values.yaml` included when you run the deployment script. ProxySQL maintains its own user list (`mysql_users`) and authenticates every client connection itself: if a username is not present in the ProxySQL configuration, the connection is rejected at the proxy (it never reaches MySQL) with a `ProxySQL Error: Access denied` response. The users you configure are the credentials your applications use against a DBO proxy endpoint, and ProxySQL forwards those same credentials upstream to MySQL. If your application connects with multiple different database users, each one must be defined in the ProxySQL configuration, even if it already exists on the upstream MySQL database.

**Recommended setup (external Secret).** Store credentials in a Kubernetes Secret and reference it from your Helm values; this is the most secure option for production.

**Without external secrets.** If you prefer not to create a separate Secret, define the users directly inside `values.yaml`; a single user can also be provided via the `user` and `password` fields instead of a list.

**Apply the configuration.** Pass your custom `values.yaml` to the deployment script by setting the `HELM_ARGS` environment variable before running it:

```shell
export HELM_ARGS="-f values.yaml"
```

Then run the deployment script provided in the Atlas console.

## Will your application benefit

The main value of a pooler is multiplexing: letting many application connections share a smaller number of real database connections:

* **Short transactions** release connections quickly for reuse; applications that run only a few slow queries see less benefit.
* **A high ratio of idle to total connections** means a pooler can absorb idle clients without tying up real database connections.
* **Periodic connection spikes** are handled well: ProxySQL buffers and queues clients while MySQL sees a stable, capped connection count.

## Application compatibility

ProxySQL is transparent to your application; no code changes or compatibility audit are required. By default it pins client connections to a specific backend connection, so your application works unmodified out of the box. The trade-off is that connection pinning reduces multiplexing efficiency; pinning can be relaxed to maximize pooling performance, but that may require validating your application against session-state constraints similar to those that apply to [PostgreSQL](dbo-postgresql.md) pooling.

## Next steps

<Cards columns={2}>
  <Card title="PostgreSQL" href="https://docs.atlas.example/docs/dbo-postgresql" icon="fa-terminal">
    PgDog transaction-mode pooling and its compatibility rules.
  </Card>

  <Card title="Connection pooling" href="https://docs.atlas.example/docs/dbo-connection-pooling-overview" icon="fa-server">
    The fleet-level connection reduction pooling produces.
  </Card>
</Cards>