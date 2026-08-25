# Semx

Semx is a semantic executor for assertions powered by coding agents: `assert()` for things that need an agent.

Semx owns the portable evaluation request—corpus, prompt, output contract, validation, and run artifacts. Pluggable runners lower that request into Codex, Claude Code, OpenCode, or Pi command invocations.

## Build

```bash
make build
```

The binary is written to `bin/semx`.

## Download the latest main build

The [`main-latest` prerelease](https://github.com/popodidi/semx/releases/tag/main-latest) contains CLI builds for Linux, macOS, and Windows on amd64 and arm64:

- [Linux amd64](https://github.com/popodidi/semx/releases/download/main-latest/semx-linux-amd64.tar.gz)
- [Linux arm64](https://github.com/popodidi/semx/releases/download/main-latest/semx-linux-arm64.tar.gz)
- [macOS amd64](https://github.com/popodidi/semx/releases/download/main-latest/semx-darwin-amd64.tar.gz)
- [macOS arm64](https://github.com/popodidi/semx/releases/download/main-latest/semx-darwin-arm64.tar.gz)
- [Windows amd64](https://github.com/popodidi/semx/releases/download/main-latest/semx-windows-amd64.zip)
- [Windows arm64](https://github.com/popodidi/semx/releases/download/main-latest/semx-windows-arm64.zip)
- [SHA-256 checksums](https://github.com/popodidi/semx/releases/download/main-latest/SHA256SUMS)

This is a mutable prerelease. Its assets are replaced after every successful build of `main`; the release notes identify the exact commit.

## Validate a configuration

```bash
./bin/semx validate examples/factuality/semx.yaml
```

A valid configuration prints `valid`. YAML fields can be overridden with dotted flags:

```bash
./bin/semx validate examples/factuality/semx.yaml \
  --runner.type=claude \
  --runner.args.model=opus
```

## Run an assertion

```bash
./bin/semx run examples/factuality/semx.yaml \
  --corpus examples/factuality/corpus \
  --out /tmp/semx-run
```

`--corpus` aliases `--corpus.path`, and `--out` aliases `--output.dir`. Runner-specific settings stay under `runner.args` and can be overridden as `--runner.args.<name>=<value>`.

A successful JSON run writes:

```text
/tmp/semx-run/
├── result.json
├── runner.stderr
├── runner.stdout
└── semx.resolved.yaml
```

Text runs write `result.txt` instead.

## Configuration

```yaml
runner:
  type: codex
  args:
    model: gpt-5.6

corpus:
  path: ./corpus

prompt:
  system: You are an LLM judge.
  user: Did the agent answer based on factual data?

output:
  dir: ./output
  format: json
  schema: ./schema.json
```

Configuration paths in YAML are resolved relative to the configuration file. CLI path overrides are resolved from the caller's working directory.

## Development

```bash
make fmt-check
make build
make test
make vet
make lint
```

`make lint` installs the pinned `golangci-lint` binary under `bin/` when needed.
