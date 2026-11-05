---
title: Decision models
description: Ask a decision model, such as a self-hosted decider or TypeSafe's Jev, for a second opinion on whether a recommendation is safe to apply.
---

A **decision model** gives a second opinion on the question the workload page exists for: *is this recommendation
safe to apply now?* It reads the same evidence as the [verdict](../concepts/verdict-and-evidence.md), and answers
**apply**, **review** or **wait**, with a probability for each.

It is advice, nothing more:

- **The decision is yours.** The model never applies, reverts or holds anything, and auto-deploy does not read it.
- **The rules stay.** The rule-based verdict is shown as before, beside the model's answer, never replaced by it.
- **Nothing is asked until you click Evaluate.** The controller never calls the model; only the server does, when a
  user asks.

## What a decision model is

Decision models (the "System One" class) read a **state** and a set of **typed questions**, each with a fixed list of
options, and return a probability for every option in one pass, without generating text. The answer is always one of
the options, and the probabilities are meant to be calibrated: a model saying 80% should be right about 80% of the
time.

resourcerer speaks their HTTP interface, `POST /v1/systemone`, which two kinds of server offer:

| Provider | What it is | Data leaves the cluster |
|---|---|---|
| `decider` | [decider](https://github.com/Mapika/decider), open weights (Apache 2.0) from 0.8B to 35B parameters, served by `decider.serve`. Runs in your cluster. | No |
| `typesafe` | TypeSafe's managed Jev API, authenticated with a token. | Yes: the state below, never names |

Prefer a self-hosted decider: no credentials to manage, and nothing about your workloads leaves the cluster.

## On the workload page

With a model configured, the Impact tab shows two blocks side by side under the **Apply** button:

| Block | Answers from | When |
|---|---|---|
| **Rules** | History coverage, the replay of usage against the recommendation, and open reliability signals. Deterministic. See [Verdict and evidence](../concepts/verdict-and-evidence.md). | Always |
| **Decision model** | The model named in the `DecisionModel`, from the state below. | When there is something to apply, and you click **Evaluate** |

The model's block shows its answer ("Safe to apply · 89%"), a bar per option, when it was asked and the version that
answered. The answer is kept for that recommendation: reloading the page shows it again, and **Re-evaluate** asks the
model again. A regenerated recommendation (a new hash) starts without an answer.

When the two blocks disagree, read the reasons under the rules: they say what to look at. The model can weigh things
the rules do not, such as a memory trend or a daily peak, and it can also be wrong; see [Limits](#limits).

Without a `DecisionModel`, the page is as before, with a "Decision models coming soon" note next to the verdict.

## Set it up

Create one cluster-scoped `DecisionModel` named `default`; the server reads it on every evaluation, so a change takes
effect on the next click.

=== "Self-hosted decider"

    Run a decider server where the resourcerer server can reach it, then point the `DecisionModel` at it:

    ```yaml title="decision-model.yaml"
    apiVersion: resourcerer.io/v1alpha1
    kind: DecisionModel
    metadata:
      name: default
    spec:
      provider: decider
      endpoint: http://decider.decider.svc:8000/v1/systemone
      model: decider-4b
      timeoutSeconds: 30
    ```

    decider publishes the weights and the server, not a container image. A minimal image:

    ```dockerfile title="Dockerfile"
    FROM python:3.12-slim
    RUN pip install --no-cache-dir "decider-ai[serve]>=1.10.0"
    # The 4B weights are fetched from Hugging Face on first start; mount a volume on /root/.cache to keep them.
    ENV DECIDER_MODEL=Mapika/decider-4b DECIDER_IMAGES=0
    CMD ["uvicorn", "decider.serve:app", "--host", "0.0.0.0", "--port", "8000"]
    ```

    `DECIDER_IMAGES=0` turns off image questions, which resourcerer does not use and which need `torchvision`. The
    server picks CUDA, then Apple's MPS, then the CPU. decider-4b takes about 8.4 GB in bf16 on a GPU; on a CPU it runs
    in float32, about twice that, and more slowly, so a smaller model (`Mapika/decider-2b`) may suit a CPU-only node.
    Check the [model cards](https://huggingface.co/Mapika/decider-4b) for current sizes and versions.

    The decider server has no authentication: keep it in-cluster, behind a `NetworkPolicy` that admits only the
    resourcerer server.

=== "TypeSafe (Jev)"

    Store the API token in a Secret **in resourcerer's namespace**, ideally synced from a secrets manager (for example
    with External Secrets), never in Helm values or Git:

    ```bash
    kubectl -n resourcerer create secret generic typesafe --from-literal=token="$TYPESAFE_TOKEN"
    ```

    ```yaml title="decision-model.yaml"
    apiVersion: resourcerer.io/v1alpha1
    kind: DecisionModel
    metadata:
      name: default
    spec:
      provider: typesafe
      endpoint: https://<the endpoint TypeSafe gives you>/v1/systemone
      model: jev
      credentialsSecretRef:
        name: typesafe
        key: token
    ```

    The token is sent as `Authorization: Bearer <token>`. The server reads only that Secret, in its own namespace;
    see [Permissions](../reference/permissions.md#server). The state below is sent to TypeSafe on every evaluation.

Every field is in [Custom resources](../reference/custom-resources.md#decisionmodel). To turn the integration off,
delete the `DecisionModel`.

## What is sent

One request per evaluation: the state of the workload, built from its `Recommendation` status and the usage history the
recommendation was computed from, and one question. Each field is a short phrase with units, written by resourcerer;
nothing is configurable yet.

| Field | From | Says |
|---|---|---|
| `workload` | The kind, the replicas, `status.impact`, `status.hpa`, `status.pricing` | Kind, replicas now and on average, the HPA and its new target, the spot and on-demand mix |
| `window` | `status.evidence` | Whether the learning period is covered, or the history is preliminary |
| `intent` | `status.settings` | The CPU and memory goals, and the level that set them |
| `usage` | Prometheus, per optimized container: the busiest pod over the learning period | p50, p95, p99 and max; max over p95; variation; a daily pattern (from 2 days of history) and a trend (from 3 days) |
| `recommendation` | `status.containers` | Every request and limit that changes, from and to |
| `replay` | `status.evidence`, `status.impact` | How often usage went above the recommendation, its peaks, and the cost before and after |
| `fixes` | `status`, the usage | The problems applying fixes: memory raised after OOM kills, a CPU limit raised where it throttled, requests raised above p95 usage |
| `reliability` | `status.oomKills`, `status.reliability` | Every signal, the ones applying fixes tagged `(fixed by applying)` |
| `history` | `status.history`, `status.lastApplyResult` | The last three applies and reverts, how long ago, and whether by a person or by auto-deploy |

An actual request, for a Deployment with an `istio-proxy` sidecar. The sidecar is in `containers.skip`: its usage and
signals are left out, and the recommendation lists it as unchanged.

```json title="POST /v1/systemone"
{
  "model": "decider-4b",
  "state": {
    "workload": "Deployment, 2 replicas now, 2 on average; runs on 50% spot, 50% on-demand nodes",
    "window": "7d learning period, fully covered",
    "intent": "CPU goal Balanced, memory goal Balanced, set at cluster level",
    "usage": {
      "api cpu": "busiest pod p50 0.101, p95 0.104, p99 0.104, max 0.105; max/p95 1x; variation 4%; no daily pattern; no trend",
      "api memory": "busiest pod p50 63Mi, p95 63Mi, p99 63Mi, max 63Mi; max/p95 1x; variation 0%; no daily pattern; no trend"
    },
    "recommendation": "api: cpu request 0.5 -> 0.115 (-77%), memory request 1Gi -> 73Mi (-93%); istio-proxy not changed",
    "replay": "over the history, the busiest pod used more than the recommended requests 0% of the time for CPU and 0% for memory, and peaked at 91% of the CPU and 86% of the memory request; $19.07 -> $6.13 a month, saves $12.94/month",
    "fixes": "nothing: no reliability problem the recommendation addresses",
    "reliability": "no OOM kills; api memory peaked at 3% of its limit; HTTP 5xx 0.2% (ok)",
    "history": "never applied"
  },
  "questions": {
    "verdict": {
      "type": "choice",
      "instructions": "Is it safe to apply the recommendation to this workload now? Weigh how the recommended requests held up against the usage history, how much of the learning period that history covers, the usage shape, what applying fixes and the reliability signals it does not fix.",
      "criteria": {
        "apply": "safe to apply now",
        "review": "a person should review it before applying",
        "wait": "wait for more history or for the reliability signals to settle"
      }
    }
  }
}
```

About 600 tokens. The answer is accepted only when its choice is one of the three options and every option has a
probability; anything else is shown as an error.

**Never sent:** the names of the workload and its namespace, who applied it, labels, annotations, raw usage series,
hashes, prices per instance type. **Sent:** container names and the names of declared reliability signals, as they
appear above. With a managed provider, keep that in mind when naming signals.

!!! tip "Why the signals applying fixes are tagged"
    Listed apart from the signals, a model counts an OOM kill against applying even when the recommendation raises
    the memory that was killed. Tagged inline (`12 OOM kills (fixed by applying)`), it reads them as reasons to apply,
    as the rules do.

## When it cannot answer

The model's block says why, and the rules are unaffected:

| Message | Cause |
|---|---|
| `ask the decision model: … connection refused` / `context deadline exceeded` | The endpoint is unreachable from the server, or slower than `timeoutSeconds`. |
| `decision model answered 401 …` | A wrong or missing token, for `typesafe`. |
| `read the decision model's credentials: …` | The Secret is missing, in another namespace, or the server may not read it. |
| `answer "…" to "verdict" is not an option` | The server answered something other than `apply`, `review` or `wait`: not a decision model, or another interface. |
| `read CPU usage: …` | Prometheus failed while the state was built. |

A failed evaluation is not kept: **Evaluate** asks again.

## Limits

- **Calibration is the model's, not ours.** A probability is only as good as the model's training on this kind of
  question. Before trusting the numbers, compare its answers with what happened after the applies you made.
- **Small models miss things.** In our tests decider-4b weighed history coverage, OOM kills and error rates well,
  but read a memory trend of 8% a day and CPU above the recommendation 30% of the time as safe. A larger model may do
  better; the rules flag the second case.
- **Answers are kept in the server's memory** for an hour per recommendation; restarting the server forgets them.
- **One model per cluster**, the `DecisionModel` named `default`.

## Try it on kind

[Try it on kind](try-on-kind.md) runs resourcerer locally; `hack/decision-model.yaml` points it at a decider running
on your machine, with the commands to start one. On Apple Silicon, decider-4b answers in one to three seconds.

## What comes next

Design issue [#26](https://github.com/gjbravi/resourcerer/issues/26): auto-deploy that can require both the rules and
the model above a confidence you choose (with a floor the cluster admin sets, and a pause when the model's version
changes), recording outcomes to measure calibration, and more questions once the verdict is calibrated.
