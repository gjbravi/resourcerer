# Uninstall

> Uninstall Orion from a cluster


Uninstall Orion when you want to stop optimizing workloads on a cluster or before you delete a cluster from Orion:

```sh
helm uninstall orion -n orion-system
```

Uninstalling removes the Orion components but leaves the Orion CRDs — and any custom resources, such as optimization configurations — in place. This is expected: it ensures your optimization configuration survives a reinstall.
