# Lights Out — Game Design

The rules and the numbers. The simulation is the API everything else is built against, so this
document is the source of truth for what a season *is*.

---

## The draft

You build a team before the season starts, then watch it race.

The draft is **five rolls**. Each lands on a real team-season out of F1 history — 1988 McLaren,
2004 Ferrari, 1967 Lotus — and from it you take exactly **one** thing:

| Item | What it does |
|---|---|
| **The car** | Power, cornering, aero, reliability |
| **A driver** | Either of the team's two. You need two, so this slot fills twice |
| **The race engineer** | Setup in qualifying, strategy in the race, cleaner operation |
| **The team principal** | Develops the car all year, lifts both drivers, holds nerve at the end |

Five rolls, five slot-places, no passing. Every roll must fill something.

**Scarcity without a budget.** A great team-season has a great car *and* a great driver *and* a
great principal, and you get one of them. Take the MP4/4 from 1988 McLaren and Senna goes
elsewhere.

**You race the real 2026 grid** — eleven teams with their actual cars, drivers, engineers and
principals. The field is identical on every seed, so no season is easier than another and the
result measures the draft alone.

**Play is unlimited.** A season is issued on request: fresh seed, fresh calendar, five fresh rolls.
The leaderboard shows how many seasons a player submitted beside their best, because a best-of-N is
partly a measure of N.

**Reliability separates the eras.** Performance ratings are relative to a car's own field, so the
1952 Ferrari 500 reads like the car that won everything. Reliability is on one absolute scale across
the roster: a 2023 Red Bull retires about twice a season across two cars, a 1967 Lotus 49 about six
times. The Lotus is one of the fastest cars in the game. Taking it is a bet.

**Development is the only thing that moves mid-season.** The principal adds rating to the car each
round. Best principal (Chapman, 95) out-develops the worst (Ugolini, 74) by 1.70 rating points on
season average and 3.40 by the final round — 2.3σ of race noise. Enough to overturn a small car
deficit, not enough to beat a great car with a mid one (car Overall spans 61–97).

---

## Circuits

Real circuits: the 2026 calendar plus tracks the championship has raced before. Each is classified
by the character that decides lap time, not by lap length — judgment calls on downforce level and
overtaking difficulty. Zandvoort is technical despite its banking; Spa is power despite its corners.

| Archetype | Pool | Chassis | Engine | Aero | Examples |
|---|---|---|---|---|---|
| Power | 9 | 0.20 | 0.55 | 0.25 | Monza, Spa, Baku, Las Vegas |
| Technical | 8 | 0.55 | 0.20 | 0.25 | Monaco, Hungaroring, Zandvoort, Imola |
| Balanced | 10 | 0.35 | 0.35 | 0.30 | Barcelona, Bahrain, Interlagos, COTA |
| High-speed | 7 | 0.30 | 0.35 | 0.35 | Silverstone, Suzuka, Albert Park |

**Every calendar is drawn to a fixed quota** — three power, three technical, two balanced, two
high-speed — then shuffled. Membership varies, shape never does. Ten power circuits would make one
rating the whole game; a fixed calendar made every season identical. The 100k-season sweep came back
identical after fictional circuits were replaced with real ones, which is what the quota is for.

---

## Resolution

Every rating is a 0–99 integer mapping straight into fixed-point space. **Overall is computed from
the sub-ratings, never authored.**

```
car(round) = power×w.engine + cornering×w.chassis + aero×w.aero
           + round × (development − 70) × DevRate

perf = car(round) + pace + leadership
     + setup      (qualifying only)
     + strategy   (race only)
     + nerve      (final three rounds only)
     + N(0, σ × consistencyFactor(driver))
```

Aero is a plain third weight, not a multiplier — nothing is allocated by the player any more, so
the degenerate concentration case cannot arise.

The spreads are deliberately unequal: ~36 rating points separate the best car from the worst, ~8 per
driver, 3–4 for an engineer. In F1 the car dominates. Two drivers is 16 points, which is why the
second seat is a decision rather than a passenger.

**Consistency modulates the driver's own σ**, so a fast erratic driver and a slower metronome can
carry the same rating and play differently. `RNG.Normal` consumes a fixed twelve uniforms whatever σ
it is given, so this changes no draw counts.

**Racecraft buys back a share of the grid penalty**, capped at half — worth most where overtaking is
hardest.

**Qualifying** ranks all 24 cars by one `perf` roll; grid position follows.

**Race** resolves in three phases:

1. **Reliability** — one draw per car. A failure is a DNF scoring nothing. First, so the rest of the
   race only simulates survivors.
2. **Race pace** — a second `perf` roll at lower σ than qualifying (a race regresses toward true
   pace), combined with grid position via the circuit's overtaking difficulty and driver racecraft.
3. **Events** — a safety car fires at a fixed per-race probability and compresses the field.

**Points**: 25-18-15-12-10-8-6-4-2-1 for the top ten of 24 cars. Both of a team's cars feed one
constructors' total.

---

## Reliability

Two independent causes share **one RNG draw** per car per race: the car retires if the uniform falls
below their sum, and the cause is whichever band it landed in. One draw keeps determinism simple and
still gives the race reel a cause to name.

```
mechanical = floor + (100 − reliability)² × quad ÷ 100,  less a share for engineer ops
driver     = base − (composure − 60) × rate
```

**Mechanical risk is quadratic in the shortfall from perfect reliability, not linear.** A linear
slope cannot do both jobs: fitted to make a 2026 car reliable it makes a 1967 car reliable too;
fitted to make a 1967 car fragile it retires the modern grid four times a season. At the linear
settings Cadillac averaged 3.99 retirements per season — 20% of its starts. The curve gives roughly
2% a start at reliability 95, 6% at 75, 16% at 55, 21% at 48.

**The engineer's operation removes a fraction of the risk, not a fixed amount**, so a great engineer
is worth much more on a fragile car.

DNFs across the current grid sit between 0.78 and 1.99 per team-season across two cars.

---

## Balance, 100,000 seasons × 5 strategies

| Strategy | Titles | Avg pts | Avg finish | Avg DNFs | Avg car |
|---|---|---|---|---|---|
| `bestavailable` — take the highest-rated thing on offer | 17.7% | 228.0 | 2.64 | 1.18 | 91 |
| `cautious` — value finishing over pace | 10.4% | 167.1 | 3.56 | 0.97 | 87 |
| `starpower` — fill both driver seats first | 10.1% | 180.9 | 3.24 | 1.42 | 89 |
| `carfirst` — always take the car | 7.7% | 138.5 | 3.94 | 1.53 | 85 |
| `first` — take the leftmost legal item | 5.3% | 127.1 | 4.10 | 1.58 | 85 |

Enumerating all 240 legal drafts wins 40% of the time. The gradient 5.3 → 17.7 → 40 is what the
draft measures.

**Findings worth keeping:**

- *Greedy early is a trap.* `starpower` ends up with worse drivers than `carfirst` — 35288 against
  35895 — because filling both seats from the first two rolls forfeits later choice. `carfirst` is
  the second-worst line in the game.
- *A greedy picker must compare like with like.* Car reliability averages in the seventies while
  principal development averages in the mid-eighties, so a naive "take the most reliable thing"
  never took a car at all. Score each item by its distance above the roster mean **for its own
  slot** — lead drivers rate higher, so a per-item baseline biases toward the number two seat.
- *σ is the wrong lever, and "perfect-play title%" is not σ-independent.* Sweeping race σ from 1.5 to
  7.5 changed scoring teams from 5 to 7 and nothing else, while inflating the measured ceiling — more
  noise mechanically raises a best-of-240. Left at 2.5 qualifying / 1.5 race.
- *The midfield lockout is structural.* Compressing the grid toward the leader by 35% and 55% moved
  Racing Bulls from 7.8 points to 7.8 and 7.4; raising midfield ratings by 6–10 moved it from 5.4 to
  5.0. The cause is not the ratings — the top four teams field eight cars and only ten positions
  score. Left alone.

---

## The field

No rival AI. The eleven teams are the real 2026 grid — McLaren, Ferrari, Red Bull, Mercedes, Aston
Martin, Alpine, Haas, Racing Bulls, Williams, Audi, Cadillac — with their actual cars, driver
pairings, engineers and principals, anchored to the real constructors' table after twelve rounds
(Mercedes 425 down to Cadillac 0). Identical on every seed, which buys three things: a constant
benchmark, so a result measures the draft and not a soft field; a goal legible cold ("beat Red Bull
and McLaren with a team assembled out of history"); and no adaptive-AI trap, where two players on one
seed would face different fields.

The 2026 entries are rollable like any other team-era. A roll can land on 2026 McLaren, and taking
their car means racing the real McLaren with a hole in it.

---

## Scoring and sharing

Season score is championship points. The leaderboard is all-time, one row per player — their best
season — ranked by points, tie-broken by wins, then podiums, then earliest submission.

```
Lights Out · Season 142
P2 of 12 · 287 pts · 1 DNF
🏁🥈🥇🥉🏁🥇✖️🥈🥇🥈
```

Three lines, no strategy spoilers, and the emoji row shows the shape of a season including the race
that went wrong. The position carries its own denominator: "P2" means nothing cold, "P2 of 12" does.
Naming your team is opt-in.

---

## Determinism contract

Every rule above must respect this, because the leaderboard depends on it.

- A season is a pure function: `(seed, []int) → SeasonResult`. No wall-clock time, no map iteration
  order, no unseeded randomness, no platform-varying floating point.
- The RNG is explicitly seeded and advances in a fixed order. **Any change to draw order is a
  breaking change** and must bump the sim version — adding a draw anywhere moves every downstream
  result.
- Seasons record the sim version they were issued under. Old runs are frozen, never re-verified
  against new code. Replacing the fictional circuits bumped the version even though no rule changed,
  because `GenerateSeason` consumed its RNG differently.
- The same source compiles to native and WASM, so floating point is avoided entirely: all `int64`
  fixed-point, normals from a sum of twelve uniforms rather than Box–Muller. The parity test exists
  to catch any regression here; the fix is always integer arithmetic, never hoping it does not
  matter.
