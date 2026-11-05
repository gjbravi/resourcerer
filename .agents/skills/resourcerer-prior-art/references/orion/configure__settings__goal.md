# Optimization goal

> Savings, Balanced, or Reliability tradeoff for recommendations


Orion always recommends values that minimally satisfy both savings and reliability, but you can set goals to shift the balance in favor of one or the other based on your risk tolerance and optimization goals.

The optimization goal applies to the entire workload.

## Optimization goal {#opt-goal}

Used in determining recommended values based on your goals and risk tolerance.

| Annotation                                    | Default value |
| --------------------------------------------- | ------------- |
| `live.orion.example/cpu.optimization-goal`    | `"Balanced"`  |
| `live.orion.example/memory.optimization-goal` | `"Balanced"`  |

### Description {#opt-goal-description}

The optimization goal for a resource indicates which of two competing criteria to maximize when computing recommended request values: cost savings or workload reliability. Orion will always produce values that minimally satisfy both criteria, but the goal setting can be used to shift the balance in favor of one or the other.

### Valid values {#opt-goal-valid-values}

- `Balanced` (default): Provides a balanced approach to achieving cost savings and workload reliability.
- `Savings`: Provides recommendations that are closer to actual CPU or memory usage. Consider this option when you want to maximize your resource savings.
- `Reliability`: Minimizes the risk of usage bursting over request values or hitting CPU or memory limits. Consider this option for business-critical applications.

### Key points {#opt-goal-key-points}

- CPU and memory optimization goals can differ. For example, if your organization can tolerate throttling when containers exceed CPU limits but cannot tolerate restarts when containers exceed memory limits, you can set a **Savings** goal for CPU and a **Reliability** goal for memory.
