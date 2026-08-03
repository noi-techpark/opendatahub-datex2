#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
#
# SPDX-License-Identifier: CC0-1.0

# Regenerates equivalence/golden/ from the real, dockerized C# datexpub app,
# run against the fixtures/ mock Open Data Hub response. Not part of normal
# `go test` - run by hand whenever fixtures/ or initdb.sql change.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EQUIV_DIR="$REPO_ROOT/equivalence"

echo "==> Building datexpub:local image"
docker build -f "$REPO_ROOT/src/common/service/DatexPub/Dockerfile" -t datexpub:local "$REPO_ROOT/src/common"

echo "==> Starting containers"
# The datexpub container runs as root, so it may have left root-owned files
# in these bind mounts from a previous run - clean up via a container
# instead of the host user, who may not have permission to remove them.
docker run --rm -v "$EQUIV_DIR:/data" alpine sh -c 'rm -rf /data/pubblicazioni /data/log'
(cd "$EQUIV_DIR" && docker compose up -d)

echo "==> Waiting for first publication cycle"
sleep 10

echo "==> datexpub logs"
(cd "$EQUIV_DIR" && docker compose logs datexpub --tail=80)

echo "==> Capturing golden XML"
rm -rf "$EQUIV_DIR/golden"
mkdir -p "$EQUIV_DIR/golden"
found=0
for f in "$EQUIV_DIR"/pubblicazioni/Invio/*/SituationPublication.xml; do
    [ -e "$f" ] || continue
    supplier="$(basename "$(dirname "$f")")"
    mkdir -p "$EQUIV_DIR/golden/$supplier"
    cp "$f" "$EQUIV_DIR/golden/$supplier/SituationPublication.xml"
    echo "captured golden/$supplier/SituationPublication.xml"
    found=1
done
if [ "$found" -eq 0 ]; then
    echo "!! no output files found under pubblicazioni/Invio/ - check the logs above" >&2
fi

echo "==> Tearing down"
(cd "$EQUIV_DIR" && docker compose down -v)
