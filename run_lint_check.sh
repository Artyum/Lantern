#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${ROOT}"

LOG="${ROOT}/report/lint_check.log"
TOTAL=4
mkdir -p report

if ! command -v go >/dev/null 2>&1; then
  echo "[ERROR] Go is not installed — install Go 1.27+" >&2
  exit 1
fi

echo ""
echo "============ LINT CHECK (Lantern) ============"
echo ""

: > "${LOG}"
{
  echo "========================================"
  echo " LINT CHECK - $(date '+%Y-%m-%d %H:%M:%S')"
  echo "========================================"
  echo ""
} >> "${LOG}"

results=()

run_step() {
  local n=$1
  shift
  local title=$1
  shift
  echo "[Step $n/$TOTAL] $title..."
  {
    echo "--- Step $n/$TOTAL: $title ---"
    date '+%Y-%m-%d %H:%M:%S'
    echo ""
  } >> "${LOG}"
  if "$@" >> "${LOG}" 2>&1; then
    results+=("0")
  else
    results+=("1")
  fi
  {
    echo ""
    echo "========================================"
    echo ""
  } >> "${LOG}"
}

run_gofmt_write() {
  gofmt -w .
}

run_gofmt_check() {
  local leftover
  leftover="$(gofmt -l .)"
  if [[ -n "${leftover}" ]]; then
    echo "${leftover}"
    return 1
  fi
}

run_step 1 "gofmt -w" run_gofmt_write
run_step 2 "gofmt -l" run_gofmt_check
run_step 3 "go vet" go vet ./...
run_step 4 "go test" go test ./... -count=1

echo ""
echo "============ SUMMARY ============"
echo ""

any_fail=0
for i in "${!results[@]}"; do
  step_no=$((i + 1))
  if [[ "${results[$i]}" -eq 0 ]]; then
    echo "  [$step_no] PASS"
  else
    echo "  [$step_no] FAIL"
    any_fail=1
  fi
done

echo ""
echo "  Log: ${LOG}"
echo ""

{
  echo ""
  echo "========================================"
  echo " SUMMARY"
  echo "========================================"
  echo ""
} >> "${LOG}"

for i in "${!results[@]}"; do
  step_no=$((i + 1))
  if [[ "${results[$i]}" -eq 0 ]]; then
    echo "  [$step_no] PASS" >> "${LOG}"
  else
    echo "  [$step_no] FAIL" >> "${LOG}"
  fi
done

if [[ "${any_fail}" -ne 0 ]]; then
  echo "[ERROR] Some steps failed — check the log."
  exit 1
fi

echo "[OK] All steps passed."
