#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${ROOT}"

mkdir -p report
LOG="${ROOT}/report/test.log"

if ! command -v go >/dev/null 2>&1; then
  echo "[ERROR] Go is not installed — install Go 1.27+" >&2
  exit 1
fi

{
  echo "========================================"
  echo " GO TEST - $(date '+%Y-%m-%d %H:%M:%S')"
  echo "========================================"
  echo ""
} > "${LOG}"

set +e
go test ./... -count=1 2>&1 | tee -a "${LOG}"
status=${PIPESTATUS[0]}
set -e

echo ""
echo "  Log: ${LOG}"
echo ""

exit "${status}"
