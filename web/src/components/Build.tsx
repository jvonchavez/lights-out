import { effectLabel, type Card } from '../game/types';

/** The five parts you assembled. The artifact worth arguing about. */
export function Build({ build }: { build: Card[] }) {
  if (build.length === 0) return null;
  return (
    <div className="rounded-lg border border-edge bg-panel p-4" data-testid="build">
      <h2 className="mb-3 text-xs font-semibold uppercase tracking-widest text-muted">Your car</h2>
      <ol className="space-y-1.5">
        {build.map((c, i) => (
          <li key={`${c.id}-${i}`} className="flex items-baseline gap-3 text-sm">
            <span className="w-4 text-right font-mono text-xs text-muted">{i + 1}</span>
            <span className="flex-1 truncate" data-testid="build-part-name">
              {c.name}
            </span>
            <span className="font-mono text-xs text-emerald-400">{effectLabel(c)}</span>
          </li>
        ))}
      </ol>
    </div>
  );
}
