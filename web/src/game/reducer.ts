import {
  ROLL_COUNT,
  SLOT_CAPACITY,
  SLOT_OF,
  type SeasonDescriptor,
  type SeasonResult,
  type Slot,
  type Standing,
  type TeamEra,
} from './types';

export type Phase = 'loading' | 'drafting' | 'reel' | 'complete' | 'error';

/** Where a run's seed came from, and therefore whether it can be scored. */
export type Mode = 'daily' | 'free';

export interface GameState {
  phase: Phase;
  mode: Mode;
  season: SeasonDescriptor | null;
  /** Zero-based index of the roll being decided. */
  roll: number;
  /** The item highlighted but not yet committed, or null. */
  pick: number | null;
  /** Committed item indices, one per roll. */
  picks: number[];
  result: SeasonResult | null;
  /** How many races the reel has played. Drives the whole animation. */
  reelRound: number;
  error: string | null;
  submitted: boolean;
}

export const initialState: GameState = {
  phase: 'loading',
  mode: 'daily',
  season: null,
  roll: 0,
  pick: null,
  picks: [],
  result: null,
  reelRound: 0,
  error: null,
  submitted: false,
};

export type Action =
  | { type: 'SEASON_LOADED'; season: SeasonDescriptor; mode: Mode }
  | { type: 'LOAD_FAILED'; error: string }
  | { type: 'PICK_ITEM'; kind: number }
  | { type: 'CONFIRM_PICK' }
  | { type: 'SEASON_RESOLVED'; result: SeasonResult }
  | { type: 'REEL_TICK' }
  | { type: 'SKIP_REEL' }
  | { type: 'SUBMITTED' }
  | { type: 'RESET' };

/** The team-era on offer at the roll being decided. */
export function currentRoll(state: GameState): TeamEra | null {
  return state.season?.rolls[state.roll] ?? null;
}

/** How many places of each slot the committed picks have filled. */
export function filledSlots(picks: number[]): Record<Slot, number> {
  const filled: Record<Slot, number> = { car: 0, driver: 0, engineer: 0, principal: 0 };
  for (const p of picks) {
    const slot = SLOT_OF[p];
    if (slot) filled[slot]++;
  }
  return filled;
}

/**
 * takeable reports whether an item can legally be taken at this roll.
 *
 * Two rules, and the second is the one that matters. A slot with no room
 * left is closed, obviously. But an item is also refused when taking it
 * would leave more empty places than there are rolls remaining -- with five
 * rolls and five places that only bites on the last roll, and it is what
 * stops the player drafting themselves into a team with no car.
 */
export function takeable(picks: number[], kind: number): boolean {
  const slot = SLOT_OF[kind];
  if (!slot) return false;
  const filled = filledSlots(picks);
  if (filled[slot] >= SLOT_CAPACITY[slot]) return false;

  const remaining = ROLL_COUNT - picks.length;
  let need = 0;
  for (const s of Object.keys(SLOT_CAPACITY) as Slot[]) {
    let short = SLOT_CAPACITY[s] - filled[s];
    if (s === slot) short--;
    if (short > 0) need += short;
  }
  return need <= remaining - 1;
}

export function draftComplete(state: GameState): boolean {
  return state.picks.length >= ROLL_COUNT;
}

/** Whether the reel has played every race it has. */
export function reelComplete(state: GameState): boolean {
  return !!state.result && state.reelRound >= state.result.races.length;
}

export function reducer(state: GameState, action: Action): GameState {
  switch (action.type) {
    case 'SEASON_LOADED':
      return { ...initialState, phase: 'drafting', season: action.season, mode: action.mode };

    case 'LOAD_FAILED':
      return { ...state, phase: 'error', error: action.error };

    case 'PICK_ITEM': {
      if (state.phase !== 'drafting') return state;
      // Illegal picks are ignored rather than clamped: a pick is an
      // identity, not a magnitude, so there is no sensible nearest value.
      if (!takeable(state.picks, action.kind)) return state;
      return { ...state, pick: action.kind };
    }

    case 'CONFIRM_PICK': {
      if (state.phase !== 'drafting' || state.pick === null) return state;
      if (!takeable(state.picks, state.pick)) return state;
      const picks = [...state.picks, state.pick];
      return {
        ...state,
        picks,
        pick: null,
        roll: picks.length,
        phase: picks.length >= ROLL_COUNT ? 'reel' : 'drafting',
      };
    }

    case 'SEASON_RESOLVED':
      return { ...state, result: action.result, phase: 'reel', reelRound: 0 };

    case 'REEL_TICK': {
      if (!state.result) return state;
      const next = Math.min(state.reelRound + 1, state.result.races.length);
      return {
        ...state,
        reelRound: next,
        phase: next >= state.result.races.length ? 'complete' : 'reel',
      };
    }

    case 'SKIP_REEL':
      if (!state.result) return state;
      return { ...state, reelRound: state.result.races.length, phase: 'complete' };

    case 'SUBMITTED':
      return { ...state, submitted: true };

    case 'RESET':
      return state.season
        ? { ...initialState, phase: 'drafting', season: state.season, mode: state.mode }
        : initialState;

    default:
      return state;
  }
}

/**
 * standingsAfter recomputes the constructors' table from the races the reel
 * has played so far. The full result is already in hand -- this exists so
 * the table climbs race by race instead of appearing finished.
 */
export function standingsAfter(result: SeasonResult, rounds: number): Standing[] {
  const byTeam = new Map<number, Standing>();
  for (const s of result.standings) {
    byTeam.set(s.team_id, { ...s, points: 0, wins: 0, podiums: 0, dnfs: 0 });
  }
  for (const race of result.races.slice(0, rounds)) {
    for (const e of race.entries) {
      const s = byTeam.get(e.team_id);
      if (!s) continue;
      s.points += e.points;
      if (e.dnf) s.dnfs++;
      else if (e.finish === 1) {
        s.wins++;
        s.podiums++;
      } else if (e.finish <= 3) s.podiums++;
    }
  }
  return [...byTeam.values()].sort(
    (a, b) =>
      b.points - a.points ||
      b.wins - a.wins ||
      b.podiums - a.podiums ||
      a.team_id - b.team_id,
  );
}
