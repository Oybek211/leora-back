#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:9090/api/v1}"
EMAIL="${LEORA_EMAIL:-test@leora.app}"
PASSWORD="${LEORA_PASSWORD:-pass1234}"
ITERATIONS="${ITERATIONS:-10}"

json_field() {
  local field="$1"
  python3 - "$field" <<'PY'
import json
import sys

field = sys.argv[1]
raw = sys.stdin.read()
data = json.loads(raw or "{}")
value = data
for part in field.split("."):
    if isinstance(value, dict) and part in value:
        value = value[part]
    else:
        value = None
        break
if value is None:
    sys.exit(2)
if isinstance(value, (dict, list)):
    print(json.dumps(value))
else:
    print(value)
PY
}

echo "Auth cycle smoke test: ${ITERATIONS} iterations"
echo "Base URL: ${BASE_URL}"
echo "Email: ${EMAIL}"

for i in $(seq 1 "${ITERATIONS}"); do
  echo "---- Iteration ${i} ----"

  login_resp="$(curl -s -X POST "${BASE_URL}/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"emailOrUsername\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")"

  login_success="$(printf '%s' "${login_resp}" | json_field success || true)"
  if [ "${login_success}" != "true" ]; then
    echo "Login failed: ${login_resp}"
    exit 1
  fi

  access_token="$(printf '%s' "${login_resp}" | json_field data.accessToken)"
  if [ -z "${access_token}" ]; then
    echo "Missing access token: ${login_resp}"
    exit 1
  fi

  me_resp="$(curl -s -X GET "${BASE_URL}/auth/me" \
    -H "Authorization: Bearer ${access_token}")"

  me_success="$(printf '%s' "${me_resp}" | json_field success || true)"
  if [ "${me_success}" != "true" ]; then
    echo "GET /auth/me failed: ${me_resp}"
    exit 1
  fi

  logout_resp="$(curl -s -X POST "${BASE_URL}/auth/logout" \
    -H "Authorization: Bearer ${access_token}")"

  logout_success="$(printf '%s' "${logout_resp}" | json_field success || true)"
  if [ "${logout_success}" != "true" ]; then
    echo "Logout failed: ${logout_resp}"
    exit 1
  fi
done

echo "Auth cycle smoke test passed."
