#!/usr/bin/env bash
# Fetch OBI integration details for a specific Go library from SUPPORT_MATRIX.md.
# Outputs only the relevant section — keeps agent context lean.
#
# Usage: obi-integration.sh <library-name>
# Example: obi-integration.sh net/http
#          obi-integration.sh gin
#          obi-integration.sh grpc
set -euo pipefail

readonly SUPPORT_MATRIX_URL="https://raw.githubusercontent.com/open-telemetry/opentelemetry-ebpf-instrumentation/main/SUPPORT_MATRIX.md"
readonly OBI_VERSION="v0.10.0"

usage() {
  echo "Usage: $(basename "$0") <library-name>" >&2
  echo "Example: $(basename "$0") net/http" >&2
  exit 1
}

[[ $# -lt 1 ]] && usage

library="${1}"

if ! command -v curl &>/dev/null; then
  echo "ERROR: curl is required" >&2
  exit 1
fi

echo "# OBI ${OBI_VERSION} — integration: ${library}"
echo "# Source: ${SUPPORT_MATRIX_URL}"
echo

# Fetch the matrix and extract rows matching the library name (case-insensitive).
# Outputs: the table header lines + any matching rows.
content=$(curl -fsSL "${SUPPORT_MATRIX_URL}" 2>/dev/null) || {
  echo "ERROR: failed to fetch SUPPORT_MATRIX.md from GitHub" >&2
  exit 1
}

# Find the section containing Go library support and extract matching rows.
# The matrix uses markdown tables; we grab header + matching data rows.
matched=$(echo "${content}" | awk -v lib="${library}" '
  BEGIN { in_go=0; header=""; printed_header=0 }
  /Go Library/ || /## Go/ { in_go=1 }
  /^## / && !/Go/ { in_go=0 }
  in_go && /^\|/ {
    if ($0 ~ /^[|][-| ]+[|]/) {
      # separator row — skip
      next
    }
    if ($0 ~ /Library|Framework|Package/) {
      header=$0
      printed_header=0
      next
    }
    if (tolower($0) ~ tolower(lib)) {
      if (!printed_header && header != "") {
        print "| Library | Min Version | Notes |"
        print "|---------|-------------|-------|"
        printed_header=1
      }
      print $0
    }
  }
')

if [[ -z "${matched}" ]]; then
  echo "No OBI integration found for '${library}'."
  echo
  echo "Check the full matrix: ${SUPPORT_MATRIX_URL}"
  echo "Supported libraries include: net/http, gin, gRPC, gorilla/mux, go-redis, Kafka, database/sql"
else
  echo "${matched}"
  echo
  echo "Note: 'zero code changes' holds for standard RED metrics."
  echo "Custom spans, business-logic events, SQL query details still require manual instrumentation."
fi
