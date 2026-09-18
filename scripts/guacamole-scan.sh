#!/usr/bin/env sh
set -e

: "${GUACAMOLE_WORKDIR:?GUACAMOLE_WORKDIR must be set}"

LAYER_DIRS_FILE="${GUACAMOLE_WORKDIR}/layer_dirs"
MODULE_DIRS_FILE="${GUACAMOLE_WORKDIR}/module_dirs"
RESULTS_FILE="${GUACAMOLE_WORKDIR}/results"

OVERALL_PASS=0
OVERALL_TOTAL=0
GUACAMOLE_EXIT=0

run_scan() {
  scan_type="$1"
  rel_dir="$2"
  abs_dir="${CI_PROJECT_DIR}/${rel_dir}"

  if [ ! -d "${abs_dir}" ]; then
    echo "Skipping missing directory: ${abs_dir}"
    return
  fi

  output=$(guacamole static "${scan_type}" -v -p "${abs_dir}" 2>&1) || GUACAMOLE_EXIT=1
  echo "${output}"

  score_pct=$(echo "${output}" | sed -n 's/^Score: \([0-9][0-9]*%\).*$/\1/p' | tail -n1)
  ratio=$(echo "${output}" | sed -n 's/^Score: [0-9][0-9]*%[^()]*(\([0-9][0-9]*\)\/\([0-9][0-9]*\)).*$/\1 \2/p' | tail -n1)

  if [ -n "${ratio}" ]; then
    pass=$(echo "${ratio}" | awk '{print $1}')
    total=$(echo "${ratio}" | awk '{print $2}')
    OVERALL_PASS=$((OVERALL_PASS + pass))
    OVERALL_TOTAL=$((OVERALL_TOTAL + total))
  fi

  failing=$(echo "${output}" | sed -n 's/^❌[[:space:]]*/❌ /p' | awk 'BEGIN { ORS="" } { if (NR > 1) printf " <br> "; printf "%s", $0 }')
  if [ -z "${failing}" ]; then
    failing="-"
  fi
  printf '%s|%s|%s|%s\n' "${scan_type}" "${rel_dir}" "${score_pct:-n/a}" "${failing}" >> "${RESULTS_FILE}"
}

while IFS= read -r dir; do
  [ -z "${dir}" ] && continue
  run_scan "layer" "${dir}"
done < "${LAYER_DIRS_FILE}"

while IFS= read -r dir; do
  [ -z "${dir}" ] && continue
  run_scan "module" "${dir}"
done < "${MODULE_DIRS_FILE}"

if [ "${OVERALL_TOTAL}" -gt 0 ]; then
  OVERALL_SCORE=$((OVERALL_PASS * 100 / OVERALL_TOTAL))
else
  OVERALL_SCORE=0
fi
printf '%d %d %d\n' "${OVERALL_SCORE}" "${OVERALL_PASS}" "${OVERALL_TOTAL}" > "${GUACAMOLE_WORKDIR}/score"
printf '%d\n' "${GUACAMOLE_EXIT}" > "${GUACAMOLE_WORKDIR}/scan_exit"
