# Semx

Semx is a semantic executor for assertions powered by coding agents: `assert()` for things that need an agent.

## Status

This repository contains the initial CLI skeleton. The assertion definition format and coding-agent integrations will be added in follow-up changes.

## Build

```bash
make build
```

The binary is written to `bin/semx`.

## Run

```bash
./bin/semx --help
./bin/semx version
```

## Development

```bash
make fmt-check
make test
make vet
make lint
```

`make lint` installs the pinned `golangci-lint` binary under `bin/` when needed.

## Project Layout

```text
cmd/semx/         CLI entry point and command routing
internal/version/ Build version metadata
```
