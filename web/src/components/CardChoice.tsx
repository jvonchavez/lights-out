import { ARCHETYPE_LABELS, effectLabel, riskPips, type Card, type Circuit } from '../game/types';

/**
 * Three dealt parts; take one and it locks. This is the whole decision.
 *
 * Randomness lands BEFORE the choice, so you are reacting to a hand you
 * were dealt rather than filling in the same form every round. The risk
 * pips are computed from the card's own effect, so what is shown is what
 * the simulation will apply.
 */
export function CardChoice({
  deal,
  picked,
  races,
  onPick,
  onConfirm,
}: {
  deal: Card[];
  picked: number | null;
  races: Circuit[];
  onPick: (i: number) => void;
  onConfirm: () => void;
}) {
  return (
    <div className="space-y-4">
      <div className="grid gap-3 sm:grid-cols-3">
        {deal.map((c, i) => {
          const pips = riskPips(c);
          const on = picked === i;
          return (
            <button
              key={c.id}
              data-testid={`card-${i}`}
              onClick={() => onPick(i)}
              aria-pressed={on}
              className={`flex h-full flex-col rounded-lg border p-4 text-left transition ${
                on
                  ? 'border-accent bg-accent/10 ring-1 ring-accent'
                  : 'border-edge bg-panel hover:border-muted'
              }`}
            >
              <h3 className="text-sm font-semibold uppercase tracking-wide">{c.name}</h3>
              <p className="mt-2 font-mono text-sm text-emerald-400">{effectLabel(c)}</p>
              <p className="mt-2 flex-1 text-xs leading-relaxed text-muted">{c.blurb}</p>
              <p className="mt-3 flex items-center gap-1 text-xs text-muted">
                <span>risk</span>
                {[1, 2, 3, 4, 5].map((n) => (
                  <span
                    key={n}
                    className={`h-1.5 w-1.5 rounded-full ${
                      n <= pips ? 'bg-accent' : 'bg-edge'
                    }`}
                  />
                ))}
              </p>
            </button>
          );
        })}
      </div>

      <button
        data-testid="confirm-pick"
        onClick={onConfirm}
        disabled={picked === null}
        className="w-full rounded bg-accent py-3 text-sm font-semibold text-white transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-40"
      >
        {picked === null
          ? 'Choose a part'
          : `Fit ${deal[picked].name} — it locks`}
      </button>

      <p className="text-center text-xs text-muted">
        Next up: {races.map((r) => `${r.name} (${ARCHETYPE_LABELS[r.archetype]})`).join(' · ')}
      </p>
    </div>
  );
}
