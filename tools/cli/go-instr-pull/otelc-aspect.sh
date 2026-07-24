#!/usr/bin/env bash
# Fetch otelc instrumentation details for a specific Go import path.
# Outputs relevant details + build command — keeps agent context lean.
#
# Usage: otelc-aspect.sh <import-path>
# Example: otelc-aspect.sh net/http
#          otelc-aspect.sh github.com/gin-gonic/gin
#          otelc-aspect.sh google.golang.org/grpc
set -euo pipefail

readonly OTELC_REPO="open-telemetry/opentelemetry-go-compile-instrumentation"
readonly OTELC_VERSION="v1.0.1"
readonly GUIDE_URL="https://raw.githubusercontent.com/${OTELC_REPO}/main/docs/instrument-guide.md"
readonly RULES_URL="https://github.com/${OTELC_REPO}/tree/main/instrumentation"

usage() {
  echo "Usage: $(basename "$0") <import-path>" >&2
  echo "Example: $(basename "$0") net/http" >&2
  exit 1
}

[[ $# -lt 1 ]] && usage

import_path="${1}"

if ! command -v curl &>/dev/null; then
  echo "ERROR: curl is required" >&2
  exit 1
fi

echo "# otelc ${OTELC_VERSION} — import: ${import_path}"
echo "# Requires: Go 1.25+"
echo

# Try to find mentions in the instrumentation guide.
guide=$(curl -fsSL "${GUIDE_URL}" 2>/dev/null) || guide=""

if [[ -n "${guide}" ]]; then
  matched=$(echo "${guide}" | grep -i "${import_path}" | head -20 || true)
  if [[ -n "${matched}" ]]; then
    echo "## Found in instrument-guide.md"
    echo "${matched}"
    echo
  fi
fi

# If nothing found, note that otelc instruments silently — no explicit config needed.
if [[ -z "${matched:-}" ]]; then
  echo "No explicit aspect entry found for '${import_path}'."
  echo
  echo "otelc instruments supported libraries automatically during build — no per-library"
  echo "configuration required. If your library is supported, spans appear automatically."
  echo
  echo "Full instrumentation rules: ${RULES_URL}"
fi

cat <<EOF
## Build with otelc

  # Install (requires Go 1.25+)
  go install go.opentelemetry.io/otelc/tool/cmd/otelc@${OTELC_VERSION}

  # Build (replaces: go build)
  otelc go build -o ./myapp ./...

  # Run (with stdout exporter for quick local check)
  OTEL_SERVICE_NAME=my-service \\
  OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 \\
  ./myapp

## Verify instrumentation is active

  otelc go build -v ./... 2>&1 | grep "injecting"

## Local trace UI (Jaeger all-in-one)

  docker run -d --name jaeger \\
    -p 16686:16686 -p 4317:4317 \\
    jaegertracing/all-in-one:latest
  open http://localhost:16686
EOF
