import { useCallback, useEffect, useReducer, useRef, useState } from 'react';
import { reducer, initialState, currentDeal, revealComplete } from './game/reducer';
import { loadSim, type SimAPI } from './game/wasm';
import {
  fetchLeaderboard,
  fetchTodaySeason,
  submitRun,
  displayName,
  setDisplayName,
} from './game/api';
import type { LeaderboardEntry } from './game/types';
import { Calendar } from './components/Calendar';
import { CardChoice } from './components/CardChoice';
import { Build } from './components/Build';
import { RaceResults } from './components/RaceResults';
import { Standings } from './components/Standings';
import { SeasonComplete } from './components/SeasonComplete';
import { Leaderboard } from './components/Leaderboard';

/** How long each race lingers before the next is revealed. */
const REVEAL_MS = 700;

export default function App() {
  const [state, dispatch] = useReducer(reducer, initialState);
  const [sim, setSim] = useState<SimAPI | null>(null);
  const [name, setName] = useState(displayName());
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [board, setBoard] = useState<LeaderboardEntry[] | null>(null);
  const simRef = useRef<SimAPI | null>(null);

  // The deals come from the API, so the first choice renders without
  // waiting on the ~1.26 MB WASM module. It is needed only once a pick is
  // committed and races have to resolve.
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

  // Each committed pick unlocks two more races. Resolve just those, so a
  // window has a visible consequence instead of ten races arriving at once.
  useEffect(() => {
    if (state.phase !== 'racing' || !state.season) return;
    const api = simRef.current;
    if (!api) return;
    try {
      const result = api.runPartial(state.season.seed, state.picks);
      dispatch({ type: 'RACES_RESOLVED', result });
    } catch (e) {
      dispatch({ type: 'LOAD_FAILED', error: (e as Error).message });
    }
  }, [state.phase, state.season, state.picks, sim]);

  // Reveal newly resolved races one at a time, then move on.
  useEffect(() => {
    if (!state.result) return;
    if (!revealComplete(state)) {
      const t = setTimeout(() => dispatch({ type: 'REVEAL_NEXT' }), REVEAL_MS);
      return () => clearTimeout(t);
    }
    if (state.phase === 'racing') {
      const t = setTimeout(() => dispatch({ type: 'NEXT_WINDOW' }), REVEAL_MS);
      return () => clearTimeout(t);
    }
  }, [state.result, state.revealed, state.phase]);

  const onSubmit = useCallback(async () => {
    if (!state.season) return;
    setSubmitting(true);
    setSubmitError(null);
    try {
      setDisplayName(name);
      await submitRun(state.season.id, state.picks, name || 'Anonymous');
      dispatch({ type: 'SUBMITTED' });
      setBoard(await fetchLeaderboard(state.season.id));
    } catch (e) {
      setSubmitError((e as Error).message);
    } finally {
      setSubmitting(false);
    }
  }, [state.season, state.picks, name]);

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
  const shown = state.result?.races.slice(0, state.revealed) ?? [];
  const playerRaces = shown.map((r) => {
    const me = r.cars.find((c) => c.team_id === 0)!;
    return { finish: me.finish, dnf: me.dnf, points: me.points };
  });
  const windowRound = season.window_rounds[state.window] ?? 0;
  const nextRaces = season.calendar.slice(windowRound, windowRound + 2);

  return (
    <Shell>
      <div className="grid gap-5 lg:grid-cols-[1fr_340px]">
        <div className="space-y-5">
          {state.phase === 'choosing' && (
            <>
              <div className="rounded-lg border border-edge bg-panel px-5 py-4">
                <p className="text-xs uppercase tracking-widest text-muted">
                  Development window {state.window + 1} of {season.deals.length}
                </p>
                <h1 className="mt-1 text-2xl font-bold">Choose one part</h1>
                <p className="mt-1 text-sm text-muted">
                  It fits for the rest of the season. Every part you bolt on raises performance and
                  lowers reliability.
                </p>
              </div>
              <CardChoice
                deal={currentDeal(state)}
                picked={state.pick}
                races={nextRaces}
                onPick={(i) => dispatch({ type: 'PICK_CARD', index: i })}
                onConfirm={() => dispatch({ type: 'CONFIRM_PICK' })}
              />
            </>
          )}

          {state.phase === 'racing' && (
            <div className="rounded-lg border border-edge bg-panel px-5 py-4" data-testid="racing">
              <p className="text-xs uppercase tracking-widest text-muted">Racing</p>
              <h1 className="mt-1 text-2xl font-bold">
                {shown.length > 0 ? shown[shown.length - 1].circuit : 'Lights out…'}
              </h1>
            </div>
          )}

          {state.phase === 'complete' && state.result && (
            <SeasonComplete
              result={state.result}
              onSubmit={onSubmit}
              submitting={submitting}
              submitted={state.submitted}
              error={submitError}
              name={name}
              onNameChange={setName}
            />
          )}

          {state.result && state.result.build.length > 0 && <Build build={state.result.build} />}
          {board && <Leaderboard entries={board} />}
          {state.phase === 'complete' && state.result && (
            <Standings standings={state.result.standings} />
          )}

          <div className="space-y-4">
            {[...shown].reverse().map((r) => (
              <RaceResults key={r.round} race={r} teams={season.field} />
            ))}
          </div>
        </div>

        <aside className="space-y-5">
          <Calendar
            calendar={season.calendar}
            round={windowRound}
            results={playerRaces.length > 0 ? playerRaces : undefined}
          />
          <div className="rounded-lg border border-edge bg-panel p-4 text-xs text-muted">
            <p>
              Season {Number(season.seed) % 1000} · sim {season.sim_version}
            </p>
            <p className="mt-2">
              Everyone in the world is dealt these same parts today. The leaderboard measures what
              you did with them.
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
