# Lights Out

A daily browser game where you run an F1 team through a 10-race season. The racing is simulated —
you never drive. Five times across the season you are dealt three parts and take one. It locks, it
fits for the rest of the year, and every part you bolt on costs reliability. A failure scores zero,
and there are only ten races.

Everyone in the world gets the same season seed each day. Same calendar, same starting ratings, same
rival AI, same random stream. The luck is identical for everyone, so the leaderboard measures
decisions and nothing else.

```
Lights Out · Season 454
P3 of 11 · 120 pts · 1 DNF
🏁🥉🥈🏁🏁🥉🏁🏁🏁✖️
```

Five clicks, about a minute. The whole calendar is visible from the first window, so you can see the
power circuits coming and take the gearbox early — or gamble that something better comes round.

## The detail worth leading with

**One simulation compiles to two targets.** `internal/sim` is a pure Go package with no I/O, no
clock, and no unseeded randomness. It compiles natively into the server binary, where it computes
the authoritative leaderboard score, and to WebAssembly, where the browser plays the whole season
locally with no network round-trips. There is one implementation, so the client and server cannot
disagree about the rules — and a parity test asserts both targets produce byte-identical JSON across
3,000 seeds.

That is what makes the leaderboard trustworthy. The client posts *only its five card indices* —
never a score. `submitRequest` has no score field at all, so a forged number has nowhere to be
decoded to and is discarded before any code could believe it. The server re-derives the deal from
the season's own seed and replays the picks itself, so scores cannot be fabricated.

They can be searched for, though, and the docs say so: 243 possible lines is trivially enumerable
when the simulation runs in your browser. Wordle is solvable too. The honest claim is that the
server is the sole authority on what a set of picks is worth — not that nobody can run a solver.

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

`cmd/balance` runs 100,000 seasons across a worker pool (~110k seasons/sec on 8 cores) and reports
how each scripted strategy fares. It has changed the game design five times — the reliability model
was mathematically degenerate, a whole budget slider turned out to be a trap and was cut, the player
had no ceiling, aero had to multiply rather than add, and the card draft needed the aero cap more
than doubled before understanding the game beat simply buying the biggest part. `docs/Game Design.md`
records each finding, the measurement behind it, and the specification it replaced.

Final spread over 100k seasons — no strategy wins more than 20.8%, and the gradient from 4.2% to
20.8% is wide enough that decisions plainly matter:

| Strategy | Titles | Avg pts | Avg finish |
|---|---|---|---|
| buy aero to the cap, then read the calendar | 20.8% | 124.7 | 3.80 |
| always take the costliest part | 14.3% | 108.4 | 4.80 |
| take what suits the next two races | 10.4% | 96.6 | 5.54 |
| always take the leftmost card | 8.1% | 90.4 | 5.91 |
| always take the cheapest part | 4.2% | 72.0 | 7.14 |

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
