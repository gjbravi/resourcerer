# Orion CLI

> Install and authenticate the Orion CLI


The Orion CLI is a standalone binary that provides the core set of capabilities for managing Orion. To connect AI agents to Orion, use the [hosted MCP server](guides__mcp-server.md).

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Getting started</div>


- If you haven't already, create a Orion account at [**app.orion.example/signup**](https://app.orion.example/signup).
- Alternatively, log in to our [**sandbox**](/docs/sandbox), then install and authenticate to explore the CLI without impacting workloads.
</div>


## Install

You can download and install using:


**Linux**

Choose one of the two methods below to install the Orion CLI.

Copy the following command to a terminal window:

```bash
# Automatically selects either AMD64 or ARM64 architecture, downloads
# the appropriate binary, then moves it to a location in PATH
{ [ "$(uname -sm)" = "Linux x86_64"  ] && curl -L https://downloads.orion.example/orion-cli/latest/orion_linux_amd64.tar.gz | tar -xz; } ||
{ [ "$(uname -sm)" = "Linux aarch64" ] && curl -L https://downloads.orion.example/orion-cli/latest/orion_linux_arm64.tar.gz | tar -xz; } &&
sudo mv orion /usr/local/bin/
```

* * *

Download the appropriate Linux binary and extract the files.

[Amd64](https://downloads.orion.example/orion-cli/latest/orion_linux_amd64.tar.gz) | [Arm64](https://downloads.orion.example/orion-cli/latest/orion_linux_arm64.tar.gz)

Then, run:

```bash
sudo mv orion /usr/local/bin/
```

**Windows**

Choose one of the two methods below to install the Orion CLI.

Download the appropriate Windows binary and extract the files.

[Amd64](https://downloads.orion.example/orion-cli/latest/orion_windows_amd64.zip) | [Arm64](https://downloads.orion.example/orion-cli/latest/orion_windows_arm64.zip)

Then, move `orion.exe` to a folder defined in your Windows PATH environment variable.

* * *

Run the following PowerShell install command:

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
If ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { Invoke-WebRequest -Uri "https://downloads.orion.example/orion-cli/latest/orion_windows_amd64.zip" -Outfile "orion.zip" }
If ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { Invoke-WebRequest -Uri "https://downloads.orion.example/orion-cli/latest/orion_windows_arm64.zip" -Outfile "orion.zip" }
Expand-Archive "orion.zip" "." -WarningVariable $w; if ($w.Count -eq 0) { Remove-Item "orion.zip" } 
```

Then, move `orion.exe` to a folder defined in your Windows PATH environment variable.

**macOS**

Choose one of the three methods below to install the Orion CLI.

Homebrew:

```sh
brew install theorion/tap/orion
```

* * *

Copy the following macOS install command to a terminal:

```bash
# Automatically selects either AMD64 or ARM64 architecture, downloads
# the appropriate binary, then moves it to a location in PATH
{ [ "$(uname -sm)" = "Darwin x86_64" ] && curl -L https://downloads.orion.example/orion-cli/latest/orion_darwin_amd64.tar.gz | tar -xz; } ||
{ [ "$(uname -sm)" = "Darwin arm64"  ] && curl -L https://downloads.orion.example/orion-cli/latest/orion_darwin_arm64.tar.gz | tar -xz; } &&
sudo mv orion /usr/local/bin/
```

* * *

Download the appropriate macOS binary and install it using your preferred binary installer.

[Amd64](https://downloads.orion.example/orion-cli/latest/orion_darwin_amd64.tar.gz) | [Arm64](https://downloads.orion.example/orion-cli/latest/orion_darwin_arm64.tar.gz)

**Docker**

We provide a container image at `registry.orion.example/library/orion-cli` (for `linux/amd64` and `linux/arm64`):

```bash
docker pull registry.orion.example/library/orion-cli
```

If you prefer to use the container image directly, set the `ORION_TOKEN` environment variable (from [`auth create -o token`](orion-cli__reference.md#orion-auth-create)).



## Authenticate

Before using other CLI commands, authenticate your session:

```sh
orion login
```

This command opens a browser and attempts to authenticate using your Orion account. If you aren't already logged in, you will be prompted to sign in.

See the [CLI reference for `orion login`](orion-cli__reference.md#orion-login) for additional details and options.

Once authenticated, you're ready to use the Orion CLI.

## Releases

For the list of CLI releases, see [Orion CLI releases](releases__orion-cli.md).

## Documentation

- Get started with common [Orion CLI workflows](orion-cli__common-cli-workflows.md)
- [Connect AI agents to Orion with MCP](guides__mcp-server.md)
- View the full [Orion CLI command reference](orion-cli__reference.md)

---

Section pages:

- [Common Orion CLI workflows](orion-cli__common-cli-workflows.md): Common Orion CLI tasks, from authentication to managing recommendations
- [Reference](orion-cli__reference.md): Commands and options for the Orion CLI
