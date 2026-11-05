---
title: Roadmap
description: What comes next, as design issues open for comments.
---

Larger features start as design issues on GitHub, open for comments before they are built. Comment on the ones that
matter to you: a concrete use case is the most useful input.

| Area | Issue | What it brings |
|---|---|---|
| Applying | [#29 ApplyPolicy](https://github.com/gjbravi/resourcerer/issues/29) | Windows, freezes and rate limits are in. Next: a disruption budget, freezes declared by a query, verification with automatic revert, and change limits per apply. |
| Access | [#33 Authentication](https://github.com/gjbravi/resourcerer/issues/33) | OIDC login is in, with the roles taken from the token's groups. Next: a trusted-header mode for proxies, per-user authorization, and an override for advisory apply windows. |
| Settings | [#30 OptimizationPolicy](https://github.com/gjbravi/resourcerer/issues/30) | Settings by selector instead of annotations, maximums, and per-container rules. |
| Costs | [#27 CostProfile](https://github.com/gjbravi/resourcerer/issues/27) | Savings Plans, Reserved Instances and discounts in cost estimates. |
| Reliability | [#28 ReliabilitySignal](https://github.com/gjbravi/resourcerer/issues/28) | Reliability signals as custom resources, so teams declare their own. |
| Decisions | [#26 Decision models](https://github.com/gjbravi/resourcerer/issues/26) | A second opinion on request is in. Next: auto-deploy that can require the rules and the model above a confidence, and measuring the model's calibration on past applies. |
| Scaling | [#39 Horizontal scaling](https://github.com/gjbravi/resourcerer/issues/39) | Full HPA support and KEDA ScaledObjects. |

Released changes are in the [changelog](https://github.com/gjbravi/resourcerer/blob/main/CHANGELOG.md).
