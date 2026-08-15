# Lights Out — Build Plan

Seven milestones. Each has a definition of done that can be checked rather than felt. The ordering
front-loads the simulation, because the sim is both the product and the riskiest thing to get wrong —
everything downstream is plumbing that can be built with confidence once the rules are settled.

**Estimates assume evenings and weekends.** They are planning aids, not commitments.

> **M0–M4 are done, and then done again.** The input surface was rebuilt twice after M4: first the
> three budget sliders became a five-window card draft, then the card draft became the pre-season
> team draft the game now is. A third change followed: the fictional circuits became real ones and
> the daily seed was removed in favour of unlimited play. `docs/Game Design.md` records each rebuild
> and the measurements that forced it. The milestones below are written as originally specified;
> the notes at the end of each say what actually shipped.

---

## M0 — Sim engine, no server *(~4 evenings)*

Build `internal/sim` as a pure package plus a `cmd/simulate` CLI that runs a season from a seed and
prints the result.

- `RunSeason(seed int64, decisions []Decision) (SeasonResult, error)` with no I/O anywhere in the package
- Explicit seeded RNG threaded through; no package-level randomness, no map-iteration-order dependence
- Circuit profiles, qualifying, race resolution, reliability, safety car, points
- Rival AI archetypes
- Table-driven unit tests per subsystem
- Golden fixtures in `testdata/` — committed seed, decisions, and full expected result

**Done when:** `go test ./internal/sim/...` passes, and running the same seed a thousand times
produces one distinct result.

---

## M1 — Balance *(~3 evenings)*

The goroutine payoff, and the milestone most likely to change the game design.

Build `cmd/balance`: run 100k seasons across a worker pool, with rival archetypes and several scripted
player strategies (all-in early, even spread, specialist, reliability-first), and report win rates and
DNF rates per strategy.

- `errgroup` over `runtime.NumCPU()` workers, results collected over a channel
- Output a CSV; analyse it in whatever is quickest

**Done when:** no single scripted strategy wins more than ~35% of seasons, and the qualifying and race
σ values are chosen from this data rather than guessed. **Expect to change `Game Design.md` here** — if
reliability investment turns out to be a trap option nobody should ever pick, cut it.

---

## M2 — Persistence *(~2 evenings)*

- Postgres schema and `golang-migrate` migrations
- `sqlc` queries for seasons, players, runs, and the leaderboard
- Docker Compose: Postgres plus the service
- `testcontainers-go` integration tests against real Postgres and real migrations

**Done when:** `docker compose up` gives a working database, and integration tests pass from a cold
start with no manual setup.

---

## M3 — API and verification *(~3 evenings)*

- `GET /api/seasons/today`, `POST /api/runs`, `GET /api/seasons/{id}/leaderboard`
- Server-side verification: re-run submitted decisions natively, ignore any client-supplied score
- Validation: decision count, budget bounds, season open, one submission per player per season
- Daily seed scheduler goroutine, idempotent against the `UNIQUE (published_at)` constraint
- Rate limiting on submission by IP

**Done when:** a submission with a hand-edited score in the payload is scored on its decisions and the
forged number is ignored — proven by a test, not by inspection.

**What shipped, after unlimited play.** The routes are `POST /api/seasons` (which mints the seed —
the only place one enters the system), `GET /api/seasons/{id}`, `POST /api/runs` and
`GET /api/leaderboard`. The scheduler goroutine and the `UNIQUE (published_at)` constraint are both
gone; "season open" is not a validation any more because a season never closes. What survived
unchanged is the part that mattered: the forged-score test, which now also asserts that five
in-range indices forming an illegal team are refused.

---

## M4 — WASM and frontend *(~5 evenings)*

The longest milestone and the one carrying the architectural centrepiece.

- `GOOS=js GOARCH=wasm` build of `internal/sim`, with a thin `syscall/js` binding layer
- **Parity test first, before any UI**: several thousand seeds through both native and WASM builds,
  asserting identical results. If floating-point divergence appears, fix it with fixed-point or
  integer arithmetic in the affected calculation — do not proceed with a known discrepancy
- React + TypeScript + Vite: season view, budget allocation, race results, standings, share card
- `embed.FS` to serve the built frontend from the Go binary
- Playwright test covering play-through and submission

**Done when:** the parity test passes, and a full season can be played and submitted in a browser.

**What shipped, after two rebuilds of the input surface.** "Budget allocation" is gone twice over.
The frontend is a draft screen (a team-season and the five things you can take from it), a race reel
that plays ten rounds in about forty-five seconds with a constructors bar chart climbing behind it,
and the result — constructors table, drivers table, share card, leaderboard. A play-again button
was added here, and later became the whole shape of the game when the daily seed was removed.

The parity test did come first, and it passed on its first run with nothing to debug — the sim was
already all `int64` fixed-point and Irwin-Hall normals, so the roster's 0–99 ratings introduced no
new floating point. It has since been rerun after every rules change and has never failed.

---

## M5 — Deploy and observability *(~3 evenings)*

- Multi-stage Dockerfile producing a `scratch` image
- `slog` JSON logging with request IDs; `/healthz` gated on a Postgres ping; `/metrics` with request
  duration, verification duration, and submission counters
- GitHub Actions: test, build, push to ECR, deploy
- App Runner service plus RDS Postgres, both deployed by hand once so the resources are understood
- AWS budget alarm

**Done when:** the game is live at a URL, `/metrics` shows real traffic, and a push to `main` deploys.

---

## M6 — Write-up *(~2 evenings)*

The milestone most likely to be skipped and most likely to matter. `_INDEX.md` notes repeatedly that
framing separates a line item from a differentiator.

Cover: why determinism makes anti-cheat trivial; compiling one simulation to two targets and the
parity problem that creates; what 100k simulated seasons revealed about the balance; and the
observability decisions.

**Done when:** it is published and linked from the game.

---

## Rough total

**~22 evenings**, call it 5–7 weeks at a realistic pace alongside a full-time job. M4 is the most
likely to overrun — and it did, twice, because the input surface was rebuilt after it rather than
during it. Both rebuilds kept the simulation's machinery entirely: fixed-point arithmetic, the
vendored PCG64, the purity test, the golden fixtures and the parity guarantee were untouched each
time. Only the meaning of a decision changed. That is the clearest evidence the boundary between
"the rules" and "the input surface" was drawn in the right place.

## Sequencing notes

**M1 before M2 is deliberate.** Balance findings can change the game rules, and changing rules after a
schema exists means migrations and invalidated fixtures. Settle the simulation before persisting
anything derived from it.

**The parity test comes before the UI in M4**, not after. Discovering a native/WASM divergence after
building a frontend against the WASM build means debugging two things at once.

**Deploy by hand before writing Terraform.** IaC written before you understand the resources it
creates is IaC you cannot debug. Terraform stays a stretch goal, and an honest one — it is better to
have no Terraform than Terraform you cannot explain.

## What would make this fail

- **Balance never lands.** If the sweeps cannot find settings where decisions matter but do not fully
  determine outcomes, the game is not fun and no amount of engineering fixes that. Mitigation: treat
  it as a real gate. *Outcome:* it landed three times, on three different games — sliders, cards, and
  the team draft — and each time the harness rather than intuition found the problem. It also caught
  two fixes that did not work: σ is the wrong lever for the midfield, and the midfield lockout is
  structural rather than a rating error. Both reverts are recorded.
- **WASM bundle size.** ~2 MB gzipped is tolerable; materially worse is not. Fallback is TinyGo,
  evaluated only against a measured number.
- **The daily seed was load-bearing, and then it was not.** M2 and M3 were built around one
  season a day: a scheduler, a date-keyed uniqueness constraint, a per-season leaderboard, and a
  closing time. Unlimited play removed all four. *Outcome:* the removal was cheap because none of
  it had reached into the simulation — `internal/sim` never knew what a day was. The cost was one
  migration and one honest paragraph about what a best-of-N board can and cannot measure.
- **Scope creep into a career mode.** Multi-season progression is the obvious next feature and would
  double the project. It is out of scope. The team draft makes the temptation worse, not better — a
  roster of 52 team-eras and 34 circuits looks like the start of a career mode. It is not one, and
  unlimited single seasons are the answer to "I want another go". Note what removing the daily seed
  did NOT do: it did not add progression, unlocks, or state that survives a run. A season is still
  a pure function of a seed and five integers.
