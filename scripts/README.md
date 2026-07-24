# Scripts

Utility scripts for developing and testing Guacamole.

## `docker-test-version.sh`

Builds a Docker image of Guacamole from a given git branch and drops you into an
interactive shell where the `guacamole` binary is on the `PATH`. Use it to try a
version end to end — in a clean, reproducible environment — before releasing it
or before running it against a real codebase.

The script can be run from anywhere: it resolves the repository root from its own
location, so the git commands and the Docker build context always resolve
correctly.

### What it does

1. Checks out the requested branch (defaults to the current one).
2. Builds the production `Dockerfile`, injecting the version, commit hash and
   build timestamp as build args (so `guacamole version` reports them).
3. Detects the host's `terraform` / `terragrunt` versions and installs the same
   ones in the image, so the `state` and `profile` modes work inside it.
4. Bakes the repo's `example/` and `tests/` folders into the image at `/example`
   and `/tests` (test-image only — **not** shipped in the production image), so
   you can run the checks out of the box.
5. Prints the built version, then starts an interactive shell.

### Usage

```bash
./scripts/docker-test-version.sh [-b <branch>] [-r <repo-path>] [-e <env-file>]
```

| Flag | Description                                                                        |
| ---- | ---------------------------------------------------------------------------------- |
| `-b` | Git branch to build (default: current branch).                                     |
| `-r` | Path to an external repository to mount in the container at `/repo`.               |
| `-e` | Path to an env file (`KEY=VALUE` per line) loaded into the container.              |
| `-h` | Show usage.                                                                        |

### Use cases

**Try the checks on the bundled fixtures (no external repo needed):**

```bash
./scripts/docker-test-version.sh
# then, inside the container:
guacamole static module -p /tests/modules/pass   # expect all ✅
guacamole static module -p /tests/modules/fail -v # expect targeted ❌
guacamole static -p /example
```

**Test a feature branch against one of your own IaC repositories:**

```bash
./scripts/docker-test-version.sh -b feat/my-new-check -r ~/projects/my-infra-repo
# then, inside the container:
guacamole static -p /repo
```

**Test the `state` / `profile` modes, which need cloud credentials:**

```bash
./scripts/docker-test-version.sh -r ~/projects/my-infra-repo -e ~/projects/my-infra-repo/.env
# then, inside the container:
guacamole state   -p /repo
guacamole profile -p /repo
```

### Requirements

- Docker
- `terraform` and `terragrunt` on the host are optional; if missing, the image is
  still built but the `state` and `profile` modes will not work inside it.

### Notes

- The image is tagged `guacamole-local:<branch>` (slashes in the branch name are
  replaced with `-`).
- Mounting an external repo (`-r`) is read via a bind mount, so the container
  sees your live files.
- `state` and `profile` run Terraform/Terragrunt and therefore need provider
  credentials — pass them with `-e`.
