#!/usr/bin/env bash
set -euo pipefail

coverage_file="${1:-coverage.out}"

if [[ ! -f "$coverage_file" ]]; then
  echo "Coverage file not found: $coverage_file" >&2
  exit 1
fi

total="$(go tool cover -func="$coverage_file" | awk '/^total:/{print $NF}')"

{
  echo "## Test coverage"
  echo ""
  echo "| Package | Coverage |"
  echo "| --- | ---: |"

  awk '
    /^mode:/ { next }
    NF < 3 { next }
    {
      split($1, parts, ":")
      file = parts[1]
      stmts = $2 + 0
      covered_block = ($3 + 0 > 0) ? stmts : 0

      if (match(file, /\/(internal|cmd|pkg)\//)) {
        pkg = substr(file, RSTART + 1)
        sub(/\/[^\/]+\.go$/, "", pkg)
      } else {
        pkg = file
        sub(/.*\//, "", pkg)
        sub(/\.go$/, "", pkg)
        pkg = "other/" pkg
      }

      total[pkg] += stmts
      hit[pkg] += covered_block
    }
    END {
      for (p in total) {
        if (total[p] > 0) {
          pct = 100 * hit[p] / total[p]
          printf "%010.2f|%s|%.1f\n", pct, p, pct
        }
      }
    }
  ' "$coverage_file" | sort -rn | while IFS='|' read -r _ pkg pct; do
    echo "| \`${pkg}\` | ${pct}% |"
  done

  echo ""
  echo "| **Total** | **${total}** |"
} >> "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY is not set}"
