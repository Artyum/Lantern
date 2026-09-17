#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${ROOT}"

if ! command -v go >/dev/null 2>&1; then
  echo "[ERROR] Go is not installed — install Go 1.27+" >&2
  exit 1
fi

command -v docker >/dev/null || {
  echo "docker command not found." >&2
  exit 1
}

command -v trivy >/dev/null || {
  echo "trivy command not found." >&2
  exit 1
}

command -v jq >/dev/null || {
  echo "jq command not found." >&2
  exit 1
}

if command -v govulncheck >/dev/null 2>&1; then
  govulncheck ./...
else
  go run golang.org/x/vuln/cmd/govulncheck@latest ./...
fi

TRIVY_IMAGE="${TRIVY_IMAGE:-lantern:local}"

docker build \
  --file Dockerfile \
  --tag "${TRIVY_IMAGE}" \
  .

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

if jq -e '
  [.Results[]?.Vulnerabilities[]?] | length > 0
' "${TRIVY_JSON}" >/dev/null; then
  echo
  echo "[TRIVY] Found HIGH/CRITICAL vulnerabilities." >&2
  exit 1
fi

echo
echo "Security audit completed successfully."
