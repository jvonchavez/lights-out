import { useCallback, useEffect, useReducer, useRef, useState } from 'react';
import { reducer, initialState, budgetFor, remaining, spent } from './game/reducer';
import { loadSim, type SimAPI } from './game/wasm';
import {
  fetchLeaderboard,
  fetchTodaySeason,
  submitRun,
  displayName,
  setDisplayName,
} from './game/api';
import type { Decision, LeaderboardEntry } from './game/types';
import { Calendar } from './components/Calendar';
import { AllocationSliders } from './components/AllocationSliders';
import { RaceResults } from './components/RaceResults';
import { Standings } from './components/Standings';
import { SeasonComplete } from './components/SeasonComplete';
import { Leaderboard } from './components/Leaderboard';

export default function App() {
  const [state, dispatch] = useReducer(reducer, initialState);
  const [sim, setSim] = useState<SimAPI | null>(null);
  const [name, setName] = useState(displayName());
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [board, setBoard] = useState<LeaderboardEntry[] | null>(null);
  const simRef = useRef<SimAPI | null>(null);

  // The calendar renders from the API response, so the season loads without
  // waiting on the ~1.25 MB WASM module. The module is fetched alongside and
  // is only needed once the tenth decision is confirmed.
  useEffect(() => {
    fetchTodaySeason()
      .then((season) => dispatch({ type: 'SEASON_LOADED', season }))
      .catch((e: Error) => dispatch({ type: 'LOAD_FAILED', error: e.message }));

    loadSim()
      .then((api) => {
        simRef.current = api;
        setSim(api);
      })
      .catch((e: Error) => dispatch({ type: 'LOAD_FAILED', error: e.message }));
  }, []);

  // Once every round is allocated, play the whole season locally. There is
  // no network round-trip: this is the same simulation the server runs.
  useEffect(() => {
    if (state.phase !== 'reviewing' || !state.season) return;
    const api = simRef.current;
    if (!api) return;
    try {
      const result = api.runSeason(state.season.seed, state.decisions);
      dispatch({ type: 'SEASON_RESOLVED', result });
    } catch (e) {
      dispatch({ type: 'LOAD_FAILED', error: (e as Error).message });
    }
  }, [state.phase, state.season, state.decisions, sim]);

  const onAllocate = useCallback((area: keyof Decision, value: number) => {
    dispatch({ type: 'ALLOCATE', area, value });
  }, []);

  const onSubmit = useCallback(async () => {
    if (!state.season) return;
    setSubmitting(true);
    setSubmitError(null);
    try {
      setDisplayName(name);
      await submitRun(state.season.id, state.decisions, name || 'Anonymous');
      dispatch({ type: 'SUBMITTED' });
      setBoard(await fetchLeaderboard(state.season.id));
    } catch (e) {
      setSubmitError((e as Error).message);
    } finally {
      setSubmitting(false);
    }
  }, [state.season, state.decisions, name]);

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

  const { season } = state;
  const circuit = season.calendar[state.round];
  const playerRaces =
    state.result?.races.map((r) => {
      const me = r.cars.find((c) => c.team_id === 0)!;
      return { finish: me.finish, dnf: me.dnf, points: me.points };
    }) ?? undefined;

  return (
    <Shell>
      <div className="grid gap-5 lg:grid-cols-[1fr_340px]">
        <div className="space-y-5">
          {state.phase === 'allocating' && (
            <>
              <div className="rounded-lg border border-edge bg-panel px-5 py-4">
                <p className="text-xs uppercase tracking-widest text-muted">
                  Round {state.round + 1} of {season.calendar.length}
                </p>
                <h1 className="mt-1 text-2xl font-bold">{circuit.name}</h1>
                <p className="mt-1 text-sm text-muted">
                  Budget {budgetFor(state)} · {remaining(state)} unspent
                </p>
              </div>

              <AllocationSliders
                allocation={state.allocation}
                budget={budgetFor(state)}
                circuit={circuit}
                onChange={onAllocate}
              />

              <button
                data-testid="confirm-race"
                onClick={() => dispatch({ type: 'CONFIRM_RACE' })}
                className="w-full rounded bg-accent py-3 text-sm font-semibold text-white transition hover:brightness-110"
              >
                {state.round + 1 === season.calendar.length
                  ? 'Run the final race'
                  : `Run race ${state.round + 1}`}
                {spent(state.allocation) === 0 && ' (spending nothing)'}
              </button>
            </>
          )}

          {state.phase === 'reviewing' && (
            <div className="rounded-lg border border-edge bg-panel p-6">
              <p className="text-muted" data-testid="simulating">
                {sim ? 'Running the season…' : 'Loading the simulation…'}
              </p>
            </div>
          )}

          {state.phase === 'complete' && state.result && (
            <>
              <SeasonComplete
                result={state.result}
                onSubmit={onSubmit}
                submitting={submitting}
                submitted={state.submitted}
                error={submitError}
                name={name}
                onNameChange={setName}
              />
              {board && <Leaderboard entries={board} />}
              <Standings standings={state.result.standings} />
              <div className="space-y-4">
                {state.result.races.map((r) => (
                  <RaceResults key={r.round} race={r} teams={season.field} />
                ))}
              </div>
            </>
          )}
        </div>

        <aside className="space-y-5">
          <Calendar calendar={season.calendar} round={state.round} results={playerRaces} />
          <div className="rounded-lg border border-edge bg-panel p-4 text-xs text-muted">
            <p>
              Season {Number(season.seed) % 1000} · sim {season.sim_version}
            </p>
            <p className="mt-2">
              Everyone in the world plays this same season today — same calendar, same rivals, same
              luck. The leaderboard measures decisions and nothing else.
            </p>
          </div>
        </aside>
      </div>
    </Shell>
  );
}

function Shell({ children }: { children: React.ReactNode }) {
  return (
    <div className="mx-auto min-h-full max-w-5xl px-4 py-8">
      <header className="mb-6 flex items-baseline gap-3">
        <span className="flex gap-1" aria-hidden>
          {[0, 1, 2, 3, 4].map((i) => (
            <span key={i} className="h-2.5 w-2.5 rounded-full bg-accent" />
          ))}
        </span>
        <h1 className="text-sm font-semibold uppercase tracking-[0.2em]">Lights Out</h1>
      </header>
      {children}
    </div>
  );
}
