---
updatedAt: 2026-09-24T07:35:04.000Z
---

# Probe and lifecycle behavior

Technical documentation explaining how Kubernetes probes and lifecycle hooks behave during container live migration, including what is skipped, what runs, and how to customize behavior for edge cases.

Container live migration preserves your container's complete runtime state: memory, filesystem, process state, and open file descriptors. Since everything has been migrated, the startup logic does not need to run again on the destination node.

This page explains what runs, what's skipped, and how to customize behavior for edge cases.

# Default behavior

| Component        | On the restored pod      | Why                                                              |
| :--------------- | :----------------------- | :--------------------------------------------------------------- |
| Init containers  | Skipped                  | Files and configuration already exist in the migrated filesystem |
| PostStart hooks  | Skipped                  | Already executed; re-running could cause duplicate side effects  |
| Startup probes   | Skipped                  | Container already started successfully                           |
| Liveness probes  | Run                      | Must verify container health after migration                     |
| Readiness probes | Run                      | Must verify container can serve traffic                          |
| PreStop hooks    | Skipped during migration | Runs normally when the pod is eventually deleted                 |

For most applications, this default behavior is correct, and no configuration changes are needed.

# Probe type matters

How a probe is skipped depends on its type:

| Probe type  | Behavior on restored pod                    |
| :---------- | :------------------------------------------ |
| `exec`      | Skipped (returns success without executing) |
| `httpGet`   | **Runs normally** (cannot be skipped)       |
| `tcpSocket` | **Runs normally** (cannot be skipped)       |
| `grpc`      | **Runs normally** (cannot be skipped)       |

**This distinction is important.** If your startup probe uses `httpGet`, `tcpSocket`, or `grpc`, it will execute on the restored pod. Ensure the probe is safe to run on an already-running container.

Most health check endpoints are idempotent and handle this correctly. If yours is not, consider switching to an exec-based probe or ensuring the endpoint handles repeated calls gracefully.

# Probe failures during migration

During the brief checkpoint window, probes cannot reach the frozen container and will time out. This does not cause problems because:

* The pod is marked for deletion, so the kubelet ignores probe failures
* Migration typically completes before `failureThreshold` is reached
* **Probe failure counts do not carry over** to the restored pod; the destination kubelet starts fresh with zero failures

If you use aggressive probe settings (for example, `periodSeconds: 1` with `failureThreshold: 3`), consider adjusting to more moderate values like `periodSeconds: 5` or `periodSeconds: 10`.

# Overriding default behavior

In rare cases, you may need to run startup logic on the restored pod. Use these annotations on your pod:

| Annotation                                                   | Effect                                                 |
| :----------------------------------------------------------- | :----------------------------------------------------- |
| `atlas.example/restore-run-startup-probe: "container1,container2"` | Forces startup probes to run for specified containers  |
| `atlas.example/restore-run-post-start-exec: "container1"`          | Forces PostStart hooks to run for specified containers |
| `atlas.example/restore-run-init-containers: "init1,init2"`         | Forces specified init containers to run                |

**Example:**

```yaml
metadata:
  annotations:
    atlas.example/restore-run-startup-probe: "app"
```

**When to use these annotations:**

* Node-specific validation (checking for hardware that may differ on the destination node)
* External service registration that relies on pod UID or pod name (which change during migration)

> ⚠️ Use with caution
>
> * Re-running PostStart hooks can cause duplicate side effects (duplicate registrations, duplicate cache entries)
> * Re-running startup probes delays pod readiness
> * Re-running init containers increases migration time and may fail if files already exist

# Frequently asked questions

<Accordion title="Will my application notice it was migrated?">
  No. The application continues from the exact point where it was frozen. Memory state, variables, open files, and network connections (with TCP migration enabled) are all preserved.

  Applications with very short TCP timeouts (under 500ms) may experience connection timeouts during the checkpoint window. Consider appropriate timeout values or retry logic for critical connections.
</Accordion>

<Accordion title="What if my init container downloads files?">
  The files are preserved. The container's entire filesystem (including files not in mounted volumes) is migrated to the destination node.
</Accordion>

<Accordion title="My PostStart hook registers with an external service. Will that still work?">
  Usually yes—the registration remains valid because pod state is preserved. However, if your registration relies on pod UID or pod name (which change during migration), consider using the override annotation or updating your service discovery to handle pod identity changes.
</Accordion>

<Accordion title="What about PreStop hooks?">
  PreStop hooks are skipped during migration because the original pod is deleted as part of the migration process, not through graceful shutdown. PreStop hooks run normally when the restored pod is eventually deleted through standard Kubernetes operations.
</Accordion>

# See also

* [Container Live Migration Overview](clm-overview.md)
* [Requirements and Limitations](clm-requirements-and-limitations.md)
* [Labels, Annotations, and Events Reference](clm-reference-labels-and-annotations.md)