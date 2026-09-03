# Lights Out — F1 Season Manager

**Started:** 2026-07-28 · **Language:** Go + TypeScript

Build an F1 team out of history, then watch it race. Five rolls, each landing on a real team-season,
and from each you take exactly one thing — their car, one of their drivers, their race engineer, or
their team principal. Your team then races the real 2026 grid over ten real circuits. You never
drive; every decision happens before the season.

Play is unlimited: a fresh seed, calendar and five rolls on request. The only constant is the
opposition, so a score measures the draft and nothing else.

## Documents

| File | Contents |
|---|---|
| `Game Design.md` | Rules, numbers, balance findings, determinism contract |
| `Architecture.md` | Services, data flow, data model, deployment |
| `Tech Stack.md` | Every technology choice and the alternative rejected |
| `Build Plan.md` | Milestones M0–M6 with a definition of done for each |

## MVP

A player loads the site, is issued a season, drafts a team across five rolls, watches ten races
resolve in a reel, and lands on the all-time leaderboard with a server-verified score. Then plays
again.

**In scope:** 10-race calendar drawn from 34 circuits, 5 draft rolls over 52 real team-eras, the 2026
grid as a fixed 11-team field, 24 cars with both drivers scoring for the constructor, qualifying and
race simulation, mechanical and driver-error retirements, in-season development, standings,
server-issued seeds, verified all-time leaderboard, share string, unlimited play.

**Out of scope:** mid-season transfers, multi-season careers, pit strategy, tyre compounds, weather,
accounts. Player identity is a UUID in `localStorage` plus a display name — no passwords, no reset
flow, no email, no PII.

## Why this project

Targets the **Go gap (#2)** from `../../Job Applications/_INDEX.md` — 10 roles, the highest-return
technical investment in the set. Go is not decoration: the sim is a pure function, batch-executed
100k times across goroutines for balance tuning, compiled to a small static container, and compiled
*again* to WebAssembly so browser and server run byte-identical code.

Secondary coverage: automated testing (#1) via golden replay fixtures, observability (#3), Docker
(#4), AWS (#7). Python is deliberately absent — `../Tennis Vision/` carries that gap.

## What the verification claims

**Scores cannot be fabricated.** The server is the sole authority on what a set of picks is worth: it
re-derives the deal from the seed, replays the picks, and computes the score. The client sends
indices and nothing else, and `submitRequest` has no score field.

**They can be searched for.** Five rolls of five items is 240 legal drafts, and the WASM module runs
locally, so anyone can enumerate every line in milliseconds. Measured: enumerating all 240 wins the
championship **40%** of the time, against 17.7% for the best scripted strategy and 5.3% for the
worst. A solver is meaningfully better than a person.

**Unlimited play adds a second limit.** A player issued any number of seasons can reroll until the
teams on offer are good and only submit those. The seed comes from the server so it cannot be
*chosen* — but it can be declined. There is no fix that keeps play unlimited, so the board shows the
number of seasons submitted beside the best score. A 260-point season in three attempts is not the
same achievement as one in three hundred.

This is accepted. Wordle's answer space is ~2,300 words and solvers crack it instantly; the fun is
doing it yourself. Replayability comes from the roster — 52 team-eras means a different five nearly
every run. Nothing resists brute force when the simulation runs client-side, and hiding the rules
from the browser would cost the dual-target design that is the point of the project.
