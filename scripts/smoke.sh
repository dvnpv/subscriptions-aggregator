#!/usr/bin/env bash

set -euo pipefail

BASE_URL="http://localhost:8080"
USER_ID="60601fee-2bf1-4721-ae6f-7636e79a0cba"

echo "==> 1. Health check"
curl -fsS "${BASE_URL}/health"
echo
echo "OK: health"

echo "==> 2. Create subscription"
CREATE_RESPONSE=$(curl -fsS -X POST "${BASE_URL}/subscriptions/" \
  -H "Content-Type: application/json" \
  -d "{
    \"service_name\": \"Netflix\",
    \"price\": 999,
    \"user_id\": \"${USER_ID}\",
    \"start_date\": \"01-2026\",
    \"end_date\": \"03-2026\"
  }")

echo "${CREATE_RESPONSE}"

ID=$(echo "${CREATE_RESPONSE}" | grep -o '"id":"[^"]*"' | head -1 | cut -d':' -f2 | tr -d '"')

if [[ -z "${ID}" ]]; then
  echo "ERROR: failed to extract id from create response"
  exit 1
fi

echo "OK: created subscription with id=${ID}"

echo "==> 3. Get subscription by id"
curl -fsS "${BASE_URL}/subscriptions/${ID}"
echo
echo "OK: get by id"

echo "==> 4. List subscriptions"
curl -fsS "${BASE_URL}/subscriptions/"
echo
echo "OK: list"

echo "==> 5. Calculate total"
curl -fsS "${BASE_URL}/subscriptions/total?from=01-2026&to=03-2026"
echo
echo "OK: total"

echo "==> 6. Update subscription"
curl -fsS -X PUT "${BASE_URL}/subscriptions/${ID}" \
  -H "Content-Type: application/json" \
  -d "{
    \"service_name\": \"Netflix Premium\",
    \"price\": 1299,
    \"user_id\": \"${USER_ID}\",
    \"start_date\": \"01-2026\",
    \"end_date\": \"04-2026\"
  }"
echo
echo "OK: update"

echo "==> 7. Delete subscription"
DELETE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "${BASE_URL}/subscriptions/${ID}")

if [[ "${DELETE_STATUS}" != "204" ]]; then
  echo "ERROR: expected 204 on delete, got ${DELETE_STATUS}"
  exit 1
fi

echo "OK: delete"

echo "==> 8. Verify deleted subscription returns 404"
GET_AFTER_DELETE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/subscriptions/${ID}")

if [[ "${GET_AFTER_DELETE_STATUS}" != "404" ]]; then
  echo "ERROR: expected 404 after delete, got ${GET_AFTER_DELETE_STATUS}"
  exit 1
fi

echo "OK: deleted entity is not found"
echo
echo "Smoke test passed successfully."