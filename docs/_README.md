# Lights Out — F1 Season Manager

**Status:** planning · **Started:** 2026-07-28 · **Language:** Go + TypeScript

A browser game where you run an F1 team through a 10-race season. The racing itself is simulated —
you never drive. Your only decisions happen *between* races: across five development windows you are
dealt three parts and take one, knowing that every part you bolt on costs reliability, and a
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

A player loads the site, sees today's season, drafts a car across five windows, watches ten races
resolve, and lands on the leaderboard against everyone else who played the same seed. Their score is
verified server-side before it counts.

**In scope:** 10-race calendar, 5 development windows dealing 3 cards each, 10 rival teams with their
own AI development strategies, qualifying and race simulation, reliability failures, championship
standings, daily seed, verified leaderboard, shareable result string.

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

## What the verification does and does not claim

The server is the sole authority on what a set of picks is worth. It re-derives the deal from the
season's seed, replays the picks, and computes the score itself; the client sends card indices and
nothing else, and `submitRequest` has no score field at all, so a forged number has nowhere to be
decoded to. **Scores cannot be fabricated.**

They can, however, be *searched for*. Five windows of three cards is 243 possible seasons, and the
WASM module runs locally, so anyone can enumerate every line in milliseconds and submit the best.
An earlier draft of this document claimed a cheater "would have to find decisions that genuinely
produce a high score, which is just playing well" — that was true of the continuous sliders and is
not true of a card draft. It is stated here rather than quietly dropped.

This is accepted. Wordle's answer space is roughly 2,300 words and solvers crack it instantly; it
remains a phenomenon because the fun is doing it yourself. Nothing resists brute force when the
simulation runs client-side, and the alternative — hiding the rules from the browser — would cost
the dual-target design that is the actual point of this project.
