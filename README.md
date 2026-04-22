# protoc-gen-aip-lint

[![CI](https://github.com/protoc-contrib/protoc-gen-aip-lint/actions/workflows/ci.yml/badge.svg)](https://github.com/protoc-contrib/protoc-gen-aip-lint/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/protoc-contrib/protoc-gen-aip-lint?include_prereleases)](https://github.com/protoc-contrib/protoc-gen-aip-lint/releases)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![protoc](https://img.shields.io/badge/protoc-compatible-blue)](https://protobuf.dev)

A [protoc](https://protobuf.dev) plugin that runs the [Google API Linter](https://github.com/googleapis/api-linter) on your Protocol Buffer files. It checks `.proto` files against the [AIP](https://aip.dev) style guidelines and reports problems in multiple output formats.

## Features

- Lint `.proto` files against the full set of AIP rules
- Multiple output formats: YAML, JSON, GitHub Actions annotations, and summary table
- Configurable via the standard `api-linter` YAML configuration file
- Works with `protoc`, `buf`, or any protoc-compatible toolchain
- CI-friendly `set-exit-status` option for gating builds on lint failures

## Installation

```bash
go install github.com/protoc-contrib/protoc-gen-aip-lint/cmd/protoc-gen-aip-lint@latest
```

## Usage

### With buf

Add the plugin to your `buf.gen.yaml`:

```yaml
version: v2
plugins:
  - protoc_builtin: cpp
    out: gen/proto/cpp
  - local: protoc-gen-aip-lint
    out: .
    opt:
      - output-format=github
      - set-exit-status
```

Then run:

```bash
buf generate
```

### With protoc

```bash
protoc \
  --aip-lint_out=. \
  --aip-lint_opt=output-format=github,set-exit-status \
  -I proto/ \
  proto/example/v1/example.proto
```

## Plugin Options

Options are passed via `--aip-lint_opt` (protoc) or `opt` (buf). Multiple
options are comma-separated.

| Option                    | Description                                                        |
| ------------------------- | ------------------------------------------------------------------ |
| `config`                  | Path to the `api-linter` YAML configuration file                   |
| `ignore-comment-disables` | Ignore inline disable comments for strict AIP enforcement          |
| `output-format`           | Output format: `yaml` (default), `json`, `github`, or `summary`    |
| `output-path`             | Output file path (defaults to stderr; stdout is reserved for protoc) |
| `set-exit-status`         | Exit with non-zero status when lint problems are found             |

## Output Formats

### YAML (default)

```yaml
- filepath: example.proto
  problems:
    - message: "Field must be in lower_snake_case."
      rule_id: "core::0140::lower-snake"
```

### JSON

```json
[
  {
    "file_path": "example.proto",
    "problems": [
      {
        "message": "Field must be in lower_snake_case.",
        "rule_id": "core::0140::lower-snake"
      }
    ]
  }
]
```

### GitHub Actions

Emits `::error` workflow commands that appear as inline annotations on pull
request diffs:

```
::error file=example.proto,line=10,col=3::core::0140::lower-snake: Field must be in lower_snake_case.
```

### Summary

A human-readable table with problem counts per rule:

```
RULE                                      COUNT
----------------------------------------  -----
core::0123::resource-pattern              1
core::0140::lower-snake                   3
----------------------------------------  -----
Total (2 files)                           4
```

## Configuration

Create an `api-linter` configuration file to enable or disable specific rules:

```yaml
---
- disabled_rules:
    - core::0140::lower-snake
- enabled_rules:
    - core::0123::resource-pattern
```

Pass the config file path as a plugin option:

```bash
protoc \
  --aip-lint_out=. \
  --aip-lint_opt=config=api-linter.yaml \
  proto/example/v1/example.proto
```

See the [api-linter configuration docs](https://linter.aip.dev/configuration)
for the full configuration format.

## CI Integration

Use `output-format=github` with `set-exit-status` to get inline annotations
and fail the build when problems are found:

```yaml
- name: Lint Protos
  run: |
    buf generate --template buf.gen.lint.yaml
```

Where `buf.gen.lint.yaml` contains:

```yaml
version: v2
plugins:
  - local: protoc-gen-aip-lint
    out: .
    opt:
      - output-format=github
      - set-exit-status
```

## Contributing

Contributions are welcome! Please open an issue or pull request.

To set up a development environment with [Nix](https://nixos.org):

```bash
nix develop
go test ./...
```

## License

[MIT](LICENSE)
