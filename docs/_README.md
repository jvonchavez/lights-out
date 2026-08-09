# Lights Out — F1 Season Manager

**Status:** planning · **Started:** 2026-07-28 · **Language:** Go + TypeScript

A browser game where you build an F1 team out of history and then watch it race. The racing itself
is simulated — you never drive. Every decision happens *before* the season: five rolls, each landing
on a real team-season, and from each you take exactly one thing — their car, one of their drivers,
their race engineer, or their team principal. Then your team races the real 2026 grid over ten
rounds, and you find out whether it wins the Constructors' Championship.

Everyone in the world gets the same season seed each day. Same calendar, same five rolls, same
random stream — and the same opposition, because the field is the actual 2026 grid rather than
anything generated. The luck is identical for everyone, so the leaderboard measures decisions and
nothing else.

---

## Documents

| File | Contents |
|---|---|
| `Game Design.md` | The rules, the numbers, the balance findings, and every design the measurements replaced |
| `Architecture.md` | Services, data flow, data model, deployment topology |
| `Tech Stack.md` | Every technology choice, its justification, and the alternative rejected |
| `Build Plan.md` | Milestones M0–M6 with a definition of done for each |

---

## MVP definition

A player loads the site, sees today's season, drafts a team across five rolls, watches ten races
resolve in a reel, and lands on the leaderboard against everyone else who played the same seed.
Their score is verified server-side before it counts. Free play gives unlimited unscored runs on
random seeds, because one season a day is not replayability.

**In scope:** 10-race calendar, 5 draft rolls over a roster of 52 real team-eras, the 2026 grid as
a fixed 11-team field, 24 cars with both of a team's drivers scoring for the constructor,
qualifying and race simulation, mechanical and driver-error retirements, in-season development,
championship standings, daily seed, verified leaderboard, shareable result string, free play.

**Out of scope:** mid-season transfers, multi-season careers, pit strategy, tyre compounds, weather,
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

They can, however, be *searched for*. Five rolls of five items is **240 legal drafts**, and the WASM
module runs locally, so anyone can enumerate every line in milliseconds and submit the best. This is
no better than the card draft's 243 and the number is not going to be talked up: an earlier draft of
this document claimed a cheater "would have to find decisions that genuinely produce a high score,
which is just playing well", which was true of the continuous sliders and has not been true since.

Measured, because the number is worth knowing: enumerating all 240 wins the championship **40%** of
the time, against 17.7% for the best scripted strategy and 5.3% for the worst. A solver is
meaningfully better than a person, and the honest ceiling is not "you cannot cheat" but "the
simulation decides, not you".

This is accepted. Wordle's answer space is roughly 2,300 words and solvers crack it instantly; it
remains a phenomenon because the fun is doing it yourself. What carries replayability instead is the
roster: 52 team-eras means the five teams you are offered are different nearly every run, and free
play makes trying another five free. Nothing resists brute force when the
simulation runs client-side, and the alternative — hiding the rules from the browser — would cost
the dual-target design that is the actual point of this project.
