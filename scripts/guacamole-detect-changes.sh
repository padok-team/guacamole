#!/usr/bin/env sh
set -e

: "${GUACAMOLE_WORKDIR:?GUACAMOLE_WORKDIR must be set}"

BASE_BRANCH="${GUACAMOLE_DIFF_BASE_BRANCH:-${CI_MERGE_REQUEST_TARGET_BRANCH_NAME:-}}"
if [ -z "${BASE_BRANCH}" ]; then
  echo "Either GUACAMOLE_DIFF_BASE_BRANCH or CI_MERGE_REQUEST_TARGET_BRANCH_NAME must be set"
  exit 1
fi

TARGET_BRANCH="origin/${BASE_BRANCH}"
if ! git rev-parse --verify "${TARGET_BRANCH}" >/dev/null 2>&1; then
  git fetch origin "${BASE_BRANCH}"
fi

MR_SHA="${GUACAMOLE_MR_SHA:-HEAD}"
echo "Base branch : ${BASE_BRANCH}"
echo "Target ref  : ${TARGET_BRANCH}"
echo "MR SHA      : ${MR_SHA}"
CHANGED_FILES=$(git diff --name-only "${TARGET_BRANCH}...${MR_SHA}")
echo "Changed files:"
echo "${CHANGED_FILES}"

LAYER_DIRS_FILE="${GUACAMOLE_WORKDIR}/layer_dirs"
MODULE_DIRS_FILE="${GUACAMOLE_WORKDIR}/module_dirs"

echo "${CHANGED_FILES}" | grep '^layers/' | while IFS= read -r file; do
  dirname "${file}"
done | sort -u > "${LAYER_DIRS_FILE}"

echo "${CHANGED_FILES}" | grep -E '^(base|functional)/' | while IFS= read -r file; do
  dirname "${file}"
done | sort -u > "${MODULE_DIRS_FILE}"

echo "Changed layers:"
cat "${LAYER_DIRS_FILE}" || true
echo "Changed modules:"
cat "${MODULE_DIRS_FILE}" || true
