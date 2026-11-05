---
updatedAt: 2026-10-08T09:12:31.000Z
---

# Audit log

Each DBO instance has its own **Audit log**, accessible from the instance-level navigation, alongside **Dashboard**, **Caches**, **Connections**, **Performance**, and **Settings** (AWS- or Google Cloud-connected instances also have a **Health** tab; see [Instance dashboard](dbo-instance-dashboard.md)). Because the audit log is per instance, check the instance where the change happened: an entry about a caching or proxy change appears on the instance it was made on, not in a fleet-wide list.

The **Audit log** records all changes made via the console UI or the API for that instance, from changing deployment details to enabling or disabling components and updating configuration, so you can track changes, troubleshoot issues, and maintain compliance.

## Interface

Entries are shown in a table with a timestamp, an operation name describing the change, and who or what initiated it: a user working in the console or a call to the API. Select an entry to expand it and see the full detail: the affected instance, the operation parameters (for example, which component was enabled or which configuration field changed), and the before and after values where the console tracks them.

For retention and exporting, see the organization-wide [audit log](audit-log.md) documentation.

## Next steps

<Cards columns={2}>
  <Card title="Instance dashboard" href="https://docs.atlas.example/docs/dbo-instance-dashboard" icon="fa-gauge-high">
    The per-instance console view this log belongs to.
  </Card>

  <Card title="Access requirements and permissions" href="https://docs.atlas.example/docs/dbo-access-requirements-and-permissions" icon="fa-key">
    What credentials each DBO action runs under.
  </Card>
</Cards>