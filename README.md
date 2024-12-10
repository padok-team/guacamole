# Guacamole 🥑

A CLI tool that runs opinionated quality checks on your IaC codebase.

Check the [IaC guild guidelines](https://padok-team.github.io/docs-terraform-guidelines/) for more information on the quality checks.

## Install

### With Homebrew

> :information_source: If you use Linux, you can install [Linuxbrew](https://docs.brew.sh/Homebrew-on-Linux)

```bash
brew tap padok-team/tap
brew install guacamole
```

### From GitHub

**Prerequisites :**

- Golang
- Terraform
- Terragrunt

One-liner installer (in `/tmp`) :

```bash
DIR=$(pwd) cd /tmp && git clone git@github.com:padok-team/guacamole.git && cd guacamole && go build && alias guacamole=/tmp/guacamole/guacamole && cd $DIR
```

For a more permanent installation, just move the `/tmp/guacamole/guacamole` binary into a directory present in your `$PATH`.

## Usage

Three modes currently exist :

- Static mode : runs quality checks on the codebase without running Terraform / Terragrunt commands

  ```bash
  guacamole static -p /path/to/your/codebase
  ```

  - By default, it will launch [module](#static-module-check) and [layer](#static-layer-check) checks
  - To launch [layer](#static-layer-check) check use `guacamole static layer`
  - To launch [module](#static-module-check) check use `guacamole static module`

- [EXPERIMENTAL] State mode : runs quality checks based on your layers' state

  We recommend using this command after checking that your codebase has been initialized properly.

  ```bash
  guacamole state -p /path/to/your/codebase
  ```

- [EXPERIMENTAL] Profile mode : creates a detailed report of the contents of your codebase

  We recommend using this command after checking that your codebase has been initialized properly.

  ```bash
  guacamole profile -p /path/to/your/codebase
  ```

A verbose mode (`-v`) exists to add more information to the output.

**Skipping individual checks**

You can use inline code comments to skip individual checks for a particular resource.

To skip a check on a given Terraform definition block resource, apply the following comment pattern inside its scope: `# guacamole-ignore:<check_id> <suppression_comment>`

    <check_id> is one of the available check scanners.
    <suppression_comment> is an optional suppression reason.

Example:

The following comment skips the `TF_NAM_001` check on the resource identified by `network`

```bash
# guacamole-ignore:TF_NAM_001 We will be creating more rg
resource "azurerm_resource_group" "network" {
  name...
```

⚠️ The following checks can't be whitelisted : `TF_MOD_002`

## List of checks

### Static module check for Terraform

- `TF_MOD_001` - [Remote module call should be pinned to a specific version](https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_versioning.html#module-layer-versioning)
- `TF_MOD_002` - [Provider should be defined by the consumer of the module](https://padok-team.github.io/docs-terraform-guidelines/terraform/donts.html#using-provider-block-in-modules)
- `TF_MOD_003` - [Required provider versions in modules should be set with ~> operator](https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_versioning.html#required-providers-version-for-modules)
- `TF_NAM_001` - [Resources in modules should be named "this" or "these" if their type is unique](https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#resource-andor-data-source-naming)
- `TF_NAM_002` - [snake_case should be used for all resource names](https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#resource-andor-data-source-naming)
- `TF_NAM_003` - [Stuttering in the naming of resources](https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#resource-andor-data-source-naming)
- `TF_NAM_004` - [Variable name's number should match its type](https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#variables)
- `TF_NAM_005` - [Resources and data sources should not be named \"this\" or \"these\" if there are more than 1 of the same type](https://padok-team.github.io/docs-terraform-guidelines/terraform/terraform_naming.html#resource-andor-data-source-naming)
- `TF_VAR_001` - [Variable should contain a description](https://padok-team.github.io/docs-terraform-guidelines/terraform/donts.html#variables)
- `TF_VAR_002` - [Variable should declare a specific type](https://padok-team.github.io/docs-terraform-guidelines/terraform/donts.html#using-type-any-in-variables)

### Static layer check for Terragrunt

- `TG_DRY_001` - [No duplicate inputs within a layer](https://padok-team.github.io/docs-terraform-guidelines/terragrunt/context_pattern.html#%EF%B8%8F-context)

### State

- `TF_MOD_004` - [Use for_each to create multiple resources of the same type](https://padok-team.github.io/docs-terraform-guidelines/terraform/iterate_on_your_resources.html)

## Demo

![Demo](/assets/demo.gif)

## License

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
