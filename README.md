# Guacamole 🥑

A CLI tool that runs opinionated quality checks on your IaC codebase.

Check the [IaC guild guidelines](https://padok-team.github.io/docs-terraform-guidelines/) for more information on the quality checks.

## Table of Contents

- [Install](#install)
  - [With Homebrew](#with-homebrew)
  - [From binary](#from-binary)
  - [From GitHub](#from-github)
- [Usage](#usage)
  - [Run in CI](#run-in-ci)
  - [Skipping individual checks](#skipping-individual-checks)
- [List of checks](#list-of-checks)
  - [Static module check for Terraform](#static-module-check-for-terraform)
  - [Static layer check for Terragrunt](#static-layer-check-for-terragrunt)
  - [State](#state)
- [Demo](#demo)
- [License](#license)

## Install

### With Homebrew

> :information_source: If you use Linux, you can install [Linuxbrew](https://docs.brew.sh/Homebrew-on-Linux)

```bash
brew tap padok-team/tap
brew trust padok-team/tap
brew install guacamole
```

### From binary

Download and install the latest binary on Linux or macOS in one command:

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]') \
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') \
VERSION=$(curl -fsSL https://api.github.com/repos/padok-team/guacamole/releases/latest | grep tag_name | cut -d '"' -f 4) \
&& curl -fsSL "https://github.com/padok-team/guacamole/releases/download/${VERSION}/guacamole_${VERSION#v}_${OS}_${ARCH}.tar.gz" \
  | tar -xz -C /tmp guacamole \
&& sudo install /tmp/guacamole /usr/local/bin/guacamole
```

To install a specific version, replace the `VERSION=...` line with `VERSION=v0.3.4` for example. Windows archives (`.zip`) are available on the [releases page](https://github.com/padok-team/guacamole/releases).

Verify the install:

```bash
guacamole version
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

  - By default, it will launch [module](#static-module-check-for-terraform) and [layer](#static-layer-check-for-terragrunt) checks
  - To launch [layer](#static-layer-check-for-terragrunt) check use `guacamole static layer`
  - To launch [module](#static-module-check-for-terraform) check use `guacamole static module`

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

- CI mode: detects changed Terraform/Terragrunt directories from your Git diff (layers under `layers/`, modules under `base/`, `functional/` or `modules/`), runs scoped static checks and can post a GitLab MR or GitHub PR comment.

  ```bash
  guacamole ci
  ```

  The platform is detected automatically: GitHub when `GITHUB_ACTIONS=true`, GitLab otherwise.

  Required CI environment:

  - `GUACAMOLE_DIFF_BASE_BRANCH`, `CI_MERGE_REQUEST_TARGET_BRANCH_NAME` (GitLab) or `GITHUB_BASE_REF` (GitHub)
  - Optional `GUACAMOLE_MR_SHA` (defaults to `HEAD`)
  - Optional `GUACAMOLE_PROJECT_DIR`, `CI_PROJECT_DIR` (GitLab) or `GITHUB_WORKSPACE` (GitHub) (defaults to current directory)

  The Git history must contain the base branch: on GitHub, use `fetch-depth: 0` with `actions/checkout`. Paths are relative to the project directory, which can be a sub-directory of the repository.

  Optional behaviour:

  - `GUACAMOLE_CI_COMMENT=false` to disable comment posting
  - `GUACAMOLE_CI_SCAN_ALL=true` to scan every tracked layer/module instead of the changed ones (no base branch needed)
  - `GUACAMOLE_CI_FAIL_ON_ERROR=false` to exit with code 0 even when checks fail

  To post a GitLab MR comment, ensure these variables are set:

  - `CI_MERGE_REQUEST_IID`
  - `CI_API_V4_URL`
  - `CI_PROJECT_ID`
  - `GUACAMOLE_GITLAB_TOKEN`

  To post a GitHub PR comment, ensure these variables are set (the comment is updated in place on each run):

  - `GITHUB_REPOSITORY` and `GITHUB_EVENT_PATH` (set by GitHub Actions), or `GUACAMOLE_GITHUB_PR_NUMBER` to target a PR explicitly
  - `GUACAMOLE_GITHUB_TOKEN` or `GITHUB_TOKEN`, with the `pull-requests: write` permission
  - Optional `GITHUB_API_URL` (defaults to `https://api.github.com`, useful for GitHub Enterprise Server)

  On GitHub, the report is also written to the job summary and the `score`, `passed` and `total` step outputs are set. See [Run in CI](#run-in-ci) for ready-to-use configurations.

A verbose mode (`-v`) exists to add more information to the output.

### Run in CI

#### GitHub Actions

Use the [guacamole-action](https://github.com/padok-team/guacamole-action) with `check_type: ci` to post the IaC score of the layers/modules changed by a pull request as a PR comment, updated in place on each push:

```yaml
on:
  pull_request:

permissions:
  contents: read
  pull-requests: write

jobs:
  guacamole:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v6
        with:
          # The base branch history is needed to compute the diff
          fetch-depth: 0

      - name: Run guacamole
        uses: padok-team/guacamole-action@v2
        with:
          check_type: ci
          # Pull requests opened from a fork get a read-only token: skip the comment for them
          comment: ${{ github.event.pull_request.head.repo.full_name == github.repository }}
```

The action also supports `scan_all`, `fail_on_error`, `base_branch` and the other checks (`static`, `state`...): see its [README](https://github.com/padok-team/guacamole-action#inputs) and [examples](https://github.com/padok-team/guacamole-action/tree/main/examples).

#### GitLab CI

```yaml
guacamole:
  stage: test
  image:
    name: ghcr.io/padok-team/guacamole:v0.4.0
    entrypoint: [""]
  script:
    - guacamole ci
  rules:
    - if: $CI_MERGE_REQUEST_IID
```

Set `GUACAMOLE_GITLAB_TOKEN` (access token with the `api` scope) in the project CI/CD variables to post the MR comment, or `GUACAMOLE_CI_COMMENT=false` to skip it.

### Skipping individual checks

You can use inline code comments to skip individual checks for a particular resource.

⚠️ Currently only supports static checks on modules for Terraform ⚠️

To skip a check on a given Terraform definition block resource, apply the following comment pattern inside its scope: `# guacamole-ignore:<check_id> <suppression_comment>`

- <check_id> is one of the available check scanners.
- <suppression_comment> is an optional suppression reason.

Example:

The following comment skips the `TF_NAM_001` check on the resource identified by `network`

```bash
# guacamole-ignore:TF_NAM_001 We will be creating more rg
resource "azurerm_resource_group" "network" {
  name...
```

You can also whitelist entire checks in **modules** by adding them to a `.guacamoleignore` file at the root of your codebase.
The format of the file should be: path of the `module - check ID` to ignore.

```bash
pathtomodule/modules/cloud-run-app TF_MOD_002,TF_MOD_001
pathtomodule/modules/network TF_NAM_001

```

This is the only way to whitelist the check `TF_MOD_002`

You can specify the path of the `.guacamoleignore` file with the `-w` flag.

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
- `TF_DAT_001` - [Data source should not depend on a value computed during apply](https://padok-team.github.io/docs-terraform-guidelines/terraform/donts.html#data-sources-depending-on-values-only-known-at-apply-time)

### Static layer check for Terragrunt

- `TG_DRY_001` - [No duplicate inputs within a layer](https://padok-team.github.io/docs-terraform-guidelines/terragrunt/context_pattern.html#%EF%B8%8F-context)
- `TG_ARC_001` - [terragrunt.hcl should be the last layer to apply](https://padok-team.github.io/docs-terraform-guidelines/terragrunt/context_pattern.html)
- `TG_DRY_002` - Terragrunt locals should not be declared if it is not worthly used

### State

- `TF_MOD_004` - [Use for_each to create multiple resources of the same type](https://padok-team.github.io/docs-terraform-guidelines/terraform/iterate_on_your_resources.html)

## Demo

![Demo](/assets/demo.gif)

## License

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
