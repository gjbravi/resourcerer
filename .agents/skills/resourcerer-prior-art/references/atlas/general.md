---
updatedAt: 2026-05-12T13:39:45.000Z
---

# General

### Does Atlas have any general recommendations for EKS to speed up cluster scaling?

Let's explain this using an example scenario:

When a user scales 1000+ pods in parallel, they take around 5 to 8 minutes to get scheduled and running as the underlying node allocation is slower. Can Atlas speed this up?

Here are a few solutions:

1. Use Bottlerocket AMI.
2. Move away from HPA to KEDA - it starts scaling sooner and removes big shocks later than when using HPA.
3. Create headroom with low-priority fake pods (those could be spun on schedule if the timing is repeatable).
4. Check if the team has some DaemonSets that take a long time to initialize the node to the Ready state.