---
updatedAt: 2026-10-08T09:12:31.000Z
---

# Security and compliance

Atlas takes data processing and storage seriously, storing only the minimum data required to provide the service. Your data and cached query responses never leave your infrastructure; only anonymized query metadata reaches the Atlas control plane.

## Where data lives

* **Cached query responses** stay in memory inside the Proxy, a self-hosted component in your Kubernetes cluster or on Amazon ECS that is subject to the security and access policies of your own infrastructure. TTL values are automatically set per query; see [Auto vs. Manual mode and override rules](dbo-auto-vs-manual-mode-and-override-rules.md) for how TTLs are managed.
* **The Atlas control plane** receives only anonymized metadata about query patterns, which powers the console's analytics and recommendations. Atlas does not access the actual data within your databases.
* **No write access:** the read-only, scoped credentials created by the connect script (an IAM role for AWS, a service account for Google Cloud) do not allow Atlas to modify your database configuration. [Access requirements and permissions](dbo-access-requirements-and-permissions.md) lists exactly what the script creates.

## SQL anonymization

Query SQL is processed in real time inside your infrastructure and anonymized at ingestion; the [Performance Advisor](dbo-performance-advisor.md) works from these anonymized templates. All literal and parameter values, including those in prepared statements, are hashed, so no personally identifiable information is stored. Database names, table names, column names, and SQL keywords are not anonymized, because they are needed to analyze query structure and target recommendations.

For example, `SELECT * FROM employees WHERE emp_no = 1001` is stored as `SELECT * FROM employees WHERE emp_no = [?]`, where `[?]` replaces the hashed value. A query with no parameters, such as `SELECT AVG(salary) FROM salaries`, is stored unchanged.

## Retention

* Anonymized query SQL is retained for 14 days.
* Aggregated metrics used for longer-term views are retained for 90 days.
* After these periods, the data is automatically deleted.

## Related resources

<Cards columns={2}>
  <Card title="Access requirements and permissions" href="https://docs.atlas.example/docs/dbo-access-requirements-and-permissions" icon="fa-key">
    Exactly what the connect script creates, per provider.
  </Card>

  <Card title="How DBO is organized" href="https://docs.atlas.example/docs/dbo-how-does-it-work" icon="fa-gear">
    What runs in your environment versus the control plane.
  </Card>
</Cards>