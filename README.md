# Lights Out

A browser game where you build a Formula 1 team out of history and then watch it race.

The draft is five rolls. Each lands on a real team-season — 1988 McLaren, 2004 Ferrari, 1967 Lotus —
and from it you take exactly **one** thing: their car, one of their two drivers, their race engineer,
or their team principal. Five rolls, five slot-places, no passing. Take the MP4/4 from 1988 McLaren
and Senna goes to somebody else; nothing has a price, and yet everything costs.

Your team then races the **real 2026 grid** — eleven teams, their actual cars, drivers, engineers and
principals — over ten real circuits, and you find out whether it wins the Constructors' Championship.
Both of your drivers score for it, which is why the second seat is a decision rather than a
passenger.

```
Lights Out · Season 454
P4 of 12 · 124 pts · 2 DNFs
🏁🏁🏁🏁🥉🏁🏁🏁🥈🏁
```

Five clicks and about a minute. Then play again: the server issues a new season on request, with a
fresh seed, a fresh calendar drawn from 34 real circuits, and five fresh rolls out of 52 team-eras.

The one thing that never varies is the opposition, so no season is easier than another and a score
measures your draft alone. The leaderboard is all-time and keeps your best season, with the number of
seasons you played shown beside it — when play is unlimited, a best-of-N is partly a measure of N.

## One simulation, two targets

`internal/sim` is a pure Go package: no I/O, no clock, no unseeded randomness. It compiles natively
into the server binary, where it computes the authoritative score, and to WebAssembly, where the
browser plays the whole season locally with no network round-trips. One implementation, so client and
server cannot disagree about the rules — and a parity test asserts both targets produce
byte-identical JSON across 3,000 seeds.

That is what makes the leaderboard trustworthy. **The server mints the seed**: a client that could
nominate its own could nominate one it had solved offline. The client posts *only its five pick
indices*, never a score — `submitRequest` has no score field at all, so a forged number has nowhere
to be decoded to. The server replays the picks against the seed it issued.

The legality rule lives in the sim rather than the API, so the browser greys out exactly what the
server would refuse. `[0,0,0,0,0]` is five in-range indices and five cars; `[1,2,1,3,4]` is three
drivers and no car. Both are 400s.

Scores can still be *searched* for: 240 legal drafts is trivially enumerable when the simulation runs
in your browser, and enumerating them wins the championship 40% of the time against 17.7% for the
best scripted line. The honest claim is that the server is the sole authority on what a set of picks
is worth — not that nobody can run a solver.

**Determinism is structural, not hoped for.** Every number is `int64` fixed-point, and the normal
distribution comes from a sum of twelve uniforms rather than Box–Muller — Go's assembly `log`, `exp`
and `cos` differ from the pure-Go versions used on `js/wasm`, and `math/rand`'s `NormFloat64` walks
straight into that. Integer arithmetic is bit-identical on every target by construction, so the
parity test passed on its first run with nothing to debug. The RNG is a vendored PCG64, because a
leaderboard is only trustworthy if a seed means the same thing next year.

## Running it

Requires Go 1.27+, Node 20+, and a container runtime (Docker Desktop, or Colima via
`brew install colima docker && colima start`).

```bash
make db        # Postgres in a container
make run       # builds the WASM module, the frontend, and the binary; serves on :8080
```

Use `make run` rather than `go build ./cmd/server` — the WASM module is a separate build target, and
a server rebuilt without it serves a client running a different ruleset from the verifier.

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
   │
   │  POST /api/seasons                             → server mints the seed
   │  POST /api/runs  { season_id, picks[] }         ← never a score
   ▼
Go binary ── internal/sim (same package, native) ──▶ authoritative result
   │         embedded frontend (embed.FS)
   ▼
Postgres ── seasons · players · runs
```

One binary carries the API and the frontend. `UNIQUE (season_id, player_id)` is the entire
anti-resubmission mechanism — enforced by the database rather than application logic that can race —
and it is also what makes unlimited play work: you cannot submit the same season twice, and playing
again issues a new one. There is no scheduler and no leader election, because a season is no longer a
day.

## What the balance runs found

`cmd/balance` runs 100,000 seasons across a worker pool and reports how each scripted strategy fares.
It has changed the game design repeatedly; `docs/Game Design.md` records the findings, including the
ones that did not work.

| Strategy | Titles | Avg pts | Avg finish |
|---|---|---|---|
| take the highest-rated thing on offer | 17.7% | 228.0 | 2.64 |
| value finishing over pace | 10.4% | 167.1 | 3.56 |
| fill both driver seats first | 10.1% | 180.9 | 3.24 |
| always take the car | 7.7% | 138.5 | 3.94 |
| take the leftmost legal item | 5.3% | 127.1 | 4.10 |

Enumerating all 240 legal drafts wins 40%, so the gradient runs 5.3 → 17.7 → 40 and the draft is
measuring real thought rather than none.

**Greed is the interesting finding, twice over.** `starpower` ends up with *worse* drivers than
`carfirst` — 35288 rating against 35895 — because filling both seats from the first two rolls
forfeits the choice later rolls would have offered; and `carfirst`, which always takes the car on
roll one, is the second-worst line in the game. Grabbing the thing you want first is not the same as
ending up with the best of it.

**Two failed fixes are recorded rather than deleted.** Seven of eleven rivals average under six points
a season, and the cause is not the ratings: compressing the grid toward the leader by 35% and 55%
moved Racing Bulls from 7.8 points to 7.8 and 7.4, and raising the midfield's own ratings by 6–10
points moved it from 5.4 to 5.0. The top four teams field eight cars and only ten positions score, so
P1–P8 are spoken for. The speculative raise was reverted: it bought nothing and made the grid less
faithful to the real 2026 standings, where the season is in fact a runaway.

## Docs

`docs/` holds the design: the rules and numbers (`Game Design.md`), the architecture and data model,
every technology choice with the alternative it rejected, and the milestone plan. The gameplay is
still being revised, so these track the current design rather than its history — superseded designs
live in git.

## Not built yet

Deployment (Dockerfile to `scratch`, GitHub Actions, App Runner + RDS) and the write-up. Out of scope
by design: mid-season transfers, multi-season careers, pit strategy, tyre compounds, weather, and
accounts. Player identity is a UUID in `localStorage` — no passwords means no reset flow, no email,
and no PII.
