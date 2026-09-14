#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
COMPOSE="${SCRIPT_DIR}/docker-compose.dev.yml"
ENV_FILE="${SCRIPT_DIR}/.env.dev"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "[ERROR] Missing ${ENV_FILE} — copy deploy/.env.dev.example" >&2
  exit 1
fi
if ! command -v docker >/dev/null 2>&1; then
  echo "[ERROR] Docker Engine with Compose v2 is required" >&2
  exit 1
fi

ACTION="${1:-}"

case "${ACTION}" in
  "" )
    BUILD_ARGS=()
    [[ "${NO_CACHE:-}" == "1" ]] && BUILD_ARGS+=(--no-cache)
    docker compose --env-file "${ENV_FILE}" -p lantern-dev -f "${COMPOSE}" build "${BUILD_ARGS[@]}"
    echo "[OK] lantern:dev — run: ./start_app.sh"
    ;;
  prod|release)
    echo "==> Building lantern:local (prod)"
    docker build -t lantern:local -f "${ROOT}/Dockerfile" "${ROOT}"
    echo "[OK] lantern:local"
    docker images lantern:local --format "    size: {{.Size}}"
    ;;
  *)
    echo "Usage: $0 [prod|release]" >&2
    exit 1
    ;;
esac
