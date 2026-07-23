import { AREA_LABELS, type Circuit, type Decision } from '../game/types';

const hint: Record<keyof Decision, string> = {
  chassis: 'Cornering — dominant at technical circuits',
  engine: 'Top speed — dominant at power circuits',
  aero: 'Multiplies the whole car, up to +15%',
};

/**
 * Four sliders sharing one budget. The maximum of each is computed from
 * what the others have already taken, so an over-budget allocation is
 * impossible to express rather than merely rejected on submit.
 */
export function AllocationSliders({
  allocation,
  budget,
  circuit,
  onChange,
}: {
  allocation: Decision;
  budget: number;
  circuit: Circuit;
  onChange: (area: keyof Decision, value: number) => void;
}) {
  const used = allocation.chassis + allocation.engine + allocation.aero;
  const left = budget - used;
  const areas = Object.keys(AREA_LABELS) as (keyof Decision)[];

  const weight: Record<keyof Decision, number> = {
    chassis: circuit.profile.chassis,
    engine: circuit.profile.engine,
    aero: circuit.profile.aero,
  };

  return (
    <div className="rounded-lg border border-edge bg-panel p-5">
      <div className="mb-4 flex items-baseline justify-between">
        <h2 className="text-xs font-semibold uppercase tracking-widest text-muted">Development</h2>
        <div className="font-mono text-sm">
          <span className={left === 0 ? 'text-emerald-400' : 'text-amber-300'}>{left}</span>
          <span className="text-muted"> / {budget} left</span>
        </div>
      </div>

      <div className="space-y-5">
        {areas.map((area) => {
          const max = allocation[area] + left;
          return (
            <div key={area}>
              <div className="mb-1 flex items-baseline justify-between gap-2">
                <label htmlFor={`slider-${area}`} className="text-sm font-medium">
                  {AREA_LABELS[area]}
                  <span className="ml-2 font-mono text-xs text-muted">
                    ×{(weight[area] / 1000).toFixed(2)}
                  </span>
                </label>
                <span className="font-mono text-sm tabular-nums">{allocation[area]}</span>
              </div>
              <input
                id={`slider-${area}`}
                data-testid={`slider-${area}`}
                type="range"
                min={0}
                max={max}
                value={allocation[area]}
                onChange={(e) => onChange(area, Number(e.target.value))}
                className="w-full accent-[var(--color-accent)]"
              />
              <p className="mt-0.5 text-xs text-muted">{hint[area]}</p>
            </div>
          );
        })}
      </div>

      <p className="mt-5 border-t border-edge pt-3 text-xs text-muted">
        Every point spent raises performance and lowers reliability — permanently, for the rest of
        the season. Risk grows with the square of what you have spent, so the last points cost far
        more than the first. A DNF scores zero, and there are only ten races.
      </p>
    </div>
  );
}
