# anytype-api-cli

A command line interface for the
[Anytype API](https://developers.anytype.io/docs/reference), written in Go.

The typed API client is **auto-generated** from Anytype's official OpenAPI spec
using [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen), so it stays
in lockstep with the published contract.

## Status

Implements global **search** (`POST /v1/search`). The architecture is set up so
more commands (spaces, objects, types, …) can be added incrementally.

## Install

Requires Go 1.24+.

```sh
go build -o anytype-api ./cmd/anytype-api
# optionally: go install ./cmd/anytype-api
```

## Authentication

The CLI reads your API token from an environment variable:

| Variable          | Required | Default                  | Purpose                              |
| ----------------- | -------- | ------------------------ | ------------------------------------ |
| `ANYTYPE_API_KEY` | yes      | —                        | Bearer token for the Anytype API     |
| `ANYTYPE_API_URL` | no       | `http://127.0.0.1:31009` | Base URL of the local Anytype server |

```sh
export ANYTYPE_API_KEY="your-token-here"
```

The Anytype desktop app exposes its API locally on `127.0.0.1:31009`. Every
request sends the required `Anytype-Version: 2025-11-08` header automatically.

## Usage

```sh
# Search every space for "roadmap"
anytype-api search roadmap

# Restrict to specific object types (repeatable), limit results
anytype-api search "launch" --type task --type page --limit 10

# Paginate
anytype-api search roadmap --limit 20 --offset 20

# Machine-readable output for scripting (pipe to jq, etc.)
anytype-api search roadmap --json
```

### `search` flags

| Flag       | Short | Default | Description                                            |
| ---------- | ----- | ------- | ------------------------------------------------------ |
| `--type`   | `-t`  | —       | Object type to include (repeatable): `page`, `task`, … |
| `--limit`  | `-L`  | `100`   | Maximum results to return (max 1000)                   |
| `--offset` |       | `0`     | Results to skip (for pagination)                       |
| `--json`   |       | `false` | Emit the raw API response as JSON                      |

File-layout objects (file, image, video, audio, pdf) are excluded by default;
pass the corresponding `--type` to include them.

## Project layout

```
api/
  openapi.yaml         # Vendored Anytype OpenAPI spec (source of truth)
  oapi-codegen.yaml    # Generator config (currently scoped to the Search tag)
internal/
  api/                 # Auto-generated client + models (do not edit by hand)
  anytype/             # Thin wrapper: env config, auth, request helpers
cmd/anytype-api/       # Cobra CLI commands
```

## Regenerating the client

After updating `api/openapi.yaml` (or widening the `include-tags` in
`api/oapi-codegen.yaml` to expose more endpoints):

```sh
go generate ./...
```

The generator is pinned as a Go tool dependency, so no separate install is
needed.
