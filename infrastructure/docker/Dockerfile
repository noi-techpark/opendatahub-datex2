# SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
#
# SPDX-License-Identifier: CC0-1.0

FROM golang:1.26-bookworm AS base

# build executable
FROM base AS build-env
WORKDIR /app
COPY src/. .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o datexpub

# build published image with binary from previous stage
FROM alpine:latest AS build
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build-env /app/datexpub .
COPY --from=build-env /app/config.example.yaml .
ENTRYPOINT [ "./datexpub" ]

FROM base AS debug
WORKDIR /code
CMD [ "go", "run", "."]
