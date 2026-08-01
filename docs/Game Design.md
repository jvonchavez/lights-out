# Lights Out — Game Design

The engineering is only worth doing if the game is worth playing. This document fixes the rules and
the numbers before any code exists, because the simulation is the API that everything else is built
against.

---

## The central decision

> **Revised after M4.** The three-slider budget split is preserved at the end of this section,
> along with why it was replaced.

Development happens in **five windows**, before rounds 1, 3, 5, 7 and 9. At each one you are dealt
**three parts** from a pool of twenty-four and you take one. It locks, and it fits for the rest of
the season. The intervening races run on whatever the car already is.

| | |
|---|---|
| **Chassis** | Cornering — dominant at technical circuits |
| **Engine** | Top speed — dominant at power circuits |
| **Aero** | Multiplies the whole car, capped at +31%, and is the only area that *improves* efficiency |

A part's cost is the development it buys, between 140 and 260 units. Cost drives risk, which is
quadratic in what you have banked across the season, so the fifth expensive part costs far more risk
than the first. Aero is credited back 30% of the pressure its own share caused, which makes an
aero-heavy card genuinely safer for its cost. **The risk pips on a card face are computed from its
effect, never authored** — what you are shown is what the simulation applies.

**What makes it a game rather than a spreadsheet:** the full calendar is visible from window one, so
you can see that rounds 4 through 7 are power circuits and take the gearbox two windows early. And
because the deal is random, every window is a hand you have to play rather than a form you fill in.
The tension is the one 82-0 runs on: take the enormous aero part now, or hope a better one comes
round while the cheap safe option sits in front of you.

Development compounds — a part fitted in window 1 pays off across nine remaining races, one fitted
in window 5 across one. Early aggression is mathematically correct and maximally risky.

### Why the sliders went

The sliders worked and were not fun. Measured against 82-0: randomness landed *after* the decision
rather than before, so nothing was ever dealt to you and every round presented an identical
surface; a season took roughly forty deliberate interactions against 82-0's five clicks; nothing had
a name you could hold an opinion about; and the trade-off could not be stated in a sentence, so the
UI had to explain squared risk in a paragraph above the confirm button.

The simulation was never the problem — M1 had already proved its decisions mattered. The input
surface was. So `internal/sim` kept its physics, its RNG, its parity guarantee and its verification
model, and only the meaning of a "decision" changed: from an arbitrary allocation to an index into a
dealt hand.

<details>
<summary>The three-slider design, superseded after M4</summary>

Before each race you receive a fixed development budget and split it across three areas:

| Area | Helps | Costs |
|---|---|---|
| **Chassis** | Cornering — dominant at technical circuits | Reliability |
| **Engine** | Top speed — dominant at power circuits | Reliability, more steeply |
| **Aero** | Multiplies the whole car, capped at +15%, and is the only area that *improves* efficiency | Reliability, mildly |

Every point spent raises performance and lowers reliability. Reliability failures are catastrophic —
a DNF scores zero, and there are only 10 races. So the game is a repeated bet: how much performance
will you buy with how much risk, given how many races are left to recover in?

**What makes it a game rather than a spreadsheet:** the full calendar is visible from race one. Each
circuit has a published profile, so you can see that rounds 4 through 7 are power circuits and invest
in engine two races early. Development compounds — a point spent in round 1 pays off across nine
remaining races, while a point in round 9 pays off across one. Early aggression is mathematically
correct and maximally risky, and that tension is the whole game.

---


</details>

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

**Settled at M1:** qualifying σ = 2.5 rating points, race σ = 1.5. Both were left at their initial
values — the sweeps showed the balance problems lay in the reliability and aero models rather than in
the noise, and changing σ would have masked them rather than fixed them.

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

> **Revised at M1.** The original model and the fourth slider are recorded at the end of this
> section, along with the measurements that replaced them.

```
failure_probability = base + (cumulative_development_spend / full_season_spend)² × pressure_quad
                      − aero_efficiency_credit
```

Development spend raises risk permanently, not just for the next race — you are running a highly
strung car for the rest of the season. This is what stops "spend everything in round 1" from being a
free win.

Risk is **convex** in cumulative spend: the first tenth of a season's development is nearly free, the
last tenth is expensive. Aero is credited back 30% of the pressure its own share of the spending
caused, which is the sense in which it "improves efficiency".

### What M1 measured, and what changed

**The linear model was degenerate.** Under `base + spend × coefficient`, risk and performance are both
linear in spend, so one side always dominates globally and there is no interior optimum. Sweeping the
coefficient from 250 to 850 per 1000 units, spending 100% of the budget was optimal at every value —
even at 4.19 DNFs per 10-race season. The trade-off the game is built around could not exist in that
form. Squaring the spend term fixes it.

**The reliability slider was a trap and is cut.** Built as specified, then measured: spending zero on
it was optimal at every pressure coefficient under the linear model *and* under the convex one, with
title chance falling monotonically as investment rose (27.0% → 23.9% → 15.4% → 7.5% → 2.5% at 0/10/20/
30/40% of budget). Reliability costs performance one-for-one while performance compounds across every
remaining race. The document's own instruction was to cut it rather than keep it out of sentiment, so
it is cut, and the game is the three sliders `_README.md` always described — with reliability as the
*cost* of development rather than a purchase.

**The player was drawn at the mean and had no ceiling.** Rivals drew starting ratings from a spread
while the player was pinned to exactly 50.0; the player scored 100.5 average points against a rival on
the same strategy scoring 102.9, yet won 1.5% of titles against its 11.2%. A car with no variance
never reaches the top of the field. The player now draws from the same distribution as the rivals —
seeded, so every player in the world still gets the same car on the same day.

**Aero now multiplies rather than only adds.** "Scales with both" was implemented as a plain third
weight, which left performance linear in the ratings and made concentration strictly better than
spreading — the optimal play collapsed to "everything into whichever area the calendar weights
highest", and the three sliders stopped being a decision. Aero rating above the baseline now
multiplies the whole weighted sum, capped at +15%. The cap is what keeps aero from simply replacing
engine as the single dominant answer, and makes "buy aero to the cap, then read the calendar" the
strongest line rather than an exploit.

### Final balance, 100,000 seasons × 5 strategies

| Strategy | Titles | Avg pts | Avg finish | Avg DNFs | Avg spend |
|---|---|---|---|---|---|
| aerofirst — buy aero to the cap, then read the calendar | 20.8% | 124.7 | 3.80 | 1.51 | 1077 |
| greedy — always take the costliest part | 14.3% | 108.4 | 4.80 | 1.74 | 1146 |
| adaptive — take what suits the next two races | 10.4% | 96.6 | 5.54 | 1.69 | 1109 |
| first — always take the leftmost card | 8.1% | 90.4 | 5.91 | 1.42 | 1006 |
| cautious — always take the cheapest part | 4.2% | 72.0 | 7.14 | 1.13 | 865 |

No strategy exceeds the ~35% ceiling, and the gradient from 4.2% to 20.8% is wide enough that
decisions plainly matter. The 79% of seasons the best line still loses is what keeps a daily seed
worth replaying.

**The card draft needed the aero cap raised**, from 150 to 310. With sliders a player could pour
unlimited budget into aero, so the cap had to be tight. A deal of three cards constrains that on its
own, and at 150 the deepest strategy barely separated from the crudest: aerofirst took 14.4% of
titles against greedy's 13.9%, meaning "buy the biggest thing" was as good as understanding that
aero compounds. Raising `PressureQuad` instead would have separated them only by pushing DNFs past
three per season. The aero cap moves performance rather than risk, so DNF rates are untouched.

Not raised further: at 400 aerofirst takes 25.1% and starts becoming the single right answer, which
is the failure M1 originally fixed.

<details>
<summary>Original specification, superseded at M1</summary>

```
failure_probability = base + (cumulative_development_spend × pressure_coefficient) − reliability_investment
```

A fourth budget option, **reliability investment**, buys the risk back down without adding
performance. It exists so the player has a way to say "I have a fast car, now let me finish races."
Whether it earns its place is an M1 balance question — if simulation shows nobody ever picks it, it
gets cut rather than kept out of sentiment.

</details>

---

## Rival AI

Ten rival teams, each with a fixed strategy archetype and a starting rating spread:

- **Aggressive** — spends heavily and early, high DNF rate, high ceiling
- **Conservative** — spreads spending evenly, rarely fails, rarely wins
- **Specialist** — pours everything into one area, dominant at circuits that suit it
- **Reactive** — invests toward whichever archetype dominates the next two circuits

Rivals are dealt **the same three cards as the player** and choose by archetype: aggressive takes
the costliest, conservative the cheapest, specialist whatever favours its area, reactive whatever
suits the races just ahead. Ties break on the lowest index, so every choice is a total order.

That makes rival behaviour legible — you can see which team took the part you passed on — and it
keeps them fully deterministic given the seed. They are not adaptive to the player, which is a
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
P2 of 11 · 287 pts · 1 DNF
🏁🥈🥇🥉🏁🥇✖️🥈🥇🥈
```

Three lines, no spoilers about the correct strategy, and the emoji row conveys the shape of a season
at a glance — including the one race that went wrong.

The position carries its own denominator. "P2" means nothing to someone who has never played;
"P2 of 11" is legible cold, which is what makes 82-0's "71-11" work as a boast.

**Naming your parts is opt-in.** On a shared daily seed your build *is* the strategy, so the default
copy stays spoiler-free and a second button appends the five part names. That is the thing worth
arguing about, and it should be a deliberate act rather than the default.

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
