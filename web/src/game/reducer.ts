import type { Card, SeasonDescriptor, SeasonResult } from './types';

export type Phase = 'loading' | 'choosing' | 'racing' | 'complete' | 'error';

export interface GameState {
  phase: Phase;
  season: SeasonDescriptor | null;
  /** Zero-based index of the development window being chosen. */
  window: number;
  /** The card highlighted but not yet committed, or null. */
  pick: number | null;
  /** Committed card indices, one per window. */
  picks: number[];
  /** Races resolved so far, from RunPartial after each pick. */
  result: SeasonResult | null;
  /** How many of result.races have been shown. Drives the reveal. */
  revealed: number;
  error: string | null;
  submitted: boolean;
}

export const initialState: GameState = {
  phase: 'loading',
  season: null,
  window: 0,
  pick: null,
  picks: [],
  result: null,
  revealed: 0,
  error: null,
  submitted: false,
};

export type Action =
  | { type: 'SEASON_LOADED'; season: SeasonDescriptor }
  | { type: 'LOAD_FAILED'; error: string }
  | { type: 'PICK_CARD'; index: number }
  | { type: 'CONFIRM_PICK' }
  | { type: 'RACES_RESOLVED'; result: SeasonResult }
  | { type: 'REVEAL_NEXT' }
  | { type: 'NEXT_WINDOW' }
  | { type: 'SUBMITTED' }
  | { type: 'RESET' };

/** The deal for the window currently being chosen. */
export function currentDeal(state: GameState): Card[] {
  if (!state.season) return [];
  return state.season.deals[state.window] ?? [];
}

/** Whether every race resolved so far has been shown. */
export function revealComplete(state: GameState): boolean {
  return !!state.result && state.revealed >= state.result.races.length;
}

export function reducer(state: GameState, action: Action): GameState {
  switch (action.type) {
    case 'SEASON_LOADED':
      return { ...initialState, phase: 'choosing', season: action.season };

    case 'LOAD_FAILED':
      return { ...state, phase: 'error', error: action.error };

    case 'PICK_CARD': {
      if (state.phase !== 'choosing') return state;
      const deal = currentDeal(state);
      // Out-of-range indices are ignored rather than clamped: a pick is an
      // identity, not a magnitude, so there is no sensible nearest value.
      if (action.index < 0 || action.index >= deal.length) return state;
      return { ...state, pick: action.index };
    }

    case 'CONFIRM_PICK': {
      if (state.phase !== 'choosing' || state.pick === null || !state.season) return state;
      return {
        ...state,
        picks: [...state.picks, state.pick],
        pick: null,
        phase: 'racing',
      };
    }

    case 'RACES_RESOLVED': {
      // Keep whatever has already been revealed: the new races append.
      const last = state.picks.length >= (state.season?.deals.length ?? 0);
      return {
        ...state,
        result: action.result,
        phase: last ? 'complete' : 'racing',
      };
    }

    case 'REVEAL_NEXT':
      if (!state.result) return state;
      return {
        ...state,
        revealed: Math.min(state.revealed + 1, state.result.races.length),
      };

    case 'NEXT_WINDOW': {
      if (!state.season) return state;
      if (state.picks.length >= state.season.deals.length) {
        return { ...state, phase: 'complete' };
      }
      return { ...state, window: state.picks.length, phase: 'choosing' };
    }

    case 'SUBMITTED':
      return { ...state, submitted: true };

    case 'RESET':
      return state.season
        ? { ...initialState, phase: 'choosing', season: state.season }
        : initialState;

    default:
      return state;
  }
}
