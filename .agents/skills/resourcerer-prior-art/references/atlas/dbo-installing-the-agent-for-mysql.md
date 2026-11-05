---
updatedAt: 2026-10-16T11:42:10.000Z
---

# Installing the agent for MySQL

[Performance Advisor](dbo-performance-advisor.md) collects its data through a lightweight agent that you deploy as a Kubernetes workload next to your MySQL database. The agent observes your query workload and metadata through read-only access and feeds the recommendations shown in the console. For PostgreSQL, see [Installing the agent for PostgreSQL](dbo-installing-the-agent-for-postgresql.md).

## Prerequisites

* MySQL 5.7 or later (required for performance schema and query digest support).
* A database user with read-only access to the performance schema and system tables.
* Helm and kubectl configured against the cluster where the agent will run.

## Step 1: Create a database user for the agent

```sql
CREATE USER 'performance_advisor_agent'@'%' IDENTIFIED BY 'your_secure_password_here';
```

Use a strong, unique password. The `'%'` host wildcard allows connections from any host; restrict it to the agent's IP or subnet in production if needed.

## Step 2: Grant required permissions

As a MySQL superuser, grant SELECT on the performance schema, which the agent uses to collect query statistics:

```sql
GRANT SELECT ON performance_schema.* TO 'performance_advisor_agent'@'%';
```

Then grant REFERENCES on each database you want analyzed. REFERENCES is only used for information\_schema lookups; it does not allow reading, modifying, or deleting table data:

```sql
GRANT REFERENCES ON your_database.* TO 'performance_advisor_agent'@'%';
```

Repeat this grant for each additional database to analyze.

## Step 3: Enable Performance Advisor in the console

MySQL instances use the same flow as PostgreSQL: on the organization **Overview**, click **Enable database optimizer** on the instance's row (the **Not onboarded** tab lists discovered MySQL instances with the same action), then tick **Performance advisor** under **Optimization features** in the **Enable Database optimization** dialog. Under **Run script below**, the **Helm** tab gives you the installation instructions specific to your instance, with your API key and organization ID already embedded in a downloadable `values.yaml` (a `db-agent:` block with `enabled: true`, `apiKey`, `apiURL`, `nameOverride`, and `organizationID`); add your MySQL connection details and credentials, then run the generated `helm upgrade --install --create-namespace -n atlas-db-optimizer ... atlas-helm/atlas-dbo -f values.yaml` command. The **Script** tab offers a copy-paste alternative that asks for a read-only database user's **Username** and **Password** instead. Finish with **I ran the script**. Your database must be onboarded to DBO first (see [Getting started](dbo-getting-started.md)).

<Image
  src="https://files.readme.io/f79c04367f97dab9f3f5e7b38bb85f8a84754d80e3a1ab2d1b5abfdd27d2e849-installing-the-agent-for-mysql-01.png"
  alt="Enable Database optimization dialog, Helm tab, with the generated values.yaml and the Install command"
  caption="The Enable Database optimization dialog's Helm tab, with the generated values.yaml and the Install command."
/>

Once the agent is installed and healthy, the status row on every instance page shows **ADVISOR AGENT** **Active**, and the instance **Settings** page shows a **Performance advisor** card with an **Active** badge.

## Step 4: Install the agent via Helm

For production deployments, store credentials in Kubernetes secrets rather than passing them directly to the Helm chart:

```shell
kubectl create secret generic atlas-db-agent-api-key \
  --namespace atlas-db-optimizer \
  --from-literal=API_KEY='your-api-key-here'

kubectl create secret generic atlas-db-agent-db-credentials \
  --namespace atlas-db-optimizer \
  --from-literal=DATABASE_USERNAME='performance_advisor_agent' \
  --from-literal=DATABASE_PASSWORD='your_secure_password_here'
```

Reference them in `values.yaml` with `apiKeySecretRef` and `database.credentialsSecretRef` (host, port 3306, database name, and protocol `mysql`). Then add the Atlas Helm repository and install the agent:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
helm upgrade --install atlas-db-agent atlas-helm/atlas-dbo \
  --namespace atlas-db-optimizer --create-namespace \
  --values values.yaml
```

Verify that the agent is running:

```shell
kubectl get pods -n atlas-db-optimizer
```

Once the pods are healthy, recommendations appear in the console within a few minutes.

## Managing the agent

To remove the agent, open the instance **Settings** page and click **Uninstall** on the **Performance advisor** card.

## Next steps

<Cards columns={2}>
  <Card title="Performance Advisor" href="https://docs.atlas.example/docs/dbo-performance-advisor" icon="fa-chart-line">
    The recommendations your agent now produces.
  </Card>

  <Card title="Installing the agent for PostgreSQL" href="https://docs.atlas.example/docs/dbo-installing-the-agent-for-postgresql" icon="fa-terminal">
    The same flow for PostgreSQL databases.
  </Card>
</Cards>