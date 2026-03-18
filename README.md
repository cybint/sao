# SAO

Spatial Awareness Operator (SAO) is an open source TAK Server replacement.

SAO is built to provide a staged migration path toward TAK Server parity,
starting with core runtime reliability and then expanding into protocol and
operational capabilities.
Its primary role is to serve as a ToC (Cursor on Target) routing server, with
added support for routing video/audio streams and establishing a bidirectional
WebSocket channel for real-time communication.

The current bootstrap includes an embedded NATS server for internal messaging.
It also supports HashiCorp plugins for runtime extensibility.
Core Cursor on Target models are defined in `internal/cot` and aligned to
MITRE CoT event semantics (v2.0 event envelope, point fields, and optional
`access`, `qos`, and `opex` attributes).

## Prerequisites

- Go 1.25+

## Quick Start

```bash
make build
make run
```

By default, the server listens on `:8080`.

Health check:

```bash
curl http://localhost:8080/healthz
```

API encoding support:

- JSON via `application/json`
- Protobuf via `application/x-protobuf` using binary `google.protobuf.Value`
  payloads for request and response bodies

Client schema registration (CoT message types):

```bash
curl -X POST http://localhost:8080/schemas \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "alpha",
    "message_type": "cot.position",
    "schema": {
      "type": "object",
      "properties": {
        "uid": {"type": "string"},
        "lat": {"type": "number"},
        "lon": {"type": "number"}
      }
    }
  }'
```

List registered schemas:

```bash
curl "http://localhost:8080/schemas?client_id=alpha"
```

Entity CRUD API:

```bash
# Create
curl -X POST http://localhost:8080/entities \
  -H "Content-Type: application/json" \
  -d '{
    "id": "entity-1",
    "type": "cot.position",
    "data": {"uid":"U1","lat":34.1,"lon":-118.2}
  }'

# Get
curl http://localhost:8080/entities/entity-1

# Update
curl -X PUT http://localhost:8080/entities/entity-1 \
  -H "Content-Type: application/json" \
  -d '{
    "type": "cot.position",
    "data": {"uid":"U1","lat":35.0,"lon":-118.2}
  }'

# Delete
curl -X DELETE http://localhost:8080/entities/entity-1
```

## Configuration

SAO now loads configuration from HCL at `/etc/sao/config.hcl` by default.
If the file does not exist, SAO creates it automatically with defaults.

Default HCL structure:

```hcl
server {
  addr = ":8080"
  shutdown_timeout = "10s"
  plugin_paths = []

  embedded_nats {
    addr = "127.0.0.1"
    port = 4222
  }
}
```

- `SAO_ADDR` - server bind address (default `:8080`)
- `SAO_SHUTDOWN_TIMEOUT` - graceful shutdown timeout (default `10s`)
- `SAO_NATS_ADDR` - embedded NATS bind host (default `127.0.0.1`)
- `SAO_NATS_PORT` - embedded NATS bind port (default `4222`)
- `SAO_PLUGIN_PATHS` - comma-separated plugin binaries to load at startup
- `SAO_UI_ADDR` - UI plugin HTTP bind address (default `:8081`)
- `SAO_CONFIG_FILE` - override config file path (default `/etc/sao/config.hcl`)

Example:

```bash
SAO_ADDR=":9090" SAO_SHUTDOWN_TIMEOUT="15s" SAO_NATS_PORT="5222" make run
```

CLI flags (urfave/cli v3):

```bash
go run ./cmd/sao server --config /etc/sao/config.hcl --addr :9090 --shutdown-timeout 15s --nats-port 5222
```

Admin UI as CLI subcommand:

```bash
go run ./cmd/sao ui --addr :8081 --nats-url nats://127.0.0.1:4222
```

Plugin example:

```bash
make plugins
SAO_PLUGIN_PATHS="./bin/example" make run
```

The host passes embedded NATS connection details to plugins at startup. The
example plugin subscribes to `sao.plugin.example.ping` and replies `pong`.

Admin UI plugin example:

```bash
make plugins
SAO_PLUGIN_PATHS="./bin/ui" SAO_UI_ADDR=":8081" make run
```

Then open `http://localhost:8081` to view the admin UI plugin page.

## Development Workflow

- `make test` - run unit tests
- `make lint` - run static checks (`go vet`)
- `make build` - compile the server binary
- `make run` - start the service locally

## Documentation

- [`docs/README.md`](docs/README.md)
- [`docs/PLAN.md`](docs/PLAN.md)
