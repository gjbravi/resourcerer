---
title: Roadmap
description: What comes next, as design issues open for comments.
---

Larger features start as design issues on GitHub, open for comments before they are built. Comment on the ones that
matter to you: a concrete use case is the most useful input.

| Area | Issue | What it brings |
|---|---|---|
| Applying | [#29 ApplyPolicy](https://github.com/gjbravi/resourcerer/issues/29) | Windows, freezes and rate limits are in. Next: a disruption budget, freezes declared by a query, verification with automatic revert, and change limits per apply. |
| Access | [#33 Authentication](https://github.com/gjbravi/resourcerer/issues/33) | OIDC login, a trusted-header mode for proxies, per-user authorization, and who applied in the history. |
| Settings | [#30 OptimizationPolicy](https://github.com/gjbravi/resourcerer/issues/30) | Settings by selector instead of annotations, maximums, and per-container rules. |
| Costs | [#27 CostProfile](https://github.com/gjbravi/resourcerer/issues/27) | Savings Plans, Reserved Instances and discounts in cost estimates. |
| Reliability | [#28 ReliabilitySignal](https://github.com/gjbravi/resourcerer/issues/28) | Reliability signals as custom resources, so teams declare their own. |
| Decisions | [#26 Decision models](https://github.com/gjbravi/resourcerer/issues/26) | Pluggable models behind the apply verdict. |
| Scaling | [#39 Horizontal scaling](https://github.com/gjbravi/resourcerer/issues/39) | Full HPA support and KEDA ScaledObjects. |
| Upgrades | [#44 CRD delivery](https://github.com/gjbravi/resourcerer/issues/44) | How a CRD change reaches an installation that already exists. |

Released changes are in the [changelog](https://github.com/gjbravi/resourcerer/blob/main/CHANGELOG.md).
