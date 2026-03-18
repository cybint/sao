# SAO Repository Bootstrap Plan

## Goal

Create a practical first-version foundation for SAO as a TAK Server replacement, with:

- a runnable Go service skeleton
- a clean project layout
- baseline quality tooling
- onboarding documentation

## Replacement Intent

For this project, "TAK Server replacement" means SAO should progressively deliver
operational parity for the most important user journeys first, then expand into
advanced TAK Server behaviors.
The primary server responsibility is ToC (Cursor on Target) routing, plus video
and audio routing with bidirectional WebSocket connectivity.

Early parity priorities:

- reliable client connectivity and session lifecycle
- ToC ingestion, routing, and fan-out distribution
- video and audio stream routing
- bidirectional WebSocket session handling
- operational visibility (health, logs, metrics)
- predictable upgrade and deployment workflows
- embedded messaging baseline using NATS
- plugin-based extensibility using HashiCorp go-plugin

## Scope

- Initialize a conventional Go project structure for a server application.
- Define a first runnable server entrypoint and health endpoint.
- Add baseline developer workflows (`build`, `run`, `test`, `lint`).
- Establish CI checks and local setup guidance.

## Target Files

- `cmd/sao/main.go`
- `internal/config/config.go`
- `internal/health/handler.go`
- `internal/server/server.go`
- `internal/health/handler_test.go`
- `internal/server/server_test.go`
- `Makefile`
- `.github/workflows/ci.yml`
- `README.md`
- `docs/README.md`

## Implementation Phases

### Phase 1: Project Skeleton

- Create module and folder layout under `cmd` and `internal`.
- Define package boundaries to support future TAK-specific features.

### Phase 2: Minimal Runnable Service

- Start an HTTP server with configurable bind address.
- Expose `GET /healthz` with a JSON status payload.
- Handle graceful shutdown on SIGINT/SIGTERM.
- Run embedded NATS in-process for baseline pub/sub infrastructure.
- Add host-side plugin lifecycle (load/start/stop) for extension hooks.

### Phase 3: Dev Tooling Baseline

- Add `Makefile` targets for build/run/test/lint.
- Add initial tests for health behavior and server serving/shutdown flow.

### Phase 4: CI Baseline

- Add a GitHub Actions workflow for test, lint, and build checks.
- Keep CI lightweight and fast for early iterations.

### Phase 5: Documentation and Onboarding

- Expand root README with setup, quick start, config, and workflows.
- Use `docs/README.md` as a docs index.

### Phase 6: TAK Replacement Milestones

- Define a compatibility matrix against core TAK Server capabilities.
- Prioritize milestone slices by operational impact and migration risk.
- Add acceptance criteria per milestone and track progress in docs.

Suggested milestone track:

- M1: Core runtime readiness
  - health, logging, config, graceful shutdown, CI
- M2: Client/session baseline
  - connection lifecycle, auth placeholder, session observability
- M3: Event flow baseline
  - ingest, validate, route, and fan-out critical event paths
- M4: Ops hardening
  - structured metrics, failure recovery behavior, load profiling
- M5: Migration readiness
  - deployment docs, rollback guidance, compatibility notes

## Out of Scope

- Full TAK protocol implementation
- Authentication and authorization design
- Persistence/storage architecture
- Production deployment manifests

## Success Criteria

- Service boots consistently with documented local workflow.
- CI passes test/lint/build on every push and pull request.
- Health endpoint and graceful shutdown are validated by tests.
- Documentation enables a new developer to run SAO in minutes.
- Replacement milestones are explicit, measurable, and versioned.
