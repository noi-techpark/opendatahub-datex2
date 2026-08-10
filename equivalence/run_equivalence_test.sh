#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
#
# SPDX-License-Identifier: CC0-1.0

# Proves the Go pipeline is equivalent to the real, dockerized C# app: fetch
# a real sample from the live Open Data Hub API, feed the identical sample
# to both, and compare their outputs. Not part of `go test ./...` or CI -
# run by hand. Requires network access and Docker.
#
# The fixture is fetched fresh every run rather than checked into git:
# several event categories (accident, animal-on-road, weather-related,
# speed-camera) are short-lived by nature, so a frozen real sample of them
# goes stale (filtered out as expired) within hours - a synthetic fixture
# with invented far-future dates avoids that, but a hand-authored fixture is
# exactly what let the CreationTime/VersionTime field-mapping bug slip
# through undetected previously. Fetching for real each run trades a fully
# reproducible fixture for one that can never silently drift from what the
# API actually returns.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EQUIV_DIR="$REPO_ROOT/equivalence"
EVENTS_URL="https://content.api.opendatahub.testingmachine.eu/v1/Announcement?pagesize=4000&tagfilter=announcement%3Atraffic-event&removenullvalues=true&getasidarray=false&source=PROVINCE_BZ&fields=Id%2CTagIds%2CGeo%2CMapping%2C_Meta%2CStartTime%2CEndTime%2CFirstImport%2CLastChange%2CDetail"

cleanup() {
  echo "==> Tearing down"
  (cd "$EQUIV_DIR" && docker compose down -v) || true
}
trap cleanup EXIT

echo "==> Fetching a real sample from the live Open Data Hub API"
mkdir -p "$EQUIV_DIR/fixtures/v1"
curl -sf "$EVENTS_URL" -o "$EQUIV_DIR/fixtures/v1/Announcement"
python3 -c "import json,sys; d=json.load(open('$EQUIV_DIR/fixtures/v1/Announcement')); print(f'  {len(d[\"Items\"])} items fetched')"

echo "==> Building datexpub:local image"
docker build -f "$REPO_ROOT/reference/common/service/DatexPub/Dockerfile" -t datexpub:local "$REPO_ROOT/reference/common"

echo "==> Starting containers"
# The datexpub container runs as root, so it may have left root-owned files
# in these bind mounts from a previous run.
docker run --rm -v "$EQUIV_DIR:/data" alpine sh -c 'rm -rf /data/pubblicazioni /data/log /data/golden'
(cd "$EQUIV_DIR" && docker compose up -d)

echo "==> Waiting for first publication cycle"
sleep 10

echo "==> datexpub logs"
(cd "$EQUIV_DIR" && docker compose logs datexpub --tail=80)

echo "==> Capturing golden XML"
mkdir -p "$EQUIV_DIR/golden"
found=0
for f in "$EQUIV_DIR"/pubblicazioni/Invio/*/SituationPublication.xml; do
    [ -e "$f" ] || continue
    supplier="$(basename "$(dirname "$f")")"
    mkdir -p "$EQUIV_DIR/golden/$supplier"
    cp "$f" "$EQUIV_DIR/golden/$supplier/SituationPublication.xml"
    echo "  captured golden/$supplier/SituationPublication.xml"
    found=1
done
if [ "$found" -eq 0 ]; then
    echo "!! no output files found under pubblicazioni/Invio/ - check the logs above" >&2
    exit 1
fi

echo "==> Running Go equivalence test"
(cd "$REPO_ROOT/src" && go test -run TestEquivalenceAgainstLegacyGolden -v ./...)
