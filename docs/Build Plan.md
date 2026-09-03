# Lights Out — Build Plan

Seven milestones, each with a checkable definition of done. The ordering front-loads the simulation:
it is both the product and the riskiest thing to get wrong, and everything downstream is plumbing.

**M0–M4 are done.** The input surface has been rebuilt since — sliders, then a card draft, then the
pre-season team draft the game now is — and the daily seed was replaced by unlimited play. Each
rebuild kept the simulation's machinery intact: fixed-point arithmetic, vendored PCG64, purity test,
golden fixtures, parity guarantee. Only the meaning of a decision changed, which is the evidence the
boundary between "the rules" and "the input surface" was drawn in the right place.

---

## M0 — Sim engine, no server ✅

`internal/sim` as a pure package plus a `cmd/simulate` CLI.

- `RunSeason(seed, picks)` with no I/O anywhere in the package
- Explicitly seeded RNG threaded through; no package-level randomness, no map-iteration dependence
- Circuit profiles, qualifying, race resolution, reliability, safety car, points
- Table-driven unit tests per subsystem; golden fixtures in `testdata/`

**Done when:** `go test ./internal/sim/...` passes, and one seed run a thousand times produces one
distinct result.

## M1 — Balance ✅

`cmd/balance`: 100k seasons across a worker pool (`errgroup` over `runtime.NumCPU()`), several
scripted strategies, CSV out.

**Done when:** no scripted strategy wins more than ~35% of seasons, and the σ values come from this
data rather than a guess.

## M2 — Persistence ✅

Postgres schema and `golang-migrate` migrations, `sqlc` queries, Docker Compose, `testcontainers-go`
integration tests against real Postgres and real migrations.

**Done when:** `docker compose up` gives a working database and integration tests pass from cold.

## M3 — API and verification ✅

Routes: `POST /api/seasons` (the only place a seed enters the system), `GET /api/seasons/{id}`,
`POST /api/runs`, `GET /api/leaderboard`. Server-side verification re-runs the submitted picks
natively and ignores anything client-supplied. Rate limiting on submission by IP.

**Done when:** a submission with a hand-edited score is scored on its picks and the forged number is
ignored — proven by a test. That test now also asserts five in-range indices forming an illegal team
are refused.

## M4 — WASM and frontend ✅

The longest milestone and the one carrying the architectural centrepiece.

- `GOOS=js GOARCH=wasm` build of `internal/sim` with a thin `syscall/js` binding layer
- **Parity test before any UI.** It passed on its first run with nothing to debug — the sim was
  already all `int64` fixed-point with Irwin-Hall normals. Rerun after every rules change since,
  never failed.
- React + TypeScript + Vite: draft screen, race reel (ten rounds in ~45 seconds with a constructors
  bar chart climbing behind it), constructors and drivers tables, share card, leaderboard
- `embed.FS` to serve the built frontend from the Go binary
- Playwright test covering play-through and submission

**Done when:** the parity test passes and a full season can be played and submitted in a browser.

## M5 — Deploy and observability

- Multi-stage Dockerfile producing a `scratch` image
- `slog` JSON logging with request IDs; `/healthz` gated on a Postgres ping; `/metrics`
- GitHub Actions: test, build, push to ECR, deploy
- App Runner plus RDS, both deployed by hand once so the resources are understood; budget alarm

**Done when:** the game is live at a URL, `/metrics` shows real traffic, and a push to `main`
deploys.

## M6 — Write-up

Why determinism makes anti-cheat trivial; one simulation compiled to two targets and the parity
problem that creates; what 100k simulated seasons revealed about balance; the observability
decisions.

**Done when:** published and linked from the game.

---

## Sequencing notes

- **M1 before M2.** Balance findings change the rules, and changing rules after a schema exists means
  migrations and invalidated fixtures.
- **Parity test before the UI in M4.** Finding a native/WASM divergence after building a frontend
  against the WASM build means debugging two things at once.
- **Deploy by hand before writing Terraform.** IaC written before you understand the resources is IaC
  you cannot debug. Terraform stays a stretch goal.

## Risks

- **Balance never lands.** If no settings make decisions matter without fully determining outcomes,
  no amount of engineering fixes it. *Outcome:* landed three times on three different games, each
  time because the harness rather than intuition found the problem — including two fixes that did not
  work (σ for the midfield, and the midfield lockout, which is structural).
- **WASM bundle size.** ~2 MB gzipped is tolerable; materially worse is not. Fallback is TinyGo,
  evaluated against a measured number.
- **Scope creep into a career mode.** 52 team-eras and 34 circuits look like the start of one. They
  are not. Unlimited single seasons are the answer to "I want another go", and removing the daily
  seed added no progression, unlocks, or state surviving a run. A season is still a pure function of
  a seed and five integers.
