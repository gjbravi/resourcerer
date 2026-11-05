---
updatedAt: 2026-10-28T09:26:36.000Z
---

# Connect AI tools to Atlas MCP

Connect your AI tools to Atlas MCP to ask questions about your clusters, costs, efficiency, and health.

Use Atlas MCP to ask questions about your clusters, costs, efficiency, and health from your preferred AI tool.

The connection provides **read-only access** to your Atlas data. It cannot modify your clusters or configuration.

## Before you start

You'll need:

* A Atlas account.
* An AI tool that supports remote MCP connections over HTTP and browser-based OAuth authentication.

Setup also uses your Atlas region. If you're not sure which one your organization uses, contact your Atlas Account Manager.

| Region | MCP server URL                  | `--api-region` value |
| ------ | ------------------------------- | -------------------- |
| US     | `https://mcp.atlas.example/mcp`       | `us`                 |
| EU     | `https://mcp.eu.atlas.example/mcp`    | `eu`                 |
| India  | `https://mcp.india.atlas.example/mcp` | `india`              |

## Which AI tools can I use?

Any MCP-compatible client can connect to Atlas, for example:

* Coding assistants and terminal tools, such as Claude Code, [Kimchi](https://kimchi.dev), Cursor, or Codex CLI.
* Desktop and web chat apps, such as Claude Desktop or ChatGPT (where custom connectors are available).
* Your own agents, built with frameworks or SDKs that support remote MCP servers.

If your tool lets you add a remote MCP server by URL, it should work with Atlas.

## Option 1: Guided setup with `atlasctl`

If you already use [atlasctl](connect-with-atlasctl.md), the Atlas CLI, this is the easiest way to connect a supported AI client.

1. Update `atlasctl` to the latest version.

2. Run the setup command in your terminal:

```shell
atlasctl mcp add
```

3. Follow the prompts. The command detects supported AI clients and guides you through configuration.

4. Sign in when prompted to complete the browser login. When the prompt appears depends on your tool: some authenticate on first use, and others ask you to authenticate the connection from the tool's MCP settings.

> 📘 Claude Desktop setup is manual
>
> `atlasctl mcp add` detects Claude Desktop but does not configure it automatically. It links to Claude's guide for [adding a custom connector](https://claude.com/docs/connectors/custom/remote-mcp) instead: follow that guide and use the MCP server URL for your region from the table above. On some workspaces, adding a custom connector requires an organization owner, who can add it for your team.

To select a region, use the `--api-region` value for your region from the table above.

## Option 2: Manual setup

Use this option if you don't have `atlasctl` or prefer to configure the connection yourself.

1. Open your AI tool's settings for adding an MCP server or connector.

2. Give the connection a recognizable name, such as **Atlas**.

3. Choose **HTTP** as the transport type, if your tool asks for it.

4. Enter the MCP server URL for your Atlas region from the table above.

5. Save the connection and complete the browser sign-in flow when prompted.

6. Enable the Atlas connection in your conversation if your AI tool requires it.

The exact names of settings and buttons vary between AI tools.

## Check that the connection works

Start a conversation with the Atlas connection enabled and ask:

> Use Atlas to list the clusters I can access and show their current state.

If your AI tool asks for permission to use Atlas tools, review and approve the request. A working connection returns a list of your actual clusters. If you recognize them, your tool is connected and has access to your Atlas data.

## Try it on a real task

Once connected, try a question from your day-to-day work:

* "Which of my clusters are driving the most cost?"
* "Where are the rightsizing opportunities?"
* "Help me understand the health of this cluster."

Name the cluster and time period where relevant. You can ask follow-up questions to explore the results.

In clients that support the command, try `/mcp prompts` to explore available prompts. For example:

```text
atlas:
   /mcp__atlas__cluster-health [cluster]
       One-page health summary for a cluster.
   /mcp__atlas__commitment-coverage [window]
       Report commitment and discount coverage against actual usage.
   /mcp__atlas__cost-spike [cluster] [since]
       Find what drove a cost increase for a cluster over a time window.
   /mcp__atlas__rightsizing-review [cluster]
       Summarize workload rightsizing opportunities by potential saving.
```

## Troubleshooting

If you can't connect, work through the following checks:

* **Region**: the server URL or `--api-region` value matches your Atlas region. If you sign in but don't see your clusters, check the region first.
* **Sign-in**: the browser sign-in completed and you approved the connection. If no browser window appeared, the pop-up was likely blocked: retry the connection and allow pop-ups for your tool.
* **Connection enabled**: the Atlas connection is enabled in your tool, and in the conversation if your tool requires it.
* **Tool permissions**: if your tool asks for permission to use Atlas tools, review and approve the request.
* **Restart**: if you configured the connection in a configuration file, fully quit and relaunch your tool. Most tools do not reload MCP configuration without a restart.

For connection errors, check the MCP or connector logs in your tool. In Cursor and Windsurf, open the **Output** panel and select **MCP**. Claude Desktop writes MCP logs to `~/Library/Logs/Claude/`. If the connection still fails, contact your Atlas Account Manager.

## Next steps

<Cards columns={2}>
  <Card title="AtlasPilot Assistant" href="https://docs.atlas.example/docs/atlaspilot-assistant" icon="fa-comments">
    Ask operational questions about your clusters directly in the Atlas console.
  </Card>
  <Card title="MCP server for full API access" href="https://docs.atlas.example/reference/mcp" icon="fa-key">
    Connect AI tools to the full Atlas API with an API key when read-only access is not enough.
  </Card>
</Cards>