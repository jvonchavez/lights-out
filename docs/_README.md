# Lights Out — F1 Season Manager

**Status:** planning · **Started:** 2026-07-28 · **Language:** Go + TypeScript

A browser game where you run an F1 team through a 10-race season. The racing itself is simulated —
you never drive. Your only decisions happen *between* races: split a fixed development budget across
chassis, engine, and aero, knowing that every point spent on performance costs reliability, and a
reliability failure scores zero.

Everyone in the world gets the same season seed each day. Same calendar, same starting ratings, same
rival AI, same random stream. The luck is identical for everyone, so the leaderboard measures
decisions and nothing else.

---

## Documents

| File | Contents |
|---|---|
| `Game Design.md` | The rules, the numbers, and why the central decision is interesting |
| `Architecture.md` | Services, data flow, data model, deployment topology |
| `Tech Stack.md` | Every technology choice, its justification, and the alternative rejected |
| `Build Plan.md` | Milestones M0–M6 with a definition of done for each |

---

## MVP definition

A player loads the site, sees today's season, plays through 10 races making one budget allocation
before each, and lands on the leaderboard against everyone else who played the same seed. Their score
is verified server-side before it counts.

**In scope:** 10-race calendar, 3 development sliders, 10 rival teams with their own AI development
strategies, qualifying and race simulation, reliability failures, championship standings, daily seed,
verified leaderboard, shareable result string.

**Out of scope:** driver transfers, multi-season careers, pit strategy, tyre compounds, weather,
accounts, and anything resembling login. Player identity is a UUID in `localStorage` and a display
name. No passwords means no password reset, no email, and no account-deletion flow.

---

## Why this project

Targets the **Go gap (#2)** from `../../Job Applications/_INDEX.md` — 10 roles, and the highest-return
technical investment in the set. Go is not decoration here: the sim runs as a pure function, gets
batch-executed 100k times across goroutines for balance tuning, compiles to a ~15 MB container, and
compiles *again* to WebAssembly so the browser and the server run byte-identical simulation code.

Secondary coverage: automated testing (#1) through golden replay fixtures, observability (#3) through
structured logging and metrics, Docker (#4), and AWS (#7).

Python is deliberately absent. `../Tennis Vision/` carries that gap, and splitting each project across
two languages would make both worse.

## The one detail worth leading with

The sim engine compiles to two targets from one source: a native binary for the server-side verifier,
and WebAssembly for the browser. The client plays the season locally at full speed with no network
round-trips; the server re-runs the submitted decisions natively and computes the authoritative score.
There is no second implementation, so there is no drift between them — and a parity test asserts that
both targets produce identical output across thousands of seeds.

That is what makes the leaderboard trustworthy, and it is the thing to talk about in an interview.
