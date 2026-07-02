# Lights Out — Game Design

The engineering is only worth doing if the game is worth playing. This document fixes the rules and
the numbers before any code exists, because the simulation is the API that everything else is built
against.

---

## The central decision

Before each race you receive a fixed development budget and split it across three areas:

| Area | Helps | Costs |
|---|---|---|
| **Chassis** | Cornering — dominant at technical circuits | Reliability |
| **Engine** | Top speed — dominant at power circuits | Reliability, more steeply |
| **Aero** | Scales with both, and is the only area that *improves* efficiency | Reliability, mildly |

Every point spent raises performance and lowers reliability. Reliability failures are catastrophic —
a DNF scores zero, and there are only 10 races. So the game is a repeated bet: how much performance
will you buy with how much risk, given how many races are left to recover in?

**What makes it a game rather than a spreadsheet:** the full calendar is visible from race one. Each
circuit has a published profile, so you can see that rounds 4 through 7 are power circuits and invest
in engine two races early. Development compounds — a point spent in round 1 pays off across nine
remaining races, while a point in round 9 pays off across one. Early aggression is mathematically
correct and maximally risky, and that tension is the whole game.

---

## Circuit profiles

Each circuit weights the three ratings. Weights sum to 1.0.

| Archetype | Chassis | Engine | Aero | Example |
|---|---|---|---|---|
| Power | 0.20 | 0.55 | 0.25 | Long straights, low downforce |
| Technical | 0.55 | 0.20 | 0.25 | Slow corners, high downforce |
| Balanced | 0.35 | 0.35 | 0.30 | Mixed |
| High-speed | 0.30 | 0.35 | 0.35 | Fast sweeping corners |

The MVP calendar is 10 fictional circuits — 3 power, 3 technical, 2 balanced, 2 high-speed — in a
fixed order that the daily seed shuffles. Fictional names avoid trademark questions entirely and cost
nothing in fun.

---

## Resolution

**Car performance** at a circuit is the weighted sum of ratings by circuit profile, plus driver skill,
plus noise:

```
perf = Σ(rating[area] × weight[area]) + driver_skill + N(0, σ)
```

σ is the single most important tuning constant in the game. Too low and the best car always wins and
your decisions stop mattering; too high and decisions drown in noise. This gets tuned empirically in
milestone M1, not guessed.

**Qualifying** ranks all 11 teams by one `perf` roll. Grid position follows.

**Race** resolves in three phases:

1. **Reliability check** — each car rolls against its own failure probability. A failure is a DNF and
   the car scores nothing. This happens first so the rest of the race only simulates survivors.
2. **Race pace** — a second `perf` roll with a *lower* σ than qualifying, because a race averages 50+
   laps and regresses toward true pace. Combined with grid position via a track-specific overtaking
   difficulty factor: at a power circuit, a fast car recovers from a bad grid slot; at a technical
   circuit, it stays stuck.
3. **Events** — a safety car fires with a fixed per-race probability and compresses the field,
   partially randomising finishing order among cars within a threshold of each other. It is a
   deliberate equaliser that keeps a dominant season from becoming boring.

**Points** use the standard 25-18-15-12-10-8-6-4-2-1 for the top ten.

---

## Reliability

```
failure_probability = base + (cumulative_development_spend × pressure_coefficient) − reliability_investment
```

Development spend raises risk permanently, not just for the next race — you are running a highly
strung car for the rest of the season. This is what stops "spend everything in round 1" from being a
free win, and it is the number most likely to need retuning after M1.

A fourth budget option, **reliability investment**, buys the risk back down without adding
performance. It exists so the player has a way to say "I have a fast car, now let me finish races."
Whether it earns its place is an M1 balance question — if simulation shows nobody ever picks it, it
gets cut rather than kept out of sentiment.

---

## Rival AI

Ten rival teams, each with a fixed strategy archetype and a starting rating spread:

- **Aggressive** — spends heavily and early, high DNF rate, high ceiling
- **Conservative** — spreads spending evenly, rarely fails, rarely wins
- **Specialist** — pours everything into one area, dominant at circuits that suit it
- **Reactive** — invests toward whichever archetype dominates the next two circuits

Rivals are fully deterministic given the seed. They are not adaptive to the player, which is a
deliberate simplification: an AI that reacts to you makes the daily seed unfair, because two players
making different early choices would face different fields.

---

## Scoring and sharing

Season score is championship points. The leaderboard ranks by points, tie-broken by wins, then by
podiums.

The share string is the point of the whole exercise and should be designed with the same care as the
sim:

```
Lights Out · Season 142
P2 · 287 pts · 1 DNF
🏁🥈🥇🥉🏁🥇✖️🥈🥇🥈
```

One line, no spoilers about the correct strategy, and the emoji row conveys the shape of a season at a
glance — including the one race that went wrong.

---

## Determinism contract

This is the constraint every rule above must respect, because the leaderboard depends on it.

- A season is a pure function: `(seed, []Decision) → SeasonResult`. No wall-clock time, no map
  iteration order, no unseeded randomness, no floating-point that varies by platform.
- The RNG is explicitly seeded and advances in a fixed order. Adding a new random draw anywhere
  changes every downstream result, so **any change to the draw order is a breaking change** and must
  bump the sim version.
- Seasons record the sim version they were generated under. Old leaderboards are never re-verified
  against new sim code; they are frozen.

Floating-point determinism deserves particular care, since the same Go source compiles to both native
and WASM. The parity test in M4 exists specifically to catch this, and if a discrepancy appears the
fix is fixed-point or integer arithmetic in the affected calculation rather than hoping it does not
matter.
