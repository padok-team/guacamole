# Guacamole 🥑

A CLI tool that runs opinionated quality checks on your IaC codebase.

Check the [IaC guild guidelines](https://padok-team.github.io/docs-terraform-guidelines/) for more information on the quality checks.

## Install

### With Homebrew

> :information_source: If you use Linux, you can install [Linuxbrew](https://docs.brew.sh/Homebrew-on-Linux)

:warning: A GitHub PAT is required because Guacamole is a private repository for the moment.

Create a GitHub classic Personal Access Token with `repo` permissions.

Export it to your env variables and in your `.zshrc`, `.bashrc` ... :

```bash
export HOMEBREW_GITHUB_API_TOKEN="***********"
```

Then you should be able to download Guacamole using the Padok tap.

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

- [EXPERIMENTAL] State mode : runs quality checks based on your layers' state
  
  We recommend to use this command after checking that your codebase has been initialized properly.


  ```bash
  guacamole state -p /path/to/your/codebase
  ```

- [EXPERIMENTAL] Profile mode : creates a detailed report of the contents of your codebase

  We recommend to use this command after checking that your codebase has been initialized properly.

  ```bash
  guacamole profile -p /path/to/your/codebase
  ```

A verbose mode (`-v`) exists to add more infos to the output.

## Demo

![Demo](/assets/demo.gif)

## List of checks

### Static

- `TF_MOD_001` - Remote module call should be pinned to a specific version
- `TF_MOD_002` - Provider should be defined by the consumer of the module
- `TF_MOD_003` - Required provider versions in modules should be set with ~> operator
- `TF_NAM_001` - Resources and datasources in modules should be named "this" or "these" if their type is unique
- `TF_NAM_002` - snake_case should be used for all resource names
- `TF_NAM_003` - Stuttering in the naming of resources
- `TF_NAM_004` - Variable name's number should match its type
- `TF_VAR_001` - Variable should contain a description
- `TF_VAR_002` - Variable should declare a specific type

### State

- `TF_MOD_004` - Use for_each to create multiple resources of the same type

## License

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
