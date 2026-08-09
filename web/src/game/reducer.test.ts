import { describe, expect, it } from 'vitest';
import {
  currentRoll,
  draftComplete,
  filledSlots,
  initialState,
  reducer,
  reelComplete,
  standingsAfter,
  takeable,
  type GameState,
} from './reducer';
import {
  carOverall,
  driverOverall,
  engineerOverall,
  ITEM_CAR,
  ITEM_DRIVER_A,
  ITEM_DRIVER_B,
  ITEM_ENGINEER,
  ITEM_PRINCIPAL,
  itemsOf,
  principalOverall,
  ROLL_COUNT,
  type SeasonDescriptor,
  type SeasonResult,
  type TeamEra,
} from './types';

function era(id: string, year = 1988): TeamEra {
  const d = (n: string) => ({
    id: n,
    name: n,
    pace: 80,
    racecraft: 80,
    consistency: 80,
    composure: 80,
  });
  return {
    id,
    team: id,
    year,
    era_id: '1980s',
    livery: '#ffffff',
    car: { id: `${id}-car`, name: `${id} car`, power: 90, cornering: 90, aero: 90, reliability: 80 },
    drivers: [d(`${id}-a`), d(`${id}-b`)],
    engineer: { id: `${id}-e`, name: `${id} eng`, setup: 80, strategy: 80, ops: 80 },
    principal: { id: `${id}-p`, name: `${id} tp`, development: 80, leadership: 80, nerve: 80 },
  };
}

const season: SeasonDescriptor = {
  id: 1,
  seed: '4242',
  sim_version: 'test',
  calendar: [],
  field: [],
  rolls: [era('a'), era('b'), era('c'), era('d'), era('e')],
  closes_at: '',
};

function loaded(): GameState {
  return reducer(initialState, { type: 'SEASON_LOADED', season, mode: 'daily' });
}

function resultWith(n: number): SeasonResult {
  return {
    sim_version: 'test',
    seed: 1,
    rolls: season.rolls,
    picks: [],
    lineup: {
      car: season.rolls[0].car,
      drivers: season.rolls[0].drivers,
      engineer: season.rolls[0].engineer,
      principal: season.rolls[0].principal,
    },
    races: Array.from({ length: n }, (_, i) => ({
      round: i + 1,
      circuit: `C${i}`,
      safety_car: false,
      entries: [
        {
          team_id: 0,
          entry: 0,
          driver_id: 'x',
          driver: 'X',
          grid: 1,
          finish: 1,
          dnf: false,
          dnf_reason: '' as const,
          points: 25,
        },
        {
          team_id: 1,
          entry: 0,
          driver_id: 'y',
          driver: 'Y',
          grid: 2,
          finish: 2,
          dnf: false,
          dnf_reason: '' as const,
          points: 18,
        },
      ],
    })),
    standings: [
      { team_id: 0, name: 'Your Team', points: 0, wins: 0, podiums: 0, dnfs: 0 },
      { team_id: 1, name: 'Rival', points: 0, wins: 0, podiums: 0, dnfs: 0 },
    ],
    drivers: [],
    player: { team_id: 0, name: 'Your Team', points: 0, wins: 0, podiums: 0, dnfs: 0 },
    player_position: 1,
    share: 'x',
  };
}

describe('loading', () => {
  it('starts drafting once the season arrives', () => {
    const s = loaded();
    expect(s.phase).toBe('drafting');
    expect(currentRoll(s)?.id).toBe('a');
  });

  it('records the mode so free play is never submitted', () => {
    const s = reducer(initialState, { type: 'SEASON_LOADED', season, mode: 'free' });
    expect(s.mode).toBe('free');
  });

  it('reports a load failure', () => {
    const s = reducer(initialState, { type: 'LOAD_FAILED', error: 'nope' });
    expect(s.phase).toBe('error');
    expect(s.error).toBe('nope');
  });
});

describe('slot legality', () => {
  it('closes a slot once it is full', () => {
    expect(takeable([ITEM_CAR], ITEM_CAR)).toBe(false);
    expect(takeable([ITEM_DRIVER_A, ITEM_DRIVER_B], ITEM_DRIVER_A)).toBe(false);
    expect(takeable([ITEM_DRIVER_A], ITEM_DRIVER_B)).toBe(true);
  });

  it('refuses a pick that would leave a slot unfillable', () => {
    // Four picks in, only the principal missing: nothing else is legal even
    // though the car slot is technically empty in this contrived case.
    const picks = [ITEM_DRIVER_A, ITEM_DRIVER_B, ITEM_ENGINEER, ITEM_CAR];
    expect(takeable(picks, ITEM_PRINCIPAL)).toBe(true);
    expect(takeable(picks, ITEM_CAR)).toBe(false);
  });

  it('never lets a draft finish with an empty slot', () => {
    // Walk every legal line and assert the team is complete at the end.
    const walk = (picks: number[]) => {
      if (picks.length === ROLL_COUNT) {
        const f = filledSlots(picks);
        expect(f).toEqual({ car: 1, driver: 2, engineer: 1, principal: 1 });
        return;
      }
      for (let k = 0; k < 5; k++) if (takeable(picks, k)) walk([...picks, k]);
    };
    walk([]);
  });
});

describe('drafting', () => {
  it('highlights a legal item and ignores an illegal one', () => {
    let s = loaded();
    s = reducer(s, { type: 'PICK_ITEM', kind: ITEM_CAR });
    expect(s.pick).toBe(ITEM_CAR);
    s = reducer(s, { type: 'CONFIRM_PICK' });
    // The car slot is full now, so picking a car again changes nothing.
    const before = s.pick;
    s = reducer(s, { type: 'PICK_ITEM', kind: ITEM_CAR });
    expect(s.pick).toBe(before);
  });

  it('advances the roll on confirm', () => {
    let s = loaded();
    s = reducer(s, { type: 'PICK_ITEM', kind: ITEM_CAR });
    s = reducer(s, { type: 'CONFIRM_PICK' });
    expect(s.roll).toBe(1);
    expect(currentRoll(s)?.id).toBe('b');
    expect(s.picks).toEqual([ITEM_CAR]);
  });

  it('never commits more picks than there are rolls', () => {
    let s = loaded();
    for (const kind of [ITEM_CAR, ITEM_DRIVER_A, ITEM_DRIVER_B, ITEM_ENGINEER, ITEM_PRINCIPAL]) {
      s = reducer(s, { type: 'PICK_ITEM', kind });
      s = reducer(s, { type: 'CONFIRM_PICK' });
    }
    expect(s.picks).toHaveLength(ROLL_COUNT);
    expect(draftComplete(s)).toBe(true);
    expect(s.phase).toBe('reel');
    s = reducer(s, { type: 'CONFIRM_PICK' });
    expect(s.picks).toHaveLength(ROLL_COUNT);
  });
});

describe('the reel', () => {
  it('plays one race per tick and finishes', () => {
    let s: GameState = { ...loaded(), phase: 'reel' };
    s = reducer(s, { type: 'SEASON_RESOLVED', result: resultWith(3) });
    expect(s.reelRound).toBe(0);
    expect(reelComplete(s)).toBe(false);
    s = reducer(s, { type: 'REEL_TICK' });
    s = reducer(s, { type: 'REEL_TICK' });
    expect(s.phase).toBe('reel');
    s = reducer(s, { type: 'REEL_TICK' });
    expect(reelComplete(s)).toBe(true);
    expect(s.phase).toBe('complete');
  });

  it('never plays past the last race', () => {
    let s: GameState = { ...loaded(), phase: 'reel' };
    s = reducer(s, { type: 'SEASON_RESOLVED', result: resultWith(2) });
    for (let i = 0; i < 10; i++) s = reducer(s, { type: 'REEL_TICK' });
    expect(s.reelRound).toBe(2);
  });

  it('skips straight to the end', () => {
    let s: GameState = { ...loaded(), phase: 'reel' };
    s = reducer(s, { type: 'SEASON_RESOLVED', result: resultWith(10) });
    s = reducer(s, { type: 'SKIP_REEL' });
    expect(s.reelRound).toBe(10);
    expect(s.phase).toBe('complete');
  });
});

describe('standingsAfter', () => {
  it('accumulates only the races played so far', () => {
    const r = resultWith(4);
    expect(standingsAfter(r, 0)[0].points).toBe(0);
    const two = standingsAfter(r, 2);
    expect(two[0]).toMatchObject({ team_id: 0, points: 50, wins: 2, podiums: 2 });
    expect(standingsAfter(r, 4)[0].points).toBe(100);
  });

  it('sorts by the same total order the sim uses', () => {
    const table = standingsAfter(resultWith(3), 3);
    expect(table.map((s) => s.team_id)).toEqual([0, 1]);
  });
});

describe('reset', () => {
  it('returns to the first roll but keeps the season', () => {
    let s = loaded();
    s = reducer(s, { type: 'PICK_ITEM', kind: ITEM_CAR });
    s = reducer(s, { type: 'CONFIRM_PICK' });
    s = reducer(s, { type: 'RESET' });
    expect(s.picks).toEqual([]);
    expect(s.phase).toBe('drafting');
    expect(s.season).toBe(season);
  });
});

describe('derived ratings', () => {
  // Overall is computed, never sent. These weights mirror roster.go, and a
  // divergence would show a number the race does not use.
  it('matches the sim formulas', () => {
    const te = era('x');
    expect(carOverall(te.car)).toBe(Math.floor((30 * 90 + 30 * 90 + 30 * 90 + 10 * 80) / 100));
    expect(driverOverall(te.drivers[0])).toBe(80);
    expect(engineerOverall(te.engineer)).toBe(80);
    expect(principalOverall(te.principal)).toBe(80);
  });

  it('offers exactly five items per roll, one per item kind', () => {
    const items = itemsOf(era('x'));
    expect(items.map((i) => i.kind)).toEqual([0, 1, 2, 3, 4]);
    expect(items.map((i) => i.slot)).toEqual(['car', 'driver', 'driver', 'engineer', 'principal']);
  });
});
