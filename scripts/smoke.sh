#!/usr/bin/env bash
set -euo pipefail

BACKEND_URL="${BACKEND_URL:-https://praxis-backend.orangepebble-10f42efc.southafricanorth.azurecontainerapps.io}"
WEB_URL="${WEB_URL:-https://praxis-web.orangepebble-10f42efc.southafricanorth.azurecontainerapps.io}"
AI_URL="${AI_URL:-https://praxis-ai.orangepebble-10f42efc.southafricanorth.azurecontainerapps.io}"

RETRIES="${RETRIES:-3}"
TIMEOUT="${TIMEOUT:-30}"

pass=0
fail=0

fetch() {
  local url=$1
  shift
  local attempt=1
  while [ "$attempt" -le "$RETRIES" ]; do
    if out=$(curl -sS --max-time "$TIMEOUT" -w '\n%{http_code}' "$@" "$url" 2>/dev/null); then
      code=$(printf '%s' "$out" | tail -n1)
      body=$(printf '%s' "$out" | sed '$d')
      printf '%s %s' "$code" "$body"
      return 0
    fi
    attempt=$((attempt + 1))
    sleep 5
  done
  printf '000 '
  return 0
}

check_status() {
  local name=$1 expected=$2 url=$3
  shift 3
  local result code
  result=$(fetch "$url" "$@")
  code=${result%% *}
  if [ "$code" = "$expected" ]; then
    echo "PASS $name ($code)"
    pass=$((pass + 1))
  else
    echo "FAIL $name (expected $expected, got $code)"
    fail=$((fail + 1))
  fi
}

check_body() {
  local name=$1 expected=$2 needle=$3 url=$4
  shift 4
  local result code body
  result=$(fetch "$url" "$@")
  code=${result%% *}
  body=${result#* }
  if [ "$code" = "$expected" ] && printf '%s' "$body" | grep -q "$needle"; then
    echo "PASS $name ($code, contains \"$needle\")"
    pass=$((pass + 1))
  else
    echo "FAIL $name (expected $expected with \"$needle\", got $code)"
    fail=$((fail + 1))
  fi
}

echo "Smoke test"
echo "  backend: $BACKEND_URL"
echo "  web:     $WEB_URL"
echo "  ai:      $AI_URL"
echo

check_body "backend health" 200 '"status":"ok"' "$BACKEND_URL/health"
check_status "backend requires auth on /v1/me" 401 "$BACKEND_URL/v1/me"
check_body "web renders backend status" 200 'Backend:' "$WEB_URL/"
check_body "ai health" 200 '"status":"ok"' "$AI_URL/health"
check_status "ai rejects unsigned internal call" 401 "$AI_URL/internal/v1/diagnostics/score" \
  -X POST -H 'Content-Type: application/json' -d '{}'

echo
echo "Passed: $pass  Failed: $fail"
[ "$fail" -eq 0 ]
