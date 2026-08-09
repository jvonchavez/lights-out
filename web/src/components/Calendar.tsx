import { ARCHETYPE_LABELS, type Circuit } from '../game/types';

const dotFor: Record<Circuit['archetype'], string> = {
  power: 'bg-rose-500',
  technical: 'bg-sky-400',
  balanced: 'bg-slate-400',
  highspeed: 'bg-amber-400',
};

/**
 * The whole calendar is visible from round one. That visibility is what
 * makes this a game rather than a spreadsheet: you can see that rounds 4
 * through 7 are power circuits and invest in engine two races early.
 */
export function Calendar({
  calendar,
  round,
  results,
}: {
  calendar: Circuit[];
  round: number;
  results?: { finish: number; dnf: boolean; points: number }[];
}) {
  return (
    <div className="rounded-lg border border-edge bg-panel p-4">
      <h2 className="mb-3 text-xs font-semibold uppercase tracking-widest text-muted">Calendar</h2>
      <ol className="space-y-1">
        {calendar.map((c, i) => {
          const done = results?.[i];
          // The round being watched, whether or not earlier rounds have
          // results yet -- during the reel the highlight IS the playhead.
          const current = i === round;
          return (
            <li
              key={c.name}
              className={`flex items-center gap-3 rounded px-2 py-1.5 text-sm ${
                current ? 'bg-accent/15 ring-1 ring-accent/40' : ''
              } ${results && !done ? 'opacity-40' : ''}`}
            >
              <span className="w-5 text-right font-mono text-xs text-muted">{i + 1}</span>
              <span className={`h-2 w-2 shrink-0 rounded-full ${dotFor[c.archetype]}`} />
              <span className="flex-1 truncate">{c.name}</span>
              <span className="text-xs text-muted">{ARCHETYPE_LABELS[c.archetype]}</span>
              {done && (
                <span
                  className={`w-12 text-right font-mono text-xs ${
                    done.dnf ? 'text-accent' : 'text-emerald-400'
                  }`}
                >
                  {done.dnf ? 'DNF' : `P${done.finish}`}
                </span>
              )}
            </li>
          );
        })}
      </ol>
    </div>
  );
}
