import type { RaceResult, Team } from '../game/types';

export function RaceResults({ race, teams }: { race: RaceResult; teams: Team[] }) {
  const name = (id: number) => teams.find((t) => t.id === id)?.name ?? `Team ${id}`;
  const classified = [...race.cars]
    .filter((c) => !c.dnf)
    .sort((a, b) => a.finish - b.finish);
  const retired = race.cars.filter((c) => c.dnf);

  return (
    <div className="rounded-lg border border-edge bg-panel p-4" data-testid="race-results">
      <div className="mb-3 flex items-baseline justify-between">
        <h2 className="text-xs font-semibold uppercase tracking-widest text-muted">
          Round {race.round} — {race.circuit}
        </h2>
        {race.safety_car && <span className="text-xs text-amber-300">Safety car</span>}
      </div>
      <ol className="space-y-0.5">
        {classified.map((c) => (
          <li
            key={c.team_id}
            className={`flex items-baseline gap-3 text-sm ${
              c.team_id === 0 ? 'font-semibold text-accent' : ''
            }`}
          >
            <span className="w-5 text-right font-mono text-xs text-muted">P{c.finish}</span>
            <span className="flex-1 truncate">{name(c.team_id)}</span>
            <span className="font-mono text-xs text-muted">from P{c.grid}</span>
            <span className="w-8 text-right font-mono tabular-nums">{c.points || ''}</span>
          </li>
        ))}
        {retired.map((c) => (
          <li
            key={c.team_id}
            className={`flex items-baseline gap-3 text-sm opacity-60 ${
              c.team_id === 0 ? 'font-semibold text-accent opacity-100' : ''
            }`}
          >
            <span className="w-5 text-right font-mono text-xs text-accent">✖</span>
            <span className="flex-1 truncate">{name(c.team_id)}</span>
            <span className="font-mono text-xs text-muted">DNF</span>
            <span className="w-8" />
          </li>
        ))}
      </ol>
    </div>
  );
}
