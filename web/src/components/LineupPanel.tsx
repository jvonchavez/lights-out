import {
  carOverall,
  driverOverall,
  engineerOverall,
  ITEM_CAR,
  ITEM_DRIVER_A,
  ITEM_ENGINEER,
  ITEM_PRINCIPAL,
  principalOverall,
  ROLL_COUNT,
  SLOT_OF,
  type TeamEra,
} from '../game/types';

interface Props {
  rolls: TeamEra[];
  picks: number[];
}

interface Filled {
  slot: string;
  name: string;
  from: string;
  overall: number;
}

/**
 * The team as it stands. It fills one row per roll, so the cost of a pick is
 * visible as the empty rows it leaves behind.
 */
export function LineupPanel({ rolls, picks }: Props) {
  const filled: Filled[] = [];
  picks.forEach((kind, i) => {
    const te = rolls[i];
    if (!te) return;
    const from = `${te.year} ${te.team}`;
    if (kind === ITEM_CAR) {
      filled.push({ slot: 'Car', name: te.car.name, from, overall: carOverall(te.car) });
    } else if (SLOT_OF[kind] === 'driver') {
      const d = te.drivers[kind === ITEM_DRIVER_A ? 0 : 1];
      filled.push({ slot: 'Driver', name: d.name, from, overall: driverOverall(d) });
    } else if (kind === ITEM_ENGINEER) {
      filled.push({
        slot: 'Engineer',
        name: te.engineer.name,
        from,
        overall: engineerOverall(te.engineer),
      });
    } else if (kind === ITEM_PRINCIPAL) {
      filled.push({
        slot: 'Principal',
        name: te.principal.name,
        from,
        overall: principalOverall(te.principal),
      });
    }
  });

  const order = ['Car', 'Driver', 'Driver', 'Engineer', 'Principal'];
  const remaining = [...order];
  for (const f of filled) {
    const i = remaining.indexOf(f.slot);
    if (i >= 0) remaining.splice(i, 1);
  }

  return (
    <div className="rounded-lg border border-edge bg-panel p-4" data-testid="lineup">
      <h2 className="mb-3 text-[11px] uppercase tracking-widest text-muted">Your team</h2>
      <ul className="space-y-2">
        {filled.map((f, i) => (
          <li key={i} className="flex items-baseline gap-2 text-sm" data-testid="lineup-row">
            <span className="w-16 shrink-0 text-[10px] uppercase tracking-widest text-muted">
              {f.slot}
            </span>
            <span className="min-w-0 flex-1">
              <span className="block truncate font-medium" data-testid="lineup-name">
                {f.name}
              </span>
              <span className="block truncate text-[11px] text-muted">{f.from}</span>
            </span>
            <span className="font-mono text-sm tabular-nums">{f.overall}</span>
          </li>
        ))}
        {remaining.map((slot, i) => (
          <li key={`empty-${i}`} className="flex items-baseline gap-2 text-sm opacity-30">
            <span className="w-16 shrink-0 text-[10px] uppercase tracking-widest text-muted">
              {slot}
            </span>
            <span className="flex-1 border-b border-dashed border-edge" />
          </li>
        ))}
      </ul>
      <p className="mt-3 text-[11px] text-muted">
        {picks.length} of {ROLL_COUNT} rolls used
      </p>
    </div>
  );
}
