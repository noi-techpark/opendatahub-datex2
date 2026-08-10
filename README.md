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
- [equivalence/](equivalence/) - proves the Go service is equivalent to the legacy .NET one for the same real input, see [equivalence/run_equivalence_test.sh](equivalence/run_equivalence_test.sh).
- [infrastructure/](infrastructure/) - Dockerfile and Helm chart used to build and deploy the Go service.

## API

Each event provider's DATEX II XML is served at
`/datex/2/{provider}/situation-publication.xml`. `/datex/2/` lists the
available providers and, for each, the DATEX II files it currently
publishes.

`/` serves interactive API docs (Redoc), and `/openapi.yaml` the underlying spec.

## Running locally

```
cp .env.example .env
docker compose up --build
```

This builds and runs the Go service from [src/](src/) against its bundled `config.example.yaml`, listening on the port configured in `.env`.

## Testing

`go test ./...` in [src/](src/) runs the unit tests. [test/run_local_test.sh](test/run_local_test.sh) is an integration test: it builds the production Docker image, runs it against the real Open Data Hub API, checks that it serves a valid DATEX II publication plus the API docs, and validates that publication against the [DATEX II 2.2.3 XSD](test/schema/DATEXIISchema_2_2_3.xsd). It needs network access, Docker, and `xmllint` (libxml2).

[equivalence/run_equivalence_test.sh](equivalence/run_equivalence_test.sh) fetches a real sample from the live Open Data Hub API, feeds it to both the Go service and the real, dockerized legacy .NET service, and asserts their DATEX II output is equivalent field by field (with a short, documented list of known legacy bugs the Go rewrite intentionally doesn't reproduce - see the comment at the top of [src/equivalence_test.go](src/equivalence_test.go)). Not part of `go test ./...` or CI, run by hand; needs network access and Docker.
