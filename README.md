<!--
SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>

SPDX-License-Identifier: CC0-1.0
-->

# opendatahub-datex2
Datex 2 node to publish traffic events from the Open Data Hub

## Repository layout

- [src/](src/) - the Go implementation of the service (this is what gets built and deployed).
- [reference/](reference/) - the legacy .NET DatexPub service, kept as a behavioral reference.
- [test/](test/) - docker-based integration test for the Go service, see [test/run_local_test.sh](test/run_local_test.sh).
- [infrastructure/](infrastructure/) - Dockerfile and Helm chart used to build and deploy the Go service.

## API

Each event provider's DATEX II XML is served at
`/datex/2/{provider}/situation-publication.xml`.

`/` serves interactive API docs (Redoc), and `/openapi.yaml` the underlying spec.

## Running locally

```
cp .env.example .env
docker compose up --build
```

This builds and runs the Go service from [src/](src/) against its bundled `config.example.yaml`, listening on the port configured in `.env`.

## Testing

`go test ./...` in [src/](src/) runs the unit tests. [test/run_local_test.sh](test/run_local_test.sh) is an integration test: it builds the production Docker image, runs it against the real Open Data Hub API, and checks that it serves a valid DATEX II publication plus the API docs. It needs network access and Docker.
