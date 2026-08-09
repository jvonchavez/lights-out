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
          │ POST /api/runs  { season_id, picks[] }
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

`RunPartial` is gone. It existed to resolve the races each in-season development window unlocked,
and there are no in-season decisions left: the team is locked in before round one and the whole
season resolves at once. The client's race-by-race reveal is presentation over a result it already
holds.

**Three independent RNG streams**, all derived from the seed: `seed` for season generation,
`seed ^ 0x5EED` for race resolution, `seed ^ 0xCA4D` for draft rolls. Keeping them separate means
editing the roster cannot shift historical race outcomes, and changing generation cannot shift what
you are offered. Draw order is the determinism contract, and three streams keep it local.

**The rival field consumes no randomness at all.** It is the real 2026 grid, identical on every
seed, so `GenerateSeason` draws nothing but the calendar order. That deleted `rivals.go` outright —
there is no rival AI, no archetype table, and no draw-order surface where rival behaviour could
perturb race RNG.

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

**Playing.** The browser fetches today's season descriptor (seed, calendar, the 2026 field, and the
five rolls), loads the WASM module, and runs the entire season client-side once the draft is
complete. No network round-trip resolves anything. The rolls come from the API as well as from the
WASM module, so the draft renders before the ~1.25 MB module has finished loading — and the server
stays the authority on what was offered.

**Free play** needs no backend at all. `GenerateSeason` is pure and compiled into the WASM module,
so the client generates its own seed and simply never posts a run. Unlimited replays cost one page
of JavaScript and no request.

**Submitting.** The client posts `{ season_id, picks[] }` — five integers, never a score. The server
re-derives the rolls from the season's seed, calls `sim.RunSeason` natively with the submitted
picks, and computes the authoritative result. The client's own score is not sent, not read, and not
trusted; there is no field on the request struct for it to occupy.

**Verification** rejects a submission when: the pick count does not match the roll count, any pick is
outside the five items a roll offers, **the picks do not form a legal team**, the season is closed,
the season was published under a different sim version, or this player has already submitted for this
season. Everything the client could lie about is recomputed from the seed.

The legality check is new and is stronger than a card draft could express. `[0,0,0,0,0]` is five
in-range indices and five cars; `[1,2,1,3,4]` is three drivers and no car. Both are 400s, and the
rule lives in `BuildLineup` inside the sim, so the browser greys out the same items the server would
refuse.

What none of this does is make the game unsearchable: 240 legal drafts is trivially enumerable
client-side, and enumerating them wins the championship 40% of the time against 17.7% for the best
scripted line. See "What the verification does and does not claim" in `_README.md`; the honest claim
is that scores cannot be fabricated, not that they cannot be searched for.

**The sim version gate is enforced, not documented.** A season is verified under the version it was
published with. Replaying the same picks under changed rules produces a different score, so a
submission to a season from another ruleset is a 409 naming both versions — otherwise a deploy would
silently corrupt an open leaderboard mid-day.

## Data model

```sql
CREATE TABLE seasons (
  id           bigserial PRIMARY KEY,
  seed         bigint      NOT NULL,
  sim_version  text        NOT NULL,
  calendar     jsonb       NOT NULL,   -- circuits and their profiles
  field        jsonb       NOT NULL,   -- the 2026 grid: cars, drivers, staff
  published_at date        NOT NULL,
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
  decisions   jsonb       NOT NULL,    -- the five picks: the only thing the client supplies
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

The schema did not change for the gameplay rebuild. `runs.decisions` was an array of five ints
before and is an array of five ints now; only their meaning moved, from card indices to item
indices. There was no migration to write.

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
