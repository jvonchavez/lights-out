import type { DriverStanding } from '../game/types';

/**
 * The drivers' table decides nothing -- the goal is the constructors'
 * championship -- but seeing which of your two carried the team is most of
 * the reason to have two.
 */
export function DriverStandings({
  drivers,
  colours,
}: {
  drivers: DriverStanding[];
  colours: Map<number, string>;
}) {
  return (
    <div className="rounded-lg border border-edge bg-panel p-4" data-testid="driver-standings">
      <h2 className="mb-3 text-xs font-semibold uppercase tracking-widest text-muted">Drivers</h2>
      <ol className="space-y-1">
        {drivers.slice(0, 10).map((d, i) => (
          <li
            key={`${d.team_id}-${d.driver_id}-${i}`}
            className={`flex items-baseline gap-2 text-sm ${
              d.team_id === 0 ? 'font-semibold text-accent' : ''
            }`}
          >
            <span className="w-5 shrink-0 font-mono tabular-nums text-muted">{i + 1}</span>
            <span
              className="h-2.5 w-1 shrink-0 rounded-sm"
              style={{ background: colours.get(d.team_id) ?? '#8b95a1' }}
            />
            <span className="min-w-0 flex-1 truncate">{d.name}</span>
            <span className="font-mono tabular-nums">{d.points}</span>
          </li>
        ))}
      </ol>
    </div>
  );
}
