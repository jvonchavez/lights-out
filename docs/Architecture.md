# Lights Out — Architecture

## Shape

One Go binary with a React frontend embedded via `embed.FS`, plus Postgres. That is the whole
system. The interesting engineering is the dual-target simulation and the verification model, not
the number of moving parts.

```
┌─────────────────────────────────────────────┐
│  Browser                                    │
│  ┌───────────────┐   ┌──────────────────┐   │
│  │ React + TS UI │──▶│ sim.wasm (Go)    │   │
│  └───────────────┘   └──────────────────┘   │
│         │  plays the full season locally    │
└─────────┼───────────────────────────────────┘
          │ POST /api/seasons  → the server mints the seed
          │ POST /api/runs  { season_id, picks[] }
          ▼
┌─────────────────────────────────────────────┐
│  Go service (single binary, AWS App Runner) │
│   ├── HTTP handlers                         │
│   ├── internal/sim  ← the same package the  │
│   │                    WASM build compiles  │
│   ├── verifier: re-runs the picks natively  │
│   ├── seed minting (crypto/rand)            │
│   └── embedded static assets (embed.FS)     │
└─────────┬───────────────────────────────────┘
          │ pgx
          ▼
   ┌──────────────┐
   │ RDS Postgres │
   └──────────────┘
```

## The dual-target simulation

`internal/sim` is pure: no I/O, no network, no clock, no logging.

```go
// picks has one entry per roll, each an index into the five items that
// roll's team-era offers.
func RunSeason(seed int64, picks []int) (SeasonResult, error)

// RollsFor is the draft: the five team-eras a seed offers.
func RollsFor(seed int64) [RollCount]TeamEra

// BuildLineup replays picks against rolls. It is the whole validation
// surface for a submission, and it lives here rather than in the API so the
// browser and the server enforce identical rules.
func BuildLineup(rolls [RollCount]TeamEra, picks []int) (Lineup, error)
```

It compiles to **native** (linked into the server, used by the verifier) and **WebAssembly**
(`GOOS=js GOARCH=wasm`, loaded by the browser, used to play). One implementation means client and
server cannot disagree about the rules; a parity test runs several thousand seeds through both and
asserts byte-identical results. Everything else is plumbing around this.

**Three independent RNG streams**, all derived from the seed: `seed` for season generation,
`seed ^ 0x5EED` for race resolution, `seed ^ 0xCA4D` for draft rolls. Separating them means editing
the roster cannot shift historical race outcomes, and changing generation cannot shift what you are
offered.

**The rival field consumes no randomness.** It is the real 2026 grid, identical on every seed, so
`GenerateSeason` draws nothing but the calendar order. There is no rival AI and no draw-order surface
where rival behaviour could perturb race RNG.

**Known cost, accepted:** Go's WASM output is large (~2 MB gzipped — the runtime and GC ship with
it). Mitigations: `-ldflags="-s -w"`, Brotli at the edge, lazy loading after first paint. Fallback is
TinyGo, evaluated only against a measured number.

## Request flow

**Starting.** `POST /api/seasons` mints a seed from `crypto/rand`, records it, and returns the
calendar, the 2026 field, and the five rolls. **This is the only place a seed enters the system** —
a client that could nominate its own could nominate one it had solved offline, and the submission
would verify perfectly. Seeds are masked below 2^53 and cross as JSON strings, so a client parsing
one as a JavaScript number still gets the season the server will verify.

**Playing.** The browser runs the entire season client-side once the draft is complete; no network
round-trip resolves anything. Rolls come from the API as well as the WASM module, so the draft
renders before the ~1.25 MB module finishes loading.

**Playing again** is another `POST /api/seasons`. There is no limit.

**Submitting.** The client posts `{ season_id, picks[] }` — five integers, never a score. The server
re-derives the rolls from the seed, calls `sim.RunSeason` natively, and computes the authoritative
result. `submitRequest` has no score field, so a forged number has nowhere to be decoded to.

**Verification** rejects a submission when the pick count does not match the roll count, any pick is
out of range, **the picks do not form a legal team**, the season was issued under a different sim
version, or the player already submitted for this season. The legality check is stronger than a card
draft could express: `[0,0,0,0,0]` is five in-range indices and five cars; `[1,2,1,3,4]` is three
drivers and no car. Both are 400s, and the rule lives in `BuildLineup` so the browser greys out
exactly what the server would refuse.

**What this does not claim.** 240 legal drafts is trivially enumerable client-side, and enumerating
them wins 40% of the time against 17.7% for the best scripted line. A server-minted seed also does
not stop a player *declining* seeds until the rolls are good. The honest claim is that scores cannot
be fabricated, not that they cannot be searched for. See `_README.md`.

**The sim version gate is enforced, not documented.** A submission to a season issued under another
ruleset is a 409 naming both versions — otherwise a deploy would silently corrupt the leaderboard.

## Data model

```sql
CREATE TABLE seasons (
  id           bigserial PRIMARY KEY,
  seed         bigint      NOT NULL,
  sim_version  text        NOT NULL,
  calendar     jsonb       NOT NULL,   -- the ten drawn circuits and their profiles
  field        jsonb       NOT NULL,   -- the 2026 grid: cars, drivers, staff
  created_at   timestamptz NOT NULL DEFAULT now()
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
  decisions   jsonb       NOT NULL,    -- the five picks: the only thing the client supplies
  points      int         NOT NULL,    -- computed server-side
  wins        int         NOT NULL,
  podiums     int         NOT NULL,
  dnfs        int         NOT NULL,
  result      jsonb       NOT NULL,    -- full race-by-race, for the share card
  created_at  timestamptz NOT NULL DEFAULT now(),
  UNIQUE (season_id, player_id)
);

CREATE INDEX runs_alltime ON runs (points DESC, wins DESC, podiums DESC, created_at ASC);
CREATE INDEX runs_by_player ON runs (player_id);
```

`UNIQUE (season_id, player_id)` is the entire anti-resubmission mechanism — enforced by the database
rather than by application logic that can race. It is also what makes unlimited play safe: the same
season cannot be banked twice, and playing again issues a new one.

`seasons.sim_version` lets the rules change without invalidating history. Old seasons are frozen,
never recomputed.

## Observability

- **Logging:** `log/slog` with the JSON handler. Every request carries a request ID; every
  submission logs season ID, player ID, verification duration, and outcome.
- **Health:** `/healthz` returns 200 only after a successful Postgres ping.
- **Metrics:** `/metrics` in Prometheus format — `http_request_duration_seconds` by route and
  status, `sim_verification_duration_seconds`, `runs_submitted_total` by outcome,
  `seasons_issued_total`.

Verification duration is the metric worth watching: CPU-bound work on the request path. If it creeps
toward hundreds of milliseconds, move verification to a worker queue. Instrumenting from day one is
what makes that a measurement rather than a guess.

## Deployment

One multi-stage Dockerfile: a builder stage compiling the WASM module, the native binary and the
frontend, then a `scratch` final stage holding one static binary with assets embedded (~20–25 MB).

**AWS App Runner** over ECS Fargate for the MVP — TLS, scaling and image deploys with far less
configuration, and nothing here needs what Fargate adds. RDS Postgres on `db.t4g.micro`. Migrating
to Fargate later is a day's work, because the container is already the deployable unit.
