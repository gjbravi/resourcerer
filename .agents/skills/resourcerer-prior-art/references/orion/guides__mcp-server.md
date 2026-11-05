# Connect AI agents to Orion with MCP

> Hosted and local MCP server setup for AI agents


Orion runs a hosted [Model Context Protocol (MCP)][mcp] server, so AI agents such as Claude Code can answer questions about your clusters, workloads, and recommendations. The hosted server is read-only by default, and you connect to it with your Orion account.

The Orion CLI also includes a [local MCP server](#use-the-local-mcp-server) for MCP clients that can't connect to the hosted server.

## Hosted server URL

```text
https://api.orion.example/mcp
```

The server uses streamable HTTP transport. Your AI agent signs you in through your browser the first time it connects; there are no API keys to create or store.

The hosted server supports Claude Code, Claude Desktop and claude.ai, ChatGPT, Codex, and VS Code. For other MCP clients, use the [local MCP server](#use-the-local-mcp-server).

## Connect your AI agent

### Claude Code

1. Add the server:
   ```shell
   claude mcp add --transport http orion https://api.orion.example/mcp
   ```
2. In Claude Code, run `/mcp`, select **orion**, and choose **Authenticate**.
3. Sign in to Orion in your browser.

### Claude Desktop and claude.ai

1. Go to **Settings** > **Connectors**, click **Add**, and choose **Add custom connector**.
2. Enter `Orion` as the name and `https://api.orion.example/mcp` as the URL. Leave the OAuth client ID and secret empty.
3. Click **Add**, then click **Connect**.
4. Sign in to Orion in your browser.

On Claude Team and Enterprise plans, an organization owner adds the connector first. Each member then clicks **Connect** and signs in with their own Orion account.

### ChatGPT

Custom MCP servers require a paid ChatGPT plan. On Business and Enterprise plans, a workspace admin may need to allow them.

1. Go to **Settings** > **Plugins** and click **Browse directory**.
2. Click **Add** and choose **Add custom MCP server**.
3. Enter `Orion` as the name and `https://api.orion.example/mcp` as the connection URL.
4. Select **I understand and want to continue**, then click **Create as a plugin**.
5. Click **Continue to Orion**.
6. Sign in to Orion in your browser.
7. In a chat, select **Orion** from the **+** menu.

### Codex

1. Add the server:
   ```shell
   codex mcp add orion --url https://api.orion.example/mcp
   ```
2. Sign in to Orion in your browser.

### VS Code

1. Open the Command Palette, run **MCP: Open User Configuration**, and add:
   ```json
   {
     "servers": {
       "orion": {
         "type": "http",
         "url": "https://api.orion.example/mcp"
       }
     }
   }
   ```
2. Run **MCP: List Servers**, select **orion**, and choose **Start Server**.
3. Allow VS Code to authenticate and open the website.
4. Sign in to Orion in your browser, then confirm opening VS Code.

### Ask your AI agent to set it up

If your AI agent can edit its own configuration, give it the details directly:

```text
Add the Orion MCP server to your configuration. It is a remote server
using streamable HTTP at https://api.orion.example/mcp, named "orion",
and it signs in with OAuth in the browser. Then connect and run its whoami
tool to confirm which Orion account you're using.
```

## Sign in and check your account

When you sign in, select your Orion organization if you belong to more than one, then authorize your AI agent's access.

Your AI agent sees the same organization and data as your Orion account.

To confirm which account your AI agent is using, ask:

```text
Which Orion account are you signed in as?
```

The response includes your name and the organizations and roles you have access to. Every Orion [user role][user-roles] can use the read tools.

## Available tools

### Read tools

| Tool | What it does |
|------|--------------|
| **Who Am I?** (`whoami`) | Shows the signed-in user and their organizations and roles |
| **Get estate overview** (`get_estate_overview`) | Returns estate-wide totals and statistics, such as workload counts, optimization scores, and cost impact |
| **List clusters** (`list_clusters`) | Lists clusters |
| **List namespaces** (`list_namespaces`) | Lists namespaces |
| **List workloads** (`list_workloads`) | Lists and filters workloads |
| **Get workload** (`get_workload`) | Returns details for one workload, including its configuration and latest recommendation |
| **List recommendations** (`list_recommendations`) | Lists the latest recommendation for each workload |

The list tools can also return optimization scores and cost, CPU, and memory impact.

### Write tools

These write tools are disabled by default and require the Operator, Manager, or Administrator role:

| Tool | What it does |
|------|--------------|
| **Apply recommendations** (`apply_recommendations`) | Marks generated recommendations as ready, so the Applier patches the matching workloads |
| **Create recommendations** (`create_recommendations`) | Creates an on-demand recommendation for each matching workload |

To enable the write tools for your organization:

1. Open the Orion configuration in your editor:
   ```shell
   orion edit configuration --managed=false
   ```
2. Remove `apply_recommendations` and `create_recommendations` from the `mcpSettings.disabledTools` list, then save and close the file.

To disable the write tools, add both names back to the list.

## Example prompts

- `How many workloads does Orion manage across my clusters?`
- `What are the top 3 most overprovisioned workloads?`
- `Which namespaces in the prod cluster have the lowest optimization scores?`
- `Show me the latest recommendation for the checkout workload.`

Example response:

```text
The top 3 by overprovisioned cost are:

  ┌───┬──────────────┬───────────┬─────────────────┬───────┬─────────────┐
  │ # │   Workload   │  Cluster  │ Overprovisioned │ Cost  │ Auto-deploy │
  ├───┼──────────────┼───────────┼─────────────────┼───────┼─────────────┤
  │ 1 │ <workload-1> │ <cluster> │ 0.185 CPU       │ $4.27 │ Off         │
  │ 2 │ <workload-2> │ <cluster> │ 0.051 CPU       │ $1.18 │ Off         │
  │ 3 │ <workload-3> │ <cluster> │ ~122 MiB memory │ $0.37 │ On          │
  └───┴──────────────┴───────────┴─────────────────┴───────┴─────────────┘

  <workload-1> and <workload-2> have auto-deploy off, so their
  recommendations aren't applied. Turning on auto-deploy would recover
  most of the overprovisioning.
```

## Use the local MCP server

The Orion CLI can run an MCP server on your machine over stdio. It works with any MCP client that supports stdio servers, including clients that can't connect to the hosted server.

Configure either the hosted server or the local server, not both.

1. [Install the Orion CLI][install-cli] and sign in:
   ```shell
   orion login
   ```
2. Add the server to your AI agent. For Claude Code:
   ```shell
   claude mcp add orion-local -- orion mcp
   ```
   For clients that use an `mcpServers` configuration file:
   ```json
   {
     "mcpServers": {
       "orion-local": {
         "command": "orion",
         "args": ["mcp"]
       }
     }
   }
   ```
   If the `orion` binary isn't on your AI agent's `PATH`, use its absolute path as the `command`.

The local server uses the account from `orion login`. Run `orion login` again to switch accounts.

## Troubleshooting

### Answers come from the wrong organization

If you configured both the hosted and local servers, they can be signed in to different Orion accounts. Remove the server you don't use.

If you belong to more than one Orion organization, you might have selected a different one when you signed in. [Sign out](#sign-out-or-switch-accounts), then sign in again and select the organization you want.

### Sign out or switch accounts

Signing in reuses any Orion session in your browser. To connect a different account, sign out of Orion in your browser first, then:

- **Claude Code:** Run `/mcp`, select **orion**, and choose **Clear authentication**. To sign in again, run `/mcp`, select **orion**, and choose **Authenticate**.
- **Claude Desktop and claude.ai:** In **Settings** > **Connectors**, click **Disconnect** on the Orion connector, then click **Connect**.
- **ChatGPT:** In **Settings** > **Plugins**, click **Uninstall** on the Orion plugin. Then click **Browse directory**, open **Orion** under **Personal**, click **Install plugin**, and click **Continue to Orion**.
- **Codex:** Run `codex mcp logout orion`, then `codex mcp login orion`.
- **VS Code:** In the Command Palette, run **Authentication: Remove Dynamic Authentication Providers** and select **Orion MCP**. Then run **MCP: List Servers**, select **orion**, and choose **Disable (Session)**. Select **orion** again, choose **Enable (Session)**, then **Start Server**, and authenticate when prompted.

## Related topics

- [Orion CLI][install-cli]
- [Claude Code MCP documentation][claude-mcp]
- [Model Context Protocol specification][mcp]

[claude-mcp]: https://code.claude.com/docs/en/mcp
[install-cli]: /docs/orion-cli/#install
[mcp]: https://modelcontextprotocol.io/
[user-roles]: /docs/reference/user-roles/
