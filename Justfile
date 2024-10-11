set quiet := true
set dotenv-load

MAIN_PACKAGE_PATH := "./cmd/api/."
BINARY_NAME := "greenlight-api"

# =========================================================================== #
# HELPERS
# =========================================================================== #

# Display help
[private]
help:
  just --list --unsorted

# Confirm or cancel recipe
_confirm:
  echo -n "Are you sure? [y/N] " && read ans && [ ${ans:-N} = y ]

# =========================================================================== #
# DEVELOPMENT
# =========================================================================== #

# Run dev server
run: 
  go run ./cmd/api

# Log into Postgres using psql 
psql:
  psql $DSN

# Preform migration on DB
db-migrations-up: _confirm
  echo 'Running up migrations...'
  migrate -path ./migrations -database $DSN up

# Create new migration files
db-migrations-new name: _confirm
  echo 'Creating migration files for {{name}}...'
  migrate create -seq -ext=.sql -dir=./migrations {{name}}

# =========================================================================== #
# QUALITY CONTROL
# =========================================================================== #

# Verify and Vet all Go files in project
audit:
  echo 'Checking module dependencies'
  go mod tidy -diff
  go mod verify
  echo 'Vetting code...'
  go vet ./...
  echo 'Running tests...'
  go test -race -vet=off ./...

# Run formatter and tidy over all Go files in project
tidy:
  echo 'Formatting .go files...'
  go fmt ./...
  echo 'Tidying module dependencies...'
  go mod tidy -v


# =========================================================================== #
# BUILD
# =========================================================================== #

# Build for current OS/Arch
build:
  go build -ldflags='-s' -o=./tmp/{{ BINARY_NAME }} {{ MAIN_PACKAGE_PATH}}
  GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./tmp/linux_amd64/{{ BINARY_NAME }} {{ MAIN_PACKAGE_PATH}}
