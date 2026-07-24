#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
#
# SPDX-License-Identifier: CC0-1.0

# Builds the DatexPub image and runs it against a throwaway Postgres
# instance defined in local-test/, using the seed data copied from the
# NODO_DATEX repo since this repo does not ship a database setup.
# Linux replacement for the start_db.bat / start_datexPub.bat scripts,
# meant for local testing only.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOCAL_TEST_DIR="$REPO_ROOT/local-test"

echo "==> Building datexpub:local image"
docker build -f "$REPO_ROOT/src/common/service/DatexPub/Dockerfile" -t datexpub:local "$REPO_ROOT/src/common"

echo "==> Starting containers"
(cd "$LOCAL_TEST_DIR" && docker compose up -d)

echo "==> Waiting for first publication cycle"
sleep 8

echo "==> datexpub logs"
(cd "$LOCAL_TEST_DIR" && docker compose logs datexpub --tail=50)

echo "==> Output files"
find "$LOCAL_TEST_DIR/pubblicazioni" -type f

echo "==> Tearing down test stack"
(cd "$LOCAL_TEST_DIR" && docker compose down -v)
