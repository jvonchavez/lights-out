import type { Decision, SeasonDescriptor, SeasonResult } from './types';

export type Phase = 'loading' | 'allocating' | 'reviewing' | 'complete' | 'error';

export interface GameState {
  phase: Phase;
  season: SeasonDescriptor | null;
  /** Zero-based index of the round being allocated. */
  round: number;
  allocation: Decision;
  decisions: Decision[];
  /** Result of the whole season, available once every round is allocated. */
  result: SeasonResult | null;
  /** Per-round view built from result, revealed one race at a time. */
  revealed: number;
  error: string | null;
  submitted: boolean;
}

export const emptyDecision: Decision = { chassis: 0, engine: 0, aero: 0 };

export const initialState: GameState = {
  phase: 'loading',
  season: null,
  round: 0,
  allocation: emptyDecision,
  decisions: [],
  result: null,
  revealed: 0,
  error: null,
  submitted: false,
};

export type Action =
  | { type: 'SEASON_LOADED'; season: SeasonDescriptor }
  | { type: 'LOAD_FAILED'; error: string }
  | { type: 'ALLOCATE'; area: keyof Decision; value: number }
  | { type: 'CONFIRM_RACE' }
  | { type: 'SEASON_RESOLVED'; result: SeasonResult }
  | { type: 'REVEAL_NEXT' }
  | { type: 'SUBMITTED' }
  | { type: 'RESET' };

export function budgetFor(state: GameState): number {
  if (!state.season) return 0;
  return state.season.budgets[state.round] ?? 0;
}

export function spent(d: Decision): number {
  return d.chassis + d.engine + d.aero;
}

export function remaining(state: GameState): number {
  return budgetFor(state) - spent(state.allocation);
}

export function reducer(state: GameState, action: Action): GameState {
  switch (action.type) {
    case 'SEASON_LOADED':
      return {
        ...initialState,
        phase: 'allocating',
        season: action.season,
        allocation: emptyDecision,
      };

    case 'LOAD_FAILED':
      return { ...state, phase: 'error', error: action.error };

    case 'ALLOCATE': {
      if (state.phase !== 'allocating') return state;
      const budget = budgetFor(state);
      const others = spent(state.allocation) - state.allocation[action.area];
      // Clamp so the four sliders can never exceed the round's budget. The
      // server rejects an over-budget allocation outright, so the UI must
      // make one impossible rather than merely discouraged.
      const value = Math.max(0, Math.min(Math.floor(action.value), budget - others));
      return { ...state, allocation: { ...state.allocation, [action.area]: value } };
    }

    case 'CONFIRM_RACE': {
      if (state.phase !== 'allocating' || !state.season) return state;
      const decisions = [...state.decisions, state.allocation];
      const last = decisions.length >= state.season.calendar.length;
      return {
        ...state,
        decisions,
        round: last ? state.round : state.round + 1,
        allocation: emptyDecision,
        phase: last ? 'reviewing' : 'allocating',
      };
    }

    case 'SEASON_RESOLVED':
      return { ...state, result: action.result, phase: 'complete', revealed: 0 };

    case 'REVEAL_NEXT':
      if (!state.result) return state;
      return {
        ...state,
        revealed: Math.min(state.revealed + 1, state.result.races.length),
      };

    case 'SUBMITTED':
      return { ...state, submitted: true };

    case 'RESET':
      return state.season
        ? { ...initialState, phase: 'allocating', season: state.season }
        : initialState;

    default:
      return state;
  }
}
