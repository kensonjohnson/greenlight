set quiet := true

MAIN_PACKAGE_PATH := "./cmd/api/."
BINARY_NAME := "greenlight-api"

[private]
help:
  just --list --unsorted

_confirm:
  echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

dev:
  air

test:
  go test -v -race -buildvsc ./...

audit:
  go mod verify
  go vet ./...

tidy:
  go fmt ./...
  go mod tidy -v

build:
  go build -o=/tmp/bin/{{ BINARY_NAME }} {{ MAIN_PACKAGE_PATH}}
