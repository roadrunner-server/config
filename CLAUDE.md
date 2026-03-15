# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

RoadRunner **config** plugin (v6). It reads YAML configuration files for the RoadRunner application server, expands environment variables, supports config file includes, and provides typed access to configuration sections via Viper. Module path: `github.com/roadrunner-server/config/v6`.

## Build & Test Commands

```bash
# Run all tests (from repo root — uses go.work)
make test                    # or: go test -v -race -tags=debug ./...

# Run tests with coverage (matches CI)
cd tests && go test -timeout 20m -v -race -cover -tags=debug \
  -failfast -coverpkg=github.com/roadrunner-server/config/v6/... \
  -coverprofile=coverage.out -covermode=atomic ./...

# Run a single test
cd tests && go test -v -race -tags=debug -run TestViperProvider_Init ./...

# Lint
golangci-lint run --build-tags=race ./...
```

## Go Workspace Layout

Uses `go.work` with two modules:

- **`.`** — the plugin library itself (`package config`): `plugin.go`, `expand.go`, `include.go`
- **`./tests`** — separate module for integration tests; has its own `go.mod` with `replace github.com/roadrunner-server/config/v6 => ../` so it always tests the local code

All test files, test fixtures (YAML configs, `.env` files, PHP test files), and mock plugins (`plugin1.go`–`plugin5.go`) live under `tests/`.

## Architecture

The plugin has three files in the root package:

| File | Purpose |
|------|---------|
| `plugin.go` | `Plugin` struct, `Init()` lifecycle, flag parsing, version checking, `UnmarshalKey`/`Get`/`Has` API |
| `expand.go` | Environment variable expansion: `${VAR}`, `$VAR`, `${VAR:-default}` syntax; iterates all Viper keys including string slices |
| `include.go` | Config file includes (`include:` key), `.env` file loading via `godotenv` (experimental feature) |

**Plugin lifecycle**: `Init()` reads the YAML file into Viper → loads `.env` file (if experimental) → expands env vars → applies CLI flag overrides → validates version key → merges included config files.

**Endure integration**: The plugin is designed to be registered with the [Endure](https://github.com/roadrunner-server/endure) dependency injection container. Other RoadRunner plugins receive it as a `Configurer` interface dependency. Tests demonstrate this pattern using real RoadRunner plugins (logger, rpc, server, kv, memory).

## Key Conventions

- **Error handling**: Uses `github.com/roadrunner-server/errors` with `errors.Op` for operation tracing. Each function defines `const op = errors.Op("...")` and wraps errors with `errors.E(op, err)`.
- **Config version**: Every config file must have `version: "3"` (string). The plugin enforces this.
- **Experimental features**: `.env` file support and `include:` require `ExperimentalFeatures: true`.
- **`mapstructure` tags**: Config structs use `mapstructure` tags (not `yaml`) since Viper unmarshals via mapstructure.

## Linting

golangci-lint v2 config in `.golangci.yml`. Pre-commit hook runs `golangci-lint --build-tags=race run`. Install hook: `./githooks-installer.sh`. Tests are **not** excluded from linting (only `dupl`, `funlen`, `gocognit`, `scopelint` are relaxed for `_test.go`).

## CI

- **linux.yml**: Tests on ubuntu with Go stable + PHP 8.5 (PHP needed for test fixture `composer.json`). Uploads coverage to Codecov.
- **linters.yml**: Runs golangci-lint on push and PR.
- **codeql-analysis.yml**: CodeQL security scanning.
