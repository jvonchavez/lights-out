# Lights Out — Architecture

## Shape

One Go binary. A React frontend embedded inside it via `embed.FS`. Postgres. That is the entire
system, and keeping it that small is a deliberate choice — the interesting engineering is the
dual-target simulation and the verification model, not the number of moving parts.

```
┌─────────────────────────────────────────────┐
│  Browser                                    │
│  ┌───────────────┐   ┌──────────────────┐   │
│  │ React + TS UI │──▶│ sim.wasm (Go)    │   │
│  └───────────────┘   └──────────────────┘   │
│         │  plays the full season locally     │
└─────────┼───────────────────────────────────┘
          │ POST /api/runs  { seed, decisions[] }
          ▼
┌─────────────────────────────────────────────┐
│  Go service (single binary, AWS App Runner) │
│   ├── HTTP handlers                          │
│   ├── internal/sim  ← the same package the   │
│   │                    WASM build compiles   │
│   ├── verifier: re-runs decisions natively   │
│   ├── daily seed scheduler                   │
│   └── embedded static assets (embed.FS)      │
└─────────┬───────────────────────────────────┘
          │ pgx
          ▼
   ┌──────────────┐
   │ RDS Postgres │
   └──────────────┘
```

## The dual-target simulation

`internal/sim` is a pure package: no I/O, no network, no clock, no logging. It exposes essentially one
function.

```go
func RunSeason(seed int64, decisions []Decision) (SeasonResult, error)
```

It compiles to two targets:

- **Native**, linked into the server binary, used by the verifier.
- **WebAssembly** (`GOOS=js GOARCH=wasm`), loaded by the browser, used to play.

Because there is one implementation, client and server cannot disagree about the rules. A parity test
runs several thousand seeds through both targets and asserts byte-identical results. This is the
architectural centrepiece — everything else is plumbing around it.

**Known cost, accepted:** Go's WASM output is large (~2 MB gzipped for a small program, because the
runtime and GC ship with it). Mitigations are `-ldflags="-s -w"`, Brotli compression at the edge, and
loading the module lazily after first paint. If it proves unacceptable, the fallback is TinyGo, which
produces far smaller binaries at the cost of some standard-library support — evaluate only if the real
number is a problem, not preemptively.

## Request flow

**Playing.** The browser fetches today's season descriptor (seed, calendar, starting ratings), loads
the WASM module, and runs the entire season client-side. Every race resolves instantly with no network
round-trip. The player's decisions accumulate in memory.

**Submitting.** The client posts `{ season_id, decisions[] }` — never a score. The server loads the
season's seed, calls `sim.RunSeason` natively with the submitted decisions, and computes the
authoritative result. The client's own score is not sent, not read, and not trusted.

**Verification** rejects a submission when: the decision count does not match the race count, any
allocation is negative or exceeds that race's budget, the season is closed, or this player has already
submitted for this season. Everything the client could lie about is recomputed from the seed.

This is the property that makes determinism load-bearing rather than a stylistic preference. A
cheater would have to find decisions that genuinely produce a high score — which is just playing well.

## Data model

```sql
CREATE TABLE seasons (
  id           bigserial PRIMARY KEY,
  seed         bigint      NOT NULL,
  sim_version  text        NOT NULL,
  calendar     jsonb       NOT NULL,   -- circuits, profiles, budgets
  field        jsonb       NOT NULL,   -- rival teams, starting ratings, AI archetypes
  published_at timestamptz NOT NULL,
  closes_at    timestamptz NOT NULL,
  UNIQUE (published_at)
);

CREATE TABLE players (
  id           uuid PRIMARY KEY,       -- generated client-side, stored in localStorage
  display_name text        NOT NULL,
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE runs (
  id          bigserial PRIMARY KEY,
  season_id   bigint      NOT NULL REFERENCES seasons(id),
  player_id   uuid        NOT NULL REFERENCES players(id),
  decisions   jsonb       NOT NULL,    -- the only thing the client supplies
  points      int         NOT NULL,    -- computed server-side
  wins        int         NOT NULL,
  podiums     int         NOT NULL,
  dnfs        int         NOT NULL,
  result      jsonb       NOT NULL,    -- full race-by-race, for the share card
  created_at  timestamptz NOT NULL DEFAULT now(),
  UNIQUE (season_id, player_id)
);

CREATE INDEX runs_leaderboard ON runs (season_id, points DESC, wins DESC, podiums DESC);
```

`seasons.sim_version` is what lets sim rules change without silently invalidating history. A season is
verified under the version it was published with; old seasons are frozen, never recomputed.

The `UNIQUE (season_id, player_id)` constraint is the entire anti-resubmission mechanism — one
submission per player per day, enforced by the database rather than by application logic that can
race.

## Daily seed scheduler

An in-process goroutine wakes once an hour, checks whether a season exists for the current UTC day,
and creates one if not. Deliberately not a separate service, not a Lambda, not a cron container.

Seasons are generated from a deterministic function of the date, so the scheduler is idempotent — two
instances racing produce the same row, and the `UNIQUE (published_at)` constraint settles the tie
harmlessly. This is what makes it safe to run more than one instance without leader election.

## Observability

- **Logging:** `log/slog` with the JSON handler. Every request carries a request ID; every submission
  logs season ID, player ID, verification duration, and outcome.
- **Health:** `/healthz` returns 200 only after a successful Postgres ping.
- **Metrics:** `/metrics` in Prometheus format via the official client —
  `http_request_duration_seconds` by route and status, `sim_verification_duration_seconds`,
  `runs_submitted_total` by outcome, and `seasons_published_total`.

Verification duration is the metric worth watching: it is CPU-bound work on the request path, and if
it ever creeps toward the hundreds of milliseconds the answer is to move verification onto a worker
queue. Instrumenting it from day one is what makes that a measurement rather than a guess.

## Deployment

Single multi-stage Dockerfile: a builder stage that compiles both the WASM module and the native
binary and builds the frontend, then a `scratch`-based final stage holding one static binary with
assets embedded. Expected image size is roughly 20–25 MB.

**AWS App Runner** over ECS Fargate for the MVP — it handles TLS, scaling, and deploys from an image
with far less configuration, and this workload has no requirement Fargate would satisfy better. RDS
Postgres on `db.t4g.micro`. If App Runner's pricing floor becomes annoying, the migration to Fargate
is a day's work because the container is already the deployable unit.
