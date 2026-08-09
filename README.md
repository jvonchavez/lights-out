# Lights Out

A daily browser game where you build a Formula 1 team out of history and then watch it race.

The draft is five rolls. Each one lands on a real team-season — 1988 McLaren, 2004 Ferrari, 1967
Lotus — and from it you take exactly **one** thing: their car, one of their two drivers, their race
engineer, or their team principal. Five rolls, five slot-places, no passing.

The scarcity needs no budget system. A great team-season has a great car *and* a great driver *and*
a great principal, and you get one of them. Roll 1988 McLaren and take the MP4/4, and Senna goes to
somebody else.

Then your team races the **real 2026 grid** — eleven teams, their actual cars, drivers, engineers
and principals — over ten rounds, and you find out whether it wins the Constructors' Championship.
Both of your drivers score for it, which is why the second seat is a decision rather than a
passenger.

```
Lights Out · Season 454
P4 of 12 · 124 pts · 2 DNFs
🏁🏁🏁🏁🥉🏁🏁🏁🥈🏁
```

Five clicks and about a minute, or forty-five seconds more if you watch the whole reel. Everyone in
the world gets the same seed each day: same calendar, same five rolls, same random stream, and the
same opposition — the field is the actual current grid rather than anything generated, so no day is
easier than another and the leaderboard measures your draft and nothing else. Free play gives
unlimited unscored runs, because one season a day is not replayability.

## The detail worth leading with

**One simulation compiles to two targets.** `internal/sim` is a pure Go package with no I/O, no
clock, and no unseeded randomness. It compiles natively into the server binary, where it computes
the authoritative leaderboard score, and to WebAssembly, where the browser plays the whole season
locally with no network round-trips. There is one implementation, so the client and server cannot
disagree about the rules — and a parity test asserts both targets produce byte-identical JSON across
3,000 seeds.

That is what makes the leaderboard trustworthy. The client posts *only its five pick indices* —
never a score. `submitRequest` has no score field at all, so a forged number has nowhere to be
decoded to and is discarded before any code could believe it. The server re-derives the rolls from
the season's own seed and replays the picks itself, so scores cannot be fabricated.

The legality rule lives in the sim rather than the API, so the browser greys out exactly the items
the server would refuse. `[0,0,0,0,0]` is five in-range indices and five cars; `[1,2,1,3,4]` is
three drivers and no car. Both are 400s — a check a card draft, with only one kind of slot, could
not express.

Scores can still be searched for, though, and the docs say so: **240 legal drafts** is trivially
enumerable when the simulation runs in your browser, and enumerating them wins the championship 40%
of the time against 17.7% for the best scripted line. Wordle is solvable too. The honest claim is
that the server is the sole authority on what a set of picks is worth — not that nobody can run a
solver.

**Determinism is structural, not hoped for.** Every number in the sim is `int64` fixed-point — the
roster's 0–99 ratings included — and the normal distribution comes from a sum of twelve uniforms
rather than Box–Muller. Go's assembly
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
make test-e2e          # Playwright: draft a team, race it, submit
make balance           # 100k seasons across a worker pool
make simulate SEED=2026 STRATEGY=bestavailable   # play a season in the terminal
```

## How it fits together

```
Browser ── React + TS ──▶ sim.wasm (Go)      plays the season locally
   │                                          (free play never leaves here)
   │  POST /api/runs  { season_id, picks[] }         ← never a score
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

`cmd/balance` runs 100,000 seasons across a worker pool and reports how each scripted strategy
fares. It has changed the game design seven times, and `docs/Game Design.md` records each finding,
the measurement behind it, and the specification it replaced — including the ones that did not work.

| Strategy | Titles | Avg pts | Avg finish |
|---|---|---|---|
| take the highest-rated thing on offer | 17.7% | 228.0 | 2.64 |
| value finishing over pace | 10.4% | 167.1 | 3.56 |
| fill both driver seats first | 10.1% | 180.9 | 3.24 |
| always take the car | 7.7% | 138.5 | 3.94 |
| take the leftmost legal item | 5.3% | 127.1 | 4.10 |

Enumerating all 240 legal drafts wins 40%, so the gradient runs 5.3 → 17.7 → 40 and the draft is
measuring real thought rather than none.

**The two most interesting findings are both about greed.** `starpower` ends up with *worse* drivers
than `carfirst` — 35288 rating against 35895 — because filling both driver seats from the first two
rolls forfeits the choice later rolls would have offered; and `carfirst`, which always takes the car
on roll one, is the second-worst line in the game. Grabbing the thing you want first is not the same
as ending up with the best of it.

**And two attempts that failed are recorded rather than deleted.** Seven of eleven rivals average
under six points a season, and the cause is not the ratings: compressing the grid toward the leader
by 35% and 55% moved Racing Bulls from 7.8 points to 7.8 and 7.4, and raising the midfield's own
ratings by 6–10 points moved it from 5.4 to 5.0. The top four teams field eight cars and only ten
positions score, so P1–P8 are spoken for. The speculative raise was reverted, because it bought
nothing and made the grid less faithful to the real 2026 standings — where the season is, in fact, a
runaway.

## Design documents

`docs/` holds the design that preceded the code: the game rules and the numbers, the architecture
and data model, every technology choice with the alternative it rejected, and the milestone plan.
They are kept current, and superseded designs are preserved in place rather than deleted —
`Game Design.md` still carries the three-slider budget split and the five-window card draft that
came before this one, each with the measurements that replaced it.

## Not built yet

Deployment (Dockerfile to `scratch`, GitHub Actions, App Runner + RDS) and the write-up. Out of
scope by design: mid-season transfers, multi-season careers, pit strategy, tyre compounds, weather,
and accounts. Player identity is a UUID in `localStorage` — no passwords means no reset flow, no email,
and no PII.
