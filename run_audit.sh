#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${ROOT}"

if ! command -v go >/dev/null 2>&1; then
  echo "[ERROR] Go is not installed — install Go 1.27+" >&2
  exit 1
fi

if command -v govulncheck >/dev/null 2>&1; then
  govulncheck ./...
else
  go run golang.org/x/vuln/cmd/govulncheck@latest ./...
fi

TRIVY_IMAGE="${TRIVY_IMAGE:-lantern:local}"
if command -v trivy >/dev/null 2>&1; then
  if docker image inspect "${TRIVY_IMAGE}" >/dev/null 2>&1; then
    if ! command -v jq >/dev/null 2>&1; then
      echo "[ERROR] jq is required for Trivy summary output" >&2
      exit 1
    fi
    TRIVY_JSON="$(mktemp)"
    trap 'rm -f "${TRIVY_JSON}"' EXIT
    trivy image \
      --format json \
      --exit-code 0 \
      --ignore-unfixed \
      --severity HIGH,CRITICAL \
      "${TRIVY_IMAGE}" \
      > "${TRIVY_JSON}"

    echo
    echo "Trivy vulnerabilities:"
    printf "%-20s %-10s %-12s %-12s\n" \
      "Library" "Severity" "Installed" "Fixed"

    jq -r '
      .Results[]?.Vulnerabilities[]? |
      [
        .PkgName,
        .Severity,
        .InstalledVersion,
        (.FixedVersion // "-")
      ] |
      @tsv
    ' "${TRIVY_JSON}" |
    while IFS=$'\t' read -r library severity installed fixed; do
      printf "%-20s %-10s %-12s %-12s\n" \
        "$library" "$severity" "$installed" "$fixed"
    done

    if jq -e '[.Results[]?.Vulnerabilities[]?] | length > 0' "${TRIVY_JSON}" >/dev/null; then
      echo
      echo "[TRIVY] Found HIGH/CRITICAL vulnerabilities." >&2
      exit 1
    fi
  else
    echo "[TRIVY] Skipping — image ${TRIVY_IMAGE} not found (build with: ./deploy/build.sh prod)"
  fi
else
  echo "[TRIVY] Skipping — trivy not installed"
fi

echo
echo "Security audit completed successfully."
