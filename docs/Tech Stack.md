# Lights Out — Technology Choices

Each row states what was chosen, why, and what was rejected. The rejections matter more than the
picks — being able to say why you did *not* use something is what separates a decision from a default.

---

## Backend

| Concern | Choice | Why | Rejected |
|---|---|---|---|
| Language | **Go 1.23+** | The gap being targeted. Also genuinely correct here: goroutines for batch balance runs, a single static binary, and a WASM target from the same source | Rust (steeper, no resume gap points), C# (no gap closed) |
| HTTP router | **`net/http` + `ServeMux`** | Go 1.22's mux gained method and path-parameter matching. A router dependency is no longer justified at this scale | chi, gin, echo — all solving a problem the standard library now solves |
| Postgres driver | **`pgx/v5`** | The de facto Go Postgres driver. Native protocol, real connection pooling, proper `jsonb` and array support | `database/sql` + `lib/pq` (unmaintained, slower) |
| Query layer | **`sqlc`** | Generates typed Go from hand-written SQL. Keeps SQL visible — which is the skill worth showing — while removing the scanning boilerplate | GORM (hides SQL, fights you at the edges), raw `pgx` (fine, but the boilerplate adds up) |
| Migrations | **`golang-migrate`** | Plain up/down SQL files, runs in CI and on startup, no framework buy-in | Atlas (more capable than this needs), ORM auto-migration (never, in anything deployed) |
| Config | **Environment variables + `envconfig`** | Twelve-factor, works identically in Compose and App Runner | Viper and config files, for a service with roughly eight settings |
| Logging | **`log/slog`** | Standard library since 1.21. Structured JSON, no dependency | zerolog, zap — both fine, neither worth a dependency now |
| Metrics | **`prometheus/client_golang`** | The default, scrapeable anywhere | OpenTelemetry — better long-term, more setup than one service needs |

## Frontend

| Concern | Choice | Why | Rejected |
|---|---|---|---|
| Framework | **React 18 + TypeScript** | Already on the resume, so zero learning cost, and this project's purpose is the Go work | Svelte, Solid — smaller and arguably nicer, but they buy learning time this project should not spend |
| Build | **Vite** | Fast, sane defaults, trivial WASM asset handling | Next.js — SSR for a client-side game that ships its own runtime makes no sense |
| Styling | **Tailwind** | Fast iteration, no naming decisions, no CSS file sprawl | CSS modules, styled-components |
| Charts | **Hand-rolled SVG and CSS** | The visualisations are a standings table, a season sparkline and a constructors bar chart that grows as the reel plays. All three are a `div` with a width or a `polyline`; a charting library would be more code than the charts | Recharts, D3 |
| Animation | **CSS keyframes** | The race reel needs one staggered slide-in and one width transition, both with a `prefers-reduced-motion` escape. A 30 kB animation library for two rules is not a trade worth making | Framer Motion, GSAP |
| State | **`useState` + `useReducer`** | The whole game state is one object mutated by one reducer. That is precisely what `useReducer` is | Redux, Zustand, Jotai |

## Testing

| Layer | Tool | What it proves |
|---|---|---|
| Unit | stdlib `testing`, table-driven | Each subsystem — points allocation, reliability, qualifying order — behaves across enumerated cases |
| Golden fixtures | stdlib + `testdata/*.json` | A committed seed and decision list still produces the exact committed result. **Any unintended rules change fails loudly** |
| Property-based | **`pgregory.net/rapid`** | Invariants over generated input: points are never negative, championship totals equal the sum of race points, every allocation respects its budget, `RunSeason` is deterministic across 1000 repeats |
| Parity | custom harness | Native and WASM builds produce identical output over several thousand seeds. Guards the architectural centrepiece |
| Integration | **`testcontainers-go`** | Real Postgres in Docker, real migrations, real queries. No mocked database |
| E2E | **Playwright** | One test: load the page, play a season, submit, appear on the leaderboard |

The golden-fixture and property-based layers together are the direct answer to the automated-testing
gap (#1) — the highest-frequency gap in `../../Job Applications/_INDEX.md`, named in 12+ folders, and
currently evidenced only by Postman validation of 30+ endpoints.

## Infrastructure

| Concern | Choice | Why | Rejected |
|---|---|---|---|
| Container | **Multi-stage Dockerfile → `scratch`** | One static binary with assets embedded. ~20–25 MB image, no OS, no shell, minimal attack surface | Alpine or distroless — sensible generally, unnecessary for a `CGO_ENABLED=0` binary |
| Local dev | **Docker Compose** | Postgres plus the service, one command, matches CI | — |
| Hosting | **AWS App Runner** | TLS, scaling, and image deploys with almost no configuration. Nothing here needs more | ECS Fargate (more knobs, none needed yet), Lambda (poor fit for a WASM-serving stateful-ish service), EC2 (managing servers is not the lesson) |
| Database | **RDS Postgres, `db.t4g.micro`** | Managed backups and a real Postgres, cheaply | Aurora Serverless (overkill and pricier at idle), self-hosted (no) |
| CI | **GitHub Actions** | Test, build both targets, push to ECR, deploy | — |
| IaC | **Terraform** *(stretch)* | Real, but only after the thing is deployed by hand once. Writing IaC before understanding the resources produces IaC you cannot debug | CDK, Pulumi |

## Deliberately not used

**Kubernetes.** One stateless container and one database. Adding K8s would be resume theatre, and
anyone competent enough to be worth impressing would recognise it as such. The Docker/K8s gap (#4)
stays partially open, and that is the correct trade.

**Redis.** No caching need. Postgres serves a leaderboard of a few thousand rows behind an index in
under a millisecond.

**Kafka or any broker.** No asynchronous work. `../Tennis Vision/` covers the event-driven gap (#8)
honestly, with a queue that exists because the workload genuinely requires one.

**An auth provider.** A UUID in `localStorage` plus a display name. No passwords means no reset flow,
no email deliverability, no session management, and no PII. Losing your device loses your streak, and
for a daily browser game that is a reasonable trade rather than a defect.

**Any LLM.** Nothing here would be improved by one. The LLM gap (#6) needs its own project rather than
a chatbot bolted onto a racing game.
