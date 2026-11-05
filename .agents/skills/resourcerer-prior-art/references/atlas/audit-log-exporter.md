---
updatedAt: 2026-05-12T13:39:45.000Z
---

# Audit log exporter

**Atlas's Audit Log Exporter** addresses users' need to export Audit Logs to an external logging system. For example, they may want to do that when aggregating Atlas logs with log entries from other systems.

It is an essential tool for users who need to maintain long-term access to their audit log data. It ensures continuous access to historical log data beyond the 90-day console retention period, allowing users to meet specific compliance requirements or data retention needs.

This Audit Logs Exporter provides such capability by fetching the logs through Atlas's public API, processing them as necessary, and delivering them to the desired external logs management system.

## The architecture of the solution

Audit Logs Exporter is based on the [OpenTelemetry](https://opentelemetry.io/) framework. We chose this technology because it is versatile in exporting, processing, and delivering instrumentation data, including logs, metrics, and traces to arbitrary systems.

<Image align="center" src="https://files.readme.io/d48c8bd-Screenshot_2023-08-29_at_10.29.14.png" />

Moreover, OpenTelemetry offers a range of components, such as receivers, processors, and exporters, that you can combine into pipelines to handle telemetry data.

In the case of Atlas Audit Logs Exporter, the following components are particularly relevant:

* **Receiver** enables collecting data from the sources system (Atlas backend).
* **Exporter** submits the logs to an external logging system, i.e., the user's tool.

Using OpenTelemetry, Atlas's Audit Logs Exporter provides a scalable and extendable solution for exporting Audit Logs. Due to the broad support of exporters, contributed by the [community](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter), we could easily integrate it with any given logs management system.

It is vital to highlight that Audit Logs Exporter is run/operated by users on their infrastructure, while Atlas provides required support for issues like setting it up, tuning, and maintenance.

## How to implement it

We released our Audit Logs Exporter as an Open Source component; you can find it on GitHub. An extensive technical description accompanies the repository's landing page.

Built as a custom OpenTelemetry Collector, Audit Logs Exporter enables the user to choose a specific destination system when creating his instance. A detailed description of this process is also present in the repository.

You can host Audit Logs Exporter as a standalone app or on Kubernetes. If you wish to host it on K8S, [use these Helm charts](https://github.com/atlas/helm-charts/tree/main/charts/atlas-audit-logs-receiver) with the example deployment described [in this thread](https://github.com/atlas/audit-logs-receiver#usage).

## How to get help

In case of any issues or questions, or if you want to contribute to the further development of Atlas's Audit Log Exporter, feel welcome to use the following forms of contact:

* **For improvements**: [raise a PR on GitHub](https://github.com/atlas/audit-logs-receiver/pulls).
* **For issues**: [please raise an issue on GitHub](https://github.com/atlas/audit-logs-receiver/issues) or contact our Customer Support Team, for example, using our [Slack channel](https://atlas-community.slack.com/).