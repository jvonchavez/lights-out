# Lights Out

A daily browser game where you run an F1 team through a 10-race season. The racing is simulated —
you never drive. Your only decisions happen *between* races: split a fixed development budget across
chassis, engine, and aero, knowing that every point spent on performance costs reliability, and a
reliability failure scores zero.

Everyone in the world gets the same season seed each day. Same calendar, same starting ratings, same
rival AI, same random stream. The luck is identical for everyone, so the leaderboard measures
decisions and nothing else.

```
Lights Out · Season 454
P3 · 120 pts · 1 DNF
🏁🥉🥈🏁🏁🥉🏁🏁🏁✖️
```

## The detail worth leading with

**One simulation compiles to two targets.** `internal/sim` is a pure Go package with no I/O, no
clock, and no unseeded randomness. It compiles natively into the server binary, where it computes
the authoritative leaderboard score, and to WebAssembly, where the browser plays the whole season
locally with no network round-trips. There is one implementation, so the client and server cannot
disagree about the rules — and a parity test asserts both targets produce byte-identical JSON across
3,000 seeds.

That is what makes the leaderboard trustworthy. The client posts *only its ten decisions* — never a
score. `submitRequest` has no score field at all, so a forged number has nowhere to be decoded to
and is discarded before any code could believe it. The server re-runs those decisions from the
season's own seed. A cheater would have to find decisions that genuinely produce a high score, which
is just playing well.

**Determinism is structural, not hoped for.** Every number in the sim is `int64` fixed-point, and
the normal distribution comes from a sum of twelve uniforms rather than Box–Muller. Go's assembly
implementations of `log`, `exp`, and `cos` differ from the pure-Go versions used on `js/wasm`, and
`math/rand`'s `NormFloat64` walks straight into that. Integer arithmetic is bit-identical on every
target by construction, so the parity test passed on its first run with nothing to debug. The RNG is
a vendored PCG64 rather than a stdlib one, because a leaderboard is only trustworthy if today's seed
means the same thing next year.

## Running it

Requires Go 1.24+, Node 20+, and a container runtime (Docker Desktop, or Colima via
`brew install colima docker && colima start`).

```bash
make db        # Postgres in a container
make run       # builds the WASM module, the frontend, and the binary; serves on :8080
```

Then open http://localhost:8080.

```bash
make test              # unit, property and golden tests — no Docker needed
make test-integration  # real Postgres in a container, real migrations
make parity            # native vs js/wasm over 3,000 seeds
make test-e2e          # Playwright: play a season and submit
make balance           # 100k seasons across a worker pool
make simulate SEED=2026 STRATEGY=aerofirst   # play a season in the terminal
```

## How it fits together

```
Browser ── React + TS ──▶ sim.wasm (Go)      plays the season locally
   │
   │  POST /api/runs  { season_id, decisions[] }      ← never a score
   ▼
Go binary ── internal/sim (same package, native) ──▶ authoritative result
   │         daily seed scheduler · embedded frontend (embed.FS)
   ▼
Postgres ── seasons · players · runs
```

One binary carries the API, the scheduler, and the frontend. `UNIQUE (season_id, player_id)` is the
entire anti-resubmission mechanism — enforced by the database rather than by application logic that
can race. `UNIQUE (published_at)` on a `date` column makes the daily scheduler idempotent, so
several instances can run it with no leader election.

## What the balance runs found

`cmd/balance` runs 100,000 seasons across a worker pool (~115k seasons/sec on 8 cores) and reports
how each scripted strategy fares. It changed the game design four times — the reliability model was
mathematically degenerate, a whole budget slider turned out to be a trap and was cut, the player had
no ceiling, and aero had to multiply rather than add. `docs/Game Design.md` records each finding,
the measurement behind it, and the original specification it replaced.

Final spread over 100k seasons — no strategy wins more than 18.9%, and the gradient from 0% to
18.9% is wide enough that decisions plainly matter:

| Strategy | Titles | Avg pts | Avg finish |
|---|---|---|---|
| aero to the cap early, then the calendar | 18.9% | 115.6 | 4.50 |
| follow the next two circuits | 13.8% | 107.1 | 4.91 |
| equal thirds | 13.0% | 106.0 | 4.95 |
| everything into engine | 1.6% | 71.3 | 7.18 |
| chassis and engine, front-loaded, no aero | 0.0% | 38.4 | 9.69 |
| spend nothing | 0.0% | 20.3 | 10.84 |

## Design documents

`docs/` holds the design that preceded the code: the game rules and the numbers, the architecture
and data model, every technology choice with the alternative it rejected, and the milestone plan.
They are kept current — `Game Design.md` was revised from the balance data rather than left as
written.

## Not built yet

Deployment (Dockerfile to `scratch`, GitHub Actions, App Runner + RDS) and the write-up. Out of
scope by design: driver transfers, multi-season careers, pit strategy, tyre compounds, weather, and
accounts. Player identity is a UUID in `localStorage` — no passwords means no reset flow, no email,
and no PII.
