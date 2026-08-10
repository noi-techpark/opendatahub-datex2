REM SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
REM
REM SPDX-License-Identifier: CC0-1.0

docker stop datexDB
docker container prune --force

docker run -d --restart=unless-stopped -p 5432:5432 --name datexDB --shm-size=1g -e PGDATA=/var/lib/postgresql/data -e POSTGRES_PASSWORD=pgpass -v "E:\TEST\NODO_DATEX\db":/var/lib/postgresql/data -v "E:\TEST\NODO_DATEX\postgres.conf":/var/lib/postgresql/postgres.conf -v "E:\TEST\NODO_DATEX\pg_hba.conf":/var/lib/postgresql/pg_hba.conf postgres:17.0

pause
