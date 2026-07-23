import { describe, it, expect } from 'vitest';
import {
  reducer,
  initialState,
  emptyDecision,
  budgetFor,
  remaining,
  spent,
  type GameState,
} from './reducer';
import type { SeasonDescriptor } from './types';

const season: SeasonDescriptor = {
  id: 1,
  seed: '2497907803379454',
  sim_version: '1.1.0',
  calendar: Array.from({ length: 10 }, (_, i) => ({
    name: `Circuit ${i + 1}`,
    archetype: 'balanced' as const,
    profile: { chassis: 350, engine: 350, aero: 300, overtake_difficulty: 280 },
  })),
  field: [],
  budgets: Array(10).fill(100),
  closes_at: '2026-09-03T00:00:00Z',
};

function loaded(): GameState {
  return reducer(initialState, { type: 'SEASON_LOADED', season });
}

describe('SEASON_LOADED', () => {
  it('moves to allocating with an empty allocation', () => {
    const s = loaded();
    expect(s.phase).toBe('allocating');
    expect(s.round).toBe(0);
    expect(s.allocation).toEqual(emptyDecision);
    expect(budgetFor(s)).toBe(100);
  });
});

describe('ALLOCATE', () => {
  it('sets a value within budget', () => {
    const s = reducer(loaded(), { type: 'ALLOCATE', area: 'chassis', value: 40 });
    expect(s.allocation.chassis).toBe(40);
    expect(remaining(s)).toBe(60);
  });

  it('clamps so the sliders can never exceed the budget', () => {
    let s = loaded();
    s = reducer(s, { type: 'ALLOCATE', area: 'chassis', value: 60 });
    s = reducer(s, { type: 'ALLOCATE', area: 'engine', value: 60 });
    expect(spent(s.allocation)).toBeLessThanOrEqual(100);
    expect(s.allocation.engine).toBe(40);
  });

  it('clamps negatives to zero', () => {
    const s = reducer(loaded(), { type: 'ALLOCATE', area: 'aero', value: -20 });
    expect(s.allocation.aero).toBe(0);
  });

  it('floors fractional values, because the server takes integers', () => {
    const s = reducer(loaded(), { type: 'ALLOCATE', area: 'aero', value: 33.7 });
    expect(s.allocation.aero).toBe(33);
  });

  it('lets an area be reduced to free up budget', () => {
    let s = loaded();
    s = reducer(s, { type: 'ALLOCATE', area: 'chassis', value: 100 });
    expect(remaining(s)).toBe(0);
    s = reducer(s, { type: 'ALLOCATE', area: 'chassis', value: 20 });
    expect(remaining(s)).toBe(80);
    s = reducer(s, { type: 'ALLOCATE', area: 'engine', value: 50 });
    expect(s.allocation.engine).toBe(50);
  });

  it('is ignored outside the allocating phase', () => {
    const s = { ...loaded(), phase: 'complete' as const };
    expect(reducer(s, { type: 'ALLOCATE', area: 'chassis', value: 50 })).toBe(s);
  });
});

describe('CONFIRM_RACE', () => {
  it('records the decision and advances the round', () => {
    let s = loaded();
    s = reducer(s, { type: 'ALLOCATE', area: 'chassis', value: 50 });
    s = reducer(s, { type: 'CONFIRM_RACE' });
    expect(s.decisions).toHaveLength(1);
    expect(s.decisions[0].chassis).toBe(50);
    expect(s.round).toBe(1);
    expect(s.allocation).toEqual(emptyDecision);
    expect(s.phase).toBe('allocating');
  });

  it('allows underspending', () => {
    let s = loaded();
    s = reducer(s, { type: 'CONFIRM_RACE' });
    expect(s.decisions[0]).toEqual(emptyDecision);
  });

  it('moves to reviewing after the final round', () => {
    let s = loaded();
    for (let i = 0; i < 10; i++) {
      s = reducer(s, { type: 'ALLOCATE', area: 'aero', value: 30 });
      s = reducer(s, { type: 'CONFIRM_RACE' });
    }
    expect(s.decisions).toHaveLength(10);
    expect(s.phase).toBe('reviewing');
    expect(s.round).toBe(9);
  });

  it('never produces an over-budget decision across a whole season', () => {
    let s = loaded();
    for (let i = 0; i < 10; i++) {
      s = reducer(s, { type: 'ALLOCATE', area: 'chassis', value: 999 });
      s = reducer(s, { type: 'ALLOCATE', area: 'engine', value: 999 });
      s = reducer(s, { type: 'ALLOCATE', area: 'aero', value: 999 });
      s = reducer(s, { type: 'CONFIRM_RACE' });
    }
    for (const d of s.decisions) {
      expect(spent(d)).toBeLessThanOrEqual(100);
    }
  });
});

describe('REVEAL_NEXT', () => {
  it('reveals races one at a time and stops at the end', () => {
    const result = {
      sim_version: '1.1.0',
      seed: 1,
      races: Array.from({ length: 3 }, (_, i) => ({
        round: i + 1,
        circuit: 'x',
        safety_car: false,
        cars: [],
      })),
      standings: [],
      player: { team_id: 0, name: 'you', points: 0, wins: 0, podiums: 0, dnfs: 0 },
      player_position: 1,
      share: '',
    };
    let s = reducer({ ...loaded(), phase: 'reviewing' }, { type: 'SEASON_RESOLVED', result });
    expect(s.phase).toBe('complete');
    expect(s.revealed).toBe(0);
    for (let i = 0; i < 5; i++) s = reducer(s, { type: 'REVEAL_NEXT' });
    expect(s.revealed).toBe(3);
  });
});

describe('RESET', () => {
  it('returns to a fresh allocation but keeps the season', () => {
    let s = loaded();
    s = reducer(s, { type: 'ALLOCATE', area: 'chassis', value: 50 });
    s = reducer(s, { type: 'CONFIRM_RACE' });
    s = reducer(s, { type: 'RESET' });
    expect(s.decisions).toEqual([]);
    expect(s.round).toBe(0);
    expect(s.phase).toBe('allocating');
    expect(s.season).toBe(season);
  });
});
