#!/usr/bin/env sh
set -e

: "${GUACAMOLE_WORKDIR:?GUACAMOLE_WORKDIR must be set}"
: "${GUACAMOLE_GITLAB_TOKEN:?GUACAMOLE_GITLAB_TOKEN must be set}"

RESULTS_FILE="${GUACAMOLE_WORKDIR}/results"
SCORE_FILE="${GUACAMOLE_WORKDIR}/score"

if [ -z "${CI_MERGE_REQUEST_IID:-}" ]; then
  exit 0
fi

COMMENT=$(mktemp)
{
  echo "### 🥑 Guacamole static checks"
  if [ -s "${RESULTS_FILE}" ]; then
    OVERALL_SCORE=$(awk '{print $1}' "${SCORE_FILE}")
    OVERALL_PASS=$(awk '{print $2}' "${SCORE_FILE}")
    OVERALL_TOTAL=$(awk '{print $3}' "${SCORE_FILE}")
    if [ "${OVERALL_SCORE}" -eq 100 ]; then
      SCORE_EMOJI="🎉"
    else
      SCORE_EMOJI="🚧"
    fi
    echo
    echo "- Global score: ${SCORE_EMOJI} ${OVERALL_SCORE}% (${OVERALL_PASS}/${OVERALL_TOTAL})"
    echo
    echo "| Scope | Path | Score | Failed rules |"
    echo "|---|---|---|---|"
    while IFS='|' read -r scope path score failing; do
      printf '| %s | `%s` | %s | %s |\n' "${scope}" "${path}" "${score}" "${failing}"
    done < "${RESULTS_FILE}"
  else
    echo
    echo 'No modified layers/modules detected under `layers/`, `base/` or `functional/`.'
  fi
} > "${COMMENT}"

BODY=$(awk 'BEGIN{ORS=""}{gsub(/%/,"%25");gsub(/&/,"%26");gsub(/\+/,"%2B");gsub(/\r/,"%0D");printf "%s%%0A",$0}' "${COMMENT}")
API_URL="${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/merge_requests/${CI_MERGE_REQUEST_IID}/notes"

AUTH_HEADER="PRIVATE-TOKEN: ${GUACAMOLE_GITLAB_TOKEN}"

wget -qO- --header="${AUTH_HEADER}" --post-data="body=${BODY}" "${API_URL}" >/dev/null
