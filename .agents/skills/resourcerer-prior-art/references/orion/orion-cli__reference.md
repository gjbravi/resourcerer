# Reference

> Commands and options for the Orion CLI


## orion

Orion Command Line Interface

### Synopsis

Automated Kubernetes Workload Rightsizing.

  Find more information at: https://docs.orion.example/

### Options

```
      --sfconfig file   configuration file to use
  -h, --help            help for orion
```

### See also

* [orion apply](#orion-apply)	 - Apply one or more recommendations
* [orion auth](#orion-auth)	 - Manage non-user authorizations
* [orion check](#orion-check)	 - Validate Orion components
* [orion completion](#orion-completion)	 - Generate the autocompletion script for the specified shell
* [orion config](#orion-config)	 - Show the CLI configuration
* [orion create](#orion-create)	 - Create a resource
* [orion delete](#orion-delete)	 - Delete a resource
* [orion describe](#orion-describe)	 - Describe a resource
* [orion edit](#orion-edit)	 - Edit a resource
* [orion get](#orion-get)	 - Fetch resources
* [orion login](#orion-login)	 - Authorize the CLI
* [orion logout](#orion-logout)	 - Revoke the CLI authorization
* [orion manage](#orion-manage)	 - Manage a resource
* [orion mcp](#orion-mcp)	 - Run an MCP server
* [orion patch](#orion-patch)	 - Deploy patches to a cluster or source files
* [orion ping](#orion-ping)	 - Connect to the Orion server (also see 'check connect')
* [orion version](#orion-version)	 - Show the version information
* [orion view](#orion-view)	 - View pages of impact data

## orion apply

Apply one or more recommendations

### Synopsis

Apply initiates the deployment of recommendations.

The apply command is used to accept recommendations, marking them as ready to be
applied by the Orion Applier.

```
orion apply ( NAME | FILTER_OPTIONS )
```

### Options

```
      --recommendation name   recommendation name to filter on
      --cluster name          cluster name to filter on
      --resource type         workload resource type to filter on
      --namespace namespace   workload namespace to filter on
      --name name             workload name to filter on
      --state value           state values to filter on; any of: Requested|Generated|FailedToGenerate|Ready|Applied|FailedToApply|LastApplied|Inactive|all|latest
  -l, --selector query        selector (label query) to filter on
      --force                 do not skip preliminary recommendations
  -h, --help                  help for apply
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface

## orion auth

Manage non-user authorizations

### Synopsis

Manage your Orion authorizations.

```
orion auth
```

### Options

```
  -h, --help   help for auth
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface
* [orion auth create](#orion-auth-create)	 - Create an authorization
* [orion auth list](#orion-auth-list)	 - List authorizations
* [orion auth revoke](#orion-auth-revoke)	 - Revoke authorizations

## orion auth create

Create an authorization

### Synopsis

Create a new Orion authorization.

```
orion auth create NAME
```

### Examples

```
  # Create a values.yaml file for installing a Orion product with Helm
  orion auth create my-token-name > orion-values.yaml

  # Create an organization wide secret with a Orion token for GitHub Actions
  gh secret set --org myOrg ORION_TOKEN <(orion auth create -o token github-actions)
```

### Options

```
  -o, --output format   output format; one of: yaml|json|token (default "yaml")
  -h, --help            help for auth create
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion auth](#orion-auth)	 - Manage non-user authorizations

## orion auth list

List authorizations

### Synopsis

List Orion authorizations.

```
orion auth list
```

### Options

```
  -o, --output format          output format; one of: json|yaml|name|wide|csv|custom-columns|go-template
      --no-headers             don't print headers
      --template go-template   template string to use for '-o go-template'
  -h, --help                   help for auth list
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion auth](#orion-auth)	 - Manage non-user authorizations

## orion auth revoke

Revoke authorizations

### Synopsis

Revokes a Orion authorization.

```
orion auth revoke AUTHORIZATION ...
```

### Options

```
  -h, --help   help for auth revoke
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion auth](#orion-auth)	 - Manage non-user authorizations

## orion check

Validate Orion components

### Synopsis

Validates components of the Orion system.

```
orion check
```

### Options

```
  -h, --help   help for check
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface
* [orion check connect](#orion-check-connect)	 - Check server connections
* [orion check version](#orion-check-version)	 - Check the CLI version

## orion check connect

Check server connections

### Synopsis

Test the connection to Orion servers.

```
orion check connect api|auth|in|registry|app|docs|downloads|localhost ...
```

### Examples

```
  # Check connectivity from inside a cluster
  kubectl run orion-check-connect --attach --rm --restart Never --quiet \
    --image registry.orion.example/library/orion-cli -- check connect
```

### Options

```
  -o, --output format          output format; one of: json|yaml|name|wide|csv|custom-columns|go-template
      --no-headers             don't print headers
      --template go-template   template string to use for '-o go-template'
  -h, --help                   help for check connect
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion check](#orion-check)	 - Validate Orion components

## orion check version

Check the CLI version

### Synopsis

Check the current CLI version.

```
orion check version
```

### Options

```
  -h, --help   help for check version
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion check](#orion-check)	 - Validate Orion components

## orion completion

Generate the autocompletion script for the specified shell

### Synopsis

Generate the autocompletion script for orion for the specified shell.
See each sub-command's help for details on how to use the generated script.

```
orion completion
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface
* [orion completion bash](#orion-completion-bash)	 - Generate the autocompletion script for bash
* [orion completion fish](#orion-completion-fish)	 - Generate the autocompletion script for fish
* [orion completion powershell](#orion-completion-powershell)	 - Generate the autocompletion script for powershell
* [orion completion zsh](#orion-completion-zsh)	 - Generate the autocompletion script for zsh

## orion completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(orion completion bash)

To load completions for every new session, execute once:

#### Linux:

	orion completion bash > /etc/bash_completion.d/orion

#### macOS:

	orion completion bash > $(brew --prefix)/etc/bash_completion.d/orion

You will need to start a new shell for this setup to take effect.

```
orion completion bash
```

### Options

```
      --no-descriptions   disable completion descriptions
  -h, --help              help for completion bash
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion completion](#orion-completion)	 - Generate the autocompletion script for the specified shell

## orion completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	orion completion fish | source

To load completions for every new session, execute once:

	orion completion fish > ~/.config/fish/completions/orion.fish

You will need to start a new shell for this setup to take effect.

```
orion completion fish
```

### Options

```
      --no-descriptions   disable completion descriptions
  -h, --help              help for completion fish
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion completion](#orion-completion)	 - Generate the autocompletion script for the specified shell

## orion completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	orion completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.

```
orion completion powershell
```

### Options

```
      --no-descriptions   disable completion descriptions
  -h, --help              help for completion powershell
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion completion](#orion-completion)	 - Generate the autocompletion script for the specified shell

## orion completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(orion completion zsh)

To load completions for every new session, execute once:

#### Linux:

	orion completion zsh > "${fpath[1]}/_orion"

#### macOS:

	orion completion zsh > $(brew --prefix)/share/zsh/site-functions/_orion

You will need to start a new shell for this setup to take effect.

```
orion completion zsh
```

### Options

```
      --no-descriptions   disable completion descriptions
  -h, --help              help for completion zsh
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion completion](#orion-completion)	 - Generate the autocompletion script for the specified shell

## orion config

Show the CLI configuration

### Synopsis

View Orion CLI configuration.

The Orion CLI loads configuration from command line flags, environment
variables, and a configuration file (in that order). The location of the file
is determined by the `--sfconfig` flag or the `ORION_CONFIG` environment
variable (collectively, "the sfconfig"). If the sfconfig DOES NOT contain a
file path separator (e.g. 'config', the default), the file is resolved relative
to the OS specific "user config" directory ('~/.config/orion' on Linux).
If the sfconfig DOES contain a path separator (e.g. './sfconfig'), the file is
resolved relative to current working directory.

The configuration file may contain sensitive information: for example, after
logging in using 'orion login' the token used to access the
Orion API will be present in the configuration file.

```
orion config
```

### Options

```
      --raw                    display the raw configuration file (may contain sensitive information!)
      --redact                 redact credentials
  -o, --output format          output format; one of: yaml|json|go-template (default "yaml")
      --template go-template   template string to use for '-o go-template'
  -h, --help                   help for config
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface

## orion create

Create a resource

### Synopsis

Creates new Orion resources.

```
orion create
```

### Options

```
  -h, --help   help for create
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface
* [orion create recommendation](#orion-create-recommendation)	 - Create recommendations

## orion create recommendation

Create recommendations

### Synopsis

Creates a new Orion recommendation.

```
orion create recommendation ( NAME | FILTER_OPTIONS )
```

### Options

```
      --cluster name               cluster name to filter on
      --resource type              workload resource type to filter on
      --namespace namespace        workload namespace to filter on
      --name name                  workload name to filter on
      --state value                state values to filter on; any of: Created|Inactive|all
  -l, --selector query             selector (label query) to filter on
      --set-label key=value        label key=value pairs to assign (default [])
      --set-annotation key=value   annotation key=value pairs to assign (default [])
      --valid-for period           the validity period for the recommendation
      --auto-deploy                automatically deploy the recommendation
      --wait                       wait for recommendations to be generated or applied (depending on auto-deploy)
      --timeout dur                duration to wait for recommendations to be applied (default 10m0s)
  -o, --output format              output format; one of: json|yaml
  -h, --help                       help for create recommendation
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion create](#orion-create)	 - Create a resource

## orion delete

Delete a resource

### Synopsis

Delete Orion resources.

```
orion delete
```

### Options

```
  -h, --help   help for delete
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface
* [orion delete cluster](#orion-delete-cluster)	 - Delete a cluster
* [orion delete workload](#orion-delete-workload)	 - Delete a workload

## orion delete cluster

Delete a cluster

### Synopsis

Delete a Orion cluster.

```
orion delete cluster ( NAME | FILTER_OPTIONS )
```

### Options

```
      --state value      state values to filter on; any of: Created|Inactive|all
  -l, --selector query   selector (label query) to filter on
  -o, --output format    output format; name
  -h, --help             help for delete cluster
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion delete](#orion-delete)	 - Delete a resource

## orion delete workload

Delete a workload

### Synopsis

Delete a Orion workload.

```
orion delete workload ( NAME | FILTER_OPTIONS )
```

### Options

```
      --cluster name          cluster name to filter on
      --resource type         workload resource type to filter on
      --namespace namespace   workload namespace to filter on
      --name name             workload name to filter on
      --state value           state values to filter on; any of: Created|Inactive|all
  -l, --selector query        selector (label query) to filter on
  -o, --output format         output format; name
  -h, --help                  help for delete workload
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion delete](#orion-delete)	 - Delete a resource

## orion describe

Describe a resource

### Synopsis

Describe one or more Orion resources.

```
orion describe
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface
* [orion describe cluster](#orion-describe-cluster)	 - Describe one or more clusters
* [orion describe namespace](#orion-describe-namespace)	 - Describe one or more namespaces
* [orion describe recommendation](#orion-describe-recommendation)	 - Describe one or more recommendations
* [orion describe workload](#orion-describe-workload)	 - Describe one or more workloads

## orion describe cluster

Describe one or more clusters

### Synopsis

Describes Orion clusters.

```
orion describe cluster ( NAME | FILTER_OPTIONS )
```

### Options

```
      --state value            state values to filter on; any of: Created|Inactive|all
  -l, --selector query         selector (label query) to filter on
  -o, --output format          output format; one of: yaml|json|go-template
      --template go-template   template string to use for '-o go-template' (default "{{ template \"ClusterDescriptor\" . }}")
  -h, --help                   help for describe cluster
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion describe](#orion-describe)	 - Describe a resource

## orion describe namespace

Describe one or more namespaces

### Synopsis

Describes a namespace for Orion workloads.

```
orion describe namespace ( NAME | FILTER_OPTIONS )
```

### Options

```
      --cluster name             cluster name to filter on
      --cluster-selector query   cluster selector (label query) to filter on
      --sort-by column           sort using column name; one of: namespace|cluster|workloads|impact
      --sort-ascending           change the sort direction
      --per-page n               number of results to show per page (max 100) (default 25)
  -o, --output format            output format; one of: yaml|json|go-template
      --template go-template     template string to use for '-o go-template' (default "{{ template \"NamespaceDescriptor\" . }}")
  -h, --help                     help for describe namespace
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion describe](#orion-describe)	 - Describe a resource

## orion describe recommendation

Describe one or more recommendations

### Synopsis

Describes Orion recommendations.

```
orion describe recommendation ( NAME | FILTER_OPTIONS )
```

### Options

```
      --recommendation name    recommendation name to filter on
      --cluster name           cluster name to filter on
      --resource type          workload resource type to filter on
      --namespace namespace    workload namespace to filter on
      --name name              workload name to filter on
      --state value            state values to filter on; any of: Requested|Generated|FailedToGenerate|Ready|Applied|FailedToApply|LastApplied|Inactive|all|latest
  -l, --selector query         selector (label query) to filter on
  -o, --output format          output format; one of: yaml|json|go-template
      --template go-template   template string to use for '-o go-template' (default "{{ template \"RecommendationDescriptor\" . }}")
  -h, --help                   help for describe recommendation
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion describe](#orion-describe)	 - Describe a resource

## orion describe workload

Describe one or more workloads

### Synopsis

Describes Orion workloads.

```
orion describe workload ( NAME | FILTER_OPTIONS )
```

### Options

```
      --cluster name           cluster name to filter on
      --resource type          workload resource type to filter on
      --namespace namespace    workload namespace to filter on
      --name name              workload name to filter on
      --state value            state values to filter on; any of: Created|Inactive|all
  -l, --selector query         selector (label query) to filter on
  -o, --output format          output format; one of: yaml|json|go-template
      --template go-template   template string to use for '-o go-template' (default "{{ template \"WorkloadDescriptor\" . }}")
  -h, --help                   help for describe workload
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion describe](#orion-describe)	 - Describe a resource

## orion edit

Edit a resource

### Synopsis

Edit Orion configuration.

```
orion edit
```

### Options

```
  -h, --help   help for edit
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface
* [orion edit cluster](#orion-edit-cluster)	 - Edit one or more clusters
* [orion edit configuration](#orion-edit-configuration)	 - Edit configuration
* [orion edit workload](#orion-edit-workload)	 - Edit one or more workloads

## orion edit cluster

Edit one or more clusters

### Synopsis

Edit configuration for a Orion cluster.

```
orion edit cluster ( NAME | FILTER_OPTIONS )
```

### Options

```
      --managed           edit the managed configuration (default true)
      --interactive       always open the editor before patching
      --set key1=val1     set values on the command line (for example, key1=val1,key2=val2,...)
      --type type         patch type; one of: json|strategic
      --patch patch       raw patch contents
  -f, --patch-file file   a file containing the patch to use
      --dry-run mode      dry run mode used to apply manifests; one of: none|server|client
      --force             ignore safety checks and force the update
      --reset-uid         reset the Kubernetes identifier associated with the cluster
      --state value       state values to filter on; any of: Created|Inactive|all
  -l, --selector query    selector (label query) to filter on
  -o, --output format     output format; one of: name|json|yaml
  -h, --help              help for edit cluster
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion edit](#orion-edit)	 - Edit a resource

## orion edit configuration

Edit configuration

### Synopsis

Edit configuration for a Orion configuration.

```
orion edit configuration
```

### Options

```
      --managed           edit the managed configuration (default true)
      --interactive       always open the editor before patching
      --set key1=val1     set values on the command line (for example, key1=val1,key2=val2,...)
      --type type         patch type; one of: json|strategic
      --patch patch       raw patch contents
  -f, --patch-file file   a file containing the patch to use
      --dry-run mode      dry run mode used to apply manifests; one of: none|server|client
      --force             ignore safety checks and force the update
  -o, --output format     output format; one of: name|json|yaml
  -h, --help              help for edit configuration
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion edit](#orion-edit)	 - Edit a resource

## orion edit workload

Edit one or more workloads

### Synopsis

Edit configuration for a Orion workload.

```
orion edit workload ( NAME | FILTER_OPTIONS )
```

### Options

```
      --managed               edit the managed configuration (default true)
      --interactive           always open the editor before patching
      --set key1=val1         set values on the command line (for example, key1=val1,key2=val2,...)
      --type type             patch type; one of: json|strategic
      --patch patch           raw patch contents
  -f, --patch-file file       a file containing the patch to use
      --dry-run mode          dry run mode used to apply manifests; one of: none|server|client
      --force                 ignore safety checks and force the update
      --cluster name          cluster name to filter on
      --resource type         workload resource type to filter on
      --namespace namespace   workload namespace to filter on
      --name name             workload name to filter on
      --state value           state values to filter on; any of: Created|Inactive|all
  -l, --selector query        selector (label query) to filter on
  -o, --output format         output format; one of: name|json|yaml
  -h, --help                  help for edit workload
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion edit](#orion-edit)	 - Edit a resource

## orion get

Fetch resources

### Synopsis

Display Orion configuration.

```
orion get
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface
* [orion get clusters](#orion-get-clusters)	 - Get clusters
* [orion get configuration](#orion-get-configuration)	 - Get the global configuration
* [orion get recommendations](#orion-get-recommendations)	 - Get recommendations
* [orion get workloads](#orion-get-workloads)	 - Get workloads

## orion get clusters

Get clusters

### Synopsis

Get configuration for a Orion clusters.

```
orion get clusters ( NAME | FILTER_OPTIONS )
```

### Options

```
      --state value               state values to filter on; any of: Created|Inactive|all
  -l, --selector query            selector (label query) to filter on
      --field-selector selector   field selector to perform client side filtering with, supports sort-by fields
      --sort-by column            sort using column name; one of: name|agent_version|applier_version|kubernetes_version|state|age
  -o, --output format             output format; one of: json|yaml|name|wide|csv|custom-columns|go-template
      --no-headers                don't print headers
      --show-labels               when printing, show all labels as the last column
      --template go-template      template string to use for '-o go-template'
  -h, --help                      help for get clusters
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion get](#orion-get)	 - Fetch resources

## orion get configuration

Get the global configuration

### Synopsis

Get configuration for a Orion configuration.

```
orion get configuration
```

### Options

```
  -o, --output format   output format; one of: json|yaml (default "json")
  -h, --help            help for get configuration
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion get](#orion-get)	 - Fetch resources

## orion get recommendations

Get recommendations

### Synopsis

Get full recommendations, including observed state and net impact.

```
orion get recommendations ( NAME | FILTER_OPTIONS )
```

### Examples

```
  # Fetch the recommendation for each deployment in the default namespace
  # NOTE: The --recommendation option cannot be used because the workload name is incomplete
  orion get recommendations my-cluster/default/deployments

  # Fetch all recommendations for a workload
  orion get recommendations my-cluster/default/deployments/nginx

  # Fetch a specific recommendation
  orion get recommendations my-cluster/default/deployments/nginx/1792004520
```

### Options

```
      --recommendation name       recommendation name to filter on
      --cluster name              cluster name to filter on
      --resource type             workload resource type to filter on
      --namespace namespace       workload namespace to filter on
      --name name                 workload name to filter on
      --state value               state values to filter on; any of: Requested|Generated|FailedToGenerate|Ready|Applied|FailedToApply|LastApplied|Inactive|all|latest
  -l, --selector query            selector (label query) to filter on
      --field-selector selector   field selector to perform client side filtering with, supports sort-by fields
      --sort-by column            sort using column name; one of: cluster|namespace|type|name|timestamp|score|net_impact|state|status|reason
  -o, --output format             output format; one of: json|yaml|name|wide|csv|custom-columns|go-template
      --no-headers                don't print headers
      --show-labels               when printing, show all labels as the last column
      --template go-template      template string to use for '-o go-template'
      --patches for[="-"]         show recommendation patches for a specific audience; one of: client-patch|pod-admission|reconciler
      --container-resources       show container level details
      --auto-scaling              show auto-scaling level details
  -h, --help                      help for get recommendations
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion get](#orion-get)	 - Fetch resources

## orion get workloads

Get workloads

### Synopsis

Get configuration for a Orion workloads.

```
orion get workloads ( NAME | FILTER_OPTIONS )
```

### Options

```
      --cluster name              cluster name to filter on
      --resource type             workload resource type to filter on
      --namespace namespace       workload namespace to filter on
      --name name                 workload name to filter on
      --state value               state values to filter on; any of: Created|Inactive|all
  -l, --selector query            selector (label query) to filter on
      --field-selector selector   field selector to perform client side filtering with, supports sort-by fields
      --sort-by column            sort using column name; one of: cluster|namespace|type|name|auto_deploy|hpa_enabled|gpu_enabled|state|age
  -o, --output format             output format; one of: json|yaml|name|wide|csv|custom-columns|go-template
      --no-headers                don't print headers
      --show-labels               when printing, show all labels as the last column
      --template go-template      template string to use for '-o go-template'
  -h, --help                      help for get workloads
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion get](#orion-get)	 - Fetch resources

## orion login

Authorize the CLI

### Synopsis

Log into your Orion account.

```
orion login [ SERVER ]
```

### Examples

```
  # Login using your workstation's default web browser.
  orion login

  # If a web browser is not locally available, or if the workstation is
  # restricted from listening on port 8085, it may be necessary to perform the
  # authorization from another device. Use the "--url" flag to get a URL that
  # can be entered into a web browser manually.
  orion login --url

  # If you are having problems logging in, try using `orion check connect`
  # to troubleshoot any connection issues.
```

### Options

```
      --org id   organization identifier to use
      --url      display the URL instead of opening a browser
      --qr       display a QR code instead of opening a browser
  -h, --help     help for login
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface

## orion logout

Revoke the CLI authorization

### Synopsis

Log out of your Orion account.

```
orion logout
```

### Options

```
  -h, --help   help for logout
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface

## orion manage

Manage a resource

### Synopsis

Manage the Orion configuration.

```
orion manage
```

### Options

```
  -h, --help   help for manage
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface
* [orion manage cluster](#orion-manage-cluster)	 - Manage one or more clusters
* [orion manage workload](#orion-manage-workload)	 - Manage one or more workloads

## orion manage cluster

Manage one or more clusters

### Synopsis

Manage the configuration of a Orion cluster.

```
orion manage cluster ( NAME | FILTER_OPTIONS ) KEY=VALUE ...
```

### Options

```
      --state value         state values to filter on; any of: Created|Inactive|all
  -l, --selector query      selector (label query) to filter on
  -o, --output format       output format; one of: name|yaml|json
  -f, --filename filename   filename to load management settings from
      --dry-run mode        dry run mode used to apply manifests; one of: none|server|client
  -h, --help                help for manage cluster
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion manage](#orion-manage)	 - Manage a resource

## orion manage workload

Manage one or more workloads

### Synopsis

Manage the configuration of a Orion workload.

```
orion manage workload ( NAME | FILTER_OPTIONS ) KEY=VALUE ...
```

### Options

```
      --cluster name          cluster name to filter on
      --resource type         workload resource type to filter on
      --namespace namespace   workload namespace to filter on
      --name name             workload name to filter on
      --state value           state values to filter on; any of: Created|Inactive|all
  -l, --selector query        selector (label query) to filter on
  -o, --output format         output format; one of: name|yaml|json
  -f, --filename filename     filename to load management settings from
      --dry-run mode          dry run mode used to apply manifests; one of: none|server|client
  -h, --help                  help for manage workload
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion manage](#orion-manage)	 - Manage a resource

## orion mcp

Run an MCP server

### Synopsis

Start a stdio MCP server that allows AI assistants to interact with the Orion platform.

```
orion mcp
```

### Options

```
      --log-level level   set the log level (default info)
  -h, --help              help for mcp
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface

## orion patch

Deploy patches to a cluster or source files

### Synopsis

Patch Kubernetes manifests using Orion recommendations.

For each of the matching recommendations, the patch command will collect the
necessary patches and consume them in one of the following modes:

When using the `--directory` flag, the patches for the specified recommendations
are written directly to the indicated directory. The `--filename-pattern` flag
accepts placeholders like `{ClusterName}`, `{Namespace}`, and `{Kind}` to split
patches between multiple files in the directory.

When using the `--filename` flag, the patches for the specified recommendations
are applied to the manifests found in the files. The patched manifests can be
written back to disk or emitted to stdout to be processed as part of a larger
pipeline.

If neither `--directory` or `--filename` are used, each patch will be applied
by calling `kubectl patch` with the appropriate arguments.

```
orion patch ( NAME | FILTER_OPTIONS )
```

### Examples

```
  # Write all the recommended patches for the cluster 'my-cluster' to disk
  # NOTE: You can also view patches using `orion get recommendations --patches my-cluster`
  orion patch --directory my-cluster-patches --filename-pattern '{Namespace}_{Kind}_{Name}.json' my-cluster

  # Preview applying recommended patches for the 'default' namespace on the cluster 'my-cluster'
  orion patch --dry-run=server --output=yaml my-cluster/default

  # Apply the above patches
  # NOTE: This will also update the Orion status to "Applied"
  orion patch my-cluster/default

  # Emit patched local manifests for 'my-cluster' and pipe them into `kubectl apply`
  orion patch --filename my-manifests.yaml my-cluster | kubectl apply -f -

  # Overwrite manifests for workloads named 'foobar' in 'my-cluster' found in a local directory
  # IMPORTANT: The `--in-place` flag does not create backups!
  orion patch --filename my-gitops-repo/ --in-place --cluster my-cluster --name foobar

  # Post render Helm output ('--filename=-' reads from stdin)
  helm template my-release charts/my-local-charts --namespace example \
    --post-renderer orion --post-renderer-args patch --post-renderer-args '--filename=-' \
    --post-renderer-args my-cluster/example --post-renderer-args --state=latest
```

### Options

```
      --recommendation name                     recommendation name to filter on
      --cluster name                            cluster name to filter on
      --resource type                           workload resource type to filter on
      --namespace namespace                     workload namespace to filter on
      --name name                               workload name to filter on
      --state value                             state values to filter on; any of: Generated|Ready|Applied|LastApplied|Inactive
  -l, --selector query                          selector (label query) to filter on
  -d, --directory dir                           directory to write patches to (instead of patching the cluster)
      --filename-pattern pattern                filename pattern to write patches to (default "{Kind}_{Name}.json")
  -f, --filename name                           file names of manifests to patch (instead of patching the cluster)
      --in-place                                patch file system resource in-place WITHOUT backups
      --cluster-context cluster=context         a cluster=context mapping of cluster name to kubeconfig context name to use (default [])
      --cluster-kubeconfig cluster=kubeconfig   a cluster=kubeconfig mapping of cluster name to kubeconfig file to use (default [])
      --dry-run mode                            dry run mode used to apply manifests; one of: none|server|client
      --kubeconfig file                         path to the kubeconfig file to use for CLI requests
      --context name                            the name of the kubeconfig context to use
  -o, --output format                           output format; one of: json|yaml
      --skip-apply                              do mark the recommendation as 'Applied' when patching cluster resources
  -h, --help                                    help for patch
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface

## orion ping

Connect to the Orion server (also see 'check connect')

### Synopsis

Check basic connectivity to the Orion cloud.

```
orion ping
```

### Options

```
  -h, --help   help for ping
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface

## orion version

Show the version information

### Synopsis

Display the full version information for the application.

```
orion version [flags]
```

### Options

```
  -h, --help                   help for version
  -o, --output format          output format
      --template go-template   template string to use for '-o go-template'
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface

## orion view

View pages of impact data

### Synopsis

View aggregated Orion recommendations by impact.

```
orion view
```

### Options

```
  -h, --help   help for view
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion](#orion)	 - Orion Command Line Interface
* [orion view clusters](#orion-view-clusters)	 - View clusters
* [orion view namespaces](#orion-view-namespaces)	 - View namespaces
* [orion view workloads](#orion-view-workloads)	 - View workloads

## orion view clusters

View clusters

### Synopsis

Get impact reports for clusters.

```
orion view clusters [ QUERY ]
```

### Options

```
  -l, --selector query         selector (label query) to filter on
      --sort-by column         sort using column name; one of: cluster|workloads|impact
      --sort-ascending         change the sort direction
      --per-page n             number of results to show per page (max 100) (default 25)
      --page number            page number to show results from
  -o, --output format          output format; one of: json|yaml|name|wide|csv|custom-columns|go-template
      --no-headers             don't print headers
      --template go-template   template string to use for '-o go-template'
  -h, --help                   help for view clusters
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion view](#orion-view)	 - View pages of impact data

## orion view namespaces

View namespaces

### Synopsis

Get impact reports for namespaces.

```
orion view namespaces [ QUERY ]
```

### Options

```
      --cluster name             cluster name to filter on
      --cluster-selector query   cluster selector (label query) to filter on
      --sort-by column           sort using column name; one of: namespace|cluster|workloads|impact
      --sort-ascending           change the sort direction
      --per-page n               number of results to show per page (max 100) (default 25)
      --page number              page number to show results from
  -o, --output format            output format; one of: json|yaml|name|wide|csv|custom-columns|go-template
      --no-headers               don't print headers
      --template go-template     template string to use for '-o go-template'
  -h, --help                     help for view namespaces
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion view](#orion-view)	 - View pages of impact data

## orion view workloads

View workloads

### Synopsis

Get impact reports for workloads.

```
orion view workloads [ QUERY ]
```

### Examples

```
  # IMPORTANT: The argument is a search query, NOT a resource name/path

  # Search for workloads that contain the text "foo"
  orion view workloads foo

  # View workloads in the "default" namespace of the "my-cluster" cluster
  orion view workloads --cluster my-cluster --namespace default
```

### Options

```
      --cluster name              cluster name to filter on
      --resource type             workload resource type to filter on
      --namespace namespace       workload namespace to filter on
      --status value              status values to filter on
      --status-reason value       status reason values to filter on
  -l, --selector query            selector (label query) to filter on
      --cluster-selector query    cluster selector (label query) to filter on
      --auto-deploy               filter on workloads with auto-deploy enabled
      --hpa                       filter on workloads with auto-scaling enabled
      --gpu                       filter on workloads with GPU enabled
      --min-oom-rate rate         filter on minimum OOM rate
      --min-cpu-throttling rate   filter on minimum CPU throttling rate
      --over-provisioned          filter on over provisioned workloads
      --under-provisioned         filter on under provisioned workloads
      --runtime classification    filter on workload runtime classification; java
      --sort-by column            sort using column name; one of: name|cluster|namespace|resource|status|impact|oom-rate|cpu-throttling-rate
      --sort-ascending            change the sort direction
      --per-page n                number of results to show per page (max 100) (default 25)
      --page number               page number to show results from
  -o, --output format             output format; one of: json|yaml|name|wide|csv|custom-columns|go-template
      --no-headers                don't print headers
      --template go-template      template string to use for '-o go-template'
      --group-by group            group results
  -h, --help                      help for view workloads
```

### Options inherited from parent commands

```
      --sfconfig file   configuration file to use
```

### See also

* [orion view](#orion-view)	 - View pages of impact data
