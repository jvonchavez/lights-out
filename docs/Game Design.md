# Lights Out — Game Design

The engineering is only worth doing if the game is worth playing. This document fixes the rules and
the numbers before any code exists, because the simulation is the API that everything else is built
against.

---

## The central decision

> **Rebuilt after M4.** The five-window card draft is preserved at the end of this section,
> along with why it was replaced.

You build a team **before the season starts**, and then you watch it race.

The draft is **five rolls**. Each one lands on a real team-season out of Formula 1 history —
1988 McLaren, 2004 Ferrari, 1967 Lotus — and from it you take exactly **one** thing:

| | |
|---|---|
| **The car** | Power, cornering, aero and reliability |
| **A driver** | Either of their two. You need two, so this slot is filled twice |
| **The race engineer** | Setup in qualifying, strategy in the race, and a cleaner operation |
| **The team principal** | Develops the car all year, lifts both drivers, holds their nerve at the end |

Five rolls, five slot-places, **no passing**. Every roll must fill something.

**The scarcity needs no budget system.** A great team-season has a great car *and* a great
driver *and* a great principal, and you get one of them. Roll 1988 McLaren and take the
MP4/4, and Senna goes to somebody else. Nothing has a price, and yet everything costs.

**What you are racing** is the real 2026 grid: eleven teams, their actual cars, drivers, race
engineers and principals. That field is identical on every seed, so no day is easier than
another and the leaderboard measures your draft and nothing else. It also makes the goal
legible without a paragraph of explanation — beat Mercedes and Ferrari over ten rounds with a
team you assembled out of history.

**Reliability is what separates the eras.** Performance ratings are relative to a car's own
field, so the Ferrari 500 that won every race it entered in 1952 reads like it. Reliability is
on one absolute scale across the whole roster, and it is brutal about the past: a 2023 Red
Bull retires about twice a season across two cars, a 1967 Lotus 49 about six times. The Lotus
is one of the fastest cars in the game. Taking it is a bet, not a purchase.

**Development is the only thing that moves after lights out in round one.** The principal adds
rating to the car every round, so a mid car under Colin Chapman out-develops a great car under
a weak principal by the closing rounds. It is the most Formula 1 thing in the design and the
reason the principal slot is not an afterthought.

### Why the card draft went

The card draft was a real improvement on the sliders and it was still the wrong game. Parts
called "Titanium Gearbox" with a +25 Engine label are abstractions nobody can hold an opinion
about, and a season was one seed a day with 243 possible lines. Measured against what makes a
daily worth replaying, it had legibility but nothing to argue about.

The simulation was never the problem. `internal/sim` kept its fixed-point arithmetic, its
vendored PCG64, its purity test, its parity guarantee and its verification model. What changed
is what a decision is made of: from an index into a hand of anonymous parts to an index into a
team you recognise.

**The honest cost.** Five rolls of five items is 240 legal drafts, down from 243 — and a
smaller space than that sounds, because the last roll is usually forced. `docs/_README.md`
states plainly that the answer space is enumerable; this makes it no better and it is written
down rather than quietly dropped. What carries replayability instead is the roster: 52
team-eras means the five teams you are offered are different nearly every run.

<details>
<summary>The five-window card draft, superseded after M4</summary>

Development happened in **five windows**, before rounds 1, 3, 5, 7 and 9. At each one you were
dealt **three parts** from a pool of twenty-four and took one. It locked, and it fitted for the
rest of the season. The intervening races ran on whatever the car already was.

| | |
|---|---|
| **Chassis** | Cornering — dominant at technical circuits |
| **Engine** | Top speed — dominant at power circuits |
| **Aero** | Multiplies the whole car, capped at +31%, and is the only area that *improves* efficiency |

A part's cost was the development it bought, between 140 and 260 units. Cost drove risk, which
was quadratic in what you had banked across the season. Aero was credited back 30% of the
pressure its own share caused.

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
</details>

## Circuit profiles

Each circuit weights the three ratings. Weights sum to 1.0.

| Archetype | Chassis | Engine | Aero | Example |
|---|---|---|---|---|
| Power | 0.20 | 0.55 | 0.25 | Long straights, low downforce |
| Technical | 0.55 | 0.20 | 0.25 | Slow corners, high downforce |
| Balanced | 0.35 | 0.35 | 0.30 | Mixed |
| High-speed | 0.30 | 0.35 | 0.35 | Fast sweeping corners |

The calendar is 10 fictional circuits — 3 power, 3 technical, 2 balanced, 2 high-speed — in a
fixed order that the daily seed shuffles. Fictional names avoid trademark questions entirely and cost
nothing in fun.

---

## Resolution

Every rating is a 0–99 integer and maps straight into the simulation's fixed-point space, so a
95-rated car corner-weights at 95.0. **Overall is computed from the sub-ratings and never
authored** — what a card face shows is derived from what the simulation applies, which is the
rule the card draft's risk pips already established.

**Car performance** at a circuit is the weighted sum of its three performance ratings, plus
whatever the principal has developed onto it:

```
car(round) = power×w.engine + cornering×w.chassis + aero×w.aero
           + round × (development − 70) × DevRate
```

Aero is a plain third weight. Under the card draft it *multiplied* the whole sum, because
performance was linear in player-allocated ratings and concentration therefore always beat
spreading. Nothing is allocated now — a car is taken whole — so the degenerate case cannot
arise and the multiplier is gone with it.

**A car's performance** adds its driver and staff on top, and only then draws noise:

```
perf = car(round) + pace + leadership
     + setup      (qualifying only)
     + strategy   (race only)
     + nerve      (final three rounds only)
     + N(0, σ × consistencyFactor(driver))
```

The spreads are deliberate and unequal: about 36 rating points separate the best car from the
worst, about 8 per driver, and 3–4 for an engineer. In Formula 1 the car dominates, and a
draft that pretended otherwise would be a lie. Two drivers is 16 points, though, which is why
the second seat is a decision rather than a passenger.

**Consistency modulates the driver's own σ.** This is what stops Overall from being the whole
story: Gilles Villeneuve and Alain Prost are not the same bet, and a fast erratic driver and a
slower metronome can carry the same rating and play completely differently over ten races.
`RNG.Normal` consumes a fixed twelve uniforms whatever σ it is given, so a per-driver σ changes
no draw counts and the determinism contract is untouched.

**Racecraft buys back a share of the grid penalty**, capped at half, so a driver who can
overtake is worth most exactly where overtaking is hardest.

**Qualifying** ranks all 24 cars by one `perf` roll. Grid position follows.

**Race** resolves in three phases:

1. **Reliability** — one draw per car (see below). A failure is a DNF and the car scores
   nothing. This happens first so the rest of the race only simulates survivors.
2. **Race pace** — a second `perf` roll at a lower σ than qualifying, because a race averages
   50+ laps and regresses toward true pace, combined with grid position via the circuit's
   overtaking difficulty and the driver's racecraft.
3. **Events** — a safety car fires at a fixed per-race probability and compresses the field.

**Points** use the standard 25-18-15-12-10-8-6-4-2-1, now for the top ten of **24 cars**. Both
of a team's cars feed one constructors' total.

---

## Reliability

> **Rebuilt after M4.** The development-spend model is recorded at the end of this section.

Two independent causes share **one RNG draw** per car per race: the car retires if the uniform
falls below their sum, and the cause is whichever band it landed in. One draw keeps the
determinism contract simple and still gives the race reel a reason to name, which is the
difference between "DNF" and "engine".

```
mechanical = floor + (100 − reliability)² × quad ÷ 100,  less a share for engineer ops
driver     = base − (composure − 60) × rate
```

**Mechanical risk is quadratic in the car's shortfall from perfect reliability, not linear.**
That is an M3 measurement, not a preference. A linear slope cannot do both jobs at once:
fitted to make a 2026 car reliable it makes a 1967 car reliable too, and fitted to make a 1967
car fragile it retires the modern grid four times a season. Measured at the linear settings,
Cadillac averaged **3.99 retirements per season — 20% of its starts**, which is not a Formula 1
team, it is attrition. The curve gives roughly 2% a start at reliability 95, 6% at 75, 16% at
55 and 21% at 48, so the modern grid finishes and the era trade-off the roster is built on
stays real.

**The engineer's operation removes a fraction of the risk rather than a fixed amount**, so a
great race engineer is worth much more on a fragile car than on one that was never going to
break.

### What M3 measured, and what changed

**The 2026 grid was rated too fragile.** Its reliability had been authored on the same scale as
the historical roster, which gave 1970s retirement rates to 2026 cars — Cadillac 3.99 a season,
Aston Martin 2.90. A struggling modern team still reaches the flag; it just reaches it
fourteenth. Raised, and DNFs across the whole grid now sit between 0.78 and 1.99 per team-season
across two cars.

**Sigma was the wrong lever and the metric was biased.** Sweeping race σ from 1.5 to 7.5
changed the number of teams that score from 5 to 7 and did nothing for the shape of the
championship. Worse, it *inflated* the measured ceiling — perfect play is the best of 240
enumerated lines, so more noise mechanically raises the best-of, and "perfect-play title%" is
not a σ-independent metric. Left at 2.5 qualifying / 1.5 race.

**The midfield lockout is structural and was left alone.** Seven of eleven rivals average under
six points a season. Two attempts to fix it by rating failed: compressing the whole grid toward
the leader by 35% and 55% moved Racing Bulls from 7.8 points to 7.8 and 7.4, and raising the
midfield's own ratings by 6–10 points moved it from 5.4 to 5.0. The cause is not the ratings.
The top four teams field **eight cars and only ten positions score**, so P1–P8 are spoken for
and everyone else contests two places. The speculative midfield raise was reverted rather than
kept, because it bought nothing and made the grid less faithful to the real standings — where
2026 is, in fact, a runaway.

### Final balance, 100,000 seasons × 5 strategies

| Strategy | Titles | Avg pts | Avg finish | Avg DNFs | Avg car |
|---|---|---|---|---|---|
| bestavailable — take the highest-rated thing on offer | 17.7% | 228.0 | 2.64 | 1.18 | 91 |
| cautious — value finishing over pace | 10.4% | 167.1 | 3.56 | 0.97 | 87 |
| starpower — fill both driver seats first | 10.1% | 180.9 | 3.24 | 1.42 | 89 |
| carfirst — always take the car | 7.7% | 138.5 | 3.94 | 1.53 | 85 |
| first — take the leftmost legal item | 5.3% | 127.1 | 4.10 | 1.58 | 85 |

Enumerating all 240 legal drafts wins the constructors' championship **40% of the time**, so
the game is winnable with real thought and not with none: the gradient from 5.3% to 17.7% to
40% is what the draft is measuring.

**Two findings worth keeping.** *Greedy early is a trap.* `starpower` ends up with worse
drivers than `carfirst` — 35288 rating against 35895 — because filling both driver seats from
the first two rolls forfeits the choice later rolls would have offered. `carfirst`, which
always takes the car on roll one, is the second-worst line in the game. Grabbing the thing you
want first is not the same as ending up with the best of it, and that is the most interesting
property the roll mechanic produces.

*A greedy picker must compare like with like.* Car reliability averages in the seventies across
the roster while principal development averages in the mid-eighties, so a naive "take the most
reliable thing" never took a car at all and ended up with whichever chassis the last roll
offered. Scoring an item by its distance above the roster mean **for its own slot** fixes it —
grouped by slot and not by item, because lead drivers rate higher on average and a per-item
baseline quietly biases every line toward the number two seat.

<details>
<summary>The development-spend risk model, superseded after M4</summary>

```
failure_probability = base + (cumulative_development_spend / full_season_spend)² × pressure_quad
                      − aero_efficiency_credit
```

Development spend raised risk permanently. Risk was **convex** in cumulative spend: the first
tenth of a season's development was nearly free, the last tenth expensive. Aero was credited
back 30% of the pressure its own share of the spending caused.

M1 found the originally specified *linear* model degenerate — spending 100% of the budget was
optimal at every pressure coefficient from 250 to 850, even at 4.19 DNFs per season, because
risk and performance were both linear in spend and there was no interior optimum. A fourth
"reliability investment" slider was built, measured, and cut: spending zero on it was optimal
everywhere, with title chance falling monotonically as investment rose.

None of it survives the rebuild, because nothing is bought during a season any more.

</details>

---

## The field

There is no rival AI. The eleven teams you race are the real 2026 grid — McLaren, Ferrari, Red
Bull, Mercedes, Aston Martin, Alpine, Haas, Racing Bulls, Williams, Audi and Cadillac — each
with its actual car, driver pairing, race engineer and team principal. Their ratings are
anchored to the real constructors' table after twelve rounds: Mercedes 425 down to Cadillac 0.

They are identical on every seed. Three things follow, and all three are improvements on the
procedural rivals they replace:

- **The benchmark is constant.** No seed is easier than another, so the leaderboard measures
  your draft and nothing else. For a shared daily seed that is strictly better than a random
  field.
- **The goal is legible cold.** "Beat Red Bull and McLaren with a team you assembled out of
  history" needs no explanation. "Finish above ten procedurally generated archetypes" did.
- **There is no adaptive-AI trap.** An AI that reacts to the player makes a shared seed unfair,
  because two players making different early choices would face different fields. A fixed grid
  cannot.

The 2026 entries are **rollable like any other team-era**. A roll can land on 2026 McLaren, and
taking their car or their driver means racing the real McLaren with a hole in it. It costs
nothing to allow and it is the best moment the mechanic produces.

---

## Scoring and sharing

Season score is championship points. The leaderboard ranks by points, tie-broken by wins, then by
podiums.

The share string is the point of the whole exercise and should be designed with the same care as the
sim:

```
Lights Out · Season 142
P2 of 12 · 287 pts · 1 DNF
🏁🥈🥇🥉🏁🥇✖️🥈🥇🥈
```

Three lines, no spoilers about the correct strategy, and the emoji row conveys the shape of a season
at a glance — including the one race that went wrong.

The position carries its own denominator. "P2" means nothing to someone who has never played;
"P2 of 12" is legible cold, which is what makes 82-0's "71-11" work as a boast.

**Naming your parts is opt-in.** On a shared daily seed your build *is* the strategy, so the default
copy stays spoiler-free and a second button appends the lineup. That is the thing worth
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
