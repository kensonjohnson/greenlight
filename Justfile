set quiet := true

MAIN_PACKAGE_PATH := "./cmd/api/."
BINARY_NAME := "greenlight-api"

[private]
help:
  just --list --unsorted

_confirm:
  echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

# Run dev server
dev:
  air

# Run all Go test files
test:
  go test -v -race -buildvsc ./...

# Verify and Vet all Go files in project
audit:
  go mod verify
  go vet ./...

# Run formatter and tidy over all Go files in project
tidy:
  go fmt ./...
  go mod tidy -v

# Build for current OS/Arch
build:
  go build -o=/tmp/bin/{{ BINARY_NAME }} {{ MAIN_PACKAGE_PATH}}

# Build for current OS/Arch and run the resulting binary
run: build
  /tmp/bin/{{ BINARY_NAME }}
