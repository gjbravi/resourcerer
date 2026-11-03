---
title: resourcerer
description: Self-hosted Kubernetes rightsizing. CPU and memory recommendations from your own Prometheus, with the savings, the risk and the reasons behind every number.
hide:
  - navigation
  - toc
  - footer
---

<div class="rz-hero">
<div class="rz-wrap rz-hero__grid">
<div>
<p class="rz-badge">Kubernetes rightsizing · Open source · Beta</p>
<h1>Right-size every workload, <span>and know why it's safe.</span></h1>
<p class="rz-lead">
resourcerer reads the usage history already in your Prometheus and recommends CPU and memory for every workload.
Each recommendation comes with <b>what it saves</b>, <b>what it risks</b> and <b>why the numbers can be trusted</b>,
then applies it on a click, from Git or automatically, inside the windows you allow.
</p>
<div class="rz-actions">
<a class="rz-btn rz-btn--primary" href="getting-started/quickstart/">Quickstart →</a>
<a class="rz-btn rz-btn--outline" href="concepts/how-it-works/">How it works</a>
<a class="rz-btn rz-btn--outline" href="https://github.com/gjbravi/resourcerer">GitHub</a>
</div>
<ul class="rz-points">
<li>One Helm chart</li>
<li>Runs in your cluster</li>
<li>No SaaS, no database</li>
</ul>
</div>

<div class="rz-verdict" aria-label="Example of a workload page">
<div class="rz-verdict__head">
<div><span class="rz-verdict__name">checkout</span> <span class="rz-verdict__ns">payments · Deployment</span></div>
<span class="rz-pill">Ready</span>
</div>
<div class="rz-costs">
<div class="rz-cost"><small>Now</small><b>$96.30</b><small>/mo</small></div>
<div class="rz-cost rz-cost--after"><small>Recommended</small><b>$51.90</b><small>/mo · saves 46%</small></div>
</div>
<div class="rz-answer"><div><b>Safe to apply</b><br><span>saves $44.40/mo</span></div><span class="rz-apply">Apply</span></div>
<ul class="rz-reasons">
<li><span class="rz-ok">✓</span> 7d of history</li>
<li><span class="rz-ok">✓</span> memory peak at 81% of the recommendation</li>
<li><span class="rz-ok">✓</span> CPU above the recommendation 3% of the time</li>
<li><span class="rz-ok">✓</span> no open reliability signals</li>
</ul>
<div class="rz-basis">cpu: p95 + 10% over 7d · memory: max + 15% over 7d</div>
</div>
</div>
</div>

<div class="rz-section">
<div class="rz-wrap">
<p class="rz-eyebrow">Why resourcerer</p>
<h2>Recommendations you can check, and apply without holding your breath.</h2>
<div class="rz-cards">
<div class="rz-card"><span class="rz-icon">⎈</span><h3>Runs in your cluster</h3><p>Three small Deployments and a <code>Recommendation</code> custom resource. Usage history comes from the Prometheus you already run.</p></div>
<div class="rz-card"><span class="rz-icon">∑</span><h3>Explainable numbers</h3><p>Every value is a percentile of usage plus a margin, such as <code>p95 + 10% over 7d</code>. No model to trust blindly.</p></div>
<div class="rz-card"><span class="rz-icon">✓</span><h3>A verdict, not just data</h3><p>"Safe to apply", "Review before applying" or "Apply to fix reliability", from history, a replay of usage and reliability signals.</p></div>
<div class="rz-card"><span class="rz-icon">$</span><h3>Costs that match your nodes</h3><p>Configured rates, or calibrated to Karpenter's prices and priced at the nodes each workload ran on, spot or on-demand.</p></div>
<div class="rz-card"><span class="rz-icon">⏱</span><h3>Applies on your terms</h3><p>Apply and revert from the UI, or auto-deploy within apply policies: windows, freezes and rate limits.</p></div>
<div class="rz-card"><span class="rz-icon">⑂</span><h3>GitOps friendly</h3><p>Settings are annotations, policies are custom resources, applies use server-side apply with their own field manager.</p></div>
</div>
</div>
</div>

<div class="rz-section rz-section--tint">
<div class="rz-wrap">
<p class="rz-eyebrow">How it works</p>
<h2>From usage history to an applied change, in four steps.</h2>
<div class="rz-steps">
<div class="rz-step"><h3>Read</h3><p>The controller reads each container's usage from the busiest pod over the learning period.</p></div>
<div class="rz-step"><h3>Recommend</h3><p>A percentile plus a margin per goal, then limits, minimums, OOM and throttling rules, and the HPA target.</p></div>
<div class="rz-step"><h3>Review</h3><p>The UI shows the cost before and after, the evidence and the verdict under the Apply button.</p></div>
<div class="rz-step"><h3>Apply</h3><p>A server-side apply and a normal rollout, recorded in a history you can revert.</p></div>
</div>
</div>
</div>

<div class="rz-section rz-cta">
<div class="rz-wrap">
<p class="rz-eyebrow">Get started</p>
<h2>Install it in a minute. Nothing changes until you click Apply.</h2>

```bash
helm install resourcerer oci://ghcr.io/gjbravi/charts/resourcerer --version <version> \
  -n resourcerer --create-namespace --set prometheus.url=http://prometheus-server.monitoring.svc
```

<div class="rz-actions">
<a class="rz-btn rz-btn--primary" href="getting-started/quickstart/">Read the quickstart →</a>
<a class="rz-btn rz-btn--outline" href="getting-started/introduction/">Introduction</a>
</div>
</div>
</div>
