import { describe, it, expect } from 'vitest';
import { reducer, initialState, currentDeal, revealComplete, type GameState } from './reducer';
import { riskPips, effectLabel, type Card, type SeasonDescriptor, type SeasonResult } from './types';

const card = (id: string, chassis: number, engine: number, aero: number): Card => ({
  id,
  name: id.toUpperCase(),
  blurb: 'a part',
  effect: { chassis, engine, aero },
});

const season: SeasonDescriptor = {
  id: 1,
  seed: '2497907803379454',
  sim_version: '2.0.0',
  calendar: Array.from({ length: 10 }, (_, i) => ({
    name: `Circuit ${i + 1}`,
    archetype: 'balanced' as const,
    profile: { chassis: 350, engine: 350, aero: 300, overtake_difficulty: 280 },
  })),
  field: [],
  deals: Array.from({ length: 5 }, (_, w) => [
    card(`a${w}`, 250, 0, 0),
    card(`b${w}`, 0, 250, 0),
    card(`c${w}`, 0, 0, 250),
  ]),
  window_rounds: [0, 2, 4, 6, 8],
  closes_at: '2026-09-03T00:00:00Z',
};

function loaded(): GameState {
  return reducer(initialState, { type: 'SEASON_LOADED', season });
}

function resultWith(races: number): SeasonResult {
  return {
    sim_version: '2.0.0',
    seed: 1,
    build: [],
    races: Array.from({ length: races }, (_, i) => ({
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
}

describe('SEASON_LOADED', () => {
  it('opens on the first window with nothing picked', () => {
    const s = loaded();
    expect(s.phase).toBe('choosing');
    expect(s.window).toBe(0);
    expect(s.pick).toBeNull();
    expect(currentDeal(s)).toHaveLength(3);
  });
});

describe('PICK_CARD', () => {
  it('highlights a card without committing it', () => {
    const s = reducer(loaded(), { type: 'PICK_CARD', index: 2 });
    expect(s.pick).toBe(2);
    expect(s.picks).toEqual([]);
  });

  it('ignores an index outside the deal', () => {
    const base = loaded();
    expect(reducer(base, { type: 'PICK_CARD', index: 3 })).toBe(base);
    expect(reducer(base, { type: 'PICK_CARD', index: -1 })).toBe(base);
  });

  it('lets the highlight change before committing', () => {
    let s = reducer(loaded(), { type: 'PICK_CARD', index: 0 });
    s = reducer(s, { type: 'PICK_CARD', index: 1 });
    expect(s.pick).toBe(1);
  });

  it('is ignored outside the choosing phase', () => {
    const s = { ...loaded(), phase: 'racing' as const };
    expect(reducer(s, { type: 'PICK_CARD', index: 1 })).toBe(s);
  });
});

describe('CONFIRM_PICK', () => {
  it('commits the highlighted card and starts racing', () => {
    let s = reducer(loaded(), { type: 'PICK_CARD', index: 2 });
    s = reducer(s, { type: 'CONFIRM_PICK' });
    expect(s.picks).toEqual([2]);
    expect(s.pick).toBeNull();
    expect(s.phase).toBe('racing');
  });

  it('does nothing when no card is highlighted', () => {
    const base = loaded();
    expect(reducer(base, { type: 'CONFIRM_PICK' })).toBe(base);
  });

  it('never commits more picks than there are windows', () => {
    let s = loaded();
    for (let w = 0; w < 5; w++) {
      s = reducer(s, { type: 'PICK_CARD', index: w % 3 });
      s = reducer(s, { type: 'CONFIRM_PICK' });
      s = reducer(s, { type: 'RACES_RESOLVED', result: resultWith((w + 1) * 2) });
      s = reducer(s, { type: 'NEXT_WINDOW' });
    }
    expect(s.picks).toHaveLength(5);
    expect(s.phase).toBe('complete');
  });
});

describe('NEXT_WINDOW', () => {
  it('advances the window to match committed picks', () => {
    let s = reducer(loaded(), { type: 'PICK_CARD', index: 0 });
    s = reducer(s, { type: 'CONFIRM_PICK' });
    s = reducer(s, { type: 'RACES_RESOLVED', result: resultWith(2) });
    s = reducer(s, { type: 'NEXT_WINDOW' });
    expect(s.window).toBe(1);
    expect(s.phase).toBe('choosing');
    expect(currentDeal(s)[0].id).toBe('a1');
  });
});

describe('REVEAL_NEXT', () => {
  it('reveals races one at a time and stops at the end', () => {
    let s = reducer(loaded(), { type: 'RACES_RESOLVED', result: resultWith(4) });
    expect(s.revealed).toBe(0);
    expect(revealComplete(s)).toBe(false);
    for (let i = 0; i < 6; i++) s = reducer(s, { type: 'REVEAL_NEXT' });
    expect(s.revealed).toBe(4);
    expect(revealComplete(s)).toBe(true);
  });

  it('does nothing before any races resolve', () => {
    const base = loaded();
    expect(reducer(base, { type: 'REVEAL_NEXT' })).toBe(base);
  });

  it('keeps what was already revealed when more races arrive', () => {
    let s = reducer(loaded(), { type: 'RACES_RESOLVED', result: resultWith(2) });
    s = reducer(s, { type: 'REVEAL_NEXT' });
    s = reducer(s, { type: 'REVEAL_NEXT' });
    expect(s.revealed).toBe(2);
    s = reducer(s, { type: 'RACES_RESOLVED', result: resultWith(4) });
    expect(s.revealed).toBe(2);
    expect(revealComplete(s)).toBe(false);
  });
});

describe('RESET', () => {
  it('returns to the first window but keeps the season', () => {
    let s = reducer(loaded(), { type: 'PICK_CARD', index: 1 });
    s = reducer(s, { type: 'CONFIRM_PICK' });
    s = reducer(s, { type: 'RESET' });
    expect(s.picks).toEqual([]);
    expect(s.window).toBe(0);
    expect(s.phase).toBe('choosing');
    expect(s.season).toBe(season);
  });
});

describe('card presentation', () => {
  it('scores an expensive card riskier than a cheap one', () => {
    expect(riskPips(card('big', 260, 0, 0))).toBeGreaterThan(riskPips(card('small', 140, 0, 0)));
  });

  it('scores an aero card safer than an engine card of equal cost', () => {
    expect(riskPips(card('aero', 0, 0, 250))).toBeLessThan(riskPips(card('eng', 0, 250, 0)));
  });

  it('always returns a pip count in range', () => {
    for (const c of [card('z', 0, 0, 0), card('big', 260, 260, 260), card('s', 140, 0, 0)]) {
      const p = riskPips(c);
      expect(p).toBeGreaterThanOrEqual(1);
      expect(p).toBeLessThanOrEqual(5);
    }
  });

  it('labels only the areas a card touches', () => {
    expect(effectLabel(card('x', 0, 120, 80))).toBe('+12 Engine · +8 Aero');
    expect(effectLabel(card('y', 250, 0, 0))).toBe('+25 Chassis');
  });
});
