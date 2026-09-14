#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY="${ROOT}/deploy"
ENV_FILE="${DEPLOY}/.env.dev"
COMPOSE="${DEPLOY}/docker-compose.dev.yml"
COMPOSE_CMD=(docker compose --env-file "${ENV_FILE}" -p lantern-dev -f "${COMPOSE}")

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "[ERROR] Missing ${ENV_FILE} — copy deploy/.env.dev.example" >&2
  exit 1
fi
if ! docker network inspect traefik-net >/dev/null 2>&1; then
  echo "[ERROR] Docker network traefik-net not found" >&2
  exit 1
fi

running="$("${COMPOSE_CMD[@]}" ps -q --status running lantern_dev || true)"
if [[ -n "${running}" ]]; then
  echo "[OK] already running"
else
  "${COMPOSE_CMD[@]}" up -d --no-recreate
fi
