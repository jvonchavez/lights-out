import { standingsAfter } from '../game/reducer';
import {
  ARCHETYPE_LABELS,
  DNF_LABELS,
  type Circuit,
  type EntryResult,
  type SeasonResult,
} from '../game/types';

/** How many finishers to show per race. The rest is 24 rows of noise. */
const SHOWN = 8;

interface Props {
  result: SeasonResult;
  calendar: Circuit[];
  /** How many races have played. The race being watched is round - 1. */
  round: number;
  colours: Map<number, string>;
  names: Map<number, string>;
  onSkip: () => void;
}

/**
 * The season, one race at a time.
 *
 * Every number here is already known -- the whole result was computed before
 * the first frame. The reel is presentation and nothing else, which is why
 * RunPartial could be deleted: there are no in-season decisions left to give
 * feedback on, so there is nothing the client needs the sim to resolve
 * incrementally.
 */
export function RaceReel({ result, calendar, round, colours, names, onSkip }: Props) {
  const idx = Math.max(0, Math.min(round - 1, result.races.length - 1));
  const race = result.races[idx];
  const circuit = calendar[idx];
  const table = standingsAfter(result, round);
  const lead = Math.max(1, table[0]?.points ?? 1);

  const classified = [...race.entries]
    .filter((e) => !e.dnf)
    .sort((a, b) => a.finish - b.finish);
  const mine = race.entries.filter((e) => e.team_id === 0);
  const shown = classified.slice(0, SHOWN);
  const missing = mine.filter((e) => !e.dnf && !shown.includes(e));

  return (
    <div className="space-y-4" data-testid="reel">
      <div className="flex items-center justify-between gap-4">
        <div>
          <p className="text-[11px] uppercase tracking-widest text-muted">
            Round {race.round} of {result.races.length}
          </p>
          <h2 className="text-2xl font-semibold tracking-tight" data-testid="reel-circuit">
            {race.circuit}
          </h2>
          <p className="text-xs text-muted">
            {circuit ? ARCHETYPE_LABELS[circuit.archetype] : ''}
            {race.safety_car && <span className="ml-2 text-amber-400">Safety car</span>}
          </p>
        </div>
        <button
          type="button"
          onClick={onSkip}
          data-testid="skip-reel"
          className="rounded border border-edge px-3 py-1.5 text-xs text-muted transition hover:border-muted hover:text-white"
        >
          Skip to result
        </button>
      </div>

      <div className="grid gap-4 md:grid-cols-[1fr_260px]">
        <ol className="rounded-lg border border-edge bg-panel">
          {shown.map((e, i) => (
            <Row key={`${e.team_id}-${e.entry}`} e={e} i={i} colours={colours} names={names} />
          ))}
          {missing.map((e) => (
            <Row
              key={`${e.team_id}-${e.entry}`}
              e={e}
              i={SHOWN}
              colours={colours}
              names={names}
              elided
            />
          ))}
          {race.entries
            .filter((e) => e.dnf)
            .map((e) => (
              <Row
                key={`dnf-${e.team_id}-${e.entry}`}
                e={e}
                i={SHOWN + 1}
                colours={colours}
                names={names}
              />
            ))
            .slice(0, 3)}
        </ol>

        <div className="rounded-lg border border-edge bg-panel p-3">
          <h3 className="mb-2 text-[11px] uppercase tracking-widest text-muted">Constructors</h3>
          <ol className="space-y-1.5">
            {table.slice(0, 8).map((s) => (
              <li key={s.team_id} className="text-[11px]">
                <div className="flex items-baseline justify-between gap-2">
                  <span className={s.team_id === 0 ? 'font-semibold text-accent' : ''}>
                    {s.name}
                  </span>
                  <span className="font-mono tabular-nums">{s.points}</span>
                </div>
                <div className="mt-0.5 h-1.5 overflow-hidden rounded-full bg-track">
                  <div
                    className="h-full rounded-full transition-[width] duration-700 ease-out"
                    style={{
                      width: `${(s.points / lead) * 100}%`,
                      background: colours.get(s.team_id) ?? '#8b95a1',
                    }}
                  />
                </div>
              </li>
            ))}
          </ol>
        </div>
      </div>
    </div>
  );
}

function Row({
  e,
  i,
  colours,
  names,
  elided,
}: {
  e: EntryResult;
  i: number;
  colours: Map<number, string>;
  names: Map<number, string>;
  elided?: boolean;
}) {
  const mine = e.team_id === 0;
  const delta = e.dnf ? 0 : e.grid - e.finish;
  return (
    <li
      // Staggered by position so the order arrives rather than appears. The
      // animation is keyed on the round in App, so it replays every race.
      style={{ animationDelay: `${i * 55}ms`, borderLeftColor: colours.get(e.team_id) ?? '#232a32' }}
      className={[
        'reel-row flex items-center gap-3 border-b border-l-2 border-edge/60 px-3 py-2 text-sm last:border-b-0',
        mine ? 'bg-accent/10' : '',
        e.dnf ? 'opacity-60' : '',
        elided ? 'border-t border-t-edge' : '',
      ].join(' ')}
    >
      <span className="w-8 shrink-0 font-mono tabular-nums text-muted">
        {e.dnf ? '—' : `P${e.finish}`}
      </span>
      <span className={`min-w-0 flex-1 truncate ${mine ? 'font-semibold text-accent' : ''}`}>
        {e.driver}
        <span className="ml-2 text-[11px] text-muted">{names.get(e.team_id)}</span>
      </span>
      {e.dnf ? (
        <span className="shrink-0 text-[11px] uppercase tracking-wide text-accent">
          {e.dnf_reason ? DNF_LABELS[e.dnf_reason] : 'DNF'}
        </span>
      ) : (
        <>
          <span
            className={`w-10 shrink-0 text-right font-mono text-[11px] tabular-nums ${
              delta > 0 ? 'text-emerald-400' : delta < 0 ? 'text-rose-400' : 'text-muted'
            }`}
          >
            {delta > 0 ? `+${delta}` : delta < 0 ? delta : '·'}
          </span>
          <span className="w-8 shrink-0 text-right font-mono tabular-nums">
            {e.points > 0 ? e.points : ''}
          </span>
        </>
      )}
    </li>
  );
}
