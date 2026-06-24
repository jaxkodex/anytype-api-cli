# anytype-api-cli

A command line interface for the
[Anytype API](https://developers.anytype.io/docs/reference), written in Go.

The typed API client is **auto-generated** from Anytype's official OpenAPI spec
using [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen), so it stays
in lockstep with the published contract.

## Status

Implements **auth** (`POST /v1/auth/challenges` and `.../api_keys`): start a
challenge and exchange the 4-digit code for an API key, **spaces** management
(list, get, create and update via `/v1/spaces` and `/v1/spaces/{space_id}`),
global **search** (`POST /v1/search`), **types** inspection
(`GET /v1/spaces/{space_id}/types` and `.../types/{type_id}`), **files**
management (upload, download and delete via `.../files` and
`.../files/{file_id}`), **lists** (collections/sets): inspect their views and
objects, and add/remove objects via `.../lists/{list_id}/...`, **properties**
management (list, get, create, update and delete via `.../properties` and
`.../properties/{property_id}`), **tags** (the selectable values of a
select/multi-select property): list, get, create, update and delete via
`.../properties/{property_id}/tags/...`, and **objects** (list, get, create,
update, delete via `.../objects` and `.../objects/{object_id}`). The architecture
is set up so more commands (members, …) can be added incrementally.

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

### `types`

Types (Page, Task, Bookmark, …) are scoped to a space, so every subcommand
requires a `--space` id.

```sh
# List every type defined in a space
anytype-api types list --space bafyre...

# Show one type's details
anytype-api types get bafyre...type-id --space bafyre...

# Machine-readable output
anytype-api types list --space bafyre... --json

# Create a type with convenience flags
anytype-api types create --space bafyre... \
  --name Task --plural Tasks --layout basic --icon ✅

# Create a type from a JSON payload (file or stdin)
cat type.json | anytype-api types create --space bafyre... --file -

# Update (rename) a type — only supplied fields change
anytype-api types update bafyre...type-id --space bafyre... --name "New name"

# Delete (archive) a type, with confirmation
anytype-api types delete bafyre...type-id --space bafyre...

# Delete without prompting (for scripts)
anytype-api types delete bafyre...type-id --space bafyre... --yes
```

| Flag       | Short | Default | Description                                             |
| ---------- | ----- | ------- | ------------------------------------------------------- |
| `--space`  | `-s`  | —       | Space id to operate on (**required**)                   |
| `--limit`  | `-L`  | `100`   | Maximum results to return (`list`)                      |
| `--offset` |       | `0`     | Results to skip (`list`)                                |
| `--file`   | `-f`  | —       | JSON payload file, `-` for stdin (`create`, `update`)   |
| `--name`   |       | —       | Type name (`create`, `update`)                          |
| `--plural` |       | —       | Plural type name (`create`, `update`)                   |
| `--key`    |       | —       | Type key in snake_case (`create`, `update`)             |
| `--layout` |       | —       | Layout: `basic`, `note`, `profile`, `action`            |
| `--icon`   |       | —       | Emoji icon for the type (`create`, `update`)            |
| `--yes`    | `-y`  | `false` | Skip the confirmation prompt (`delete`)                 |
| `--json`   |       | `false` | Emit the raw API response as JSON                       |

The type definition for `create`/`update` can come from a `--file` JSON payload
(matching the API's `CreateTypeRequest`/`UpdateTypeRequest` shape), from the
convenience flags, or both. When combined, flags take precedence over fields in
the payload, so a file can serve as a template you tweak per invocation.

### `files`

Files are scoped to a space, so every subcommand requires a `--space` id.

```sh
# Upload a local file
anytype-api files upload ./photo.png --space bafyre...

# Machine-readable output for the uploaded file object
anytype-api files upload ./photo.png --space bafyre... --json

# Download a file to a sensible local filename
anytype-api files download bafyre...file-id --space bafyre...

# Download to a specific path
anytype-api files download bafyre...file-id --space bafyre... --output photo.png

# Stream to stdout for piping (also the default when stdout is not a terminal)
anytype-api files download bafyre...file-id --space bafyre... --output - > photo.png

# Delete (move to bin), with confirmation
anytype-api files delete bafyre...file-id --space bafyre...

# Delete permanently, without prompting (for scripts)
anytype-api files delete bafyre...file-id --space bafyre... --skip-bin --yes
```

| Flag         | Short | Default | Description                                      |
| ------------ | ----- | ------- | ------------------------------------------------ |
| `--space`    | `-s`  | —       | Space id to operate on (**required**)            |
| `--output`   | `-o`  | —       | Destination path, `-` for stdout (`download`)    |
| `--skip-bin` |       | `false` | Permanently delete instead of the bin (`delete`) |
| `--yes`      | `-y`  | `false` | Skip the confirmation prompt (`delete`)          |
| `--json`     |       | `false` | Emit the raw API response as JSON (`upload`)     |

`download` writes to `--output` when given (`-` means stdout). Without
`--output`, it streams to stdout when that is not a terminal (so the command
pipes safely), otherwise it writes to a file named after the file id with an
extension inferred from the response media type.

### `lists`

Lists (collections and sets) are scoped to a space, so every subcommand requires
a `--space` id. Find a list id by searching the space for objects of type
`collection` or `set`.

```sh
# List the views defined for a list
anytype-api lists views bafyre...list-id --space bafyre...

# List the objects in a view (filtered and sorted by that view)
anytype-api lists objects bafyre...list-id --space bafyre... --view 67bf3f21...

# Add one or more objects to a collection
anytype-api lists add bafyre...list-id --space bafyre... bafyreA... bafyreB...

# Remove an object from a collection, with confirmation
anytype-api lists remove bafyre...list-id bafyreA... --space bafyre...

# Remove without prompting (for scripts)
anytype-api lists remove bafyre...list-id bafyreA... --space bafyre... --yes

# Machine-readable output
anytype-api lists views bafyre...list-id --space bafyre... --json
```

| Flag       | Short | Default | Description                                         |
| ---------- | ----- | ------- | --------------------------------------------------- |
| `--space`  | `-s`  | —       | Space id the list belongs to (**required**)         |
| `--view`   |       | —       | View id to filter/sort by (**required**, `objects`) |
| `--limit`  | `-L`  | `100`   | Maximum results to return (`views`, `objects`)      |
| `--offset` |       | `0`     | Results to skip (`views`, `objects`)                |
| `--yes`    | `-y`  | `false` | Skip the confirmation prompt (`remove`)             |
| `--json`   |       | `false` | Emit the raw API response as JSON                   |

Only collections can be modified with `add`/`remove`; the objects of a set are
determined by its query.

## Project layout

```
api/
  openapi.yaml         # Vendored Anytype OpenAPI spec (source of truth)
  oapi-codegen.yaml    # Generator config (scoped to the Search, Types and Lists tags)
internal/
  api/                 # Auto-generated client + models (do not edit by hand)
  anytype/             # Thin wrapper: env config, auth, request helpers
                       #   client.go holds the shared Client; per-resource
                       #   methods live in client_<resource>.go
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

`internal/api/anytype.gen.go` is fully derived from the spec, so it should never
be merged textually. `.gitattributes` marks it `merge=ours`; enable that driver
once per clone:

```sh
git config merge.ours.driver true
```

After rebasing a feature branch, regenerate rather than resolving conflicts in
the generated file by hand:

```sh
go generate ./...
```
