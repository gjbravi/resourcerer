---
updatedAt: 2026-10-16T11:42:10.000Z
---

# Installing the agent for PostgreSQL

[Performance Advisor](dbo-performance-advisor.md) collects its data through a lightweight agent that you deploy as a Kubernetes workload next to your PostgreSQL database. The agent observes your query workload and metadata through read-only access and feeds the recommendations shown in the console. For MySQL, see [Installing the agent for MySQL](dbo-installing-the-agent-for-mysql.md).

## Prerequisites

* PostgreSQL with the `pg_stat_statements` extension enabled.
* A database user with read-only access to system catalogs and statistics.
* Helm and kubectl configured against the cluster where the agent will run.

## Step 1: Enable the pg\_stat\_statements extension

Check whether the extension is already enabled:

```sql
SELECT * FROM pg_extension WHERE extname = 'pg_stat_statements';
```

If it is not, create it as a superuser:

```sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
```

Enabling it may require adding `pg_stat_statements` to `shared_preload_libraries` in your PostgreSQL configuration and restarting the database. On AWS RDS, set `shared_preload_libraries` via a parameter group and reboot the instance.

## Step 2: Create a database user for the agent

```sql
CREATE USER performance_advisor_agent WITH PASSWORD 'your_secure_password_here';
GRANT pg_read_all_stats TO performance_advisor_agent;
```

The `pg_read_all_stats` role gives the agent access to the query statistics views without any write privileges.

## Step 3: Enable Performance Advisor in the console

1. On the organization **Overview**, click **Enable database optimizer** on the instance's row (**Not onboarded** tab).

2. In the **Enable Database optimization** dialog, confirm **Performance advisor** is ticked under **Optimization features**. It is pre-selected and marked **Recommended**.

3. Under **Run script below**, choose how to install:
   * **Helm**: download the provided `values.yaml`, add your database connection details and credentials, and run the generated `helm upgrade --install` command.
   * **Script**: copy the `curl` command and run it as-is. It asks for a read-only database user's **Username** and **Password**. Credentials are shown once and not stored.

4. Finish with **I ran the script**.

Your database must be onboarded to DBO first (see [Getting started](dbo-getting-started.md)).

### The Helm `values.yaml`

The downloadable `values.yaml` already contains your API key and organization ID in a `db-agent:` block:

```yaml
db-agent:
  enabled: true
  apiKey: <your-api-key>
  apiURL: https://api.atlas.example
  nameOverride: ""
  organizationID: <your-organization-id>
```

<Image
  src="https://files.readme.io/2b85152f651bc42e00472103c912fb07937a8e7e53c30aff9b814b6a7ed07058-postgresql-01.png"
  alt="Enable Database optimization dialog, Helm tab, with the Configure database permissions, Access repository, Customize values, and Install steps"
  caption="The Enable Database optimization dialog's Helm tab, with the Configure database permissions, Access repository, Customize values, and Install steps."
/>

The **Script** tab is the copy-paste alternative, asking only for the read-only agent user's **Username** and **Password**:

<Image
  src="https://files.readme.io/fa36ac7736bd2517a0e74342b2542e7ba032c78015236686e31326c718314ace-postgresql-02.png"
  alt="Enable Database optimization dialog, Script tab, with the Configure user permission and Install steps"
  caption="The Enable Database optimization dialog's Script tab, with the Configure user permission and Install steps."
/>

Once the agent is installed and healthy, the status row on every instance page shows **ADVISOR AGENT** **Active**, and the instance **Settings** page shows a **Performance advisor** card with an **Active** badge:

<Image
  src="https://files.readme.io/e6bddf292c0d3e617ff6283a288c5cc877a032e062934d40f4a686826dce5c09-installing-the-agent-for-postgresql-03.png"
  alt="Instance Settings page with the Performance advisor card showing an Active badge"
  caption="The instance Settings page's Performance advisor card, showing an Active badge once the agent is healthy."
/>

## Step 4: Install the agent via Helm

Store credentials in Kubernetes secrets rather than passing them directly to the Helm chart:

```shell
kubectl create secret generic atlas-db-agent-api-key \
  --namespace atlas-db-optimizer \
  --from-literal=API_KEY='your-api-key-here'

kubectl create secret generic atlas-db-agent-db-credentials \
  --namespace atlas-db-optimizer \
  --from-literal=DATABASE_USERNAME='performance_advisor_agent' \
  --from-literal=DATABASE_PASSWORD='your_secure_password_here'
```

Then add the Atlas Helm repository and install the agent:

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

## Optional connection methods

* **Cloud SQL Proxy (GCP)**: connect through the Cloud SQL instance's Connection Name instead of a direct IP.
* **AWS RDS IAM authentication**: passwordless authentication via an IAM role bound to the db-agent service account (IRSA), instead of a static username and password. Create the database user with the `rds_iam` role (`CREATE USER performance_advisor_agent; GRANT rds_iam TO performance_advisor_agent;`), grant the same `pg_read_all_stats` permissions as in step 2, enable IAM database authentication on the RDS instance, and give the IAM role the `rds-db:connect` permission.

## Next steps

<Cards columns={2}>
  <Card title="Performance Advisor" href="https://docs.atlas.example/docs/dbo-performance-advisor" icon="fa-chart-line">
    The recommendations your agent now produces.
  </Card>

  <Card title="Installing the agent for MySQL" href="https://docs.atlas.example/docs/dbo-installing-the-agent-for-mysql" icon="fa-terminal">
    The same flow for MySQL databases.
  </Card>
</Cards>