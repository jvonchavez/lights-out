import { useCallback, useEffect, useMemo, useReducer, useRef, useState } from 'react';
import {
  fetchLeaderboard,
  fetchTodaySeason,
  displayName,
  setDisplayName,
  submitRun,
} from './game/api';
import {
  currentRoll,
  initialState,
  reducer,
  takeable,
  type GameState,
} from './game/reducer';
import { loadSim, type SimAPI } from './game/wasm';
import {
  ROLL_COUNT,
  itemsOf,
  type LeaderboardEntry,
  type SeasonDescriptor,
} from './game/types';
import { Calendar } from './components/Calendar';
import { DriverStandings } from './components/DriverStandings';
import { Leaderboard } from './components/Leaderboard';
import { LineupPanel } from './components/LineupPanel';
import { RaceReel } from './components/RaceReel';
import { SeasonComplete } from './components/SeasonComplete';
import { Standings } from './components/Standings';
import { TeamEraCard } from './components/TeamEraCard';

/**
 * How long each race holds the screen. Ten races plus a beat at the end is
 * about forty-five seconds, which is the target: long enough that a DNF
 * lands as a moment, short enough that a second run is free.
 */
const REEL_MS = 4000;

export default function App() {
  const [state, dispatch] = useReducer(reducer, initialState);
  const [sim, setSim] = useState<SimAPI | null>(null);
  const simRef = useRef<SimAPI | null>(null);
  const [name, setName] = useState(displayName);
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [board, setBoard] = useState<LeaderboardEntry[] | null>(null);

  // Two independent loads. The season renders the draft immediately; the
  // WASM module is ~1.25 MB and is only needed once five picks are in.
  useEffect(() => {
    fetchTodaySeason()
      .then((season) => dispatch({ type: 'SEASON_LOADED', season, mode: 'daily' }))
      .catch((e: Error) => dispatch({ type: 'LOAD_FAILED', error: e.message }));
    loadSim()
      .then((api) => {
        simRef.current = api;
        setSim(api);
      })
      .catch((e: Error) => dispatch({ type: 'LOAD_FAILED', error: e.message }));
  }, []);

  // The draft is complete: resolve the whole season at once. There are no
  // in-season decisions, so there is nothing to resolve incrementally.
  useEffect(() => {
    if (state.phase !== 'reel' || state.result || !state.season) return;
    const api = simRef.current;
    if (!api) return;
    try {
      const result = api.runSeason(state.season.seed, state.picks);
      dispatch({ type: 'SEASON_RESOLVED', result });
    } catch (e) {
      dispatch({ type: 'LOAD_FAILED', error: (e as Error).message });
    }
  }, [state.phase, state.result, state.season, state.picks, sim]);

  // The reel itself: one race per tick until the season runs out.
  useEffect(() => {
    if (state.phase !== 'reel' || !state.result) return;
    if (state.reelRound >= state.result.races.length) return;
    const t = setTimeout(() => dispatch({ type: 'REEL_TICK' }), REEL_MS);
    return () => clearTimeout(t);
  }, [state.phase, state.result, state.reelRound]);

  /**
   * Free play generates its own seed and never posts a run. GenerateSeason
   * is a pure function compiled into the WASM module, so this needs no
   * backend at all -- which is why unlimited replays cost nothing.
   */
  const playFree = useCallback(() => {
    const api = simRef.current;
    if (!api) return;
    const seed = String(Math.floor(Math.random() * 9_007_199_254_740_991));
    try {
      const gen = api.generateSeason(seed);
      const season: SeasonDescriptor = {
        id: 0,
        seed,
        sim_version: gen.sim_version,
        calendar: gen.calendar,
        field: gen.rivals,
        rolls: gen.rolls,
        closes_at: '',
      };
      setBoard(null);
      setSubmitError(null);
      dispatch({ type: 'SEASON_LOADED', season, mode: 'free' });
    } catch (e) {
      dispatch({ type: 'LOAD_FAILED', error: (e as Error).message });
    }
  }, []);

  async function submit() {
    if (!state.season || !state.result) return;
    setSubmitting(true);
    setSubmitError(null);
    try {
      setDisplayName(name);
      await submitRun(state.season.id, state.picks, name);
      dispatch({ type: 'SUBMITTED' });
      setBoard(await fetchLeaderboard(state.season.id));
    } catch (e) {
      setSubmitError((e as Error).message);
    } finally {
      setSubmitting(false);
    }
  }

  // Team colours and names, keyed by team ID, for the reel and the tables.
  const { colours, names } = useMemo(() => {
    const colours = new Map<number, string>([[0, '#e10600']]);
    const names = new Map<number, string>([[0, 'Your Team']]);
    for (const t of state.season?.field ?? []) {
      colours.set(t.id, t.livery);
      names.set(t.id, t.name);
    }
    return { colours, names };
  }, [state.season]);

  if (state.phase === 'error') {
    return (
      <Shell>
        <p className="text-accent" data-testid="error">
          {state.error}
        </p>
      </Shell>
    );
  }

  if (!state.season) {
    return (
      <Shell>
        <p className="text-muted">Loading today&rsquo;s season…</p>
      </Shell>
    );
  }

  const season = state.season;
  const roll = currentRoll(state);

  return (
    <Shell mode={state.mode} version={sim?.version ?? season.sim_version} seed={season.seed}>
      <div className="grid gap-5 lg:grid-cols-[1fr_320px]">
        <main className="space-y-5">
          {state.phase === 'drafting' && roll && (
            <Draft state={state} dispatch={dispatch} />
          )}

          {state.phase === 'reel' && state.result && (
            <div key={state.reelRound}>
              <RaceReel
                result={state.result}
                calendar={season.calendar}
                round={Math.max(1, state.reelRound + 1)}
                colours={colours}
                names={names}
                onSkip={() => dispatch({ type: 'SKIP_REEL' })}
              />
            </div>
          )}

          {state.phase === 'reel' && !state.result && (
            <p className="text-muted" data-testid="resolving">
              Lights out…
            </p>
          )}

          {state.phase === 'complete' && state.result && (
            <>
              <SeasonComplete
                result={state.result}
                mode={state.mode}
                onPlayAgain={playFree}
                onSubmit={submit}
                submitting={submitting}
                submitted={state.submitted}
                error={submitError}
                name={name}
                onNameChange={setName}
              />
              <Standings standings={state.result.standings} />
              <DriverStandings drivers={state.result.drivers} colours={colours} />
              {board && <Leaderboard entries={board} />}
            </>
          )}
        </main>

        <aside className="space-y-5">
          <LineupPanel rolls={season.rolls} picks={state.picks} />
          <Calendar
            calendar={season.calendar}
            round={state.reelRound}
            results={
              state.result
                ? state.result.races.slice(0, state.reelRound).map((r) => {
                    const mine = r.entries.filter((e) => e.team_id === 0);
                    const best = mine.filter((e) => !e.dnf).sort((a, b) => a.finish - b.finish)[0];
                    return {
                      finish: best?.finish ?? 0,
                      dnf: mine.length > 0 && mine.every((e) => e.dnf),
                      points: mine.reduce((n, e) => n + e.points, 0),
                    };
                  })
                : undefined
            }
          />
        </aside>
      </div>
    </Shell>
  );
}

function Draft({
  state,
  dispatch,
}: {
  state: GameState;
  dispatch: (a: { type: 'PICK_ITEM'; kind: number } | { type: 'CONFIRM_PICK' }) => void;
}) {
  const roll = currentRoll(state);
  if (!roll) return null;
  const picked = state.pick;
  const item = picked === null ? null : itemsOf(roll).find((i) => i.kind === picked);

  return (
    <>
      <div className="rounded-lg border border-edge bg-panel px-4 py-3">
        <p className="text-[11px] uppercase tracking-widest text-muted">
          Roll {state.picks.length + 1} of {ROLL_COUNT}
        </p>
        <p className="mt-1 text-sm">
          Take <span className="font-semibold">one</span> thing from this team. It locks, and you
          roll again.
        </p>
      </div>

      <TeamEraCard
        era={roll}
        picked={picked}
        canTake={(kind) => takeable(state.picks, kind)}
        onPick={(kind) => dispatch({ type: 'PICK_ITEM', kind })}
      />

      <button
        type="button"
        data-testid="confirm-pick"
        disabled={picked === null}
        onClick={() => dispatch({ type: 'CONFIRM_PICK' })}
        className="w-full rounded bg-accent px-4 py-3 text-sm font-semibold text-white transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-40"
      >
        {item ? `Sign ${item.name} — it locks` : 'Choose one'}
      </button>
    </>
  );
}

function Shell({
  children,
  mode,
  version,
  seed,
}: {
  children: React.ReactNode;
  mode?: 'daily' | 'free';
  version?: string;
  seed?: string;
}) {
  const label = seed ? Number(BigInt(seed) % 1000n) : null;
  return (
    <div className="mx-auto max-w-5xl px-4 py-6">
      <header className="mb-5 flex flex-wrap items-baseline justify-between gap-2">
        <h1 className="text-lg font-bold tracking-tight">
          Lights Out
          {label !== null && (
            <span className="ml-2 text-sm font-normal text-muted">
              {mode === 'free' ? 'Free play' : `Season ${label}`}
            </span>
          )}
        </h1>
        {version && <span className="font-mono text-[11px] text-muted">sim {version}</span>}
      </header>
      {children}
    </div>
  );
}
