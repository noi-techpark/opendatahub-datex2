#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
#
# SPDX-License-Identifier: CC0-1.0

# End-to-end integration test: builds the production Docker image, runs it
# against the real Open Data Hub API using the bundled config.example.yaml,
# and checks that it actually serves a DATEX II publication plus the API
# docs. Requires network access to reach the Open Data Hub API, and xmllint
# (libxml2) to validate the publication against the DATEX II XSD.

set -euo pipefail

command -v xmllint >/dev/null || { echo "xmllint not found (install libxml2-utils/libxml2)"; exit 1; }

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCHEMA="$TEST_DIR/schema/DATEXIISchema_2_2_3.xsd"
BASE_URL="http://localhost:8090"
PROVIDER_PATH="/datex/2/province-bz/situation-publication.xml"

cleanup() {
  echo "==> Tearing down test stack"
  (cd "$TEST_DIR" && docker compose down -v) || true
}
trap cleanup EXIT

echo "==> Building and starting the service"
(cd "$TEST_DIR" && docker compose up --build -d)

echo "==> Waiting for the first publish cycle"
published=false
for _ in $(seq 1 30); do
  if (cd "$TEST_DIR" && docker compose logs app) | grep -q "published .* events to 1 recipients"; then
    published=true
    break
  fi
  sleep 1
done
if [ "$published" != true ]; then
  echo "FAIL: no publish cycle observed within 30s"
  (cd "$TEST_DIR" && docker compose logs app)
  exit 1
fi

check_status() {
  local path="$1" want="$2" got
  got="$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL$path")"
  if [ "$got" != "$want" ]; then
    echo "FAIL: GET $path returned $got, want $want"
    exit 1
  fi
  echo "OK: GET $path -> $got"
}

check_status "/" 200
check_status "/openapi.yaml" 200
check_status "/datex/2/" 200
check_status "$PROVIDER_PATH" 200
check_status "/not-a-configured-path" 404

echo "==> Checking the provider list includes the configured feed"
providers="$(curl -s "$BASE_URL/datex/2/")"
if ! grep -q "$PROVIDER_PATH" <<<"$providers"; then
  echo "FAIL: /datex/2/ does not list $PROVIDER_PATH"
  echo "$providers"
  exit 1
fi

echo "==> Checking the response is a DATEX II publication"
body="$(curl -s "$BASE_URL$PROVIDER_PATH")"
if ! grep -q "<d2LogicalModel" <<<"$body"; then
  echo "FAIL: response is not a d2LogicalModel document"
  echo "$body"
  exit 1
fi
if ! grep -q "SituationPublication" <<<"$body"; then
  echo "FAIL: response has no SituationPublication payload"
  echo "$body"
  exit 1
fi

echo "==> Validating the response against the DATEX II 2.2.3 XSD"
if ! xmllint --noout --schema "$SCHEMA" - <<<"$body"; then
  echo "FAIL: response does not validate against $SCHEMA"
  exit 1
fi

echo "==> All checks passed"
