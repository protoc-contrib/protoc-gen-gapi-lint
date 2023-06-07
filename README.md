# protoc-gen-gapi-lint

A protoc plugin that runs [api-linter](https://github.com/googleapis/api-linter).

## Usage

```
# buf.gen.yaml
version: v1

managed:
  enabled: true

plugins:
  - name: gapi-lint
    out: internal/proto
    opt:
      - paths=source_relative
      - config=buf.lint.yaml
```
