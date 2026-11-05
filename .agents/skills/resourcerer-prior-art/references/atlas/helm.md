---
updatedAt: 2026-05-12T13:39:45.000Z
---

# Helm

### Does Atlas sometimes remotely update versions of Atlas AI helm charts?

Atlas updates the Evictor chart from the autoscaler to match the latest version available if it’s Atlas-managed. The cluster-controller component management is always Atlas-managed, so it will always be updated automatically.

<br />

***

### Do Agent, Cluster Controller, Spot Handler, Evictor, Sentinel get updates automatically?

We have a process for updating the cluster-controller Helm chart we can trigger it from our side.

The following components are not updated automatically:

* Spot handler,
* Atlas agent.