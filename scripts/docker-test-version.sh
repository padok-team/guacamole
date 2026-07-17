#!/usr/bin/env bash
set -euo pipefail

# Resolve the repository root so the script works from any working directory:
# the git commands, the Docker build context (`.`) and the COPY paths below all
# assume they run from the root of the guacamole repository.
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

usage() {
  echo "Usage: $0 [-b <branch>] [-r <repo-path>] [-e <env-file>]"
  echo ""
  echo "  -b  Git branch to build (default: current branch)"
  echo "  -r  Absolute or relative path to a repository to mount in the container at /repo"
  echo "  -e  Path to an env file to load into the container (KEY=VALUE, one per line)"
  echo ""
  echo "Examples:"
  echo "  $0 -b main"
  echo "  $0 -b feat/my-feature -r ~/projects/my-infra-repo"
  echo "  $0 -r ~/projects/my-infra-repo -e ~/projects/my-infra-repo/.env"
  exit 1
}

BRANCH=""
REPO_PATH=""
ENV_FILE=""

while getopts "b:r:e:h" opt; do
  case $opt in
    b) BRANCH="$OPTARG" ;;
    r) REPO_PATH="$OPTARG" ;;
    e) ENV_FILE="$OPTARG" ;;
    h) usage ;;
    *) usage ;;
  esac
done

# Default to current branch
if [[ -z "$BRANCH" ]]; then
  BRANCH=$(git rev-parse --abbrev-ref HEAD)
fi

echo "Checking out branch '$BRANCH'..."
git checkout "$BRANCH"

VERSION=$(git rev-parse --abbrev-ref HEAD | tr '/' '-')
COMMIT_HASH=$(git rev-parse HEAD)
BUILD_TIMESTAMP=$(date +%s)
IMAGE="guacamole-local:$VERSION"

TERRAGRUNT_VERSION=""
if command -v terragrunt &>/dev/null; then
  TERRAGRUNT_VERSION=$(terragrunt --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
  echo "Detected terragrunt version: $TERRAGRUNT_VERSION"
else
  echo "Warning: terragrunt not found on host — 'guacamole state' and 'guacamole profile' will not work inside the container"
fi

TERRAFORM_VERSION=""
if command -v terraform &>/dev/null; then
  TERRAFORM_VERSION=$(terraform version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
  echo "Detected terraform version: $TERRAFORM_VERSION"
else
  echo "Warning: terraform not found on host — 'guacamole state' and 'guacamole profile' will not work inside the container"
fi

BUILD_ARGS=(
  --build-arg VERSION="$VERSION"
  --build-arg COMMIT_HASH="$COMMIT_HASH"
  --build-arg BUILD_TIMESTAMP="$BUILD_TIMESTAMP"
)
[[ -n "$TERRAGRUNT_VERSION" ]] && BUILD_ARGS+=(--build-arg TERRAGRUNT_VERSION="$TERRAGRUNT_VERSION")
[[ -n "$TERRAFORM_VERSION" ]] && BUILD_ARGS+=(--build-arg TERRAFORM_VERSION="$TERRAFORM_VERSION")

echo "Building image $IMAGE..."
docker build "${BUILD_ARGS[@]}" -t "$IMAGE" .

# Bake the example codebase and the test fixtures into the local test image so
# guacamole can be tried out of the box (e.g. on /example or /tests) without
# mounting an external repo. This is only added to the local test image, not to
# the production Dockerfile.
echo "Adding example codebase and test fixtures to image (/example, /tests)..."
docker build -t "$IMAGE" -f - . <<EOF
FROM $IMAGE
COPY example /example
COPY tests /tests
EOF

echo ""
echo "Built version:"
docker run --rm "$IMAGE" version

DOCKER_ARGS=(--rm -it --entrypoint /bin/sh)

if [[ -n "$ENV_FILE" ]]; then
  ABS_ENV_FILE=$(realpath "$ENV_FILE")
  if [[ ! -f "$ABS_ENV_FILE" ]]; then
    echo "Error: env file '$ABS_ENV_FILE' does not exist"
    exit 1
  fi
  DOCKER_ARGS+=(--env-file "$ABS_ENV_FILE")
  echo "Loading env vars from: $ABS_ENV_FILE"
fi

echo ""
echo "The example codebase is baked into the image at /example. Try:"
echo "  guacamole static -p /example"
echo "  guacamole static module -p /example"
echo ""
echo "The test fixtures are baked in at /tests (pass/ must be all ✅, fail/ all ❌):"
echo "  guacamole static module -p /tests/modules/pass"
echo "  guacamole static module -p /tests/modules/fail -v"
echo "  guacamole static layer  -p /tests/layers/pass"
echo "  guacamole static layer  -p /tests/layers/fail -v"

if [[ -n "$REPO_PATH" ]]; then
  ABS_REPO_PATH=$(realpath "$REPO_PATH")
  if [[ ! -d "$ABS_REPO_PATH" ]]; then
    echo "Error: repo path '$ABS_REPO_PATH' does not exist or is not a directory"
    exit 1
  fi
  DOCKER_ARGS+=(-v "$ABS_REPO_PATH:/repo")
  echo ""
  echo "Mounting repo: $ABS_REPO_PATH -> /repo"
  echo ""
  echo "Once inside the container, try:"
  echo "  guacamole static -p /repo"
  echo "  guacamole state -p /repo"
  echo "  guacamole profile -p /repo"
fi

echo ""
echo "Starting interactive shell (guacamole is in PATH)..."
docker run "${DOCKER_ARGS[@]}" "$IMAGE"
